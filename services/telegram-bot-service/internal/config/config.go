package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"
)

type Config struct {
	// Telegram bot configuration
	TelegramBotToken    string
	TelegramBotMode     string
	TelegramPollTimeout int
	TelegramWebhookURL  string

	// Web app configuration
	WebAppBaseURL string

	// Service configuration
	ServicePort string
}

// Load loads configuration from environment variables
func Load() (*Config, error) {
	cfg := &Config{}

	// Required fields
	cfg.TelegramBotToken = os.Getenv("TELEGRAM_BOT_TOKEN")
	if cfg.TelegramBotToken == "" {
		return nil, fmt.Errorf("TELEGRAM_BOT_TOKEN is required")
	}

	cfg.TelegramBotMode = strings.ToLower(os.Getenv("TELEGRAM_BOT_MODE"))
	if cfg.TelegramBotMode == "" {
		cfg.TelegramBotMode = "longpoll" // default mode
	}
	if cfg.TelegramBotMode != "webhook" && cfg.TelegramBotMode != "longpoll" {
		return nil, fmt.Errorf("TELEGRAM_BOT_MODE must be 'webhook' or 'longpoll', got: %s", cfg.TelegramBotMode)
	}

	cfg.WebAppBaseURL = os.Getenv("WEBAPP_BASE_URL")
	if cfg.WebAppBaseURL == "" {
		return nil, fmt.Errorf("WEBAPP_BASE_URL is required")
	}

	// Mode-specific configuration
	if cfg.TelegramBotMode == "webhook" {
		cfg.TelegramWebhookURL = os.Getenv("TELEGRAM_WEBHOOK_URL")
		if cfg.TelegramWebhookURL == "" {
			return nil, fmt.Errorf("TELEGRAM_WEBHOOK_URL is required for webhook mode")
		}

		cfg.ServicePort = os.Getenv("SERVICE_PORT")
		if cfg.ServicePort == "" {
			cfg.ServicePort = "8080" // default port
		}
	} else {
		// Long polling mode
		pollTimeoutStr := os.Getenv("TELEGRAM_POLL_TIMEOUT")
		if pollTimeoutStr == "" {
			cfg.TelegramPollTimeout = 30 // default timeout
		} else {
			timeout, err := strconv.Atoi(pollTimeoutStr)
			if err != nil {
				return nil, fmt.Errorf("invalid TELEGRAM_POLL_TIMEOUT: %v", err)
			}
			if timeout < 0 || timeout > 60 {
				return nil, fmt.Errorf("TELEGRAM_POLL_TIMEOUT must be between 0 and 60, got: %d", timeout)
			}
			cfg.TelegramPollTimeout = timeout
		}
	}

	return cfg, nil
}

// Validate performs additional validation on the configuration
func (c *Config) Validate() error {
	// Validate bot token format
	if !strings.HasPrefix(c.TelegramBotToken, "bot") && !strings.Contains(c.TelegramBotToken, ":") {
		return fmt.Errorf("invalid bot token format")
	}

	// Validate URLs
	if c.TelegramBotMode == "webhook" {
		if !strings.HasPrefix(c.TelegramWebhookURL, "https://") {
			return fmt.Errorf("webhook URL must use HTTPS")
		}
	}

	if !strings.HasPrefix(c.WebAppBaseURL, "http://") && !strings.HasPrefix(c.WebAppBaseURL, "https://") {
		return fmt.Errorf("WebApp base URL must start with http:// or https://")
	}

	// Remove trailing slashes for consistency
	c.WebAppBaseURL = strings.TrimRight(c.WebAppBaseURL, "/")
	if c.TelegramWebhookURL != "" {
		c.TelegramWebhookURL = strings.TrimRight(c.TelegramWebhookURL, "/")
	}

	return nil
}

// String returns a string representation of the config (for logging, without sensitive data)
func (c *Config) String() string {
	tokenMasked := "***"
	if len(c.TelegramBotToken) > 10 {
		tokenMasked = c.TelegramBotToken[:6] + "***"
	}

	return fmt.Sprintf(
		"Config{Mode: %s, Token: %s, WebApp: %s, Port: %s, PollTimeout: %d}",
		c.TelegramBotMode, tokenMasked, c.WebAppBaseURL, c.ServicePort, c.TelegramPollTimeout,
	)
}