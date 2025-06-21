package services

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"log/slog"
	"net/url"
	"os"
	"sort"
	"strings"
	"testing"
	"time"
)

func TestNewTelegramValidator(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	validator := NewTelegramValidator("test_token", logger)

	if validator == nil {
		t.Fatal("Expected validator to be created")
	}

	if validator.botToken != "test_token" {
		t.Errorf("Expected bot token 'test_token', got '%s'", validator.botToken)
	}

	if validator.logger == nil {
		t.Error("Expected logger to be set")
	}
}

func TestParseInitData(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	validator := NewTelegramValidator("test_token", logger)

	tests := []struct {
		name     string
		initData string
		wantErr  bool
		checkFn  func(*TelegramData) error
	}{
		{
			name:     "valid data",
			initData: `user=%7B%22id%22%3A123456789%2C%22first_name%22%3A%22John%22%2C%22last_name%22%3A%22Doe%22%2C%22username%22%3A%22johndoe%22%2C%22language_code%22%3A%22en%22%7D&auth_date=1703243400&hash=abcd1234`,
			wantErr:  false,
			checkFn: func(data *TelegramData) error {
				if data.User.ID != 123456789 {
					return fmt.Errorf("expected user ID 123456789, got %d", data.User.ID)
				}
				if data.User.FirstName != "John" {
					return fmt.Errorf("expected first name 'John', got '%s'", data.User.FirstName)
				}
				if data.AuthDate != 1703243400 {
					return fmt.Errorf("expected auth_date 1703243400, got %d", data.AuthDate)
				}
				if data.Hash != "abcd1234" {
					return fmt.Errorf("expected hash 'abcd1234', got '%s'", data.Hash)
				}
				return nil
			},
		},
		{
			name:     "missing user field",
			initData: `auth_date=1703243400&hash=abcd1234`,
			wantErr:  true,
		},
		{
			name:     "missing auth_date field",
			initData: `user=%7B%22id%22%3A123456789%2C%22first_name%22%3A%22John%22%7D&hash=abcd1234`,
			wantErr:  true,
		},
		{
			name:     "missing hash field",
			initData: `user=%7B%22id%22%3A123456789%2C%22first_name%22%3A%22John%22%7D&auth_date=1703243400`,
			wantErr:  true,
		},
		{
			name:     "invalid user JSON",
			initData: `user=invalid_json&auth_date=1703243400&hash=abcd1234`,
			wantErr:  true,
		},
		{
			name:     "invalid auth_date format",
			initData: `user=%7B%22id%22%3A123456789%2C%22first_name%22%3A%22John%22%7D&auth_date=invalid&hash=abcd1234`,
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			data, err := validator.parseInitData(tt.initData)

			if (err != nil) != tt.wantErr {
				t.Errorf("parseInitData() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr && tt.checkFn != nil {
				if err := tt.checkFn(data); err != nil {
					t.Errorf("parseInitData() validation failed: %v", err)
				}
			}
		})
	}
}

