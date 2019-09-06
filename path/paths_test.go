package path_test

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	"github.com/t94j0/satellite/net/http/httptest"
	. "github.com/t94j0/satellite/path"
)

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

func (t TempDir) Close() error {
	return os.RemoveAll(t.Path)
}

// Tests
func TestNew_none(t *testing.T) {
	dir, err := NewTempDir()
	if err != nil {
		t.Error(err)
	}

	if _, err := New(dir.Path); err != nil {
		t.Error(err)
	}
}

func TestPaths_Len_zero(t *testing.T) {
	tmpdir, err := NewTempDir()
	if err != nil {
		t.Error(err)
	}
	paths, err := New(tmpdir.Path)
	if err != nil {
		t.Error(err)
	}
	if paths.Len() != 0 {
		t.Fail()
	}
	if err := tmpdir.Close(); err != nil {
		t.Error(err)
	}
}

func TestPaths_Len_one(t *testing.T) {
	tmpdir, err := NewTempDir()
	if err != nil {
		t.Error(err)
	}
	if err := tmpdir.CreateFile("one.info", ""); err != nil {
		t.Error(err)
	}
	if err := tmpdir.CreateFile("one", ""); err != nil {
		t.Error(err)
	}
	paths, err := New(tmpdir.Path)
	if err != nil {
		t.Error(err)
	}
	if paths.Len() != 1 {
		t.Fail()
	}
	if err := tmpdir.Close(); err != nil {
		t.Error(err)
	}
}

func TestNew_one(t *testing.T) {
	tmpdir, err := NewTempDir()
	if err != nil {
		t.Error(err)
	}
	if err := tmpdir.CreateFile("one.info", ""); err != nil {
		t.Error(err)
	}
	if _, err := New(tmpdir.Path); err != nil {
		t.Error(err)
	}
	if err := tmpdir.Close(); err != nil {
		t.Error(err)
	}
}

func TestNew_proxy(t *testing.T) {
	tmpdir, err := NewTempDir()
	if err != nil {
		t.Error(err)
	}
	proxyContent := `
- path: /
  proxy: http://google.com`
	if err := tmpdir.CreateFile(".proxy.yml", proxyContent); err != nil {
		t.Error(err)
	}
	if _, err := New(tmpdir.Path); err != nil {
		t.Error(err)
	}
	if err := tmpdir.Close(); err != nil {
		t.Error(err)
	}
}

func TestPaths_new_testproxyyml(t *testing.T) {
	tmpdir, err := NewTempDir()
	if err != nil {
		t.Error(err)
	}
	oneData := `
- path: /`
	if err := tmpdir.CreateFile(".proxy.yml", oneData); err != nil {
		t.Error(err)
	}
	paths, err := New(tmpdir.Path)
	if err != nil {
		t.Error(err)
	}
	if _, ok := paths.Match("/"); !ok {
		t.Fail()
	}
	if err := tmpdir.Close(); err != nil {
		t.Error(err)
	}
}

func TestPaths_Add(t *testing.T) {
	tmpdir, err := NewTempDir()
	if err != nil {
		t.Error(err)
	}
	paths, err := New(tmpdir.Path)
	if err != nil {
		t.Error(err)
	}
	pathData := `
authorized_useragents:
  - none`
	newPath, err := NewPathData([]byte(pathData))
	if err != nil {
		t.Error(err)
	}
	paths.Add("/", newPath)
	paths.Remove("/")
	if err := tmpdir.Close(); err != nil {
		t.Error(err)
	}
}

func TestPaths_MatchAndServe_file_success(t *testing.T) {
	// Create project directory
	tmpdir, err := NewTempDir()
	if err != nil {
		t.Error(err)
	}
	if err := tmpdir.CreateFile("one", "Hello!"); err != nil {
		t.Error(err)
	}
	if err := tmpdir.CreateFile("one.info", ""); err != nil {
		t.Error(err)
	}

	// Create paths object
	paths, err := New(tmpdir.Path)
	if err != nil {
		t.Error(err)
	}

	// Create HTTP request
	req := httptest.NewRequest("GET", "/one", nil)
	w := httptest.NewRecorder()

	// Execute request
	didMatch, err := paths.MatchAndServe(w, req)
	if err != nil {
		t.Error(err)
	}
	if didMatch == false || w.Code != 200 || w.Body.String() != "Hello!" {
		t.Fail()
	}

	// Close
	if err := tmpdir.Close(); err != nil {
		t.Error(err)
	}
}

