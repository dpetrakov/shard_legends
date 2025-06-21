package handlers

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/shard-legends/telegram-bot-service/internal/telegram"
)

func TestNewWebhookHandler(t *testing.T) {
	bot := &telegram.Bot{} // Mock bot for testing

	handler := NewWebhookHandler(bot)

	tests := []struct {
		name           string
		method         string
		body           string
		expectedStatus int
	}{
		{
			name:           "GET method not allowed",
			method:         "GET",
			body:           "",
			expectedStatus: http.StatusMethodNotAllowed,
		},
		{
			name:           "PUT method not allowed", 
			method:         "PUT",
			body:           "",
			expectedStatus: http.StatusMethodNotAllowed,
		},
		{
			name:           "POST with valid JSON",
			method:         "POST",
			body:           `{"update_id": 123}`,
			expectedStatus: http.StatusOK,
		},
		{
			name:           "POST with invalid JSON",
			method:         "POST", 
			body:           "invalid json",
			expectedStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(tt.method, "/webhook", strings.NewReader(tt.body))
			w := httptest.NewRecorder()

			handler(w, req)

			if w.Code != tt.expectedStatus {
				t.Errorf("expected status %d, got %d", tt.expectedStatus, w.Code)
			}
		})
	}
}

func TestWebhookHandler_ContentType(t *testing.T) {
	bot := &telegram.Bot{}
	handler := NewWebhookHandler(bot)

	req := httptest.NewRequest("POST", "/webhook", strings.NewReader(`{"update_id": 123}`))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	handler(w, req)

	// Should handle JSON content type properly
	if w.Code != http.StatusOK && w.Code != http.StatusBadRequest {
		t.Errorf("unexpected status for JSON content: %d", w.Code)
	}
}