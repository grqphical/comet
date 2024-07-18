package config

import (
	"errors"

	"github.com/spf13/viper"
)

type Backend struct {
	Type        string `mapstructure:"type"`
	RouteFilter string `mapstructure:"route_filter"`
	StripFilter bool   `mapstructure:"strip_filter"`
	// These fields should only be populated if the type is "proxy"
	Address        string   `mapstructure:"address"`
	HealthEndpoint string   `mapstructure:"health_endpoint"`
	CheckHealth    bool     `mapstructure:"check_health"`
	HiddenRoutes   []string `mapstructure:"hidden_routes"`
	// These fields should only be populated if the type is "staticfs"
	Directory string `mapstructure:"directory"`
}

var Backends []Backend

func ReadConfig() error {
	viper.SetConfigName("comet")
	viper.SetConfigType("toml")
	viper.AddConfigPath(".")
	err := viper.ReadInConfig()
	if err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {
			return errors.New("no comet.toml file found")
		} else {
			return err
		}
	}

	err = viper.UnmarshalKey("backend", &Backends)
	if err != nil {
		return err
	}

	return nil
}