func TestPaths_MatchAndServe_file_success_headers(t *testing.T) {
	// Create project directory
	tmpdir, err := NewTempDir()
	if err != nil {
		t.Error(err)
	}
	if err := tmpdir.CreateFile("one", "Hello!"); err != nil {
		t.Error(err)
	}
	oneData := `
content_type: application/json
`
	if err := tmpdir.CreateFile("one.info", oneData); err != nil {
		t.Error(err)
	}

	// Create paths object
	paths, err := New(tmpdir.Path)
	if err != nil {
		t.Error(err)
	}

	// Create HTTP request
	req := httptest.NewRequest("GET", "/one", nil)
	w := httptest.NewRecorder()

	// Execute request
	didMatch, err := paths.MatchAndServe(w, req)
	if err != nil {
		t.Error(err)
	}
	if didMatch == false || w.Code != 200 || w.Body.String() != "Hello!" {
		t.Fail()
	}

	// Close
	if err := tmpdir.Close(); err != nil {
		t.Error(err)
	}
}

func TestPaths_MatchAndServe_file_failure_redirect(t *testing.T) {
	// Create project directory
	tmpdir, err := NewTempDir()
	if err != nil {
		t.Error(err)
	}
	if err := tmpdir.CreateFile("one", "Hello!"); err != nil {
		t.Error(err)
	}
	oneData := `
authorized_useragents:
  - none
on_failure:
  redirect: https://aws.amazon.com`
	if err := tmpdir.CreateFile("one.info", oneData); err != nil {
		t.Error(err)
	}

	// Create paths object
	paths, err := New(tmpdir.Path)
	if err != nil {
		t.Error(err)
	}

	// Create HTTP request
	req := httptest.NewRequest("GET", "/one", nil)
	req.Header.Add("User-Agent", "wont_match")
	w := httptest.NewRecorder()

	// Execute request
	didMatch, err := paths.MatchAndServe(w, req)
	if err != nil {
		t.Error(err)
	}
	if didMatch == false || w.Code != 301 {
		t.Fail()
	}

	// Close
	if err := tmpdir.Close(); err != nil {
		t.Error(err)
	}
}

func TestPaths_MatchAndServe_file_failure_render(t *testing.T) {
	// Create project directory
	tmpdir, err := NewTempDir()
	if err != nil {
		t.Error(err)
	}
	if err := tmpdir.CreateFile("one", "Hello!"); err != nil {
		t.Error(err)
	}
	oneData := `
authorized_useragents:
  - none
on_failure:
  render: /index.html`
	if err := tmpdir.CreateFile("one.info", oneData); err != nil {
		t.Error(err)
	}
	if err := tmpdir.CreateFile("index.html.info", ""); err != nil {
		t.Error(err)
	}
	if err := tmpdir.CreateFile("index.html", "Hello!"); err != nil {
		t.Error(err)
	}

	// Create paths object
	paths, err := New(tmpdir.Path)
	if err != nil {
		t.Error(err)
	}

	// Create HTTP request
	req := httptest.NewRequest("GET", "/one", nil)
	req.Header.Add("User-Agent", "wont_match")
	w := httptest.NewRecorder()

	// Execute request
	didMatch, err := paths.MatchAndServe(w, req)
	if err != nil {
		t.Error(err)
	}
	if didMatch == false || w.Code != 200 || w.Body.String() != "Hello!" {
		t.Fail()
	}

	// Close
	if err := tmpdir.Close(); err != nil {
		t.Error(err)
	}
}

func TestPaths_MatchAndServe_notfound(t *testing.T) {
	// Create project directory
	tmpdir, err := NewTempDir()
	if err != nil {
		t.Error(err)
	}

	// Create paths object
	paths, err := New(tmpdir.Path)
	if err != nil {
		t.Error(err)
	}

	// Create HTTP request
	req := httptest.NewRequest("GET", "/one", nil)
	req.Header.Add("User-Agent", "wont_match")
	w := httptest.NewRecorder()

	// Execute request
	didMatch, err := paths.MatchAndServe(w, req)
	if err != nil {
		t.Error(err)
	}
	if didMatch == true {
		t.Fail()
	}

	// Close
	if err := tmpdir.Close(); err != nil {
		t.Error(err)
	}
}

func TestPaths_MatchAndServe_file_failure_render_notfound(t *testing.T) {
	// Create project directory
	tmpdir, err := NewTempDir()
	if err != nil {
		t.Error(err)
	}
	if err := tmpdir.CreateFile("one", "Hello!"); err != nil {
		t.Error(err)
	}
	oneData := `
authorized_useragents:
  - none
on_failure:
  render: /index.html`
	if err := tmpdir.CreateFile("one.info", oneData); err != nil {
		t.Error(err)
	}

	// Create paths object
	paths, err := New(tmpdir.Path)
	if err != nil {
		t.Error(err)
	}

	// Create HTTP request
	req := httptest.NewRequest("GET", "/one", nil)
	req.Header.Add("User-Agent", "wont_match")
	w := httptest.NewRecorder()

	// Execute request
	if _, err := paths.MatchAndServe(w, req); err.Error() != "path not found" {
		t.Fail()
	}

	// Close
	if err := tmpdir.Close(); err != nil {
		t.Error(err)
	}
}

