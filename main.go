package main

import (
	log "github.com/sirupsen/logrus"

	"github.com/pkg/errors"
)

// ProjectName is the current project name
const ProjectName string = "DEITYSHADOW"

var paths *Paths

func main() {
	config, err := Config()
	if err != nil {
		log.Fatal(err)
	}

	serverPath := config.GetString("server_path")
	listen := config.GetString("listen")
	certPath := config.GetString("ssl.cert")
	keyPath := config.GetString("ssl.key")
	serverHeader := config.GetString("server_header")
	managementIP := config.GetString("management.ip")
	managementPath := config.GetString("management.path")
	notFoundRedirect := config.GetString("not_found.redirect")
	notFoundRender := config.GetString("not_found.render")
	indexPath := config.GetString("index")
	redirectHTTP := config.GetBool("redirect_http")

	log.Infof("Using config file %s", config.ConfigFileUsed())
	log.Infof("Using server path %s", serverPath)

	paths, err = NewPaths(serverPath)
	if err != nil {
		log.Fatal(err)
	}

	if notFoundRedirect == "" && notFoundRender == "" {
		log.Warn("Use not_found handlers for opsec")
	}

	log.Printf("Loaded %d path(s)", paths.Len())

	go func() {
		if err := createWatcher(serverPath, "1s", func() error {
			return paths.Reload()
		}); err != nil {
			log.Fatal(err)
		}
	}()

	server, err := NewServer(
		serverPath,
		listen,
		keyPath,
		certPath,
		serverHeader,
		managementIP,
		managementPath,
		indexPath,
		notFoundRedirect,
		notFoundRender,
		redirectHTTP,
	)
	if err != nil {
		log.Fatal(errors.Wrap(err, "server configuration error"))
	}

	log.Infof("Listening HTTPS on port %s", config.GetString("listen"))
	if err := server.Start(); err != nil {
		log.Fatal(err)
	}
}
