package api

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"mime"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/ccawmiku/disk-insight/internal/analytics"
	"github.com/ccawmiku/disk-insight/internal/model"
	"github.com/ccawmiku/disk-insight/internal/scanner"
	"github.com/ccawmiku/disk-insight/internal/store"
)

type Server struct {
	store     *store.Store
	scanner   *scanner.Manager
	dashboard *analytics.DashboardService
	webPath   string
	logger    *slog.Logger
}

func New(dataStore *store.Store, scanManager *scanner.Manager, dashboard *analytics.DashboardService, webPath string, logger *slog.Logger) http.Handler {
	server := &Server{store: dataStore, scanner: scanManager, dashboard: dashboard, webPath: webPath, logger: logger}
	mux := http.NewServeMux()
	mux.HandleFunc("GET /api/v1/health", server.health)
	mux.HandleFunc("GET /api/v1/roots", server.roots)
	mux.HandleFunc("GET /api/v1/settings", server.getSettings)
	mux.HandleFunc("PUT /api/v1/settings", server.putSettings)
	mux.HandleFunc("GET /api/v1/scans/progress", server.scanProgress)
	mux.HandleFunc("POST /api/v1/scans", server.startScans)
	mux.HandleFunc("DELETE /api/v1/scans/{rootID}", server.cancelScan)
	mux.HandleFunc("GET /api/v1/scan-errors", server.scanErrors)
	mux.HandleFunc("GET /api/v1/tree", server.tree)
	mux.HandleFunc("GET /api/v1/dashboard", server.getDashboard)
	mux.HandleFunc("GET /", server.static)
	return server.securityHeaders(server.logRequests(mux))
}

func (s *Server) health(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, map[string]any{"status": "ok", "version": "dev"})
}

func (s *Server) roots(w http.ResponseWriter, request *http.Request) {
	roots, err := s.store.Roots(request.Context())
	if err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	writeJSON(w, http.StatusOK, roots)
}

func (s *Server) getSettings(w http.ResponseWriter, request *http.Request) {
	settings, err := s.store.Settings(request.Context())
	if err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	writeJSON(w, http.StatusOK, settings)
}

func (s *Server) putSettings(w http.ResponseWriter, request *http.Request) {
	var settings model.Settings
	if err := decodeJSON(request, &settings); err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}
	if err := validateSettings(settings); err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}
	if err := s.store.UpdateSettings(request.Context(), settings); err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	writeJSON(w, http.StatusOK, settings)
}

func (s *Server) scanProgress(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, s.scanner.Progress())
}

func (s *Server) startScans(w http.ResponseWriter, request *http.Request) {
	var body struct {
		RootIDs []int64 `json:"rootIds"`
	}
	if request.ContentLength > 0 {
		if err := decodeJSON(request, &body); err != nil {
			writeError(w, http.StatusBadRequest, err)
			return
		}
	}
	settings, err := s.store.Settings(request.Context())
	if err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	roots, err := s.store.RootConfigs(request.Context())
	if err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	selected := make(map[int64]struct{}, len(body.RootIDs))
	for _, id := range body.RootIDs {
		selected[id] = struct{}{}
	}
	started := make([]int64, 0, len(roots))
	for _, root := range roots {
		if len(selected) > 0 {
			if _, ok := selected[root.ID]; !ok {
				continue
			}
		}
		if err := s.scanner.Start(root, settings.Exclude); err != nil {
			if errors.Is(err, scanner.ErrAlreadyRunning) {
				continue
			}
			writeError(w, http.StatusInternalServerError, err)
			return
		}
		started = append(started, root.ID)
	}
	if len(started) == 0 {
		writeError(w, http.StatusConflict, errors.New("no eligible scan root was started"))
		return
	}
	writeJSON(w, http.StatusAccepted, map[string]any{"startedRootIds": started})
}

func (s *Server) cancelScan(w http.ResponseWriter, request *http.Request) {
	rootID, err := strconv.ParseInt(request.PathValue("rootID"), 10, 64)
	if err != nil || rootID <= 0 {
		writeError(w, http.StatusBadRequest, errors.New("invalid root id"))
		return
	}
	if !s.scanner.Cancel(rootID) {
		writeError(w, http.StatusNotFound, errors.New("scan is not running"))
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (s *Server) scanErrors(w http.ResponseWriter, request *http.Request) {
	rootID, err := requiredInt64(request, "rootId")
	if err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}
	items, err := s.store.ScanErrors(request.Context(), rootID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	writeJSON(w, http.StatusOK, items)
}

func (s *Server) tree(w http.ResponseWriter, request *http.Request) {
	rootID, err := requiredInt64(request, "rootId")
	if err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}
	nodes, err := s.dashboard.Tree(request.Context(), rootID, request.URL.Query().Get("path"))
	if err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}
	writeJSON(w, http.StatusOK, nodes)
}

