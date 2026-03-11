package handlers

import (
	"net/http"
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
	userID := c.GetUint("userId")
	coinID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid coin ID"})
		return
	}

	var coin models.Coin
	if err := database.DB.Where("id = ? AND user_id = ?", coinID, userID).Preload("Images").First(&coin).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Coin not found"})
		return
	}

	if len(coin.Images) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Coin has no images to analyze"})
		return
	}

	// Read Ollama settings from DB (with fallback to env/defaults)
	ollamaURL := services.GetSetting(services.SettingOllamaURL)
	ollamaModel := services.GetSetting(services.SettingOllamaModel)
	customPrompt := services.GetSetting(services.SettingAIPrompt)

	ollamaSvc := services.NewOllamaService(ollamaURL)

	var imagePaths []string
	for _, img := range coin.Images {
		imagePaths = append(imagePaths, img.FilePath)
	}

	analysis, err := ollamaSvc.AnalyzeCoinImages(imagePaths, coin, ollamaModel, customPrompt)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "AI analysis failed: " + err.Error()})
		return
	}

	database.DB.Model(&coin).Update("ai_analysis", analysis)
	coin.AIAnalysis = analysis

	c.JSON(http.StatusOK, gin.H{
		"analysis": analysis,
		"coin":     coin,
	})
}
