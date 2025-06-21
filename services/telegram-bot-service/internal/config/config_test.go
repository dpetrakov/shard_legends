package config

import (
	"os"
	"testing"
)

func TestLoad_Success(t *testing.T) {
	tests := []struct {
		name     string
		envVars  map[string]string
		expected *Config
	}{
		{
			name: "webhook mode with all required fields",
			envVars: map[string]string{
				"TELEGRAM_BOT_TOKEN":   "bot123456:ABC-DEF1234ghIkl-zyx57W2v1u123ew11",
				"TELEGRAM_BOT_MODE":    "webhook",
				"WEBAPP_BASE_URL":      "https://example.com",
				"TELEGRAM_WEBHOOK_URL": "https://api.example.com/webhook",
				"SERVICE_PORT":         "9090",
			},
			expected: &Config{
				TelegramBotToken:    "bot123456:ABC-DEF1234ghIkl-zyx57W2v1u123ew11",
				TelegramBotMode:     "webhook",
				WebAppBaseURL:       "https://example.com",
				TelegramWebhookURL:  "https://api.example.com/webhook",
				ServicePort:         "9090",
				TelegramPollTimeout: 0,
			},
		},
		{
			name: "webhook mode with default port",
			envVars: map[string]string{
				"TELEGRAM_BOT_TOKEN":   "bot123456:ABC-DEF1234ghIkl-zyx57W2v1u123ew11",
				"TELEGRAM_BOT_MODE":    "webhook",
				"WEBAPP_BASE_URL":      "https://example.com",
				"TELEGRAM_WEBHOOK_URL": "https://api.example.com/webhook",
			},
			expected: &Config{
				TelegramBotToken:    "bot123456:ABC-DEF1234ghIkl-zyx57W2v1u123ew11",
				TelegramBotMode:     "webhook",
				WebAppBaseURL:       "https://example.com",
				TelegramWebhookURL:  "https://api.example.com/webhook",
				ServicePort:         "8080",
				TelegramPollTimeout: 0,
			},
		},
		{
			name: "longpoll mode with custom timeout",
			envVars: map[string]string{
				"TELEGRAM_BOT_TOKEN":    "bot123456:ABC-DEF1234ghIkl-zyx57W2v1u123ew11",
				"TELEGRAM_BOT_MODE":     "longpoll",
				"WEBAPP_BASE_URL":       "https://example.com",
				"TELEGRAM_POLL_TIMEOUT": "45",
			},
			expected: &Config{
				TelegramBotToken:    "bot123456:ABC-DEF1234ghIkl-zyx57W2v1u123ew11",
				TelegramBotMode:     "longpoll",
				WebAppBaseURL:       "https://example.com",
				TelegramPollTimeout: 45,
			},
		},
		{
			name: "longpoll mode with default timeout",
			envVars: map[string]string{
				"TELEGRAM_BOT_TOKEN": "bot123456:ABC-DEF1234ghIkl-zyx57W2v1u123ew11",
				"TELEGRAM_BOT_MODE":  "longpoll",
				"WEBAPP_BASE_URL":    "https://example.com",
			},
			expected: &Config{
				TelegramBotToken:    "bot123456:ABC-DEF1234ghIkl-zyx57W2v1u123ew11",
				TelegramBotMode:     "longpoll",
				WebAppBaseURL:       "https://example.com",
				TelegramPollTimeout: 30,
			},
		},
		{
			name: "default mode (longpoll) when mode not specified",
			envVars: map[string]string{
				"TELEGRAM_BOT_TOKEN": "bot123456:ABC-DEF1234ghIkl-zyx57W2v1u123ew11",
				"WEBAPP_BASE_URL":    "https://example.com",
			},
			expected: &Config{
				TelegramBotToken:    "bot123456:ABC-DEF1234ghIkl-zyx57W2v1u123ew11",
				TelegramBotMode:     "longpoll",
				WebAppBaseURL:       "https://example.com",
				TelegramPollTimeout: 30,
			},
		},
		{
			name: "mode is case insensitive",
			envVars: map[string]string{
				"TELEGRAM_BOT_TOKEN": "bot123456:ABC-DEF1234ghIkl-zyx57W2v1u123ew11",
				"TELEGRAM_BOT_MODE":  "WEBHOOK",
				"WEBAPP_BASE_URL":    "https://example.com",
				"TELEGRAM_WEBHOOK_URL": "https://api.example.com/webhook",
			},
			expected: &Config{
				TelegramBotToken:    "bot123456:ABC-DEF1234ghIkl-zyx57W2v1u123ew11",
				TelegramBotMode:     "webhook",
				WebAppBaseURL:       "https://example.com",
				TelegramWebhookURL:  "https://api.example.com/webhook",
				ServicePort:         "8080",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Clear environment
			clearEnv()
			defer clearEnv()

			// Set test environment variables
			for k, v := range tt.envVars {
				os.Setenv(k, v)
			}

			// Load config
			cfg, err := Load()
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			// Compare configs
			if !compareConfigs(cfg, tt.expected) {
				t.Errorf("config mismatch:\ngot:  %+v\nwant: %+v", cfg, tt.expected)
			}
		})
	}
}

