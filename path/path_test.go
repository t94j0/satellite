package path_test

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"net/url"
	"os"
	"strings"
	"testing"

	"github.com/t94j0/satellite/net/http"
	"github.com/t94j0/satellite/net/http/httptest"
	. "github.com/t94j0/satellite/path"
)

func TestNewPath(t *testing.T) {
	file, err := ioutil.TempFile("", "satellitetest")
	if err != nil {
		t.Error(err)
	}
	defer os.Remove(file.Name())

	if _, err := file.WriteString(`
content_type: application/x-www-form-urlencoded
serve: 1
`); err != nil {
		t.Error(err)
	}
	if err := file.Close(); err != nil {
		t.Error(err)
	}
	if _, err := NewPath(file.Name()); err != nil {
		t.Error(err)
	}
}

func TestNewPath_path_not_exist(t *testing.T) {
	_, err := NewPath("/this-file-should-not-exist")
	if err == nil {
		t.Error(err)
	}
}

func TestNewPathData_bad_yaml(t *testing.T) {
	data := `abc:abc`
	if _, err := NewPathData([]byte(data)); err == nil {
		t.Error(err)
	}
}

func TestNewPathArray(t *testing.T) {
	file, err := ioutil.TempFile("", "satellitetest")
	if err != nil {
		t.Error(err)
	}
	defer os.Remove(file.Name())

	if _, err := file.WriteString(`
- path: /index
  content_type: application/x-www-form-urlencoded
  serve: 1
- path: /payload
`); err != nil {
		t.Error(err)
	}
	if err := file.Close(); err != nil {
		t.Error(err)
	}
	if _, err := NewPathArray(file.Name()); err != nil {
		t.Error(err)
	}
}

func TestNewPathArrayData_fail(t *testing.T) {
	data := `abc:abc`
	if _, err := NewPathArrayData([]byte(data)); err == nil {
		t.Error(err)
	}
}

func TestPath_ContentHeaders_type(t *testing.T) {
	file, err := ioutil.TempFile("", "satellitetest")
	if err != nil {
		t.Error(err)
	}
	defer os.Remove(file.Name())

	if _, err := file.WriteString(`
content_type: application/x-www-form-urlencoded
`); err != nil {
		t.Error(err)
	}
	if err := file.Close(); err != nil {
		t.Error(err)
	}
	path, err := NewPath(file.Name())
	if err != nil {
		t.Error(err)
	}
	headers := path.ContentHeaders()
	if headers["Content-Type"] != "application/x-www-form-urlencoded" {
		t.Fail()
	}
}

func TestPath_ContentHeaders_disposition(t *testing.T) {
	file, err := ioutil.TempFile("", "satellitetest")
	if err != nil {
		t.Error(err)
	}
	defer os.Remove(file.Name())

	if _, err := file.WriteString(`
disposition:
  type: application/msword
  file_name: file.doc
`); err != nil {
		t.Error(err)
	}
	if err := file.Close(); err != nil {
		t.Error(err)
	}
	path, err := NewPath(file.Name())
	if err != nil {
		t.Error(err)
	}
	headers := path.ContentHeaders()
	if headers["Content-Disposition"] != "application/msword; filename=\"file.doc\"" {
		t.Fail()
	}
}

func TestPath_ContentHeaders_disposition_typeonly(t *testing.T) {
	file, err := ioutil.TempFile("", "satellitetest")
	if err != nil {
		t.Error(err)
	}
	defer os.Remove(file.Name())

	if _, err := file.WriteString(`
disposition:
  type: application/msword
`); err != nil {
		t.Error(err)
	}
	if err := file.Close(); err != nil {
		t.Error(err)
	}
	path, err := NewPath(file.Name())
	if err != nil {
		t.Error(err)
	}
	headers := path.ContentHeaders()
	if headers["Content-Disposition"] != "application/msword" {
		t.Fail()
	}
}

