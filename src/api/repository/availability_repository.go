package repository

import (
	"github.com/briandenicola/ancient-coins-api/models"
	"gorm.io/gorm"
)

// AvailabilityRepository encapsulates all availability-check related DB operations.
type AvailabilityRepository struct {
	db *gorm.DB
}

// NewAvailabilityRepository creates a new AvailabilityRepository.
func NewAvailabilityRepository(db *gorm.DB) *AvailabilityRepository {
	return &AvailabilityRepository{db: db}
}

// CreateRun inserts a new availability run.
func (r *AvailabilityRepository) CreateRun(run *models.AvailabilityRun) error {
	return r.db.Create(run).Error
}

// CompleteRun updates a run's stats and completion timestamp.
func (r *AvailabilityRepository) CompleteRun(run *models.AvailabilityRun) error {
	return r.db.Model(run).Updates(map[string]interface{}{
		"coins_checked": run.CoinsChecked,
		"available":     run.Available,
		"unavailable":   run.Unavailable,
		"unknown":       run.Unknown,
		"errors":        run.Errors,
		"duration_ms":   run.DurationMs,
		"completed_at":  run.CompletedAt,
	}).Error
}

// CreateResult inserts a single availability check result.
func (r *AvailabilityRepository) CreateResult(result *models.AvailabilityResult) error {
	return r.db.Create(result).Error
}

// UpdateResult updates an existing availability check result (used by agent escalation).
func (r *AvailabilityRepository) UpdateResult(result *models.AvailabilityResult) error {
	return r.db.Model(result).Updates(map[string]interface{}{
		"status":     result.Status,
		"reason":     result.Reason,
		"agent_used": result.AgentUsed,
	}).Error
}

// ListRuns returns paginated availability runs, newest first.
func (r *AvailabilityRepository) ListRuns(page, limit int) ([]models.AvailabilityRun, int64, error) {
	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 20
	}

	var total int64
	if err := r.db.Model(&models.AvailabilityRun{}).Count(&total).Error; err != nil {
		return nil, 0, err
	}

	var runs []models.AvailabilityRun
	offset := (page - 1) * limit
	err := r.db.Preload("User", func(db *gorm.DB) *gorm.DB {
		return db.Select("id, username")
	}).Order("started_at DESC").Offset(offset).Limit(limit).Find(&runs).Error
	return runs, total, err
}

// GetRunWithResults returns a single run with all its per-coin results.
func (r *AvailabilityRepository) GetRunWithResults(runID uint) (*models.AvailabilityRun, error) {
	var run models.AvailabilityRun
	err := r.db.Preload("Results").Preload("User", func(db *gorm.DB) *gorm.DB {
		return db.Select("id, username")
	}).First(&run, runID).Error
	if err != nil {
		return nil, err
	}
	return &run, nil
}
