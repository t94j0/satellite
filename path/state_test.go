package path_test

import (
	"io/ioutil"
	"os"
	"testing"

	"github.com/t94j0/satellite/net/http"
	. "github.com/t94j0/satellite/path"
)

// Helper
func TemporaryDB() (*State, string, error) {
	dirname, err := ioutil.TempDir("", "satellitetest")
	if err != nil {
		return nil, "", err
	}

	state, err := NewState(dirname)
	if err != nil {
		return nil, "", err
	}

	return state, dirname, nil
}

func RemoveDB(dirname string) error {
	return os.RemoveAll(dirname)
}

// Tests
func TestNewState(t *testing.T) {
	_, file, err := TemporaryDB()
	if err != nil {
		t.Error(err)
		t.FailNow()
	}
	if err := RemoveDB(file); err != nil {
		t.Error(err)
	}
}

func TestNewState_badpath(t *testing.T) {
	if _, err := NewState("/this-path-should-exist"); err == nil {
		t.Fail()
	}
}

func TestState_Hit(t *testing.T) {
	state, file, err := TemporaryDB()
	if err != nil {
		t.Error(err)
		t.FailNow()
	}

	req, err := http.NewRequest("GET", "/", nil)
	if err != nil {
		t.Error(err)
	}

	if err := state.Hit(req); err != nil {
		t.Error(err)
	}

	if err := RemoveDB(file); err != nil {
		t.Error(err)
	}
}

func TestState_gethits_none(t *testing.T) {
	state, file, err := TemporaryDB()
	if err != nil {
		t.Error(err)
		t.FailNow()
	}

	times, err := state.GetHits("/")
	if err != nil {
		t.Error(err)
	}
	if times != 0 {
		t.Fail()
	}

	if err := RemoveDB(file); err != nil {
		t.Error(err)
	}
}

func TestState_gethits_one(t *testing.T) {
	state, file, err := TemporaryDB()
	if err != nil {
		t.Error(err)
		t.FailNow()
	}

	req, err := http.NewRequest("GET", "/", nil)
	if err != nil {
		t.Error(err)
	}

	if err := state.Hit(req); err != nil {
		t.Error(err)
	}

	times, err := state.GetHits("/")
	if err != nil {
		t.Error(err)
	}

	if times != 1 {
		t.Fail()
	}

	if err := RemoveDB(file); err != nil {
		t.Error(err)
	}
}

func TestState_gethits_two(t *testing.T) {
	state, file, err := TemporaryDB()
	if err != nil {
		t.Error(err)
		t.FailNow()
	}
	req, err := http.NewRequest("GET", "/", nil)
	if err != nil {
		t.Error(err)
	}

	if err := state.Hit(req); err != nil {
		t.Error(err)
	}
	if err := state.Hit(req); err != nil {
		t.Error(err)
	}
	times, err := state.GetHits("/")
	if err != nil {
		t.Error(err)
	}
	if times != 2 {
		t.Fail()
	}

	if err := RemoveDB(file); err != nil {
		t.Error(err)
	}
}

func TestState_hit_urlfail(t *testing.T) {
	state, file, err := TemporaryDB()
	if err != nil {
		t.Error(err)
		t.FailNow()
	}
	req := &http.Request{}
	if err := state.Hit(req); err != ErrNoURL {
		t.Fail()
	}

	if err := RemoveDB(file); err != nil {
		t.Error(err)
	}
}
