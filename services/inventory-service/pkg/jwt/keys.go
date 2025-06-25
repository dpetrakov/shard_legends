package jwt

import (
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"
)

// LoadPublicKeyFromFile loads an RSA public key from a PEM file
func LoadPublicKeyFromFile(keyPath string) (*rsa.PublicKey, error) {
	// Read the key file
	keyData, err := os.ReadFile(keyPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read public key file: %w", err)
	}

	// Parse the key data using shared function
	return parsePublicKeyPEM(keyData)
}

// LoadPublicKeyFromAuthService loads an RSA public key from Auth Service endpoint
func LoadPublicKeyFromAuthService(authServiceURL string) (*rsa.PublicKey, error) {
	if authServiceURL == "" {
		authServiceURL = "http://auth-service:8080"
	}
	
	endpoint := authServiceURL + "/public-key.pem"
	
	// Create HTTP client with timeout
	client := &http.Client{
		Timeout: 10 * time.Second,
	}
	
	// Make request to auth service
	resp, err := client.Get(endpoint)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch public key from auth service: %w", err)
	}
	defer resp.Body.Close()
	
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("auth service returned status %d when fetching public key", resp.StatusCode)
	}
	
	// Read response body
	keyData, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read public key response: %w", err)
	}
	
	// Parse the key data using existing function logic
	return parsePublicKeyPEM(keyData)
}

// parsePublicKeyPEM parses PEM-encoded public key data
func parsePublicKeyPEM(keyData []byte) (*rsa.PublicKey, error) {
	// Decode PEM block
	block, _ := pem.Decode(keyData)
	if block == nil {
		return nil, fmt.Errorf("failed to decode public key PEM")
	}

	// Parse the public key
	var publicKey *rsa.PublicKey
	switch block.Type {
	case "PUBLIC KEY":
		key, err := x509.ParsePKIXPublicKey(block.Bytes)
		if err != nil {
			return nil, fmt.Errorf("failed to parse PKIX public key: %w", err)
		}
		var ok bool
		publicKey, ok = key.(*rsa.PublicKey)
		if !ok {
			return nil, fmt.Errorf("key is not an RSA public key")
		}
	case "RSA PUBLIC KEY":
		key, err := x509.ParsePKCS1PublicKey(block.Bytes)
		if err != nil {
			return nil, fmt.Errorf("failed to parse PKCS1 public key: %w", err)
		}
		publicKey = key
	default:
		return nil, fmt.Errorf("unsupported key type: %s", block.Type)
	}

	return publicKey, nil
}