package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/briandenicola/ancient-coins-api/database"
	"github.com/briandenicola/ancient-coins-api/models"
	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
)

type UserHandler struct {
	UploadDir string
}

func NewUserHandler(uploadDir string) *UserHandler {
	return &UserHandler{UploadDir: uploadDir}
}

// ChangePassword allows a user to change their own password.
//
//	@Summary		Change password
//	@Description	Change the authenticated user's password. Requires the current password for verification.
//	@Tags			User
//	@Accept			json
//	@Produce		json
//	@Param			body	body		ChangePasswordRequest	true	"Password change request"
//	@Success		200		{object}	PasswordChangedResponse
//	@Failure		400		{object}	ErrorResponse
//	@Failure		401		{object}	ErrorResponse
//	@Failure		404		{object}	ErrorResponse
//	@Failure		500		{object}	ErrorResponse
//	@Security		BearerAuth
//	@Router			/auth/change-password [post]
func (h *UserHandler) ChangePassword(c *gin.Context) {
	userID := c.GetUint("userId")

	var req struct {
		CurrentPassword string `json:"currentPassword" binding:"required"`
		NewPassword     string `json:"newPassword" binding:"required,min=6"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var user models.User
	if err := database.DB.First(&user, userID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.CurrentPassword)); err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Current password is incorrect"})
		return
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(req.NewPassword), bcrypt.DefaultCost)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to hash password"})
		return
	}

	database.DB.Model(&user).Update("password_hash", string(hash))
	c.JSON(http.StatusOK, gin.H{"message": "Password changed"})
}

// GetMe returns the current authenticated user's info.
//
//	@Summary		Get current user
//	@Description	Returns profile information for the authenticated user.
//	@Tags			User
//	@Produce		json
//	@Success		200	{object}	UserInfoResponse
//	@Failure		401	{object}	ErrorResponse
//	@Failure		404	{object}	ErrorResponse
//	@Security		BearerAuth
//	@Router			/auth/me [get]
func (h *UserHandler) GetMe(c *gin.Context) {
	userID := c.GetUint("userId")

	var user models.User
	if err := database.DB.First(&user, userID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"id":        user.ID,
		"username":  user.Username,
		"role":      user.Role,
		"createdAt": user.CreatedAt,
	})
}

// ExportCollection exports the current user's coins and images as a zip archive.
//
//	@Summary		Export collection
//	@Description	Downloads all coins and images as a ZIP archive containing coins.json and image files.
//	@Tags			User
//	@Produce		application/zip
//	@Success		200	"ZIP archive"
//	@Failure		401	{object}	ErrorResponse
//	@Security		BearerAuth
//	@Router			/user/export [get]
func (h *UserHandler) ExportCollection(c *gin.Context) {
	userID := c.GetUint("userId")

	var coins []models.Coin
	database.DB.Where("user_id = ?", userID).Preload("Images").Find(&coins)

	filename := fmt.Sprintf("my-coins-export-%s.zip", time.Now().Format("2006-01-02"))
	writeCollectionZip(c, coins, h.UploadDir, filename)
}

// ImportCollection imports coins from JSON for the current user.
//
//	@Summary		Import collection
//	@Description	Imports coins from a JSON array. Each coin is created fresh with a new ID for the authenticated user.
//	@Tags			User
//	@Accept			json
//	@Produce		json
//	@Param			body	body		[]models.Coin	true	"Array of coins to import"
//	@Success		200		{object}	ImportResponse
//	@Failure		400		{object}	ErrorResponse
//	@Failure		401		{object}	ErrorResponse
//	@Security		BearerAuth
//	@Router			/user/import [post]
func (h *UserHandler) ImportCollection(c *gin.Context) {
	userID := c.GetUint("userId")

	var coins []models.Coin
	if err := json.NewDecoder(c.Request.Body).Decode(&coins); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid JSON format"})
		return
	}

	imported := 0
	for _, coin := range coins {
		coin.ID = 0
		coin.UserID = userID
		// Reset image associations for clean import
		coin.Images = nil
		if err := database.DB.Create(&coin).Error; err == nil {
			imported++
		}
	}

	c.JSON(http.StatusOK, gin.H{"message": "Import complete", "imported": imported})
}
