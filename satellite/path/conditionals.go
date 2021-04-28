package path

import (
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"net"
	"os/exec"
	"regexp"
	"strings"

	"github.com/imdario/mergo"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	"github.com/t94j0/satellite/net/http"
	"github.com/t94j0/satellite/net/http/httputil"
	"github.com/t94j0/satellite/satellite/geoip"
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
	GeoIP       struct {
		AuthorizedCountries []string `yaml:"authorized_countries"`
		BlacklistCountries  []string `yaml:"blacklist_countries"`
	} `yaml:"geoip"`
}

// NewRequestConditions creates an object based on a YAML blob
func NewRequestConditions(data []byte) (RequestConditions, error) {
	var conditions RequestConditions

	if err := yaml.Unmarshal(data, &conditions); err != nil {
		return conditions, err
	}

	regexes := append(conditions.AuthorizedUserAgents, conditions.BlacklistUserAgents...)

	for _, ua := range regexes {
		if _, err := regexp.Compile(ua); err != nil {
			return conditions, errors.New(fmt.Sprintf("%s is not valid regex", ua))
		}
	}

	return conditions, nil
}

// MergeRequestConditions merges a list of RequestCondition. They are applied starting from the first to the last. It will overwrite later RequestCondition
func MergeRequestConditions(conds ...RequestConditions) (RequestConditions, error) {
	var target RequestConditions
	for _, c := range conds {
		if err := mergo.Merge(&target, c, mergo.WithOverride, mergo.WithAppendSlice); err != nil {
			return target, err
		}
	}
	return target, nil
}

func parseRemoteAddr(ipPort string) net.IP {
	targetIP := strings.Split(ipPort, ":")[0]
	return net.ParseIP(targetIP)
}

func (c *RequestConditions) authorizedUserAgents(req *http.Request) bool {
	correctAgent := false
	userAgent := req.UserAgent()

	if len(c.AuthorizedUserAgents) == 0 {
		log.Trace("No Authorized User Agents")
		return true
	}

	for _, u := range c.AuthorizedUserAgents {
		re := regexp.MustCompile(u)
		if re.MatchString(userAgent) {
			log.WithFields(log.Fields{
				"user_agent":        userAgent,
				"target_user_agent": u,
			}).Debug("Matched User Agent")
			correctAgent = true
		} else {
			log.WithFields(log.Fields{
				"user_agent":        userAgent,
				"target_user_agent": u,
			}).Trace("Did not match authorized User Agent")
		}
	}

	return correctAgent
}

func (c *RequestConditions) blacklistUserAgents(req *http.Request) bool {
	if len(c.BlacklistUserAgents) == 0 {
		log.Trace("No Blacklist User Agents")
		return true
	}

	for _, u := range c.BlacklistUserAgents {
		re := regexp.MustCompile(u)
		if re.MatchString(req.UserAgent()) {
			log.WithFields(log.Fields{
				"user_agent":        req.UserAgent(),
				"target_user_agent": u,
			}).Debug("Blacklisted User Agent")
			return false
		}
		log.WithFields(log.Fields{
			"user_agent":        req.UserAgent(),
			"target_user_agent": u,
		}).Trace("Did not match blacklisted User Agent")
	}

	return true
}

func (c *RequestConditions) authorizedIPRange(req *http.Request) bool {
	targetHost := parseRemoteAddr(req.RemoteAddr)
	correctRange := false

	if len(c.AuthorizedIPRange) == 0 {
		log.Trace("No authorized IP ranges")
		return true
	}

	for _, r := range c.AuthorizedIPRange {
		if strings.Contains(r, "/") {
			_, tmpRange, err := net.ParseCIDR(r)
			if err != nil {
				log.WithFields(log.Fields{
					"ip": r,
				}).Debug("Could not parse IP range")
				return false
			}
			if tmpRange.Contains(targetHost) {
				log.WithFields(log.Fields{
					"ip": r,
				}).Debug("Matched authorized IP range")
				correctRange = true
			} else {
				log.WithFields(log.Fields{
					"ip": r,
				}).Trace("Did not match authorized IP range")
			}
		} else {
			if net.ParseIP(r).Equal(targetHost) {
				log.WithFields(log.Fields{
					"ip": r,
				}).Debug("Matched authorized IP range")
				correctRange = true
			} else {
				log.WithFields(log.Fields{
					"ip": r,
				}).Trace("Did not match authorized IP range")
			}
		}
	}

	return correctRange
}

