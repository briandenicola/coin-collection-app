package repository

import (
	"github.com/briandenicola/ancient-coins-api/models"
	"gorm.io/gorm"
)

// JournalRepository encapsulates all journal-related database operations.
type JournalRepository struct {
	db *gorm.DB
}

// NewJournalRepository creates a new JournalRepository.
func NewJournalRepository(db *gorm.DB) *JournalRepository {
	return &JournalRepository{db: db}
}

// CoinExists checks if a coin belongs to the given user.
func (r *JournalRepository) CoinExists(coinID, userID uint) bool {
	var count int64
	r.db.Model(&models.Coin{}).Where("id = ? AND user_id = ?", coinID, userID).Count(&count)
	return count > 0
}

// GetEntries returns all journal entries for a coin owned by the user, newest first.
func (r *JournalRepository) GetEntries(coinID, userID uint) ([]models.CoinJournal, error) {
	var entries []models.CoinJournal
	err := r.db.Where("coin_id = ? AND user_id = ?", coinID, userID).
		Order("created_at DESC").
		Find(&entries).Error
	return entries, err
}

// CreateEntry inserts a new journal entry.
func (r *JournalRepository) CreateEntry(entry *models.CoinJournal) error {
	return r.db.Create(entry).Error
}

// DeleteEntry removes a journal entry owned by the user. Returns rows affected.
func (r *JournalRepository) DeleteEntry(entryID, userID uint) (int64, error) {
	result := r.db.Where("id = ? AND user_id = ?", entryID, userID).Delete(&models.CoinJournal{})
	return result.RowsAffected, result.Error
}
