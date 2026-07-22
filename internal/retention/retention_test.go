package retention

import (
	"testing"
	"time"
)

func TestKeepIDsUsesRollingBuckets(t *testing.T) {
	now := time.Date(2026, 7, 22, 12, 0, 0, 0, time.UTC)
	points := []Point{
		{ID: 1, CompletedAt: now.Add(-2 * time.Hour)},
		{ID: 2, CompletedAt: now.Add(-4 * time.Hour)},
		{ID: 3, CompletedAt: now.Add(-2 * 24 * time.Hour)},
		{ID: 4, CompletedAt: now.Add(-10 * 24 * time.Hour)},
		{ID: 5, CompletedAt: now.Add(-11 * 24 * time.Hour)},
		{ID: 6, CompletedAt: now.Add(-45 * 24 * time.Hour)},
		{ID: 7, CompletedAt: now.Add(-50 * 24 * time.Hour)},
		{ID: 8, CompletedAt: now.Add(-400 * 24 * time.Hour)},
		{ID: 9, CompletedAt: now.Add(-500 * 24 * time.Hour)},
	}
	keep := KeepIDs(now, points)
	if _, ok := keep[1]; !ok {
		t.Fatal("newest daily point was not kept")
	}
	if _, ok := keep[2]; ok {
		t.Fatal("older point in the same day was kept")
	}
	if _, ok := keep[3]; !ok {
		t.Fatal("point from another recent day was not kept")
	}
	if _, ok := keep[4]; !ok {
		t.Fatal("newest weekly point was not kept")
	}
	if _, ok := keep[5]; ok {
		t.Fatal("older point from the same week was kept")
	}
	if _, ok := keep[6]; !ok {
		t.Fatal("monthly point was not kept")
	}
	if _, ok := keep[7]; ok {
		t.Fatal("older point from the same month was kept")
	}
	if _, ok := keep[8]; !ok {
		t.Fatal("yearly point was not kept")
	}
	if _, ok := keep[9]; ok {
		t.Fatal("older point from the same year was kept")
	}
}
