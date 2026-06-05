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
	Port int
	Mode string
}

type AuthConfig struct {
	AdminToken string
}

type DataConfig struct {
	SchoolsFile string
}

type AppConfig struct {
	Server ServerConfig
	Auth   AuthConfig
	Data   DataConfig
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
			Port: getEnvInt("LUMINOUS_SERVER_PORT", 8080),
			Mode: getEnv("LUMINOUS_SERVER_MODE", "debug"),
		},
		Auth: AuthConfig{
			AdminToken: os.Getenv("LUMINOUS_AUTH_ADMIN_TOKEN"),
		},
		Data: DataConfig{
			SchoolsFile: getEnv("LUMINOUS_DATA_SCHOOLS_FILE", "./data/schools.json"),
		},
	}

	if Cfg.Auth.AdminToken == "" {
		return errors.New("LUMINOUS_AUTH_ADMIN_TOKEN is required; set it in your environment or .env file")
	}

	return nil
}
