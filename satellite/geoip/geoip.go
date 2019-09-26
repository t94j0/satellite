package geoip

import (
	"net"
	"os"

	gip "github.com/oschwald/geoip2-golang"
)

// DB holds the DB reader
type DB struct {
	db *gip.Reader
}

// New creates a new DB reader based on the mmdb path
func New(dbpath string) (DB, error) {
	if _, err := os.Stat(dbpath); os.IsNotExist(err) {
		return DB{}, os.ErrNotExist
	}

	var geoip DB
	db, err := gip.Open(dbpath)
	if err != nil {
		return DB{}, err
	}
	geoip.db = db
	return geoip, nil
}

// HasDB returns true when the DB was configured properly
func (g DB) HasDB() bool {
	return g.db != nil
}

// CountryCode returns the ISO country code of the target IP
func (g DB) CountryCode(ip net.IP) (string, error) {
	c, err := g.db.Country(ip)
	if err != nil {
		return "", err
	}

	return c.Country.IsoCode, nil
}
