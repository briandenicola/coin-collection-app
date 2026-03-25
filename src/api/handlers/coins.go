package handlers

import (
	"net/http"
	"strconv"
	"time"

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

// List returns a paginated list of coins for the authenticated user.
//
//	@Summary		List coins
//	@Description	Returns a paginated, filterable list of coins belonging to the authenticated user.
//	@Tags			Coins
//	@Produce		json
//	@Param			category	query		string	false	"Filter by category (Roman, Greek, Byzantine, Modern, Other)"
//	@Param			search		query		string	false	"Search across name, denomination, ruler, era, mint, inscriptions, notes"
//	@Param			wishlist	query		string	false	"Filter by wishlist status"	Enums(true, false)
//	@Param			sold		query		string	false	"Filter by sold status"	Enums(true, false)
//	@Param			page		query		int		false	"Page number"	default(1)
//	@Param			limit		query		int		false	"Items per page (max 100)"	default(50)
//	@Param			sort		query		string	false	"Sort field"	Enums(created_at, updated_at, current_value)	default(updated_at)
//	@Param			order		query		string	false	"Sort order"	Enums(asc, desc)	default(desc)
//	@Success		200			{object}	CoinListResponse
//	@Failure		401			{object}	ErrorResponse
//	@Failure		500			{object}	ErrorResponse
//	@Security		BearerAuth
//	@Router			/coins [get]
func (h *CoinHandler) List(c *gin.Context) {
	userID := c.GetUint("userId")
	category := c.Query("category")
	search := c.Query("search")
	wishlist := c.Query("wishlist")
	sold := c.Query("sold")
	sortField := c.DefaultQuery("sort", "updated_at")
	sortOrder := c.DefaultQuery("order", "desc")
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

	if sold == "true" {
		query = query.Where("is_sold = ?", true)
	} else if sold == "false" {
		query = query.Where("is_sold = ?", false)
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

	allowedSortFields := map[string]string{
		"created_at":    "created_at",
		"updated_at":    "updated_at",
		"current_value": "current_value",
	}
	column, ok := allowedSortFields[sortField]
	if !ok {
		column = "updated_at"
	}
	if sortOrder != "asc" && sortOrder != "desc" {
		sortOrder = "desc"
	}
	orderClause := column + " " + sortOrder

	var coins []models.Coin
	if err := query.Order(orderClause).Offset(offset).Limit(limit).Find(&coins).Error; err != nil {
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

// Get returns a single coin by ID for the authenticated user.
//
//	@Summary		Get a coin
//	@Description	Returns a single coin with its images, owned by the authenticated user.
//	@Tags			Coins
//	@Produce		json
//	@Param			id	path		int	true	"Coin ID"
//	@Success		200	{object}	models.Coin
//	@Failure		400	{object}	ErrorResponse
//	@Failure		401	{object}	ErrorResponse
//	@Failure		404	{object}	ErrorResponse
//	@Failure		500	{object}	ErrorResponse
//	@Security		BearerAuth
//	@Router			/coins/{id} [get]
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

// Create adds a new coin for the authenticated user.
//
//	@Summary		Create a coin
//	@Description	Creates a new coin record for the authenticated user.
//	@Tags			Coins
//	@Accept			json
//	@Produce		json
//	@Param			body	body		models.Coin	true	"Coin data"
//	@Success		201		{object}	models.Coin
//	@Failure		400		{object}	ErrorResponse
//	@Failure		401		{object}	ErrorResponse
//	@Failure		500		{object}	ErrorResponse
//	@Security		BearerAuth
//	@Router			/coins [post]
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
	RecordValueSnapshot(userID)
	c.JSON(http.StatusCreated, coin)
}

// Update modifies an existing coin owned by the authenticated user.
//
//	@Summary		Update a coin
//	@Description	Updates an existing coin record. Only the coin owner can update it.
//	@Tags			Coins
//	@Accept			json
//	@Produce		json
//	@Param			id		path		int			true	"Coin ID"
//	@Param			body	body		models.Coin	true	"Updated coin data"
//	@Success		200		{object}	models.Coin
//	@Failure		400		{object}	ErrorResponse
//	@Failure		401		{object}	ErrorResponse
//	@Failure		404		{object}	ErrorResponse
//	@Failure		500		{object}	ErrorResponse
//	@Security		BearerAuth
//	@Router			/coins/{id} [put]
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

	oldValue := existing.CurrentValue

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

	// Track manual value changes
	if updates.CurrentValue != nil {
		newVal := *updates.CurrentValue
		oldVal := 0.0
		if oldValue != nil {
			oldVal = *oldValue
		}
		if newVal != oldVal {
			database.DB.Create(&models.CoinValueHistory{
				CoinID:     existing.ID,
				UserID:     userID,
				Value:      newVal,
				Confidence: "manual",
				RecordedAt: time.Now(),
			})
		}
	}

	database.DB.Preload("Images").First(&existing, existing.ID)
	RecordValueSnapshot(userID)
	c.JSON(http.StatusOK, existing)
}

// Purchase marks a wishlist coin as purchased (moves it to the collection).
//
//	@Summary		Mark coin as purchased
//	@Description	Sets isWishlist to false, moving the coin from wishlist to collection. Only the coin owner can do this.
//	@Tags			Coins
//	@Produce		json
//	@Param			id	path		int	true	"Coin ID"
//	@Success		200	{object}	models.Coin
//	@Failure		400	{object}	ErrorResponse
//	@Failure		401	{object}	ErrorResponse
//	@Failure		404	{object}	ErrorResponse
//	@Failure		500	{object}	ErrorResponse
//	@Security		BearerAuth
//	@Router			/coins/{id}/purchase [post]
func (h *CoinHandler) Purchase(c *gin.Context) {
	userID := c.GetUint("userId")
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid coin ID"})
		return
	}

	var coin models.Coin
	if err := database.DB.Where("id = ? AND user_id = ?", id, userID).First(&coin).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Coin not found"})
		return
	}

	if err := database.DB.Model(&coin).Update("is_wishlist", false).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to mark as purchased"})
		return
	}

	database.DB.Preload("Images").First(&coin, coin.ID)
	RecordValueSnapshot(userID)
	c.JSON(http.StatusOK, coin)
}

