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
	config.AddConfigPath("$HOME/" + ProjectName)
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

	paths, err = NewPaths(serverPath)
	if err != nil {
		log.Fatal(err)
	}

	go createWatcher(serverPath, func() error {
		return paths.Reload()
	})

	server := NewServer(listen, certPath, keyPath)
	log.Printf("Listening on port %s", config.GetString("listen"))
	if err := server.Start(); err != nil {
		log.Fatal(err)
	}
}
