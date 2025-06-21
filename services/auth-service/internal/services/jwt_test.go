package services

import (
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// setupTestJWTService creates a JWT service for testing with temporary key files
func setupTestJWTService(t *testing.T) (*JWTService, string) {
	t.Helper()

	// Create temporary directory for test keys
	tempDir := t.TempDir()

	keyPaths := KeyPaths{
		PrivateKeyPath: filepath.Join(tempDir, "test_private.pem"),
		PublicKeyPath:  filepath.Join(tempDir, "test_public.pem"),
	}

	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}))

	service, err := NewJWTService(keyPaths, "test-issuer", 24, logger)
	if err != nil {
		t.Fatalf("Failed to create JWT service: %v", err)
	}

	return service, tempDir
}

func TestNewJWTService(t *testing.T) {
	tests := []struct {
		name        string
		issuer      string
		expiryHours int
		wantErr     bool
	}{
		{
			name:        "valid parameters",
			issuer:      "test-issuer",
			expiryHours: 24,
			wantErr:     false,
		},
		{
			name:        "empty issuer",
			issuer:      "",
			expiryHours: 24,
			wantErr:     false, // Empty issuer is allowed
		},
		{
			name:        "zero expiry hours",
			issuer:      "test-issuer",
			expiryHours: 0,
			wantErr:     false, // Zero expiry is allowed (means no expiry)
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tempDir := t.TempDir()
			keyPaths := KeyPaths{
				PrivateKeyPath: filepath.Join(tempDir, "private.pem"),
				PublicKeyPath:  filepath.Join(tempDir, "public.pem"),
			}

			logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}))

			service, err := NewJWTService(keyPaths, tt.issuer, tt.expiryHours, logger)

			if tt.wantErr {
				if err == nil {
					t.Error("Expected error but got none")
				}
				return
			}

			if err != nil {
				t.Errorf("Unexpected error: %v", err)
				return
			}

			if service == nil {
				t.Error("Expected service but got nil")
				return
			}

			// Verify service properties
			if service.issuer != tt.issuer {
				t.Errorf("Expected issuer %q, got %q", tt.issuer, service.issuer)
			}

			if service.expiryHours != tt.expiryHours {
				t.Errorf("Expected expiryHours %d, got %d", tt.expiryHours, service.expiryHours)
			}

			// Verify keys were generated
			if service.privateKey == nil {
				t.Error("Private key should not be nil")
			}

			if service.publicKey == nil {
				t.Error("Public key should not be nil")
			}

			// Verify key files were created
			if _, err := os.Stat(keyPaths.PrivateKeyPath); os.IsNotExist(err) {
				t.Error("Private key file was not created")
			}

			if _, err := os.Stat(keyPaths.PublicKeyPath); os.IsNotExist(err) {
				t.Error("Public key file was not created")
			}
		})
	}
}

func TestGenerateToken(t *testing.T) {
	service, _ := setupTestJWTService(t)

	tests := []struct {
		name       string
		telegramID int64
		wantErr    bool
	}{
		{
			name:       "valid telegram ID",
			telegramID: 12345678,
			wantErr:    false,
		},
		{
			name:       "another valid telegram ID",
			telegramID: 98765432,
			wantErr:    false,
		},
		{
			name:       "zero telegram ID",
			telegramID: 0,
			wantErr:    true, // Zero is invalid - token will be generated but validation will fail
		},
		{
			name:       "negative telegram ID",
			telegramID: -12345,
			wantErr:    true, // Negative is invalid - token will be generated but validation will fail
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			token, err := service.GenerateToken(tt.telegramID)

			// Token generation should always succeed
			if err != nil {
				t.Errorf("Unexpected error during token generation: %v", err)
				return
			}

			if token == "" {
				t.Error("Expected non-empty token")
				return
			}

			// Verify token format (should have 3 parts separated by dots)
			parts := strings.Split(token, ".")
			if len(parts) != 3 {
				t.Errorf("Expected 3 token parts, got %d", len(parts))
			}

			// Verify we can parse the token
			claims, err := service.ValidateToken(token)

			if tt.wantErr {
				// For invalid telegram IDs, validation should fail
				if err == nil {
					t.Error("Expected validation error but got none")
				}
				return
			}

			if err != nil {
				t.Errorf("Failed to validate generated token: %v", err)
				return
			}

			// Verify claims
			if claims.TelegramID != tt.telegramID {
				t.Errorf("Expected telegram_id %d, got %d", tt.telegramID, claims.TelegramID)
			}

			if claims.Issuer != service.issuer {
				t.Errorf("Expected issuer %q, got %q", service.issuer, claims.Issuer)
			}

			if claims.JTI == "" {
				t.Error("Expected non-empty JTI")
			}

			// Verify expiration time
			if claims.ExpiresAt == nil {
				t.Error("Expected expiration time")
			} else {
				expectedExpiry := time.Now().Add(time.Duration(service.expiryHours) * time.Hour)
				timeDiff := claims.ExpiresAt.Time.Sub(expectedExpiry).Abs()
				if timeDiff > time.Minute {
					t.Errorf("Expiration time difference too large: %v", timeDiff)
				}
			}
		})
	}
}

