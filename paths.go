package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/apcera/util/iprange"
	"github.com/t94j0/array"
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

type ErrID struct {
	id  uint
	err string
}

func (i *ErrID) Error() string {
	return fmt.Sprintf("ID %d %s", i.id, i.err)
}

type ErrIPRange struct {
	id      uint
	ipRange string
}

func (i *ErrIPRange) Error() string {
	return fmt.Sprintf("IP range for ID %d has an invalid IP range of %s", i.id, i.ipRange)
}

type ErrPathNotExist struct {
	id uint
}

func (i *ErrPathNotExist) Error() string {
	return fmt.Sprintf("File for ID %d does not exist", i.id)
}

type ErrIDNotExist struct {
	path string
}

func (i *ErrIDNotExist) Error() string {
	return fmt.Sprintf("No ID exists for file %s", i.path)
}

func (paths *Paths) verify() error {
	checkedIDs := make([]uint, 0)
	for _, path := range paths.list {
		// Check ID
		if path.ID == 0 {
			return &ErrIDNotExist{path.FullPath}
		}
		// Check ID parameters
		if array.In(path.ID, checkedIDs) {
			return &ErrID{path.ID, "used multiple times"}
		}
		checkedIDs = append(checkedIDs, path.ID)

		// Check Authorized IPs
		for _, r := range path.AuthorizedIPRange {
			if _, err := iprange.ParseIPRange(r); err != nil {
				return &ErrIPRange{path.ID, r}
			}
		}

		// Check path of file exists
		if _, err := os.Stat(path.FullPath); os.IsNotExist(err) {
			return &ErrPathNotExist{path.ID}
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

		paths.list[requestPath] = tmpPath
		return nil
	}); err != nil {
		return err
	}

	return paths.verify()
}
