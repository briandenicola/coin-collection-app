package models

import "time"

type AgentConversation struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	UserID    uint      `gorm:"not null;index" json:"userId"`
	User      User      `gorm:"foreignKey:UserID" json:"-"`
	Title     string    `gorm:"not null" json:"title"`
	Messages  string    `gorm:"type:text;not null" json:"messages"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
}