func TestValidateToken(t *testing.T) {
	service, _ := setupTestJWTService(t)

	// Generate a valid token first
	telegramID := int64(12345678)
	validToken, err := service.GenerateToken(telegramID)
	if err != nil {
		t.Fatalf("Failed to generate test token: %v", err)
	}

	// Create an expired token for testing
	expiredClaims := &JWTClaims{
		TelegramID: telegramID,
		JTI:        "test-jti",
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    service.issuer,
			Subject:   "12345678",
			IssuedAt:  jwt.NewNumericDate(time.Now().Add(-2 * time.Hour)),
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(-1 * time.Hour)), // Expired 1 hour ago
		},
	}
	expiredToken := jwt.NewWithClaims(jwt.SigningMethodRS256, expiredClaims)
	expiredTokenString, err := expiredToken.SignedString(service.privateKey)
	if err != nil {
		t.Fatalf("Failed to create expired token: %v", err)
	}

	tests := []struct {
		name    string
		token   string
		wantErr bool
	}{
		{
			name:    "valid token",
			token:   validToken,
			wantErr: false,
		},
		{
			name:    "empty token",
			token:   "",
			wantErr: true,
		},
		{
			name:    "malformed token",
			token:   "invalid.token.format",
			wantErr: true,
		},
		{
			name:    "expired token",
			token:   expiredTokenString,
			wantErr: true,
		},
		{
			name:    "token with invalid signature",
			token:   validToken[:len(validToken)-10] + "tamperedxx",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			claims, err := service.ValidateToken(tt.token)

			if tt.wantErr {
				if err == nil {
					t.Error("Expected error but got none")
				}
				return
			}

			if err != nil {
				t.Errorf("Unexpected error: %v", err)
				return
			}

			if claims == nil {
				t.Error("Expected claims but got nil")
				return
			}

			// Verify claims for valid tokens
			if tt.token == validToken {
				if claims.TelegramID != telegramID {
					t.Errorf("Expected telegram_id %d, got %d", telegramID, claims.TelegramID)
				}

				if claims.Issuer != service.issuer {
					t.Errorf("Expected issuer %q, got %q", service.issuer, claims.Issuer)
				}
			}
		})
	}
}

func TestValidateClaims(t *testing.T) {
	service, _ := setupTestJWTService(t)

	tests := []struct {
		name    string
		claims  *JWTClaims
		wantErr bool
	}{
		{
			name: "valid claims",
			claims: &JWTClaims{
				TelegramID: 12345678,
				JTI:        "test-jti",
				RegisteredClaims: jwt.RegisteredClaims{
					Issuer:  service.issuer,
					Subject: "12345678",
				},
			},
			wantErr: false,
		},
		{
			name: "invalid issuer",
			claims: &JWTClaims{
				TelegramID: 12345678,
				JTI:        "test-jti",
				RegisteredClaims: jwt.RegisteredClaims{
					Issuer:  "wrong-issuer",
					Subject: "12345678",
				},
			},
			wantErr: true,
		},
		{
			name: "invalid telegram_id (zero)",
			claims: &JWTClaims{
				TelegramID: 0,
				JTI:        "test-jti",
				RegisteredClaims: jwt.RegisteredClaims{
					Issuer:  service.issuer,
					Subject: "0",
				},
			},
			wantErr: true,
		},
		{
			name: "invalid telegram_id (negative)",
			claims: &JWTClaims{
				TelegramID: -12345,
				JTI:        "test-jti",
				RegisteredClaims: jwt.RegisteredClaims{
					Issuer:  service.issuer,
					Subject: "-12345",
				},
			},
			wantErr: true,
		},
		{
			name: "missing JTI",
			claims: &JWTClaims{
				TelegramID: 12345678,
				JTI:        "",
				RegisteredClaims: jwt.RegisteredClaims{
					Issuer:  service.issuer,
					Subject: "12345678",
				},
			},
			wantErr: true,
		},
		{
			name: "subject mismatch",
			claims: &JWTClaims{
				TelegramID: 12345678,
				JTI:        "test-jti",
				RegisteredClaims: jwt.RegisteredClaims{
					Issuer:  service.issuer,
					Subject: "87654321", // Different from telegram_id
				},
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := service.validateClaims(tt.claims)

			if tt.wantErr {
				if err == nil {
					t.Error("Expected error but got none")
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error: %v", err)
				}
			}
		})
	}
}

func TestGetPublicKeyPEM(t *testing.T) {
	service, _ := setupTestJWTService(t)

	publicKeyPEM, err := service.GetPublicKeyPEM()
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
		return
	}

	if publicKeyPEM == "" {
		t.Error("Expected non-empty public key PEM")
		return
	}

	// Verify PEM format
	if !strings.Contains(publicKeyPEM, "-----BEGIN PUBLIC KEY-----") {
		t.Error("Expected PEM to contain BEGIN PUBLIC KEY header")
	}

	if !strings.Contains(publicKeyPEM, "-----END PUBLIC KEY-----") {
		t.Error("Expected PEM to contain END PUBLIC KEY footer")
	}
}

