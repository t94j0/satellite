package path

import (
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/pkg/errors"
	"github.com/t94j0/satellite/net/http"
)

// Paths is the compilation of parsed paths
type Paths struct {
	base      string
	list      map[string]*Path
	proxyPath string
	state     *State
}

// New creates a new Paths variable from the specified base path
func New(base string) (*Paths, error) {
	list := make(map[string]*Path)

	state, err := NewState(path.Join(base, ".db"))
	if err != nil {
		return nil, err
	}

	ret := &Paths{
		base:      base,
		list:      list,
		proxyPath: path.Join(base, "/.proxy.yml"),
		state:     state,
	}

	if err := ret.Reload(); err != nil {
		return ret, err
	}

	return ret, nil
}

// Len gets the number of paths
func (paths *Paths) Len() int {
	return len(paths.list)
}

// AddProxyList is a flat list of proxies to add in YAML format
func (paths *Paths) AddProxyList(path string) error {
	pathArr, err := NewPathArray(path)
	if err != nil {
		return err
	}

	for _, p := range pathArr {
		paths.list[p.Path] = p
	}

	return nil
}

// Add adds a new Path to the global paths list
func (paths *Paths) Add(path string, pathData *Path) {
	paths.list[path] = pathData
}

// Match matches a page given a URI. It returns the specified Path and a boolean
// value to determine if there was a page that matched the URI
func (paths *Paths) Match(URI string) (*Path, bool) {
	p, b := paths.list[URI]
	return p, b
}

// Remove removes a path from the list of usable paths
func (paths *Paths) Remove(path string) {
	// Remove path from list in memory
	delete(paths.list, path)
	paths.state.Remove(path)
}

func (paths *Paths) RemoveDir(dir string) {
	for k := range paths.list {
		if strings.HasPrefix(k, dir) {
			paths.Remove(k)
		}
	}
}

// Reload refreshes the list of paths internally to Paths
func (paths *Paths) Reload() error {
	paths.list = make(map[string]*Path)
	if err := filepath.Walk(paths.base, func(oPath string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if strings.HasSuffix(oPath, ".info") {
			tmpPath, err := NewPath(oPath)
			if err != nil {
				return err
			}

			fullPath := strings.TrimSuffix(oPath, ".info")
			tmpPath.HostedFile = fullPath

			requestPath := strings.TrimPrefix(fullPath, paths.base)
			tmpPath.Path = requestPath

			paths.list[requestPath] = tmpPath
		} else {
			var tmpPath Path
			tmpPath.HostedFile = oPath

			requestPath := strings.TrimPrefix(oPath, paths.base)
			tmpPath.Path = requestPath

			paths.list[requestPath] = &tmpPath
		}
		return nil
	}); err != nil {
		return err
	}

	// Add proxy path if it exists
	proxyPath := paths.proxyPath
	if _, err := os.Stat(proxyPath); !os.IsNotExist(err) {
		if err := paths.AddProxyList(proxyPath); err != nil {
			return err
		}
	}

	paths.Remove("")
	paths.RemoveDir("/.db")
	paths.Remove("/.proxy.yml")

	return nil
}

// Serve serves a page without checking conditionals
func (paths *Paths) Serve(w http.ResponseWriter, req *http.Request) error {
	writeHeaders := func(w http.ResponseWriter, headers map[string]string) {
		for name, value := range headers {
			w.Header().Add(name, value)
		}
	}
	uri := req.URL.Path
	targetPath, exists := paths.Match(uri)
	if !exists {
		return errors.New("not_found render page not found")
	}

	writeHeaders(w, targetPath.ContentHeaders())
	if err := targetPath.ServeHTTP(w, req); err != nil {
		return err
	}
	return nil
}

// MatchAndServe matches a path, determines if the path should be served, and serves the file based on an HTTP request. If a failure occurs, this function will serve failed pages.
//
// This is a helper function which combines already-exposed functions to make file serving easy.
//
// Returns true when the file was served and false when a 404 page should be returned
func (paths *Paths) MatchAndServe(w http.ResponseWriter, req *http.Request) (bool, error) {
	writeHeaders := func(w http.ResponseWriter, headers map[string]string) {
		for name, value := range headers {
			w.Header().Add(name, value)
		}
	}

	uri := req.URL.Path
	targetPath, exists := paths.Match(uri)
	if !exists {
		return false, nil
	}

	shouldHost := targetPath.ShouldHost(req, paths.state)

	if !shouldHost {
		if targetPath.OnFailure.Redirect != "" {
			http.Redirect(w, req, targetPath.OnFailure.Redirect, http.StatusMovedPermanently)
			return true, nil
		} else if targetPath.OnFailure.Render != "" {
			newPath, found := paths.Match(targetPath.OnFailure.Render)
			if !found {
				return false, errors.New("path not found")
			}
			if err := newPath.ServeHTTP(w, req); err != nil {
				return false, err
			}
			return true, nil
		}
		return false, nil
	}

	writeHeaders(w, targetPath.ContentHeaders())
	if err := targetPath.ServeHTTP(w, req); err != nil {
		return false, err
	}
	return true, nil
}
