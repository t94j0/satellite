package main

import (
	"crypto/md5"
	"encoding/hex"
	"io/ioutil"
	"log"
	"net"
	"strings"

	"github.com/apcera/util/iprange"
	"github.com/t94j0/ja3-server/net/http"
	"gopkg.in/yaml.v2"
)

type Path struct {
	// FullPath is the path of the file to host
	FullPath string
	// AddHeaders are a dict of headers to add to every request
	AddHeaders map[string]string `yaml:"add_headers"`
	// AddHeadersSuccess are a dict of headers to add to every successful request
	AddHeadersSuccess map[string]string `yaml:"add_headers_success"`
	// AddHeadersFailure are a dict of headers to add to every hit, but failed header
	AddHeadersFailure map[string]string `yaml:"add_headers_failure"`
	// AddHeadersNotExist are a dict of headers to add to every 404
	AddHeadersNotExist map[string]string `yaml:"add_headers_not_exist"`
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
	// Once files will only be served once before they are no longer able to be accessed
	Once bool `yaml:"once"`
}

func NewPath(path string) (Path, error) {
	var infoPath Path

	data, err := ioutil.ReadFile(path)
	if err != nil {
		return infoPath, err
	}

	if err := yaml.Unmarshal(data, &infoPath); err != nil {
		return infoPath, err
	}

	return infoPath, nil
}

func getHost(inp string) net.IP {
	host := strings.Split(inp, ":")[0]
	return net.ParseIP(host)
}

func (f Path) ShouldHost(req *http.Request) bool {
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

	targetHost := getHost(req.RemoteAddr)
	correctRange := false
	if len(f.AuthorizedIPRange) != 0 {
		for _, r := range f.AuthorizedIPRange {
			tmpRange, err := iprange.ParseIPRange(r)
			// TODO: Validate IP ranges earlier
			if err != nil {
				log.Fatal(err)
			}
			if tmpRange.Contains(targetHost) {
				correctRange = true
			}
		}
	} else {
		correctRange = true
	}

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

	// Check JA3
	hash := md5.Sum([]byte(req.JA3Fingerprint))
	out := make([]byte, 32)
	hex.Encode(out, hash[:])
	ja3 := string(out)
	log.Println("JA3:", ja3)

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

	return correctAgent && correctRange && correctMethods && correctHeaders && correctJA3
}
