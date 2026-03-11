package handlers

import (
	"net/http"
	"strconv"

	"github.com/briandenicola/ancient-coins-api/database"
	"github.com/briandenicola/ancient-coins-api/models"
	"github.com/briandenicola/ancient-coins-api/services"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type CoinHandler struct{}

func NewCoinHandler() *CoinHandler {
	return &CoinHandler{}
}

func (h *CoinHandler) List(c *gin.Context) {
	userID := c.GetUint("userId")
	category := c.Query("category")
	search := c.Query("search")
	wishlist := c.Query("wishlist")
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "50"))

	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 50
	}
	offset := (page - 1) * limit

	query := database.DB.Where("user_id = ?", userID).Preload("Images")

	if category != "" {
		query = query.Where("category = ?", category)
	}

	if wishlist == "true" {
		query = query.Where("is_wishlist = ?", true)
	} else if wishlist == "false" {
		query = query.Where("is_wishlist = ?", false)
	}

	if search != "" {
		searchTerm := "%" + search + "%"
		query = query.Where(
			database.DB.Where("name LIKE ?", searchTerm).
				Or("denomination LIKE ?", searchTerm).
				Or("ruler LIKE ?", searchTerm).
				Or("era LIKE ?", searchTerm).
				Or("mint LIKE ?", searchTerm).
				Or("obverse_inscription LIKE ?", searchTerm).
				Or("reverse_inscription LIKE ?", searchTerm).
				Or("notes LIKE ?", searchTerm).
				Or("rarity_rating LIKE ?", searchTerm),
		)
	}

	var total int64
	query.Model(&models.Coin{}).Count(&total)

	var coins []models.Coin
	if err := query.Order("updated_at DESC").Offset(offset).Limit(limit).Find(&coins).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch coins"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"coins": coins,
		"total": total,
		"page":  page,
		"limit": limit,
	})
}

func (h *CoinHandler) Get(c *gin.Context) {
	userID := c.GetUint("userId")
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid coin ID"})
		return
	}

	var coin models.Coin
	if err := database.DB.Where("id = ? AND user_id = ?", id, userID).Preload("Images").First(&coin).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "Coin not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch coin"})
		return
	}

	c.JSON(http.StatusOK, coin)
}

func (h *CoinHandler) Create(c *gin.Context) {
	logger := services.AppLogger
	userID := c.GetUint("userId")

	var coin models.Coin
	if err := c.ShouldBindJSON(&coin); err != nil {
		logger.Warn("coins", "Create failed - invalid JSON (user %d): %v", userID, err)
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	coin.UserID = userID
	coin.ID = 0

	logger.Debug("coins", "Creating coin '%s' for user %d", coin.Name, userID)

	if err := database.DB.Create(&coin).Error; err != nil {
		logger.Error("coins", "Failed to create coin: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create coin"})
		return
	}

	logger.Info("coins", "Created coin %d '%s' for user %d", coin.ID, coin.Name, userID)
	database.DB.Preload("Images").First(&coin, coin.ID)
	c.JSON(http.StatusCreated, coin)
}

func (h *CoinHandler) Update(c *gin.Context) {
	userID := c.GetUint("userId")
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid coin ID"})
		return
	}

	var existing models.Coin
	if err := database.DB.Where("id = ? AND user_id = ?", id, userID).First(&existing).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Coin not found"})
		return
	}

	var updates models.Coin
	if err := c.ShouldBindJSON(&updates); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	updates.ID = existing.ID
	updates.UserID = userID

	if err := database.DB.Model(&existing).Updates(updates).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update coin"})
		return
	}

	database.DB.Preload("Images").First(&existing, existing.ID)
	c.JSON(http.StatusOK, existing)
}

func (h *CoinHandler) Delete(c *gin.Context) {
	userID := c.GetUint("userId")
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid coin ID"})
		return
	}

	result := database.DB.Where("id = ? AND user_id = ?", id, userID).Delete(&models.Coin{})
	if result.RowsAffected == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "Coin not found"})
		return
	}

	// Clean up associated images
	database.DB.Where("coin_id = ?", id).Delete(&models.CoinImage{})

	c.JSON(http.StatusOK, gin.H{"message": "Coin deleted"})
}

func (h *CoinHandler) Stats(c *gin.Context) {
	userID := c.GetUint("userId")

	var totalCoins int64
	var totalWishlist int64
	database.DB.Model(&models.Coin{}).Where("user_id = ? AND is_wishlist = ?", userID, false).Count(&totalCoins)
	database.DB.Model(&models.Coin{}).Where("user_id = ? AND is_wishlist = ?", userID, true).Count(&totalWishlist)

	type categoryCount struct {
		Category string `json:"category"`
		Count    int64  `json:"count"`
	}
	var byCategory []categoryCount
	database.DB.Model(&models.Coin{}).
		Select("category, count(*) as count").
		Where("user_id = ? AND is_wishlist = ?", userID, false).
		Group("category").
		Scan(&byCategory)

	type materialCount struct {
		Material string `json:"material"`
		Count    int64  `json:"count"`
	}
	var byMaterial []materialCount
	database.DB.Model(&models.Coin{}).
		Select("material, count(*) as count").
		Where("user_id = ? AND is_wishlist = ?", userID, false).
		Group("material").
		Scan(&byMaterial)

	type valueSummary struct {
		TotalPurchasePrice float64 `json:"totalPurchasePrice"`
		TotalCurrentValue  float64 `json:"totalCurrentValue"`
		AvgPurchasePrice   float64 `json:"avgPurchasePrice"`
		AvgCurrentValue    float64 `json:"avgCurrentValue"`
	}
	var values valueSummary
	database.DB.Model(&models.Coin{}).
		Select("COALESCE(SUM(purchase_price), 0) as total_purchase_price, COALESCE(SUM(current_value), 0) as total_current_value, COALESCE(AVG(purchase_price), 0) as avg_purchase_price, COALESCE(AVG(current_value), 0) as avg_current_value").
		Where("user_id = ? AND is_wishlist = ?", userID, false).
		Scan(&values)

	c.JSON(http.StatusOK, gin.H{
		"totalCoins":    totalCoins,
		"totalWishlist": totalWishlist,
		"byCategory":    byCategory,
		"byMaterial":    byMaterial,
		"values":        values,
	})
}

// Suggestions returns distinct values for autocomplete fields
func (h *CoinHandler) Suggestions(c *gin.Context) {
	userID := c.GetUint("userId")
	field := c.Query("field")
	q := c.Query("q")

	var column string
	switch field {
	case "name":
		column = "name"
	case "denomination":
		column = "denomination"
	case "purchaseLocation":
		column = "purchase_location"
	default:
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid field"})
		return
	}

	var values []string
	query := database.DB.Model(&models.Coin{}).
		Where("user_id = ? AND "+column+" != ''", userID).
		Distinct(column).
		Order(column)

	if q != "" {
		query = query.Where(column+" LIKE ?", "%"+q+"%")
	}

	query.Limit(20).Pluck(column, &values)
	c.JSON(http.StatusOK, values)
}
