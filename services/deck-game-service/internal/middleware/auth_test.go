package middleware

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/json"
	"encoding/pem"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/shard-legends/deck-game-service/internal/auth"
	keyProvider "github.com/shard-legends/deck-game-service/pkg/jwt"
	"github.com/shard-legends/deck-game-service/pkg/logger"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockRedisClient is a mock implementation of RedisInterface
type MockRedisClient struct {
	mock.Mock
}

func (m *MockRedisClient) IsTokenRevoked(ctx context.Context, tokenID string) (bool, error) {
	args := m.Called(ctx, tokenID)
	return args.Bool(0), args.Error(1)
}

// MockKeyProvider is a mock implementation for testing
type MockKeyProvider struct {
	mock.Mock
}

func (m *MockKeyProvider) ValidateToken(tokenString string) (*jwt.Token, error) {
	args := m.Called(tokenString)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*jwt.Token), args.Error(1)
}

// Helper functions for testing
func generateRSAKeyPair() (*rsa.PrivateKey, *rsa.PublicKey, error) {
	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return nil, nil, err
	}
	return privateKey, &privateKey.PublicKey, nil
}

func createTestToken(privateKey *rsa.PrivateKey, claims jwt.MapClaims) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodRS256, claims)
	return token.SignedString(privateKey)
}

func marshalPublicKeyToPEM(publicKey *rsa.PublicKey) ([]byte, error) {
	pubKeyBytes, err := x509.MarshalPKIXPublicKey(publicKey)
	if err != nil {
		return nil, err
	}

	pubKeyPEM := pem.EncodeToMemory(&pem.Block{
		Type:  "PUBLIC KEY",
		Bytes: pubKeyBytes,
	})

	return pubKeyPEM, nil
}

func setupTestKeyProvider(publicKey *rsa.PublicKey) (*keyProvider.KeyProvider, error) {
	// Create a test HTTP server that serves the public key
	pubKeyPEM, err := marshalPublicKeyToPEM(publicKey)
	if err != nil {
		return nil, err
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/x-pem-file")
		w.Write(pubKeyPEM)
	}))

	kp := keyProvider.NewKeyProvider(server.URL)
	return kp, nil
}

func TestJWTAuthMiddleware_AuthenticateJWT_Success(t *testing.T) {
	// Setup
	gin.SetMode(gin.TestMode)

	privateKey, publicKey, err := generateRSAKeyPair()
	assert.NoError(t, err)

	kp, err := setupTestKeyProvider(publicKey)
	assert.NoError(t, err)

	mockRedis := new(MockRedisClient)
	testLogger := logger.NewLogger("test")

	middleware := NewJWTAuthMiddleware(kp, mockRedis, testLogger)

	// Create valid token
	claims := jwt.MapClaims{
		"sub":         "user123",
		"telegram_id": float64(987654321),
		"jti":         "token123",
		"exp":         time.Now().Add(time.Hour).Unix(),
		"iat":         time.Now().Unix(),
	}

	tokenString, err := createTestToken(privateKey, claims)
	assert.NoError(t, err)

	// Mock Redis response
	mockRedis.On("IsTokenRevoked", mock.Anything, "token123").Return(false, nil)

	// Setup HTTP request
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest("GET", "/test", nil)
	c.Request.Header.Set("Authorization", "Bearer "+tokenString)

	// Setup route handler
	authenticated := false
	testHandler := func(c *gin.Context) {
		user, exists := c.Get("user")
		assert.True(t, exists)
		assert.IsType(t, &auth.UserContext{}, user)

		userCtx := user.(*auth.UserContext)
		assert.Equal(t, "user123", userCtx.UserID)
		assert.Equal(t, int64(987654321), userCtx.TelegramID)

		jti, exists := c.Get("jti")
		assert.True(t, exists)
		assert.Equal(t, "token123", jti)

		authenticated = true
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	}

	// Execute middleware and handler
	middleware.AuthenticateJWT()(c)
	if !c.IsAborted() {
		testHandler(c)
	}

	// Assertions
	assert.True(t, authenticated)
	assert.Equal(t, http.StatusOK, w.Code)
	mockRedis.AssertExpectations(t)
}

