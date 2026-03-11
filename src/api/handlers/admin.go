package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/briandenicola/ancient-coins-api/database"
	"github.com/briandenicola/ancient-coins-api/models"
	"github.com/briandenicola/ancient-coins-api/services"
	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
)

type AdminHandler struct{}

func NewAdminHandler() *AdminHandler {
	return &AdminHandler{}
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

// ListUsers returns all users for admin management
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

// DeleteUser removes a user and their coins
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

// ResetPassword allows admin to set a new password for a user
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

// GetSettings returns all app settings
func (h *AdminHandler) GetSettings(c *gin.Context) {
	settings := services.GetAllSettings()
	c.JSON(http.StatusOK, settings)
}

// UpdateSettings updates app settings
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

	c.JSON(http.StatusOK, gin.H{"message": "Settings updated"})
}

// ExportCollection returns all coins for a user as JSON
func (h *AdminHandler) ExportAllData(c *gin.Context) {
	var coins []models.Coin
	database.DB.Preload("Images").Find(&coins)

	c.Header("Content-Disposition", "attachment; filename=ancient-coins-export.json")
	c.JSON(http.StatusOK, coins)
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
