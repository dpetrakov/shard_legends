package services

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

// JWTClaims represents the custom claims for JWT tokens
type JWTClaims struct {
	TelegramID int64  `json:"telegram_id"`
	JTI        string `json:"jti"`
	jwt.RegisteredClaims
}

// JWTService handles JWT token generation and validation
type JWTService struct {
	privateKey  *rsa.PrivateKey
	publicKey   *rsa.PublicKey
	issuer      string
	expiryHours int
	logger      *slog.Logger
	keyPaths    KeyPaths
}

// KeyPaths holds the paths to private and public key files
type KeyPaths struct {
	PrivateKeyPath string
	PublicKeyPath  string
}

// NewJWTService creates a new JWT service instance
func NewJWTService(keyPaths KeyPaths, issuer string, expiryHours int, logger *slog.Logger) (*JWTService, error) {
	service := &JWTService{
		issuer:      issuer,
		expiryHours: expiryHours,
		logger:      logger,
		keyPaths:    keyPaths,
	}

	// Load or generate RSA keys
	if err := service.loadOrGenerateKeys(); err != nil {
		return nil, fmt.Errorf("failed to load or generate RSA keys: %w", err)
	}

	logger.Info("JWT Service initialized",
		"issuer", issuer,
		"expiry_hours", expiryHours,
		"private_key_path", keyPaths.PrivateKeyPath,
		"public_key_path", keyPaths.PublicKeyPath)

	return service, nil
}

// loadOrGenerateKeys loads existing RSA keys or generates new ones if they don't exist
func (j *JWTService) loadOrGenerateKeys() error {
	// Try to load existing keys first
	if j.keysExist() {
		j.logger.Info("Loading existing RSA keys")
		return j.loadKeys()
	}

	// Generate new keys if they don't exist
	j.logger.Info("Generating new RSA keys")
	return j.generateAndSaveKeys()
}

// keysExist checks if both private and public key files exist
func (j *JWTService) keysExist() bool {
	_, privateErr := os.Stat(j.keyPaths.PrivateKeyPath)
	_, publicErr := os.Stat(j.keyPaths.PublicKeyPath)
	return privateErr == nil && publicErr == nil
}

// generateAndSaveKeys generates new RSA key pair and saves them to files
func (j *JWTService) generateAndSaveKeys() error {
	// Generate RSA private key (2048 bits as required)
	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return fmt.Errorf("failed to generate RSA private key: %w", err)
	}

	j.privateKey = privateKey
	j.publicKey = &privateKey.PublicKey

	// Save private key
	if err := j.savePrivateKey(); err != nil {
		return fmt.Errorf("failed to save private key: %w", err)
	}

	// Save public key
	if err := j.savePublicKey(); err != nil {
		return fmt.Errorf("failed to save public key: %w", err)
	}

	j.logger.Info("RSA keys generated and saved successfully")
	return nil
}

// savePrivateKey saves the private key to PEM file
func (j *JWTService) savePrivateKey() error {
	// Create directory if it doesn't exist
	if err := j.ensureKeyDirectory(j.keyPaths.PrivateKeyPath); err != nil {
		return err
	}

	// Encode private key to PKCS#1 ASN.1 DER format
	privateKeyDER := x509.MarshalPKCS1PrivateKey(j.privateKey)

	// Create PEM block
	privateKeyPEM := &pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: privateKeyDER,
	}

	// Write to file
	file, err := os.OpenFile(j.keyPaths.PrivateKeyPath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0600)
	if err != nil {
		return fmt.Errorf("failed to create private key file: %w", err)
	}
	defer file.Close()

	if err := pem.Encode(file, privateKeyPEM); err != nil {
		return fmt.Errorf("failed to encode private key: %w", err)
	}

	return nil
}

