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
	port     string
	keyPath  string
	certPath string
}

func NewServer(port, certPath, keyPath string) Server {
	return Server{
		port:     port,
		keyPath:  keyPath,
		certPath: certPath,
	}
}

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

func (s Server) writeHeaders(w http.ResponseWriter, headers map[string]string) {
	for name, value := range headers {
		w.Header().Add(name, value)
	}
}

func (s Server) matchHandler(w http.ResponseWriter, req *http.Request, path Path) {

	s.writeHeaders(w, path.AddHeadersSuccess)
	data, err := ioutil.ReadFile(path.FullPath)
	if err != nil {
		log.Println(err)
		return
	}
	io.WriteString(w, string(data))
	if path.Once {
		paths.Remove(path)
	}
	return
}

func (s Server) noMatchHandler(w http.ResponseWriter, req *http.Request, path Path) {
	s.writeHeaders(w, path.AddHeadersFailure)
	io.WriteString(w, "Errored\n")
}

func (s Server) doesNotExistHandler(w http.ResponseWriter, req *http.Request, path Path) {
	s.writeHeaders(w, path.AddHeadersNotExist)
	// TODO: Write a 404 handler
	io.WriteString(w, "404\n")
}

func (s Server) handler(w http.ResponseWriter, req *http.Request) {
	path, exists := paths.Match(req.URL.Path)

	s.writeHeaders(w, path.AddHeaders)
	if !exists {
		s.doesNotExistHandler(w, req, path)
	} else if file.ShouldHost(req) {
		s.matchHandler(w, req, path)
	} else {
		s.noMatchHandler(w, req, path)
	}
}
