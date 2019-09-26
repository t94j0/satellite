package server

import (
	"os"

	"github.com/pkg/errors"
	"github.com/t94j0/satellite/crypto/tls"
)

type SSL struct {
	keyPath  string
	certPath string
}

func NewSSL(keyPath, certPath string) (SSL, error) {
	if _, err := os.Stat(keyPath); os.IsNotExist(err) {
		return SSL{}, errors.Wrap(err, "SSL key not found")
	}
	if _, err := os.Stat(certPath); os.IsNotExist(err) {
		return SSL{}, errors.Wrap(err, "SSL cert not found")
	}

	return SSL{keyPath, certPath}, nil
}

func (s SSL) CreateTLSConfig() (*tls.Config, error) {
	cert, err := tls.LoadX509KeyPair(s.certPath, s.keyPath)
	if err != nil {
		return nil, err
	}
	tlsConfig := &tls.Config{Certificates: []tls.Certificate{cert}}
	return tlsConfig, nil
}
