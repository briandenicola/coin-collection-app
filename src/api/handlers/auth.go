package handlers

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"net/http"
	"time"

	"github.com/briandenicola/ancient-coins-api/database"
	"github.com/briandenicola/ancient-coins-api/models"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

const (
	accessTokenDuration  = 15 * time.Minute
	refreshTokenDuration = 30 * 24 * time.Hour // 30 days
)

type AuthHandler struct {
	JWTSecret string
}

type loginRequest struct {
	Username string `json:"username" binding:"required" example:"admin"`
	Password string `json:"password" binding:"required" example:"password123"`
}

type registerRequest struct {
	Username string `json:"username" binding:"required,min=3" example:"admin"`
	Password string `json:"password" binding:"required,min=6" example:"password123"`
	Email    string `json:"email" binding:"required,email" example:"user@example.com"`
}

type refreshRequest struct {
	RefreshToken string `json:"refreshToken" binding:"required"`
}

func NewAuthHandler(jwtSecret string) *AuthHandler {
	return &AuthHandler{JWTSecret: jwtSecret}
}

// Register creates a new user account. The first user registered becomes an admin.
//
//	@Summary		Register a new user
//	@Description	Create a new user account. The first registered user is assigned the admin role. Returns access token (15min) and refresh token (30 days).
//	@Tags			Auth
//	@Accept			json
//	@Produce		json
//	@Param			body	body		registerRequest	true	"Registration credentials"
//	@Success		201		{object}	AuthResponse
//	@Failure		400		{object}	ErrorResponse
//	@Failure		409		{object}	ErrorResponse
//	@Failure		500		{object}	ErrorResponse
//	@Router			/auth/register [post]
func (h *AuthHandler) Register(c *gin.Context) {
	var req registerRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to hash password"})
		return
	}

	// First user becomes admin
	var count int64
	database.DB.Model(&models.User{}).Count(&count)
	role := models.RoleUser
	if count == 0 {
		role = models.RoleAdmin
	}

	user := models.User{
		Username:     req.Username,
		Email:        req.Email,
		PasswordHash: string(hash),
		Role:         role,
	}

	if err := database.DB.Create(&user).Error; err != nil {
		c.JSON(http.StatusConflict, gin.H{"error": "Username already exists"})
		return
	}

	h.issueTokens(c, user, http.StatusCreated)
}

// Login authenticates a user and returns JWT access and refresh tokens.
//
//	@Summary		Login
//	@Description	Authenticate with username and password. Returns access token (15min) and refresh token (30 days).
//	@Tags			Auth
//	@Accept			json
//	@Produce		json
//	@Param			body	body		loginRequest	true	"Login credentials"
//	@Success		200		{object}	AuthResponse
//	@Failure		400		{object}	ErrorResponse
//	@Failure		401		{object}	ErrorResponse
//	@Failure		500		{object}	ErrorResponse
//	@Router			/auth/login [post]
func (h *AuthHandler) Login(c *gin.Context) {
	var req loginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var user models.User
	if err := database.DB.Where("username = ?", req.Username).First(&user).Error; err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid credentials"})
		return
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.Password)); err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid credentials"})
		return
	}

	h.issueTokens(c, user, http.StatusOK)
}

// Refresh exchanges a valid refresh token for a new access token and refresh token (rolling).
//
//	@Summary		Refresh tokens
//	@Description	Exchange a valid refresh token for new access and refresh tokens. The old refresh token is revoked (rolling refresh).
//	@Tags			Auth
//	@Accept			json
//	@Produce		json
//	@Param			body	body		refreshRequest	true	"Refresh token"
//	@Success		200		{object}	AuthResponse
//	@Failure		400		{object}	ErrorResponse
//	@Failure		401		{object}	ErrorResponse
//	@Failure		500		{object}	ErrorResponse
//	@Router			/auth/refresh [post]
func (h *AuthHandler) Refresh(c *gin.Context) {
	var req refreshRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Hash the provided token and look it up
	hash := sha256.Sum256([]byte(req.RefreshToken))
	tokenHash := hex.EncodeToString(hash[:])

	var rt models.RefreshToken
	if err := database.DB.Where("token_hash = ? AND revoked_at IS NULL", tokenHash).First(&rt).Error; err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid refresh token"})
		return
	}

	// Check expiry
	if time.Now().After(rt.ExpiresAt) {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Refresh token expired"})
		return
	}

	// Revoke the old refresh token
	now := time.Now()
	database.DB.Model(&rt).Update("revoked_at", &now)

	// Look up user
	var user models.User
	if err := database.DB.First(&user, rt.UserID).Error; err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not found"})
		return
	}

	h.issueTokens(c, user, http.StatusOK)
}

// issueTokens generates and returns both access and refresh tokens.
func (h *AuthHandler) issueTokens(c *gin.Context, user models.User, statusCode int) {
	accessToken, err := h.generateAccessToken(user)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate access token"})
		return
	}

	refreshToken, err := h.generateRefreshToken(user)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate refresh token"})
		return
	}

	c.JSON(statusCode, gin.H{
		"token":        accessToken,
		"refreshToken": refreshToken,
		"user": gin.H{
			"id":         user.ID,
			"username":   user.Username,
			"role":       user.Role,
			"email":      user.Email,
			"avatarPath": user.AvatarPath,
			"isPublic":   user.IsPublic,
			"bio":        user.Bio,
		},
	})
}

func (h *AuthHandler) generateAccessToken(user models.User) (string, error) {
	claims := jwt.MapClaims{
		"userId":   user.ID,
		"username": user.Username,
		"role":     string(user.Role),
		"exp":      time.Now().Add(accessTokenDuration).Unix(),
		"iat":      time.Now().Unix(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(h.JWTSecret))
}

func (h *AuthHandler) generateRefreshToken(user models.User) (string, error) {
	// Generate a random 32-byte token
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	plainToken := "rt_" + hex.EncodeToString(b)

	// Store hashed version in DB
	hash := sha256.Sum256([]byte(plainToken))
	tokenHash := hex.EncodeToString(hash[:])

	rt := models.RefreshToken{
		UserID:    user.ID,
		TokenHash: tokenHash,
		ExpiresAt: time.Now().Add(refreshTokenDuration),
	}
	if err := database.DB.Create(&rt).Error; err != nil {
		return "", err
	}

	return plainToken, nil
}

// NeedsSetup returns whether the first user has been created yet.
//
//	@Summary		Check setup status
//	@Description	Returns whether the application needs initial setup (no users exist yet).
//	@Tags			Auth
//	@Produce		json
//	@Success		200	{object}	SetupResponse
//	@Router			/auth/setup [get]
func (h *AuthHandler) NeedsSetup(c *gin.Context) {
	var count int64
	database.DB.Model(&models.User{}).Count(&count)
	c.JSON(http.StatusOK, gin.H{"needsSetup": count == 0})
}
