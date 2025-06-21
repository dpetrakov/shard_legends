package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gorilla/mux"
	"github.com/joho/godotenv"
	"github.com/shard-legends/telegram-bot-service/internal/config"
	"github.com/shard-legends/telegram-bot-service/internal/handlers"
	"github.com/shard-legends/telegram-bot-service/internal/telegram"
)

func main() {
	// Load environment variables from .env file in development
	if err := godotenv.Load(); err != nil {
		log.Printf("Error loading .env file: %v", err)
	}

	// Initialize logger
	log.SetFlags(log.LstdFlags | log.Lshortfile)
	log.Println("Starting Telegram Bot Service...")

	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// Validate configuration
	if err := cfg.Validate(); err != nil {
		log.Fatalf("Invalid configuration: %v", err)
	}

	log.Printf("Configuration loaded: %s", cfg)

	// Initialize Telegram bot
	bot, err := telegram.NewBot(cfg)
	if err != nil {
		log.Fatalf("Failed to create bot: %v", err)
	}

	// Setup HTTP server for webhook mode or health checks
	var httpServer *http.Server
	if cfg.TelegramBotMode == "webhook" {
		router := mux.NewRouter()
		
		// Health check endpoint
		router.HandleFunc("/health", handlers.NewHealthHandler(cfg.TelegramBotMode)).Methods("GET")
		
		// Webhook endpoint
		router.HandleFunc("/webhook", handlers.NewWebhookHandler(bot)).Methods("POST")
		
		httpServer = &http.Server{
			Addr:         fmt.Sprintf(":%s", cfg.ServicePort),
			Handler:      router,
			ReadTimeout:  15 * time.Second,
			WriteTimeout: 15 * time.Second,
			IdleTimeout:  60 * time.Second,
		}

		// Start HTTP server in goroutine
		go func() {
			log.Printf("Starting HTTP server on port %s", cfg.ServicePort)
			if err := httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
				log.Fatalf("Failed to start HTTP server: %v", err)
			}
		}()
	} else {
		// For longpoll mode, still start a minimal HTTP server for health checks
		http.HandleFunc("/health", handlers.NewHealthHandler(cfg.TelegramBotMode))
		
		httpServer = &http.Server{
			Addr:         ":8080",
			ReadTimeout:  15 * time.Second,
			WriteTimeout: 15 * time.Second,
			IdleTimeout:  60 * time.Second,
		}

		// Start HTTP server in goroutine
		go func() {
			log.Println("Starting health check server on port 8080")
			if err := httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
				log.Fatalf("Failed to start health check server: %v", err)
			}
		}()
	}

	// Start bot in appropriate mode
	botCtx, botCancel := context.WithCancel(context.Background())
	go func() {
		if err := bot.Start(botCtx); err != nil && err != context.Canceled {
			log.Printf("Bot error: %v", err)
		}
	}()

	// Setup signal handling for graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	// Wait for shutdown signal
	sig := <-sigChan
	log.Printf("Received signal: %v. Shutting down...", sig)

	// Graceful shutdown with timeout
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer shutdownCancel()

	// Stop bot first
	botCancel()
	bot.Stop()

	// Cleanup webhook if needed
	if err := bot.CleanupWebhook(); err != nil {
		log.Printf("Failed to cleanup webhook: %v", err)
	}

	// Stop HTTP server
	if httpServer != nil {
		if err := httpServer.Shutdown(shutdownCtx); err != nil {
			log.Printf("HTTP server shutdown error: %v", err)
		} else {
			log.Println("HTTP server gracefully stopped")
		}
	}

	log.Println("Telegram Bot Service stopped")
}