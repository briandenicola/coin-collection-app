package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/briandenicola/ancient-coins-api/database"
	"github.com/briandenicola/ancient-coins-api/models"
	"github.com/briandenicola/ancient-coins-api/services"
	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
)

type AdminHandler struct {
	UploadDir string
}

func NewAdminHandler(uploadDir string) *AdminHandler {
	return &AdminHandler{UploadDir: uploadDir}
}

// AdminRequired middleware ensures only admin users can access
func AdminRequired() gin.HandlerFunc {
	return func(c *gin.Context) {
		role, _ := c.Get("userRole")
		if role != string(models.RoleAdmin) {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "Admin access required"})
			return
		}
		c.Next()
	}
}

// ListUsers returns all users for admin management.
//
//	@Summary		List all users
//	@Description	Returns a list of all registered users. Admin only.
//	@Tags			Admin
//	@Produce		json
//	@Success		200	{array}		UserDTO
//	@Failure		401	{object}	ErrorResponse
//	@Failure		403	{object}	ErrorResponse
//	@Security		BearerAuth
//	@Router			/admin/users [get]
func (h *AdminHandler) ListUsers(c *gin.Context) {
	var users []models.User
	database.DB.Find(&users)

	type userDTO struct {
		ID        uint           `json:"id"`
		Username  string         `json:"username"`
		Role      models.UserRole `json:"role"`
		CreatedAt string         `json:"createdAt"`
	}

	var result []userDTO
	for _, u := range users {
		result = append(result, userDTO{
			ID:        u.ID,
			Username:  u.Username,
			Role:      u.Role,
			CreatedAt: u.CreatedAt.Format("2006-01-02T15:04:05Z"),
		})
	}

	c.JSON(http.StatusOK, result)
}

// DeleteUser removes a user and their coins.
//
//	@Summary		Delete a user
//	@Description	Deletes a user and all their associated coins and images. Cannot delete yourself. Admin only.
//	@Tags			Admin
//	@Produce		json
//	@Param			id	path		int	true	"User ID"
//	@Success		200	{object}	MessageResponse
//	@Failure		400	{object}	ErrorResponse
//	@Failure		401	{object}	ErrorResponse
//	@Failure		403	{object}	ErrorResponse
//	@Failure		404	{object}	ErrorResponse
//	@Security		BearerAuth
//	@Router			/admin/users/{id} [delete]
func (h *AdminHandler) DeleteUser(c *gin.Context) {
	adminID := c.GetUint("userId")
	targetID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
		return
	}

	if uint(targetID) == adminID {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Cannot delete your own account"})
		return
	}

	// Delete user's coin images, coins, then user
	var coinIDs []uint
	database.DB.Model(&models.Coin{}).Where("user_id = ?", targetID).Pluck("id", &coinIDs)
	if len(coinIDs) > 0 {
		database.DB.Where("coin_id IN ?", coinIDs).Delete(&models.CoinImage{})
	}
	database.DB.Where("user_id = ?", targetID).Delete(&models.Coin{})

	result := database.DB.Delete(&models.User{}, targetID)
	if result.RowsAffected == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "User deleted"})
}

