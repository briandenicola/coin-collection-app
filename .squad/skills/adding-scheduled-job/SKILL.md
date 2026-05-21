# Skill: Adding a Scheduled Background Job to the Go API

**Category:** Backend Development  
**Applies to:** Ancient Coins Go API  
**Last Updated:** 2026-05-21  

## When to Use This Skill

Use this recipe when you need to add a new recurring background task that runs on a schedule (e.g., daily cleanup, periodic checks, scheduled notifications).

## Prerequisites

- Existing scheduler pattern in the codebase (e.g., `availability_scheduler.go`, `valuation_scheduler.go`)
- Database repository for any data the scheduler needs to query or update
- Service layer for any business logic (e.g., sending notifications)
- Settings service for configuration

## Step-by-Step Recipe

### 1. Create the Scheduler Service

**File:** `src/api/services/{feature}_scheduler.go`

**Pattern:**
```go
package services

import (
	"sync"
	"time"
	"github.com/briandenicola/ancient-coins-api/repository"
)

type FeatureScheduler struct {
	repo        *repository.FeatureRepository
	settingsSvc *SettingsService
	logger      *Logger
	stopCh      chan struct{}
	once        sync.Once
}

func NewFeatureScheduler(
	repo *repository.FeatureRepository,
	settingsSvc *SettingsService,
	logger *Logger,
) *FeatureScheduler {
	return &FeatureScheduler{
		repo:        repo,
		settingsSvc: settingsSvc,
		logger:      logger,
		stopCh:      make(chan struct{}),
	}
}

func (s *FeatureScheduler) Start() {
	s.logger.Info("scheduler", "Feature scheduler started")
	
	// Initial startup delay
	select {
	case <-time.After(30 * time.Second):
	case <-s.stopCh:
		return
	}
	
	for {
		wait := s.timeUntilNextRun()
		s.logger.Info("scheduler", "Next feature check in %s", wait)
		
		select {
		case <-time.After(wait):
		case <-s.stopCh:
			s.logger.Info("scheduler", "Feature scheduler stopped")
			return
		}
		
		s.runCycle()
	}
}

func (s *FeatureScheduler) Stop() {
	s.once.Do(func() { close(s.stopCh) })
}

func (s *FeatureScheduler) timeUntilNextRun() time.Duration {
	now := time.Now()
	startHour, startMin := s.getStartTime()
	interval := s.getInterval()
	
	anchor := time.Date(now.Year(), now.Month(), now.Day(), startHour, startMin, 0, 0, now.Location())
	if anchor.After(now) {
		return anchor.Sub(now)
	}
	
	elapsed := now.Sub(anchor)
	periods := int(elapsed/interval) + 1
	next := anchor.Add(time.Duration(periods) * interval)
	return next.Sub(now)
}

func (s *FeatureScheduler) getStartTime() (int, int) {
	raw := s.settingsSvc.GetSetting(SettingFeatureCheckStartTime)
	var h, m int
	if _, err := fmt.Sscanf(raw, "%d:%d", &h, &m); err != nil || h < 0 || h > 23 || m < 0 || m > 59 {
		return 8, 0 // default
	}
	return h, m
}

func (s *FeatureScheduler) getInterval() time.Duration {
	minStr := s.settingsSvc.GetSetting(SettingFeatureCheckInterval)
	mins, err := strconv.Atoi(minStr)
	if err != nil || mins < 5 {
		mins = 1440 // default: 24 hours
	}
	return time.Duration(mins) * time.Minute
}

func (s *FeatureScheduler) runCycle() {
	enabled := s.settingsSvc.GetSetting(SettingFeatureCheckEnabled)
	if enabled != "true" {
		s.logger.Debug("scheduler", "Feature check disabled, skipping cycle")
		return
	}
	
	s.logger.Info("scheduler", "Starting scheduled feature check cycle")
	
	// Your business logic here
	// Query data, process, send notifications, etc.
	
	s.logger.Info("scheduler", "Feature check cycle complete")
}
```

**Key Points:**
- Use `sync.Once` in `Stop()` to prevent double-close panics
- Always check the `Enabled` setting in `runCycle()`
- Log at Info level for start/stop/cycle, Debug for skipped cycles
- Use `time.After` with `select` to allow clean shutdown

### 2. Add Settings Constants

**File:** `src/api/services/settings_service.go`

**Add three constants:**
```go
const (
	// ... existing constants
	SettingFeatureCheckEnabled  = "FeatureCheckEnabled"
	SettingFeatureCheckInterval = "FeatureCheckInterval"
	SettingFeatureCheckStartTime = "FeatureCheckStartTime"
)
```

**Add defaults to the map:**
```go
var settingDefaults = map[string]string{
	// ... existing defaults
	SettingFeatureCheckEnabled: "false",
	SettingFeatureCheckInterval: "1440", // minutes
	SettingFeatureCheckStartTime: "08:00",
}
```

**Naming Convention:**
- `{Feature}CheckEnabled` — Boolean string (`"true"` / `"false"`)
- `{Feature}CheckInterval` — Integer string (minutes for sub-daily, or use `IntervalDays` suffix for daily+)
- `{Feature}CheckStartTime` — HH:MM format (24-hour)

### 3. Add Repository Methods (if needed)

If your scheduler needs to query data, add methods to the appropriate repository:

```go
// GetItemsRequiringAction returns items that need processing today.
func (r *FeatureRepository) GetItemsRequiringAction() ([]models.Feature, error) {
	var items []models.Feature
	now := time.Now()
	startOfDay := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
	endOfDay := startOfDay.Add(24 * time.Hour)
	
	err := r.db.Where("status = ? AND action_date >= ? AND action_date < ?",
		"pending", startOfDay, endOfDay).
		Order("user_id ASC").
		Find(&items).Error
	return items, err
}
```

**Testing:** Always add a unit test for new repository methods using the in-memory SQLite pattern.

### 4. Wire the Scheduler in main.go

**Location:** After Ollama check, before "Application ready" log.

**Steps:**
1. Ensure all dependencies (repos, services) are already created before this point
2. Construct the scheduler
3. Start it in a goroutine

```go
// Check Ollama connectivity at startup
func() {
	// ... ollama check
}()

// Start wishlist availability scheduler
scheduler := services.NewAvailabilityScheduler(availSvc, coinRepo, settingsSvc, logger)
go scheduler.Start()

// Start YOUR NEW SCHEDULER HERE
featureScheduler := services.NewFeatureScheduler(featureRepo, settingsSvc, logger)
go featureScheduler.Start()

logger.Info("startup", "Application ready")
```

**Note:** If the scheduler needs a repository that's defined inside the `protected` route group, move the repository creation up before the schedulers (see auction ending scheduler example).

### 5. Update README.md

**File:** `src/api/README.md`

Add your scheduler to the "Background Schedulers" section:

```markdown
## Background Schedulers

The API runs [N] background schedulers that start automatically on server startup:

...existing schedulers...

N. **Feature Scheduler** — Brief description. Configured via `FeatureCheckEnabled`, `FeatureCheckStartTime`, and `FeatureCheckInterval` settings. Sends notifications when X happens.
```

### 6. Test & Verify

**Commands:**
```bash
cd src/api
go vet ./...       # Must pass
go test -v ./...   # All tests must pass
```

**Manual Testing:**
1. Start the server
2. Check logs for "Feature scheduler started"
3. Set the enabled flag via Admin Settings
4. Verify cycle runs at the configured time
5. Check notification delivery

## Common Patterns

### Daily Cadence
Default interval: `1440` minutes (24 hours)  
Default start time: `08:00` (8 AM local)

### Idempotency (Daily Jobs)
If your scheduler runs more than once per day but should only act once per day per user:

**In-memory map:**
```go
type FeatureScheduler struct {
	// ...
	lastNotified map[uint]string // userID -> date string (YYYY-MM-DD)
	mu          sync.RWMutex
}

func (s *FeatureScheduler) runCycle() {
	today := time.Now().Format("2006-01-02")
	
	for userID, items := range userItems {
		s.mu.RLock()
		lastDate := s.lastNotified[userID]
		s.mu.RUnlock()
		
		if lastDate == today {
			continue // already notified today
		}
		
		// Do the work
		s.notifyUser(userID, items)
		
		s.mu.Lock()
		s.lastNotified[userID] = today
		s.mu.Unlock()
	}
}
```

**When to use in-memory vs. DB column:**
- In-memory: Simple, daily cadence, acceptable to lose state on restart
- DB column: Need persistent tracking, multiple services, or auditing

### Grouped Notifications
If notifying multiple users, group by user and send one consolidated message per user (not one per item):

```go
// Group by user
userItems := make(map[uint][]models.Feature)
for _, item := range items {
	userItems[item.UserID] = append(userItems[item.UserID], item)
}

// Send one notification per user
for userID, items := range userItems {
	s.notifyUser(userID, items)
}
```

### Pushover Integration
```go
func (s *FeatureScheduler) notifyUser(userID uint, items []models.Feature) {
	user, err := s.userRepo.FindByID(userID)
	if err != nil || !user.PushoverEnabled || user.PushoverUserKey == "" {
		return
	}
	
	title := "Feature Alert"
	message := fmt.Sprintf("%d items require attention:\n\n", len(items))
	for _, item := range items {
		message += fmt.Sprintf("• %s\n", item.Name)
	}
	
	go func() {
		s.pushoverSvc.SendNotification(user.PushoverUserKey, title, message, "")
	}()
}
```

## Architecture Checklist

- [ ] Scheduler uses constructor injection
- [ ] Settings follow naming convention (`{Feature}CheckEnabled`, etc.)
- [ ] Repository owns all GORM queries
- [ ] `sync.Once` used in `Stop()` to prevent double-close
- [ ] Logs at appropriate levels (Info for lifecycle, Debug for skipped cycles)
- [ ] No new HTTP endpoints (uses existing settings API)
- [ ] Tests added for new repository methods
- [ ] `go vet ./...` passes
- [ ] `go test -v ./...` passes
- [ ] README.md updated

## Examples in Codebase

1. **`services/availability_scheduler.go`** — Wishlist availability checking (2-hour interval)
2. **`services/valuation_scheduler.go`** — Collection valuation (7-day interval)
3. **`services/auction_ending_scheduler.go`** — Auction ending alerts (24-hour interval, in-memory idempotency)

## Common Pitfalls

1. **Forgetting `sync.Once` in Stop()** — Causes double-close panic on shutdown
2. **Not checking the Enabled setting** — Scheduler runs even when disabled
3. **Defining repo inside protected group** — Move it before schedulers if needed
4. **Hardcoding interval/start time** — Always read from settings
5. **Not handling missing Pushover config** — Check `PushoverEnabled` and key presence
6. **Sending one notification per item** — Batch per user instead

## Related Skills

- Adding a new repository method with tests
- Adding app settings (see `settings_service.go`)
- Integrating Pushover notifications (see `pushover_service.go`)