func TestGetKeyID(t *testing.T) {
	service, _ := setupTestJWTService(t)

	keyID, err := service.GetKeyID()
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
		return
	}

	if keyID == "" {
		t.Error("Expected non-empty key ID")
		return
	}

	// Verify key ID format (should be hex string)
	if len(keyID) != 16 { // 8 bytes = 16 hex characters
		t.Errorf("Expected key ID length 16, got %d", len(keyID))
	}
}

func TestKeyGeneration(t *testing.T) {
	tempDir := t.TempDir()

	keyPaths := KeyPaths{
		PrivateKeyPath: filepath.Join(tempDir, "private.pem"),
		PublicKeyPath:  filepath.Join(tempDir, "public.pem"),
	}

	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}))

	// First service - should generate keys
	service1, err := NewJWTService(keyPaths, "test-issuer", 24, logger)
	if err != nil {
		t.Fatalf("Failed to create first JWT service: %v", err)
	}

	// Verify keys were generated and are RSA 2048-bit
	if service1.privateKey.N.BitLen() != 2048 {
		t.Errorf("Expected 2048-bit RSA key, got %d-bit", service1.privateKey.N.BitLen())
	}

	// Second service - should load existing keys
	service2, err := NewJWTService(keyPaths, "test-issuer", 24, logger)
	if err != nil {
		t.Fatalf("Failed to create second JWT service: %v", err)
	}

	// Verify keys are the same
	if service1.privateKey.N.Cmp(service2.privateKey.N) != 0 {
		t.Error("Private keys should be identical when loaded from same files")
	}

	if service1.publicKey.N.Cmp(service2.publicKey.N) != 0 {
		t.Error("Public keys should be identical when loaded from same files")
	}
}

func TestKeyFilePermissions(t *testing.T) {
	tempDir := t.TempDir()

	keyPaths := KeyPaths{
		PrivateKeyPath: filepath.Join(tempDir, "private.pem"),
		PublicKeyPath:  filepath.Join(tempDir, "public.pem"),
	}

	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}))

	_, err := NewJWTService(keyPaths, "test-issuer", 24, logger)
	if err != nil {
		t.Fatalf("Failed to create JWT service: %v", err)
	}

	// Check private key file permissions (should be 0600)
	privateInfo, err := os.Stat(keyPaths.PrivateKeyPath)
	if err != nil {
		t.Fatalf("Failed to stat private key file: %v", err)
	}

	if privateInfo.Mode().Perm() != 0600 {
		t.Errorf("Expected private key file permissions 0600, got %o", privateInfo.Mode().Perm())
	}

	// Check public key file permissions (should be 0644)
	publicInfo, err := os.Stat(keyPaths.PublicKeyPath)
	if err != nil {
		t.Fatalf("Failed to stat public key file: %v", err)
	}

	if publicInfo.Mode().Perm() != 0644 {
		t.Errorf("Expected public key file permissions 0644, got %o", publicInfo.Mode().Perm())
	}
}

func TestTokenRoundTrip(t *testing.T) {
	service, _ := setupTestJWTService(t)

	telegramIDs := []int64{1, 12345678, 9876543210}

	for _, telegramID := range telegramIDs {
		t.Run(fmt.Sprintf("telegram_id_%d", telegramID), func(t *testing.T) {
			// Generate token
			token, err := service.GenerateToken(telegramID)
			if err != nil {
				t.Fatalf("Failed to generate token: %v", err)
			}

			// Validate token
			claims, err := service.ValidateToken(token)
			if err != nil {
				t.Fatalf("Failed to validate token: %v", err)
			}

			// Verify round-trip integrity
			if claims.TelegramID != telegramID {
				t.Errorf("Token round-trip failed: expected telegram_id %d, got %d", telegramID, claims.TelegramID)
			}

			// Verify all required claims are present
			if claims.Issuer == "" {
				t.Error("Issuer claim is missing")
			}

			if claims.Subject == "" {
				t.Error("Subject claim is missing")
			}

			if claims.JTI == "" {
				t.Error("JTI claim is missing")
			}

			if claims.IssuedAt == nil {
				t.Error("IssuedAt claim is missing")
			}

			if claims.ExpiresAt == nil {
				t.Error("ExpiresAt claim is missing")
			}
		})
	}
}

// Benchmark tests
func BenchmarkGenerateToken(b *testing.B) {
	service, _ := setupTestJWTService(&testing.T{})
	telegramID := int64(12345678)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := service.GenerateToken(telegramID)
		if err != nil {
			b.Fatalf("Failed to generate token: %v", err)
		}
	}
}

func BenchmarkValidateToken(b *testing.B) {
	service, _ := setupTestJWTService(&testing.T{})
	telegramID := int64(12345678)

	// Generate a token to validate
	token, err := service.GenerateToken(telegramID)
	if err != nil {
		b.Fatalf("Failed to generate test token: %v", err)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := service.ValidateToken(token)
		if err != nil {
			b.Fatalf("Failed to validate token: %v", err)
		}
	}
}
