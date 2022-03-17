package handlers

import (
	"crypto/md5"
	"encoding/hex"
	"io"
	"net"
	"strings"

	log "github.com/sirupsen/logrus"
	"github.com/t94j0/satellite/net/http"
	"github.com/t94j0/satellite/satellite/geoip"
	"github.com/t94j0/satellite/satellite/path"
	"github.com/t94j0/satellite/satellite/util"
)

// RootHandler is the Handler function for all incoming http requests
type RootHandler struct {
	defaultIndex string
	serverHeader string
	paths        *path.Paths
	notFound     util.NotFound
}

// NewRootHandler creates a new RootHandler object
func NewRootHandler(ps *path.Paths, notFound util.NotFound, defaultIndex, serverHeader string) RootHandler {
	return RootHandler{
		defaultIndex: defaultIndex,
		serverHeader: serverHeader,
		paths:        ps,
		notFound:     notFound,
	}
}

// ServeHTTP redirects the task of handling based on
// if the file exist, the file should be hosted (based on Path rules), and if
// the file should not be hosted
func (h RootHandler) ServeHTTP(w http.ResponseWriter, req *http.Request) {

	// Redirect to specified index
	if req.URL.Path == "/" && h.defaultIndex != "" {
		req.URL.Path = h.defaultIndex
	}

	if h.serverHeader != "" {
		w.Header().Add("Server", h.serverHeader)
	}

	served, err := h.paths.MatchAndServe(w, req)
	if err != nil {
		log.Error(err)
	}
	if !served {
		log.Debug("File not found. Redirecting to not_found")
		h.log(req, 301)
		h.notExistHandler(w, req)
	} else {
		h.log(req, 200)
	}
}

func (h RootHandler) notExistHandler(w http.ResponseWriter, req *http.Request) {
	if h.notFound.Redirect != "" {
		http.Redirect(w, req, h.notFound.Redirect, http.StatusMovedPermanently)
	} else if h.notFound.Render != "" {
		req.URL.Path = h.notFound.Render
		if err := h.paths.Serve(w, req); err != nil {
			log.Error(err)
		}
	} else {
		w.WriteHeader(http.StatusNotFound)
		io.WriteString(w, "404\n")
	}
}

func getJA3(req *http.Request) string {
	hash := md5.Sum([]byte(req.JA3Fingerprint))
	out := make([]byte, 32)
	hex.Encode(out, hash[:])
	return string(out)
}

func parseRemoteAddr(ipPort string) net.IP {
	targetIP := strings.Split(ipPort, ":")[0]
	return net.ParseIP(targetIP)
}

func getCountryCode(remoteAddr string, gip *geoip.DB) (string, error) {
	targetHost := parseRemoteAddr(remoteAddr)
	if gip.HasDB() {
		cc, err := gip.CountryCode(targetHost)
		if err != nil {
			return "", err
		}
		return cc, nil
	}
	return "", nil
}

func (h RootHandler) log(req *http.Request, respCode int) {
	ja3 := getJA3(req)
	cc, err := getCountryCode(req.RemoteAddr, &h.paths.GeoipDB)
	if err != nil {
		log.Error(err)
	}
	log.WithFields(log.Fields{
		"method":      req.Method,
		"host":        req.Host,
		"remote_addr": req.RemoteAddr,
		"req_uri":     req.RequestURI,
		"ja3":         ja3,
		"response":    respCode,
		"user_agent":  req.UserAgent(),
		"geo_ip":      cc,
	}).Info("request")
}
