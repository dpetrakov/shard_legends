package test

import (
	"bytes"
	"crypto/rand"
	"crypto/rsa"
	"encoding/json"
	"net/http"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestServerSeparation проверяет разделение эндпоинтов между публичным и внутренним серверами
func TestServerSeparation(t *testing.T) {
	// Предполагаем, что тестовый инстанс запущен локально
	publicBaseURL := "http://localhost:8080"
	internalBaseURL := "http://localhost:8081"

	// Генерируем тестовые ключи для JWT
	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	require.NoError(t, err)

	tests := []struct {
		name           string
		endpoint       string
		method         string
		publicServer   bool
		internalServer bool
		requiresJWT    bool
		requiresServiceJWT bool
		payload        interface{}
	}{
		{
			name:           "Health endpoint on internal server",
			endpoint:       "/health",
			method:         "GET",
			publicServer:   false,
			internalServer: true,
			requiresJWT:    false,
			requiresServiceJWT: false,
		},
		{
			name:           "Metrics endpoint on internal server",
			endpoint:       "/metrics",
			method:         "GET",
			publicServer:   false,
			internalServer: true,
			requiresJWT:    false,
			requiresServiceJWT: false,
		},
		{
			name:           "Get user inventory on public server",
			endpoint:       "/api/inventory",
			method:         "GET",
			publicServer:   true,
			internalServer: false,
			requiresJWT:    true,
			requiresServiceJWT: false,
		},
		{
			name:           "Reserve items on internal server only",
			endpoint:       "/api/inventory/reserve",
			method:         "POST",
			publicServer:   false,
			internalServer: true,
			requiresJWT:    false,
			requiresServiceJWT: true,
			payload: map[string]interface{}{
				"user_id":      "123e4567-e89b-12d3-a456-426614174000",
				"operation_id": "123e4567-e89b-12d3-a456-426614174001",
				"items": []map[string]interface{}{
					{
						"item_id":  "123e4567-e89b-12d3-a456-426614174002",
						"quantity": 1,
					},
				},
			},
		},
		{
			name:           "Return reserved items on internal server only",
			endpoint:       "/api/inventory/return-reserve",
			method:         "POST",
			publicServer:   false,
			internalServer: true,
			requiresJWT:    false,
			requiresServiceJWT: true,
			payload: map[string]interface{}{
				"user_id":      "123e4567-e89b-12d3-a456-426614174000",
				"operation_id": "123e4567-e89b-12d3-a456-426614174001",
			},
		},
		{
			name:           "Consume reserved items on internal server only",
			endpoint:       "/api/inventory/consume-reserve",
			method:         "POST",
			publicServer:   false,
			internalServer: true,
			requiresJWT:    false,
			requiresServiceJWT: true,
			payload: map[string]interface{}{
				"user_id":      "123e4567-e89b-12d3-a456-426614174000",
				"operation_id": "123e4567-e89b-12d3-a456-426614174001",
			},
		},
		{
			name:           "Add items on internal server only",
			endpoint:       "/api/inventory/add-items",
			method:         "POST",
			publicServer:   false,
			internalServer: true,
			requiresJWT:    false,
			requiresServiceJWT: true,
			payload: map[string]interface{}{
				"user_id":        "123e4567-e89b-12d3-a456-426614174000",
				"section":        "main",
				"operation_type": "chest_reward",
				"operation_id":   "123e4567-e89b-12d3-a456-426614174003",
				"items": []map[string]interface{}{
					{
						"item_id":  "123e4567-e89b-12d3-a456-426614174002",
						"quantity": 1,
					},
				},
			},
		},
		{
			name:           "Admin adjust on public server",
			endpoint:       "/api/inventory/admin/adjust",
			method:         "POST",
			publicServer:   true,
			internalServer: false,
			requiresJWT:    true,
			requiresServiceJWT: false,
			payload: map[string]interface{}{
				"user_id": "123e4567-e89b-12d3-a456-426614174000",
				"section": "main",
				"items": []map[string]interface{}{
					{
						"item_id":         "123e4567-e89b-12d3-a456-426614174002",
						"quantity_change": 1,
					},
				},
				"reason": "Test adjustment for E2E test",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Тест на публичном сервере
			if tt.publicServer {
				t.Run("Public server accessible", func(t *testing.T) {
					url := publicBaseURL + tt.endpoint
					var token string
					
					if tt.requiresJWT {
						token = createUserJWT(t, privateKey)
					}
					
					resp, err := makeRequest(t, tt.method, url, tt.payload, token)
					require.NoError(t, err)
					defer resp.Body.Close()

					// Эндпоинт должен быть доступен (не 404)
					assert.NotEqual(t, http.StatusNotFound, resp.StatusCode, 
						"Endpoint %s should be accessible on public server", tt.endpoint)
					
					if tt.requiresJWT && token == "" {
						// Без JWT должен возвращать 401
						assert.Equal(t, http.StatusUnauthorized, resp.StatusCode,
							"Endpoint %s should require authentication on public server", tt.endpoint)
					}
				})
			} else {
				t.Run("Public server NOT accessible", func(t *testing.T) {
					url := publicBaseURL + tt.endpoint
					resp, err := makeRequest(t, tt.method, url, tt.payload, "")
					require.NoError(t, err)
					defer resp.Body.Close()

					// Эндпоинт не должен быть доступен на публичном сервере
					assert.Equal(t, http.StatusNotFound, resp.StatusCode,
						"Endpoint %s should NOT be accessible on public server", tt.endpoint)
				})
			}

			// Тест на внутреннем сервере
			if tt.internalServer {
				t.Run("Internal server accessible", func(t *testing.T) {
					url := internalBaseURL + tt.endpoint
					var token string
					
					if tt.requiresServiceJWT {
						token = createServiceJWT(t, privateKey)
					} else if tt.requiresJWT {
						token = createUserJWT(t, privateKey)
					}
					
					resp, err := makeRequest(t, tt.method, url, tt.payload, token)
					require.NoError(t, err)
					defer resp.Body.Close()

					// Эндпоинт должен быть доступен (не 404)
					assert.NotEqual(t, http.StatusNotFound, resp.StatusCode,
						"Endpoint %s should be accessible on internal server", tt.endpoint)
					
					if tt.requiresServiceJWT && token == "" {
						// Без Service JWT должен возвращать 401
						assert.Equal(t, http.StatusUnauthorized, resp.StatusCode,
							"Endpoint %s should require service authentication on internal server", tt.endpoint)
					}
				})
			} else {
				t.Run("Internal server NOT accessible", func(t *testing.T) {
					url := internalBaseURL + tt.endpoint
					resp, err := makeRequest(t, tt.method, url, tt.payload, "")
					require.NoError(t, err)
					defer resp.Body.Close()

					// Эндпоинт не должен быть доступен на внутреннем сервере
					assert.Equal(t, http.StatusNotFound, resp.StatusCode,
						"Endpoint %s should NOT be accessible on internal server", tt.endpoint)
				})
			}
		})
	}
}

// TestServiceJWTWithoutInternalRole проверяет, что Service JWT без роли 'internal' отклоняется
func TestServiceJWTWithoutInternalRole(t *testing.T) {
	internalBaseURL := "http://localhost:8081"
	
	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	require.NoError(t, err)

	// Создаем Service JWT без роли 'internal'
	claims := jwt.MapClaims{
		"sub":     "service-test",
		"service": "test-service",
		"roles":   []interface{}{"service", "user"}, // Нет роли 'internal'
		"jti":     "test-jti-no-internal",
		"exp":     time.Now().Add(time.Hour).Unix(),
		"iat":     time.Now().Unix(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodRS256, claims)
	tokenString, err := token.SignedString(privateKey)
	require.NoError(t, err)

	// Пытаемся обратиться к внутреннему эндпоинту
	payload := map[string]interface{}{
		"user_id":      "123e4567-e89b-12d3-a456-426614174000",
		"operation_id": "123e4567-e89b-12d3-a456-426614174001",
		"items": []map[string]interface{}{
			{
				"item_id":  "123e4567-e89b-12d3-a456-426614174002",
				"quantity": 1,
			},
		},
	}

	url := internalBaseURL + "/api/inventory/reserve"
	resp, err := makeRequest(t, "POST", url, payload, "Bearer "+tokenString)
	require.NoError(t, err)
	defer resp.Body.Close()

	// Должен возвращать 403 Forbidden
	assert.Equal(t, http.StatusForbidden, resp.StatusCode,
		"Service JWT without 'internal' role should be rejected with 403")
}

// createUserJWT создает JWT токен для пользователя
func createUserJWT(t *testing.T, privateKey *rsa.PrivateKey) string {
	claims := jwt.MapClaims{
		"sub":         "user-123",
		"telegram_id": float64(123456789),
		"jti":         "test-user-jti",
		"exp":         time.Now().Add(time.Hour).Unix(),
		"iat":         time.Now().Unix(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodRS256, claims)
	tokenString, err := token.SignedString(privateKey)
	require.NoError(t, err)
	return "Bearer " + tokenString
}

// createServiceJWT создает Service JWT токен с ролью 'internal'
func createServiceJWT(t *testing.T, privateKey *rsa.PrivateKey) string {
	claims := jwt.MapClaims{
		"sub":     "service-test",
		"service": "test-service",
		"roles":   []interface{}{"internal", "service"},
		"jti":     "test-service-jti",
		"exp":     time.Now().Add(time.Hour).Unix(),
		"iat":     time.Now().Unix(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodRS256, claims)
	tokenString, err := token.SignedString(privateKey)
	require.NoError(t, err)
	return "Bearer " + tokenString
}

// makeRequest выполняет HTTP запрос с опциональным JWT токеном
func makeRequest(t *testing.T, method, url string, payload interface{}, authToken string) (*http.Response, error) {
	var body *bytes.Buffer
	
	if payload != nil {
		jsonPayload, err := json.Marshal(payload)
		require.NoError(t, err)
		body = bytes.NewBuffer(jsonPayload)
	} else {
		body = &bytes.Buffer{}
	}

	req, err := http.NewRequest(method, url, body)
	require.NoError(t, err)

	if payload != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	
	if authToken != "" {
		req.Header.Set("Authorization", authToken)
	}

	client := &http.Client{Timeout: 10 * time.Second}
	return client.Do(req)
}

// TestHealthAndMetricsEndpoints проверяет доступность health и metrics эндпоинтов
func TestHealthAndMetricsEndpoints(t *testing.T) {
	internalBaseURL := "http://localhost:8081"

	t.Run("Health endpoint", func(t *testing.T) {
		resp, err := http.Get(internalBaseURL + "/health")
		require.NoError(t, err)
		defer resp.Body.Close()

		// Health должен возвращать 200 или 503 (в зависимости от состояния зависимостей)
		assert.True(t, resp.StatusCode == http.StatusOK || resp.StatusCode == http.StatusServiceUnavailable,
			"Health endpoint should return 200 or 503, got %d", resp.StatusCode)
	})

	t.Run("Metrics endpoint", func(t *testing.T) {
		resp, err := http.Get(internalBaseURL + "/metrics")
		require.NoError(t, err)
		defer resp.Body.Close()

		// Metrics должен возвращать 200 и содержать Prometheus метрики
		assert.Equal(t, http.StatusOK, resp.StatusCode,
			"Metrics endpoint should return 200")

		// Проверяем Content-Type
		contentType := resp.Header.Get("Content-Type")
		assert.Contains(t, contentType, "text/plain",
			"Metrics should return text/plain content type")
	})
}

// TestInternalEndpointsRequireServiceJWT проверяет, что внутренние эндпоинты отклоняют запросы без Service-JWT
func TestInternalEndpointsRequireServiceJWT(t *testing.T) {
	internalBaseURL := "http://localhost:8081"

	// Список внутренних эндпоинтов, которые требуют Service-JWT
	internalEndpoints := []struct {
		endpoint string
		method   string
		payload  interface{}
	}{
		{
			endpoint: "/api/inventory/reserve",
			method:   "POST",
			payload: map[string]interface{}{
				"user_id":      "123e4567-e89b-12d3-a456-426614174000",
				"operation_id": "123e4567-e89b-12d3-a456-426614174001",
				"items": []map[string]interface{}{
					{
						"item_id":  "123e4567-e89b-12d3-a456-426614174002",
						"quantity": 1,
					},
				},
			},
		},
		{
			endpoint: "/api/inventory/return-reserve",
			method:   "POST",
			payload: map[string]interface{}{
				"user_id":      "123e4567-e89b-12d3-a456-426614174000",
				"operation_id": "123e4567-e89b-12d3-a456-426614174001",
			},
		},
		{
			endpoint: "/api/inventory/consume-reserve",
			method:   "POST",
			payload: map[string]interface{}{
				"user_id":      "123e4567-e89b-12d3-a456-426614174000",
				"operation_id": "123e4567-e89b-12d3-a456-426614174001",
			},
		},
		{
			endpoint: "/api/inventory/add-items",
			method:   "POST",
			payload: map[string]interface{}{
				"user_id":        "123e4567-e89b-12d3-a456-426614174000",
				"section":        "main",
				"operation_type": "chest_reward",
				"operation_id":   "123e4567-e89b-12d3-a456-426614174003",
				"items": []map[string]interface{}{
					{
						"item_id":  "123e4567-e89b-12d3-a456-426614174002",
						"quantity": 1,
					},
				},
			},
		},
	}

	for _, endpoint := range internalEndpoints {
		t.Run(endpoint.endpoint+" without Service-JWT", func(t *testing.T) {
			url := internalBaseURL + endpoint.endpoint
			
			// Запрос без токена - должен возвращать 401
			resp, err := makeRequest(t, endpoint.method, url, endpoint.payload, "")
			require.NoError(t, err)
			defer resp.Body.Close()

			assert.Equal(t, http.StatusUnauthorized, resp.StatusCode,
				"Endpoint %s should require Service-JWT authentication", endpoint.endpoint)
		})

		t.Run(endpoint.endpoint+" with user JWT (should fail)", func(t *testing.T) {
			privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
			require.NoError(t, err)
			
			userToken := createUserJWT(t, privateKey)
			url := internalBaseURL + endpoint.endpoint
			
			// Запрос с пользовательским JWT (не Service-JWT) - должен возвращать 401 или 403
			resp, err := makeRequest(t, endpoint.method, url, endpoint.payload, userToken)
			require.NoError(t, err)
			defer resp.Body.Close()

			assert.True(t, resp.StatusCode == http.StatusUnauthorized || resp.StatusCode == http.StatusForbidden,
				"Endpoint %s should reject user JWT and require Service-JWT, got %d", endpoint.endpoint, resp.StatusCode)
		})

		t.Run(endpoint.endpoint+" with valid Service-JWT", func(t *testing.T) {
			privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
			require.NoError(t, err)
			
			serviceToken := createServiceJWT(t, privateKey)
			url := internalBaseURL + endpoint.endpoint
			
			// Запрос с валидным Service-JWT - не должен возвращать 401/403 (может быть 400 или 404 из-за некорректных данных)
			resp, err := makeRequest(t, endpoint.method, url, endpoint.payload, serviceToken)
			require.NoError(t, err)
			defer resp.Body.Close()

			assert.NotEqual(t, http.StatusUnauthorized, resp.StatusCode,
				"Endpoint %s should accept valid Service-JWT (got %d)", endpoint.endpoint, resp.StatusCode)
			assert.NotEqual(t, http.StatusForbidden, resp.StatusCode,
				"Endpoint %s should accept valid Service-JWT (got %d)", endpoint.endpoint, resp.StatusCode)
		})
	}
}

// TestCrossServerIsolation проверяет, что серверы действительно изолированы
func TestCrossServerIsolation(t *testing.T) {
	publicBaseURL := "http://localhost:8080"
	internalBaseURL := "http://localhost:8081"

	// Список эндпоинтов, которые должны быть доступны только на внутреннем сервере
	internalOnlyEndpoints := []string{
		"/api/inventory/reserve",
		"/api/inventory/return-reserve",
		"/api/inventory/consume-reserve",
		"/api/inventory/add-items",
		"/health",
		"/metrics",
	}

	// Список эндпоинтов, которые должны быть доступны только на публичном сервере
	publicOnlyEndpoints := []string{
		"/api/inventory",
		"/api/inventory/admin/adjust",
	}

	t.Run("Internal endpoints not accessible on public server", func(t *testing.T) {
		for _, endpoint := range internalOnlyEndpoints {
			t.Run(endpoint, func(t *testing.T) {
				resp, err := http.Get(publicBaseURL + endpoint)
				require.NoError(t, err)
				defer resp.Body.Close()

				assert.Equal(t, http.StatusNotFound, resp.StatusCode,
					"Internal endpoint %s should not be accessible on public server", endpoint)
			})
		}
	})

	t.Run("Public endpoints not accessible on internal server", func(t *testing.T) {
		for _, endpoint := range publicOnlyEndpoints {
			t.Run(endpoint, func(t *testing.T) {
				resp, err := http.Get(internalBaseURL + endpoint)
				require.NoError(t, err)
				defer resp.Body.Close()

				assert.Equal(t, http.StatusNotFound, resp.StatusCode,
					"Public endpoint %s should not be accessible on internal server", endpoint)
			})
		}
	})
}