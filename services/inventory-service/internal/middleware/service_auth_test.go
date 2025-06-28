package middleware

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"fmt"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockRedisClient для тестирования
type MockRedisClient struct {
	mock.Mock
}

func (m *MockRedisClient) IsJWTRevoked(ctx context.Context, jti string) (bool, error) {
	args := m.Called(ctx, jti)
	return args.Bool(0), args.Error(1)
}

func TestServiceJWTAuthMiddleware_AuthenticateServiceJWT(t *testing.T) {
	// Генерируем тестовые RSA ключи
	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	assert.NoError(t, err)
	publicKey := &privateKey.PublicKey

	// Настраиваем логгер
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelError}))

	// Создаем мок Redis клиента
	mockRedis := new(MockRedisClient)

	// Создаем middleware
	middleware := NewServiceJWTAuthMiddleware(publicKey, mockRedis, logger)

	// Создаем Gin router для тестов
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name           string
		setupToken     func() string
		setupRedis     func()
		expectedStatus int
		expectedError  string
	}{
		{
			name: "Valid service token with internal role",
			setupToken: func() string {
				claims := jwt.MapClaims{
					"sub":     "service-production",
					"service": "production-service",
					"roles":   []interface{}{"internal", "service"},
					"jti":     "test-jti-123",
					"exp":     time.Now().Add(time.Hour).Unix(),
					"iat":     time.Now().Unix(),
				}
				token := jwt.NewWithClaims(jwt.SigningMethodRS256, claims)
				tokenString, _ := token.SignedString(privateKey)
				return "Bearer " + tokenString
			},
			setupRedis: func() {
				mockRedis.On("IsJWTRevoked", mock.Anything, "test-jti-123").Return(false, nil)
			},
			expectedStatus: http.StatusOK,
		},
		{
			name: "Missing Authorization header",
			setupToken: func() string {
				return ""
			},
			setupRedis:     func() {},
			expectedStatus: http.StatusUnauthorized,
			expectedError:  "missing_service_token",
		},
		{
			name: "Invalid Bearer format",
			setupToken: func() string {
				return "InvalidToken"
			},
			setupRedis:     func() {},
			expectedStatus: http.StatusUnauthorized,
			expectedError:  "invalid_service_token_format",
		},
		{
			name: "Token without internal role",
			setupToken: func() string {
				claims := jwt.MapClaims{
					"sub":     "service-production",
					"service": "production-service",
					"roles":   []interface{}{"service", "user"},
					"jti":     "test-jti-456",
					"exp":     time.Now().Add(time.Hour).Unix(),
					"iat":     time.Now().Unix(),
				}
				token := jwt.NewWithClaims(jwt.SigningMethodRS256, claims)
				tokenString, _ := token.SignedString(privateKey)
				return "Bearer " + tokenString
			},
			setupRedis: func() {
				// Для теста без роли 'internal' Redis не должен вызываться,
				// так как проверка роли происходит до проверки отзыва
			},
			expectedStatus: http.StatusForbidden,
			expectedError:  "insufficient_service_permissions",
		},
		{
			name: "Revoked token",
			setupToken: func() string {
				claims := jwt.MapClaims{
					"sub":     "service-production",
					"service": "production-service",
					"roles":   []interface{}{"internal"},
					"jti":     "test-jti-revoked",
					"exp":     time.Now().Add(time.Hour).Unix(),
					"iat":     time.Now().Unix(),
				}
				token := jwt.NewWithClaims(jwt.SigningMethodRS256, claims)
				tokenString, _ := token.SignedString(privateKey)
				return "Bearer " + tokenString
			},
			setupRedis: func() {
				mockRedis.On("IsJWTRevoked", mock.Anything, "test-jti-revoked").Return(true, nil)
			},
			expectedStatus: http.StatusUnauthorized,
			expectedError:  "service_token_revoked",
		},
		{
			name: "Expired token",
			setupToken: func() string {
				claims := jwt.MapClaims{
					"sub":     "service-production",
					"service": "production-service",
					"roles":   []interface{}{"internal"},
					"jti":     "test-jti-expired",
					"exp":     time.Now().Add(-time.Hour).Unix(), // Expired 1 hour ago
					"iat":     time.Now().Add(-2 * time.Hour).Unix(),
				}
				token := jwt.NewWithClaims(jwt.SigningMethodRS256, claims)
				tokenString, _ := token.SignedString(privateKey)
				return "Bearer " + tokenString
			},
			setupRedis:     func() {},
			expectedStatus: http.StatusUnauthorized,
			expectedError:  "invalid_service_token_signature",
		},
		{
			name: "Missing JTI claim",
			setupToken: func() string {
				claims := jwt.MapClaims{
					"sub":     "service-production",
					"service": "production-service",
					"roles":   []interface{}{"internal"},
					"exp":     time.Now().Add(time.Hour).Unix(),
					"iat":     time.Now().Unix(),
					// Missing jti
				}
				token := jwt.NewWithClaims(jwt.SigningMethodRS256, claims)
				tokenString, _ := token.SignedString(privateKey)
				return "Bearer " + tokenString
			},
			setupRedis:     func() {},
			expectedStatus: http.StatusUnauthorized,
			expectedError:  "missing_service_token_id",
		},
		{
			name: "Missing roles claim",
			setupToken: func() string {
				claims := jwt.MapClaims{
					"sub":     "service-production",
					"service": "production-service",
					"jti":     "test-jti-no-roles",
					"exp":     time.Now().Add(time.Hour).Unix(),
					"iat":     time.Now().Unix(),
					// Missing roles
				}
				token := jwt.NewWithClaims(jwt.SigningMethodRS256, claims)
				tokenString, _ := token.SignedString(privateKey)
				return "Bearer " + tokenString
			},
			setupRedis:     func() {},
			expectedStatus: http.StatusForbidden,
			expectedError:  "missing_service_roles",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Сбрасываем мок перед каждым тестом
			mockRedis.ExpectedCalls = nil
			mockRedis.Calls = nil

			// Настраиваем мок
			tt.setupRedis()

			// Создаем новый router для каждого теста
			router := gin.New()

			// Тестовый endpoint
			router.POST("/test", middleware.AuthenticateServiceJWT(), func(c *gin.Context) {
				// Проверяем, что контекст правильно установлен
				serviceName, exists := c.Get("service_name")
				if exists {
					assert.NotEmpty(t, serviceName)
				}

				serviceJTI, exists := c.Get("service_jti")
				if exists {
					assert.NotEmpty(t, serviceJTI)
				}

				c.JSON(http.StatusOK, gin.H{"status": "success"})
			})

			// Создаем запрос
			req := httptest.NewRequest("POST", "/test", nil)

			// Устанавливаем Authorization header если он есть
			token := tt.setupToken()
			if token != "" {
				req.Header.Set("Authorization", token)
			}

			// Выполняем запрос
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			// Проверяем результат
			assert.Equal(t, tt.expectedStatus, w.Code)

			// Если ожидается ошибка, проверяем ее наличие в ответе
			if tt.expectedError != "" {
				assert.Contains(t, w.Body.String(), tt.expectedError)
			}

			// Проверяем, что все ожидаемые вызовы мока были выполнены
			mockRedis.AssertExpectations(t)
		})
	}
}

