package retention

import (
	"fmt"
	"sort"
	"time"
)

type Point struct {
	ID          int64
	CompletedAt time.Time
}

// KeepIDs applies the rolling history policy: one point per day for the last
// seven days, per ISO week through day 31, per month through one year, and per
// year thereafter. The newest point in each bucket wins.
func KeepIDs(now time.Time, points []Point) map[int64]struct{} {
	sorted := append([]Point(nil), points...)
	sort.Slice(sorted, func(i, j int) bool { return sorted[i].CompletedAt.After(sorted[j].CompletedAt) })
	keep := make(map[int64]struct{}, len(sorted))
	buckets := make(map[string]struct{}, len(sorted))
	for _, point := range sorted {
		age := now.Sub(point.CompletedAt)
		var key string
		switch {
		case age <= 7*24*time.Hour:
			key = "d:" + point.CompletedAt.Format("2006-01-02")
		case age <= 31*24*time.Hour:
			year, week := point.CompletedAt.ISOWeek()
			key = fmt.Sprintf("w:%04d:%02d", year, week)
		case age <= 365*24*time.Hour:
			key = "m:" + point.CompletedAt.Format("2006-01")
		default:
			key = "y:" + point.CompletedAt.Format("2006")
		}
		if _, exists := buckets[key]; exists {
			continue
		}
		buckets[key] = struct{}{}
		keep[point.ID] = struct{}{}
	}
	return keep
}
