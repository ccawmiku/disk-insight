package analytics

// NiceAxisMax returns a readable upper boundary above maxBytes. It uses
// binary units while keeping the visible numbers familiar: 3.3 MiB -> 5 MiB,
// 6 GiB -> 10 GiB, 11 GiB -> 15 GiB, and 983 MiB -> 1 GiB.
func NiceAxisMax(maxBytes int64) int64 {
	if maxBytes <= 0 {
		return 5 * 1024
	}
	unit := int64(1)
	for _, candidate := range []int64{1 << 40, 1 << 30, 1 << 20, 1 << 10, 1} {
		if maxBytes >= candidate {
			unit = candidate
			break
		}
	}
	value := float64(maxBytes) / float64(unit)
	if value >= 900 && unit < 1<<40 {
		return 1024 * unit
	}
	step := 5.0
	switch {
	case value >= 500:
		step = 100
	case value >= 100:
		step = 50
	case value >= 20:
		step = 10
	}
	boundary := int64(value/step+1) * int64(step) * unit
	if boundary <= maxBytes {
		boundary += int64(step) * unit
	}
	if boundary >= 1024*unit && unit < 1<<40 {
		return 1024 * unit
	}
	return boundary
}
