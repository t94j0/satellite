package geoip_test

import (
	"errors"
	"fmt"
	"net"
	"os"
	"path/filepath"
	"testing"

	. "github.com/t94j0/satellite/satellite/geoip"
)

func createGeoIP() (DB, error) {
	wd, err := os.Getwd()
	if err != nil {
		return DB{}, err
	}

	fp := filepath.Join(wd, "..", "..", ".config", "var", "lib", "satellite", "GeoLite2-Country.mmdb")

	gip, err := New(fp)
	if err != nil {
		return DB{}, err
	}

	return gip, nil
}

func testIPCountry(ip, targetCountry string) error {
	gip, err := createGeoIP()
	if err != nil {
		return err
	}

	country, err := gip.CountryCode(net.ParseIP(ip))
	if err != nil {
		return err
	}

	if targetCountry != country {
		return errors.New(fmt.Sprintf("Address %s was said to be in %s", ip, country))
	}

	return nil
}

func TestNew(t *testing.T) {
	if _, err := createGeoIP(); err != nil {
		t.Error(err)
	}
}

func TestNew_CountryCode(t *testing.T) {
	if err := testIPCountry("104.222.16.238", "US"); err != nil {
		t.Error(err)
	}
}
