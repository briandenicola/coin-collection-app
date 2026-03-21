package models

import "time"

type CoinComment struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	CoinID    uint      `gorm:"not null;index" json:"coinId"`
	Coin      Coin      `gorm:"foreignKey:CoinID" json:"-"`
	UserID    uint      `gorm:"not null;index" json:"userId"`
	User      User      `gorm:"foreignKey:UserID" json:"-"`
	Comment   string    `gorm:"type:text;not null" json:"comment"`
	Rating    int       `gorm:"default:0" json:"rating"`
	CreatedAt time.Time `json:"createdAt"`
}
