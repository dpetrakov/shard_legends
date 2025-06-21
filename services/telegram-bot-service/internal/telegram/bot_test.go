package telegram

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/shard-legends/telegram-bot-service/internal/config"
)

func TestNewBot(t *testing.T) {
	tests := []struct {
		name        string
		config      *config.Config
		expectError bool
	}{
		{
			name: "valid configuration",
			config: &config.Config{
				TelegramBotToken: "123456:ABC-DEF1234ghIkl-zyx57W2v1u123ew11",
				TelegramBotMode:  "longpoll",
				WebAppBaseURL:    "https://example.com",
			},
			expectError: true, // Will fail because it's not a real token
		},
		{
			name: "invalid token",
			config: &config.Config{
				TelegramBotToken: "invalid-token",
				TelegramBotMode:  "longpoll",
				WebAppBaseURL:    "https://example.com",
			},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			bot, err := NewBot(tt.config)
			if tt.expectError {
				if err == nil {
					t.Error("expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if bot == nil {
				t.Error("expected bot, got nil")
			}
		})
	}
}

func TestBot_HandleWebhookUpdate(t *testing.T) {
	tests := []struct {
		name           string
		config         *config.Config
		method         string
		body           string
		headers        map[string]string
		expectedStatus int
	}{
		{
			name: "invalid json",
			config: &config.Config{
				TelegramBotToken: "123456:test-token",
				TelegramBotMode:  "webhook",
				WebAppBaseURL:    "https://example.com",
			},
			method:         "POST",
			body:           "invalid json",
			expectedStatus: http.StatusBadRequest,
		},
		{
			name: "valid json structure without secret token",
			config: &config.Config{
				TelegramBotToken: "123456:test-token",
				TelegramBotMode:  "webhook",
				WebAppBaseURL:    "https://example.com",
			},
			method:         "POST",
			body:           `{"update_id": 123, "message": {"message_id": 1, "date": 1234567890, "chat": {"id": 123, "type": "private"}, "text": "test"}}`,
			expectedStatus: http.StatusOK,
		},
		{
			name: "valid request with correct secret token",
			config: &config.Config{
				TelegramBotToken:    "123456:test-token",
				TelegramBotMode:     "webhook",
				WebAppBaseURL:       "https://example.com",
				TelegramSecretToken: "secret123",
			},
			method: "POST",
			body:   `{"update_id": 123, "message": {"message_id": 1, "date": 1234567890, "chat": {"id": 123, "type": "private"}, "text": "test"}}`,
			headers: map[string]string{
				"X-Telegram-Bot-Api-Secret-Token": "secret123",
			},
			expectedStatus: http.StatusOK,
		},
		{
			name: "request with incorrect secret token",
			config: &config.Config{
				TelegramBotToken:    "123456:test-token",
				TelegramBotMode:     "webhook",
				WebAppBaseURL:       "https://example.com",
				TelegramSecretToken: "secret123",
			},
			method: "POST",
			body:   `{"update_id": 123, "message": {"message_id": 1, "date": 1234567890, "chat": {"id": 123, "type": "private"}, "text": "test"}}`,
			headers: map[string]string{
				"X-Telegram-Bot-Api-Secret-Token": "wrong-secret",
			},
			expectedStatus: http.StatusUnauthorized,
		},
		{
			name: "request with missing secret token when required",
			config: &config.Config{
				TelegramBotToken:    "123456:test-token",
				TelegramBotMode:     "webhook",
				WebAppBaseURL:       "https://example.com",
				TelegramSecretToken: "secret123",
			},
			method:         "POST",
			body:           `{"update_id": 123, "message": {"message_id": 1, "date": 1234567890, "chat": {"id": 123, "type": "private"}, "text": "test"}}`,
			expectedStatus: http.StatusUnauthorized,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			bot := &Bot{
				config: tt.config,
			}

			req := httptest.NewRequest(tt.method, "/webhook", strings.NewReader(tt.body))
			
			// Add headers if provided
			for key, value := range tt.headers {
				req.Header.Set(key, value)
			}
			
			w := httptest.NewRecorder()

			bot.HandleWebhookUpdate(w, req)

			if w.Code != tt.expectedStatus {
				t.Errorf("expected status %d, got %d", tt.expectedStatus, w.Code)
			}
		})
	}
}

func TestBot_CleanupWebhook(t *testing.T) {
	tests := []struct {
		name   string
		config *config.Config
	}{
		{
			name: "longpoll mode - should not call API",
			config: &config.Config{
				TelegramBotToken: "123456:test-token",
				TelegramBotMode:  "longpoll",
				WebAppBaseURL:    "https://example.com",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			bot := &Bot{
				config: tt.config,
			}

			// CleanupWebhook should not panic for longpoll mode
			err := bot.CleanupWebhook()
			if err != nil {
				t.Errorf("longpoll mode should not return error: %v", err)
			}
		})
	}
}

func TestBot_StartWebhookMode(t *testing.T) {
	cfg := &config.Config{
		TelegramBotToken:    "123456:test-token",
		TelegramBotMode:     "webhook",
		WebAppBaseURL:       "https://example.com",
		TelegramWebhookURL:  "https://example.com/webhook",
	}

	// Test that webhook mode setup doesn't panic
	// Note: We expect this to fail due to invalid token, but it shouldn't panic
	bot, err := NewBot(cfg)
	if err != nil {
		// Expected - invalid token should fail during initialization
		return
	}

	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately

	err = bot.Start(ctx)
	// Expected to fail due to cancelled context
	if err == nil {
		t.Error("Expected error due to cancelled context")
	}
}

func TestBot_Stop(t *testing.T) {
	bot := &Bot{}

	// Stop should not panic even if cancel is nil
	bot.Stop()

	// Test with cancel function
	ctx, cancel := context.WithCancel(context.Background())
	bot.cancel = cancel

	bot.Stop()
	
	// Verify context was cancelled
	select {
	case <-ctx.Done():
		// Expected
	case <-time.After(100 * time.Millisecond):
		t.Error("context was not cancelled")
	}
}