package models

import "time"

type CoinValueHistory struct {
	ID         uint      `gorm:"primaryKey" json:"id"`
	CoinID     uint      `gorm:"not null;index" json:"coinId"`
	UserID     uint      `gorm:"not null;index" json:"userId"`
	Value      float64   `gorm:"not null" json:"value"`
	Confidence string    `gorm:"type:varchar(20);not null" json:"confidence"`
	RecordedAt time.Time `gorm:"not null;index" json:"recordedAt"`
}