func TestJWTAuthMiddleware_AuthenticateJWT_MissingAuthHeader(t *testing.T) {
	// Setup
	gin.SetMode(gin.TestMode)

	mockRedis := new(MockRedisClient)
	testLogger := logger.NewLogger("test")

	// Use mock key provider for this test
	mockKP := new(MockKeyProvider)
	middleware := NewJWTAuthMiddleware(mockKP, mockRedis, testLogger)

	// Setup HTTP request without Authorization header
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest("GET", "/test", nil)

	// Execute middleware
	middleware.AuthenticateJWT()(c)

	// Assertions
	assert.True(t, c.IsAborted())
	assert.Equal(t, http.StatusUnauthorized, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, "missing_token", response["error"])
}

func TestJWTAuthMiddleware_AuthenticateJWT_InvalidTokenFormat(t *testing.T) {
	// Setup
	gin.SetMode(gin.TestMode)

	mockRedis := new(MockRedisClient)
	testLogger := logger.NewLogger("test")

	mockKP := new(MockKeyProvider)
	middleware := NewJWTAuthMiddleware(mockKP, mockRedis, testLogger)

	// Setup HTTP request with invalid token format
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest("GET", "/test", nil)
	c.Request.Header.Set("Authorization", "InvalidFormat token123")

	// Execute middleware
	middleware.AuthenticateJWT()(c)

	// Assertions
	assert.True(t, c.IsAborted())
	assert.Equal(t, http.StatusUnauthorized, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, "invalid_token_format", response["error"])
}

func TestJWTAuthMiddleware_AuthenticateJWT_TokenRevoked(t *testing.T) {
	// Setup
	gin.SetMode(gin.TestMode)

	privateKey, publicKey, err := generateRSAKeyPair()
	assert.NoError(t, err)

	kp, err := setupTestKeyProvider(publicKey)
	assert.NoError(t, err)

	mockRedis := new(MockRedisClient)
	testLogger := logger.NewLogger("test")

	middleware := NewJWTAuthMiddleware(kp, mockRedis, testLogger)

	// Create valid token
	claims := jwt.MapClaims{
		"sub":         "user123",
		"telegram_id": float64(987654321),
		"jti":         "revoked_token",
		"exp":         time.Now().Add(time.Hour).Unix(),
		"iat":         time.Now().Unix(),
	}

	tokenString, err := createTestToken(privateKey, claims)
	assert.NoError(t, err)

	// Mock Redis response - token is revoked
	mockRedis.On("IsTokenRevoked", mock.Anything, "revoked_token").Return(true, nil)

	// Setup HTTP request
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest("GET", "/test", nil)
	c.Request.Header.Set("Authorization", "Bearer "+tokenString)

	// Execute middleware
	middleware.AuthenticateJWT()(c)

	// Assertions
	assert.True(t, c.IsAborted())
	assert.Equal(t, http.StatusUnauthorized, w.Code)

	var response map[string]interface{}
	err = json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, "token_revoked", response["error"])

	mockRedis.AssertExpectations(t)
}

func TestJWTAuthMiddleware_AuthenticateJWT_MissingClaims(t *testing.T) {
	// Setup
	gin.SetMode(gin.TestMode)

	privateKey, publicKey, err := generateRSAKeyPair()
	assert.NoError(t, err)

	kp, err := setupTestKeyProvider(publicKey)
	assert.NoError(t, err)

	mockRedis := new(MockRedisClient)
	testLogger := logger.NewLogger("test")

	middleware := NewJWTAuthMiddleware(kp, mockRedis, testLogger)

	// Create token with missing claims
	claims := jwt.MapClaims{
		"exp": time.Now().Add(time.Hour).Unix(),
		"iat": time.Now().Unix(),
		// Missing: sub, telegram_id, jti
	}

	tokenString, err := createTestToken(privateKey, claims)
	assert.NoError(t, err)

	// Setup HTTP request
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest("GET", "/test", nil)
	c.Request.Header.Set("Authorization", "Bearer "+tokenString)

	// Execute middleware
	middleware.AuthenticateJWT()(c)

	// Assertions
	assert.True(t, c.IsAborted())
	assert.Equal(t, http.StatusUnauthorized, w.Code)

	var response map[string]interface{}
	err = json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, "missing_token_id", response["error"])
}
