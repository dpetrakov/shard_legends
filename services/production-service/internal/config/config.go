package config

import (
	"fmt"
	"strings"
	"time"

	"github.com/spf13/viper"
)

// Config represents the complete application configuration
type Config struct {
	Server           ServerConfig           `mapstructure:"server"`
	Database         DatabaseConfig         `mapstructure:"database"`
	Redis            RedisConfig           `mapstructure:"redis"`
	Auth             AuthConfig            `mapstructure:"auth"`
	Logging          LoggingConfig         `mapstructure:"logging"`
	ExternalServices ExternalServicesConfig `mapstructure:"external_services"`
	Timeouts         TimeoutsConfig        `mapstructure:"timeouts"`
	Cleanup          CleanupConfig         `mapstructure:"cleanup"`
	Metrics          MetricsConfig         `mapstructure:"metrics"`
}

// ServerConfig contains HTTP server configuration
type ServerConfig struct {
	Host         string        `mapstructure:"host"`
	Port         string        `mapstructure:"port"`
	InternalPort string        `mapstructure:"internal_port"`
	ReadTimeout  time.Duration `mapstructure:"read_timeout"`
	WriteTimeout time.Duration `mapstructure:"write_timeout"`
	IdleTimeout  time.Duration `mapstructure:"idle_timeout"`
}

// DatabaseConfig contains database connection configuration
type DatabaseConfig struct {
	URL               string        `mapstructure:"url"`
	MaxConnections    int           `mapstructure:"max_connections"`
	MaxIdleTime       time.Duration `mapstructure:"max_idle_time"`
	HealthCheckPeriod time.Duration `mapstructure:"health_check_period"`
	PingTimeout       time.Duration `mapstructure:"ping_timeout"`
}

// RedisConfig contains Redis connection configuration
type RedisConfig struct {
	URL            string        `mapstructure:"url"`
	AuthURL        string        `mapstructure:"auth_url"`
	MaxConnections int           `mapstructure:"max_connections"`
	ReadTimeout    time.Duration `mapstructure:"read_timeout"`
	WriteTimeout   time.Duration `mapstructure:"write_timeout"`
	MaxRetries     int           `mapstructure:"max_retries"`
	PingTimeout    time.Duration `mapstructure:"ping_timeout"`
}

// AuthConfig contains authentication configuration
type AuthConfig struct {
	PublicKeyURL   string        `mapstructure:"public_key_url"`
	CacheTTL       time.Duration `mapstructure:"cache_ttl"`
	RefreshInterval time.Duration `mapstructure:"refresh_interval"`
}

// LoggingConfig contains logging configuration
type LoggingConfig struct {
	Level string `mapstructure:"level"`
}

// ExternalServicesConfig contains external service URLs and configuration
type ExternalServicesConfig struct {
	InventoryService ExternalServiceConfig `mapstructure:"inventory_service"`
	UserService      ExternalServiceConfig `mapstructure:"user_service"`
}

// ExternalServiceConfig contains configuration for an external service
type ExternalServiceConfig struct {
	BaseURL string        `mapstructure:"base_url"`
	Timeout time.Duration `mapstructure:"timeout"`
}

// TimeoutsConfig contains various timeout configurations
type TimeoutsConfig struct {
	HTTPMiddleware     time.Duration `mapstructure:"http_middleware"`
	JWTValidatorClient time.Duration `mapstructure:"jwt_validator_client"`
	GracefulShutdown   time.Duration `mapstructure:"graceful_shutdown"`
	DatabaseHealth     time.Duration `mapstructure:"database_health"`
	RedisHealth        time.Duration `mapstructure:"redis_health"`
}

// CleanupConfig contains configuration for background cleanup processes
type CleanupConfig struct {
	OrphanedTaskTimeout time.Duration `mapstructure:"orphaned_task_timeout"`
	CleanupInterval     time.Duration `mapstructure:"cleanup_interval"`
	CleanupTimeout      time.Duration `mapstructure:"cleanup_timeout"`
}

// MetricsConfig contains metrics collection configuration
type MetricsConfig struct {
	UpdateInterval time.Duration `mapstructure:"update_interval"`
}

// Load loads configuration from file and environment variables
func Load() (*Config, error) {
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath(".")
	viper.AddConfigPath("./configs")
	viper.AddConfigPath("/etc/production-service")

	// Set environment variable prefix and key replacement
	viper.SetEnvPrefix("PROD_SVC")
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	viper.AutomaticEnv()
	
	// Explicitly bind environment variables for better reliability
	viper.BindEnv("database.url", "PROD_SVC_DATABASE_URL")
	viper.BindEnv("redis.url", "PROD_SVC_REDIS_URL")
	viper.BindEnv("redis.auth_url", "PROD_SVC_REDIS_AUTH_URL")
	viper.BindEnv("server.port", "PROD_SVC_SERVER_PORT")
	viper.BindEnv("server.internal_port", "PROD_SVC_SERVER_INTERNAL_PORT")
	viper.BindEnv("auth.public_key_url", "PROD_SVC_AUTH_PUBLIC_KEY_URL")
	viper.BindEnv("external_services.inventory_service.base_url", "PROD_SVC_EXTERNAL_SERVICES_INVENTORY_SERVICE_BASE_URL")
	viper.BindEnv("external_services.user_service.base_url", "PROD_SVC_EXTERNAL_SERVICES_USER_SERVICE_BASE_URL")

	// Set defaults
	setDefaults()

	// Try to read config file (optional)
	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			return nil, fmt.Errorf("failed to read config file: %w", err)
		}
		// Config file not found is OK, we'll use defaults + env vars
	}

	var cfg Config
	if err := viper.Unmarshal(&cfg); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config: %w", err)
	}

	if err := cfg.Validate(); err != nil {
		return nil, fmt.Errorf("config validation failed: %w", err)
	}

	return &cfg, nil
}

