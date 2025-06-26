package middleware

import (
	"net/http"
	"strings"

	"github.com/shard-legends/production-service/internal/auth"
	"github.com/shard-legends/production-service/pkg/jwt"
	"github.com/shard-legends/production-service/pkg/logger"
	"go.uber.org/zap"
)

func Auth(validator *jwt.Validator) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			authHeader := r.Header.Get("Authorization")
			if authHeader == "" {
				http.Error(w, "Missing authorization header", http.StatusUnauthorized)
				return
			}

			tokenString := strings.TrimPrefix(authHeader, "Bearer ")
			if tokenString == authHeader {
				http.Error(w, "Invalid authorization header format", http.StatusUnauthorized)
				return
			}

			claims, err := validator.ValidateToken(r.Context(), tokenString)
			if err != nil {
				logger.Debug("Token validation failed",
					zap.String("error", err.Error()),
					zap.String("path", r.URL.Path),
				)
				http.Error(w, "Invalid token", http.StatusUnauthorized)
				return
			}

			userCtx := &auth.UserContext{
				UserID:     claims.Subject,
				TelegramID: claims.TelegramID,
				IsAdmin:    false,
			}

			ctx := auth.WithUser(r.Context(), userCtx)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func AdminAuth(validator *jwt.Validator) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			authHeader := r.Header.Get("Authorization")
			if authHeader == "" {
				http.Error(w, "Missing authorization header", http.StatusUnauthorized)
				return
			}

			tokenString := strings.TrimPrefix(authHeader, "Bearer ")
			if tokenString == authHeader {
				http.Error(w, "Invalid authorization header format", http.StatusUnauthorized)
				return
			}

			claims, err := validator.ValidateToken(r.Context(), tokenString)
			if err != nil {
				logger.Debug("Admin token validation failed",
					zap.String("error", err.Error()),
					zap.String("path", r.URL.Path),
				)
				http.Error(w, "Invalid token", http.StatusUnauthorized)
				return
			}

			// TODO: Check admin role when roles are implemented
			// For now, we'll accept any valid token for admin endpoints
			userCtx := &auth.UserContext{
				UserID:     claims.Subject,
				TelegramID: claims.TelegramID,
				IsAdmin:    true,
			}

			ctx := auth.WithUser(r.Context(), userCtx)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}
