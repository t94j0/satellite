package handler

import (
	"encoding/base64"
	"encoding/json"
	"io"
	"io/ioutil"
	"path/filepath"

	"github.com/apcera/util/iprange"
	log "github.com/sirupsen/logrus"
	"github.com/t94j0/ja3-server/net/http"
	"github.com/t94j0/satellite/path"
	"github.com/t94j0/satellite/util"
	"gopkg.in/yaml.v2"
)

// Management is the interface for managing routes
type Management struct {
	paths      *path.Paths
	serverPath string
	ipRange    string
}

// NewManagementHandler creates the management routes
func NewManagementHandler(paths *path.Paths, serverPath, ipRange string) http.Handler {
	mgmt := Management{paths, serverPath, ipRange}
	mux := http.NewServeMux()
	mux.HandleFunc("/reset", mgmt.reset)
	mux.HandleFunc("/new", mgmt.upload)
	mux.HandleFunc("/", mgmt.get)
	return mux
}

// allowed determines if a request is allowed to access the management portal
func (m Management) allowed(req *http.Request) (bool, error) {
	mgmtRange, err := iprange.ParseIPRange(m.ipRange)
	if err != nil {
		return false, err
	}
	targetHost := util.GetHost(req)
	return mgmtRange.Contains(targetHost), nil
}

// get gets all paths in JSON format
func (m Management) get(w http.ResponseWriter, req *http.Request) {
	isAllowed, err := m.allowed(req)
	if err != nil || !isAllowed {
		m.doesNotExistHandler(w, req)
		return
	}

	outPaths := m.paths.Out()
	w.Header().Add("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(outPaths); err != nil {
		log.Println(err)
		m.doesNotExistHandler(w, req)
		return
	}
}

// reset resets a given path to 0 clicks and serves the page
func (m Management) reset(w http.ResponseWriter, req *http.Request) {
	isAllowed, err := m.allowed(req)
	if err != nil || !isAllowed {
		m.doesNotExistHandler(w, req)
		return
	}

	type Body struct {
		Path string `json:"path"`
	}
	var mgmt Body

	if req.Body == nil {
		log.Println("No HTTP body")
		m.doesNotExistHandler(w, req)
		return
	}
	if err := json.NewDecoder(req.Body).Decode(&mgmt); err != nil {
		log.Println(err)
		m.doesNotExistHandler(w, req)
		return
	}

	path, _ := m.paths.Match(mgmt.Path)

	path.NotServing = false
	path.TimesServed = 0

	if err := path.Write(); err != nil {
		log.Println(err)
		m.doesNotExistHandler(w, req)
		return
	}

}

// upload allows the user to upload a new path
func (m Management) upload(w http.ResponseWriter, req *http.Request) {
	isAllowed, err := m.allowed(req)
	if err != nil || !isAllowed {
		m.doesNotExistHandler(w, req)
		return
	}

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

	path := filepath.Join(m.serverPath, newPath.Path.Path)

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

// doesNotExistHandler is 404
func (m Management) doesNotExistHandler(w http.ResponseWriter, req *http.Request) {
	io.WriteString(w, "404\n")
}
