package main

import (
	"errors"

	"github.com/spf13/viper"
)

// ErrNoConfigFound is given when no configuration file is found
var ErrNoConfigFound = errors.New("Config file not found. Read the README to learn how to configure the satellite service")

// Config sets the config
func Config() (*viper.Viper, error) {
	config := viper.New()

	config.SetDefault("server_root", "/var/www/html")
	config.SetDefault("listen", "127.0.0.1:8080")

	config.SetConfigName("config")
	config.AddConfigPath("$HOME/.config/" + ProjectName)
	config.AddConfigPath("$HOME/." + ProjectName)
	config.AddConfigPath("/etc/" + ProjectName)

	err := config.ReadInConfig()
	switch err.(type) {
	case viper.ConfigFileNotFoundError:
		return nil, ErrNoConfigFound
	}
	if err != nil {
		return nil, err
	}

	return config, nil
}
