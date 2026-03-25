package repository

import (
	"github.com/briandenicola/ancient-coins-api/models"
	"gorm.io/gorm"
)

// ConversationRepository encapsulates all conversation-related database operations.
type ConversationRepository struct {
	db *gorm.DB
}

// NewConversationRepository creates a new ConversationRepository.
func NewConversationRepository(db *gorm.DB) *ConversationRepository {
	return &ConversationRepository{db: db}
}

// List returns all conversations for a user, newest first.
func (r *ConversationRepository) List(userID uint) ([]models.AgentConversation, error) {
	var conversations []models.AgentConversation
	err := r.db.Where("user_id = ?", userID).
		Order("updated_at DESC").
		Find(&conversations).Error
	return conversations, err
}

// FindByID returns a single conversation owned by the user.
func (r *ConversationRepository) FindByID(id, userID uint) (*models.AgentConversation, error) {
	var conv models.AgentConversation
	err := r.db.Where("id = ? AND user_id = ?", id, userID).First(&conv).Error
	if err != nil {
		return nil, err
	}
	return &conv, nil
}

// Save updates an existing conversation.
func (r *ConversationRepository) Save(conv *models.AgentConversation) error {
	return r.db.Save(conv).Error
}

// Create inserts a new conversation.
func (r *ConversationRepository) Create(conv *models.AgentConversation) error {
	return r.db.Create(conv).Error
}

// Delete removes a conversation owned by the user. Returns rows affected.
func (r *ConversationRepository) Delete(id, userID uint) (int64, error) {
	result := r.db.Where("id = ? AND user_id = ?", id, userID).Delete(&models.AgentConversation{})
	return result.RowsAffected, result.Error
}
