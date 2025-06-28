package services

import (
	"crypto/hmac"
	"log/slog"
	"net/url"
	"os"
	"sort"
	"strings"
	"testing"
)

func TestValidateSignatureWithMultipleTokens(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))

	// Test tokens
	token1 := "123456789:ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefgh"
	token2 := "987654321:XYZabcdefghIJKLMNOPQRSTUVWXYZABCD"
	tokens := []string{token1, token2}

	validator := NewTelegramValidator(tokens, logger)

	// Create test data
	userData := `{"id":123456789,"first_name":"John","last_name":"Doe","username":"johndoe","language_code":"en"}`
	authDate := "1703243400"
	queryID := "test_query_id"

	// Build initData without hash
	values := url.Values{}
	values.Set("user", userData)
	values.Set("auth_date", authDate)
	values.Set("query_id", queryID)

	// Calculate valid hash for token1
	var pairs []string
	for key, valueSlice := range values {
		if len(valueSlice) > 0 {
			pairs = append(pairs, key+"="+valueSlice[0])
		}
	}
	sort.Strings(pairs)
	dataCheckString := strings.Join(pairs, "\n")

	secretKey1 := validator.generateSecretKeyForToken(token1)
	validHash1 := validator.calculateHMAC(dataCheckString, secretKey1)

	// Calculate valid hash for token2
	secretKey2 := validator.generateSecretKeyForToken(token2)
	validHash2 := validator.calculateHMAC(dataCheckString, secretKey2)

	// Build complete initData
	values.Set("hash", validHash1)
	initDataWithToken1Hash := values.Encode()

	values.Set("hash", validHash2)
	initDataWithToken2Hash := values.Encode()

	tests := []struct {
		name     string
		initData string
		hash     string
		wantErr  bool
	}{
		{
			name:     "valid signature with first token",
			initData: initDataWithToken1Hash,
			hash:     validHash1,
			wantErr:  false,
		},
		{
			name:     "valid signature with second token",
			initData: initDataWithToken2Hash,
			hash:     validHash2,
			wantErr:  false,
		},
		{
			name:     "invalid signature",
			initData: initDataWithToken1Hash,
			hash:     "invalid_hash",
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validator.validateSignatureWithMultipleTokens(tt.initData, tt.hash)
			if (err != nil) != tt.wantErr {
				t.Errorf("validateSignatureWithMultipleTokens() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestGenerateSecretKeyForToken(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	validator := NewTelegramValidator([]string{"test_token"}, logger)

	token1 := "token1"
	token2 := "token2"

	secretKey1 := validator.generateSecretKeyForToken(token1)
	secretKey2 := validator.generateSecretKeyForToken(token2)

	if len(secretKey1) != 32 { // SHA256 hash is 32 bytes
		t.Errorf("Expected secret key length 32, got %d", len(secretKey1))
	}

	if len(secretKey2) != 32 { // SHA256 hash is 32 bytes
		t.Errorf("Expected secret key length 32, got %d", len(secretKey2))
	}

	// Different tokens should produce different keys
	if hmac.Equal(secretKey1, secretKey2) {
		t.Error("Expected different secret keys for different tokens")
	}

	// Same token should produce same key
	secretKey1Again := validator.generateSecretKeyForToken(token1)
	if !hmac.Equal(secretKey1, secretKey1Again) {
		t.Error("Expected same secret key for same token")
	}
}

func TestMultipleTokensIntegration(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))

	// Test with multiple tokens
	token1 := "123456789:ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefgh"
	token2 := "987654321:XYZabcdefghIJKLMNOPQRSTUVWXYZABCD"
	tokens := []string{token1, token2}

	validator := NewTelegramValidator(tokens, logger)

	// Test that validator accepts data signed with either token
	userData := `{"id":123456789,"first_name":"John","last_name":"Doe","username":"johndoe","language_code":"en"}`

	// Test with token1
	values1 := url.Values{}
	values1.Set("user", userData)
	values1.Set("auth_date", "1703243400")

	var pairs1 []string
	for key, valueSlice := range values1 {
		if len(valueSlice) > 0 {
			pairs1 = append(pairs1, key+"="+valueSlice[0])
		}
	}
	sort.Strings(pairs1)
	dataCheckString1 := strings.Join(pairs1, "\n")

	secretKey1 := validator.generateSecretKeyForToken(token1)
	validHash1 := validator.calculateHMAC(dataCheckString1, secretKey1)
	values1.Set("hash", validHash1)
	initData1 := values1.Encode()

	// Test with token2
	values2 := url.Values{}
	values2.Set("user", userData)
	values2.Set("auth_date", "1703243400")

	var pairs2 []string
	for key, valueSlice := range values2 {
		if len(valueSlice) > 0 {
			pairs2 = append(pairs2, key+"="+valueSlice[0])
		}
	}
	sort.Strings(pairs2)
	dataCheckString2 := strings.Join(pairs2, "\n")

	secretKey2 := validator.generateSecretKeyForToken(token2)
	validHash2 := validator.calculateHMAC(dataCheckString2, secretKey2)
	values2.Set("hash", validHash2)
	initData2 := values2.Encode()

	// Both should validate successfully
	err1 := validator.validateSignatureWithMultipleTokens(initData1, validHash1)
	if err1 != nil {
		t.Errorf("Expected token1 signature to validate, got error: %v", err1)
	}

	err2 := validator.validateSignatureWithMultipleTokens(initData2, validHash2)
	if err2 != nil {
		t.Errorf("Expected token2 signature to validate, got error: %v", err2)
	}
}
