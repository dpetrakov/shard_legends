package telegram

import (
	"testing"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/shard-legends/telegram-bot-service/internal/config"
)

func TestBot_handleCommand(t *testing.T) {
	cfg := &config.Config{
		WebAppBaseURL: "https://example.com",
	}
	
	bot := &Bot{
		config: cfg,
	}

	tests := []struct {
		name     string
		update   tgbotapi.Update
		expected bool // whether command was handled
	}{
		{
			name: "start command",
			update: tgbotapi.Update{
				Message: &tgbotapi.Message{
					MessageID: 1,
					From:      &tgbotapi.User{ID: 123},
					Chat:      &tgbotapi.Chat{ID: 123},
					Text:      "/start",
					Entities: []tgbotapi.MessageEntity{
						{Type: "bot_command", Offset: 0, Length: 6},
					},
				},
			},
			expected: true,
		},
		{
			name: "start command with game parameter",
			update: tgbotapi.Update{
				Message: &tgbotapi.Message{
					MessageID: 2,
					From:      &tgbotapi.User{ID: 123},
					Chat:      &tgbotapi.Chat{ID: 123},
					Text:      "/start game",
					Entities: []tgbotapi.MessageEntity{
						{Type: "bot_command", Offset: 0, Length: 6},
					},
				},
			},
			expected: true,
		},
		{
			name: "unknown command",
			update: tgbotapi.Update{
				Message: &tgbotapi.Message{
					MessageID: 3,
					From:      &tgbotapi.User{ID: 123},
					Chat:      &tgbotapi.Chat{ID: 123},
					Text:      "/unknown",
					Entities: []tgbotapi.MessageEntity{
						{Type: "bot_command", Offset: 0, Length: 8},
					},
				},
			},
			expected: false,
		},
		{
			name: "not a command",
			update: tgbotapi.Update{
				Message: &tgbotapi.Message{
					MessageID: 4,
					From:      &tgbotapi.User{ID: 123},
					Chat:      &tgbotapi.Chat{ID: 123},
					Text:      "hello world",
				},
			},
			expected: false,
		},
		{
			name: "empty message",
			update: tgbotapi.Update{
				Message: nil,
			},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := bot.handleCommand(tt.update)
			if result != tt.expected {
				t.Errorf("handleCommand() = %v, expected %v", result, tt.expected)
			}
		})
	}
}

func TestBot_handleStartCommand(t *testing.T) {
	cfg := &config.Config{
		WebAppBaseURL: "https://example.com",
	}
	
	bot := &Bot{
		config: cfg,
	}

	tests := []struct {
		name   string
		update tgbotapi.Update
	}{
		{
			name: "regular start command",
			update: tgbotapi.Update{
				Message: &tgbotapi.Message{
					MessageID: 1,
					From:      &tgbotapi.User{ID: 123},
					Chat:      &tgbotapi.Chat{ID: 123},
					Text:      "/start",
				},
			},
		},
		{
			name: "start command with game parameter",
			update: tgbotapi.Update{
				Message: &tgbotapi.Message{
					MessageID: 2,
					From:      &tgbotapi.User{ID: 123},
					Chat:      &tgbotapi.Chat{ID: 123},
					Text:      "/start game",
				},
			},
		},
		{
			name: "start command with extra spaces",
			update: tgbotapi.Update{
				Message: &tgbotapi.Message{
					MessageID: 3,
					From:      &tgbotapi.User{ID: 123},
					Chat:      &tgbotapi.Chat{ID: 123},
					Text:      "/start   game  ",
				},
			},
		},
		{
			name: "empty message",
			update: tgbotapi.Update{
				Message: nil,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// This will fail due to no API connection, but should not panic
			bot.handleStartCommand(tt.update)
		})
	}
}