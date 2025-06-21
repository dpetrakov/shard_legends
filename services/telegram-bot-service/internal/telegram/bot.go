package telegram

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/shard-legends/telegram-bot-service/internal/config"
)

type Bot struct {
	api    *tgbotapi.BotAPI
	config *config.Config
	cancel context.CancelFunc
}

func NewBot(cfg *config.Config) (*Bot, error) {
	api, err := tgbotapi.NewBotAPI(cfg.TelegramBotToken)
	if err != nil {
		return nil, fmt.Errorf("failed to create bot API: %w", err)
	}

	log.Printf("Authorized on account %s", api.Self.UserName)

	return &Bot{
		api:    api,
		config: cfg,
	}, nil
}

func (b *Bot) Start(ctx context.Context) error {
	ctx, cancel := context.WithCancel(ctx)
	b.cancel = cancel

	if b.config.TelegramBotMode == "webhook" {
		return b.startWebhook(ctx)
	} else {
		return b.startLongPolling(ctx)
	}
}

func (b *Bot) Stop() {
	if b.cancel != nil {
		b.cancel()
	}
}

func (b *Bot) startWebhook(ctx context.Context) error {
	log.Println("Setting up webhook mode...")

	// Remove existing webhook first
	_, err := b.api.Request(tgbotapi.DeleteWebhookConfig{DropPendingUpdates: true})
	if err != nil {
		log.Printf("Warning: failed to delete existing webhook: %v", err)
	}

	// Set new webhook
	webhookConfig, err := tgbotapi.NewWebhook(b.config.TelegramWebhookURL)
	if err != nil {
		return fmt.Errorf("failed to create webhook config: %w", err)
	}
	_, err = b.api.Request(webhookConfig)
	if err != nil {
		return fmt.Errorf("failed to set webhook: %w", err)
	}

	log.Printf("Webhook set to: %s", b.config.TelegramWebhookURL)

	// Webhook will be handled by HTTP server in main.go
	// Just wait for context cancellation
	<-ctx.Done()
	return ctx.Err()
}

func (b *Bot) startLongPolling(ctx context.Context) error {
	log.Println("Starting long polling mode...")

	// Remove webhook if exists
	_, err := b.api.Request(tgbotapi.DeleteWebhookConfig{DropPendingUpdates: true})
	if err != nil {
		log.Printf("Warning: failed to delete webhook: %v", err)
	}

	// Configure update settings
	updateConfig := tgbotapi.NewUpdate(0)
	updateConfig.Timeout = b.config.TelegramPollTimeout

	updates := b.api.GetUpdatesChan(updateConfig)

	for {
		select {
		case <-ctx.Done():
			log.Println("Stopping long polling...")
			b.api.StopReceivingUpdates()
			return ctx.Err()

		case update := <-updates:
			go b.handleUpdate(update)
		}
	}
}

func (b *Bot) HandleWebhookUpdate(w http.ResponseWriter, r *http.Request) {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		log.Printf("Failed to read request body: %v", err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	var update tgbotapi.Update
	if err := json.Unmarshal(body, &update); err != nil {
		log.Printf("Failed to unmarshal webhook update: %v", err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	go b.handleUpdate(update)
	w.WriteHeader(http.StatusOK)
}

func (b *Bot) handleUpdate(update tgbotapi.Update) {
	// Handle commands first
	if b.handleCommand(update) {
		return
	}
	
	// If not a command, handle as echo
	b.handleEcho(update)
}

func (b *Bot) CleanupWebhook() error {
	if b.config.TelegramBotMode == "webhook" {
		log.Println("Cleaning up webhook...")
		_, err := b.api.Request(tgbotapi.DeleteWebhookConfig{DropPendingUpdates: false})
		if err != nil {
			return fmt.Errorf("failed to delete webhook: %w", err)
		}
		log.Println("Webhook cleaned up")
	}
	return nil
}