package main

import (
	"log"

	"github.com/pkg/errors"

	"github.com/spf13/viper"
)

const ProjectName string = "DEITYSHADOW"

var paths *Paths

func Config() (*viper.Viper, error) {
	config := viper.New()

	config.SetDefault("server_path", "/var/www/html")
	config.SetDefault("listen", "127.0.0.1:8080")

	config.SetConfigName("config")
	config.AddConfigPath("$HOME/.config/" + ProjectName)
	config.AddConfigPath("$HOME/." + ProjectName)
	config.AddConfigPath("/etc/" + ProjectName)

	if err := config.ReadInConfig(); err != nil {
		return config, err
	}

	return config, nil
}

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

	log.Printf("Using config file %s", config.ConfigFileUsed())
	log.Printf("Using server path %s", serverPath)

	paths, err = NewPaths(serverPath)
	if err != nil {
		log.Fatal(err)
	}

	log.Printf("Loaded %d path(s)", paths.Len())

	go func() {
		if err := createWatcher(serverPath, "1s", func() error {
			return paths.Reload()
		}); err != nil {
			log.Fatal(err)
		}
	}()

	serverConfig := ServerConfig{
		Port:             listen,
		KeyPath:          keyPath,
		CertPath:         certPath,
		ServerHeader:     serverHeader,
		ManagementIP:     managementIP,
		ManagementPath:   managementPath,
		IndexPath:        indexPath,
		NotFoundRedirect: notFoundRedirect,
		NotFoundRender:   notFoundRender,
	}

	server, err := NewServer(serverConfig)
	if err != nil {
		log.Fatal(errors.Wrap(err, "server configuration error"))
	}

	log.Printf("Listening on port %s", config.GetString("listen"))
	if err := server.Start(); err != nil {
		log.Fatal(err)
	}
}
