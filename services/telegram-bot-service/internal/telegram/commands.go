package telegram

import (
	"fmt"
	"log"
	"strings"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

func (b *Bot) handleStartCommand(update tgbotapi.Update) {
	if update.Message == nil || update.Message.From == nil || update.Message.Chat == nil {
		return
	}

	// Extract command arguments for deep linking
	args := strings.TrimSpace(strings.TrimPrefix(update.Message.Text, "/start"))
	
	var webappURL string
	var messageText string

	if args == "game" || args == " game" {
		// Deep link: /start game
		webappURL = fmt.Sprintf("%s?start=game", b.config.WebAppBaseURL)
		messageText = "🎮 Добро пожаловать в Shard Legends: Clan Wars!\n\n" +
			"Готовы начать эпическое приключение? Нажмите кнопку ниже, чтобы запустить игру!"
	} else {
		// Regular /start command
		webappURL = b.config.WebAppBaseURL
		messageText = "🌟 Добро пожаловать в Shard Legends: Clan Wars!\n\n" +
			"Присоединяйтесь к увлекательному миру стратегических битв и кланов. " +
			"Откройте веб-приложение для начала игры!"
	}

	// Create inline keyboard with URL button (WebApp functionality)
	keyboard := tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonURL("🚀 Открыть игру", webappURL),
		),
	)

	msg := tgbotapi.NewMessage(update.Message.Chat.ID, messageText)
	msg.ReplyMarkup = keyboard
	msg.ParseMode = tgbotapi.ModeHTML

	if _, err := b.api.Send(msg); err != nil {
		log.Printf("Failed to send start message: %v", err)
	}

	log.Printf("Sent start command response to user %d (args: '%s')", update.Message.From.ID, args)
}

func (b *Bot) handleCommand(update tgbotapi.Update) bool {
	if update.Message == nil || update.Message.From == nil || !update.Message.IsCommand() {
		return false
	}

	command := update.Message.Command()
	log.Printf("Received command: /%s from user %d", command, update.Message.From.ID)

	switch command {
	case "start":
		b.handleStartCommand(update)
		return true
	default:
		// Unknown command - let it fall through to echo handler
		return false
	}
}