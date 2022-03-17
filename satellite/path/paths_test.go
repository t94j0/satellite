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

const Sentinal = "sentinal"

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

func (t TempDir) CreateDirectory(name string) {
	fullpath := filepath.Join(t.Path, name)
	os.Mkdir(fullpath, 0777)
}

func (t TempDir) CreateIndexFile() {
	t.CreateFile("index.html", Sentinal)
}

func (t TempDir) CreatePathList(content string) {
	t.CreateFile("pathList.yml", content)
}

func (t TempDir) CreatePathListIndex(content ...string) {
	pathContent := `- path: /index.html
  hosted_file: /index.html`
	for _, c := range content {
		pathContent += "\n  " + c
	}
	t.CreatePathList(pathContent)
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
	tmpdir.CreateIndexFile()
	pathsContent := `- path: /index.html
  hosted_file: /index.html`
	tmpdir.CreatePathList(pathsContent)

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
	tmpdir.CreatePathList("- path: /index.html")
	if _, err := NewDefaultTest(tmpdir.Path); err != nil {
		t.Error(err)
	}
}

func TestNew_different_hosted(t *testing.T) {
	tmpdir, err := NewTempDir()
	if err != nil {
		t.Error(err)
	}
	defer tmpdir.Close()
	tmpdir.CreatePathList(`- path: /index.html
  hosted_file: abc`)
	tmpdir.CreateFile("abc", Sentinal)

	paths, err := NewDefaultTest(tmpdir.Path)
	if err != nil {
		t.Error(err)
	}

	req := httptest.NewRequest("GET", "/index.html", nil)
	w := httptest.NewRecorder()
	didMatch, err := paths.MatchAndServe(w, req)
	if err != nil {
		t.Error(err)
	}

	if !didMatch {
		t.Fail()
	}

	data, err := ioutil.ReadAll(w.Result().Body)
	if err != nil {
		t.Fail()
	}
	if string(data) != Sentinal {
		t.Fail()
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
	tmpdir.CreatePathList(proxyContent)
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
	tmpdir.CreateFile("pathList.yml", oneData)
	paths, err := NewDefaultTest(tmpdir.Path)
	if err != nil {
		t.Error(err)
	}
	if _, ok := paths.Match("/"); !ok {
		t.Fail()
	}
}

func TestPaths_MatchAndServe_file_success(t *testing.T) {
	// Create project directory
	tmpdir, err := NewTempDir()
	if err != nil {
		t.Error(err)
	}
	defer tmpdir.Close()

	tmpdir.CreateIndexFile()
	tmpdir.CreatePathListIndex()

	// Create paths object
	paths, err := NewDefaultTest(tmpdir.Path)
	if err != nil {
		t.Error(err)
	}

	// Create HTTP request
	req := httptest.NewRequest("GET", "/index.html", nil)
	w := httptest.NewRecorder()

	// Execute request
	didMatch, err := paths.MatchAndServe(w, req)
	if err != nil {
		t.Error(err)
	}
	if didMatch == false || w.Code != 200 || w.Body.String() != Sentinal {
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

	tmpdir.CreateIndexFile()
	tmpdir.CreatePathListIndex("content_type: application/json")

	// Create paths object
	paths, err := NewDefaultTest(tmpdir.Path)
	if err != nil {
		t.Error(err)
	}

	// Create HTTP request
	req := httptest.NewRequest("GET", "/index.html", nil)
	w := httptest.NewRecorder()

	// Execute request
	didMatch, err := paths.MatchAndServe(w, req)
	if err != nil {
		t.Error(err)
	}
	if didMatch == false || w.Code != 200 || w.Body.String() != Sentinal {
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

	tmpdir.CreateIndexFile()
	pathData := `- path: /index.html
  hosted_file: /index.html
  authorized_useragents:
    - none
  on_failure:
    redirect: https://aws.amazon.com`
	tmpdir.CreatePathList(pathData)

	// Create paths object
	paths, err := NewDefaultTest(tmpdir.Path)
	if err != nil {
		t.Error(err)
	}

	// Create HTTP request
	req := httptest.NewRequest("GET", "/index.html", nil)
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
	// defer tmpdir.Close()
	const SentinalOne = Sentinal + "1"
	tmpdir.CreateIndexFile()
	tmpdir.CreateFile("one", SentinalOne)
	indexData := `- path: /index.html
  hosted_file: /index.html
  authorized_useragents:
    - none
  on_failure:
    render: /one

- path: /one
  hosted_file: /one`
	tmpdir.CreatePathList(indexData)

	// Create paths object
	paths, err := NewDefaultTest(tmpdir.Path)
	if err != nil {
		t.Error(err)
	}

	// Create HTTP request
	req := httptest.NewRequest("GET", "/index.html", nil)
	req.Header.Add("User-Agent", "wont_match")
	w := httptest.NewRecorder()

	// Execute request
	didMatch, err := paths.MatchAndServe(w, req)
	if err != nil {
		t.Error(err)
	}
	if didMatch == false || w.Code != 200 || w.Body.String() != SentinalOne {
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
	req := httptest.NewRequest("GET", "/index.html", nil)
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
	tmpdir.CreateIndexFile()
	oneData := `- path: /index.html
  hosted_file: /index.html
  authorized_useragents:
    - none
  on_failure:
    render: /memer.html`
	tmpdir.CreatePathList(oneData)

	// Create paths object
	paths, err := NewDefaultTest(tmpdir.Path)
	if err != nil {
		t.Error(err)
	}

	// Create HTTP request
	req := httptest.NewRequest("GET", "/index.html", nil)
	req.Header.Add("User-Agent", "wont_match")
	w := httptest.NewRecorder()

	// Execute request
	// if _, err := paths.MatchAndServe(w, req); err.Error() != "path not found" {
	// 	t.Fail()
	// }
	if _, err := paths.MatchAndServe(w, req); err != nil {
		if err.Error() != "path does not exist" {
			t.Fail()
		}
	}
}

func TestPaths_MatchAndServe_file_failure_render_meme(t *testing.T) {
	// Create project directory
	tmpdir, err := NewTempDir()
	if err != nil {
		t.Error(err)
	}
	defer tmpdir.Close()
	tmpdir.CreateIndexFile()
	tmpdir.CreatePathList(`- path: /index.html
  hosted_file: /index.html
  authorized_useragents:
    - none`)
	// Create paths object
	paths, err := NewDefaultTest(tmpdir.Path)
	if err != nil {
		t.Error(err)
	}

	// Create HTTP request
	req := httptest.NewRequest("GET", "/index.html", nil)
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
	tmpdir.CreateIndexFile()
	tmpdir.CreatePathListIndex()

	// Create paths object
	paths, err := NewDefaultTest(tmpdir.Path)
	if err != nil {
		t.Error(err)
	}

	// Create HTTP request
	req := httptest.NewRequest("GET", "/index.html", nil)
	w := httptest.NewRecorder()

	// Execute request
	if err := paths.Serve(w, req); err != nil {
		t.Error(err)
	}

	if w.Code != 200 || w.Body.String() != Sentinal {
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
	req := httptest.NewRequest("GET", "/index.html", nil)
	w := httptest.NewRecorder()

	// Execute request
	if err := paths.Serve(w, req); err != nil {
		if err.Error() != "not_found render page not found" {
			t.Fail()
		}
	}
}

func TestPaths_Serve_success_headers(t *testing.T) {
	// Create project directory
	tmpdir, err := NewTempDir()
	if err != nil {
		t.Error(err)
	}
	defer tmpdir.Close()
	tmpdir.CreateIndexFile()
	tmpdir.CreatePathListIndex("content_type: application/json")

	// Create paths object
	paths, err := NewDefaultTest(tmpdir.Path)
	if err != nil {
		t.Error(err)
	}

	// Create HTTP request
	req := httptest.NewRequest("GET", "/index.html", nil)
	w := httptest.NewRecorder()

	// Execute request
	if err := paths.Serve(w, req); err != nil {
		t.Error(err)
	}

	contentHeader := w.Header().Get("Content-Type")
	if w.Code != 200 || w.Body.String() != Sentinal || contentHeader != "application/json" {
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
	tmpdir.CreateIndexFile()

	// Create paths object
	paths, err := NewDefaultTest(tmpdir.Path)
	if err != nil {
		t.Error(err)
	}

	// Create HTTP request
	req := httptest.NewRequest("GET", "/index.html", nil)
	w := httptest.NewRecorder()

	// Execute request
	didMatch, err := paths.MatchAndServe(w, req)
	if err != nil {
		t.Error(err)
	}
	if didMatch == false || w.Code != 200 || w.Body.String() != Sentinal {
		t.Fail()
	}
}

func TestPaths_MatchAndServe_glob_file(t *testing.T) {
	// Create project directory
	tmpdir, err := NewTempDir()
	if err != nil {
		t.Error(err)
	}
	defer tmpdir.Close()
	tmpdir.CreateFile("first.html", Sentinal)
	tmpdir.CreateFile("second.html", "lol")

	pathList := `- path: /*.html
  hosted_file: /first.html`

	tmpdir.CreatePathList(pathList)

	// Create paths object
	paths, err := NewDefaultTest(tmpdir.Path)
	if err != nil {
		t.Error(err)
	}

	// Create HTTP request
	req := httptest.NewRequest("GET", "/second.html", nil)
	w := httptest.NewRecorder()

	// Execute request
	didMatch, err := paths.MatchAndServe(w, req)
	if err != nil {
		t.Error(err)
	}
	if didMatch == false || w.Code != 200 || w.Body.String() != Sentinal {
		t.Fail()
	}
}

func buildTestEnv() (TempDir, error) {
	tmpdir, err := NewTempDir()
	if err != nil {
		return TempDir{}, err
	}
	tmpdir.CreateDirectory("testdir1")
	tmpdir.CreateFile("/testdir1/first.html", Sentinal)
	tmpdir.CreateFile("/testdir1/second.html", Sentinal+"4")
	tmpdir.CreateDirectory("testdir2")
	tmpdir.CreateFile("/testdir2/first.html", Sentinal+"2")
	tmpdir.CreateFile("/second.html", Sentinal+"3")

	return tmpdir, nil
}

func TestPaths_MatchAndServe_glob_extensions_block(t *testing.T) {
	tmpdir, err := buildTestEnv()
	if err != nil {
		t.Error(err)
	}
	defer tmpdir.Close()

	pathList := `- path: /**.html
  authorized_methods: [PUT]`
	tmpdir.CreatePathList(pathList)

	// Create paths object
	paths, err := NewDefaultTest(tmpdir.Path)
	if err != nil {
		t.Error(err)
	}

	// Make GET request to path that should only be PUT
	req := httptest.NewRequest("GET", "/testdir1/first.html", nil)
	w := httptest.NewRecorder()

	// Execute request
	didMatch, err := paths.MatchAndServe(w, req)
	if err != nil {
		t.Error(err)
	}
	if didMatch {
		t.Fatal("Request should not have matched due to glob block on GET")
	}
}

func TestPaths_MatchAndServe_glob_extensions_block_multiple(t *testing.T) {
	tmpdir, err := buildTestEnv()
	if err != nil {
		t.Error(err)
	}
	defer tmpdir.Close()

	pathList := `- path: /**.html
  authorized_methods: [PUT]
- path: /testdir2/*.html
  authorized_methods: [GET]`
	tmpdir.CreatePathList(pathList)

	// Create paths object
	paths, err := NewDefaultTest(tmpdir.Path)
	if err != nil {
		t.Error(err)
	}

	// Make GET request to path that should only be PUT
	req := httptest.NewRequest("GET", "/testdir1/first.html", nil)
	w := httptest.NewRecorder()

	// Execute request
	didMatch, err := paths.MatchAndServe(w, req)
	if err != nil {
		t.Error(err)
	}
	if didMatch {
		t.Fatal("GET request to /testdir1/first.html should not have matched")
	}

	// Make GET request to path that should only be PUT
	req2 := httptest.NewRequest("GET", "/testdir2/first.html", nil)
	w2 := httptest.NewRecorder()

	// Execute request
	didMatch2, err := paths.MatchAndServe(w2, req2)
	if err != nil {
		t.Error(err)
	}
	if !didMatch2 {
		t.Fatal("GET request to /testdir2/second.html should have matched")
	}
}

func TestPaths_MatchAndServe_jlob_directory(t *testing.T) {
	// Create project directory
	tmpdir, err := buildTestEnv()
	if err != nil {
		t.Fatal(err)
	}
	defer tmpdir.Close()

	pathList := `- path: /testdir1/*
  hosted_file: /testdir1/first.html`

	tmpdir.CreatePathList(pathList)

	// Create paths object
	paths, err := NewDefaultTest(tmpdir.Path)
	if err != nil {
		t.Error(err)
	}

	// Create HTTP request
	req := httptest.NewRequest("GET", "/testdir1/second.html", nil)
	w := httptest.NewRecorder()

	// Execute request
	didMatch, err := paths.MatchAndServe(w, req)
	if err != nil {
		t.Error(err)
	}
	if !didMatch {
		t.Error("Request should have matched:", didMatch)
	}
	if w.Code != 200 {
		t.Error("Request should have been 200:", w.Code)
	}
	if w.Body.String() != Sentinal {
		t.Error("/testdir1/second.html should have returned", Sentinal, "but returned", w.Body.String())
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
	serverRoot.CreateIndexFile()

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

	req := httptest.NewRequest("GET", "/index.html", nil)
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
	serverRoot.CreateIndexFile()

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
	req := httptest.NewRequest("GET", "/index.html", nil)
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
	reqTwo := httptest.NewRequest("GET", "/index.html", nil)
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
