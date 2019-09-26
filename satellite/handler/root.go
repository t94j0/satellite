package handler

import (
	"crypto/md5"
	"encoding/hex"
	"errors"
	"io"
	"net"
	rhttp "net/http"

	log "github.com/sirupsen/logrus"
	"github.com/t94j0/satellite/crypto/tls"
	"github.com/t94j0/satellite/net/http"
	"github.com/t94j0/satellite/satellite/path"
)

// Server is used to serve HTTP(S)
type Server struct {
	paths            *path.Paths
	serverPath       string
	port             string
	keyPath          string
	certPath         string
	serverHeader     string
	indexPath        string
	notFoundRedirect string
	notFoundRender   string
	redirectHTTP     bool
	identifier       *path.ClientID
}

// ErrNotFoundConfig is an error when both redirection options are present
var ErrNotFoundConfig = errors.New("both not_found redirect and render cannot be set at the same time")

// New creates a new Server object
func New(paths *path.Paths, serverPath, port, keyPath, certPath, serverHeader, indexPath, notFoundRedirect, notFoundRender string, redirectHTTP bool) (Server, error) {
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
			s.createHTTPRedirect()
		}()
	}

	mux := http.NewServeMux()
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
func (s Server) createHTTPRedirect() {
	rhttp.ListenAndServe(":80", rhttp.HandlerFunc(func(w rhttp.ResponseWriter, req *rhttp.Request) {
		target := "https://" + req.Host + req.URL.Path
		if len(req.URL.RawQuery) > 0 {
			target += "?" + req.URL.RawQuery
		}
		rhttp.Redirect(w, req, target, rhttp.StatusTemporaryRedirect)
	}))
}

func getJA3(req *http.Request) string {
	hash := md5.Sum([]byte(req.JA3Fingerprint))
	out := make([]byte, 32)
	hex.Encode(out, hash[:])
	return string(out)
}

// handler manages all path handling. It redirects the task of handling based on
// if the file exist, the file should be hosted (based on Path rules), and if
// the file should not be hosted
func (s Server) handler(w http.ResponseWriter, req *http.Request) {
	ja3 := getJA3(req)

	log.WithFields(log.Fields{
		"method":      req.Method,
		"host":        req.Host,
		"remote_addr": req.RemoteAddr,
		"req_uri":     req.RequestURI,
		"ja3":         ja3,
	}).Info("request")

	// Redirect to specified index
	if req.URL.Path == "/" && s.indexPath != "" {
		req.URL.Path = s.indexPath
	}

	if s.serverHeader != "" {
		w.Header().Add("Server", s.serverHeader)
	}

	served, err := s.paths.MatchAndServe(w, req)
	if err != nil {
		log.Error(err)
	}
	if !served {
		s.doesNotExistHandler(w, req)
	}
}

// doesNotExistHandler is 404
func (s Server) doesNotExistHandler(w http.ResponseWriter, req *http.Request) {
	if s.notFoundRedirect != "" {
		http.Redirect(w, req, s.notFoundRedirect, http.StatusMovedPermanently)
		return
	} else if s.notFoundRender != "" {
		if err := s.paths.Serve(w, req); err != nil {
			log.Error(err)
		}
		return
	}
	w.WriteHeader(http.StatusNotFound)
	io.WriteString(w, "404\n")
}
