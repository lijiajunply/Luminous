package config

import (
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

type XAUATConfig struct {
	BaseURL            string `mapstructure:"base_url"`
	LoginURL           string `mapstructure:"login_url"`
	OldBusURL          string `mapstructure:"old_bus_url"`
	NewBusURL          string `mapstructure:"new_bus_url"`
	SemesterStart      string `mapstructure:"semester_start"`
	SemesterEnd        string `mapstructure:"semester_end"`
	PaymentOAuthSecret string `mapstructure:"payment_oauth_secret"`
}

type SchoolsConfig struct {
	XAUAT XAUATConfig `mapstructure:"xauat"`
}

type AppConfig struct {
	Server  ServerConfig  `mapstructure:"server"`
	Auth    AuthConfig    `mapstructure:"auth"`
	Data    DataConfig    `mapstructure:"data"`
	Schools SchoolsConfig `mapstructure:"schools"`
}

var Cfg *AppConfig

func LoadConfig() error {
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath(".")

	viper.SetEnvPrefix("LUMINOUS")
	viper.AutomaticEnv()

	viper.SetDefault("server.port", 8080)
	viper.SetDefault("server.mode", "debug")
	viper.SetDefault("auth.admin_token", "")
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

	return nil
}
