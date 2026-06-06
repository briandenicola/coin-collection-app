package services

import (
	"fmt"
	"sync"
	"time"
)

// SetSnapshotScheduler captures daily set valuation snapshots when enabled.
type SetSnapshotScheduler struct {
	setSvc      *SetService
	settingsSvc *SettingsService
	logger      *Logger
	stopCh      chan struct{}
	once        sync.Once
}

func NewSetSnapshotScheduler(setSvc *SetService, settingsSvc *SettingsService, logger *Logger) *SetSnapshotScheduler {
	return &SetSnapshotScheduler{
		setSvc:      setSvc,
		settingsSvc: settingsSvc,
		logger:      logger,
		stopCh:      make(chan struct{}),
	}
}

func (s *SetSnapshotScheduler) Start() {
	s.logger.Info("set-snapshot-scheduler", "Set snapshot scheduler started")
	for {
		wait := s.timeUntilNextRun()
		select {
		case <-time.After(wait):
		case <-s.stopCh:
			return
		}
		if s.settingsSvc.GetSetting(SettingSetSnapshotEnabled) != "true" {
			continue
		}
		if err := s.setSvc.CreateSnapshotsForAllUsers(); err != nil {
			s.logger.Error("set-snapshot-scheduler", "Set snapshot cycle failed: %v", err)
		}
	}
}

func (s *SetSnapshotScheduler) Stop() {
	s.once.Do(func() { close(s.stopCh) })
}

func (s *SetSnapshotScheduler) timeUntilNextRun() time.Duration {
	now := time.Now()
	h, m := 4, 0
	if _, err := fmt.Sscanf(s.settingsSvc.GetSetting(SettingSetSnapshotStartTime), "%d:%d", &h, &m); err != nil || h < 0 || h > 23 || m < 0 || m > 59 {
		h, m = 4, 0
	}
	next := time.Date(now.Year(), now.Month(), now.Day(), h, m, 0, 0, now.Location())
	if !next.After(now) {
		next = next.Add(24 * time.Hour)
	}
	return next.Sub(now)
}
