package handlers

import (
	"encoding/base64"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
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

// Upload adds an image to a coin.
//
//	@Summary		Upload a coin image
//	@Description	Upload an image file for a specific coin. Supports setting image type and primary flag.
//	@Tags			Images
//	@Accept			multipart/form-data
//	@Produce		json
//	@Param			id			path		int		true	"Coin ID"
//	@Param			image		formData	file	true	"Image file"
//	@Param			imageType	formData	string	false	"Image type"	Enums(obverse, reverse, detail, other)	default(other)
//	@Param			isPrimary	formData	string	false	"Set as primary image"	Enums(true, false)	default(false)
//	@Success		201			{object}	models.CoinImage
//	@Failure		400			{object}	ErrorResponse
//	@Failure		401			{object}	ErrorResponse
//	@Failure		404			{object}	ErrorResponse
//	@Failure		500			{object}	ErrorResponse
//	@Security		BearerAuth
//	@Router			/coins/{id}/images [post]
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

type base64ImageRequest struct {
	Image         string `json:"image" binding:"required" example:"/9j/4AAQSkZJRgABAQ..."`
	FileExtension string `json:"fileExtension" binding:"required" example:".jpg"`
	ImageType     string `json:"imageType" example:"obverse"`
	IsPrimary     bool   `json:"isPrimary" example:"false"`
}

// UploadBase64 adds an image to a coin from a base64-encoded string.
//
//	@Summary		Upload a coin image (base64)
//	@Description	Upload a raw base64-encoded image for a specific coin. The image field must contain only raw base64 data (no data URI prefix). The fileExtension field (e.g. ".jpg", ".png") tells the API how to save the file. Max 20MB decoded.
//	@Tags			Images
//	@Accept			json
//	@Produce		json
//	@Param			id		path		int					true	"Coin ID"
//	@Param			body	body		base64ImageRequest	true	"Base64 image data"
//	@Success		201		{object}	models.CoinImage
//	@Failure		400		{object}	ErrorResponse
//	@Failure		401		{object}	ErrorResponse
//	@Failure		404		{object}	ErrorResponse
//	@Failure		500		{object}	ErrorResponse
//	@Security		BearerAuth
//	@Security		ApiKeyAuth
//	@Router			/coins/{id}/images/base64 [post]
func (h *ImageHandler) UploadBase64(c *gin.Context) {
	logger := services.AppLogger
	userID := c.GetUint("userId")
	coinID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid coin ID"})
		return
	}

	var req base64ImageRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Validate file extension
	ext := req.FileExtension
	if !strings.HasPrefix(ext, ".") {
		ext = "." + ext
	}
	allowed := map[string]bool{".jpg": true, ".jpeg": true, ".png": true, ".gif": true, ".webp": true}
	if !allowed[strings.ToLower(ext)] {
		c.JSON(http.StatusBadRequest, gin.H{"error": "fileExtension must be one of: .jpg, .jpeg, .png, .gif, .webp"})
		return
	}

	// Verify ownership
	var coin models.Coin
	if err := database.DB.Where("id = ? AND user_id = ?", coinID, userID).First(&coin).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Coin not found"})
		return
	}

	decoded, err := base64.StdEncoding.DecodeString(req.Image)
	if err != nil {
		decoded, err = base64.RawStdEncoding.DecodeString(req.Image)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid base64 image data"})
			return
		}
	}

	const maxSize = 20 * 1024 * 1024
	if len(decoded) > maxSize {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Image exceeds 20MB limit"})
		return
	}

	logger.Debug("images", "Base64 upload for coin %d: %d bytes decoded", coinID, len(decoded))

	imageType := models.ImageType("other")
	if req.ImageType != "" {
		imageType = models.ImageType(req.ImageType)
	}

	coinDir := filepath.Join(h.UploadDir, fmt.Sprintf("coin-%d", coinID))
	if err := os.MkdirAll(coinDir, 0755); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create upload directory"})
		return
	}

	filename := fmt.Sprintf("%d-%s%s", time.Now().UnixNano(), imageType, ext)
	filePath := filepath.Join(coinDir, filename)

	if err := os.WriteFile(filePath, decoded, 0644); err != nil {
		logger.Error("images", "Failed to write base64 image to %s: %v", filePath, err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save image"})
		return
	}

	if req.IsPrimary {
		database.DB.Model(&models.CoinImage{}).Where("coin_id = ?", coinID).Update("is_primary", false)
	}

	image := models.CoinImage{
		CoinID:    uint(coinID),
		FilePath:  filepath.ToSlash(filepath.Join(fmt.Sprintf("coin-%d", coinID), filename)),
		ImageType: imageType,
		IsPrimary: req.IsPrimary,
	}

	if err := database.DB.Create(&image).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save image record"})
		return
	}

	logger.Info("images", "Uploaded base64 %s image for coin %d: %s", imageType, coinID, image.FilePath)
	c.JSON(http.StatusCreated, image)
}

// Delete removes an image from a coin.
//
//	@Summary		Delete a coin image
//	@Description	Deletes an image from a coin. Removes the file from disk and the database record.
//	@Tags			Images
//	@Produce		json
//	@Param			id		path		int	true	"Coin ID"
//	@Param			imageId	path		int	true	"Image ID"
//	@Success		200		{object}	ImageDeletedResponse
//	@Failure		400		{object}	ErrorResponse
//	@Failure		401		{object}	ErrorResponse
//	@Failure		404		{object}	ErrorResponse
//	@Security		BearerAuth
//	@Router			/coins/{id}/images/{imageId} [delete]
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

// ProxyImage fetches an external image URL and streams it back to the client.
//
//	@Summary		Proxy an external image
//	@Description	Fetches an image from an external URL and streams it to the client. Limited to 20MB. Only http/https URLs with image content types are allowed.
//	@Tags			Images
//	@Produce		image/*
//	@Param			url	query	string	true	"External image URL"
//	@Success		200	"Image binary data"
//	@Failure		400	{object}	ErrorResponse
//	@Failure		401	{object}	ErrorResponse
//	@Failure		502	{object}	ErrorResponse
//	@Security		BearerAuth
//	@Router			/proxy-image [get]
func (h *ImageHandler) ProxyImage(c *gin.Context) {
	logger := services.AppLogger

	imageURL := c.Query("url")
	if imageURL == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "url parameter is required"})
		return
	}

	if !strings.HasPrefix(imageURL, "http://") && !strings.HasPrefix(imageURL, "https://") {
		c.JSON(http.StatusBadRequest, gin.H{"error": "URL must start with http:// or https://"})
		return
	}

	logger.Debug("images", "Proxying image from %s", imageURL)

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Get(imageURL)
	if err != nil {
		logger.Warn("images", "Failed to fetch image from %s: %v", imageURL, err)
		c.JSON(http.StatusBadGateway, gin.H{"error": "Failed to fetch image"})
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		c.JSON(http.StatusBadGateway, gin.H{"error": fmt.Sprintf("Remote server returned %d", resp.StatusCode)})
		return
	}

	contentType := resp.Header.Get("Content-Type")
	if !strings.HasPrefix(contentType, "image/") {
		c.JSON(http.StatusBadRequest, gin.H{"error": "URL does not point to an image"})
		return
	}

	// Limit to 20MB
	const maxSize = 20 * 1024 * 1024
	c.Header("Content-Type", contentType)
	c.Status(http.StatusOK)
	io.Copy(c.Writer, io.LimitReader(resp.Body, maxSize))
}
