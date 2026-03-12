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

	// Filter by side if requested
	side := c.Query("side")
	var analyzeImages []models.CoinImage
	if side == "obverse" || side == "reverse" {
		for _, img := range coin.Images {
			if string(img.ImageType) == side {
				analyzeImages = append(analyzeImages, img)
			}
		}
		if len(analyzeImages) == 0 {
			logger.Warn("analysis", "Coin %d has no %s image", coinID, side)
			c.JSON(http.StatusBadRequest, gin.H{"error": "No " + side + " image found"})
			return
		}
		logger.Info("analysis", "Analyzing %s side only for coin %d", side, coinID)
	} else {
		analyzeImages = coin.Images
	}

	// Read Ollama settings from DB (with fallback to env/defaults)
	ollamaURL := services.GetSetting(services.SettingOllamaURL)
	ollamaModel := services.GetSetting(services.SettingOllamaModel)
	ollamaTimeout, _ := strconv.Atoi(services.GetSetting(services.SettingOllamaTimeout))

	// Pick the prompt based on which side is being analyzed
	var customPrompt string
	switch side {
	case "obverse":
		customPrompt = services.GetSetting(services.SettingObversePrompt)
	case "reverse":
		customPrompt = services.GetSetting(services.SettingReversePrompt)
	default:
		customPrompt = services.GetSetting(services.SettingObversePrompt)
	}

	logger.Debug("analysis", "Ollama URL: %s, Model: %s, Timeout: %ds", ollamaURL, ollamaModel, ollamaTimeout)
	logger.Debug("analysis", "Side: %s, Custom prompt: [%s]", side, customPrompt)

	ollamaSvc := services.NewOllamaService(ollamaURL, ollamaTimeout)

	var imagePaths []string
	for _, img := range analyzeImages {
		p := filepath.Join("uploads", img.FilePath)
		imagePaths = append(imagePaths, p)
		logger.Trace("analysis", "Image path: %s", p)
	}

	logger.Info("analysis", "Sending %d image(s) to Ollama for coin %d", len(imagePaths), coinID)
	analysis, err := ollamaSvc.AnalyzeCoinImages(imagePaths, coin, ollamaModel, customPrompt)
	if err != nil {
		logger.Error("analysis", "AI analysis failed for coin %d: %v", coinID, err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "AI analysis failed: " + err.Error()})
		return
	}

	logger.Info("analysis", "Analysis complete for coin %d (%d chars)", coinID, len(analysis))
	logger.Trace("analysis", "Analysis result: %s", analysis)

	// Store in the appropriate field based on side
	switch side {
	case "obverse":
		database.DB.Model(&coin).Update("obverse_analysis", analysis)
		coin.ObverseAnalysis = analysis
	case "reverse":
		database.DB.Model(&coin).Update("reverse_analysis", analysis)
		coin.ReverseAnalysis = analysis
	default:
		database.DB.Model(&coin).Update("ai_analysis", analysis)
		coin.AIAnalysis = analysis
	}

	c.JSON(http.StatusOK, gin.H{
		"analysis": analysis,
		"side":     side,
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
	ollamaTimeout, _ := strconv.Atoi(services.GetSetting(services.SettingOllamaTimeout))
	customPrompt := services.GetSetting(services.SettingTextExtractionPrompt)

	logger.Debug("extract-text", "Sending to Ollama: URL=%s, Model=%s, Timeout=%ds", ollamaURL, ollamaModel, ollamaTimeout)
	logger.Debug("extract-text", "Custom extraction prompt: [%s]", customPrompt)

	ollamaSvc := services.NewOllamaService(ollamaURL, ollamaTimeout)
	text, err := ollamaSvc.ExtractTextFromImage(imageData, ollamaModel, customPrompt)
	if err != nil {
		logger.Error("extract-text", "Text extraction failed: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Text extraction failed: " + err.Error()})
		return
	}

	logger.Info("extract-text", "Extraction complete (%d chars)", len(text))
	logger.Trace("extract-text", "Extracted text: %s", text)

	c.JSON(http.StatusOK, gin.H{"text": text})
}

// OllamaStatus checks Ollama connectivity and model availability
func (h *AnalysisHandler) OllamaStatus(c *gin.Context) {
	logger := services.AppLogger
	logger.Debug("ollama-status", "Checking Ollama status")

	ollamaURL := services.GetSetting(services.SettingOllamaURL)
	ollamaModel := services.GetSetting(services.SettingOllamaModel)

	ollamaSvc := services.NewOllamaService(ollamaURL, 10)
	available, message := ollamaSvc.CheckModel(ollamaModel)

	logger.Info("ollama-status", "Ollama available=%v, model=%s, message=%s", available, ollamaModel, message)

	c.JSON(http.StatusOK, gin.H{
		"available": available,
		"model":     ollamaModel,
		"url":       ollamaURL,
		"message":   message,
	})
}
