package models

import "time"

type ValueSnapshot struct {
	ID            uint      `gorm:"primaryKey" json:"id"`
	UserID        uint      `gorm:"not null;index" json:"userId"`
	TotalValue    float64   `json:"totalValue"`
	TotalInvested float64   `json:"totalInvested"`
	CoinCount     int64     `json:"coinCount"`
	RecordedAt    time.Time `gorm:"not null;index" json:"recordedAt"`
}
