package jwt

import (
	"crypto/rsa"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// KeyProvider manages RSA public keys for JWT verification
type KeyProvider struct {
	publicKey   *rsa.PublicKey
	keyURL      string
	lastFetched time.Time
	cacheTTL    time.Duration
	httpClient  *http.Client
}

// NewKeyProvider creates a new KeyProvider instance
func NewKeyProvider(keyURL string) *KeyProvider {
	return &KeyProvider{
		keyURL:   keyURL,
		cacheTTL: 5 * time.Minute, // Cache key for 5 minutes
		httpClient: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

// GetPublicKey returns the RSA public key, fetching it if necessary
func (kp *KeyProvider) GetPublicKey() (*rsa.PublicKey, error) {
	// Check if key is cached and still valid
	if kp.publicKey != nil && time.Since(kp.lastFetched) < kp.cacheTTL {
		return kp.publicKey, nil
	}

	// Fetch new key
	key, err := kp.fetchPublicKey()
	if err != nil {
		// If fetch fails but we have a cached key, use it
		if kp.publicKey != nil {
			return kp.publicKey, nil
		}
		return nil, err
	}

	kp.publicKey = key
	kp.lastFetched = time.Now()
	return key, nil
}

// fetchPublicKey fetches the public key from the remote URL
func (kp *KeyProvider) fetchPublicKey() (*rsa.PublicKey, error) {
	resp, err := kp.httpClient.Get(kp.keyURL)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch public key: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to fetch public key: HTTP %d", resp.StatusCode)
	}

	keyData, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read public key response: %w", err)
	}

	publicKey, err := jwt.ParseRSAPublicKeyFromPEM(keyData)
	if err != nil {
		return nil, fmt.Errorf("failed to parse RSA public key: %w", err)
	}

	return publicKey, nil
}

// ValidateToken validates a JWT token using the public key
func (kp *KeyProvider) ValidateToken(tokenString string) (*jwt.Token, error) {
	publicKey, err := kp.GetPublicKey()
	if err != nil {
		return nil, fmt.Errorf("failed to get public key: %w", err)
	}

	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		// Verify signing method
		if _, ok := token.Method.(*jwt.SigningMethodRSA); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return publicKey, nil
	})

	if err != nil {
		return nil, fmt.Errorf("failed to parse token: %w", err)
	}

	if !token.Valid {
		return nil, fmt.Errorf("invalid token")
	}

	return token, nil
}

// ExtractUserID extracts user ID from JWT claims
func ExtractUserID(token *jwt.Token) (string, error) {
	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return "", fmt.Errorf("invalid token claims")
	}

	userID, ok := claims["sub"].(string)
	if !ok {
		return "", fmt.Errorf("user ID not found in token")
	}

	return userID, nil
}
