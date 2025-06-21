package middleware

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/shard-legends/auth-service/internal/services"
)

// JWTPublicKeyMiddleware provides middleware for exposing the JWT public key
type JWTPublicKeyMiddleware struct {
	jwtService *services.JWTService
}

// NewJWTPublicKeyMiddleware creates a new JWT public key middleware
func NewJWTPublicKeyMiddleware(jwtService *services.JWTService) *JWTPublicKeyMiddleware {
	return &JWTPublicKeyMiddleware{
		jwtService: jwtService,
	}
}

// PublicKeyHandler returns the JWT public key in PEM format
func (m *JWTPublicKeyMiddleware) PublicKeyHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		publicKeyPEM, err := m.jwtService.GetPublicKeyPEM()
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "Failed to retrieve public key",
			})
			return
		}

		keyID, err := m.jwtService.GetKeyID()
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "Failed to retrieve key ID",
			})
			return
		}

		// Return public key information in JSON Web Key Set (JWKS) compatible format
		c.JSON(http.StatusOK, gin.H{
			"keys": []gin.H{
				{
					"kty": "RSA",        // Key Type
					"use": "sig",        // Public Key Use
					"alg": "RS256",      // Algorithm
					"kid": keyID,        // Key ID
					"pem": publicKeyPEM, // PEM format for easy consumption
				},
			},
		})
	}
}

// PublicKeyPEMHandler returns only the PEM format public key (simpler endpoint)
func (m *JWTPublicKeyMiddleware) PublicKeyPEMHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		publicKeyPEM, err := m.jwtService.GetPublicKeyPEM()
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "Failed to retrieve public key",
			})
			return
		}

		// Set content type to text/plain for PEM format
		c.Header("Content-Type", "text/plain")
		c.String(http.StatusOK, publicKeyPEM)
	}
}
