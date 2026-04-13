package services

import (
	"strconv"
	"time"

	"github.com/briandenicola/ancient-coins-api/repository"
)

// AvailabilityScheduler runs periodic wishlist availability checks.
type AvailabilityScheduler struct {
	svc      *AvailabilityService
	coinRepo *repository.CoinRepository
	logger   *Logger
	stopCh   chan struct{}
}

// NewAvailabilityScheduler creates a new scheduler.
func NewAvailabilityScheduler(
	svc *AvailabilityService,
	coinRepo *repository.CoinRepository,
) *AvailabilityScheduler {
	return &AvailabilityScheduler{
		svc:      svc,
		coinRepo: coinRepo,
		logger:   AppLogger,
		stopCh:   make(chan struct{}),
	}
}

// Start begins the periodic check loop. Call from a goroutine.
func (s *AvailabilityScheduler) Start() {
	s.logger.Info("scheduler", "Wishlist availability scheduler started")

	// Initial delay to let the app finish startup
	select {
	case <-time.After(30 * time.Second):
	case <-s.stopCh:
		return
	}

	for {
		s.runCycle()

		interval := s.getInterval()
		s.logger.Info("scheduler", "Next availability check in %s", interval)

		select {
		case <-time.After(interval):
		case <-s.stopCh:
			s.logger.Info("scheduler", "Scheduler stopped")
			return
		}
	}
}

// Stop signals the scheduler to shut down.
func (s *AvailabilityScheduler) Stop() {
	close(s.stopCh)
}

// runCycle executes one full availability check for all users.
func (s *AvailabilityScheduler) runCycle() {
	enabled := GetSetting(SettingWishlistCheckEnabled)
	if enabled != "true" {
		s.logger.Debug("scheduler", "Wishlist checking disabled, skipping cycle")
		return
	}

	s.logger.Info("scheduler", "Starting scheduled availability check cycle")

	coins, err := s.coinRepo.GetAllWishlistWithURLs()
	if err != nil {
		s.logger.Error("scheduler", "Failed to fetch all wishlist coins: %s", err)
		return
	}

	if len(coins) == 0 {
		s.logger.Info("scheduler", "No wishlist coins with URLs found")
		return
	}

	// Group coins by user
	userCoins := make(map[uint]bool)
	for _, coin := range coins {
		userCoins[coin.UserID] = true
	}

	s.logger.Info("scheduler", "Found %d coins across %d users", len(coins), len(userCoins))

	for userID := range userCoins {
		_, err := s.svc.CheckWishlistForUser(userID, "scheduled", nil)
		if err != nil {
			s.logger.Error("scheduler", "Scheduled check failed for user %d: %s", userID, err)
		}
	}

	s.logger.Info("scheduler", "Scheduled availability check cycle complete")
}

// getInterval returns the configured check interval.
func (s *AvailabilityScheduler) getInterval() time.Duration {
	hoursStr := GetSetting(SettingWishlistCheckInterval)
	hours, err := strconv.Atoi(hoursStr)
	if err != nil || hours < 1 {
		hours = 24
	}
	return time.Duration(hours) * time.Hour
}
