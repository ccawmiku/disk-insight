package classify

import "testing"

func TestCategory(t *testing.T) {
	tests := map[string]string{
		"movie.MKV":      Video,
		"song.flac":      Audio,
		"photo.avif":     Image,
		"report.docx":    Document,
		"backup.7z":      Archive,
		"main.go":        Code,
		"events.parquet": Data,
		"setup.exe":      Executable,
		"server.vhdx":    DiskImage,
		"display.woff2":  Font,
		"part.step":      Design3D,
		"README":         Other,
		"stream.ts":      Video,
	}
	for name, want := range tests {
		if got := Category(name); got != want {
			t.Errorf("Category(%q) = %q, want %q", name, got, want)
		}
	}
}
