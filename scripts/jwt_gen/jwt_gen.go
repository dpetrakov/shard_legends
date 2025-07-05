package main

import (
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"io/ioutil"
	"log"
	"path/filepath"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

// Константы, полученные из базы данных dev-стенда
const (
	issuer            = "shard-legends-auth"
	telegramID  int64 = 56851083
	expiryHours       = 720 // 30 дней = 30 * 24 часа

	userIDFile     = "user_id"
	privateKeyFile = "private_key"
	tokenOutFile   = "token.jwt"
)

// JWTClaims описывает кастомные поля токена
type JWTClaims struct {
	TelegramID int64  `json:"telegram_id"`
	JTI        string `json:"jti"`
	jwt.RegisteredClaims
}

// generateJWTToken генерирует JWT токен с указанными файлами и telegram_id
func generateJWTToken(userIDFilePath, privateKeyFilePath, tokenOutFilePath string, telegramIDForToken int64) error {
	// Читаем user_id
	userIDBytes, err := ioutil.ReadFile(userIDFilePath)
	if err != nil {
		return fmt.Errorf("failed to read user_id file %s: %v", userIDFilePath, err)
	}
	userIDStr := strings.TrimSpace(string(userIDBytes))
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		return fmt.Errorf("invalid user_id in file %s: %v", userIDFilePath, err)
	}

	// Читаем приватный ключ
	keyBytes, err := ioutil.ReadFile(privateKeyFilePath)
	if err != nil {
		return fmt.Errorf("failed to read private key file %s: %v", privateKeyFilePath, err)
	}

	block, _ := pem.Decode(keyBytes)
	if block == nil || block.Type != "RSA PRIVATE KEY" {
		return fmt.Errorf("failed to decode PEM private key from %s", privateKeyFilePath)
	}

	privKey, err := x509.ParsePKCS1PrivateKey(block.Bytes)
	if err != nil {
		return fmt.Errorf("failed to parse RSA private key from %s: %v", privateKeyFilePath, err)
	}

	now := time.Now()
	expiresAt := now.Add(expiryHours * time.Hour)

	claims := &JWTClaims{
		TelegramID: telegramIDForToken,
		JTI:        uuid.New().String(),
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    issuer,
			Subject:   userID.String(),
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(expiresAt),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodRS256, claims)
	tokenString, err := token.SignedString(privKey)
	if err != nil {
		return fmt.Errorf("failed to sign token: %v", err)
	}

	// Записываем токен в файл
	if err := ioutil.WriteFile(tokenOutFilePath, []byte(tokenString), 0644); err != nil {
		return fmt.Errorf("failed to write token to file %s: %v", tokenOutFilePath, err)
	}

	absPath, _ := filepath.Abs(tokenOutFilePath)
	fmt.Printf("JWT токен успешно сгенерирован и сохранён в %s\n", absPath)
	return nil
}

func main() {
	// Генерируем основной токен (существующая логика)
	fmt.Println("Генерация основного токена...")
	if err := generateJWTToken(userIDFile, privateKeyFile, tokenOutFile, telegramID); err != nil {
		log.Fatalf("failed to generate main token: %v", err)
	}

	// Генерируем дополнительный токен для ssh0_dev
	fmt.Println("\nГенерация токена для ssh0_dev...")
	ssh0UserIDFile := "user_id_ssh0_dev"
	ssh0PrivateKeyFile := "private_key_ssh0_dev"
	ssh0TokenOutFile := "token_ssh0_dev.jwt"
	ssh0TelegramID := int64(218635402) // Другой telegram_id для ssh0_dev

	if err := generateJWTToken(ssh0UserIDFile, ssh0PrivateKeyFile, ssh0TokenOutFile, ssh0TelegramID); err != nil {
		log.Fatalf("failed to generate ssh0_dev token: %v", err)
	}

	fmt.Println("\n✅ Все токены успешно сгенерированы!")
}
