package repository

import "gorm.io/gorm"

// OwnedBy scopes a query to records belonging to the given user.
func OwnedBy(userID uint) func(db *gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		return db.Where("user_id = ?", userID)
	}
}

// ByID scopes a query to a specific record by primary key.
func ByID(id uint) func(db *gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		return db.Where("id = ?", id)
	}
}

// OwnedByID scopes a query to a specific record owned by the given user.
func OwnedByID(id uint, userID uint) func(db *gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		return db.Where("id = ? AND user_id = ?", id, userID)
	}
}

// ActiveCollection scopes to coins that are not wishlist and not sold.
func ActiveCollection(userID uint) func(db *gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		return db.Where("user_id = ? AND is_wishlist = ? AND is_sold = ?", userID, false, false)
	}
}

// PublicCoins scopes to coins that are visible to other users.
func PublicCoins(userID uint) func(db *gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		return db.Where("user_id = ? AND is_wishlist = false AND is_sold = false AND is_private = false", userID)
	}
}

// ByCoinID scopes a query to records associated with a specific coin.
func ByCoinID(coinID uint) func(db *gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		return db.Where("coin_id = ?", coinID)
	}
}
