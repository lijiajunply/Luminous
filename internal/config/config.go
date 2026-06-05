package config

import (
	"fmt"
	"strings"

	"github.com/spf13/viper"
)

type ServerConfig struct {
	Port       int    `mapstructure:"port"`
	Mode       string `mapstructure:"mode"`
	CORSOrigin string `mapstructure:"cors_origin"`
}

type AuthConfig struct {
	AdminToken string `mapstructure:"admin_token"`
}

type ReleaseConfig struct {
	APIURL    string `mapstructure:"api_url"`
	AppUUID   string `mapstructure:"app_uuid"`
	ChannelID string `mapstructure:"channel_id"`
}

type DataConfig struct {
	SchoolsFile string `mapstructure:"schools_file"`
}

type DatabaseConfig struct {
	DSN          string `mapstructure:"dsn"`
	Host         string `mapstructure:"host"`
	Port         int    `mapstructure:"port"`
	User         string `mapstructure:"user"`
	Password     string `mapstructure:"password"`
	DBName       string `mapstructure:"dbname"`
	SSLMode      string `mapstructure:"sslmode"`
	PoolMaxConns int32  `mapstructure:"pool_max_conns"`
	PoolMinConns int32  `mapstructure:"pool_min_conns"`
}

type AppConfig struct {
	Server   ServerConfig   `mapstructure:"server"`
	Auth     AuthConfig     `mapstructure:"auth"`
	Data     DataConfig     `mapstructure:"data"`
	Database DatabaseConfig `mapstructure:"database"`
	Release  ReleaseConfig  `mapstructure:"release"`
}

var Cfg *AppConfig

func LoadConfig() error {
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath(".")

	viper.SetEnvPrefix("LUMINOUS")
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	viper.AutomaticEnv()

	viper.SetDefault("server.port", 8080)
	viper.SetDefault("server.mode", "debug")
	viper.SetDefault("server.cors_origin", "*")
	viper.SetDefault("auth.admin_token", "")
	viper.SetDefault("data.schools_file", "./data/schools.json")
	viper.SetDefault("release.api_url", "")
	viper.SetDefault("release.app_uuid", "5f278ffc-5a70-4805-a6bf-0543040981a8")
	viper.SetDefault("release.channel_id", "9e1a198a-a0c2-4017-b492-f2d0e5bee437")

	viper.SetDefault("database.host", "localhost")
	viper.SetDefault("database.port", 5432)
	viper.SetDefault("database.user", "luminous")
	viper.SetDefault("database.password", "luminous")
	viper.SetDefault("database.dbname", "luminous")
	viper.SetDefault("database.sslmode", "disable")
	viper.SetDefault("database.pool_max_conns", 20)
	viper.SetDefault("database.pool_min_conns", 5)

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