func (c *RequestConditions) blacklistIPRange(req *http.Request) bool {
	targetHost := parseRemoteAddr(req.RemoteAddr)

	if len(c.BlacklistIPRange) == 0 {
		return true
	}

	for _, r := range c.BlacklistIPRange {
		_, tmpRange, err := net.ParseCIDR(r)
		if err == nil {
			if tmpRange.Contains(targetHost) {
				log.WithFields(log.Fields{
					"ip": r,
				}).Debug("Matched blacklisted IP range")
				return false
			}
			log.WithFields(log.Fields{
				"ip": r,
			}).Trace("Did not match blacklisted IP range")
		} else {
			if net.ParseIP(r).Equal(targetHost) {
				log.WithFields(log.Fields{
					"ip": r,
				}).Debug("Matched blacklisted IP range")
				return false
			}
			log.WithFields(log.Fields{
				"ip": r,
			}).Trace("Did not match blacklisted IP range")
		}
	}
	return true
}

func (c *RequestConditions) authorizedMethods(req *http.Request) bool {
	correctMethods := false

	if len(c.AuthorizedMethods) == 0 {
		log.Trace("No authorized methods")
		return true
	}

	for _, m := range c.AuthorizedMethods {
		if req.Method == m {
			log.WithFields(log.Fields{
				"method": m,
			}).Debug("Matched HTTP method")
			correctMethods = true
		}
		log.WithFields(log.Fields{
			"method": m,
		}).Trace("Did not match HTTP method")
	}

	return correctMethods
}

func (c *RequestConditions) authorizedHeaders(req *http.Request) bool {
	correctHeaders := false

	if len(c.AuthorizedHeaders) == 0 {
		log.Trace("No authorized methods")
		return true
	}

	for k, v := range c.AuthorizedHeaders {
		if req.Header.Get(k) == v {
			log.WithFields(log.Fields{
				"header_key":   k,
				"header_value": v,
			}).Debug("Matched header")
			correctHeaders = true
		}
		log.WithFields(log.Fields{
			"header_key":   k,
			"header_value": v,
		}).Trace("Did not match header")
	}

	return correctHeaders
}

func (c *RequestConditions) authorizedJA3(req *http.Request) bool {
	hash := md5.Sum([]byte(req.JA3Fingerprint))
	out := make([]byte, 32)
	hex.Encode(out, hash[:])
	ja3 := string(out)

	correctJA3 := false

	if len(c.AuthorizedJA3) == 0 {
		log.Trace("No authorized JA3 signatures")
		return true
	}

	for _, j := range c.AuthorizedJA3 {
		if ja3 == j {
			log.WithFields(log.Fields{
				"target_ja3": j,
				"req_ja3":    ja3,
			}).Debug("Authorized JA3 signature matched")
			correctJA3 = true
		} else {
			log.WithFields(log.Fields{
				"target_ja3": j,
				"req_ja3":    ja3,
			}).Trace("Authorized JA3 signature did not match")
		}
	}

	return correctJA3
}

func (c *RequestConditions) authorizedExec(req *http.Request) bool {
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
	return correctExec
}

