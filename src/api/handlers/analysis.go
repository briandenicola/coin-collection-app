package handlers

import (
	"io"
	"net/http"
	"path/filepath"
	"strconv"

	"github.com/briandenicola/ancient-coins-api/database"
	"github.com/briandenicola/ancient-coins-api/models"
	"github.com/briandenicola/ancient-coins-api/services"
	"github.com/gin-gonic/gin"
)

type AnalysisHandler struct{}

func NewAnalysisHandler() *AnalysisHandler {
	return &AnalysisHandler{}
}

func (h *AnalysisHandler) Analyze(c *gin.Context) {
	logger := services.AppLogger
	userID := c.GetUint("userId")
	coinID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		logger.Warn("analysis", "Invalid coin ID param: %s", c.Param("id"))
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid coin ID"})
		return
	}

	logger.Info("analysis", "Starting analysis for coin %d (user %d)", coinID, userID)

	var coin models.Coin
	if err := database.DB.Where("id = ? AND user_id = ?", coinID, userID).Preload("Images").First(&coin).Error; err != nil {
		logger.Warn("analysis", "Coin %d not found for user %d: %v", coinID, userID, err)
		c.JSON(http.StatusNotFound, gin.H{"error": "Coin not found"})
		return
	}

	logger.Debug("analysis", "Coin loaded: %s, images: %d", coin.Name, len(coin.Images))

	if len(coin.Images) == 0 {
		logger.Warn("analysis", "Coin %d has no images to analyze", coinID)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Coin has no images to analyze"})
		return
	}

	// Read Ollama settings from DB (with fallback to env/defaults)
	ollamaURL := services.GetSetting(services.SettingOllamaURL)
	ollamaModel := services.GetSetting(services.SettingOllamaModel)
	customPrompt := services.GetSetting(services.SettingAIPrompt)

	logger.Debug("analysis", "Ollama URL: %s, Model: %s", ollamaURL, ollamaModel)
	logger.Trace("analysis", "Custom prompt length: %d", len(customPrompt))

	ollamaSvc := services.NewOllamaService(ollamaURL)

	var imagePaths []string
	for _, img := range coin.Images {
		p := filepath.Join("uploads", img.FilePath)
		imagePaths = append(imagePaths, p)
		logger.Trace("analysis", "Image path: %s", p)
	}

	logger.Info("analysis", "Sending %d images to Ollama for coin %d", len(imagePaths), coinID)
	analysis, err := ollamaSvc.AnalyzeCoinImages(imagePaths, coin, ollamaModel, customPrompt)
	if err != nil {
		logger.Error("analysis", "AI analysis failed for coin %d: %v", coinID, err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "AI analysis failed: " + err.Error()})
		return
	}

	logger.Info("analysis", "Analysis complete for coin %d (%d chars)", coinID, len(analysis))
	logger.Trace("analysis", "Analysis result: %s", analysis)

	database.DB.Model(&coin).Update("ai_analysis", analysis)
	coin.AIAnalysis = analysis

	c.JSON(http.StatusOK, gin.H{
		"analysis": analysis,
		"coin":     coin,
	})
}

// ExtractText accepts an image upload and returns extracted text via Ollama
func (h *AnalysisHandler) ExtractText(c *gin.Context) {
	logger := services.AppLogger
	logger.Info("extract-text", "Text extraction request received")

	file, err := c.FormFile("image")
	if err != nil {
		logger.Warn("extract-text", "No image file in request: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "No image file provided"})
		return
	}

	logger.Debug("extract-text", "File: %s, size: %d bytes", file.Filename, file.Size)

	f, err := file.Open()
	if err != nil {
		logger.Error("extract-text", "Failed to open uploaded file: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to read image"})
		return
	}
	defer f.Close()

	imageData, err := io.ReadAll(f)
	if err != nil {
		logger.Error("extract-text", "Failed to read image data: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to read image data"})
		return
	}

	ollamaURL := services.GetSetting(services.SettingOllamaURL)
	ollamaModel := services.GetSetting(services.SettingOllamaModel)

	logger.Debug("extract-text", "Sending to Ollama: URL=%s, Model=%s", ollamaURL, ollamaModel)

	ollamaSvc := services.NewOllamaService(ollamaURL)
	text, err := ollamaSvc.ExtractTextFromImage(imageData, ollamaModel)
	if err != nil {
		logger.Error("extract-text", "Text extraction failed: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Text extraction failed: " + err.Error()})
		return
	}

	logger.Info("extract-text", "Extraction complete (%d chars)", len(text))
	logger.Trace("extract-text", "Extracted text: %s", text)

	c.JSON(http.StatusOK, gin.H{"text": text})
}
