package config

import (
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/joho/godotenv"
)

type Config struct {
	Server           ServerConfig
	Database         DatabaseConfig
	Redis            RedisConfig
	Auth             AuthConfig
	Logging          LoggingConfig
	Services         ServicesConfig
	ExternalServices ExternalServicesConfig
}

type ServerConfig struct {
	Host         string
	Port         string
	ReadTimeout  time.Duration
	WriteTimeout time.Duration
	IdleTimeout  time.Duration
}

type DatabaseConfig struct {
	URL            string
	MaxConnections int
	MaxIdleTime    time.Duration
}

type RedisConfig struct {
	URL            string
	AuthURL        string // Отдельная база для JWT revocation (база 0)
	MaxConnections int
	ReadTimeout    time.Duration
	WriteTimeout   time.Duration
}

type AuthConfig struct {
	PublicKeyURL string
	CacheTTL     time.Duration
}

type LoggingConfig struct {
	Level string
}

type ServicesConfig struct {
	InventoryServiceURL string
	AuthServiceURL      string
}

type ExternalServicesConfig struct {
	InventoryService ExternalServiceConfig
	UserService      ExternalServiceConfig
}

type ExternalServiceConfig struct {
	BaseURL string
}

func Load() (*Config, error) {
	_ = godotenv.Load()

	cfg := &Config{
		Server: ServerConfig{
			Host:         getEnv("PRODUCTION_SERVICE_HOST", "0.0.0.0"),
			Port:         getEnv("PRODUCTION_SERVICE_PORT", "8082"),
			ReadTimeout:  getDurationEnv("SERVER_READ_TIMEOUT", 15*time.Second),
			WriteTimeout: getDurationEnv("SERVER_WRITE_TIMEOUT", 15*time.Second),
			IdleTimeout:  getDurationEnv("SERVER_IDLE_TIMEOUT", 60*time.Second),
		},
		Database: DatabaseConfig{
			URL:            getEnv("DATABASE_URL", ""),
			MaxConnections: getIntEnv("DATABASE_MAX_CONNECTIONS", 25),
			MaxIdleTime:    getDurationEnv("DATABASE_MAX_IDLE_TIME", 5*time.Minute),
		},
		Redis: RedisConfig{
			URL:            getEnv("REDIS_URL", ""),
			AuthURL:        getEnv("REDIS_AUTH_URL", "redis://redis:6379/0"),
			MaxConnections: getIntEnv("REDIS_MAX_CONNECTIONS", 10),
			ReadTimeout:    getDurationEnv("REDIS_READ_TIMEOUT", 3*time.Second),
			WriteTimeout:   getDurationEnv("REDIS_WRITE_TIMEOUT", 3*time.Second),
		},
		Auth: AuthConfig{
			PublicKeyURL: getEnv("AUTH_PUBLIC_KEY_URL", "http://auth-service:8080/public-key.pem"),
			CacheTTL:     getDurationEnv("AUTH_CACHE_TTL", 1*time.Hour),
		},
		Logging: LoggingConfig{
			Level: getEnv("LOG_LEVEL", "info"),
		},
		Services: ServicesConfig{
			InventoryServiceURL: getEnv("INVENTORY_SERVICE_URL", "http://inventory-service:8081"),
			AuthServiceURL:      getEnv("AUTH_SERVICE_URL", "http://auth-service:8080"),
		},
		ExternalServices: ExternalServicesConfig{
			InventoryService: ExternalServiceConfig{
				BaseURL: getEnv("INVENTORY_SERVICE_URL", "http://inventory-service:8081"),
			},
			UserService: ExternalServiceConfig{
				BaseURL: getEnv("USER_SERVICE_URL", "http://user-service:8080"),
			},
		},
	}

	if err := cfg.Validate(); err != nil {
		return nil, fmt.Errorf("config validation failed: %w", err)
	}

	return cfg, nil
}

func (c *Config) Validate() error {
	if c.Database.URL == "" {
		return fmt.Errorf("DATABASE_URL is required")
	}
	if c.Redis.URL == "" {
		return fmt.Errorf("REDIS_URL is required")
	}
	return nil
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getIntEnv(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if intValue, err := strconv.Atoi(value); err == nil {
			return intValue
		}
	}
	return defaultValue
}

func getDurationEnv(key string, defaultValue time.Duration) time.Duration {
	if value := os.Getenv(key); value != "" {
		if duration, err := time.ParseDuration(value); err == nil {
			return duration
		}
	}
	return defaultValue
}
