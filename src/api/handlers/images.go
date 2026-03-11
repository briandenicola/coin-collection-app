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
	"github.com/gin-gonic/gin"
)

type ImageHandler struct {
	UploadDir string
}

func NewImageHandler(uploadDir string) *ImageHandler {
	return &ImageHandler{UploadDir: uploadDir}
}

func (h *ImageHandler) Upload(c *gin.Context) {
	userID := c.GetUint("userId")
	coinID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid coin ID"})
		return
	}

	// Verify ownership
	var coin models.Coin
	if err := database.DB.Where("id = ? AND user_id = ?", coinID, userID).First(&coin).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Coin not found"})
		return
	}

	file, err := c.FormFile("image")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "No image file provided"})
		return
	}

	imageType := models.ImageType(c.DefaultPostForm("imageType", "other"))
	isPrimary := c.DefaultPostForm("isPrimary", "false") == "true"

	// Create upload directory for this coin
	coinDir := filepath.Join(h.UploadDir, fmt.Sprintf("coin-%d", coinID))
	if err := os.MkdirAll(coinDir, 0755); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create upload directory"})
		return
	}

	// Generate unique filename
	ext := filepath.Ext(file.Filename)
	filename := fmt.Sprintf("%d-%s%s", time.Now().UnixNano(), imageType, ext)
	filePath := filepath.Join(coinDir, filename)

	if err := c.SaveUploadedFile(file, filePath); err != nil {
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
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save image record"})
		return
	}

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
