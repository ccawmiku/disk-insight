package analytics

import "testing"

func TestNiceAxisMax(t *testing.T) {
	const (
		KiB = int64(1 << 10)
		MiB = int64(1 << 20)
		GiB = int64(1 << 30)
	)
	tests := []struct {
		name string
		max  int64
		want int64
	}{
		{name: "zero", max: 0, want: 5 * KiB},
		{name: "3.3 MiB", max: 33 * MiB / 10, want: 5 * MiB},
		{name: "6 GiB", max: 6 * GiB, want: 10 * GiB},
		{name: "11 GiB", max: 11 * GiB, want: 15 * GiB},
		{name: "67 GiB", max: 67 * GiB, want: 70 * GiB},
		{name: "983 MiB", max: 983 * MiB, want: GiB},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if got := NiceAxisMax(test.max); got != test.want {
				t.Fatalf("NiceAxisMax(%d) = %d, want %d", test.max, got, test.want)
			}
		})
	}
}

func TestBinsIncludeMaximum(t *testing.T) {
	samples := []fileSample{{size: 0}, {size: 10}, {size: 100}}
	for _, scale := range []string{"linear", "log"} {
		points := buildSizePoints(samples, 100, scale, 3, 110)
		var count int64
		for _, point := range points {
			count += point.Count
		}
		if count != 3 {
			t.Fatalf("%s bins contain %d samples, want 3", scale, count)
		}
		if points[len(points)-1].CumulativeCount != 100 {
			t.Fatalf("%s cumulative count = %v, want 100", scale, points[len(points)-1].CumulativeCount)
		}
	}
}

func TestModificationHeatmapUsesEqualIntervalsAndTracksBytes(t *testing.T) {
	samples := []fileSample{
		{size: 100, age: 0},
		{size: 200, age: 600},
	}
	points := buildAgePoints(samples, "log")
	if len(points) != 60 {
		t.Fatalf("heatmap bins = %d, want 60", len(points))
	}
	var count, bytes int64
	for index, point := range points {
		count += point.Count
		bytes += point.Bytes
		if index > 0 {
			previousWidth := points[index-1].UpperSeconds
			if index > 1 {
				previousWidth -= points[index-2].UpperSeconds
			}
			width := point.UpperSeconds - points[index-1].UpperSeconds
			if width < previousWidth-1 || width > previousWidth+1 {
				t.Fatalf("bucket %d width = %d, previous = %d", index, width, previousWidth)
			}
		}
	}
	if count != 2 || bytes != 300 {
		t.Fatalf("heatmap totals = %d files, %d bytes", count, bytes)
	}
}
