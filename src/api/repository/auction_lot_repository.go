package repository

import (
	"github.com/briandenicola/ancient-coins-api/models"
	"gorm.io/gorm"
)

// AuctionLotRepository encapsulates all auction-lot-related database operations.
type AuctionLotRepository struct {
	db *gorm.DB
}

// NewAuctionLotRepository creates a new AuctionLotRepository.
func NewAuctionLotRepository(db *gorm.DB) *AuctionLotRepository {
	return &AuctionLotRepository{db: db}
}

// AuctionLotListFilters holds filtering/sorting options for listing auction lots.
type AuctionLotListFilters struct {
	Status    string
	Search    string
	SortField string
	SortOrder string
	Page      int
	Limit     int
}

// List returns a paginated list of auction lots for the given user.
func (r *AuctionLotRepository) List(userID uint, filters AuctionLotListFilters) ([]models.AuctionLot, int64, error) {
	var lots []models.AuctionLot
	var total int64

	query := r.db.Model(&models.AuctionLot{}).Scopes(OwnedBy(userID))

	if filters.Status != "" {
		query = query.Where("status = ?", filters.Status)
	}
	if filters.Search != "" {
		like := "%" + filters.Search + "%"
		query = query.Where("title LIKE ? OR description LIKE ? OR auction_house LIKE ?", like, like, like)
	}

	query.Count(&total)

	sortField := "updated_at"
	if filters.SortField != "" {
		allowed := map[string]bool{
			"created_at": true, "updated_at": true, "sale_date": true,
			"estimate": true, "current_bid": true, "lot_number": true,
		}
		if allowed[filters.SortField] {
			sortField = filters.SortField
		}
	}
	sortOrder := "desc"
	if filters.SortOrder == "asc" {
		sortOrder = "asc"
	}

	limit := 50
	if filters.Limit > 0 && filters.Limit <= 100 {
		limit = filters.Limit
	}
	page := 1
	if filters.Page > 0 {
		page = filters.Page
	}
	offset := (page - 1) * limit

	err := query.Order(sortField + " " + sortOrder).Limit(limit).Offset(offset).Find(&lots).Error
	return lots, total, err
}

// GetByID returns a single auction lot owned by the given user.
func (r *AuctionLotRepository) GetByID(id, userID uint) (*models.AuctionLot, error) {
	var lot models.AuctionLot
	err := r.db.Scopes(OwnedByID(id, userID)).First(&lot).Error
	if err != nil {
		return nil, err
	}
	return &lot, nil
}

// GetByURL finds an auction lot by its NumisBids URL for the given user.
func (r *AuctionLotRepository) GetByURL(url string, userID uint) (*models.AuctionLot, error) {
	var lot models.AuctionLot
	err := r.db.Where("numis_bids_url = ? AND user_id = ?", url, userID).First(&lot).Error
	if err != nil {
		return nil, err
	}
	return &lot, nil
}

// Create inserts a new auction lot.
func (r *AuctionLotRepository) Create(lot *models.AuctionLot) error {
	return r.db.Create(lot).Error
}

// Update saves changes to an existing auction lot.
func (r *AuctionLotRepository) Update(lot *models.AuctionLot, updates *models.AuctionLot) error {
	return r.db.Model(lot).Updates(updates).Error
}

// UpdateFields updates specific fields on an auction lot.
func (r *AuctionLotRepository) UpdateFields(lot *models.AuctionLot, fields map[string]interface{}) error {
	return r.db.Model(lot).Updates(fields).Error
}

// Delete removes an auction lot.
func (r *AuctionLotRepository) Delete(id, userID uint) (int64, error) {
	result := r.db.Where("id = ? AND user_id = ?", id, userID).Delete(&models.AuctionLot{})
	return result.RowsAffected, result.Error
}

// Upsert creates or updates an auction lot by its NumisBids URL for the given user.
func (r *AuctionLotRepository) Upsert(lot *models.AuctionLot) error {
	existing, err := r.GetByURL(lot.NumisBidsURL, lot.UserID)
	if err != nil {
		// Not found — create
		return r.Create(lot)
	}
	// Update fields that may have changed
	updates := map[string]interface{}{
		"current_bid":   lot.CurrentBid,
		"estimate":      lot.Estimate,
		"title":         lot.Title,
		"description":   lot.Description,
		"image_url":     lot.ImageURL,
		"auction_house": lot.AuctionHouse,
		"sale_name":     lot.SaleName,
		"currency":      lot.Currency,
		"lot_number":    lot.LotNumber,
	}
	return r.UpdateFields(existing, updates)
}