func TestLoad_Errors(t *testing.T) {
	tests := []struct {
		name        string
		envVars     map[string]string
		expectedErr string
	}{
		{
			name:        "missing bot token",
			envVars:     map[string]string{},
			expectedErr: "TELEGRAM_BOT_TOKEN is required",
		},
		{
			name: "missing webapp base URL",
			envVars: map[string]string{
				"TELEGRAM_BOT_TOKEN": "bot123456:ABC-DEF1234ghIkl-zyx57W2v1u123ew11",
			},
			expectedErr: "WEBAPP_BASE_URL is required",
		},
		{
			name: "invalid bot mode",
			envVars: map[string]string{
				"TELEGRAM_BOT_TOKEN": "bot123456:ABC-DEF1234ghIkl-zyx57W2v1u123ew11",
				"TELEGRAM_BOT_MODE":  "invalid",
				"WEBAPP_BASE_URL":    "https://example.com",
			},
			expectedErr: "TELEGRAM_BOT_MODE must be 'webhook' or 'longpoll'",
		},
		{
			name: "webhook mode missing webhook URL",
			envVars: map[string]string{
				"TELEGRAM_BOT_TOKEN": "bot123456:ABC-DEF1234ghIkl-zyx57W2v1u123ew11",
				"TELEGRAM_BOT_MODE":  "webhook",
				"WEBAPP_BASE_URL":    "https://example.com",
			},
			expectedErr: "TELEGRAM_WEBHOOK_URL is required for webhook mode",
		},
		{
			name: "invalid poll timeout (not a number)",
			envVars: map[string]string{
				"TELEGRAM_BOT_TOKEN":    "bot123456:ABC-DEF1234ghIkl-zyx57W2v1u123ew11",
				"TELEGRAM_BOT_MODE":     "longpoll",
				"WEBAPP_BASE_URL":       "https://example.com",
				"TELEGRAM_POLL_TIMEOUT": "abc",
			},
			expectedErr: "invalid TELEGRAM_POLL_TIMEOUT",
		},
		{
			name: "poll timeout too high",
			envVars: map[string]string{
				"TELEGRAM_BOT_TOKEN":    "bot123456:ABC-DEF1234ghIkl-zyx57W2v1u123ew11",
				"TELEGRAM_BOT_MODE":     "longpoll",
				"WEBAPP_BASE_URL":       "https://example.com",
				"TELEGRAM_POLL_TIMEOUT": "100",
			},
			expectedErr: "TELEGRAM_POLL_TIMEOUT must be between 0 and 60",
		},
		{
			name: "poll timeout negative",
			envVars: map[string]string{
				"TELEGRAM_BOT_TOKEN":    "bot123456:ABC-DEF1234ghIkl-zyx57W2v1u123ew11",
				"TELEGRAM_BOT_MODE":     "longpoll",
				"WEBAPP_BASE_URL":       "https://example.com",
				"TELEGRAM_POLL_TIMEOUT": "-1",
			},
			expectedErr: "TELEGRAM_POLL_TIMEOUT must be between 0 and 60",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Clear environment
			clearEnv()
			defer clearEnv()

			// Set test environment variables
			for k, v := range tt.envVars {
				os.Setenv(k, v)
			}

			// Load config
			_, err := Load()
			if err == nil {
				t.Fatal("expected error, got nil")
			}

			if err.Error() != tt.expectedErr && !contains(err.Error(), tt.expectedErr) {
				t.Errorf("error mismatch:\ngot:  %v\nwant: %v", err.Error(), tt.expectedErr)
			}
		})
	}
}

