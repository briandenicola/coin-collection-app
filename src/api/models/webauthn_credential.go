package models

import "time"

type WebAuthnCredential struct {
	ID              uint      `gorm:"primaryKey" json:"id"`
	UserID          uint      `gorm:"not null;index" json:"userId"`
	CredentialID    string    `gorm:"not null;uniqueIndex" json:"credentialId"`
	PublicKey       []byte    `gorm:"not null" json:"-"`
	AttestationType string    `json:"attestationType"`
	SignCount       uint32    `gorm:"default:0" json:"signCount"`
	Name            string    `json:"name"`
	CreatedAt       time.Time `json:"createdAt"`
}
