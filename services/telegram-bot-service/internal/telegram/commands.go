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
	
	var messageText string
	var startParam string

	if args == "game" || args == " game" {
		// Deep link: /start game
		startParam = "game"
		messageText = "🎮 Добро пожаловать в Shard Legends: Clan Wars!\n\n" +
			"Готовы начать эпическое приключение? Нажмите кнопку ниже, чтобы запустить игру!"
	} else {
		// Regular /start command
		messageText = "🌟 Добро пожаловать в Shard Legends: Clan Wars!\n\n" +
			"Присоединяйтесь к увлекательному миру стратегических битв и кланов. " +
			"Откройте мини-приложение для начала игры!"
	}

	// Create inline keyboard with Web App button for Telegram Mini App
	var keyboard tgbotapi.InlineKeyboardMarkup
	var webAppURL string
	
	if startParam != "" {
		// Use Web App with start parameter
		webAppURL = fmt.Sprintf("%s?start=%s", b.config.WebAppBaseURL, startParam)
	} else {
		// Use Web App without parameters
		webAppURL = b.config.WebAppBaseURL
	}

	// Create WebApp button - use URL button as fallback if WebApp not supported
	webAppButton := tgbotapi.InlineKeyboardButton{
		Text: "🚀 Открыть игру",
		URL:  &webAppURL,
	}

	keyboard = tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(webAppButton),
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