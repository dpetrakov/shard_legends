package jwt

import (
	"context"
	"crypto/rsa"
	"encoding/pem"
	"fmt"
	"io"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/shard-legends/production-service/internal/database"
	"github.com/shard-legends/production-service/pkg/logger"
	"go.uber.org/zap"
)

type CustomClaims struct {
	jwt.RegisteredClaims
	TelegramID int64 `json:"telegram_id"`
}

type Validator struct {
	publicKey    *rsa.PublicKey
	publicKeyURL string
	redis        *database.RedisClient
	mu           sync.RWMutex
	cacheTTL     time.Duration
}

func NewValidator(publicKeyURL string, redis *database.RedisClient, cacheTTL time.Duration) *Validator {
	return &Validator{
		publicKeyURL: publicKeyURL,
		redis:        redis,
		cacheTTL:     cacheTTL,
	}
}

func (v *Validator) Initialize(ctx context.Context) error {
	publicKey, err := v.fetchPublicKey(ctx)
	if err != nil {
		return fmt.Errorf("failed to fetch public key: %w", err)
	}

	v.mu.Lock()
	v.publicKey = publicKey
	v.mu.Unlock()

	logger.Info("JWT validator initialized with public key from auth service")
	return nil
}

func (v *Validator) fetchPublicKey(ctx context.Context) (*rsa.PublicKey, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, v.publicKeyURL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch public key: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	keyData, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read public key: %w", err)
	}

	block, _ := pem.Decode(keyData)
	if block == nil {
		return nil, fmt.Errorf("failed to decode PEM block")
	}

	publicKey, err := jwt.ParseRSAPublicKeyFromPEM(keyData)
	if err != nil {
		return nil, fmt.Errorf("failed to parse RSA public key: %w", err)
	}

	return publicKey, nil
}

func (v *Validator) ValidateToken(ctx context.Context, tokenString string) (*CustomClaims, error) {
	tokenString = strings.TrimPrefix(tokenString, "Bearer ")
	tokenString = strings.TrimSpace(tokenString)

	v.mu.RLock()
	publicKey := v.publicKey
	v.mu.RUnlock()

	if publicKey == nil {
		return nil, fmt.Errorf("public key not initialized")
	}

	token, err := jwt.ParseWithClaims(tokenString, &CustomClaims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodRSA); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return publicKey, nil
	})

	if err != nil {
		return nil, fmt.Errorf("failed to parse token: %w", err)
	}

	claims, ok := token.Claims.(*CustomClaims)
	if !ok || !token.Valid {
		return nil, fmt.Errorf("invalid token claims")
	}

	if claims.ID == "" {
		return nil, fmt.Errorf("missing jti claim")
	}

	revoked, err := v.redis.IsJWTRevoked(ctx, claims.ID)
	if err != nil {
		logger.Error("Failed to check token revocation", zap.Error(err))
		// В случае ошибки Redis продолжаем - не блокируем пользователей
	} else if revoked {
		return nil, fmt.Errorf("token has been revoked")
	}

	return claims, nil
}

func (v *Validator) RefreshPublicKey(ctx context.Context) error {
	publicKey, err := v.fetchPublicKey(ctx)
	if err != nil {
		return fmt.Errorf("failed to refresh public key: %w", err)
	}

	v.mu.Lock()
	v.publicKey = publicKey
	v.mu.Unlock()

	logger.Info("JWT public key refreshed")
	return nil
}

func (v *Validator) GetPublicKey() *rsa.PublicKey {
	v.mu.RLock()
	defer v.mu.RUnlock()
	return v.publicKey
}