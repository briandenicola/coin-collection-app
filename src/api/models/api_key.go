package models

import "time"

type ApiKey struct {
	ID        uint       `gorm:"primaryKey" json:"id"`
	UserID    uint       `gorm:"not null;index" json:"userId"`
	User      User       `gorm:"foreignKey:UserID" json:"-"`
	KeyHash   string     `gorm:"not null;uniqueIndex" json:"-"`
	KeyPrefix string     `gorm:"not null" json:"keyPrefix"`
	Name      string     `gorm:"not null" json:"name"`
	CreatedAt time.Time  `json:"createdAt"`
	LastUsedAt *time.Time `json:"lastUsedAt"`
	RevokedAt *time.Time `json:"revokedAt"`
}
