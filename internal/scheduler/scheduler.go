package scheduler

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/ccawmiku/disk-insight/internal/model"
	"github.com/ccawmiku/disk-insight/internal/scanner"
	"github.com/ccawmiku/disk-insight/internal/store"
)

type Scheduler struct {
	store   *store.Store
	scanner *scanner.Manager
	mu      sync.Mutex
	lastKey string
}

func New(dataStore *store.Store, scanManager *scanner.Manager) *Scheduler {
	return &Scheduler{store: dataStore, scanner: scanManager}
}

func (s *Scheduler) Run(ctx context.Context) {
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()
	s.tick(ctx, time.Now())
	for {
		select {
		case <-ctx.Done():
			return
		case now := <-ticker.C:
			s.tick(ctx, now)
		}
	}
}

func (s *Scheduler) tick(ctx context.Context, now time.Time) {
	settings, err := s.store.Settings(ctx)
	if err != nil || settings.ScheduleKind == "off" {
		return
	}
	location, err := time.LoadLocation(settings.Timezone)
	if err != nil {
		return
	}
	local := now.In(location)
	if local.Format("15:04") != settings.ScheduleTime {
		return
	}
	if settings.ScheduleKind == "weekly" && weekdayNumber(local.Weekday()) != settings.ScheduleDay {
		return
	}
	key := fmt.Sprintf("%s:%s:%d", local.Format("2006-01-02"), settings.ScheduleKind, settings.ScheduleDay)
	s.mu.Lock()
	if s.lastKey == key {
		s.mu.Unlock()
		return
	}
	s.lastKey = key
	s.mu.Unlock()
	roots, err := s.store.RootConfigs(ctx)
	if err != nil {
		return
	}
	for _, root := range roots {
		_ = s.scanner.Start(root, settings.Exclude)
	}
}

func weekdayNumber(day time.Weekday) int {
	if day == time.Sunday {
		return 7
	}
	return int(day)
}

func StartInitialScans(ctx context.Context, dataStore *store.Store, manager *scanner.Manager) error {
	roots, err := dataStore.Roots(ctx)
	if err != nil {
		return err
	}
	configs, err := dataStore.RootConfigs(ctx)
	if err != nil {
		return err
	}
	settings, err := dataStore.Settings(ctx)
	if err != nil {
		return err
	}
	byID := make(map[int64]model.RootConfig, len(configs))
	for _, config := range configs {
		byID[config.ID] = config
	}
	for _, root := range roots {
		if root.CurrentScanID == nil {
			if config, ok := byID[root.ID]; ok {
				_ = manager.Start(config, settings.Exclude)
			}
		}
	}
	return nil
}
