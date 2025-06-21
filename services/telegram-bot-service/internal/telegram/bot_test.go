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
	// Create a mock bot for testing webhook handling
	cfg := &config.Config{
		TelegramBotToken: "123456:test-token",
		TelegramBotMode:  "webhook",
		WebAppBaseURL:    "https://example.com",
	}

	// We can't create a real bot without a valid token, so we'll test the handler structure
	// In a real scenario, this would require mocking the telegram API

	tests := []struct {
		name           string
		method         string
		body           string
		expectedStatus int
	}{
		{
			name:           "invalid json",
			method:         "POST",
			body:           "invalid json",
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "valid json structure",
			method:         "POST",
			body:           `{"update_id": 123, "message": {"message_id": 1, "date": 1234567890, "chat": {"id": 123, "type": "private"}, "text": "test"}}`,
			expectedStatus: http.StatusOK,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a mock bot (this would normally require proper initialization)
			bot := &Bot{
				config: cfg,
			}

			req := httptest.NewRequest(tt.method, "/webhook", strings.NewReader(tt.body))
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

	bot := &Bot{
		config: cfg,
	}

	// Test that webhook mode setup doesn't panic
	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately

	err := bot.Start(ctx)
	// Expected to fail due to invalid token or cancelled context, but shouldn't panic
	_ = err
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