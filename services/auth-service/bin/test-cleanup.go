package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/google/uuid"
	"github.com/shard-legends/auth-service/internal/config"
	"github.com/shard-legends/auth-service/internal/storage"
	"github.com/shard-legends/auth-service/pkg/utils"
)

func main() {
	// Initialize logger
	logger := utils.NewLogger()

	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// Initialize Redis token storage
	tokenStorage, err := storage.NewRedisTokenStorage(cfg.RedisURL, cfg.RedisMaxConns, logger, nil)
	if err != nil {
		log.Fatalf("Failed to connect to Redis: %v", err)
	}
	defer tokenStorage.Close()

	ctx := context.Background()

	// Create test expired tokens
	userID := uuid.New()
	telegramID := int64(999999999)

	// Create tokens that are already expired
	expiredTime := time.Now().Add(-time.Hour) // Expired 1 hour ago
	
	fmt.Println("Creating test expired tokens...")
	
	for i := 0; i < 3; i++ {
		jti := uuid.New().String()
		
		err := tokenStorage.StoreActiveToken(ctx, jti, userID, telegramID, expiredTime)
		if err != nil {
			log.Printf("Failed to store expired token %d: %v", i+1, err)
		} else {
			fmt.Printf("Created expired token %d: %s\n", i+1, jti)
		}
	}

	// Create one revoked expired token
	revokedJTI := uuid.New().String()
	err = tokenStorage.StoreActiveToken(ctx, revokedJTI, userID, telegramID, expiredTime)
	if err != nil {
		log.Printf("Failed to store token for revocation: %v", err)
	} else {
		// Revoke it
		err = tokenStorage.RevokeToken(ctx, revokedJTI)
		if err != nil {
			log.Printf("Failed to revoke token: %v", err)
		} else {
			fmt.Printf("Created expired revoked token: %s\n", revokedJTI)
		}
	}

	// Check counts before cleanup
	activeCount, _ := tokenStorage.GetActiveTokenCount(ctx)
	fmt.Printf("\nBefore cleanup - Active tokens: %d\n", activeCount)

	// Run cleanup
	fmt.Println("\nRunning cleanup...")
	cleanedCount, err := tokenStorage.CleanupExpiredTokens(ctx)
	if err != nil {
		log.Fatalf("Cleanup failed: %v", err)
	}

	// Check counts after cleanup
	activeCountAfter, _ := tokenStorage.GetActiveTokenCount(ctx)
	fmt.Printf("\nAfter cleanup:")
	fmt.Printf("\n- Cleaned tokens: %d", cleanedCount)
	fmt.Printf("\n- Active tokens remaining: %d\n", activeCountAfter)
	
	fmt.Println("\nCleanup test completed!")
}