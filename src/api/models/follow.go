package models

import "time"

type Follow struct {
	ID          uint      `gorm:"primaryKey" json:"id"`
	FollowerID  uint      `gorm:"not null;uniqueIndex:idx_follow" json:"followerId"`
	Follower    User      `gorm:"foreignKey:FollowerID" json:"-"`
	FollowingID uint      `gorm:"not null;uniqueIndex:idx_follow" json:"followingId"`
	Following   User      `gorm:"foreignKey:FollowingID" json:"-"`
	CreatedAt   time.Time `json:"createdAt"`
}
