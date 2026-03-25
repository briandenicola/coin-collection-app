package middleware

import (
	"crypto/sha256"
	"encoding/hex"
	"net/http"
	"strings"
	"time"

	"github.com/briandenicola/ancient-coins-api/models"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"gorm.io/gorm"
)

func AuthRequired(jwtSecret string, db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Try API key auth first
		apiKey := c.GetHeader("X-API-Key")
		if apiKey != "" {
			if authenticateApiKey(c, apiKey, db) {
				c.Next()
				return
			}
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Invalid API key"})
			return
		}

		// Fall back to JWT bearer auth
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Authorization header required"})
			return
		}

		tokenString := strings.TrimPrefix(authHeader, "Bearer ")
		if tokenString == authHeader {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Bearer token required"})
			return
		}

		token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, jwt.ErrSignatureInvalid
			}
			return []byte(jwtSecret), nil
		})

		if err != nil || !token.Valid {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Invalid token"})
			return
		}

		claims, ok := token.Claims.(jwt.MapClaims)
		if !ok {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Invalid token claims"})
			return
		}

		userID, ok := claims["userId"].(float64)
		if !ok {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Invalid user ID in token"})
			return
		}

		role, _ := claims["role"].(string)

		c.Set("userId", uint(userID))
		c.Set("userRole", role)
		c.Next()
	}
}

func authenticateApiKey(c *gin.Context, plainKey string, db *gorm.DB) bool {
	hash := sha256.Sum256([]byte(plainKey))
	keyHash := hex.EncodeToString(hash[:])

	var apiKey models.ApiKey
	if err := db.Where("key_hash = ? AND revoked_at IS NULL", keyHash).First(&apiKey).Error; err != nil {
		return false
	}

	// Look up the user to get their role
	var user models.User
	if err := db.First(&user, apiKey.UserID).Error; err != nil {
		return false
	}

	// Update last used timestamp
	now := time.Now()
	db.Model(&apiKey).Update("last_used_at", &now)

	c.Set("userId", apiKey.UserID)
	c.Set("userRole", string(user.Role))
	return true
}
