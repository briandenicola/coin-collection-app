package repository

import (
	"github.com/briandenicola/ancient-coins-api/models"
	"gorm.io/gorm"
)

// WebAuthnRepository encapsulates database operations for WebAuthn credentials.
type WebAuthnRepository struct {
	db *gorm.DB
}

// NewWebAuthnRepository creates a new WebAuthnRepository.
func NewWebAuthnRepository(db *gorm.DB) *WebAuthnRepository {
	return &WebAuthnRepository{db: db}
}

// FindUserByID finds a user by primary key.
func (r *WebAuthnRepository) FindUserByID(userID uint) (*models.User, error) {
	var user models.User
	if err := r.db.First(&user, userID).Error; err != nil {
		return nil, err
	}
	return &user, nil
}

// FindUserByUsername finds a user by username.
func (r *WebAuthnRepository) FindUserByUsername(username string) (*models.User, error) {
	var user models.User
	if err := r.db.Where("username = ?", username).First(&user).Error; err != nil {
		return nil, err
	}
	return &user, nil
}

// LoadCredentials returns all WebAuthn credentials for a user.
func (r *WebAuthnRepository) LoadCredentials(userID uint) ([]models.WebAuthnCredential, error) {
	var creds []models.WebAuthnCredential
	if err := r.db.Where("user_id = ?", userID).Find(&creds).Error; err != nil {
		return nil, err
	}
	return creds, nil
}

// CreateCredential persists a new WebAuthn credential.
func (r *WebAuthnRepository) CreateCredential(cred *models.WebAuthnCredential) error {
	return r.db.Create(cred).Error
}

// UpdateSignCount updates the sign count for a credential identified by credential ID and user ID.
func (r *WebAuthnRepository) UpdateSignCount(credentialID string, userID uint, signCount uint32) error {
	return r.db.Model(&models.WebAuthnCredential{}).
		Where("credential_id = ? AND user_id = ?", credentialID, userID).
		Update("sign_count", signCount).Error
}

// DeleteCredential deletes a credential by ID and user ID. Returns rows affected.
func (r *WebAuthnRepository) DeleteCredential(credID string, userID uint) (int64, error) {
	result := r.db.Where("id = ? AND user_id = ?", credID, userID).Delete(&models.WebAuthnCredential{})
	return result.RowsAffected, result.Error
}

// CountCredentials returns the number of credentials for a user.
func (r *WebAuthnRepository) CountCredentials(userID uint) (int64, error) {
	var count int64
	if err := r.db.Model(&models.WebAuthnCredential{}).Where("user_id = ?", userID).Count(&count).Error; err != nil {
		return 0, err
	}
	return count, nil
}
