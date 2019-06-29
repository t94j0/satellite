package handler

import (
	"io"
	"io/ioutil"
	"net/url"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	"github.com/t94j0/ja3-server/crypto/tls"
	"github.com/t94j0/ja3-server/net/http"
	"github.com/t94j0/ja3-server/net/http/httputil"
	"github.com/t94j0/satellite/path"
)

// RootHandler is used to serve HTTP(S)
type RootHandler struct {
	paths            *path.Paths
	serverHeader     string
	managementIP     string
	managementPath   string
	indexPath        string
	notFoundRedirect string
	notFoundRender   string
	redirectHTTP     bool
	identifier       *path.ClientID
}

// New creates a new RootHandler object
func NewRoot(paths *path.Paths, cfg *viper.Viper) RootHandler {
	return RootHandler{
		paths:            paths,
		serverHeader:     cfg.GetString("server_header"),
		managementIP:     cfg.GetString("management.ip"),
		managementPath:   cfg.GetString("management.path"),
		indexPath:        cfg.GetString("index"),
		notFoundRedirect: cfg.GetString("not_found.redirect"),
		notFoundRender:   cfg.GetString("not_found.render"),
		identifier:       path.NewClientID(),
	}
}

// handler manages all path handling. It redirects the task of handling based on
// if the file exist, the file should be hosted (based on Path rules), and if
// the file should not be hosted
func (s RootHandler) ServeHTTP(w http.ResponseWriter, req *http.Request) {
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

func (s RootHandler) writeHeaders(w http.ResponseWriter, headers map[string]string) {
	for name, value := range headers {
		w.Header().Add(name, value)
	}
}

// render will render the target path no matter what
func (s RootHandler) render(w http.ResponseWriter, req *http.Request, path *path.Path) {
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
func (s RootHandler) matchHandler(w http.ResponseWriter, req *http.Request, path *path.Path) {
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
func (s RootHandler) noMatchHandler(w http.ResponseWriter, req *http.Request, path *path.Path) {
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
func (s RootHandler) doesNotExistHandler(w http.ResponseWriter, req *http.Request) {
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
