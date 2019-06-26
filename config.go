package main

import "github.com/spf13/viper"

// Config sets the config
func Config() (*viper.Viper, error) {
	config := viper.New()

	config.SetDefault("server_root", "/var/www/html")
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
