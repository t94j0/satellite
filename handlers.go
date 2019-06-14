package main

import (
	"encoding/json"
	"io"
	"io/ioutil"
	"log"
	"net"

	"github.com/apcera/util/iprange"
	"github.com/t94j0/ja3-server/crypto/tls"
	"github.com/t94j0/ja3-server/net/http"
)

// Server is used to serve HTTP(S)
type Server struct {
	port           string
	keyPath        string
	certPath       string
	serverHeader   string
	managementIP   string
	managementPath string
	identifier     *ClientID
}

// NewServer creates a new Server object
func NewServer(port, certPath, keyPath, serverHeader, managementIP, managementPath string) Server {
	return Server{
		port:           port,
		keyPath:        keyPath,
		certPath:       certPath,
		serverHeader:   serverHeader,
		managementIP:   managementIP,
		managementPath: managementPath,
		identifier:     NewClientID(),
	}
}

// Start makes the server begin listening
func (s Server) Start() error {
	mux := http.NewServeMux()
	if s.managementPath != "" {
		mux.HandleFunc(s.managementPath, s.managementHandler)
	}
	mux.HandleFunc("/", s.handler)
	//handler := http.HandlerFunc(s.handler)
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

// matchHandler is the handler for if the file exists and the rules matched the
// request. This means serve the target file
func (s Server) matchHandler(w http.ResponseWriter, req *http.Request, path *Path) {
	s.writeHeaders(w, path.AddHeadersSuccess)
	s.writeHeaders(w, path.ContentHeaders())
	data, err := ioutil.ReadFile(path.FullPath)
	if err != nil {
		log.Println(err)
		return
	}
	io.WriteString(w, string(data))
	if path.ShouldRemove() {
		path.Remove()
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

func (s Server) managementHandler(w http.ResponseWriter, req *http.Request) {
	mgmtRange, err := iprange.ParseIPRange(s.managementIP)
	if err != nil {
		log.Println(err)
	}
	targetHost := getHost(req)
	if !mgmtRange.Contains(targetHost) {
		s.doesNotExistHandler(w, req)
		return
	}

	if req.Method == "GET" {
		s.managementGetHandler(w, req)
	} else if req.Method == "POST" {
		s.managementPostHandler(w, req)
	}
}

func (s Server) managementGetHandler(w http.ResponseWriter, req *http.Request) {
	outPaths := paths.Out()
	w.Header().Add("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(outPaths); err != nil {
		log.Println(err)
		s.doesNotExistHandler(w, req)
		return
	}
}

func (s Server) managementPostHandler(w http.ResponseWriter, req *http.Request) {
	type Management struct {
		ID    uint `json:"id"`
		Reset bool `json:"reset"`
	}
	var mgmt Management

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

	path := paths.GetID(mgmt.ID)
	if mgmt.Reset {
		path.NotServing = false
		path.TimesServed = 0
		if err := path.Write(); err != nil {
			log.Println(err)
			s.doesNotExistHandler(w, req)
			return
		}
	}
}
