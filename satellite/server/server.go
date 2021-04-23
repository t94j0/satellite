package server

import (
	"net"
	rhttp "net/http"

	"github.com/t94j0/satellite/crypto/tls"
	"github.com/t94j0/satellite/net/http"
	"github.com/t94j0/satellite/satellite/handlers"
	"github.com/t94j0/satellite/satellite/path"
	"github.com/t94j0/satellite/satellite/util"
)

// Server is used to serve HTTP(S)
type Server struct {
	paths        *path.Paths
	ssl          SSL
	nf           util.NotFound
	serverPath   string
	port         string
	serverHeader string
	indexPath    string
	redirectHTTP bool
	identifier   *path.ClientID
}

// New creates a new Server object
func New(paths *path.Paths, ssl SSL, nf util.NotFound, serverPath, port, serverHeader, indexPath string, redirectHTTP bool) (Server, error) {
	return Server{
		paths:        paths,
		serverPath:   serverPath,
		port:         port,
		ssl:          ssl,
		nf:           nf,
		serverHeader: serverHeader,
		indexPath:    indexPath,

		redirectHTTP: redirectHTTP,
		identifier:   path.NewClientID(),
	}, nil
}

// Start makes the server begin listening
func (s Server) Start() error {
	if s.redirectHTTP {
		go func() {
			s.createHTTPRedirect()
		}()
	}

	rootHandler := handlers.NewRootHandler(s.paths, s.nf, s.indexPath, s.serverHeader)

	mux := http.NewServeMux()
	mux.Handle("/", http.Handler(rootHandler))

	return s.serveHTTPS(mux)
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

// serveHTTPS serves the mux with HTTPS
func (s Server) serveHTTPS(mux *http.ServeMux) error {
	server := &http.Server{Addr: s.port, Handler: mux}
	ln, err := net.Listen("tcp", s.port)
	if err != nil {
		return err
	}
	defer ln.Close()

	tlsConfig, err := s.ssl.CreateTLSConfig()
	if err != nil {
		return err
	}

	tlsListener := tls.NewListener(ln, tlsConfig)
	return server.Serve(tlsListener)
}
