package telegram

import (
	"testing"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/shard-legends/telegram-bot-service/internal/config"
)

func TestBot_handleEcho(t *testing.T) {
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
			name: "text message",
			update: tgbotapi.Update{
				Message: &tgbotapi.Message{
					MessageID: 1,
					From:      &tgbotapi.User{ID: 123},
					Chat:      &tgbotapi.Chat{ID: 123},
					Text:      "Hello, world!",
				},
			},
		},
		{
			name: "sticker message",
			update: tgbotapi.Update{
				Message: &tgbotapi.Message{
					MessageID: 2,
					From:      &tgbotapi.User{ID: 123},
					Chat:      &tgbotapi.Chat{ID: 123},
					Sticker: &tgbotapi.Sticker{
						FileID: "sticker123",
					},
				},
			},
		},
		{
			name: "photo message",
			update: tgbotapi.Update{
				Message: &tgbotapi.Message{
					MessageID: 3,
					From:      &tgbotapi.User{ID: 123},
					Chat:      &tgbotapi.Chat{ID: 123},
					Photo: []tgbotapi.PhotoSize{
						{FileID: "photo123", Width: 100, Height: 100},
						{FileID: "photo456", Width: 200, Height: 200},
					},
					Caption: "Test photo",
				},
			},
		},
		{
			name: "document message",
			update: tgbotapi.Update{
				Message: &tgbotapi.Message{
					MessageID: 4,
					From:      &tgbotapi.User{ID: 123},
					Chat:      &tgbotapi.Chat{ID: 123},
					Document: &tgbotapi.Document{
						FileID:   "doc123",
						FileName: "test.pdf",
					},
				},
			},
		},
		{
			name: "voice message",
			update: tgbotapi.Update{
				Message: &tgbotapi.Message{
					MessageID: 5,
					From:      &tgbotapi.User{ID: 123},
					Chat:      &tgbotapi.Chat{ID: 123},
					Voice: &tgbotapi.Voice{
						FileID: "voice123",
					},
				},
			},
		},
		{
			name: "video message",
			update: tgbotapi.Update{
				Message: &tgbotapi.Message{
					MessageID: 6,
					From:      &tgbotapi.User{ID: 123},
					Chat:      &tgbotapi.Chat{ID: 123},
					Video: &tgbotapi.Video{
						FileID: "video123",
					},
				},
			},
		},
		{
			name: "animation message",
			update: tgbotapi.Update{
				Message: &tgbotapi.Message{
					MessageID: 7,
					From:      &tgbotapi.User{ID: 123},
					Chat:      &tgbotapi.Chat{ID: 123},
					Animation: &tgbotapi.Animation{
						FileID: "animation123",
					},
				},
			},
		},
		{
			name: "audio message",
			update: tgbotapi.Update{
				Message: &tgbotapi.Message{
					MessageID: 8,
					From:      &tgbotapi.User{ID: 123},
					Chat:      &tgbotapi.Chat{ID: 123},
					Audio: &tgbotapi.Audio{
						FileID: "audio123",
					},
				},
			},
		},
		{
			name: "empty message",
			update: tgbotapi.Update{
				Message: nil,
			},
		},
		{
			name: "message without from",
			update: tgbotapi.Update{
				Message: &tgbotapi.Message{
					MessageID: 10,
					Chat:      &tgbotapi.Chat{ID: 123},
					Text:      "test",
				},
			},
		},
		{
			name: "unsupported message type (location)",
			update: tgbotapi.Update{
				Message: &tgbotapi.Message{
					MessageID: 9,
					From:      &tgbotapi.User{ID: 123},
					Chat:      &tgbotapi.Chat{ID: 123},
					Location: &tgbotapi.Location{
						Latitude:  55.7558,
						Longitude: 37.6176,
					},
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// This will fail due to no API connection, but should not panic
			bot.handleEcho(tt.update)
		})
	}
}

func TestBot_echoTextMessage(t *testing.T) {
	bot := &Bot{}

	// Test should not panic even with nil API
	bot.echoTextMessage(123, 456, &tgbotapi.Message{
		Text: "test message",
	})
}

func TestBot_echoStickerMessage(t *testing.T) {
	bot := &Bot{}

	// Test should not panic even with nil API
	bot.echoStickerMessage(123, 456, &tgbotapi.Message{
		Sticker: &tgbotapi.Sticker{
			FileID: "sticker123",
		},
	})
}

func TestBot_echoPhotoMessage(t *testing.T) {
	bot := &Bot{}

	// Test should not panic even with nil API
	bot.echoPhotoMessage(123, 456, &tgbotapi.Message{
		Photo: []tgbotapi.PhotoSize{
			{FileID: "photo123", Width: 100, Height: 100},
		},
		Caption: "Test caption",
	})
}

func TestBot_echoDocumentMessage(t *testing.T) {
	bot := &Bot{}

	// Test should not panic even with nil API
	bot.echoDocumentMessage(123, 456, &tgbotapi.Message{
		Document: &tgbotapi.Document{
			FileID:   "doc123",
			FileName: "test.pdf",
		},
	})
}

func TestBot_echoVoiceMessage(t *testing.T) {
	bot := &Bot{}

	// Test should not panic even with nil API
	bot.echoVoiceMessage(123, 456, &tgbotapi.Message{
		Voice: &tgbotapi.Voice{
			FileID: "voice123",
		},
	})
}

func TestBot_echoVideoMessage(t *testing.T) {
	bot := &Bot{}

	// Test should not panic even with nil API
	bot.echoVideoMessage(123, 456, &tgbotapi.Message{
		Video: &tgbotapi.Video{
			FileID: "video123",
		},
	})
}

func TestBot_echoAnimationMessage(t *testing.T) {
	bot := &Bot{}

	// Test should not panic even with nil API
	bot.echoAnimationMessage(123, 456, &tgbotapi.Message{
		Animation: &tgbotapi.Animation{
			FileID: "animation123",
		},
	})
}

func TestBot_echoAudioMessage(t *testing.T) {
	bot := &Bot{}

	// Test should not panic even with nil API
	bot.echoAudioMessage(123, 456, &tgbotapi.Message{
		Audio: &tgbotapi.Audio{
			FileID: "audio123",
		},
	})
}