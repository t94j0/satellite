package config

import "github.com/spf13/viper"

// Config sets the config
func Config(projectName string) (*viper.Viper, error) {
	config := viper.New()

	config.SetDefault("server_root", "/var/www/html")
	config.SetDefault("listen", "127.0.0.1:8080")

	config.SetConfigName("config")
	config.AddConfigPath("$HOME/.config/" + projectName)
	config.AddConfigPath("$HOME/." + projectName)
	config.AddConfigPath("/etc/" + projectName)

	if err := config.ReadInConfig(); err != nil {
		return config, err
	}

	if err := verify(config); err != nil {
		return nil, err
	}

	return config, nil
}