// savePublicKey saves the public key to PEM file
func (j *JWTService) savePublicKey() error {
	// Create directory if it doesn't exist
	if err := j.ensureKeyDirectory(j.keyPaths.PublicKeyPath); err != nil {
		return err
	}

	// Encode public key to PKIX ASN.1 DER format
	publicKeyDER, err := x509.MarshalPKIXPublicKey(j.publicKey)
	if err != nil {
		return fmt.Errorf("failed to marshal public key: %w", err)
	}

	// Create PEM block
	publicKeyPEM := &pem.Block{
		Type:  "PUBLIC KEY",
		Bytes: publicKeyDER,
	}

	// Write to file
	file, err := os.OpenFile(j.keyPaths.PublicKeyPath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0644)
	if err != nil {
		return fmt.Errorf("failed to create public key file: %w", err)
	}
	defer file.Close()

	if err := pem.Encode(file, publicKeyPEM); err != nil {
		return fmt.Errorf("failed to encode public key: %w", err)
	}

	return nil
}

// loadKeys loads existing RSA keys from PEM files
func (j *JWTService) loadKeys() error {
	// Load private key
	if err := j.loadPrivateKey(); err != nil {
		return fmt.Errorf("failed to load private key: %w", err)
	}

	// Load public key
	if err := j.loadPublicKey(); err != nil {
		return fmt.Errorf("failed to load public key: %w", err)
	}

	return nil
}

// loadPrivateKey loads the private key from PEM file
func (j *JWTService) loadPrivateKey() error {
	keyData, err := os.ReadFile(j.keyPaths.PrivateKeyPath)
	if err != nil {
		return fmt.Errorf("failed to read private key file: %w", err)
	}

	block, _ := pem.Decode(keyData)
	if block == nil {
		return fmt.Errorf("failed to decode private key PEM")
	}

	if block.Type != "RSA PRIVATE KEY" {
		return fmt.Errorf("invalid private key type: %s", block.Type)
	}

	privateKey, err := x509.ParsePKCS1PrivateKey(block.Bytes)
	if err != nil {
		return fmt.Errorf("failed to parse private key: %w", err)
	}

	j.privateKey = privateKey
	return nil
}

// loadPublicKey loads the public key from PEM file
func (j *JWTService) loadPublicKey() error {
	keyData, err := os.ReadFile(j.keyPaths.PublicKeyPath)
	if err != nil {
		return fmt.Errorf("failed to read public key file: %w", err)
	}

	block, _ := pem.Decode(keyData)
	if block == nil {
		return fmt.Errorf("failed to decode public key PEM")
	}

	if block.Type != "PUBLIC KEY" {
		return fmt.Errorf("invalid public key type: %s", block.Type)
	}

	publicKey, err := x509.ParsePKIXPublicKey(block.Bytes)
	if err != nil {
		return fmt.Errorf("failed to parse public key: %w", err)
	}

	rsaPublicKey, ok := publicKey.(*rsa.PublicKey)
	if !ok {
		return fmt.Errorf("key is not RSA public key")
	}

	j.publicKey = rsaPublicKey
	return nil
}

// ensureKeyDirectory creates the directory for key files if it doesn't exist
func (j *JWTService) ensureKeyDirectory(keyPath string) error {
	dir := filepath.Dir(keyPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create key directory: %w", err)
	}
	return nil
}

// TokenInfo represents generated token information
type TokenInfo struct {
	Token     string
	JTI       string
	ExpiresAt time.Time
}

// GenerateToken generates a new JWT token for the given user ID and telegram ID
func (j *JWTService) GenerateToken(userID uuid.UUID, telegramID int64) (*TokenInfo, error) {
	now := time.Now()
	expirationTime := now.Add(time.Duration(j.expiryHours) * time.Hour)

	// Create claims with all required fields
	claims := &JWTClaims{
		TelegramID: telegramID,
		JTI:        uuid.New().String(), // JWT ID for token uniqueness
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    j.issuer,                           // iss
			Subject:   userID.String(),                    // sub - using UUID as per RFC 7519 standards
			IssuedAt:  jwt.NewNumericDate(now),            // iat
			ExpiresAt: jwt.NewNumericDate(expirationTime), // exp
		},
	}

	// Create token with RS256 algorithm
	token := jwt.NewWithClaims(jwt.SigningMethodRS256, claims)

	// Sign token with private key
	tokenString, err := token.SignedString(j.privateKey)
	if err != nil {
		j.logger.Error("Failed to sign JWT token", 
			"error", err, 
			"user_id", userID.String(),
			"telegram_id", telegramID)
		return nil, fmt.Errorf("failed to sign token: %w", err)
	}

	j.logger.Info("JWT token generated successfully",
		"user_id", userID.String(),
		"telegram_id", telegramID,
		"jti", claims.JTI,
		"expires_at", expirationTime)

	return &TokenInfo{
		Token:     tokenString,
		JTI:       claims.JTI,
		ExpiresAt: expirationTime,
	}, nil
}

