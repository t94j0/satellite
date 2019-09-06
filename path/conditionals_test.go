package path_test

import (
	"fmt"
	"io/ioutil"
	"os"
	"testing"

	"github.com/t94j0/satellite/net/http"

	. "github.com/t94j0/satellite/path"
)

func TestNewRequestConditions(t *testing.T) {
	data := ""
	if _, err := NewRequestConditions([]byte(data)); err != nil {
		t.Error(err)
	}
}

func TestNewRequestConditions_fail(t *testing.T) {
	data := "abc:abc"
	if _, err := NewRequestConditions([]byte(data)); err == nil {
		t.Fail()
	}
}

func TestRequestConditions_ShouldHost_auth_ua_succeed(t *testing.T) {
	// Create HTTP Request
	header := http.Header(make(map[string][]string))
	header.Add("User-Agent", "none")
	mockRequest := &http.Request{Header: header}

	state, file, err := TemporaryDB()
	if err != nil {
		t.Error(err)
	}

	// Create RequestConditions object
	data := `
authorized_useragents:
  - none
`
	conditions, err := NewRequestConditions([]byte(data))
	if err != nil {
		t.Error(err)
	}
	if !conditions.ShouldHost(mockRequest, state) {
		t.Fail()
	}

	if err := RemoveDB(file); err != nil {
		t.Error(err)
	}
}

func TestRequestConditions_ShouldHost_auth_ua_regex(t *testing.T) {
	// Create HTTP Request
	header := http.Header(make(map[string][]string))
	header.Add("User-Agent", "none")
	mockRequest := &http.Request{Header: header}

	state, file, err := TemporaryDB()
	if err != nil {
		t.Error(err)
	}

	// Create RequestConditions object
	data := `
authorized_useragents:
  - non[e|a]
`
	conditions, err := NewRequestConditions([]byte(data))
	if err != nil {
		t.Error(err)
	}
	if !conditions.ShouldHost(mockRequest, state) {
		t.Fail()
	}

	if err := RemoveDB(file); err != nil {
		t.Error(err)
	}
}

