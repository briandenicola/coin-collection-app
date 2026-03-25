package services

import (
	"fmt"
	"time"

	"github.com/briandenicola/ancient-coins-api/models"
	"github.com/briandenicola/ancient-coins-api/repository"
)

// CoinService handles coin business logic and orchestrates repository calls.
type CoinService struct {
	repo *repository.CoinRepository
}

// NewCoinService creates a new CoinService.
func NewCoinService(repo *repository.CoinRepository) *CoinService {
	return &CoinService{repo: repo}
}

// CreateCoin creates a coin and records a value snapshot.
func (s *CoinService) CreateCoin(coin *models.Coin) error {
	if err := s.repo.Create(coin); err != nil {
		return err
	}
	s.repo.RecordValueSnapshot(coin.UserID)
	return nil
}

// UpdateCoin applies updates to an existing coin. If the current value changed
// and the source is not "estimate", it records a value history entry and a
// journal entry. A value snapshot is always recorded afterward.
func (s *CoinService) UpdateCoin(existing *models.Coin, updates *models.Coin, userID uint, source string) error {
	oldValue := existing.CurrentValue

	if err := s.repo.Update(existing, updates); err != nil {
		return err
	}

	if updates.CurrentValue != nil {
		newVal := *updates.CurrentValue
		oldVal := 0.0
		if oldValue != nil {
			oldVal = *oldValue
		}
		if newVal != oldVal && source != "estimate" {
			s.repo.RecordValueHistory(&models.CoinValueHistory{
				CoinID:     existing.ID,
				UserID:     userID,
				Value:      newVal,
				Confidence: "manual",
				RecordedAt: time.Now(),
			})
			s.repo.CreateJournalEntry(&models.CoinJournal{
				CoinID: existing.ID,
				UserID: userID,
				Entry:  fmt.Sprintf("Current value updated manually: $%.2f", newVal),
			})
		}
	}

	s.repo.RecordValueSnapshot(userID)
	return nil
}

// DeleteCoin deletes a coin and records a value snapshot if rows were affected.
// Returns the number of rows affected.
func (s *CoinService) DeleteCoin(id, userID uint) (int64, error) {
	rows, err := s.repo.Delete(id, userID)
	if err != nil {
		return 0, err
	}
	if rows > 0 {
		s.repo.RecordValueSnapshot(userID)
	}
	return rows, nil
}

// PurchaseCoin marks a wishlist coin as purchased and records a value snapshot.
func (s *CoinService) PurchaseCoin(coin *models.Coin, userID uint) error {
	if err := s.repo.UpdateField(coin, "is_wishlist", false); err != nil {
		return err
	}
	s.repo.RecordValueSnapshot(userID)
	return nil
}

// SellCoin applies sale updates to a coin and records a value snapshot.
func (s *CoinService) SellCoin(coin *models.Coin, updates map[string]interface{}, userID uint) error {
	if err := s.repo.UpdateFields(coin, updates); err != nil {
		return err
	}
	s.repo.RecordValueSnapshot(userID)
	return nil
}
