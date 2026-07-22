package scanner

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/ccawmiku/disk-insight/internal/analytics"
	"github.com/ccawmiku/disk-insight/internal/model"
	"github.com/ccawmiku/disk-insight/internal/store"
)

func TestScanBuildsQueryableSnapshot(t *testing.T) {
	rootPath := t.TempDir()
	mustWrite(t, filepath.Join(rootPath, "movie.mp4"), 4096)
	mustWrite(t, filepath.Join(rootPath, "docs", "report.pdf"), 1024)
	mustWrite(t, filepath.Join(rootPath, "docs", "notes.txt"), 512)
	mustWrite(t, filepath.Join(rootPath, ".hidden", "secret.mp3"), 2048)

	dataStore, err := store.Open(filepath.Join(t.TempDir(), "scan.db"))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = dataStore.Close() })
	if err := dataStore.SyncRoots(context.Background(), []model.RootConfig{{Name: "Test", Path: rootPath, Enabled: true}}); err != nil {
		t.Fatal(err)
	}
	roots, err := dataStore.RootConfigs(context.Background())
	if err != nil || len(roots) != 1 {
		t.Fatalf("roots: %v, %v", roots, err)
	}
	manager := New(dataStore, nil)
	if err := manager.Run(context.Background(), roots[0], nil); err != nil {
		t.Fatal(err)
	}

	dashboard, err := analytics.NewDashboardService(dataStore).Build(context.Background(), roots[0].ID, "", nil, "linear", "linear")
	if err != nil {
		t.Fatal(err)
	}
	if dashboard.Summary.FileCount != 3 {
		t.Fatalf("file count = %d, want 3", dashboard.Summary.FileCount)
	}
	if dashboard.Summary.DirectoryCount != 1 {
		t.Fatalf("directory count = %d, want 1", dashboard.Summary.DirectoryCount)
	}
	if dashboard.Summary.LogicalSize != 5632 {
		t.Fatalf("logical size = %d, want 5632", dashboard.Summary.LogicalSize)
	}
	if len(dashboard.Categories) != 2 {
		t.Fatalf("categories = %#v", dashboard.Categories)
	}
	if len(dashboard.TopFiles) != 3 || dashboard.TopFiles[0].Name != "movie.mp4" {
		t.Fatalf("top files = %#v", dashboard.TopFiles)
	}
	tree, err := analytics.NewDashboardService(dataStore).Tree(context.Background(), roots[0].ID, "")
	if err != nil || len(tree) != 1 || tree[0].Path != "docs" {
		t.Fatalf("tree = %#v, err = %v", tree, err)
	}
}

func TestFailedReplacementKeepsLastCompleteSnapshot(t *testing.T) {
	rootPath := t.TempDir()
	mustWrite(t, filepath.Join(rootPath, "one.txt"), 10)
	dataStore, err := store.Open(filepath.Join(t.TempDir(), "scan.db"))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = dataStore.Close() })
	if err := dataStore.SyncRoots(context.Background(), []model.RootConfig{{Name: "Test", Path: rootPath, Enabled: true}}); err != nil {
		t.Fatal(err)
	}
	roots, _ := dataStore.RootConfigs(context.Background())
	manager := New(dataStore, nil)
	if err := manager.Run(context.Background(), roots[0], nil); err != nil {
		t.Fatal(err)
	}
	cancelled, cancel := context.WithCancel(context.Background())
	cancel()
	if err := manager.Run(cancelled, roots[0], nil); err == nil {
		t.Fatal("expected cancelled scan to fail")
	}
	dashboard, err := analytics.NewDashboardService(dataStore).Build(context.Background(), roots[0].ID, "", nil, "linear", "linear")
	if err != nil || dashboard.Summary.FileCount != 1 {
		t.Fatalf("last snapshot was not preserved: count=%d err=%v", dashboard.Summary.FileCount, err)
	}
}

func mustWrite(t *testing.T, path string, size int) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, make([]byte, size), 0o644); err != nil {
		t.Fatal(err)
	}
}