func TestServiceJWTAuthMiddleware_ContextValues(t *testing.T) {
	// Генерируем тестовые RSA ключи
	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	assert.NoError(t, err)
	publicKey := &privateKey.PublicKey

	// Настраиваем логгер
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelError}))

	// Создаем мок Redis клиента
	mockRedis := new(MockRedisClient)
	mockRedis.On("IsJWTRevoked", mock.Anything, "test-context-jti").Return(false, nil)

	// Создаем middleware
	middleware := NewServiceJWTAuthMiddleware(publicKey, mockRedis, logger)

	// Создаем валидный токен
	claims := jwt.MapClaims{
		"sub":     "service-inventory",
		"service": "inventory-service",
		"roles":   []interface{}{"internal", "service"},
		"jti":     "test-context-jti",
		"exp":     time.Now().Add(time.Hour).Unix(),
		"iat":     time.Now().Unix(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodRS256, claims)
	tokenString, err := token.SignedString(privateKey)
	assert.NoError(t, err)

	// Создаем Gin router
	gin.SetMode(gin.TestMode)
	router := gin.New()

	// Тестовый endpoint
	router.POST("/test", middleware.AuthenticateServiceJWT(), func(c *gin.Context) {
		// Проверяем контекст
		serviceName, exists := c.Get("service_name")
		assert.True(t, exists)
		assert.Equal(t, "inventory-service", serviceName)

		serviceJTI, exists := c.Get("service_jti")
		assert.True(t, exists)
		assert.Equal(t, "test-context-jti", serviceJTI)

		serviceRoles, exists := c.Get("service_roles")
		assert.True(t, exists)
		roles := serviceRoles.([]interface{})
		assert.Contains(t, roles, "internal")
		assert.Contains(t, roles, "service")

		c.JSON(http.StatusOK, gin.H{"status": "success"})
	})

	// Создаем запрос
	req := httptest.NewRequest("POST", "/test", nil)
	req.Header.Set("Authorization", "Bearer "+tokenString)

	// Выполняем запрос
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Проверяем результат
	assert.Equal(t, http.StatusOK, w.Code)
	mockRedis.AssertExpectations(t)
}

func TestServiceJWTAuthMiddleware_RedisError(t *testing.T) {
	// Генерируем тестовые RSA ключи
	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	assert.NoError(t, err)
	publicKey := &privateKey.PublicKey

	// Настраиваем логгер
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelError}))

	// Создаем мок Redis клиента с ошибкой
	mockRedis := new(MockRedisClient)
	mockRedis.On("IsJWTRevoked", mock.Anything, "test-redis-error-jti").Return(false, fmt.Errorf("redis connection error"))

	// Создаем middleware
	middleware := NewServiceJWTAuthMiddleware(publicKey, mockRedis, logger)

	// Создаем валидный токен
	claims := jwt.MapClaims{
		"sub":     "service-test",
		"service": "test-service",
		"roles":   []interface{}{"internal"},
		"jti":     "test-redis-error-jti",
		"exp":     time.Now().Add(time.Hour).Unix(),
		"iat":     time.Now().Unix(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodRS256, claims)
	tokenString, err := token.SignedString(privateKey)
	assert.NoError(t, err)

	// Создаем Gin router
	gin.SetMode(gin.TestMode)
	router := gin.New()

	// Тестовый endpoint
	router.POST("/test", middleware.AuthenticateServiceJWT(), func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "success"})
	})

	// Создаем запрос
	req := httptest.NewRequest("POST", "/test", nil)
	req.Header.Set("Authorization", "Bearer "+tokenString)

	// Выполняем запрос
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Проверяем, что запрос прошел успешно несмотря на ошибку Redis
	// (должно только записать предупреждение в лог)
	assert.Equal(t, http.StatusOK, w.Code)
	mockRedis.AssertExpectations(t)
}
