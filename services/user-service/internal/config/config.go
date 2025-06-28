package config

import (
	"os"
	"strconv"
)

type Config struct {
	Server struct {
		Port         string
		InternalPort string
		Host         string
	}
	Auth struct {
		PublicKeyURL string
		RedisURL     string
	}
}

func Load() (*Config, error) {
	cfg := &Config{}

	// Настройки сервера
	cfg.Server.Port = getEnvWithDefault("USER_SERVICE_PORT", "8080")
	cfg.Server.InternalPort = getEnvWithDefault("USER_SERVICE_INTERNAL_PORT", "8090")
	cfg.Server.Host = getEnvWithDefault("USER_SERVICE_HOST", "0.0.0.0")

	// Настройки аутентификации
	cfg.Auth.PublicKeyURL = getEnvWithDefault("AUTH_SERVICE_PUBLIC_KEY_URL", "http://auth-service:8090/public-key.pem")
	cfg.Auth.RedisURL = getEnvWithDefault("REDIS_AUTH_URL", "redis://redis:6379/0")

	return cfg, nil
}

func getEnvWithDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getEnvIntWithDefault(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if intValue, err := strconv.Atoi(value); err == nil {
			return intValue
		}
	}
	return defaultValue
}