// ResetPassword allows admin to set a new password for a user.
//
//	@Summary		Reset user password
//	@Description	Sets a new password for the specified user. Admin only.
//	@Tags			Admin
//	@Accept			json
//	@Produce		json
//	@Param			id		path		int						true	"User ID"
//	@Param			body	body		ResetPasswordRequest	true	"New password"
//	@Success		200		{object}	MessageResponse
//	@Failure		400		{object}	ErrorResponse
//	@Failure		401		{object}	ErrorResponse
//	@Failure		403		{object}	ErrorResponse
//	@Failure		404		{object}	ErrorResponse
//	@Failure		500		{object}	ErrorResponse
//	@Security		BearerAuth
//	@Router			/admin/users/{id}/reset-password [post]
func (h *AdminHandler) ResetPassword(c *gin.Context) {
	targetID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
		return
	}

	var req struct {
		NewPassword string `json:"newPassword" binding:"required,min=6"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(req.NewPassword), bcrypt.DefaultCost)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to hash password"})
		return
	}

	result := database.DB.Model(&models.User{}).Where("id = ?", targetID).Update("password_hash", string(hash))
	if result.RowsAffected == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Password reset"})
}

// GetSettings returns all app settings.
//
//	@Summary		Get application settings
//	@Description	Returns all application settings merged with defaults. Admin only.
//	@Tags			Admin
//	@Produce		json
//	@Success		200	{object}	map[string]string
//	@Failure		401	{object}	ErrorResponse
//	@Failure		403	{object}	ErrorResponse
//	@Security		BearerAuth
//	@Router			/admin/settings [get]
func (h *AdminHandler) GetSettings(c *gin.Context) {
	settings := services.GetAllSettings()
	c.JSON(http.StatusOK, settings)
}

// GetSettingDefaults returns the built-in default values for all settings.
//
//	@Summary		Get setting defaults
//	@Description	Returns the built-in default values for all application settings. Admin only.
//	@Tags			Admin
//	@Produce		json
//	@Success		200	{object}	map[string]string
//	@Failure		401	{object}	ErrorResponse
//	@Failure		403	{object}	ErrorResponse
//	@Security		BearerAuth
//	@Router			/admin/settings/defaults [get]
func (h *AdminHandler) GetSettingDefaults(c *gin.Context) {
	c.JSON(http.StatusOK, services.GetSettingDefaults())
}

// UpdateSettings updates app settings.
//
//	@Summary		Update application settings
//	@Description	Updates one or more application settings. Syncs log level if changed. Admin only.
//	@Tags			Admin
//	@Accept			json
//	@Produce		json
//	@Param			body	body		[]SettingInput	true	"Settings to update"
//	@Success		200		{object}	SettingsUpdateResponse
//	@Failure		400		{object}	ErrorResponse
//	@Failure		401		{object}	ErrorResponse
//	@Failure		403		{object}	ErrorResponse
//	@Failure		500		{object}	ErrorResponse
//	@Security		BearerAuth
//	@Router			/admin/settings [put]
func (h *AdminHandler) UpdateSettings(c *gin.Context) {
	var settings []struct {
		Key   string `json:"key" binding:"required"`
		Value string `json:"value"`
	}
	if err := c.ShouldBindJSON(&settings); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	for _, s := range settings {
		if err := services.SetSetting(s.Key, s.Value); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save setting: " + s.Key})
			return
		}
	}

	// Sync log level if it was updated
	services.SyncLogLevel()

	c.JSON(http.StatusOK, gin.H{"message": "Settings updated"})
}

// GetLogs returns recent application logs.
//
//	@Summary		Get application logs
//	@Description	Returns recent in-memory application logs, optionally filtered by level. Admin only.
//	@Tags			Admin
//	@Produce		json
//	@Param			limit	query		int		false	"Maximum number of log entries"	default(500)
//	@Param			level	query		string	false	"Filter by log level"	Enums(trace, debug, info, warn, error)
//	@Success		200		{object}	LogsResponse
//	@Failure		401		{object}	ErrorResponse
//	@Failure		403		{object}	ErrorResponse
//	@Security		BearerAuth
//	@Router			/admin/logs [get]
func (h *AdminHandler) GetLogs(c *gin.Context) {
	limit := 500
	if l := c.Query("limit"); l != "" {
		if n, err := strconv.Atoi(l); err == nil && n > 0 {
			limit = n
		}
	}

	level := c.Query("level")
	logs := services.AppLogger.GetLogs(limit)

	if level != "" {
		filtered := make([]services.LogEntry, 0)
		for _, entry := range logs {
			if entry.Level == level {
				filtered = append(filtered, entry)
			}
		}
		logs = filtered
	}

	c.JSON(http.StatusOK, gin.H{
		"logs":     logs,
		"count":    len(logs),
		"logLevel": services.AppLogger.GetLevel(),
	})
}

// ExportAllData exports all coins and images as a zip archive
func (h *AdminHandler) ExportAllData(c *gin.Context) {
	var coins []models.Coin
	database.DB.Preload("Images").Find(&coins)

	filename := fmt.Sprintf("ancient-coins-export-%s.zip", time.Now().Format("2006-01-02"))
	writeCollectionZip(c, coins, h.UploadDir, filename)
}

// ImportCollection imports coins from JSON for a specific user (admin)
func (h *AdminHandler) ImportData(c *gin.Context) {
	var coins []models.Coin
	if err := json.NewDecoder(c.Request.Body).Decode(&coins); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid JSON"})
		return
	}

	imported := 0
	for _, coin := range coins {
		coin.ID = 0
		if err := database.DB.Create(&coin).Error; err == nil {
			imported++
		}
	}

	c.JSON(http.StatusOK, gin.H{"message": "Import complete", "imported": imported})
}