func TestValidate(t *testing.T) {
	tests := []struct {
		name        string
		config      *Config
		expectedErr string
		expectValid bool
	}{
		{
			name: "valid webhook config",
			config: &Config{
				TelegramBotToken:   "bot123456:ABC-DEF1234ghIkl-zyx57W2v1u123ew11",
				TelegramBotMode:    "webhook",
				WebAppBaseURL:      "https://example.com/",
				TelegramWebhookURL: "https://api.example.com/webhook/",
			},
			expectValid: true,
		},
		{
			name: "valid longpoll config",
			config: &Config{
				TelegramBotToken: "123456:ABC-DEF1234ghIkl-zyx57W2v1u123ew11",
				TelegramBotMode:  "longpoll",
				WebAppBaseURL:    "http://localhost:3000",
			},
			expectValid: true,
		},
		{
			name: "invalid bot token format",
			config: &Config{
				TelegramBotToken: "invalid-token",
				TelegramBotMode:  "longpoll",
				WebAppBaseURL:    "https://example.com",
			},
			expectedErr: "invalid bot token format",
		},
		{
			name: "webhook URL not HTTPS",
			config: &Config{
				TelegramBotToken:   "bot123456:ABC-DEF1234ghIkl-zyx57W2v1u123ew11",
				TelegramBotMode:    "webhook",
				WebAppBaseURL:      "https://example.com",
				TelegramWebhookURL: "http://api.example.com/webhook",
			},
			expectedErr: "webhook URL must use HTTPS",
		},
		{
			name: "invalid WebApp URL",
			config: &Config{
				TelegramBotToken: "bot123456:ABC-DEF1234ghIkl-zyx57W2v1u123ew11",
				TelegramBotMode:  "longpoll",
				WebAppBaseURL:    "example.com",
			},
			expectedErr: "WebApp base URL must start with http:// or https://",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.config.Validate()
			if tt.expectValid {
				if err != nil {
					t.Fatalf("unexpected error: %v", err)
				}
				// Check trailing slashes are removed
				if tt.config.WebAppBaseURL[len(tt.config.WebAppBaseURL)-1] == '/' {
					t.Error("WebAppBaseURL should have trailing slash removed")
				}
				if tt.config.TelegramWebhookURL != "" && tt.config.TelegramWebhookURL[len(tt.config.TelegramWebhookURL)-1] == '/' {
					t.Error("TelegramWebhookURL should have trailing slash removed")
				}
			} else {
				if err == nil {
					t.Fatal("expected error, got nil")
				}
				if err.Error() != tt.expectedErr {
					t.Errorf("error mismatch:\ngot:  %v\nwant: %v", err.Error(), tt.expectedErr)
				}
			}
		})
	}
}

func TestString(t *testing.T) {
	tests := []struct {
		name     string
		config   *Config
		expected string
	}{
		{
			name: "webhook mode",
			config: &Config{
				TelegramBotToken:   "bot123456:ABC-DEF1234ghIkl-zyx57W2v1u123ew11",
				TelegramBotMode:    "webhook",
				WebAppBaseURL:      "https://example.com",
				ServicePort:        "8080",
				TelegramPollTimeout: 0,
			},
			expected: "Config{Mode: webhook, Token: bot123***, WebApp: https://example.com, Port: 8080, PollTimeout: 0}",
		},
		{
			name: "longpoll mode",
			config: &Config{
				TelegramBotToken:    "bot123456:ABC-DEF1234ghIkl-zyx57W2v1u123ew11",
				TelegramBotMode:     "longpoll",
				WebAppBaseURL:       "https://example.com",
				TelegramPollTimeout: 30,
			},
			expected: "Config{Mode: longpoll, Token: bot123***, WebApp: https://example.com, Port: , PollTimeout: 30}",
		},
		{
			name: "short token",
			config: &Config{
				TelegramBotToken:    "short",
				TelegramBotMode:     "longpoll",
				WebAppBaseURL:       "https://example.com",
				TelegramPollTimeout: 30,
			},
			expected: "Config{Mode: longpoll, Token: ***, WebApp: https://example.com, Port: , PollTimeout: 30}",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.config.String()
			if result != tt.expected {
				t.Errorf("String() mismatch:\ngot:  %s\nwant: %s", result, tt.expected)
			}
		})
	}
}

// Helper functions

func clearEnv() {
	os.Unsetenv("TELEGRAM_BOT_TOKEN")
	os.Unsetenv("TELEGRAM_BOT_MODE")
	os.Unsetenv("TELEGRAM_POLL_TIMEOUT")
	os.Unsetenv("TELEGRAM_WEBHOOK_URL")
	os.Unsetenv("WEBAPP_BASE_URL")
	os.Unsetenv("SERVICE_PORT")
}

func compareConfigs(a, b *Config) bool {
	if a == nil || b == nil {
		return a == b
	}
	return a.TelegramBotToken == b.TelegramBotToken &&
		a.TelegramBotMode == b.TelegramBotMode &&
		a.TelegramPollTimeout == b.TelegramPollTimeout &&
		a.TelegramWebhookURL == b.TelegramWebhookURL &&
		a.WebAppBaseURL == b.WebAppBaseURL &&
		a.ServicePort == b.ServicePort
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && s[:len(substr)] == substr
}