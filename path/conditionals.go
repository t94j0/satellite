package path

import (
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"net"
	"os/exec"
	"regexp"
	"strings"

	"github.com/pkg/errors"
	"github.com/t94j0/satellite/net/http"
	"github.com/t94j0/satellite/net/http/httputil"
	"gopkg.in/yaml.v2"
)

// RequestConditions are the conditions in the http.Request object
type RequestConditions struct {
	// AUserAgent is the authorized user agents for a file
	AuthorizedUserAgents []string `yaml:"authorized_useragents,omitempty"`
	// BlacklistUserAgents are blacklisted user agents
	BlacklistUserAgents []string `yaml:"blacklist_useragents,omitempty"`
	// AuthorizedIPRange is the authorized range of IPs who are allowed to access a file
	AuthorizedIPRange []string `yaml:"authorized_iprange,omitempty"`
	// BlacklistIPRange are blacklisted IPs
	BlacklistIPRange []string `yaml:"blacklist_iprange,omitempty"`
	// AuthorizedMethods are the HTTP methods which can access the page
	AuthorizedMethods []string `yaml:"authorized_methods,omitempty"`
	// AuthorizedHeaders are HTTP headers which must be present in order to access a file
	AuthorizedHeaders map[string]string `yaml:"authorized_headers,omitempty"`
	// AuthorizedJA3 are valid JA3 hashes
	AuthorizedJA3 []string `yaml:"authorized_ja3,omitempty"`
	// Exec file executes script/binary and checks stdout
	Exec struct {
		ScriptPath string `yaml:"script"`
		Output     string `yaml:"output"`
	} `yaml:"exec,omitempty"`
	// NotServing does not serve the page when NotServing is true
	NotServing bool `yaml:"not_serving,omitempty"`
	// Serve is the number of times the file should be served
	Serve uint64 `yaml:"serve,omitempty"`
	// PrereqPaths path of hits that need to happen before the current one will succeed
	PrereqPaths []string `yaml:"prereq,omitempty"`
}

// NewRequestConditions creates an object based on a YAML blob
func NewRequestConditions(data []byte) (RequestConditions, error) {
	var conditions RequestConditions

	if err := yaml.Unmarshal(data, &conditions); err != nil {
		return conditions, err
	}

	for _, ua := range conditions.AuthorizedUserAgents {
		if _, err := regexp.Compile(ua); err != nil {
			return conditions, errors.New(fmt.Sprintf("%s is not valid regex", ua))
		}
	}

	for _, ua := range conditions.BlacklistUserAgents {
		if _, err := regexp.Compile(ua); err != nil {
			return conditions, errors.New(fmt.Sprintf("%s is not valid regex", ua))
		}
	}

	return conditions, nil
}

// ShouldHost returns when an HTTP request should be hosted or not
func (c *RequestConditions) ShouldHost(req *http.Request, state *State) bool {
	// Not Serving
	if c.NotServing {
		return false
	}

	// Agent
	correctAgent := false
	if len(c.AuthorizedUserAgents) != 0 {
		for _, u := range c.AuthorizedUserAgents {
			re := regexp.MustCompile(u)
			if re.MatchString(req.UserAgent()) {
				correctAgent = true
			}
		}
	} else {
		correctAgent = true
	}

	// Blacklist User Agents
	if len(c.BlacklistUserAgents) != 0 {
		for _, u := range c.BlacklistUserAgents {
			re := regexp.MustCompile(u)
			if re.MatchString(req.UserAgent()) {
				return false
			}
		}
	}

	// IP Range
	targetHost := net.ParseIP(req.RemoteAddr)
	correctRange := false
	if len(c.AuthorizedIPRange) != 0 {
		for _, r := range c.AuthorizedIPRange {
			if strings.Contains(r, "/") {
				_, tmpRange, err := net.ParseCIDR(r)
				if err != nil {
					return false
				}
				if tmpRange.Contains(targetHost) {
					correctRange = true
				}
			} else {
				if net.ParseIP(r).Equal(targetHost) {
					correctRange = true
				}
			}
		}
	} else {
		correctRange = true
	}

	// Blacklist IP range
	if len(c.BlacklistIPRange) != 0 {
		for _, r := range c.BlacklistIPRange {
			_, tmpRange, err := net.ParseCIDR(r)
			if err == nil {
				if tmpRange.Contains(targetHost) {
					return false
				}
			} else {
				if net.ParseIP(r).Equal(targetHost) {
					return false
				}
			}
		}
	}

	// Method
	correctMethods := false
	if len(c.AuthorizedMethods) != 0 {
		for _, m := range c.AuthorizedMethods {
			if req.Method == m {
				correctMethods = true
			}
		}
	} else {
		correctMethods = true
	}

	// Headers
	correctHeaders := false
	if len(c.AuthorizedHeaders) != 0 {
		for k, v := range c.AuthorizedHeaders {
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

	if len(c.AuthorizedJA3) != 0 {
		for _, j := range c.AuthorizedJA3 {
			if ja3 == j {
				correctJA3 = true
			}
		}
	} else {
		correctJA3 = true
	}

	// Exec
	correctExec := false
	if c.Exec.ScriptPath != "" {
		cmd := exec.Command(c.Exec.ScriptPath)

		stdin, err := cmd.StdinPipe()
		if err != nil {
			return false
		}

		go func() {
			defer stdin.Close()
			dump, err := httputil.DumpRequest(req, true)
			if err == nil {
				stdin.Write(dump)
			}
		}()

		out, err := cmd.CombinedOutput()
		if err != nil {
			return false
		}

		if c.Exec.Output == strings.TrimSuffix(string(out), "\n") {
			correctExec = true
		}
	} else {
		correctExec = true
	}

	// Serve
	correctServe := true
	if c.Serve != 0 && req.URL != nil {
		hits, err := state.GetHits(req.URL.Path)
		if err != nil {
			correctServe = false
		}
		if hits >= c.Serve {
			correctServe = false
		}
	}

	// Prereq Paths
	filledPrereq := true
	if len(c.PrereqPaths) != 0 {
		filledPrereq = state.MatchPaths(targetHost, c.PrereqPaths)
	}

	return correctAgent && correctRange && correctMethods && correctHeaders && correctJA3 && correctExec && correctServe && filledPrereq
}
