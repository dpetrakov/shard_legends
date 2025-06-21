package telegram

import (
	"fmt"
	"log"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

func (b *Bot) handleStartCommand(update tgbotapi.Update) {
	if update.Message == nil || update.Message.From == nil || update.Message.Chat == nil {
		return
	}

	messageText := "🌟 Добро пожаловать в Shard Legends: Clan Wars!\n\n" +
		"Присоединяйтесь к увлекательному миру стратегических битв и кланов. " +
		"Откройте мини-приложение для начала игры!"

	// Create inline keyboard with Mini App button
	miniAppURL := fmt.Sprintf("https://t.me/%s/%s", b.config.BotUsername, b.config.MiniAppShortName)

	// Create URL button that opens the mini app directly
	miniAppButton := tgbotapi.InlineKeyboardButton{
		Text: "🚀 Открыть игру",
		URL:  &miniAppURL,
	}

	keyboard := tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(miniAppButton),
	)

	msg := tgbotapi.NewMessage(update.Message.Chat.ID, messageText)
	msg.ReplyMarkup = keyboard
	msg.ParseMode = tgbotapi.ModeHTML

	if _, err := b.api.Send(msg); err != nil {
		log.Printf("Failed to send start message: %v", err)
	}

	log.Printf("Sent start command response to user %d", update.Message.From.ID)
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