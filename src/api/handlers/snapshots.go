package handlers

import (
	"time"

	"github.com/briandenicola/ancient-coins-api/database"
	"github.com/briandenicola/ancient-coins-api/models"
)

// RecordValueSnapshot captures the current total value, invested amount,
// and coin count for a user. Should be called after coin mutations.
func RecordValueSnapshot(userID uint) {
	type result struct {
		TotalValue    float64
		TotalInvested float64
		CoinCount     int64
	}
	var r result
	database.DB.Model(&models.Coin{}).
		Select("COALESCE(SUM(current_value), 0) as total_value, COALESCE(SUM(purchase_price), 0) as total_invested, COUNT(*) as coin_count").
		Where("user_id = ? AND is_wishlist = ?", userID, false).
		Scan(&r)

	snapshot := models.ValueSnapshot{
		UserID:        userID,
		TotalValue:    r.TotalValue,
		TotalInvested: r.TotalInvested,
		CoinCount:     r.CoinCount,
		RecordedAt:    time.Now(),
	}
	database.DB.Create(&snapshot)
}
