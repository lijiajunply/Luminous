package config

import (
	"bufio"
	"errors"
	"fmt"
	"os"
	"strconv"
	"strings"
)

type ServerConfig struct {
	Port          int
	Mode          string
	CORSOrigin    string
	TLSCert       string
	TLSKey        string
	TrustedProxies string
}

type AuthConfig struct {
	AdminToken string
}

type DatabaseConfig struct {
	DSN          string
	PoolMaxConns int32
	PoolMinConns int32
}

type ReleaseConfig struct {
	APIURL    string
	AppUUID   string
	ChannelID string
}

type RateLimitConfig struct {
	Rate  int
	Burst int
}

type AppConfig struct {
	Server    ServerConfig
	Auth      AuthConfig
	Database  DatabaseConfig
	Release   ReleaseConfig
	RateLimit RateLimitConfig
}

var Cfg *AppConfig

func loadEnvFile(path string) error {
	f, err := os.Open(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return fmt.Errorf("open .env file: %w", err)
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		parts := strings.SplitN(line, "=", 2)
		if len(parts) != 2 {
			continue
		}
		key := strings.TrimSpace(parts[0])
		value := strings.TrimSpace(parts[1])
		value = strings.Trim(value, `"'`)
		if _, exists := os.LookupEnv(key); exists {
			continue
		}
		os.Setenv(key, value)
	}
	return scanner.Err()
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func getEnvInt(key string, fallback int) int {
	if v := os.Getenv(key); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			return n
		}
	}
	return fallback
}

func LoadConfig() error {
	if err := loadEnvFile(".env"); err != nil {
		return fmt.Errorf("load .env file: %w", err)
	}

	Cfg = &AppConfig{
		Server: ServerConfig{
			Port:          getEnvInt("LUMINOUS_SERVER_PORT", 8080),
			Mode:          getEnv("LUMINOUS_SERVER_MODE", "release"),
			CORSOrigin:    getEnv("LUMINOUS_SERVER_CORS_ORIGIN", ""),
			TLSCert:       getEnv("LUMINOUS_SERVER_TLS_CERT", ""),
			TLSKey:        getEnv("LUMINOUS_SERVER_TLS_KEY", ""),
			TrustedProxies: getEnv("LUMINOUS_SERVER_TRUSTED_PROXIES", ""),
		},
		Auth: AuthConfig{
			AdminToken: getEnv("LUMINOUS_AUTH_ADMIN_TOKEN", ""),
		},
		Database: DatabaseConfig{
			DSN:          getEnv("LUMINOUS_DATABASE_DSN", ""),
			PoolMaxConns: int32(getEnvInt("LUMINOUS_DATABASE_POOL_MAX_CONNS", 20)),
			PoolMinConns: int32(getEnvInt("LUMINOUS_DATABASE_POOL_MIN_CONNS", 5)),
		},
		Release: ReleaseConfig{
			APIURL:    getEnv("LUMINOUS_RELEASE_API_URL", ""),
			AppUUID:   getEnv("LUMINOUS_RELEASE_APP_UUID", "5f278ffc-5a70-4805-a6bf-0543040981a8"),
			ChannelID: getEnv("LUMINOUS_RELEASE_CHANNEL_ID", "9e1a198a-a0c2-4017-b492-f2d0e5bee437"),
		},
		RateLimit: RateLimitConfig{
			Rate:  getEnvInt("LUMINOUS_RATE_LIMIT_RATE", 10),
			Burst: getEnvInt("LUMINOUS_RATE_LIMIT_BURST", 30),
		},
	}

	if Cfg.Auth.AdminToken == "" {
		return errors.New("LUMINOUS_AUTH_ADMIN_TOKEN is required")
	}

	return nil
}