func (c *RequestConditions) serveLimit(req *http.Request, state *State) bool {
	correctServe := true
	if c.Serve != 0 && req.URL != nil {
		hits, err := state.GetHits(req.URL.Path)
		if err != nil {
			log.WithFields(log.Fields{
				"error": err,
			}).Debug("Error getting times served")
			correctServe = false
		}
		if hits >= c.Serve {
			log.WithFields(log.Fields{
				"serve_limit":  c.Serve,
				"times_served": hits,
			}).Debug("Route exceeds times served")
			correctServe = false
		} else {
			log.WithFields(log.Fields{
				"serve_limit":  c.Serve,
				"times_served": hits,
			}).Trace("Route served")
		}
	}
	return correctServe
}

func (c *RequestConditions) prereqMatch(req *http.Request, state *State) bool {
	filledPrereq := true
	if len(c.PrereqPaths) == 0 {
		return true
	}

	targetHost := parseRemoteAddr(req.RemoteAddr)
	filledPrereq = state.MatchPaths(targetHost, c.PrereqPaths)
	if filledPrereq {
		log.WithFields(log.Fields{
			"prereqs": c.PrereqPaths,
		}).Debug("Matched prerequisites")
	} else {
		log.WithFields(log.Fields{
			"prereqs": c.PrereqPaths,
		}).Debug("Did not match prerequisites")
	}

	return filledPrereq
}

func (c *RequestConditions) geoipMatch(req *http.Request, gip geoip.DB) bool {
	targetHost := parseRemoteAddr(req.RemoteAddr)
	correctGeoIP := true
	if gip.HasDB() {
		cc, err := gip.CountryCode(targetHost)
		if err != nil {
			log.WithFields(log.Fields{
				"error": err,
			}).Debug("Error getting country code")
			return false
		}

		// Authorized GeoIP
		if len(c.GeoIP.AuthorizedCountries) != 0 {
			correctGeoIP = false
			for _, targetCC := range c.GeoIP.AuthorizedCountries {
				if cc == targetCC {
					log.WithFields(log.Fields{
						"target_countrycode": targetCC,
						"countrycode":        cc,
					}).Debug("Matched authorized country code")
					correctGeoIP = true
				} else {
					log.WithFields(log.Fields{
						"target_countrycode": targetCC,
						"countrycode":        cc,
					}).Trace("Did not match authorized country code")
				}
			}
		}

		// Blacklist GeoIP
		if len(c.GeoIP.BlacklistCountries) != 0 {
			for _, targetCC := range c.GeoIP.BlacklistCountries {
				if targetCC == cc {
					log.WithFields(log.Fields{
						"target_countrycode": targetCC,
						"countrycode":        cc,
					}).Debug("Matched blacklist country code")
					return false
				}
				log.WithFields(log.Fields{
					"target_countrycode": targetCC,
					"countrycode":        cc,
				}).Trace("Did not match blacklist country code")
			}
		}
	}
	return correctGeoIP
}

// ShouldHost returns when an HTTP request should be hosted or not
func (c *RequestConditions) ShouldHost(req *http.Request, state *State, gip geoip.DB) bool {
	// Not Serving
	if c.NotServing {
		log.Trace("Not serving")
		return false
	}

	if ok := c.authorizedUserAgents(req); !ok {
		return false
	}

	if ok := c.blacklistUserAgents(req); !ok {
		return false
	}

	if ok := c.authorizedIPRange(req); !ok {
		return false
	}

	if ok := c.blacklistIPRange(req); !ok {
		return false
	}

	if ok := c.authorizedMethods(req); !ok {
		return false
	}

	if ok := c.authorizedHeaders(req); !ok {
		return false
	}

	if ok := c.authorizedJA3(req); !ok {
		return false
	}

	if ok := c.authorizedExec(req); !ok {
		return false
	}

	if ok := c.serveLimit(req, state); !ok {
		return false
	}

	// Prereq Paths
	if ok := c.prereqMatch(req, state); !ok {
		return false
	}

	if ok := c.geoipMatch(req, gip); !ok {
		return false
	}

	return true
}
