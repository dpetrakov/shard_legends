package middleware

import (
	"context"
	"crypto/rsa"
	"fmt"
	"log/slog"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

// ServiceJWTAuthMiddleware provides JWT authentication for internal service endpoints
type ServiceJWTAuthMiddleware struct {
	publicKey   *rsa.PublicKey
	redisClient RedisInterface
	logger      *slog.Logger
}

// NewServiceJWTAuthMiddleware creates a new service JWT authentication middleware
func NewServiceJWTAuthMiddleware(publicKey *rsa.PublicKey, redisClient RedisInterface, logger *slog.Logger) *ServiceJWTAuthMiddleware {
	return &ServiceJWTAuthMiddleware{
		publicKey:   publicKey,
		redisClient: redisClient,
		logger:      logger,
	}
}

// AuthenticateServiceJWT validates service JWT tokens with 'internal' role
func (m *ServiceJWTAuthMiddleware) AuthenticateServiceJWT() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Extract JWT token from Authorization header
		tokenString := c.GetHeader("Authorization")
		if tokenString == "" {
			m.logger.Error("Missing Authorization header for service endpoint")
			c.JSON(http.StatusUnauthorized, gin.H{
				"error":   "missing_service_token",
				"message": "Missing Authorization header for service endpoint",
			})
			c.Abort()
			return
		}

		// Remove "Bearer " prefix
		if strings.HasPrefix(tokenString, "Bearer ") {
			tokenString = tokenString[7:]
		} else {
			m.logger.Error("Invalid service token format", "token_prefix", tokenString[:min(10, len(tokenString))])
			c.JSON(http.StatusUnauthorized, gin.H{
				"error":   "invalid_service_token_format",
				"message": "Invalid Bearer token format for service endpoint",
			})
			c.Abort()
			return
		}

		// Validate JWT signature with public key from Auth Service
		token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
			// Check signing algorithm
			if _, ok := token.Method.(*jwt.SigningMethodRSA); !ok {
				return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
			}
			return m.publicKey, nil
		})

		if err != nil {
			m.logger.Error("Service JWT validation failed", "error", err)
			c.JSON(http.StatusUnauthorized, gin.H{
				"error":   "invalid_service_token_signature",
				"message": "Invalid service JWT signature",
			})
			c.Abort()
			return
		}

		if !token.Valid {
			m.logger.Error("Service token is not valid")
			c.JSON(http.StatusUnauthorized, gin.H{
				"error":   "invalid_service_token",
				"message": "Service token is not valid",
			})
			c.Abort()
			return
		}

		// Extract claims
		claims, ok := token.Claims.(jwt.MapClaims)
		if !ok {
			m.logger.Error("Failed to parse service token claims")
			c.JSON(http.StatusUnauthorized, gin.H{
				"error":   "invalid_service_token_claims",
				"message": "Failed to parse service token claims",
			})
			c.Abort()
			return
		}

		// Check for 'internal' role in the token
		roles, ok := claims["roles"].([]interface{})
		if !ok {
			m.logger.Error("Missing roles in service token")
			c.JSON(http.StatusForbidden, gin.H{
				"error":   "missing_service_roles",
				"message": "Missing roles in service token",
			})
			c.Abort()
			return
		}

		hasInternalRole := false
		for _, role := range roles {
			if roleStr, ok := role.(string); ok && roleStr == "internal" {
				hasInternalRole = true
				break
			}
		}

		if !hasInternalRole {
			m.logger.Error("Service token does not have 'internal' role", "roles", roles)
			c.JSON(http.StatusForbidden, gin.H{
				"error":   "insufficient_service_permissions",
				"message": "Service token does not have required 'internal' role",
			})
			c.Abort()
			return
		}

		// Check token revocation in Redis
		jti, ok := claims["jti"].(string)
		if !ok {
			m.logger.Error("Missing JTI in service token")
			c.JSON(http.StatusUnauthorized, gin.H{
				"error":   "missing_service_token_id",
				"message": "Missing JTI in service token",
			})
			c.Abort()
			return
		}

		// Check in Redis: EXISTS revoked:{jti}
		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer cancel()

		isRevoked, err := m.redisClient.IsJWTRevoked(ctx, jti)
		if err != nil {
			m.logger.Warn("Failed to check service token revocation in Redis", "jti", jti, "error", err)
			// Don't fail the request if Redis is unavailable, just log the warning
		} else if isRevoked {
			m.logger.Error("Service token has been revoked", "jti", jti)
			c.JSON(http.StatusUnauthorized, gin.H{
				"error":   "service_token_revoked",
				"message": "Service token has been revoked",
			})
			c.Abort()
			return
		}

		// Extract service information
		serviceName, _ := claims["service"].(string)
		if serviceName == "" {
			serviceName = "unknown"
		}

		// Save service context
		c.Set("service_name", serviceName)
		c.Set("service_jti", jti)
		c.Set("service_roles", roles)

		m.logger.Info("Service authenticated successfully",
			"service", serviceName,
			"jti", jti,
			"roles", roles)

		c.Next()
	}
}