// Sell marks a collection coin as sold with a sale price.
//
//	@Summary		Mark coin as sold
//	@Description	Sets isSold to true with a sold price and date, moving the coin from collection to sold gallery.
//	@Tags			Coins
//	@Accept			json
//	@Produce		json
//	@Param			id		path		int						true	"Coin ID"
//	@Param			body	body		object{soldPrice float64}	true	"Sale details"
//	@Success		200		{object}	models.Coin
//	@Failure		400		{object}	ErrorResponse
//	@Failure		401		{object}	ErrorResponse
//	@Failure		404		{object}	ErrorResponse
//	@Failure		500		{object}	ErrorResponse
//	@Security		BearerAuth
//	@Router			/coins/{id}/sell [post]
func (h *CoinHandler) Sell(c *gin.Context) {
	userID := c.GetUint("userId")
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid coin ID"})
		return
	}

	var body struct {
		SoldPrice *float64 `json:"soldPrice"`
		SoldTo    string   `json:"soldTo"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	var coin models.Coin
	if err := database.DB.Where("id = ? AND user_id = ?", id, userID).First(&coin).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Coin not found"})
		return
	}

	now := time.Now()
	updates := map[string]interface{}{
		"is_sold":    true,
		"sold_price": body.SoldPrice,
		"sold_date":  &now,
		"sold_to":    body.SoldTo,
	}
	if err := database.DB.Model(&coin).Updates(updates).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to mark as sold"})
		return
	}

	database.DB.Preload("Images").First(&coin, coin.ID)
	RecordValueSnapshot(userID)
	c.JSON(http.StatusOK, coin)
}

// Delete removes a coin and its associated images.
//
//	@Summary		Delete a coin
//	@Description	Deletes a coin and all associated image records. Only the coin owner can delete it.
//	@Tags			Coins
//	@Produce		json
//	@Param			id	path		int	true	"Coin ID"
//	@Success		200	{object}	CoinDeletedResponse
//	@Failure		400	{object}	ErrorResponse
//	@Failure		401	{object}	ErrorResponse
//	@Failure		404	{object}	ErrorResponse
//	@Security		BearerAuth
//	@Router			/coins/{id} [delete]
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

	RecordValueSnapshot(userID)
	c.JSON(http.StatusOK, gin.H{"message": "Coin deleted"})
}

// Stats returns aggregate statistics for the authenticated user's collection.
//
//	@Summary		Get collection statistics
//	@Description	Returns counts by category/material, total/average values, and wishlist count.
//	@Tags			Coins
//	@Produce		json
//	@Success		200	{object}	StatsResponse
//	@Failure		401	{object}	ErrorResponse
//	@Security		BearerAuth
//	@Router			/stats [get]
func (h *CoinHandler) Stats(c *gin.Context) {
	userID := c.GetUint("userId")

	var totalCoins int64
	var totalWishlist int64
	var totalSold int64
	database.DB.Model(&models.Coin{}).Where("user_id = ? AND is_wishlist = ? AND is_sold = ?", userID, false, false).Count(&totalCoins)
	database.DB.Model(&models.Coin{}).Where("user_id = ? AND is_wishlist = ?", userID, true).Count(&totalWishlist)
	database.DB.Model(&models.Coin{}).Where("user_id = ? AND is_sold = ?", userID, true).Count(&totalSold)

	type categoryCount struct {
		Category string `json:"category"`
		Count    int64  `json:"count"`
	}
	var byCategory []categoryCount
	database.DB.Model(&models.Coin{}).
		Select("category, count(*) as count").
		Where("user_id = ? AND is_wishlist = ? AND is_sold = ?", userID, false, false).
		Group("category").
		Scan(&byCategory)

	type materialCount struct {
		Material string `json:"material"`
		Count    int64  `json:"count"`
	}
	var byMaterial []materialCount
	database.DB.Model(&models.Coin{}).
		Select("material, count(*) as count").
		Where("user_id = ? AND is_wishlist = ? AND is_sold = ?", userID, false, false).
		Group("material").
		Scan(&byMaterial)

	type gradeCount struct {
		Grade string `json:"grade"`
		Count int64  `json:"count"`
	}
	var byGrade []gradeCount
	database.DB.Model(&models.Coin{}).
		Select("grade, count(*) as count").
		Where("user_id = ? AND is_wishlist = ? AND is_sold = ? AND grade != ''", userID, false, false).
		Group("grade").
		Order("count DESC").
		Scan(&byGrade)

	type eraCount struct {
		Era   string `json:"era"`
		Count int64  `json:"count"`
	}
	var byEra []eraCount
	database.DB.Model(&models.Coin{}).
		Select("era, count(*) as count").
		Where("user_id = ? AND is_wishlist = ? AND is_sold = ? AND era != ''", userID, false, false).
		Group("era").
		Order("count DESC").
		Scan(&byEra)

	type rulerCount struct {
		Ruler string `json:"ruler"`
		Count int64  `json:"count"`
	}
	var byRuler []rulerCount
	database.DB.Model(&models.Coin{}).
		Select("ruler, count(*) as count").
		Where("user_id = ? AND is_wishlist = ? AND is_sold = ? AND ruler != ''", userID, false, false).
		Group("ruler").
		Order("count DESC").
		Limit(10).
		Scan(&byRuler)

	type priceRange struct {
		Range string `json:"range"`
		Count int64  `json:"count"`
	}
	var byPriceRange []priceRange
	database.DB.Model(&models.Coin{}).
		Select(`CASE
			WHEN purchase_price < 50 THEN 'Under $50'
			WHEN purchase_price >= 50 AND purchase_price < 200 THEN '$50 - $200'
			WHEN purchase_price >= 200 AND purchase_price < 500 THEN '$200 - $500'
			WHEN purchase_price >= 500 AND purchase_price < 1000 THEN '$500 - $1K'
			ELSE '$1K+'
		END as ` + "`range`" + `, count(*) as count`).
		Where("user_id = ? AND is_wishlist = ? AND is_sold = ? AND purchase_price IS NOT NULL", userID, false, false).
		Group("`range`").
		Scan(&byPriceRange)

	type valueSummary struct {
		TotalPurchasePrice float64 `json:"totalPurchasePrice"`
		TotalCurrentValue  float64 `json:"totalCurrentValue"`
		AvgPurchasePrice   float64 `json:"avgPurchasePrice"`
		AvgCurrentValue    float64 `json:"avgCurrentValue"`
	}
	var values valueSummary
	database.DB.Model(&models.Coin{}).
		Select("COALESCE(SUM(purchase_price), 0) as total_purchase_price, COALESCE(SUM(current_value), 0) as total_current_value, COALESCE(AVG(purchase_price), 0) as avg_purchase_price, COALESCE(AVG(current_value), 0) as avg_current_value").
		Where("user_id = ? AND is_wishlist = ? AND is_sold = ?", userID, false, false).
		Scan(&values)

	// Sold coins value summary
	type soldSummary struct {
		TotalSoldPrice    float64 `json:"totalSoldPrice"`
		TotalPurchaseCost float64 `json:"totalPurchaseCost"`
	}
	var soldValues soldSummary
	database.DB.Model(&models.Coin{}).
		Select("COALESCE(SUM(sold_price), 0) as total_sold_price, COALESCE(SUM(purchase_price), 0) as total_purchase_cost").
		Where("user_id = ? AND is_sold = ?", userID, true).
		Scan(&soldValues)

	c.JSON(http.StatusOK, gin.H{
		"totalCoins":    totalCoins,
		"totalWishlist": totalWishlist,
		"totalSold":     totalSold,
		"byCategory":    byCategory,
		"byMaterial":    byMaterial,
		"byGrade":       byGrade,
		"byEra":         byEra,
		"byRuler":       byRuler,
		"byPriceRange":  byPriceRange,
		"values":        values,
		"soldValues":    soldValues,
	})
}

// Suggestions returns distinct values for autocomplete fields.
//
//	@Summary		Get autocomplete suggestions
//	@Description	Returns distinct values for the specified field, optionally filtered by a search query. Limited to 20 results.
//	@Tags			Coins
//	@Produce		json
//	@Param			field	query		string	true	"Field to get suggestions for"	Enums(name, denomination, purchaseLocation)
//	@Param			q		query		string	false	"Search filter"
//	@Success		200		{array}		string
//	@Failure		400		{object}	ErrorResponse
//	@Failure		401		{object}	ErrorResponse
//	@Security		BearerAuth
//	@Router			/suggestions [get]
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
	case "ruler":
		column = "ruler"
	case "era":
		column = "era"
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

// ValueHistory returns historical value snapshots for the authenticated user.
//
//	@Summary		Get value history
//	@Description	Returns value snapshots over time for charting collection value trends.
//	@Tags			Coins
//	@Produce		json
//	@Success		200	{array}		models.ValueSnapshot
//	@Failure		401	{object}	ErrorResponse
//	@Security		BearerAuth
//	@Router			/value-history [get]
func (h *CoinHandler) ValueHistory(c *gin.Context) {
	userID := c.GetUint("userId")

	var snapshots []models.ValueSnapshot
	database.DB.Where("user_id = ?", userID).
		Order("recorded_at ASC").
		Find(&snapshots)

	c.JSON(http.StatusOK, snapshots)
}

// CoinValueHistory returns value history entries for a specific coin.
//
//	@Summary		Get coin value history
//	@Description	Returns value tracking entries for a coin, ordered oldest first.
//	@Tags			Coins
//	@Produce		json
//	@Param			id	path		int	true	"Coin ID"
//	@Success		200	{array}		models.CoinValueHistory
//	@Failure		401	{object}	ErrorResponse
//	@Failure		404	{object}	ErrorResponse
//	@Security		BearerAuth
//	@Router			/coins/{id}/value-history [get]
func (h *CoinHandler) CoinValueHistory(c *gin.Context) {
	userID := c.GetUint("userId")
	coinID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid coin ID"})
		return
	}

	var count int64
	database.DB.Model(&models.Coin{}).Where("id = ? AND user_id = ?", coinID, userID).Count(&count)
	if count == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "Coin not found"})
		return
	}

	var entries []models.CoinValueHistory
	database.DB.Where("coin_id = ? AND user_id = ?", coinID, userID).
		Order("recorded_at ASC").
		Find(&entries)

	c.JSON(http.StatusOK, entries)
}
