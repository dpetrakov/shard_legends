package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"
)

type Config struct {
	// Service configuration
	ServiceHost         string
	ServicePort         string
	InternalServicePort string

	// Database configuration
	DatabaseURL      string
	DatabaseMaxConns int

	// Redis configuration (only for JWT auth)
	RedisAuthURL  string
	RedisMaxConns int

	// External services
	ProductionInternalURL string
	ProductionExternalURL string
	InventoryInternalURL  string

	// Business logic configuration
	CooldownSec        int
	DailyChestRecipeID string

	// JWT configuration
	AuthPublicKeyURL string

	// Logging configuration
	LogLevel string
}

// Load loads configuration from environment variables
func Load() (*Config, error) {
	cfg := &Config{}

	// Service configuration
	cfg.ServicePort = getEnvOrDefault("PORT_PUBLIC", "8080")
	cfg.InternalServicePort = getEnvOrDefault("PORT_INTERNAL", "8090")
	cfg.ServiceHost = "0.0.0.0"

	// Database configuration
	cfg.DatabaseURL = os.Getenv("DATABASE_URL")
	if cfg.DatabaseURL == "" {
		return nil, fmt.Errorf("DATABASE_URL is required")
	}

	maxConnStr := getEnvOrDefault("DATABASE_MAX_CONNECTIONS", "10")
	maxConns, err := strconv.Atoi(maxConnStr)
	if err != nil {
		return nil, fmt.Errorf("invalid DATABASE_MAX_CONNECTIONS: %v", err)
	}
	cfg.DatabaseMaxConns = maxConns

	// Redis configuration (only for JWT auth)
	cfg.RedisAuthURL = getEnvOrDefault("REDIS_AUTH_URL", "redis://redis:6379/0")

	redisMaxConnStr := getEnvOrDefault("REDIS_MAX_CONNECTIONS", "10")
	redisMaxConns, err := strconv.Atoi(redisMaxConnStr)
	if err != nil {
		return nil, fmt.Errorf("invalid REDIS_MAX_CONNECTIONS: %v", err)
	}
	cfg.RedisMaxConns = redisMaxConns

	// External services
	cfg.ProductionInternalURL = getEnvOrDefault("PRODUCTION_INTERNAL_URL", "http://production-service:8090")
	cfg.ProductionExternalURL = getEnvOrDefault("PRODUCTION_EXTERNAL_URL", "http://production-service:8080")
	cfg.InventoryInternalURL = getEnvOrDefault("INVENTORY_INTERNAL_URL", "http://inventory-service:8080")

	// Business logic configuration
	cooldownStr := getEnvOrDefault("COOLDOWN_SEC", "30")
	cooldown, err := strconv.Atoi(cooldownStr)
	if err != nil {
		return nil, fmt.Errorf("invalid COOLDOWN_SEC: %v", err)
	}
	cfg.CooldownSec = cooldown

	cfg.DailyChestRecipeID = getEnvOrDefault("DAILY_CHEST_RECIPE_ID", "9b9a4a62-7e79-4f1c-8dbe-62784be4c9d2")

	// JWT configuration
	cfg.AuthPublicKeyURL = getEnvOrDefault("AUTH_PUBLIC_KEY_URL", "http://auth-service:8090/public-key.pem")

	// Logging configuration
	cfg.LogLevel = getEnvOrDefault("LOG_LEVEL", "info")

	return cfg, nil
}

// Validate performs validation on the configuration
func (c *Config) Validate() error {
	// Validate database URL format
	if !strings.HasPrefix(c.DatabaseURL, "postgresql://") && !strings.HasPrefix(c.DatabaseURL, "postgres://") {
		return fmt.Errorf("DATABASE_URL must start with postgresql:// or postgres://")
	}

	// Validate Redis Auth URL format
	if !strings.HasPrefix(c.RedisAuthURL, "redis://") {
		return fmt.Errorf("REDIS_AUTH_URL must start with redis://")
	}

	// Validate numeric ranges
	if c.DatabaseMaxConns < 1 || c.DatabaseMaxConns > 100 {
		return fmt.Errorf("DATABASE_MAX_CONNECTIONS must be between 1 and 100")
	}

	if c.RedisMaxConns < 1 || c.RedisMaxConns > 100 {
		return fmt.Errorf("REDIS_MAX_CONNECTIONS must be between 1 and 100")
	}

	if c.CooldownSec < 0 || c.CooldownSec > 3600 {
		return fmt.Errorf("COOLDOWN_SEC must be between 0 and 3600")
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
		"Config{Host: %s, Port: %s, InternalPort: %s, LogLevel: %s, CooldownSec: %d, RecipeID: %s, DB: %s, RedisAuth: %s}",
		c.ServiceHost, c.ServicePort, c.InternalServicePort, c.LogLevel, c.CooldownSec, c.DailyChestRecipeID,
		maskURL(c.DatabaseURL), maskURL(c.RedisAuthURL),
	)
}

// Helper functions

func getEnvOrDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
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
