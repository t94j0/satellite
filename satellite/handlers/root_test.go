package handlers_test

import (
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
	"testing"

	"github.com/t94j0/satellite/net/http/httptest"
	. "github.com/t94j0/satellite/satellite/handlers"
	"github.com/t94j0/satellite/satellite/path"
	"github.com/t94j0/satellite/satellite/util"
)

var RedirectNotFound = util.NotFound{Redirect: "/index.html", Render: ""}
var RenderNotFound = util.NotFound{Redirect: "", Render: "/index.html"}
var NoNotFound = util.NotFound{Redirect: "", Render: ""}

// TempDir Helper
type TempDir struct {
	Path string
}

func NewTempDir() (TempDir, error) {
	var td TempDir
	dir, err := ioutil.TempDir("", "")
	if err != nil {
		return td, nil
	}
	td.Path = dir
	return td, nil
}

func (t TempDir) CreateFile(name, content string) error {
	fullpath := filepath.Join(t.Path, name)
	if err := ioutil.WriteFile(fullpath, []byte(content), 0666); err != nil {
		return err
	}
	return nil
}

func (t TempDir) CreateFiles(paths map[string]string) error {
	for k, v := range paths {
		if err := t.CreateFile(k, v); err != nil {
			return err
		}
	}
	return nil
}

func (t TempDir) Paths() (*path.Paths, error) {
	return path.NewDefaultTest(t.Path)
}

func (t TempDir) Close() error {
	return os.RemoveAll(t.Path)
}

// Tests
func TestRootHandler_ServeHTTP_exact(t *testing.T) {
	td, err := NewTempDir()
	if err != nil {
		t.Error(err)
	}
	defer td.Close()
	td.CreateFiles(map[string]string{
		"/index.html": "Hello!",
	})
	paths, err := td.Paths()
	if err != nil {
		t.Error(err)
	}
	handler := NewRootHandler(paths, RedirectNotFound, "/index.html", "Server")

	req := httptest.NewRequest("GET", "/index.html", nil)
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	resp := w.Result()
	body, _ := ioutil.ReadAll(resp.Body)

	if resp.StatusCode != http.StatusOK || string(body) != "Hello!" {
		t.Fail()
	}
}

func TestRootHandler_ServeHTTP_redirect(t *testing.T) {
	td, err := NewTempDir()
	if err != nil {
		t.Error(err)
	}
	defer td.Close()
	td.CreateFiles(map[string]string{
		"/index.html": "Hello!",
	})
	paths, err := td.Paths()
	if err != nil {
		t.Error(err)
	}
	handler := NewRootHandler(paths, RedirectNotFound, "/index.html", "Server")

	req := httptest.NewRequest("GET", "/", nil)
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	resp := w.Result()
	body, _ := ioutil.ReadAll(resp.Body)

	if resp.StatusCode != http.StatusOK || string(body) != "Hello!" {
		t.Fail()
	}
}

func TestRootHandler_ServeHTTP_notfound_redirect(t *testing.T) {
	td, err := NewTempDir()
	if err != nil {
		t.Error(err)
	}
	defer td.Close()
	td.CreateFiles(map[string]string{})
	paths, err := td.Paths()
	if err != nil {
		t.Error(err)
	}

	handler := NewRootHandler(paths, RedirectNotFound, "/index.html", "Server")

	req := httptest.NewRequest("GET", "/", nil)
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	resp := w.Result()

	if resp.StatusCode != http.StatusMovedPermanently {
		t.Fail()
	}

	if resp.Header.Get("Location") != "/index.html" {
		t.Fail()
	}
}

func TestRootHandler_ServeHTTP_notfound_render(t *testing.T) {
	td, err := NewTempDir()
	if err != nil {
		t.Error(err)
	}
	defer td.Close()
	td.CreateFiles(map[string]string{
		"index.html": "not found",
	})
	paths, err := td.Paths()
	if err != nil {
		t.Error(err)
	}

	handler := NewRootHandler(paths, RenderNotFound, "/abc123", "Server")

	req := httptest.NewRequest("GET", "/", nil)
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	resp := w.Result()
	body, _ := ioutil.ReadAll(resp.Body)

	if resp.StatusCode != 200 && string(body) != "not found" {
		t.Fail()
	}
}

func TestRootHandler_ServeHTTP_notfound_404(t *testing.T) {
	td, err := NewTempDir()
	if err != nil {
		t.Error(err)
	}
	defer td.Close()
	paths, err := td.Paths()
	if err != nil {
		t.Error(err)
	}

	handler := NewRootHandler(paths, NoNotFound, "/index.html", "Server")

	req := httptest.NewRequest("GET", "/abc", nil)
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	resp := w.Result()
	body, _ := ioutil.ReadAll(resp.Body)

	if resp.StatusCode != http.StatusNotFound && string(body) != "404" {
		t.Fail()
	}
}
