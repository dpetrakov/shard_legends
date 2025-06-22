package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"
)

type Config struct {
	// Service configuration
	ServiceHost string
	ServicePort string

	// Database configuration
	DatabaseURL      string
	DatabaseMaxConns int

	// Redis configuration
	RedisURL                string
	RedisMaxConns           int
	RedisPersistenceEnabled bool
	RedisSaveInterval       int
	RedisAOFEnabled         bool

	// JWT configuration
	JWTPrivateKeyPath string
	JWTPublicKeyPath  string
	JWTIssuer         string
	JWTExpiryHours    int

	// Telegram configuration
	TelegramBotTokens []string

	// Security configuration
	RateLimitRequests int
	RateLimitWindow   time.Duration

	// Token cleanup configuration
	TokenCleanupIntervalHours int
	TokenCleanupTimeoutMins   int
}

// Load loads configuration from environment variables
func Load() (*Config, error) {
	cfg := &Config{}

	// Service configuration
	cfg.ServiceHost = os.Getenv("AUTH_SERVICE_HOST")
	if cfg.ServiceHost == "" {
		cfg.ServiceHost = "0.0.0.0"
	}

	cfg.ServicePort = os.Getenv("AUTH_SERVICE_PORT")
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

	cfg.RedisPersistenceEnabled = strings.ToLower(os.Getenv("REDIS_PERSISTENCE_ENABLED")) == "true"

	saveIntervalStr := os.Getenv("REDIS_SAVE_INTERVAL")
	if saveIntervalStr == "" {
		cfg.RedisSaveInterval = 60
	} else {
		saveInterval, err := strconv.Atoi(saveIntervalStr)
		if err != nil {
			return nil, fmt.Errorf("invalid REDIS_SAVE_INTERVAL: %v", err)
		}
		cfg.RedisSaveInterval = saveInterval
	}

	cfg.RedisAOFEnabled = strings.ToLower(os.Getenv("REDIS_AOF_ENABLED")) == "true"

	// JWT configuration
	cfg.JWTPrivateKeyPath = os.Getenv("JWT_PRIVATE_KEY_PATH")
	if cfg.JWTPrivateKeyPath == "" {
		cfg.JWTPrivateKeyPath = "/etc/auth/private_key.pem"
	}

	cfg.JWTPublicKeyPath = os.Getenv("JWT_PUBLIC_KEY_PATH")
	if cfg.JWTPublicKeyPath == "" {
		cfg.JWTPublicKeyPath = "/etc/auth/public_key.pem"
	}

	cfg.JWTIssuer = os.Getenv("JWT_ISSUER")
	if cfg.JWTIssuer == "" {
		cfg.JWTIssuer = "shard-legends-auth"
	}

	jwtExpiryStr := os.Getenv("JWT_EXPIRY_HOURS")
	if jwtExpiryStr == "" {
		cfg.JWTExpiryHours = 24
	} else {
		jwtExpiry, err := strconv.Atoi(jwtExpiryStr)
		if err != nil {
			return nil, fmt.Errorf("invalid JWT_EXPIRY_HOURS: %v", err)
		}
		cfg.JWTExpiryHours = jwtExpiry
	}

	// Telegram configuration
	primaryToken := os.Getenv("TELEGRAM_BOT_TOKEN")
	if primaryToken == "" {
		return nil, fmt.Errorf("TELEGRAM_BOT_TOKEN is required")
	}
	cfg.TelegramBotTokens = []string{primaryToken}

	// Add additional bot tokens if provided
	secondaryToken := os.Getenv("TELEGRAM_BOT_TOKEN_SECONDARY")
	if secondaryToken != "" {
		cfg.TelegramBotTokens = append(cfg.TelegramBotTokens, secondaryToken)
	}

	// Security configuration
	rateLimitStr := os.Getenv("RATE_LIMIT_REQUESTS")
	if rateLimitStr == "" {
		cfg.RateLimitRequests = 10
	} else {
		rateLimit, err := strconv.Atoi(rateLimitStr)
		if err != nil {
			return nil, fmt.Errorf("invalid RATE_LIMIT_REQUESTS: %v", err)
		}
		cfg.RateLimitRequests = rateLimit
	}

	rateLimitWindowStr := os.Getenv("RATE_LIMIT_WINDOW")
	if rateLimitWindowStr == "" {
		cfg.RateLimitWindow = 60 * time.Second
	} else {
		rateLimitWindow, err := time.ParseDuration(rateLimitWindowStr)
		if err != nil {
			return nil, fmt.Errorf("invalid RATE_LIMIT_WINDOW: %v", err)
		}
		cfg.RateLimitWindow = rateLimitWindow
	}

	// Token cleanup configuration
	cleanupIntervalStr := os.Getenv("TOKEN_CLEANUP_INTERVAL_HOURS")
	if cleanupIntervalStr == "" {
		cfg.TokenCleanupIntervalHours = 1
	} else {
		cleanupInterval, err := strconv.Atoi(cleanupIntervalStr)
		if err != nil {
			return nil, fmt.Errorf("invalid TOKEN_CLEANUP_INTERVAL_HOURS: %v", err)
		}
		cfg.TokenCleanupIntervalHours = cleanupInterval
	}

	cleanupTimeoutStr := os.Getenv("TOKEN_CLEANUP_TIMEOUT_MINUTES")
	if cleanupTimeoutStr == "" {
		cfg.TokenCleanupTimeoutMins = 5
	} else {
		cleanupTimeout, err := strconv.Atoi(cleanupTimeoutStr)
		if err != nil {
			return nil, fmt.Errorf("invalid TOKEN_CLEANUP_TIMEOUT_MINUTES: %v", err)
		}
		cfg.TokenCleanupTimeoutMins = cleanupTimeout
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

	// Validate Telegram bot tokens format
	for i, token := range c.TelegramBotTokens {
		if !strings.Contains(token, ":") {
			return fmt.Errorf("invalid TELEGRAM_BOT_TOKEN format at index %d", i)
		}
	}

	// Validate numeric ranges
	if c.DatabaseMaxConns < 1 || c.DatabaseMaxConns > 100 {
		return fmt.Errorf("DATABASE_MAX_CONNECTIONS must be between 1 and 100")
	}

	if c.RedisMaxConns < 1 || c.RedisMaxConns > 100 {
		return fmt.Errorf("REDIS_MAX_CONNECTIONS must be between 1 and 100")
	}

	if c.JWTExpiryHours < 1 || c.JWTExpiryHours > 168 {
		return fmt.Errorf("JWT_EXPIRY_HOURS must be between 1 and 168 (1 week)")
	}

	if c.RateLimitRequests < 1 || c.RateLimitRequests > 1000 {
		return fmt.Errorf("RATE_LIMIT_REQUESTS must be between 1 and 1000")
	}

	if c.TokenCleanupIntervalHours < 1 || c.TokenCleanupIntervalHours > 24 {
		return fmt.Errorf("TOKEN_CLEANUP_INTERVAL_HOURS must be between 1 and 24")
	}

	if c.TokenCleanupTimeoutMins < 1 || c.TokenCleanupTimeoutMins > 60 {
		return fmt.Errorf("TOKEN_CLEANUP_TIMEOUT_MINUTES must be between 1 and 60")
	}

	return nil
}

// String returns a string representation of the config (for logging, without sensitive data)
func (c *Config) String() string {
	var maskedTokens []string
	for _, token := range c.TelegramBotTokens {
		tokenMasked := "***"
		if len(token) > 10 {
			tokenMasked = token[:6] + "***"
		}
		maskedTokens = append(maskedTokens, tokenMasked)
	}

	return fmt.Sprintf(
		"Config{Host: %s, Port: %s, JWT: %s, Tokens: %v, DB: %s, Redis: %s}",
		c.ServiceHost, c.ServicePort, c.JWTIssuer, maskedTokens,
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
