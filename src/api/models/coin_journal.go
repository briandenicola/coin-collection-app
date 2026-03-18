package models

import "time"

type CoinJournal struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	CoinID    uint      `gorm:"not null;index" json:"coinId"`
	UserID    uint      `gorm:"not null" json:"userId"`
	Entry     string    `gorm:"type:text;not null" json:"entry"`
	CreatedAt time.Time `json:"createdAt"`
}