func (s *Server) getDashboard(w http.ResponseWriter, request *http.Request) {
	rootID, err := requiredInt64(request, "rootId")
	if err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}
	var categories []string
	for _, value := range strings.Split(request.URL.Query().Get("categories"), ",") {
		if value = strings.TrimSpace(value); value != "" {
			categories = append(categories, value)
		}
	}
	dashboard, err := s.dashboard.Build(request.Context(), rootID, request.URL.Query().Get("path"), categories, request.URL.Query().Get("sizeScale"), request.URL.Query().Get("ageScale"))
	if err != nil {
		if errors.Is(err, analytics.ErrNoScan) {
			writeError(w, http.StatusNotFound, err)
		} else {
			writeError(w, http.StatusBadRequest, err)
		}
		return
	}
	writeJSON(w, http.StatusOK, dashboard)
}

func (s *Server) static(w http.ResponseWriter, request *http.Request) {
	if request.URL.Path == "/api" || strings.HasPrefix(request.URL.Path, "/api/") {
		http.NotFound(w, request)
		return
	}
	relative := strings.TrimPrefix(filepath.Clean(request.URL.Path), string(filepath.Separator))
	if relative == "." || relative == "" {
		relative = "index.html"
	}
	if relative == ".." || strings.HasPrefix(relative, ".."+string(filepath.Separator)) || filepath.IsAbs(relative) {
		http.NotFound(w, request)
		return
	}
	filePath := filepath.Join(s.webPath, relative)
	if info, err := os.Stat(filePath); err != nil || info.IsDir() {
		filePath = filepath.Join(s.webPath, "index.html")
	}
	file, err := os.Open(filePath)
	if err != nil {
		writeError(w, http.StatusServiceUnavailable, errors.New("web application has not been built"))
		return
	}
	defer file.Close()
	info, err := file.Stat()
	if err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	if contentType := mime.TypeByExtension(filepath.Ext(filePath)); contentType != "" {
		w.Header().Set("Content-Type", contentType)
	}
	if strings.Contains(relative, "assets") {
		w.Header().Set("Cache-Control", "public, max-age=31536000, immutable")
	} else {
		w.Header().Set("Cache-Control", "no-cache")
	}
	http.ServeContent(w, request, info.Name(), info.ModTime(), file)
}

func (s *Server) securityHeaders(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, request *http.Request) {
		w.Header().Set("X-Content-Type-Options", "nosniff")
		w.Header().Set("X-Frame-Options", "DENY")
		w.Header().Set("Referrer-Policy", "no-referrer")
		w.Header().Set("Permissions-Policy", "camera=(), microphone=(), geolocation=()")
		w.Header().Set("Content-Security-Policy", "default-src 'self'; script-src 'self'; style-src 'self' 'unsafe-inline'; img-src 'self' data:; connect-src 'self'; font-src 'self'")
		next.ServeHTTP(w, request)
	})
}

func (s *Server) logRequests(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, request *http.Request) {
		started := time.Now()
		next.ServeHTTP(w, request)
		s.logger.Debug("request", "method", request.Method, "path", request.URL.Path, "duration", time.Since(started))
	})
}

func validateSettings(settings model.Settings) error {
	if settings.ScheduleKind != "daily" && settings.ScheduleKind != "weekly" && settings.ScheduleKind != "off" {
		return errors.New("scheduleKind must be daily, weekly, or off")
	}
	if _, err := time.Parse("15:04", settings.ScheduleTime); err != nil {
		return errors.New("scheduleTime must use HH:MM")
	}
	if settings.ScheduleDay < 1 || settings.ScheduleDay > 7 {
		return errors.New("scheduleDay must be between 1 and 7")
	}
	if _, err := time.LoadLocation(settings.Timezone); err != nil {
		return errors.New("unknown timezone")
	}
	if len(settings.Exclude) > 100 {
		return errors.New("too many exclusion rules")
	}
	return nil
}

func requiredInt64(request *http.Request, name string) (int64, error) {
	value, err := strconv.ParseInt(request.URL.Query().Get(name), 10, 64)
	if err != nil || value <= 0 {
		return 0, fmt.Errorf("%s must be a positive integer", name)
	}
	return value, nil
}

func decodeJSON(request *http.Request, target any) error {
	decoder := json.NewDecoder(io.LimitReader(request.Body, 1<<20))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(target); err != nil {
		return fmt.Errorf("invalid JSON body: %w", err)
	}
	if decoder.Decode(&struct{}{}) != io.EOF {
		return errors.New("request body must contain one JSON object")
	}
	return nil
}

func writeJSON(w http.ResponseWriter, status int, value any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(value)
}

func writeError(w http.ResponseWriter, status int, err error) {
	writeJSON(w, status, map[string]string{"error": err.Error()})
}

func Shutdown(ctx context.Context, server *http.Server) error { return server.Shutdown(ctx) }