func TestPath_ShouldHost(t *testing.T) {
	// Create Request
	req, err := http.NewRequest("GET", "/", nil)
	if err != nil {
		t.Error(err)
	}
	req.Header.Add("User-Agent", "none")

	// Create State
	state, file, err := TemporaryDB()
	if err != nil {
		t.Error(err)
	}

	// Create Path
	data := `
authorized_useragents:
  - none`

	path, err := NewPathData([]byte(data))
	if err != nil {
		t.Error(err)
	}

	// Execute ShouldHost
	shouldHost := path.ShouldHost(req, state)

	if !shouldHost {
		t.Fail()
	}

	// Remove file
	if err := RemoveDB(file); err != nil {
		t.Error(err)
	}
}

func TestPath_Render(t *testing.T) {
	// Create Request
	req, err := http.NewRequest("GET", "/", nil)
	if err != nil {
		t.Error(err)
	}
	req.Header.Add("User-Agent", "none")

	// Create readable file
	serveFile, err := ioutil.TempFile("", "")
	if err != nil {
		t.Error(err)
	}
	if _, err := serveFile.WriteString("Hello!"); err != nil {
		t.Error(err)
	}
	defer os.Remove(serveFile.Name())
	if err := serveFile.Close(); err != nil {
		t.Error(err)
	}

	// Create Path
	data := fmt.Sprintf("hosted_file: %s\n", serveFile.Name())
	data += "authorized_useragents:\n"
	data += "  - none\n"

	path, err := NewPathData([]byte(data))
	if err != nil {
		t.Error(err)
	}

	// Create handler
	handlerFunc := func(w http.ResponseWriter, req *http.Request) {
		if err := path.ServeHTTP(w, req); err != nil {
			t.Error(err)
		}
	}
	handler := http.HandlerFunc(handlerFunc)

	// Test HTTP connection
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)
	if rr.Body.String() != "Hello!" {
		t.Fail()
	}
}

func TestPath_Render_nopath(t *testing.T) {
	// Create Request
	req, err := http.NewRequest("GET", "/", nil)
	if err != nil {
		t.Error(err)
	}
	req.Header.Add("User-Agent", "none")

	// Create Path
	data := "hosted_file: /path-does-not-exist\n"
	data += "authorized_useragents:\n"
	data += "  - none\n"

	path, err := NewPathData([]byte(data))
	if err != nil {
		t.Error(err)
	}

	// Create handler
	handlerFunc := func(w http.ResponseWriter, req *http.Request) {
		if err := path.ServeHTTP(w, req); err == nil {
			t.Fail()
		}
	}
	handler := http.HandlerFunc(handlerFunc)

	// Test HTTP connection
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)
	if rr.Body.String() == "Hello!" {
		t.Fail()
	}
}

func TestPath_Render_proxy(t *testing.T) {
	// Create Request
	req := httptest.NewRequest("GET", "/", nil)
	req.Header.Add("User-Agent", "none")

	// Create Path
	data := "proxy: https://www.google.com/\n"
	data += "authorized_useragents:\n"
	data += "  - none\n"

	path, err := NewPathData([]byte(data))
	if err != nil {
		t.Error(err)
	}

	// Create handler
	handlerFunc := func(w http.ResponseWriter, req *http.Request) {
		if err := path.ServeHTTP(w, req); err != nil {
			t.Error(err)
		}
	}
	handler := http.HandlerFunc(handlerFunc)

	// Test HTTP connection
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

}

func TestPath_Render_proxyhost_fail(t *testing.T) {
	// Create Request
	req, err := http.NewRequest("GET", "/", nil)
	if err != nil {
		t.Error(err)
	}
	req.Header.Add("User-Agent", "none")

	// Create Path
	data := "proxy: abc\n"
	data += "authorized_useragents:\n"
	data += "  - none\n"

	path, err := NewPathData([]byte(data))
	if err != nil {
		t.Error(err)
	}

	// Create handler
	handlerFunc := func(w http.ResponseWriter, req *http.Request) {
		if err := path.ServeHTTP(w, req); err == nil {
			t.Fail()
		}
	}
	handler := http.HandlerFunc(handlerFunc)

	// Test HTTP connection
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)
}