func TestValidateRequiredFields(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	validator := NewTelegramValidator("test_token", logger)

	tests := []struct {
		name    string
		data    *TelegramData
		wantErr bool
	}{
		{
			name: "valid data",
			data: &TelegramData{
				User:     &TelegramUser{ID: 123, FirstName: "John"},
				AuthDate: 1703243400,
				Hash:     "abcd1234",
			},
			wantErr: false,
		},
		{
			name: "missing user",
			data: &TelegramData{
				AuthDate: 1703243400,
				Hash:     "abcd1234",
			},
			wantErr: true,
		},
		{
			name: "missing auth_date",
			data: &TelegramData{
				User: &TelegramUser{ID: 123, FirstName: "John"},
				Hash: "abcd1234",
			},
			wantErr: true,
		},
		{
			name: "missing hash",
			data: &TelegramData{
				User:     &TelegramUser{ID: 123, FirstName: "John"},
				AuthDate: 1703243400,
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validator.validateRequiredFields(tt.data)
			if (err != nil) != tt.wantErr {
				t.Errorf("validateRequiredFields() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestValidateAuthDate(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	validator := NewTelegramValidator("test_token", logger)

	now := time.Now()
	validTime := now.Add(-1 * time.Hour).Unix()
	tooOldTime := now.Add(-25 * time.Hour).Unix()
	futureTime := now.Add(1 * time.Hour).Unix()

	tests := []struct {
		name     string
		authDate int64
		wantErr  bool
	}{
		{
			name:     "valid recent time",
			authDate: validTime,
			wantErr:  false,
		},
		{
			name:     "too old time",
			authDate: tooOldTime,
			wantErr:  true,
		},
		{
			name:     "future time",
			authDate: futureTime,
			wantErr:  true,
		},
		{
			name:     "current time",
			authDate: now.Unix(),
			wantErr:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validator.validateAuthDate(tt.authDate)
			if (err != nil) != tt.wantErr {
				t.Errorf("validateAuthDate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestValidateUserData(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	validator := NewTelegramValidator("test_token", logger)

	tests := []struct {
		name    string
		user    *TelegramUser
		wantErr bool
	}{
		{
			name: "valid user",
			user: &TelegramUser{
				ID:           123456789,
				IsBot:        false,
				FirstName:    "John",
				LastName:     "Doe",
				Username:     "johndoe",
				LanguageCode: "en",
			},
			wantErr: false,
		},
		{
			name: "minimal valid user",
			user: &TelegramUser{
				ID:        123456789,
				IsBot:     false,
				FirstName: "John",
			},
			wantErr: false,
		},
		{
			name: "negative user ID",
			user: &TelegramUser{
				ID:        -123,
				IsBot:     false,
				FirstName: "John",
			},
			wantErr: true,
		},
		{
			name: "zero user ID",
			user: &TelegramUser{
				ID:        0,
				IsBot:     false,
				FirstName: "John",
			},
			wantErr: true,
		},
		{
			name: "bot user",
			user: &TelegramUser{
				ID:        123456789,
				IsBot:     true,
				FirstName: "BotName",
			},
			wantErr: true,
		},
		{
			name: "empty first name",
			user: &TelegramUser{
				ID:        123456789,
				IsBot:     false,
				FirstName: "",
			},
			wantErr: true,
		},
		{
			name: "whitespace only first name",
			user: &TelegramUser{
				ID:        123456789,
				IsBot:     false,
				FirstName: "   ",
			},
			wantErr: true,
		},
		{
			name: "too long first name",
			user: &TelegramUser{
				ID:        123456789,
				IsBot:     false,
				FirstName: strings.Repeat("a", 101),
			},
			wantErr: true,
		},
		{
			name: "too long last name",
			user: &TelegramUser{
				ID:        123456789,
				IsBot:     false,
				FirstName: "John",
				LastName:  strings.Repeat("b", 101),
			},
			wantErr: true,
		},
		{
			name: "too long username",
			user: &TelegramUser{
				ID:        123456789,
				IsBot:     false,
				FirstName: "John",
				Username:  strings.Repeat("c", 101),
			},
			wantErr: true,
		},
		{
			name: "too long language code",
			user: &TelegramUser{
				ID:           123456789,
				IsBot:        false,
				FirstName:    "John",
				LanguageCode: strings.Repeat("d", 11),
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validator.validateUserData(tt.user)
			if (err != nil) != tt.wantErr {
				t.Errorf("validateUserData() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestGenerateSecretKey(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	validator := NewTelegramValidator("test_token", logger)

	secretKey := validator.generateSecretKey()

	if len(secretKey) != 32 { // SHA256 hash is 32 bytes
		t.Errorf("Expected secret key length 32, got %d", len(secretKey))
	}

	// Test that the same token produces the same key
	secretKey2 := validator.generateSecretKey()
	if !hmac.Equal(secretKey, secretKey2) {
		t.Error("Expected same secret key for same token")
	}

	// Test that different tokens produce different keys
	validator2 := NewTelegramValidator("different_token", logger)
	secretKey3 := validator2.generateSecretKey()
	if hmac.Equal(secretKey, secretKey3) {
		t.Error("Expected different secret keys for different tokens")
	}
}

func TestCalculateHMAC(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	validator := NewTelegramValidator("test_token", logger)

	data := "test_data"
	key := []byte("test_key")

	hash := validator.calculateHMAC(data, key)

	// Calculate expected hash manually
	h := hmac.New(sha256.New, key)
	h.Write([]byte(data))
	expected := hex.EncodeToString(h.Sum(nil))

	if hash != expected {
		t.Errorf("Expected hash %s, got %s", expected, hash)
	}

	// Test that same data produces same hash
	hash2 := validator.calculateHMAC(data, key)
	if hash != hash2 {
		t.Error("Expected same hash for same data")
	}

	// Test that different data produces different hash
	hash3 := validator.calculateHMAC("different_data", key)
	if hash == hash3 {
		t.Error("Expected different hash for different data")
	}
}

func TestValidateSignature(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	botToken := "123456789:ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefgh"
	validator := NewTelegramValidator(botToken, logger)

	// Create test data with valid signature
	userData := `{"id":123456789,"first_name":"John","last_name":"Doe","username":"johndoe","language_code":"en"}`
	authDate := "1703243400"
	queryID := "test_query_id"

	// Build initData without hash
	values := url.Values{}
	values.Set("user", userData)
	values.Set("auth_date", authDate)
	values.Set("query_id", queryID)

	// Calculate valid hash
	var pairs []string
	for key, valueSlice := range values {
		if len(valueSlice) > 0 {
			pairs = append(pairs, key+"="+valueSlice[0])
		}
	}
	sort.Strings(pairs)
	dataCheckString := strings.Join(pairs, "\n")

	secretKey := validator.generateSecretKey()
	validHash := validator.calculateHMAC(dataCheckString, secretKey)

	// Build complete initData with hash
	values.Set("hash", validHash)
	initDataWithValidHash := values.Encode()

	tests := []struct {
		name     string
		initData string
		hash     string
		wantErr  bool
	}{
		{
			name:     "valid signature",
			initData: initDataWithValidHash,
			hash:     validHash,
			wantErr:  false,
		},
		{
			name:     "invalid signature",
			initData: initDataWithValidHash,
			hash:     "invalid_hash",
			wantErr:  true,
		},
		{
			name:     "malformed initData",
			initData: "invalid%data",
			hash:     "some_hash",
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validator.validateSignature(tt.initData, tt.hash)
			if (err != nil) != tt.wantErr {
				t.Errorf("validateSignature() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestValidateTelegramDataIntegration(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	botToken := "123456789:ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefgh"
	validator := NewTelegramValidator(botToken, logger)

	// Test with current timestamp
	currentTime := time.Now().Unix()

	// Create test initData with valid signature
	userData := `{"id":123456789,"first_name":"John","last_name":"Doe","username":"johndoe","language_code":"en","is_premium":true}`

	values := url.Values{}
	values.Set("user", userData)
	values.Set("auth_date", fmt.Sprintf("%d", currentTime))
	values.Set("query_id", "test_query")

	// Calculate hash for this data
	var pairs []string
	for key, valueSlice := range values {
		if len(valueSlice) > 0 {
			pairs = append(pairs, key+"="+valueSlice[0])
		}
	}
	sort.Strings(pairs)
	dataCheckString := strings.Join(pairs, "\n")

	secretKey := validator.generateSecretKey()
	validHash := validator.calculateHMAC(dataCheckString, secretKey)
	values.Set("hash", validHash)

	validInitData := values.Encode()

	// Create test data with old timestamp (>24h)
	oldTime := time.Now().Add(-25 * time.Hour).Unix()
	valuesOld := url.Values{}
	valuesOld.Set("user", userData)
	valuesOld.Set("auth_date", fmt.Sprintf("%d", oldTime))

	var pairsOld []string
	for key, valueSlice := range valuesOld {
		if len(valueSlice) > 0 {
			pairsOld = append(pairsOld, key+"="+valueSlice[0])
		}
	}
	sort.Strings(pairsOld)
	dataCheckStringOld := strings.Join(pairsOld, "\n")
	validHashOld := validator.calculateHMAC(dataCheckStringOld, secretKey)
	valuesOld.Set("hash", validHashOld)
	oldInitData := valuesOld.Encode()

	// Create test data with invalid user (bot)
	botUserData := `{"id":123456789,"is_bot":true,"first_name":"BotName"}`
	valuesBot := url.Values{}
	valuesBot.Set("user", botUserData)
	valuesBot.Set("auth_date", fmt.Sprintf("%d", currentTime))

	var pairsBot []string
	for key, valueSlice := range valuesBot {
		if len(valueSlice) > 0 {
			pairsBot = append(pairsBot, key+"="+valueSlice[0])
		}
	}
	sort.Strings(pairsBot)
	dataCheckStringBot := strings.Join(pairsBot, "\n")
	validHashBot := validator.calculateHMAC(dataCheckStringBot, secretKey)
	valuesBot.Set("hash", validHashBot)
	botInitData := valuesBot.Encode()

	tests := []struct {
		name     string
		initData string
		wantErr  bool
		checkFn  func(*TelegramData) error
	}{
		{
			name:     "valid complete data",
			initData: validInitData,
			wantErr:  false,
			checkFn: func(data *TelegramData) error {
				if data.User.ID != 123456789 {
					return fmt.Errorf("expected user ID 123456789, got %d", data.User.ID)
				}
				if data.User.FirstName != "John" {
					return fmt.Errorf("expected first name 'John', got '%s'", data.User.FirstName)
				}
				if data.User.LastName != "Doe" {
					return fmt.Errorf("expected last name 'Doe', got '%s'", data.User.LastName)
				}
				if data.User.Username != "johndoe" {
					return fmt.Errorf("expected username 'johndoe', got '%s'", data.User.Username)
				}
				if !data.User.IsPremium {
					return fmt.Errorf("expected user to be premium")
				}
				return nil
			},
		},
		{
			name:     "old auth date",
			initData: oldInitData,
			wantErr:  true,
		},
		{
			name:     "bot user",
			initData: botInitData,
			wantErr:  true,
		},
		{
			name:     "completely invalid data",
			initData: "invalid_data",
			wantErr:  true,
		},
		{
			name:     "empty data",
			initData: "",
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			data, err := validator.ValidateTelegramData(tt.initData)

			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateTelegramData() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr && tt.checkFn != nil {
				if err := tt.checkFn(data); err != nil {
					t.Errorf("ValidateTelegramData() validation failed: %v", err)
				}
			}
		})
	}
}
