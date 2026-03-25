package repository

import (
	"github.com/briandenicola/ancient-coins-api/models"
	"gorm.io/gorm"
)

// AdminRepository encapsulates database operations for the admin handler.
type AdminRepository struct {
	db *gorm.DB
}

// NewAdminRepository creates a new AdminRepository.
func NewAdminRepository(db *gorm.DB) *AdminRepository {
	return &AdminRepository{db: db}
}

// ListUsers returns all users.
func (r *AdminRepository) ListUsers() ([]models.User, error) {
	var users []models.User
	if err := r.db.Find(&users).Error; err != nil {
		return nil, err
	}
	return users, nil
}

// DeleteUserCascade deletes a user and all associated data in a transaction.
func (r *AdminRepository) DeleteUserCascade(userID uint) (int64, error) {
	var rowsAffected int64
	err := r.db.Transaction(func(tx *gorm.DB) error {
		var coinIDs []uint
		tx.Model(&models.Coin{}).Where("user_id = ?", userID).Pluck("id", &coinIDs)
		if len(coinIDs) > 0 {
			tx.Where("coin_id IN ?", coinIDs).Delete(&models.CoinImage{})
			tx.Where("coin_id IN ?", coinIDs).Delete(&models.CoinJournal{})
		}
		tx.Where("user_id = ?", userID).Delete(&models.Coin{})
		tx.Where("user_id = ?", userID).Delete(&models.AgentConversation{})
		tx.Where("user_id = ?", userID).Delete(&models.ValueSnapshot{})
		tx.Where("user_id = ?", userID).Delete(&models.ApiKey{})
		tx.Where("user_id = ?", userID).Delete(&models.RefreshToken{})
		tx.Where("user_id = ?", userID).Delete(&models.WebAuthnCredential{})

		result := tx.Delete(&models.User{}, userID)
		rowsAffected = result.RowsAffected
		return result.Error
	})
	return rowsAffected, err
}

// ResetPassword updates a user's password hash. Returns rows affected.
func (r *AdminRepository) ResetPassword(userID uint, passwordHash string) (int64, error) {
	result := r.db.Model(&models.User{}).Where("id = ?", userID).Update("password_hash", passwordHash)
	return result.RowsAffected, result.Error
}

// ExportAllCoins returns all coins with their images preloaded.
func (r *AdminRepository) ExportAllCoins() ([]models.Coin, error) {
	var coins []models.Coin
	if err := r.db.Preload("Images").Find(&coins).Error; err != nil {
		return nil, err
	}
	return coins, nil
}

// ImportCoin creates a single coin record.
func (r *AdminRepository) ImportCoin(coin *models.Coin) error {
	return r.db.Create(coin).Error
}
