package main

import (
	"io"
	"io/ioutil"
	"log"
	"net"

	"github.com/t94j0/ja3-server/crypto/tls"
	"github.com/t94j0/ja3-server/net/http"
)

// Server is used to serve HTTP(S)
type Server struct {
	port       string
	keyPath    string
	certPath   string
	identifier *ClientID
}

// NewServer creates a new Server object
func NewServer(port, certPath, keyPath string) Server {
	return Server{
		port:       port,
		keyPath:    keyPath,
		certPath:   certPath,
		identifier: NewClientID(),
	}
}

// Start makes the server begin listening
func (s Server) Start() error {
	handler := http.HandlerFunc(s.handler)
	server := &http.Server{Addr: s.port, Handler: handler}

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
	err = server.Serve(tlsListener)
	if err != nil {
		return err
	}

	ln.Close()

	return nil
}

// handler manages all path handling. It redirects the task of handling based on
// if the file exist, the file should be hosted (based on Path rules), and if
// the file should not be hosted
func (s Server) handler(w http.ResponseWriter, req *http.Request) {
	path, exists := paths.Match(req.URL.Path)

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

// matchHandler is the handler for if the file exists and the rules matched the
// request. This means serve the target file
func (s Server) matchHandler(w http.ResponseWriter, req *http.Request, path *Path) {
	s.writeHeaders(w, path.AddHeadersSuccess)
	if path.Download {
		w.Header().Add("Content-Type", "application/octet-stream")
	} else if path.ContentType != "" {
		w.Header().Add("Content-Type", path.ContentType)
	}
	data, err := ioutil.ReadFile(path.FullPath)
	if err != nil {
		log.Println(err)
		return
	}
	io.WriteString(w, string(data))
	if path.ShouldRemove() {
		paths.Remove(path)
	}
	return
}

// noMatchHandler is the handler used when the file exists, but the rules
// determine the request does not match
func (s Server) noMatchHandler(w http.ResponseWriter, req *http.Request, path *Path) {
	// TODO: Write a failure handler
	s.writeHeaders(w, path.AddHeadersFailure)
	io.WriteString(w, "Errored\n")
}

// doesNotExistHandler is 404
func (s Server) doesNotExistHandler(w http.ResponseWriter, req *http.Request) {
	// TODO: Write a 404 handler
	io.WriteString(w, "404\n")
}