// ValidateToken validates a JWT token and returns the claims if valid
func (j *JWTService) ValidateToken(tokenString string) (*JWTClaims, error) {
	// Parse token with public key
	token, err := jwt.ParseWithClaims(tokenString, &JWTClaims{}, func(token *jwt.Token) (any, error) {
		// Verify signing method
		if _, ok := token.Method.(*jwt.SigningMethodRSA); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return j.publicKey, nil
	})

	if err != nil {
		j.logger.Error("Failed to parse JWT token", "error", err)
		return nil, fmt.Errorf("invalid token: %w", err)
	}

	// Extract claims
	claims, ok := token.Claims.(*JWTClaims)
	if !ok || !token.Valid {
		j.logger.Error("Invalid JWT token claims")
		return nil, fmt.Errorf("invalid token claims")
	}

	// Additional validation
	if err := j.validateClaims(claims); err != nil {
		j.logger.Error("JWT token claims validation failed", "error", err)
		return nil, err
	}

	j.logger.Debug("JWT token validated successfully",
		"telegram_id", claims.TelegramID,
		"jti", claims.JTI,
		"subject", claims.Subject)

	return claims, nil
}

// validateClaims performs additional validation on JWT claims
func (j *JWTService) validateClaims(claims *JWTClaims) error {
	// Validate issuer
	if claims.Issuer != j.issuer {
		return fmt.Errorf("invalid issuer: expected %s, got %s", j.issuer, claims.Issuer)
	}

	// Validate telegram_id
	if claims.TelegramID <= 0 {
		return fmt.Errorf("invalid telegram_id: %d", claims.TelegramID)
	}

	// Validate JTI is not empty
	if claims.JTI == "" {
		return fmt.Errorf("missing JTI claim")
	}

	// Validate subject is a valid UUID (as per RFC 7519 standards)
	if _, err := uuid.Parse(claims.Subject); err != nil {
		return fmt.Errorf("invalid subject UUID: %s", claims.Subject)
	}

	return nil
}

// GetPublicKeyPEM returns the public key in PEM format for sharing with other services
func (j *JWTService) GetPublicKeyPEM() (string, error) {
	if j.publicKey == nil {
		return "", fmt.Errorf("public key not loaded")
	}

	// Encode public key to PKIX ASN.1 DER format
	publicKeyDER, err := x509.MarshalPKIXPublicKey(j.publicKey)
	if err != nil {
		return "", fmt.Errorf("failed to marshal public key: %w", err)
	}

	// Create PEM block
	publicKeyPEM := &pem.Block{
		Type:  "PUBLIC KEY",
		Bytes: publicKeyDER,
	}

	// Encode to PEM string
	pemBytes := pem.EncodeToMemory(publicKeyPEM)
	return string(pemBytes), nil
}

// GetKeyID returns a unique identifier for the current key pair
func (j *JWTService) GetKeyID() (string, error) {
	if j.publicKey == nil {
		return "", fmt.Errorf("public key not loaded")
	}

	// Use first 8 characters of public key fingerprint as key ID
	publicKeyDER, err := x509.MarshalPKIXPublicKey(j.publicKey)
	if err != nil {
		return "", fmt.Errorf("failed to marshal public key: %w", err)
	}

	// Simple hash-based key ID
	hash := fmt.Sprintf("%x", publicKeyDER[:8])
	return hash, nil
}
