package path

import (
	"bytes"
	"encoding/binary"
	"net"

	_ "github.com/mattn/go-sqlite3"
	"github.com/pkg/errors"
	"github.com/prologic/bitcask"
	"github.com/t94j0/satellite/net/http"
)

// State contains all state for Paths configuration
type State struct {
	db *bitcask.Bitcask
	// PathIdentifier is the global ClientID
	pathIdentifier *ClientID
}

// NewState creates the prereqs for managing state in Satellite
func NewState(dbPath string) (*State, error) {
	state := &State{db: nil, pathIdentifier: NewClientID()}

	database, err := bitcask.Open(dbPath)
	if err != nil {
		return nil, err
	}
	state.db = database

	return state, nil
}

// exists checks if a path exists in the db
func (s *State) exists(path string) bool {
	return s.db.Has([]byte(path))
}

// createPath creates a path in the database
func (s *State) create(path string) error {
	buf := make([]byte, binary.MaxVarintLen64)
	binary.PutUvarint(buf, 1)
	return s.db.Put([]byte(path), buf)
}

// incrementServed increments the times_served for a path
func (s *State) incrementServed(path string) error {
	n, err := s.db.Get([]byte(path))
	if err != nil {
		return err
	}

	timesRead, err := binary.ReadUvarint(bytes.NewBuffer(n))
	if err != nil {
		return err
	}

	newTimesRead := make([]byte, binary.MaxVarintLen64)
	binary.PutUvarint(newTimesRead, timesRead+1)
	return s.db.Put([]byte(path), newTimesRead)
}

// ErrNoURL is returned when a request has no URL in the request
var ErrNoURL = errors.New("No URL for request")

// Hit will create a path in the DB if it does not exist, and increment the times_served if it does exist
func (s *State) Hit(req *http.Request) error {
	if req.URL == nil {
		return ErrNoURL
	}
	path := req.URL.Path

	// ClientID Hit
	remoteAddr := net.ParseIP(req.RemoteAddr)
	s.pathIdentifier.Hit(remoteAddr, path)

	// DB Hit
	if exists := s.exists(path); !exists {
		return s.create(path)
	}

	return s.incrementServed(path)
}

// GetHits gets the number of hits for a target path using the SQLite DB
func (s *State) GetHits(path string) (uint64, error) {
	if !s.exists(path) {
		return 0, nil
	}

	n, err := s.db.Get([]byte(path))
	if err != nil {
		return 0, err
	}

	timesRead, err := binary.ReadUvarint(bytes.NewBuffer(n))
	if err != nil {
		return 0, err
	}
	return timesRead, nil
}

// MatchPaths checks if an IP has hit the specified paths in order to make sure an IP can access a page
func (s *State) MatchPaths(ip net.IP, paths []string) bool {
	return s.pathIdentifier.Match(ip, paths)
}

// Remove removes path from DB
func (s *State) Remove(path string) error {
	return s.db.Delete([]byte(path))
}