func TestPaths_MatchAndServe_file_failure_render_meme(t *testing.T) {
	// Create project directory
	tmpdir, err := NewTempDir()
	if err != nil {
		t.Error(err)
	}
	if err := tmpdir.CreateFile("one", "Hello!"); err != nil {
		t.Error(err)
	}
	oneData := `
authorized_useragents:
  - none`
	if err := tmpdir.CreateFile("one.info", oneData); err != nil {
		t.Error(err)
	}

	// Create paths object
	paths, err := New(tmpdir.Path)
	if err != nil {
		t.Error(err)
	}

	// Create HTTP request
	req := httptest.NewRequest("GET", "/one", nil)
	req.Header.Add("User-Agent", "wont_match")
	w := httptest.NewRecorder()

	// Execute request
	didHost, err := paths.MatchAndServe(w, req)
	if err != nil {
		t.Error(err)
	}
	if didHost {
		t.Fail()
	}

	// Close
	if err := tmpdir.Close(); err != nil {
		t.Error(err)
	}
}

func TestPaths_Serve_success(t *testing.T) {
	// Create project directory
	tmpdir, err := NewTempDir()
	if err != nil {
		t.Error(err)
	}
	if err := tmpdir.CreateFile("one", "Hello!"); err != nil {
		t.Error(err)
	}
	if err := tmpdir.CreateFile("one.info", ""); err != nil {
		t.Error(err)
	}

	// Create paths object
	paths, err := New(tmpdir.Path)
	if err != nil {
		t.Error(err)
	}

	// Create HTTP request
	req := httptest.NewRequest("GET", "/one", nil)
	w := httptest.NewRecorder()

	// Execute request
	if err := paths.Serve(w, req); err != nil {
		t.Error(err)
	}

	if w.Code != 200 || w.Body.String() != "Hello!" {
		t.Fail()
	}

	// Close
	if err := tmpdir.Close(); err != nil {
		t.Error(err)
	}
}

func TestPaths_Serve_notfound(t *testing.T) {
	// Create project directory
	tmpdir, err := NewTempDir()
	if err != nil {
		t.Error(err)
	}

	// Create paths object
	paths, err := New(tmpdir.Path)
	if err != nil {
		t.Error(err)
	}

	// Create HTTP request
	req := httptest.NewRequest("GET", "/one", nil)
	w := httptest.NewRecorder()

	// Execute request
	if err := paths.Serve(w, req); err.Error() != "not_found render page not found" {
		t.Fail()
	}

	// Close
	if err := tmpdir.Close(); err != nil {
		t.Error(err)
	}
}

func TestPaths_Serve_success_headers(t *testing.T) {
	// Create project directory
	tmpdir, err := NewTempDir()
	if err != nil {
		t.Error(err)
	}
	if err := tmpdir.CreateFile("one", "Hello!"); err != nil {
		t.Error(err)
	}
	oneData := `
content_type: application/json
`
	if err := tmpdir.CreateFile("one.info", oneData); err != nil {
		t.Error(err)
	}

	// Create paths object
	paths, err := New(tmpdir.Path)
	if err != nil {
		t.Error(err)
	}

	// Create HTTP request
	req := httptest.NewRequest("GET", "/one", nil)
	w := httptest.NewRecorder()

	// Execute request
	if err := paths.Serve(w, req); err != nil {
		t.Error(err)
	}

	contentHeader := w.Header().Get("Content-Type")
	if w.Code != 200 || w.Body.String() != "Hello!" || contentHeader != "application/json" {
		t.Fail()
	}

	// Close
	if err := tmpdir.Close(); err != nil {
		t.Error(err)
	}
}

func TestPaths_MatchAndServe_file_noinfo(t *testing.T) {
	// Create project directory
	tmpdir, err := NewTempDir()
	if err != nil {
		t.Error(err)
	}
	if err := tmpdir.CreateFile("one", "Hello!"); err != nil {
		t.Error(err)
	}

	// Create paths object
	paths, err := New(tmpdir.Path)
	if err != nil {
		t.Error(err)
	}

	// Create HTTP request
	req := httptest.NewRequest("GET", "/one", nil)
	w := httptest.NewRecorder()

	// Execute request
	didMatch, err := paths.MatchAndServe(w, req)
	if err != nil {
		t.Error(err)
	}
	if didMatch == false || w.Code != 200 || w.Body.String() != "Hello!" {
		t.Fail()
	}

	// Close
	if err := tmpdir.Close(); err != nil {
		t.Error(err)
	}
}
