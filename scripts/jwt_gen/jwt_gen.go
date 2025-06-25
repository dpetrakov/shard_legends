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
	expiryHours       = 24

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

func main() {
	// Читаем user_id
	userIDBytes, err := ioutil.ReadFile(userIDFile)
	if err != nil {
		log.Fatalf("failed to read user_id file: %v", err)
	}
	userIDStr := strings.TrimSpace(string(userIDBytes))
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		log.Fatalf("invalid user_id in file: %v", err)
	}

	// Читаем приватный ключ
	keyBytes, err := ioutil.ReadFile(privateKeyFile)
	if err != nil {
		log.Fatalf("failed to read private key file: %v", err)
	}

	block, _ := pem.Decode(keyBytes)
	if block == nil || block.Type != "RSA PRIVATE KEY" {
		log.Fatalf("failed to decode PEM private key")
	}

	privKey, err := x509.ParsePKCS1PrivateKey(block.Bytes)
	if err != nil {
		log.Fatalf("failed to parse RSA private key: %v", err)
	}

	now := time.Now()
	expiresAt := now.Add(expiryHours * time.Hour)

	claims := &JWTClaims{
		TelegramID: telegramID,
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
		log.Fatalf("failed to sign token: %v", err)
	}

	// Записываем токен в файл
	if err := ioutil.WriteFile(tokenOutFile, []byte(tokenString), 0644); err != nil {
		log.Fatalf("failed to write token to file: %v", err)
	}

	absPath, _ := filepath.Abs(tokenOutFile)
	fmt.Printf("JWT токен успешно сгенерирован и сохранён в %s\n", absPath)
}
