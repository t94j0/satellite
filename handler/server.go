package handler

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"io"
	"io/ioutil"
	"net"
	"net/url"
	"path/filepath"

	"github.com/apcera/util/iprange"
	log "github.com/sirupsen/logrus"
	"github.com/t94j0/ja3-server/crypto/tls"
	"github.com/t94j0/ja3-server/net/http"
	"github.com/t94j0/ja3-server/net/http/httputil"
	"github.com/t94j0/satellite/path"
	"github.com/t94j0/satellite/util"
	"gopkg.in/yaml.v2"
)

// Server is used to serve HTTP(S)
type Server struct {
	paths            *path.Paths
	serverPath       string
	port             string
	keyPath          string
	certPath         string
	serverHeader     string
	managementIP     string
	managementPath   string
	indexPath        string
	notFoundRedirect string
	notFoundRender   string
	redirectHTTP     bool
	identifier       *path.ClientID
}

// ErrNotFoundConfig is an error when both redirection options are present
var ErrNotFoundConfig = errors.New("both not_found redirect and render cannot be set at the same time")

// NewServer creates a new Server object
func New(paths *path.Paths, serverPath, port, keyPath, certPath, serverHeader, managementIP, managementPath, indexPath, notFoundRedirect, notFoundRender string, redirectHTTP bool) (Server, error) {
	if notFoundRedirect != "" && notFoundRender != "" {
		return Server{}, ErrNotFoundConfig
	}
	return Server{
		paths:            paths,
		serverPath:       serverPath,
		port:             port,
		keyPath:          keyPath,
		certPath:         certPath,
		serverHeader:     serverHeader,
		managementIP:     managementIP,
		managementPath:   managementPath,
		indexPath:        indexPath,
		notFoundRedirect: notFoundRedirect,
		notFoundRender:   notFoundRender,
		redirectHTTP:     redirectHTTP,
		identifier:       path.NewClientID(),
	}, nil
}

// Start makes the server begin listening
func (s Server) Start() error {
	// HTTP
	if s.redirectHTTP {
		go func() {
			if err := s.createHTTPRedirect(); err != nil {
				log.Error(err)
			}
		}()
	}

	mux := http.NewServeMux()
	if s.managementPath != "" {
		mux.HandleFunc(s.managementPath, s.managementHandler)
		mux.HandleFunc(s.managementPath+"/reset", s.resetHandler)
		mux.HandleFunc(s.managementPath+"/new", s.uploadHandler)
	}
	mux.HandleFunc("/", s.handler)

	server := &http.Server{Addr: s.port, Handler: mux}
	ln, err := net.Listen("tcp", s.port)
	if err != nil {
		return err
	}
	defer ln.Close()
	cert, err := tls.LoadX509KeyPair(s.certPath, s.keyPath)
	if err != nil {
		return err
	}
	tlsConfig := tls.Config{Certificates: []tls.Certificate{cert}}
	tlsListener := tls.NewListener(ln, &tlsConfig)
	return server.Serve(tlsListener)
}

// createHTTPRedirect creates a HTTP listener to redirect to HTTPS
func (s Server) createHTTPRedirect() error {
	log.Debug("HTTP redirection not implemented")
	return nil
}

// handler manages all path handling. It redirects the task of handling based on
// if the file exist, the file should be hosted (based on Path rules), and if
// the file should not be hosted
func (s Server) handler(w http.ResponseWriter, req *http.Request) {
	log.WithFields(log.Fields{
		"method":      req.Method,
		"host":        req.Host,
		"remote_addr": req.RemoteAddr,
		"req_uri":     req.RequestURI,
	}).Info("request")

	if req.URL.Path == "/" && s.indexPath != "" {
		req.URL.Path = s.indexPath
	}

	path, exists := s.paths.Match(req.URL.Path)

	if s.serverHeader != "" {
		w.Header().Add("Server", s.serverHeader)
	}

	if !exists {
		s.doesNotExistHandler(w, req)
		return
	}

	s.writeHeaders(w, path.AddHeaders)

	isHost, err := path.ShouldHost(req, s.identifier)
	if err != nil {
		log.Println(err)
	}

	if isHost {
		s.matchHandler(w, req, path)
	} else {
		s.noMatchHandler(w, req, path)
	}
}

func (s Server) writeHeaders(w http.ResponseWriter, headers map[string]string) {
	for name, value := range headers {
		w.Header().Add(name, value)
	}
}

