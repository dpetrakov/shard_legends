package middleware

import (
	"context"
	"crypto/rsa"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/redis/go-redis/v9"
	"user-service/pkg/logger"
)

// UserContext holds user information from JWT token
type UserContext struct {
	UserID     string `json:"user_id"`
	TelegramID int64  `json:"telegram_id"`
}

// RedisInterface defines the methods needed from Redis for JWT operations
type RedisInterface interface {
	Exists(ctx context.Context, keys ...string) *redis.IntCmd
}

// JWTAuthMiddleware provides JWT authentication for user service
type JWTAuthMiddleware struct {
	publicKey   *rsa.PublicKey
	redisClient RedisInterface
	logger      *logger.Logger
}

// NewJWTAuthMiddleware creates a new JWT authentication middleware
func NewJWTAuthMiddleware(publicKey *rsa.PublicKey, redisClient RedisInterface, logger *logger.Logger) *JWTAuthMiddleware {
	return &JWTAuthMiddleware{
		publicKey:   publicKey,
		redisClient: redisClient,
		logger:      logger,
	}
}

// AuthenticateJWT validates JWT tokens from Authorization header
func (m *JWTAuthMiddleware) AuthenticateJWT() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 1. Извлечь JWT токен из Authorization header
		tokenString := c.GetHeader("Authorization")
		if tokenString == "" {
			m.logger.Error("Missing Authorization header")
			c.JSON(http.StatusUnauthorized, gin.H{
				"error":   "missing_token",
				"message": "Missing Authorization header",
			})
			c.Abort()
			return
		}

		// Удаление префикса "Bearer "
		if strings.HasPrefix(tokenString, "Bearer ") {
			tokenString = tokenString[7:]
		} else {
			m.logger.Error("Invalid token format")
			c.JSON(http.StatusUnauthorized, gin.H{
				"error":   "invalid_token_format",
				"message": "Invalid Bearer token format",
			})
			c.Abort()
			return
		}

		// 2. Валидация JWT подписи с публичным ключом от Auth Service
		token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
			// Проверка алгоритма подписи
			if _, ok := token.Method.(*jwt.SigningMethodRSA); !ok {
				return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
			}
			return m.publicKey, nil
		})

		if err != nil {
			m.logger.Error("JWT validation failed", map[string]interface{}{"error": err.Error()})
			c.JSON(http.StatusUnauthorized, gin.H{
				"error":   "invalid_token_signature",
				"message": "Invalid JWT signature",
			})
			c.Abort()
			return
		}

		if !token.Valid {
			m.logger.Error("Token is not valid")
			c.JSON(http.StatusUnauthorized, gin.H{
				"error":   "invalid_token",
				"message": "Token is not valid",
			})
			c.Abort()
			return
		}

		// 3. Извлечение claims
		claims, ok := token.Claims.(jwt.MapClaims)
		if !ok {
			m.logger.Error("Failed to parse token claims")
			c.JSON(http.StatusUnauthorized, gin.H{
				"error":   "invalid_token_claims",
				"message": "Failed to parse token claims",
			})
			c.Abort()
			return
		}

		// 4. Проверка отзыва токена в Redis
		jti, ok := claims["jti"].(string)
		if !ok {
			m.logger.Error("Missing JTI in token")
			c.JSON(http.StatusUnauthorized, gin.H{
				"error":   "missing_token_id",
				"message": "Missing JTI in token",
			})
			c.Abort()
			return
		}

		// Проверка в Redis: EXISTS revoked:{jti}
		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer cancel()

		revokedKey := "revoked:" + jti
		isRevoked := m.redisClient.Exists(ctx, revokedKey).Val()
		if isRevoked > 0 {
			m.logger.Error("Token has been revoked", map[string]interface{}{"jti": jti})
			c.JSON(http.StatusUnauthorized, gin.H{
				"error":   "token_revoked",
				"message": "Token has been revoked",
			})
			c.Abort()
			return
		}

		// 5. Извлечение данных пользователя
		userID, ok := claims["sub"].(string)
		if !ok {
			m.logger.Error("Missing user_id in token")
			c.JSON(http.StatusUnauthorized, gin.H{
				"error":   "missing_user_id",
				"message": "Missing user_id in token claims",
			})
			c.Abort()
			return
		}

		telegramID, ok := claims["telegram_id"].(float64)
		if !ok {
			m.logger.Error("Missing telegram_id in token")
			c.JSON(http.StatusUnauthorized, gin.H{
				"error":   "missing_telegram_id",
				"message": "Missing telegram_id in token claims",
			})
			c.Abort()
			return
		}

		// 6. Сохранение в контексте запроса
		user := &UserContext{
			UserID:     userID,
			TelegramID: int64(telegramID),
		}

		c.Set("user", user)
		c.Set("jti", jti)

		m.logger.Info("User authenticated successfully", map[string]interface{}{
			"user_id":     userID,
			"telegram_id": int64(telegramID),
			"jti":         jti,
		})

		c.Next()
	}
}

// GetUserFromContext extracts user information from gin context
func GetUserFromContext(c *gin.Context) (*UserContext, bool) {
	user, exists := c.Get("user")
	if !exists {
		return nil, false
	}
	userCtx, ok := user.(*UserContext)
	return userCtx, ok
}