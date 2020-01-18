package path_test

import (
	"net"
	"testing"

	. "github.com/t94j0/satellite/satellite/path"
)

func TestClientID(t *testing.T) {
	cid := NewClientID()
	if cid == nil {
		t.Fail()
	}
}

func TestClientID_hit(t *testing.T) {
	cid := NewClientID()
	cid.Hit(net.ParseIP("127.0.0.1"), "/")
}

func TestClientID_match_none(t *testing.T) {
	cid := NewClientID()
	if !cid.Match(net.ParseIP("127.0.0.1"), []string{}) {
		t.Fail()
	}
}

func TestClientID_match_one(t *testing.T) {
	cid := NewClientID()
	cid.Hit(net.ParseIP("127.0.0.1"), "/")
	if !cid.Match(net.ParseIP("127.0.0.1"), []string{"/"}) {
		t.Fail()
	}
}

func TestClientID_match_one_fail(t *testing.T) {
	cid := NewClientID()
	cid.Hit(net.ParseIP("127.0.0.1"), "/test")
	if cid.Match(net.ParseIP("127.0.0.1"), []string{"/"}) {
		t.Fail()
	}
}

func TestClientID_match_not_in_list(t *testing.T) {
	cid := NewClientID()
	if cid.Match(net.ParseIP("127.0.0.1"), []string{"/"}) {
		t.Fail()
	}
}

func TestClientID_match_mod(t *testing.T) {
	cid := NewClientID()
	cid.Hit(net.ParseIP("127.0.0.1"), "/")
	if cid.Match(net.ParseIP("127.0.0.1"), []string{"/", "/test"}) {
		t.Fail()
	}
}

func TestClientID_match_three(t *testing.T) {
	cid := NewClientID()
	cid.Hit(net.ParseIP("127.0.0.1"), "/one")
	cid.Hit(net.ParseIP("127.0.0.1"), "/two")
	cid.Hit(net.ParseIP("127.0.0.1"), "/three")
	if !cid.Match(net.ParseIP("127.0.0.1"), []string{"/one", "/two", "/three"}) {
		t.Fail()
	}
}

func TestClientID_match_three_mod(t *testing.T) {
	cid := NewClientID()
	cid.Hit(net.ParseIP("127.0.0.1"), "/one")
	cid.Hit(net.ParseIP("127.0.0.1"), "/one")
	cid.Hit(net.ParseIP("127.0.0.1"), "/two")
	cid.Hit(net.ParseIP("127.0.0.1"), "/three")
	if !cid.Match(net.ParseIP("127.0.0.1"), []string{"/one", "/two", "/three"}) {
		t.Fail()
	}
}

func TestClientID_match_four_mod_fail(t *testing.T) {
	cid := NewClientID()
	cid.Hit(net.ParseIP("127.0.0.1"), "/one")
	cid.Hit(net.ParseIP("127.0.0.1"), "/one")
	cid.Hit(net.ParseIP("127.0.0.1"), "/two")
	cid.Hit(net.ParseIP("127.0.0.1"), "/three")
	cid.Hit(net.ParseIP("127.0.0.1"), "/one")
	if cid.Match(net.ParseIP("127.0.0.1"), []string{"/one", "/two", "/three"}) {
		t.Fail()
	}
}
