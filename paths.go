package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/apcera/util/iprange"
)

// Paths is the compilation of parsed paths
type Paths struct {
	base string
	list map[string]*Path
}

// NewPaths creates a new Paths variable from the specified base path
func NewPaths(base string) (*Paths, error) {
	list := make(map[string]*Path)
	ret := &Paths{
		base: base,
		list: list,
	}
	if err := ret.Reload(); err != nil {
		return ret, err
	}
	if err := ret.createCompleted(); err != nil {
		return ret, err
	}
	return ret, nil
}

// Add adds a new Path to the global paths list
func (paths *Paths) Add(id string, path *Path) {
	paths.list[id] = path
}

// Out returns the list as a non-pointer map
func (paths *Paths) Out() map[string]Path {
	retPaths := make(map[string]Path, 0)
	for k, path := range paths.list {
		retPaths[k] = *path
	}
	return retPaths
}

// Match matches a page given a URI. It returns the specified Path and a boolean
// value to determine if there was a page that matched the URI
func (paths *Paths) Match(URI string) (*Path, bool) {
	p, b := paths.list[URI]
	return p, b
}

// Len returns the number of successfully parsed paths
func (paths *Paths) Len() int {
	return len(paths.list)
}

func (paths *Paths) createCompleted() error {
	name := filepath.Join(paths.base, "done")
	if _, err := os.Stat(name); !os.IsNotExist(err) {
		// The directory already exists. Nothing to be done
		return nil
	}

	return os.Mkdir(name, os.FileMode(0777))
}

// Remove removes the specified path from the list of paths
func (paths *Paths) Remove(path *Path) {
	// Remove path from list
	for k, p := range paths.list {
		if p.FullPath == path.FullPath {
			delete(paths.list, k)
			break
		}
	}
	// TODO: Put removed Path into the `done` directory
}

// ErrPath is errors describing the path
type ErrPath struct {
	path string
	err  string
}

func (i *ErrPath) Error() string {
	return fmt.Sprintf("%s %s", i.path, i.err)
}

func (paths *Paths) verify() error {
	for _, path := range paths.list {
		// Check Authorized IPs
		for _, r := range path.AuthorizedIPRange {
			if _, err := iprange.ParseIPRange(r); err != nil {
				return &ErrPath{path.Path, "has an invalid IP range: " + r}
			}
		}

		// Check path of file exists
		if _, err := os.Stat(path.FullPath); os.IsNotExist(err) {
			return &ErrPath{path.Path, "does not have a source file associated"}
		}
	}
	return nil
}

// Reload refreshes the list of paths internally to Paths
func (paths *Paths) Reload() error {
	paths.list = make(map[string]*Path)
	if err := filepath.Walk(paths.base, func(oPath string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		// Only parse .info files
		if !strings.HasSuffix(oPath, ".info") {
			return nil
		}

		tmpPath, err := NewPath(oPath)
		if err != nil {
			return err
		}

		fullPath := strings.TrimSuffix(oPath, ".info")
		requestPath := strings.TrimPrefix(fullPath, paths.base)
		tmpPath.FullPath = fullPath
		tmpPath.Path = requestPath

		paths.list[requestPath] = tmpPath
		return nil
	}); err != nil {
		return err
	}

	return paths.verify()
}
