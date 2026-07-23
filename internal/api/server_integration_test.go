package api

import (
	"context"
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/ccawmiku/disk-insight/internal/analytics"
	"github.com/ccawmiku/disk-insight/internal/model"
	"github.com/ccawmiku/disk-insight/internal/scanner"
	"github.com/ccawmiku/disk-insight/internal/store"
)

func TestServerExposesScannedDashboardAndSettings(t *testing.T) {
	rootPath := t.TempDir()
	if err := os.WriteFile(filepath.Join(rootPath, "clip.mp4"), make([]byte, 2048), 0o644); err != nil {
		t.Fatal(err)
	}
	dataStore, err := store.Open(filepath.Join(t.TempDir(), "api.db"))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = dataStore.Close() })
	if err := dataStore.SyncRoots(context.Background(), []model.RootConfig{{Name: "Media", Path: rootPath, Enabled: true}}); err != nil {
		t.Fatal(err)
	}
	roots, _ := dataStore.RootConfigs(context.Background())
	dashboards := analytics.NewDashboardService(dataStore)
	manager := scanner.New(dataStore, dashboards.Invalidate)
	if err := manager.Run(context.Background(), roots[0], nil); err != nil {
		t.Fatal(err)
	}
	webPath := t.TempDir()
	if err := os.WriteFile(filepath.Join(webPath, "index.html"), []byte("<!doctype html><title>Disk Insight</title>"), 0o644); err != nil {
		t.Fatal(err)
	}
	handler := New(dataStore, manager, dashboards, webPath, slog.New(slog.NewTextHandler(io.Discard, nil)))

	request := httptest.NewRequest(http.MethodGet, "/api/v1/dashboard?rootId=1&sizeScale=linear&ageScale=linear", nil)
	response := httptest.NewRecorder()
	handler.ServeHTTP(response, request)
	if response.Code != http.StatusOK {
		t.Fatalf("dashboard status = %d, body = %s", response.Code, response.Body.String())
	}
	var dashboard model.Dashboard
	if err := json.NewDecoder(response.Body).Decode(&dashboard); err != nil {
		t.Fatal(err)
	}
	if dashboard.Summary.FileCount != 1 || len(dashboard.Size) != 80 {
		t.Fatalf("unexpected dashboard: %#v", dashboard)
	}

	badSettings := strings.NewReader(`{"scheduleKind":"hourly","scheduleTime":"03:00","scheduleDay":1,"timezone":"Asia/Shanghai","theme":"tropical-coral","language":"zh-CN","exclude":[]}`)
	request = httptest.NewRequest(http.MethodPut, "/api/v1/settings", badSettings)
	response = httptest.NewRecorder()
	handler.ServeHTTP(response, request)
	if response.Code != http.StatusBadRequest {
		t.Fatalf("invalid settings status = %d", response.Code)
	}

	request = httptest.NewRequest(http.MethodGet, "/", nil)
	response = httptest.NewRecorder()
	handler.ServeHTTP(response, request)
	if response.Code != http.StatusOK || !strings.Contains(response.Body.String(), "Disk Insight") {
		t.Fatalf("static response = %d %q", response.Code, response.Body.String())
	}
}

func TestServerReturnsNoScanBeforeInitialSnapshotCompletes(t *testing.T) {
	rootPath := t.TempDir()
	dataStore, err := store.Open(filepath.Join(t.TempDir(), "api.db"))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = dataStore.Close() })
	if err := dataStore.SyncRoots(context.Background(), []model.RootConfig{{Name: "Empty", Path: rootPath, Enabled: true}}); err != nil {
		t.Fatal(err)
	}

	dashboards := analytics.NewDashboardService(dataStore)
	manager := scanner.New(dataStore, dashboards.Invalidate)
	handler := New(dataStore, manager, dashboards, t.TempDir(), slog.New(slog.NewTextHandler(io.Discard, nil)))
	request := httptest.NewRequest(http.MethodGet, "/api/v1/dashboard?rootId=1", nil)
	response := httptest.NewRecorder()

	handler.ServeHTTP(response, request)

	if response.Code != http.StatusNotFound {
		t.Fatalf("dashboard status = %d, body = %s", response.Code, response.Body.String())
	}
	if !strings.Contains(response.Body.String(), analytics.ErrNoScan.Error()) {
		t.Fatalf("dashboard body = %s", response.Body.String())
	}

	request = httptest.NewRequest(http.MethodGet, "/api/v1/scan-errors?rootId=1", nil)
	response = httptest.NewRecorder()
	handler.ServeHTTP(response, request)
	if response.Code != http.StatusOK || strings.TrimSpace(response.Body.String()) != "[]" {
		t.Fatalf("scan errors response = %d %q", response.Code, response.Body.String())
	}
}
