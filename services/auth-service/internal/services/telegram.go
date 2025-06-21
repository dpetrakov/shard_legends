package services

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/url"
	"sort"
	"strconv"
	"strings"
	"time"
)

// TelegramUser represents a Telegram user from initData
type TelegramUser struct {
	ID           int64  `json:"id"`
	IsBot        bool   `json:"is_bot"`
	FirstName    string `json:"first_name"`
	LastName     string `json:"last_name,omitempty"`
	Username     string `json:"username,omitempty"`
	LanguageCode string `json:"language_code,omitempty"`
	IsPremium    bool   `json:"is_premium,omitempty"`
	PhotoURL     string `json:"photo_url,omitempty"`
}

// TelegramData represents parsed Telegram Web App initData
type TelegramData struct {
	User       *TelegramUser `json:"user"`
	AuthDate   int64         `json:"auth_date"`
	Hash       string        `json:"hash"`
	QueryID    string        `json:"query_id,omitempty"`
	StartParam string        `json:"start_param,omitempty"`
	Signature  string        `json:"signature,omitempty"`
}

// TelegramValidator handles validation of Telegram Web App data
type TelegramValidator struct {
	botToken string
	logger   *slog.Logger
}

// NewTelegramValidator creates a new Telegram data validator
func NewTelegramValidator(botToken string, logger *slog.Logger) *TelegramValidator {
	return &TelegramValidator{
		botToken: botToken,
		logger:   logger,
	}
}

// ValidateTelegramData validates Telegram Web App initData according to official algorithm
func (tv *TelegramValidator) ValidateTelegramData(initData string) (*TelegramData, error) {
	tv.logger.Debug("Starting Telegram data validation", "data_length", len(initData))

	// Parse URL-encoded data
	parsedData, err := tv.parseInitData(initData)
	if err != nil {
		tv.logger.Error("Failed to parse initData", "error", err)
		return nil, fmt.Errorf("invalid initData format: %w", err)
	}

	// Validate required fields
	if err := tv.validateRequiredFields(parsedData); err != nil {
		tv.logger.Error("Required fields validation failed", "error", err)
		return nil, err
	}

	// Validate auth_date (not older than 24 hours)
	if err := tv.validateAuthDate(parsedData.AuthDate); err != nil {
		tv.logger.Error("Auth date validation failed", "auth_date", parsedData.AuthDate, "error", err)
		return nil, err
	}

	// Validate HMAC signature
	if err := tv.validateSignature(initData, parsedData.Hash); err != nil {
		tv.logger.Error("HMAC signature validation failed", "error", err)
		return nil, err
	}

	// Validate user data structure
	if err := tv.validateUserData(parsedData.User); err != nil {
		tv.logger.Error("User data validation failed", "error", err)
		return nil, err
	}

	tv.logger.Info("Telegram data validation successful",
		"user_id", parsedData.User.ID,
		"username", parsedData.User.Username,
		"first_name", parsedData.User.FirstName)

	return parsedData, nil
}

// parseInitData parses URL-encoded initData string
func (tv *TelegramValidator) parseInitData(initData string) (*TelegramData, error) {
	values, err := url.ParseQuery(initData)
	if err != nil {
		return nil, fmt.Errorf("failed to parse query string: %w", err)
	}

	data := &TelegramData{}

	// Parse user field (required JSON object)
	userStr := values.Get("user")
	if userStr == "" {
		return nil, fmt.Errorf("missing required field: user")
	}

	user := &TelegramUser{}
	if err := json.Unmarshal([]byte(userStr), user); err != nil {
		return nil, fmt.Errorf("invalid user JSON: %w", err)
	}
	data.User = user

	// Parse auth_date (required)
	authDateStr := values.Get("auth_date")
	if authDateStr == "" {
		return nil, fmt.Errorf("missing required field: auth_date")
	}

	authDate, err := strconv.ParseInt(authDateStr, 10, 64)
	if err != nil {
		return nil, fmt.Errorf("invalid auth_date format: %w", err)
	}
	data.AuthDate = authDate

	// Parse hash (required)
	hash := values.Get("hash")
	if hash == "" {
		return nil, fmt.Errorf("missing required field: hash")
	}
	data.Hash = hash

	// Parse optional fields
	data.QueryID = values.Get("query_id")
	data.StartParam = values.Get("start_param")
	data.Signature = values.Get("signature")

	return data, nil
}

