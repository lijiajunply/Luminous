package config

import (
	"fmt"
	"os"
	"strings"

	"github.com/spf13/viper"
)

type ServerConfig struct {
	Port          int    `mapstructure:"port"`
	Mode          string `mapstructure:"mode"`
	CORSOrigin    string `mapstructure:"cors_origin"`
	TLSCert       string `mapstructure:"tls_cert"`
	TLSKey        string `mapstructure:"tls_key"`
	TrustedProxies string `mapstructure:"trusted_proxies"`
}

type AuthConfig struct {
	AdminToken string `mapstructure:"admin_token"`
}

type ReleaseConfig struct {
	APIURL    string `mapstructure:"api_url"`
	AppUUID   string `mapstructure:"app_uuid"`
	ChannelID string `mapstructure:"channel_id"`
}

type DatabaseConfig struct {
	DSN          string `mapstructure:"dsn"`
	PoolMaxConns int32  `mapstructure:"pool_max_conns"`
	PoolMinConns int32  `mapstructure:"pool_min_conns"`
}

type RateLimitConfig struct {
	Rate  int `mapstructure:"rate"`
	Burst int `mapstructure:"burst"`
}

type AppConfig struct {
	Server    ServerConfig    `mapstructure:"server"`
	Auth      AuthConfig      `mapstructure:"auth"`
	Database  DatabaseConfig  `mapstructure:"database"`
	Release   ReleaseConfig   `mapstructure:"release"`
	RateLimit RateLimitConfig `mapstructure:"rate_limit"`
}

var Cfg *AppConfig

func LoadConfig() error {
	viper.SetEnvPrefix("LUMINOUS")
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	viper.AutomaticEnv()

	viper.SetDefault("server.port", 8080)
	viper.SetDefault("server.mode", "release")
	viper.SetDefault("server.cors_origin", "")
	viper.SetDefault("auth.admin_token", "")
	viper.SetDefault("release.api_url", "")
	viper.SetDefault("release.app_uuid", "5f278ffc-5a70-4805-a6bf-0543040981a8")
	viper.SetDefault("release.channel_id", "9e1a198a-a0c2-4017-b492-f2d0e5bee437")

	viper.SetDefault("database.pool_max_conns", 20)
	viper.SetDefault("database.pool_min_conns", 5)
	viper.SetDefault("rate_limit.rate", 10)
	viper.SetDefault("rate_limit.burst", 30)

	// Optional YAML config file for local development.
	// Set LUMINOUS_CONFIG_PATH to load a file, otherwise pure env vars.
	if p := os.Getenv("LUMINOUS_CONFIG_PATH"); p != "" {
		viper.SetConfigFile(p)
		if err := viper.ReadInConfig(); err != nil {
			return fmt.Errorf("read config file %s: %w", p, err)
		}
	}

	Cfg = &AppConfig{}
	if err := viper.Unmarshal(Cfg); err != nil {
		return fmt.Errorf("unmarshal config: %w", err)
	}

	return nil
}