// render will render the target path no matter what
func (s Server) render(w http.ResponseWriter, req *http.Request, path *path.Path) {
	if path.ProxyHost != "" {
		proxyURL, err := url.Parse(path.ProxyHost)
		if err != nil {
			log.Error(err)
		}
		proxy := httputil.NewSingleHostReverseProxy(proxyURL)
		tr := &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		}
		proxy.Transport = tr
		proxy.ServeHTTP(w, req)
	} else {
		data, err := ioutil.ReadFile(path.FullPath)
		if err != nil {
			log.Error(err)
			return
		}
		io.WriteString(w, string(data))
	}
}

// matchHandler is the handler for if the file exists and the rules matched the
// request. This means serve the target file
func (s Server) matchHandler(w http.ResponseWriter, req *http.Request, path *path.Path) {
	s.writeHeaders(w, path.AddHeadersSuccess)
	s.writeHeaders(w, path.ContentHeaders())
	s.render(w, req, path)
	if path.ShouldRemove() {
		path.Remove()
	}
	return
}

// noMatchHandler is the handler used when the file exists, but the rules
// determine the request does not match
func (s Server) noMatchHandler(w http.ResponseWriter, req *http.Request, path *path.Path) {
	s.writeHeaders(w, path.AddHeadersFailure)
	if path.OnFailure.Redirect != "" {
		http.Redirect(w, req, path.OnFailure.Redirect, 301)
	} else if path.OnFailure.Render != "" {
		newPath, found := s.paths.Match(path.OnFailure.Render)
		if !found {
			log.Println("Error: failure path not found")
			io.WriteString(w, "Errored\n")
			return
		}
		s.render(w, req, newPath)
	} else {
		s.doesNotExistHandler(w, req)
	}
}

// doesNotExistHandler is 404
func (s Server) doesNotExistHandler(w http.ResponseWriter, req *http.Request) {
	if s.notFoundRedirect != "" {
		http.Redirect(w, req, s.notFoundRedirect, 301)
	} else if s.notFoundRender != "" {
		path, found := s.paths.Match(s.notFoundRender)
		if !found {
			log.Println("Error: not_found render page not found")
			return
		}
		s.matchHandler(w, req, path)
	} else {
		io.WriteString(w, "404\n")
	}
}

// managementHandler handles management information
func (s Server) managementEnabled(req *http.Request) (bool, error) {
	mgmtRange, err := iprange.ParseIPRange(s.managementIP)
	if err != nil {
		return false, err
	}
	targetHost := util.GetHost(req)
	return mgmtRange.Contains(targetHost), nil
}

func (s Server) managementHandler(w http.ResponseWriter, req *http.Request) {
	outPaths := s.paths.Out()
	w.Header().Add("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(outPaths); err != nil {
		log.Println(err)
		s.doesNotExistHandler(w, req)
		return
	}
}

func (s Server) resetHandler(w http.ResponseWriter, req *http.Request) {
	type Body struct {
		Path string `json:"path"`
	}
	var mgmt Body

	if req.Body == nil {
		log.Println("No HTTP body")
		s.doesNotExistHandler(w, req)
		return
	}
	if err := json.NewDecoder(req.Body).Decode(&mgmt); err != nil {
		log.Println(err)
		s.doesNotExistHandler(w, req)
		return
	}

	path, _ := s.paths.Match(mgmt.Path)

	path.NotServing = false
	path.TimesServed = 0

	if err := path.Write(); err != nil {
		log.Println(err)
		s.doesNotExistHandler(w, req)
		return
	}
}

func (s Server) uploadHandler(w http.ResponseWriter, req *http.Request) {
	type Body struct {
		Path path.Path `json:"path"`
		File string    `json:"file"`
	}

	var newPath Body

	// Decode body
	if err := json.NewDecoder(req.Body).Decode(newPath); err != nil {
		log.Println(err)
		http.Error(w, "failure to decode body", http.StatusBadRequest)
		return
	}

	// Decode file as b64
	fileData, err := base64.StdEncoding.DecodeString(newPath.File)
	if err != nil {
		log.Println(err)
		http.Error(w, "failure to base64 decode body", http.StatusBadRequest)
		return
	}

	path := filepath.Join(s.serverPath, newPath.Path.Path)

	// Write data file
	if err := ioutil.WriteFile(path, fileData, 0644); err != nil {
		log.Println(err)
		http.Error(w, "failure to write data file", http.StatusBadRequest)
		return
	}

	// Marshal info file to yaml
	infoData, err := yaml.Marshal(&newPath.Path)
	if err != nil {
		log.Println(err)
		http.Error(w, "failure marshal info to yaml", http.StatusBadRequest)
		return
	}

	// Write YAML
	if err := ioutil.WriteFile(path+".info", infoData, 0644); err != nil {
		log.Println(err)
		http.Error(w, "failure to write data file", http.StatusBadRequest)
		return
	}
}