// validateRequiredFields checks that all required fields are present and valid
func (tv *TelegramValidator) validateRequiredFields(data *TelegramData) error {
	if data.User == nil {
		return fmt.Errorf("user data is required")
	}

	if data.AuthDate == 0 {
		return fmt.Errorf("auth_date is required")
	}

	if data.Hash == "" {
		return fmt.Errorf("hash is required")
	}

	return nil
}

// validateAuthDate checks that auth_date is not older than 24 hours
func (tv *TelegramValidator) validateAuthDate(authDate int64) error {
	authTime := time.Unix(authDate, 0)
	now := time.Now()

	// Check if auth_date is in the future (invalid)
	if authTime.After(now) {
		return fmt.Errorf("auth_date cannot be in the future")
	}

	// Check if auth_date is older than 24 hours
	maxAge := 24 * time.Hour
	if now.Sub(authTime) > maxAge {
		return fmt.Errorf("auth_date is too old, must be within 24 hours")
	}

	return nil
}

// validateSignature validates HMAC-SHA256 signature according to Telegram algorithm
func (tv *TelegramValidator) validateSignature(initData, receivedHash string) error {
	// Step 1: Parse initData and remove hash parameter
	values, err := url.ParseQuery(initData)
	if err != nil {
		return fmt.Errorf("failed to parse initData for signature validation: %w", err)
	}

	// Remove hash from values
	values.Del("hash")

	// Step 2: Create data-check-string
	var pairs []string
	for key, valueSlice := range values {
		if len(valueSlice) > 0 {
			pairs = append(pairs, key+"="+valueSlice[0])
		}
	}

	// Sort alphabetically
	sort.Strings(pairs)

	// Join with newlines
	dataCheckString := strings.Join(pairs, "\n")

	// Step 3: Generate secret key
	secretKey := tv.generateSecretKey()

	// Step 4: Calculate HMAC-SHA256
	calculatedHash := tv.calculateHMAC(dataCheckString, secretKey)

	// Step 5: Compare hashes
	if calculatedHash != receivedHash {
		tv.logger.Error("HMAC signature mismatch",
			"calculated", calculatedHash,
			"received", receivedHash,
			"data_check_string", dataCheckString)
		return fmt.Errorf("invalid HMAC signature")
	}

	return nil
}

// generateSecretKey generates secret key from bot token
func (tv *TelegramValidator) generateSecretKey() []byte {
	h := hmac.New(sha256.New, []byte("WebAppData"))
	h.Write([]byte(tv.botToken))
	return h.Sum(nil)
}

// calculateHMAC calculates HMAC-SHA256 for given data and key
func (tv *TelegramValidator) calculateHMAC(data string, key []byte) string {
	h := hmac.New(sha256.New, key)
	h.Write([]byte(data))
	return hex.EncodeToString(h.Sum(nil))
}

// validateUserData validates user object structure and required fields
func (tv *TelegramValidator) validateUserData(user *TelegramUser) error {
	if user.ID <= 0 {
		return fmt.Errorf("user.id must be positive")
	}

	if user.IsBot {
		return fmt.Errorf("bots are not allowed")
	}

	if strings.TrimSpace(user.FirstName) == "" {
		return fmt.Errorf("user.first_name is required")
	}

	// Validate field lengths to prevent extremely long data
	if len(user.FirstName) > 100 {
		return fmt.Errorf("user.first_name too long (max 100 characters)")
	}

	if len(user.LastName) > 100 {
		return fmt.Errorf("user.last_name too long (max 100 characters)")
	}

	if len(user.Username) > 100 {
		return fmt.Errorf("user.username too long (max 100 characters)")
	}

	if len(user.LanguageCode) > 10 {
		return fmt.Errorf("user.language_code too long (max 10 characters)")
	}

	return nil
}
