package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"
)

type Config struct {
	// Service configuration
	ServiceHost string
	ServicePort string

	// Database configuration
	DatabaseURL      string
	DatabaseMaxConns int

	// Redis configuration
	RedisURL      string
	RedisMaxConns int

	// Logging configuration
	LogLevel string

	// Metrics configuration
	MetricsPort string

	// JWT configuration
	JWTPublicKeyPath string
}

// Load loads configuration from environment variables
func Load() (*Config, error) {
	cfg := &Config{}

	// Service configuration
	cfg.ServiceHost = os.Getenv("INVENTORY_SERVICE_HOST")
	if cfg.ServiceHost == "" {
		cfg.ServiceHost = "0.0.0.0"
	}

	cfg.ServicePort = os.Getenv("INVENTORY_SERVICE_PORT")
	if cfg.ServicePort == "" {
		cfg.ServicePort = "8080"
	}

	// Database configuration
	cfg.DatabaseURL = os.Getenv("DATABASE_URL")
	if cfg.DatabaseURL == "" {
		return nil, fmt.Errorf("DATABASE_URL is required")
	}

	maxConnStr := os.Getenv("DATABASE_MAX_CONNECTIONS")
	if maxConnStr == "" {
		cfg.DatabaseMaxConns = 10
	} else {
		maxConns, err := strconv.Atoi(maxConnStr)
		if err != nil {
			return nil, fmt.Errorf("invalid DATABASE_MAX_CONNECTIONS: %v", err)
		}
		cfg.DatabaseMaxConns = maxConns
	}

	// Redis configuration
	cfg.RedisURL = os.Getenv("REDIS_URL")
	if cfg.RedisURL == "" {
		return nil, fmt.Errorf("REDIS_URL is required")
	}

	redisMaxConnStr := os.Getenv("REDIS_MAX_CONNECTIONS")
	if redisMaxConnStr == "" {
		cfg.RedisMaxConns = 10
	} else {
		redisMaxConns, err := strconv.Atoi(redisMaxConnStr)
		if err != nil {
			return nil, fmt.Errorf("invalid REDIS_MAX_CONNECTIONS: %v", err)
		}
		cfg.RedisMaxConns = redisMaxConns
	}

	// Logging configuration
	cfg.LogLevel = os.Getenv("LOG_LEVEL")
	if cfg.LogLevel == "" {
		cfg.LogLevel = "info"
	}

	// Metrics configuration
	cfg.MetricsPort = os.Getenv("METRICS_PORT")
	if cfg.MetricsPort == "" {
		cfg.MetricsPort = "9090"
	}

	// JWT configuration
	cfg.JWTPublicKeyPath = os.Getenv("JWT_PUBLIC_KEY_PATH")
	if cfg.JWTPublicKeyPath == "" {
		cfg.JWTPublicKeyPath = "/etc/auth/public_key.pem"
	}

	return cfg, nil
}

// Validate performs validation on the configuration
func (c *Config) Validate() error {
	// Validate database URL format
	if !strings.HasPrefix(c.DatabaseURL, "postgresql://") && !strings.HasPrefix(c.DatabaseURL, "postgres://") {
		return fmt.Errorf("DATABASE_URL must start with postgresql:// or postgres://")
	}

	// Validate Redis URL format
	if !strings.HasPrefix(c.RedisURL, "redis://") {
		return fmt.Errorf("REDIS_URL must start with redis://")
	}

	// Validate numeric ranges
	if c.DatabaseMaxConns < 1 || c.DatabaseMaxConns > 100 {
		return fmt.Errorf("DATABASE_MAX_CONNECTIONS must be between 1 and 100")
	}

	if c.RedisMaxConns < 1 || c.RedisMaxConns > 100 {
		return fmt.Errorf("REDIS_MAX_CONNECTIONS must be between 1 and 100")
	}

	// Validate log level
	validLevels := []string{"debug", "info", "warn", "error"}
	validLevel := false
	for _, level := range validLevels {
		if c.LogLevel == level {
			validLevel = true
			break
		}
	}
	if !validLevel {
		return fmt.Errorf("LOG_LEVEL must be one of: %s", strings.Join(validLevels, ", "))
	}

	return nil
}

// String returns a string representation of the config (for logging, without sensitive data)
func (c *Config) String() string {
	return fmt.Sprintf(
		"Config{Host: %s, Port: %s, LogLevel: %s, MetricsPort: %s, DB: %s, Redis: %s}",
		c.ServiceHost, c.ServicePort, c.LogLevel, c.MetricsPort,
		maskURL(c.DatabaseURL), maskURL(c.RedisURL),
	)
}

// maskURL masks sensitive information in URLs
func maskURL(url string) string {
	if strings.Contains(url, "@") {
		parts := strings.Split(url, "@")
		if len(parts) >= 2 {
			return parts[0][:strings.Index(parts[0], "://")+3] + "***@" + parts[1]
		}
	}
	return url
}