package main

import (
	"os"
	"path/filepath"
	"strings"
)

type Paths struct {
	base string
	list map[string]Path
}

func NewPaths(base string) (*Paths, error) {
	list := make(map[string]Path)
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

func (paths *Paths) Match(page string) (Path, bool) {
	p, b := paths.list[page]
	return p, b
}

func (paths *Paths) createCompleted() error {
	name := filepath.Join(paths.base, "done")
	if _, err := os.Stat(name); !os.IsNotExist(err) {
		// The directory already exists. Nothing to be done
		return nil
	}

	return os.Mkdir(name, os.FileMode(0777))
}

func (paths *Paths) Remove(path Path) {
	// Remove path from list
	for k, p := range paths.list {
		if p.FullPath == path.FullPath {
			delete(paths.list, k)
			break
		}
	}
}

func (paths *Paths) Reload() error {
	err := filepath.Walk(paths.base, func(oPath string, info os.FileInfo, err error) error {
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

		fullPath := strings.TrimRight(oPath, ".info")
		requestPath := "/" + strings.TrimLeft(fullPath, paths.base)
		tmpPath.FullPath = fullPath

		paths.list[requestPath] = tmpPath
		return nil
	})

	if err != nil {
		return err
	}

	return nil
}
