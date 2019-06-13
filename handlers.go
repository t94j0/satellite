package main

import (
	"io"
	"io/ioutil"
	"log"
	"net"

	"github.com/CapacitorSet/ja3-server/crypto/tls"
	"github.com/CapacitorSet/ja3-server/net/http"
)

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

func (s Server) matchHandler(w http.ResponseWriter, req *http.Request, path Path) {
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

func (s Server) noMatchHandler(w http.ResponseWriter, req *http.Request) {
	io.WriteString(w, "Errored\n")
}

func (s Server) doesNotExistHandler(w http.ResponseWriter, req *http.Request) {
	// TODO: Write a 404 handler
	io.WriteString(w, "404\n")
	return
}

func (s Server) handler(w http.ResponseWriter, req *http.Request) {
	file, exists := paths.Match(req.URL.Path)

	if !exists {
		s.doesNotExistHandler(w, req)
	} else if file.ShouldHost(req) {
		s.matchHandler(w, req, file)
	} else {
		s.noMatchHandler(w, req)
	}
}