func TestNewPathData_invalidyaml(t *testing.T) {
	data := "abc"
	if _, err := NewPathData([]byte(data)); err == nil {
		t.Fail()
	}
}

func TestNewPathArray_nofile(t *testing.T) {
	if _, err := NewPathArray("/this-file-should-not-exist"); err == nil {
		t.Fail()
	}
}

func TestPath_Render_credentialcapture(t *testing.T) {
	// Create data values
	bodyVal := url.Values{}
	bodyVal.Set("username", "test1")
	bodyVal.Set("password", "test2")

	// Create request
	req := httptest.NewRequest("POST", "/", bytes.NewBufferString(bodyVal.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	// Create temp file for writing
	outFile, err := ioutil.TempFile("", "")
	if err != nil {
		t.Error(err)
	}

	// Create Path
	data := "credential_capture:\n"
	data += fmt.Sprintf("  file_output: %s\n", outFile.Name())

	path, err := NewPathData([]byte(data))
	if err != nil {
		t.Error(err)
	}

	// Create handler
	handlerFunc := func(w http.ResponseWriter, req *http.Request) {
		if err := path.ServeHTTP(w, req); err != nil {
			t.Error(err)
		}
	}
	handler := http.HandlerFunc(handlerFunc)

	// Test HTTP connection
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	outputData, err := ioutil.ReadAll(outFile)
	if err != nil {
		t.Error(err)
	}
	outputDataString := strings.TrimRight(string(outputData), "\n")

	outValues, err := url.ParseQuery(outputDataString)
	if err != nil {
		t.Error(err)
	}

	if outValues.Get("username") != "test1" || outValues.Get("password") != "test2" {
		t.Fail()
	}

	if err := outFile.Close(); err != nil {
		t.Error(err)
	}

	// Clean up files
	if err := os.Remove(outFile.Name()); err != nil {
		t.Error(err)
	}
}

func TestPath_Render_credentialcapture_nil(t *testing.T) {
	// Create request
	req := httptest.NewRequest("POST", "/", nil)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	// Create temp file for writing
	outFile, err := ioutil.TempFile("", "")
	if err != nil {
		t.Error(err)
	}

	// Create Path
	data := "credential_capture:\n"
	data += fmt.Sprintf("  file_output: %s\n", outFile.Name())

	path, err := NewPathData([]byte(data))
	if err != nil {
		t.Error(err)
	}

	// Create handler
	handlerFunc := func(w http.ResponseWriter, req *http.Request) {
		if err := path.ServeHTTP(w, req); err != nil {
			t.Error(err)
		}
	}
	handler := http.HandlerFunc(handlerFunc)

	// Test HTTP connection
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	outputData, err := ioutil.ReadAll(outFile)
	if err != nil {
		t.Error(err)
	}

	if string(outputData) != "\n" {
		t.Fail()
	}

	if err := outFile.Close(); err != nil {
		t.Error(err)
	}

	// Clean up files
	if err := os.Remove(outFile.Name()); err != nil {
		t.Error(err)
	}
}

func TestPath_Render_credentialcapture_badfile(t *testing.T) {
	// Create request
	req := httptest.NewRequest("POST", "/", nil)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	// Create Path
	data := "credential_capture:\n"
	data += "  file_output: /root/root/root/root/cannot-write\n"

	path, err := NewPathData([]byte(data))
	if err != nil {
		t.Error(err)
	}

	// Create handler
	handlerFunc := func(w http.ResponseWriter, req *http.Request) {
		if err := path.ServeHTTP(w, req); err == nil {
			t.Fail()
		}
	}
	handler := http.HandlerFunc(handlerFunc)

	// Test HTTP connection
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)
}
