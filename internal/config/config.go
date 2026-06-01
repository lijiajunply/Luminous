package config

import (
	"errors"
	"fmt"

	"github.com/spf13/viper"
)

type ServerConfig struct {
	Port int    `mapstructure:"port"`
	Mode string `mapstructure:"mode"`
}

type AuthConfig struct {
	AdminToken string `mapstructure:"admin_token"`
}

type DataConfig struct {
	SchoolsFile string `mapstructure:"schools_file"`
}

type AppConfig struct {
	Server ServerConfig `mapstructure:"server"`
	Auth   AuthConfig   `mapstructure:"auth"`
	Data   DataConfig   `mapstructure:"data"`
}

var Cfg *AppConfig

func LoadConfig() error {
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath(".")

	viper.SetEnvPrefix("LUMINOUS")
	viper.AutomaticEnv()

	viper.BindEnv("server.port")
	viper.BindEnv("server.mode")
	viper.BindEnv("auth.admin_token")
	viper.BindEnv("data.schools_file")

	viper.SetDefault("server.port", 8080)
	viper.SetDefault("server.mode", "debug")
	viper.SetDefault("data.schools_file", "./data/schools.json")

	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			return fmt.Errorf("read config file: %w", err)
		}
	}

	Cfg = &AppConfig{}
	if err := viper.Unmarshal(Cfg); err != nil {
		return fmt.Errorf("unmarshal config: %w", err)
	}

	if Cfg.Auth.AdminToken == "" {
		return errors.New("auth.admin_token must not be empty; set it in config.yaml or via LUMINOUS_AUTH_ADMIN_TOKEN env var")
	}

	return nil
}
