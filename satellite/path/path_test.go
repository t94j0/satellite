package path_test

import (
	"io/ioutil"
	"os"
	"testing"

	"github.com/t94j0/satellite/net/http"
	"github.com/t94j0/satellite/satellite/geoip"
	. "github.com/t94j0/satellite/satellite/path"
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
	shouldHost := path.ShouldHost(req, state, geoip.DB{})

	if !shouldHost {
		t.Fail()
	}

	// Remove file
	if err := RemoveDB(file); err != nil {
		t.Error(err)
	}
}
