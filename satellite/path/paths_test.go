package path_test

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	"github.com/t94j0/satellite/net/http/httptest"
	. "github.com/t94j0/satellite/satellite/path"
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

func (t TempDir) CreateFile(name, content string) {
	fullpath := filepath.Join(t.Path, name)
	ioutil.WriteFile(fullpath, []byte(content), 0666)
}

func (t TempDir) Close() error {
	return os.RemoveAll(t.Path)
}

func makeUABlacklist(userAgents ...string) string {
	tgt := "blacklist_useragents:\n"
	for _, u := range userAgents {
		tgt += fmt.Sprintf("  - %s\n", u)
	}
	return tgt
}

// Tests
func TestNew_none(t *testing.T) {
	dir, err := NewTempDir()
	if err != nil {
		t.Error(err)
	}

	if _, err := NewDefaultTest(dir.Path); err != nil {
		t.Error(err)
	}
}

func TestPaths_Len_zero(t *testing.T) {
	tmpdir, err := NewTempDir()
	if err != nil {
		t.Error(err)
	}
	defer tmpdir.Close()

	paths, err := NewDefaultTest(tmpdir.Path)
	if err != nil {
		t.Error(err)
	}

	if paths.Len() != 0 {
		t.Fail()
	}
}

func TestPaths_Len_one(t *testing.T) {
	tmpdir, err := NewTempDir()
	if err != nil {
		t.Error(err)
	}
	defer tmpdir.Close()
	tmpdir.CreateFile("one.info", "")
	tmpdir.CreateFile("one", "")

	paths, err := NewDefaultTest(tmpdir.Path)
	if err != nil {
		t.Error(err)
	}
	if paths.Len() != 1 {
		t.Fail()
	}
}

func TestNew_one(t *testing.T) {
	tmpdir, err := NewTempDir()
	if err != nil {
		t.Error(err)
	}
	defer tmpdir.Close()
	tmpdir.CreateFile("one.info", "")
	if _, err := NewDefaultTest(tmpdir.Path); err != nil {
		t.Error(err)
	}
}

func TestNew_proxy(t *testing.T) {
	tmpdir, err := NewTempDir()
	if err != nil {
		t.Error(err)
	}
	defer tmpdir.Close()
	proxyContent := `
- path: /
  proxy: http://google.com`
	tmpdir.CreateFile(".proxy.yml", proxyContent)
	if _, err := NewDefaultTest(tmpdir.Path); err != nil {
		t.Error(err)
	}
}

func TestPaths_new_testproxyyml(t *testing.T) {
	tmpdir, err := NewTempDir()
	if err != nil {
		t.Error(err)
	}
	defer tmpdir.Close()
	oneData := `
- path: /`
	tmpdir.CreateFile(".proxy.yml", oneData)
	paths, err := NewDefaultTest(tmpdir.Path)
	if err != nil {
		t.Error(err)
	}
	if _, ok := paths.Match("/"); !ok {
		t.Fail()
	}
}

func TestPaths_Add(t *testing.T) {
	tmpdir, err := NewTempDir()
	if err != nil {
		t.Error(err)
	}
	defer tmpdir.Close()
	paths, err := NewDefaultTest(tmpdir.Path)
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
}

func TestPaths_MatchAndServe_file_success(t *testing.T) {
	// Create project directory
	tmpdir, err := NewTempDir()
	if err != nil {
		t.Error(err)
	}
	defer tmpdir.Close()
	tmpdir.CreateFile("one", "Hello!")
	tmpdir.CreateFile("one.info", "")

	// Create paths object
	paths, err := NewDefaultTest(tmpdir.Path)
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
}

func TestPaths_MatchAndServe_file_success_headers(t *testing.T) {
	// Create project directory
	tmpdir, err := NewTempDir()
	if err != nil {
		t.Error(err)
	}
	defer tmpdir.Close()
	tmpdir.CreateFile("one", "Hello!")
	oneData := `
content_type: application/json
`
	tmpdir.CreateFile("one.info", oneData)

	// Create paths object
	paths, err := NewDefaultTest(tmpdir.Path)
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

}

func TestPaths_MatchAndServe_file_failure_redirect(t *testing.T) {
	// Create project directory
	tmpdir, err := NewTempDir()
	if err != nil {
		t.Error(err)
	}
	defer tmpdir.Close()
	tmpdir.CreateFile("one", "Hello!")
	oneData := `
authorized_useragents:
  - none
on_failure:
  redirect: https://aws.amazon.com`
	tmpdir.CreateFile("one.info", oneData)

	// Create paths object
	paths, err := NewDefaultTest(tmpdir.Path)
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
}

func TestPaths_MatchAndServe_file_failure_render(t *testing.T) {
	// Create project directory
	tmpdir, err := NewTempDir()
	if err != nil {
		t.Error(err)
	}
	defer tmpdir.Close()
	tmpdir.CreateFile("one", "Hello!")
	oneData := `
authorized_useragents:
  - none
on_failure:
  render: /index.html`
	tmpdir.CreateFile("one.info", oneData)
	tmpdir.CreateFile("index.html.info", "")
	tmpdir.CreateFile("index.html", "Hello!")

	// Create paths object
	paths, err := NewDefaultTest(tmpdir.Path)
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
}

func TestPaths_MatchAndServe_notfound(t *testing.T) {
	// Create project directory
	tmpdir, err := NewTempDir()
	if err != nil {
		t.Error(err)
	}
	defer tmpdir.Close()

	// Create paths object
	paths, err := NewDefaultTest(tmpdir.Path)
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
}