func TestRequestConditions_ShouldHost_auth_ua_regex_fail(t *testing.T) {
	data := `
authorized_useragents:
  - none\`
	if _, err := NewRequestConditions([]byte(data)); err == nil {
		t.Fail()
	}
}

func TestRequestConditions_ShouldHost_bl_ua_regex_fail(t *testing.T) {
	data := `
blacklist_useragents:
  - none\`
	if _, err := NewRequestConditions([]byte(data)); err == nil {
		t.Fail()
	}
}

func TestRequestConditions_ShouldHost_auth_ua_fail(t *testing.T) {
	// Create HTTP Request
	header := http.Header(make(map[string][]string))
	header.Add("User-Agent", "none")
	mockRequest := &http.Request{Header: header}

	state, file, err := TemporaryDB()
	if err != nil {
		t.Error(err)
	}

	// Create RequestConditions object
	data := `
authorized_useragents:
  - not_correct
`
	conditions, err := NewRequestConditions([]byte(data))
	if err != nil {
		t.Error(err)
	}
	if conditions.ShouldHost(mockRequest, state) {
		t.Fail()
	}

	if err := RemoveDB(file); err != nil {
		t.Error(err)
	}
}

func TestRequestConditions_ShouldHost_bl_ua_succeed(t *testing.T) {
	// Create HTTP Request
	header := http.Header(make(map[string][]string))
	header.Add("User-Agent", "none")
	mockRequest := &http.Request{Header: header}

	state, file, err := TemporaryDB()
	if err != nil {
		t.Error(err)
	}

	// Create RequestConditions object
	data := `
blacklist_useragents:
  - not_correct
`
	conditions, err := NewRequestConditions([]byte(data))
	if err != nil {
		t.Error(err)
	}
	if !conditions.ShouldHost(mockRequest, state) {
		t.Fail()
	}

	if err := RemoveDB(file); err != nil {
		t.Error(err)
	}
}

func TestRequestConditions_ShouldHost_bl_ua_fail(t *testing.T) {
	// Create HTTP Request
	header := http.Header(make(map[string][]string))
	header.Add("User-Agent", "none")
	mockRequest := &http.Request{Header: header}

	state, file, err := TemporaryDB()
	if err != nil {
		t.Error(err)
	}

	// Create RequestConditions object
	data := `
blacklist_useragents:
  - none
`
	conditions, err := NewRequestConditions([]byte(data))
	if err != nil {
		t.Error(err)
	}
	if conditions.ShouldHost(mockRequest, state) {
		t.Fail()
	}

	if err := RemoveDB(file); err != nil {
		t.Error(err)
	}
}

func TestRequestConditions_ShouldHost_ip_auth_succeed(t *testing.T) {
	// Create HTTP Request
	mockRequest := &http.Request{RemoteAddr: "127.0.0.1"}

	state, file, err := TemporaryDB()
	if err != nil {
		t.Error(err)
	}

	data := `
authorized_iprange:
  - 127.0.0.1
`

	conditions, err := NewRequestConditions([]byte(data))
	if err != nil {
		t.Error(err)
	}
	if !conditions.ShouldHost(mockRequest, state) {
		t.Fail()
	}

	if err := RemoveDB(file); err != nil {
		t.Error(err)
	}
}

func TestRequestConditions_ShouldHost_ip_auth_fail(t *testing.T) {
	// Create HTTP Request
	mockRequest := &http.Request{RemoteAddr: "127.0.0.2"}

	state, file, err := TemporaryDB()
	if err != nil {
		t.Error(err)
	}

	data := `
authorized_iprange:
  - 127.0.0.1`

	conditions, err := NewRequestConditions([]byte(data))
	if err != nil {
		t.Error(err)
	}
	if conditions.ShouldHost(mockRequest, state) {
		t.Fail()
	}

	if err := RemoveDB(file); err != nil {
		t.Error(err)
	}

}

func TestRequestConditions_ShouldHost_ip_auth_cidr_succeed(t *testing.T) {
	// Create HTTP Request
	mockRequest := &http.Request{RemoteAddr: "127.0.0.1"}

	state, file, err := TemporaryDB()
	if err != nil {
		t.Error(err)
	}

	data := `
authorized_iprange:
  - 127.0.0.1/24`

	conditions, err := NewRequestConditions([]byte(data))
	if err != nil {
		t.Error(err)
	}
	if !conditions.ShouldHost(mockRequest, state) {
		t.Fail()
	}

	if err := RemoveDB(file); err != nil {
		t.Error(err)
	}
}

func TestRequestConditions_ShouldHost_ip_auth_cidr_fail(t *testing.T) {
	// Create HTTP Request
	mockRequest := &http.Request{RemoteAddr: "127.0.1.1"}

	state, file, err := TemporaryDB()
	if err != nil {
		t.Error(err)
	}

	data := `
authorized_iprange:
  - 127.0.0.1/24`

	conditions, err := NewRequestConditions([]byte(data))
	if err != nil {
		t.Error(err)
	}
	if conditions.ShouldHost(mockRequest, state) {
		t.Fail()
	}

	if err := RemoveDB(file); err != nil {
		t.Error(err)
	}
}

func TestRequestConditions_ShouldHost_ip_auth_wrongcidr(t *testing.T) {
	// Create HTTP Request
	mockRequest := &http.Request{RemoteAddr: "127.0.1.1"}

	state, file, err := TemporaryDB()
	if err != nil {
		t.Error(err)
	}

	data := `
authorized_iprange:
  - 127.0/0.1/24`

	conditions, err := NewRequestConditions([]byte(data))
	if err != nil {
		t.Error(err)
	}
	if conditions.ShouldHost(mockRequest, state) {
		t.Fail()
	}

	if err := RemoveDB(file); err != nil {
		t.Error(err)
	}
}

func TestRequestConditions_ShouldHost_ip_bl_succeed(t *testing.T) {
	// Create HTTP Request
	mockRequest := &http.Request{RemoteAddr: "127.0.0.1"}

	state, file, err := TemporaryDB()
	if err != nil {
		t.Error(err)
	}

	data := `
blacklist_iprange:
  - 127.0.0.1`

	conditions, err := NewRequestConditions([]byte(data))
	if err != nil {
		t.Error(err)
	}
	if conditions.ShouldHost(mockRequest, state) {
		t.Fail()
	}

	if err := RemoveDB(file); err != nil {
		t.Error(err)
	}
}

func TestRequestConditions_ShouldHost_ip_bl_fail(t *testing.T) {
	// Create HTTP Request
	mockRequest := &http.Request{RemoteAddr: "127.0.0.2"}

	state, file, err := TemporaryDB()
	if err != nil {
		t.Error(err)
	}

	data := `
blacklist_iprange:
  - 127.0.0.1`

	conditions, err := NewRequestConditions([]byte(data))
	if err != nil {
		t.Error(err)
	}
	if !conditions.ShouldHost(mockRequest, state) {
		t.Fail()
	}

	if err := RemoveDB(file); err != nil {
		t.Error(err)
	}
}

func TestRequestConditions_ShouldHost_ip_bl_cidr_success(t *testing.T) {
	// Create HTTP Request
	mockRequest := &http.Request{RemoteAddr: "127.0.0.5"}

	state, file, err := TemporaryDB()
	if err != nil {
		t.Error(err)
	}

	data := `
blacklist_iprange:
  - 127.0.0.1/24`

	conditions, err := NewRequestConditions([]byte(data))
	if err != nil {
		t.Error(err)
	}
	if conditions.ShouldHost(mockRequest, state) {
		t.Fail()
	}

	if err := RemoveDB(file); err != nil {
		t.Error(err)
	}
}

func TestRequestConditions_ShouldHost_ip_bl_cidr_fail(t *testing.T) {
	// Create HTTP Request
	mockRequest := &http.Request{RemoteAddr: "127.0.1.1"}

	state, file, err := TemporaryDB()
	if err != nil {
		t.Error(err)
	}

	data := `
blacklist_iprange:
  - 127.0.0.1/24`

	conditions, err := NewRequestConditions([]byte(data))
	if err != nil {
		t.Error(err)
	}
	if !conditions.ShouldHost(mockRequest, state) {
		t.Fail()
	}

	if err := RemoveDB(file); err != nil {
		t.Error(err)
	}
}

func TestRequestConditions_ShouldHost_method_auth_succeed(t *testing.T) {
	// Create HTTP Request
	mockRequest := &http.Request{Method: "GET"}

	state, file, err := TemporaryDB()
	if err != nil {
		t.Error(err)
	}

	data := `
authorized_methods:
  - GET`

	conditions, err := NewRequestConditions([]byte(data))
	if err != nil {
		t.Error(err)
	}
	if !conditions.ShouldHost(mockRequest, state) {
		t.Fail()
	}

	if err := RemoveDB(file); err != nil {
		t.Error(err)
	}

}

func TestRequestConditions_ShouldHost_method_auth_fail(t *testing.T) {
	// Create HTTP Request
	mockRequest := &http.Request{Method: "POST"}

	state, file, err := TemporaryDB()
	if err != nil {
		t.Error(err)
	}

	data := `
authorized_methods:
  - GET`

	conditions, err := NewRequestConditions([]byte(data))
	if err != nil {
		t.Error(err)
	}
	if conditions.ShouldHost(mockRequest, state) {
		t.Fail()
	}

	if err := RemoveDB(file); err != nil {
		t.Error(err)
	}
}

func TestRequestConditions_ShouldHost_header_auth_succeed(t *testing.T) {
	// Create HTTP Request
	header := http.Header(make(map[string][]string))
	header.Add("Header", "test")
	mockRequest := &http.Request{Header: header}

	state, file, err := TemporaryDB()
	if err != nil {
		t.Error(err)
	}

	data := `
authorized_headers:
  Header: test
`

	conditions, err := NewRequestConditions([]byte(data))
	if err != nil {
		t.Error(err)
	}
	if !conditions.ShouldHost(mockRequest, state) {
		t.Fail()
	}

	if err := RemoveDB(file); err != nil {
		t.Error(err)
	}
}

func TestRequestConditions_ShouldHost_header_auth_fail(t *testing.T) {
	// Create HTTP Request
	header := http.Header(make(map[string][]string))
	header.Add("Header", "none")
	mockRequest := &http.Request{Header: header}

	state, file, err := TemporaryDB()
	if err != nil {
		t.Error(err)
	}

	data := `
authorized_headers:
  Header: test
`

	conditions, err := NewRequestConditions([]byte(data))
	if err != nil {
		t.Error(err)
	}
	if conditions.ShouldHost(mockRequest, state) {
		t.Fail()
	}

	if err := RemoveDB(file); err != nil {
		t.Error(err)
	}
}

func TestRequestConditions_ShouldHost_ja3(t *testing.T) {
	// TODO: Add tests for JA3

}

func TestRequestConditions_ShouldHost_exec_succeed(t *testing.T) {
	// Create HTTP Request
	mockRequest, err := http.NewRequest("GET", "/", nil)
	if err != nil {
		t.Error(err)
	}

	state, file, err := TemporaryDB()
	if err != nil {
		t.Error(err)
	}

	// Add script
	data := "#!/usr/bin/env python\nprint('ok')"
	shellfile, err := ioutil.TempFile("", "file")
	if err != nil {
		t.Error(err)
	}
	defer os.Remove(shellfile.Name())

	if _, err := shellfile.Write([]byte(data)); err != nil {
		t.Error(err)
	}

	if err := shellfile.Chmod(0777); err != nil {
		t.Error(err)
	}

	if err := shellfile.Close(); err != nil {
		t.Error(err)
	}

	// Execute
	content := "exec:\n"
	content += fmt.Sprintf("  script: %s\n", shellfile.Name())
	content += "  output: ok"

	conditions, err := NewRequestConditions([]byte(content))
	if err != nil {
		t.Error(err)
	}
	if !conditions.ShouldHost(mockRequest, state) {
		t.Fail()
	}

	if err := RemoveDB(file); err != nil {
		t.Error(err)
	}
}

func TestRequestConditions_ShouldHost_exec_fail(t *testing.T) {
	// Create HTTP Request
	mockRequest, err := http.NewRequest("GET", "/", nil)
	if err != nil {
		t.Error(err)
	}

	state, file, err := TemporaryDB()
	if err != nil {
		t.Error(err)
	}

	// Add script
	data := "#!/usr/bin/env python\nprint('not_ok')"
	shellfile, err := ioutil.TempFile("", "file")
	if err != nil {
		t.Error(err)
	}
	defer os.Remove(shellfile.Name())

	if _, err := shellfile.Write([]byte(data)); err != nil {
		t.Error(err)
	}

	if err := shellfile.Chmod(0777); err != nil {
		t.Error(err)
	}

	if err := shellfile.Close(); err != nil {
		t.Error(err)
	}

	// Execute
	content := "exec:\n"
	content += fmt.Sprintf("  script: %s\n", shellfile.Name())
	content += "  output: ok"

	conditions, err := NewRequestConditions([]byte(content))
	if err != nil {
		t.Error(err)
	}
	if conditions.ShouldHost(mockRequest, state) {
		t.Fail()
	}

	if err := RemoveDB(file); err != nil {
		t.Error(err)
	}
}

func TestRequestConditions_ShouldHost_notserving(t *testing.T) {
	// Create HTTP Request
	mockRequest, err := http.NewRequest("GET", "/", nil)
	if err != nil {
		t.Error(err)
	}

	state, file, err := TemporaryDB()
	if err != nil {
		t.Error(err)
	}

	// Create RequestConditions object
	data := `
not_serving: true`

	conditions, err := NewRequestConditions([]byte(data))
	if err != nil {
		t.Error(err)
	}
	if conditions.ShouldHost(mockRequest, state) {
		t.Fail()
	}

	if err := RemoveDB(file); err != nil {
		t.Error(err)
	}
}

func TestRequestConditions_ShouldHost_serve_one_succeed(t *testing.T) {
	// Create HTTP Request
	mockRequest, err := http.NewRequest("GET", "/", nil)
	if err != nil {
		t.Error(err)
	}

	state, file, err := TemporaryDB()
	if err != nil {
		t.Error(err)
	}

	// Create RequestConditions object
	data := `
serve: 1`

	conditions, err := NewRequestConditions([]byte(data))
	if err != nil {
		t.Error(err)
	}
	if !conditions.ShouldHost(mockRequest, state) {
		t.Fail()
	}

	if err := RemoveDB(file); err != nil {
		t.Error(err)
	}
}

func TestRequestConditions_ShouldHost_serve_one_fail(t *testing.T) {
	// Create HTTP Request
	mockRequest, err := http.NewRequest("GET", "/", nil)
	if err != nil {
		t.Error(err)
	}

	state, file, err := TemporaryDB()
	if err != nil {
		t.Error(err)
	}

	if err := state.Hit(mockRequest); err != nil {
		t.Error(err)
	}

	// Create RequestConditions object
	data := `
serve: 1`

	conditions, err := NewRequestConditions([]byte(data))
	if err != nil {
		t.Error(err)
	}
	if conditions.ShouldHost(mockRequest, state) {
		t.Fail()
	}

	if err := RemoveDB(file); err != nil {
		t.Error(err)
	}
}

func TestRequestConditions_ShouldHost_prereq_none(t *testing.T) {
	// Create HTTP Request
	mockRequest, err := http.NewRequest("GET", "/", nil)
	if err != nil {
		t.Error(err)
	}

	state, file, err := TemporaryDB()
	if err != nil {
		t.Error(err)
	}

	if err := state.Hit(mockRequest); err != nil {
		t.Error(err)
	}

	// Create RequestConditions object
	data := `
prereq:`

	conditions, err := NewRequestConditions([]byte(data))
	if err != nil {
		t.Error(err)
	}
	if !conditions.ShouldHost(mockRequest, state) {
		t.Fail()
	}

	if err := RemoveDB(file); err != nil {
		t.Error(err)
	}
}

func TestRequestConditions_ShouldHost_prereq_one_succeed(t *testing.T) {
	state, file, err := TemporaryDB()
	if err != nil {
		t.Error(err)
	}

	firstHit, err := http.NewRequest("GET", "/", nil)
	if err != nil {
		t.Error(err)
	}

	payloadHit, err := http.NewRequest("GET", "/payload", nil)
	if err != nil {
		t.Error(err)
	}

	if err := state.Hit(firstHit); err != nil {
		t.Error(err)
	}

	// Create RequestConditions object
	data := `
prereq:
  - /`

	conditions, err := NewRequestConditions([]byte(data))
	if err != nil {
		t.Error(err)
	}
	if !conditions.ShouldHost(payloadHit, state) {
		t.Fail()
	}

	if err := RemoveDB(file); err != nil {
		t.Error(err)
	}
}

func TestRequestConditions_ShouldHost_prereq_one_fail(t *testing.T) {
	state, file, err := TemporaryDB()
	if err != nil {
		t.Error(err)
	}

	firstHit, err := http.NewRequest("GET", "/one", nil)
	if err != nil {
		t.Error(err)
	}

	payloadHit, err := http.NewRequest("GET", "/two", nil)
	if err != nil {
		t.Error(err)
	}

	if err := state.Hit(firstHit); err != nil {
		t.Error(err)
	}

	// Create RequestConditions object
	data := `
prereq:
  - /`

	conditions, err := NewRequestConditions([]byte(data))
	if err != nil {
		t.Error(err)
	}
	if conditions.ShouldHost(payloadHit, state) {
		t.Fail()
	}

	if err := RemoveDB(file); err != nil {
		t.Error(err)
	}
}
