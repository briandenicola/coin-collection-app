package handlers

import (
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"time"

	"github.com/briandenicola/ancient-coins-api/database"
	"github.com/briandenicola/ancient-coins-api/models"
	"github.com/briandenicola/ancient-coins-api/services"
	"github.com/gin-gonic/gin"
)

type ImageHandler struct {
	UploadDir string
}

func NewImageHandler(uploadDir string) *ImageHandler {
	return &ImageHandler{UploadDir: uploadDir}
}

func (h *ImageHandler) Upload(c *gin.Context) {
	logger := services.AppLogger
	userID := c.GetUint("userId")
	coinID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid coin ID"})
		return
	}

	logger.Debug("images", "Upload request for coin %d (user %d)", coinID, userID)

	// Verify ownership
	var coin models.Coin
	if err := database.DB.Where("id = ? AND user_id = ?", coinID, userID).First(&coin).Error; err != nil {
		logger.Warn("images", "Coin %d not found for user %d", coinID, userID)
		c.JSON(http.StatusNotFound, gin.H{"error": "Coin not found"})
		return
	}

	file, err := c.FormFile("image")
	if err != nil {
		logger.Warn("images", "No image file in upload request: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "No image file provided"})
		return
	}

	logger.Debug("images", "Received file: %s (%d bytes)", file.Filename, file.Size)

	imageType := models.ImageType(c.DefaultPostForm("imageType", "other"))
	isPrimary := c.DefaultPostForm("isPrimary", "false") == "true"

	// Create upload directory for this coin
	coinDir := filepath.Join(h.UploadDir, fmt.Sprintf("coin-%d", coinID))
	if err := os.MkdirAll(coinDir, 0755); err != nil {
		logger.Error("images", "Failed to create directory %s: %v", coinDir, err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create upload directory"})
		return
	}

	// Generate unique filename
	ext := filepath.Ext(file.Filename)
	filename := fmt.Sprintf("%d-%s%s", time.Now().UnixNano(), imageType, ext)
	filePath := filepath.Join(coinDir, filename)

	if err := c.SaveUploadedFile(file, filePath); err != nil {
		logger.Error("images", "Failed to save file to %s: %v", filePath, err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save image"})
		return
	}

	// If this is primary, unset other primary images
	if isPrimary {
		database.DB.Model(&models.CoinImage{}).Where("coin_id = ?", coinID).Update("is_primary", false)
	}

	image := models.CoinImage{
		CoinID:    uint(coinID),
		FilePath:  filepath.ToSlash(filepath.Join(fmt.Sprintf("coin-%d", coinID), filename)),
		ImageType: imageType,
		IsPrimary: isPrimary,
	}

	if err := database.DB.Create(&image).Error; err != nil {
		logger.Error("images", "Failed to save image record: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save image record"})
		return
	}

	logger.Info("images", "Uploaded %s image for coin %d: %s", imageType, coinID, image.FilePath)
	c.JSON(http.StatusCreated, image)
}

func (h *ImageHandler) Delete(c *gin.Context) {
	userID := c.GetUint("userId")
	coinID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid coin ID"})
		return
	}
	imageID, err := strconv.ParseUint(c.Param("imageId"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid image ID"})
		return
	}

	// Verify ownership
	var coin models.Coin
	if err := database.DB.Where("id = ? AND user_id = ?", coinID, userID).First(&coin).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Coin not found"})
		return
	}

	var image models.CoinImage
	if err := database.DB.Where("id = ? AND coin_id = ?", imageID, coinID).First(&image).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Image not found"})
		return
	}

	// Delete file from disk
	os.Remove(filepath.Join(h.UploadDir, image.FilePath))

	database.DB.Delete(&image)
	c.JSON(http.StatusOK, gin.H{"message": "Image deleted"})
}