func TestPaths_MatchAndServe_file_failure_render_notfound(t *testing.T) {
	// Create project directory
	tmpdir, err := NewTempDir()
	if err != nil {
		t.Error(err)
	}
	defer tmpdir.Close()
	tmpdir.CreateFile("one", "Hello!")
	oneData := `
authorized_useragents:
  - none
on_failure:
  render: /index.html`
	tmpdir.CreateFile("one.info", oneData)

	// Create paths object
	paths, err := NewDefaultTest(tmpdir.Path)
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
}

func TestPaths_MatchAndServe_file_failure_render_meme(t *testing.T) {
	// Create project directory
	tmpdir, err := NewTempDir()
	if err != nil {
		t.Error(err)
	}
	defer tmpdir.Close()
	tmpdir.CreateFile("one", "Hello!")
	oneData := `
authorized_useragents:
  - none`
	tmpdir.CreateFile("one.info", oneData)
	// Create paths object
	paths, err := NewDefaultTest(tmpdir.Path)
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
}

func TestPaths_Serve_success(t *testing.T) {
	// Create project directory
	tmpdir, err := NewTempDir()
	if err != nil {
		t.Error(err)
	}
	defer tmpdir.Close()
	tmpdir.CreateFile("one", "Hello!")
	tmpdir.CreateFile("one.info", "")

	// Create paths object
	paths, err := NewDefaultTest(tmpdir.Path)
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
}

func TestPaths_Serve_notfound(t *testing.T) {
	// Create project directory
	tmpdir, err := NewTempDir()
	if err != nil {
		t.Error(err)
	}
	defer tmpdir.Close()

	// Create paths object
	paths, err := NewDefaultTest(tmpdir.Path)
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
}

func TestPaths_Serve_success_headers(t *testing.T) {
	// Create project directory
	tmpdir, err := NewTempDir()
	if err != nil {
		t.Error(err)
	}
	defer tmpdir.Close()
	tmpdir.CreateFile("one", "Hello!")
	oneData := `
content_type: application/json
`
	tmpdir.CreateFile("one.info", oneData)

	// Create paths object
	paths, err := NewDefaultTest(tmpdir.Path)
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
}

func TestPaths_MatchAndServe_file_noinfo(t *testing.T) {
	// Create project directory
	tmpdir, err := NewTempDir()
	if err != nil {
		t.Error(err)
	}
	defer tmpdir.Close()
	tmpdir.CreateFile("one", "Hello!")

	// Create paths object
	paths, err := NewDefaultTest(tmpdir.Path)
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
}

func TestPaths_Reload_globalconditionals_makeNone(t *testing.T) {
	serverRoot, err := NewTempDir()
	if err != nil {
		t.Error(err)
	}
	defer serverRoot.Close()

	condsRoot, err := NewTempDir()
	if err != nil {
		t.Error(err)
	}
	defer condsRoot.Close()

	if _, err := NewDefault(serverRoot.Path, condsRoot.Path); err != nil {
		t.Error(err)
	}
}

func TestPaths_Reload_globalconditionals_gcpNotExist(t *testing.T) {
	serverRoot, err := NewTempDir()
	if err != nil {
		t.Error(err)
	}
	defer serverRoot.Close()

	if _, err := NewDefault(serverRoot.Path, "/tmp/does/not/exist"); err != nil {
		t.Error(err)
	}
}

func TestPaths_Reload_globalconditionals_one(t *testing.T) {
	serverRoot, err := NewTempDir()
	if err != nil {
		t.Error(err)
	}
	defer serverRoot.Close()
	serverRoot.CreateFile("one", "hello")

	condsRoot, err := NewTempDir()
	if err != nil {
		t.Error(err)
	}
	defer condsRoot.Close()

	uaBlock := makeUABlacklist("target")
	condsRoot.CreateFile("test.yml", uaBlock)

	paths, err := NewDefault(serverRoot.Path, condsRoot.Path)
	if err != nil {
		t.Error(err)
	}

	req := httptest.NewRequest("GET", "/one", nil)
	w := httptest.NewRecorder()

	req.Header.Set("User-Agent", "target")

	didMatch, err := paths.MatchAndServe(w, req)
	if err != nil {
		t.Error(err)
	}

	if didMatch {
		t.Fail()
	}
}

func TestPaths_Reload_globalconditionals_two(t *testing.T) {
	serverRoot, err := NewTempDir()
	if err != nil {
		t.Error(err)
	}
	defer serverRoot.Close()
	serverRoot.CreateFile("one", "hello")

	condsRoot, err := NewTempDir()
	if err != nil {
		t.Error(err)
	}
	defer condsRoot.Close()

	uaBlockTarget := makeUABlacklist("target")
	condsRoot.CreateFile("test.yml", uaBlockTarget)

	uaBlockTargetOne := makeUABlacklist("target1")
	condsRoot.CreateFile("test1.yml", uaBlockTargetOne)

	paths, err := NewDefault(serverRoot.Path, condsRoot.Path)
	if err != nil {
		t.Error(err)
	}

	// target request
	req := httptest.NewRequest("GET", "/one", nil)
	w := httptest.NewRecorder()
	req.Header.Set("User-Agent", "target")
	didMatch, err := paths.MatchAndServe(w, req)
	if err != nil {
		t.Error(err)
	}
	if didMatch {
		t.Fail()
	}

	// target1 request
	reqTwo := httptest.NewRequest("GET", "/one", nil)
	wTwo := httptest.NewRecorder()
	reqTwo.Header.Set("User-Agent", "target1")
	didMatchTwo, err := paths.MatchAndServe(wTwo, reqTwo)
	if err != nil {
		t.Error(err)
	}
	if didMatchTwo {
		t.Fail()
	}
}