// setDefaults sets default values for configuration
func setDefaults() {
	// Server defaults
	viper.SetDefault("server.host", "0.0.0.0")
	viper.SetDefault("server.read_timeout", "15s")
	viper.SetDefault("server.write_timeout", "15s")
	viper.SetDefault("server.idle_timeout", "60s")

	// Database defaults (conservative values, no unsafe defaults)
	viper.SetDefault("database.max_connections", 25)
	viper.SetDefault("database.max_idle_time", "5m")
	viper.SetDefault("database.health_check_period", "1m")
	viper.SetDefault("database.ping_timeout", "5s")

	// Redis defaults
	viper.SetDefault("redis.max_connections", 10)
	viper.SetDefault("redis.read_timeout", "3s")
	viper.SetDefault("redis.write_timeout", "3s")
	viper.SetDefault("redis.max_retries", 3)
	viper.SetDefault("redis.ping_timeout", "5s")

	// Auth defaults (no unsafe URL defaults)
	viper.SetDefault("auth.cache_ttl", "1h")
	viper.SetDefault("auth.refresh_interval", "24h")

	// Logging defaults
	viper.SetDefault("logging.level", "info")

	// External services defaults (timeouts only, no URL defaults)
	viper.SetDefault("external_services.inventory_service.timeout", "10s")
	viper.SetDefault("external_services.user_service.timeout", "5s")

	// Timeout defaults
	viper.SetDefault("timeouts.http_middleware", "60s")
	viper.SetDefault("timeouts.jwt_validator_client", "10s")
	viper.SetDefault("timeouts.graceful_shutdown", "30s")
	viper.SetDefault("timeouts.database_health", "2s")
	viper.SetDefault("timeouts.redis_health", "2s")

	// Cleanup defaults
	viper.SetDefault("cleanup.orphaned_task_timeout", "5m")
	viper.SetDefault("cleanup.cleanup_interval", "5m")
	viper.SetDefault("cleanup.cleanup_timeout", "5m")

	// Metrics defaults
	viper.SetDefault("metrics.update_interval", "10s")
}

// Validate validates the configuration and ensures required fields are present
func (c *Config) Validate() error {
	// Required fields that must be set via environment variables
	requiredFields := map[string]string{
		"database.url":                          "PROD_SVC_DATABASE_URL",
		"redis.url":                            "PROD_SVC_REDIS_URL",
		"redis.auth_url":                       "PROD_SVC_REDIS_AUTH_URL",
		"server.port":                          "PROD_SVC_SERVER_PORT",
		"server.internal_port":                 "PROD_SVC_SERVER_INTERNAL_PORT",
		"auth.public_key_url":                  "PROD_SVC_AUTH_PUBLIC_KEY_URL",
		"external_services.inventory_service.base_url": "PROD_SVC_EXTERNAL_SERVICES_INVENTORY_SERVICE_BASE_URL",
		"external_services.user_service.base_url":      "PROD_SVC_EXTERNAL_SERVICES_USER_SERVICE_BASE_URL",
	}

	for field, envVar := range requiredFields {
		if !viper.IsSet(field) {
			return fmt.Errorf("required configuration field '%s' is not set (use environment variable %s)", field, envVar)
		}
		
		// Check if value is empty string
		if viper.GetString(field) == "" {
			return fmt.Errorf("required configuration field '%s' cannot be empty (set environment variable %s)", field, envVar)
		}
	}

	// Validate timeout values are reasonable
	timeouts := map[string]time.Duration{
		"server.read_timeout":    c.Server.ReadTimeout,
		"server.write_timeout":   c.Server.WriteTimeout,
		"database.ping_timeout":  c.Database.PingTimeout,
		"redis.ping_timeout":     c.Redis.PingTimeout,
		"cleanup.cleanup_interval": c.Cleanup.CleanupInterval,
	}

	for name, timeout := range timeouts {
		if timeout <= 0 {
			return fmt.Errorf("timeout '%s' must be positive, got %v", name, timeout)
		}
		if timeout > 10*time.Minute {
			return fmt.Errorf("timeout '%s' seems too large, got %v", name, timeout)
		}
	}

	// Validate numeric values
	if c.Database.MaxConnections <= 0 {
		return fmt.Errorf("database.max_connections must be positive, got %d", c.Database.MaxConnections)
	}
	if c.Redis.MaxConnections <= 0 {
		return fmt.Errorf("redis.max_connections must be positive, got %d", c.Redis.MaxConnections)
	}
	if c.Redis.MaxRetries < 0 {
		return fmt.Errorf("redis.max_retries cannot be negative, got %d", c.Redis.MaxRetries)
	}

	return nil
}

// GetCleanupConfig returns cleanup configuration for the cleanup service
func (c *Config) GetCleanupConfig() CleanupConfig {
	return c.Cleanup
}