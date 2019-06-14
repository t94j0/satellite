package main

import (
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"io/ioutil"
	"net"
	"os/exec"
	"strings"

	"github.com/apcera/util/iprange"
	"github.com/t94j0/ja3-server/net/http"
	"gopkg.in/yaml.v2"
)

// Path is an available path that can be accessed on the server
type Path struct {
	// FullPath is the path of the file to host
	FullPath string
	// ID of path
	ID uint
	// AddHeaders are a dict of headers to add to every request
	AddHeaders map[string]string `yaml:"add_headers"`
	// AddHeadersSuccess are a dict of headers to add to every successful request
	AddHeadersSuccess map[string]string `yaml:"add_headers_success"`
	// AddHeadersFailure are a dict of headers to add to every hit, but failed header
	AddHeadersFailure map[string]string `yaml:"add_headers_failure"`
	// AUserAgent is the authorized user agents for a file
	AuthorizedUserAgents []string `yaml:"authorized_useragents"`
	// AuthorizedIPRange is the authorized range of IPs who are allowed to access a file
	AuthorizedIPRange []string `yaml:"authorized_iprange"`
	// AuthorizedMethods are the HTTP methods which can access the page
	AuthorizedMethods []string `yaml:"authorized_methods"`
	// AuthorizedHeaders are HTTP headers which must be present in order to access a file
	AuthorizedHeaders map[string]string `yaml:"authorized_headers"`
	// AuthorizedJA3 are valid JA3 hashes
	AuthorizedJA3 []string `yaml:"authorized_ja3"`
	// BlacklistIPRange are blacklisted IPs
	BlacklistIPRange []string `yaml:"blacklist_iprange"`
	// Serve is the number of times the file should be served
	Serve uint `yaml:"serve"`
	// timesServed is the number of times served so far
	timesServed uint
	// ContentType tells the browser what content should be parsed. A list of MIME
	// types can be found here: https://www.freeformatter.com/mime-types-list.html
	ContentType string `yaml:"content_type"`
	// Disposition sets the Content-Disposition header
	Disposition struct {
		// Type is the type of disposition. Usually either inline or attachment
		Type string `yaml:"type"`
		// FileName is the name of the file if Content.Type is attachment
		FileName string `yaml:"file_name"`
	} `yaml:"disposition"`
	// IDs of hits that need to happen before the current one will succeed
	PrereqIDs []uint `yaml:"prereq"`
	Exec      struct {
		ScriptPath string `yaml:"script"`
		Output     string `yaml:"output"`
	} `yaml:"exec"`
}

// NewPath parses a .info file in the base path directory
func NewPath(path string) (*Path, error) {
	var infoPath Path

	data, err := ioutil.ReadFile(path)
	if err != nil {
		return &infoPath, err
	}

	if err := yaml.Unmarshal(data, &infoPath); err != nil {
		return &infoPath, err
	}

	return &infoPath, nil
}

// ContentHeaders sets the Content-Type and Content-Disposition headers.
func (f *Path) ContentHeaders() map[string]string {
	headers := make(map[string]string)

	if f.ContentType != "" {
		headers["Content-Type"] = f.ContentType
	}

	if f.Disposition.Type != "" {
		if f.Disposition.FileName != "" {
			headers["Content-Disposition"] = fmt.Sprintf("%s; filename=\"%s\"", f.Disposition.Type, f.Disposition.FileName)
		} else {
			headers["Content-Disposition"] = f.Disposition.Type
		}
	}

	return headers
}

func getHost(inp string) net.IP {
	split := strings.Split(inp, ":")
	filter := split[:len(split)-1]
	full := strings.Join(filter, ":")
	trimmedr := strings.TrimRight(full, "]")
	trimmed := strings.TrimLeft(trimmedr, "[")
	return net.ParseIP(trimmed)
}

// ShouldRemove determines if a path should be removed because the number of
// times served has been reached
func (f *Path) ShouldRemove() bool {
	if f.Serve == 0 {
		return false
	}
	return f.timesServed >= f.Serve
}

// ShouldHost does the checking to see if the requested file should be given to a target
// TODO: Make this function less ass
func (f *Path) ShouldHost(req *http.Request, identifier *ClientID) (bool, error) {
	// Don't serve if it's been served too many times
	if f.ShouldRemove() {
		return false, nil
	}

	// Agent
	correctAgent := false
	if len(f.AuthorizedUserAgents) != 0 {
		for _, u := range f.AuthorizedUserAgents {
			if req.UserAgent() == u {
				correctAgent = true
			}
		}
	} else {
		correctAgent = true
	}

	// IP Range
	targetHost := getHost(req.RemoteAddr)
	correctRange := false
	if len(f.AuthorizedIPRange) != 0 {
		for _, r := range f.AuthorizedIPRange {
			tmpRange, err := iprange.ParseIPRange(r)
			if err != nil {
				return false, err
			}
			if tmpRange.Contains(targetHost) {
				correctRange = true
			}
		}
	} else {
		correctRange = true
	}

	// Blacklist IP range
	if len(f.BlacklistIPRange) != 0 {
		for _, r := range f.BlacklistIPRange {
			tmpRange, err := iprange.ParseIPRange(r)
			if err != nil {
				return false, err
			}
			if tmpRange.Contains(targetHost) {
				return false, nil
			}
		}
	}

	// Method
	correctMethods := false
	if len(f.AuthorizedMethods) != 0 {
		for _, m := range f.AuthorizedMethods {
			if req.Method == m {
				correctMethods = true
			}
		}
	} else {
		correctMethods = true
	}

	// Headers
	correctHeaders := false
	if len(f.AuthorizedHeaders) != 0 {
		for k, v := range f.AuthorizedHeaders {
			if req.Header.Get(k) == v {
				correctHeaders = true
			}
		}
	} else {
		correctHeaders = true
	}

	// JA3
	hash := md5.Sum([]byte(req.JA3Fingerprint))
	out := make([]byte, 32)
	hex.Encode(out, hash[:])
	ja3 := string(out)

	correctJA3 := false

	if len(f.AuthorizedJA3) != 0 {
		for _, j := range f.AuthorizedJA3 {
			if ja3 == j {
				correctJA3 = true
			}
		}
	} else {
		correctJA3 = true
	}

	// Exec
	correctExec := false
	if f.Exec.ScriptPath != "" {
		out, err := exec.Command(f.Exec.ScriptPath).Output()
		if err != nil {
			return false, err
		}
		if f.Exec.Output == strings.TrimSuffix(string(out), "\n") {
			correctExec = true
		}
	} else {
		correctExec = true
	}

	// Prereq
	filledPrereq := true
	if len(f.PrereqIDs) != 0 {
		filledPrereq = identifier.Match(targetHost, f.PrereqIDs)
	}

	didSucceed := correctAgent && correctRange && correctMethods && correctHeaders && correctJA3 && filledPrereq && correctExec

	if didSucceed {
		f.timesServed += 1
		identifier.Hit(targetHost, f.ID)
	}

	return didSucceed, nil
}
