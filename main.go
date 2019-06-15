package main

import (
	"log"

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
	keyPath := config.GetString("ssl.key")
	certPath := config.GetString("ssl.cert")
	serverHeader := config.GetString("server_header")
	managementIP := config.GetString("management.ip")
	managementPath := config.GetString("management.path")
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

	server := NewServer(listen, certPath, keyPath, serverHeader, managementIP, managementPath, indexPath)
	log.Printf("Listening on port %s", config.GetString("listen"))
	if err := server.Start(); err != nil {
		log.Fatal(err)
	}
}
