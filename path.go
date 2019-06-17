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
	"github.com/t94j0/ja3-server/net/http/httputil"
	"gopkg.in/yaml.v2"
)

// Path is an available path that can be accessed on the server
type Path struct {
	Path string `yaml:"-"`
	// FullPath is the path of the file to host
	FullPath string `yaml:"-"`
	// NotServing does not serve the page when NotServing is true
	NotServing bool `yaml:"not_serving,omitempty"`
	// AddHeaders are a dict of headers to add to every request
	AddHeaders map[string]string `yaml:"add_headers,omitempty"`
	// AddHeadersSuccess are a dict of headers to add to every successful request
	AddHeadersSuccess map[string]string `yaml:"add_headers_success,omitempty"`
	// AddHeadersFailure are a dict of headers to add to every hit, but failed header
	AddHeadersFailure map[string]string `yaml:"add_headers_failure,omitempty"`
	// AUserAgent is the authorized user agents for a file
	AuthorizedUserAgents []string `yaml:"authorized_useragents,omitempty"`
	// AuthorizedIPRange is the authorized range of IPs who are allowed to access a file
	AuthorizedIPRange []string `yaml:"authorized_iprange,omitempty"`
	// AuthorizedMethods are the HTTP methods which can access the page
	AuthorizedMethods []string `yaml:"authorized_methods,omitempty"`
	// AuthorizedHeaders are HTTP headers which must be present in order to access a file
	AuthorizedHeaders map[string]string `yaml:"authorized_headers,omitempty"`
	// AuthorizedJA3 are valid JA3 hashes
	AuthorizedJA3 []string `yaml:"authorized_ja3,omitempty"`
	// BlacklistIPRange are blacklisted IPs
	BlacklistIPRange []string `yaml:"blacklist_iprange,omitempty"`
	// Serve is the number of times the file should be served
	Serve uint `yaml:"serve,omitempty"`
	// TimesServed is the number of times served so far
	TimesServed uint `yaml:"times_served,omitempty"`
	// ContentType tells the browser what content should be parsed. A list of MIME
	// types can be found here: https://www.freeformatter.com/mime-types-list.html
	ContentType string `yaml:"content_type,omitempty"`
	// Disposition sets the Content-Disposition header
	Disposition struct {
		// Type is the type of disposition. Usually either inline or attachment
		Type string `yaml:"type"`
		// FileName is the name of the file if Content.Type is attachment
		FileName string `yaml:"file_name"`
	} `yaml:"disposition,omitempty"`
	// PrereqPaths path of hits that need to happen before the current one will succeed
	PrereqPaths []string `yaml:"prereq,omitempty"`
	Exec        struct {
		ScriptPath string `yaml:"script"`
		Output     string `yaml:"output"`
	} `yaml:"exec,omitempty"`
	// OnFailure instructs the Path what to do when a failure occurs
	OnFailure struct {
		// Redirect will redirect the user with a 301 to a target address
		Redirect string `yaml:"redirect"`
		// Render will render the following path
		Render string `yaml:"render"`
	} `yaml:"on_failure"`
}

// NewPath parses a .info file in the base path directory
func NewPath(infoPath string) (*Path, error) {
	var newInfo Path

	data, err := ioutil.ReadFile(infoPath)
	if err != nil {
		return &infoPath, err
	}

	if err := yaml.Unmarshal(data, &newInfo); err != nil {
		return &newInfo, err
	}

	return &newInfo, nil
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

func getHost(req *http.Request) net.IP {
	inp := req.RemoteAddr
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
	return f.TimesServed >= f.Serve
}

// Remove removes the ability to access the file
func (f *Path) Remove() error {
	f.NotServing = true
	return f.Write()
}

// Write Info path to file
func (f *Path) Write() error {
	out, err := yaml.Marshal(f)
	if err != nil {
		return err
	}
	return ioutil.WriteFile(f.FullPath+".info", out, 0644)
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
	targetHost := getHost(req)
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
		cmd := exec.Command(f.Exec.ScriptPath)

		stdin, err := cmd.StdinPipe()
		if err != nil {
			return false, err
		}

		go func() {
			defer stdin.Close()
			dump, _ := httputil.DumpRequest(req, true)
			stdin.Write(dump)
		}()

		out, err := cmd.CombinedOutput()
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
	if len(f.PrereqPaths) != 0 {
		filledPrereq = identifier.Match(targetHost, f.PrereqPaths)
	}

	didSucceed := correctAgent && correctRange && correctMethods && correctHeaders && correctJA3 && filledPrereq && correctExec

	if didSucceed {
		f.TimesServed++
		identifier.Hit(targetHost, f.Path)
		if err := f.Write(); err != nil {
			return false, err
		}
	}

	return didSucceed, nil
}
