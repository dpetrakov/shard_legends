package handlers

import (
	"net/http"

	"github.com/shard-legends/telegram-bot-service/internal/telegram"
)

func NewWebhookHandler(bot *telegram.Bot) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}

		bot.HandleWebhookUpdate(w, r)
	}
}