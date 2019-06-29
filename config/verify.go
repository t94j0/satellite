package config

import (
	"errors"

	"github.com/spf13/viper"
)

// ErrNotFoundConfig is an error when both redirection options are present
var ErrNotFoundConfig = errors.New("both not_found redirect and render cannot be set at the same time")

func verify(cfg *viper.Viper) error {
	// both not_found options cannot be true at the same time
	if cfg.IsSet("not_found.render") && cfg.IsSet("not_found.redirect") {
		return ErrNotFoundConfig
	}
	return nil
}
