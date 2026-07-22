package store

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/ccawmiku/disk-insight/internal/model"
	"github.com/ccawmiku/disk-insight/internal/retention"
	_ "modernc.org/sqlite"
)

type Store struct {
	db *sql.DB
}

type Entry struct {
	Path           string
	ParentPath     string
	Name           string
	Kind           string
	Category       string
	Size           int64
	AllocatedSize  *int64
	ModifiedAt     time.Time
	Identity       string
	RecursiveFiles int64
	RecursiveDirs  int64
	RecursiveSize  int64
}

type RunSummary struct {
	Files         int64
	Directories   int64
	LogicalSize   int64
	AllocatedSize *int64
	Errors        int64
	LargestName   string
	LargestSize   int64
}

func Open(path string) (*Store, error) {
	if path == "" {
		return nil, errors.New("database path is required")
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return nil, fmt.Errorf("create database directory: %w", err)
	}
	dsn := "file:" + filepath.ToSlash(path) + "?_pragma=busy_timeout(5000)&_pragma=foreign_keys(1)&_pragma=journal_mode(WAL)"
	db, err := sql.Open("sqlite", dsn)
	if err != nil {
		return nil, fmt.Errorf("open database: %w", err)
	}
	db.SetMaxOpenConns(4)
	db.SetMaxIdleConns(4)
	store := &Store{db: db}
	if err := store.migrate(context.Background()); err != nil {
		db.Close()
		return nil, err
	}
	return store, nil
}

func (s *Store) Close() error { return s.db.Close() }

func (s *Store) migrate(ctx context.Context) error {
	const schema = `
CREATE TABLE IF NOT EXISTS roots (
  id INTEGER PRIMARY KEY,
  path TEXT NOT NULL UNIQUE,
  name TEXT NOT NULL,
  enabled INTEGER NOT NULL DEFAULT 1,
  current_scan_id INTEGER,
  created_at TEXT NOT NULL
);
CREATE TABLE IF NOT EXISTS scan_runs (
  id INTEGER PRIMARY KEY,
  root_id INTEGER NOT NULL REFERENCES roots(id),
  status TEXT NOT NULL,
  started_at TEXT NOT NULL,
  completed_at TEXT,
  file_count INTEGER NOT NULL DEFAULT 0,
  directory_count INTEGER NOT NULL DEFAULT 0,
  logical_size INTEGER NOT NULL DEFAULT 0,
  allocated_size INTEGER,
  error_count INTEGER NOT NULL DEFAULT 0,
  largest_name TEXT NOT NULL DEFAULT '',
  largest_size INTEGER NOT NULL DEFAULT 0,
  error_message TEXT NOT NULL DEFAULT ''
);
CREATE TABLE IF NOT EXISTS entries (
  id INTEGER PRIMARY KEY,
  run_id INTEGER NOT NULL REFERENCES scan_runs(id) ON DELETE CASCADE,
  root_id INTEGER NOT NULL REFERENCES roots(id),
  path TEXT NOT NULL,
  parent_path TEXT NOT NULL,
  name TEXT NOT NULL,
  kind TEXT NOT NULL CHECK(kind IN ('file','directory')),
  category TEXT NOT NULL DEFAULT '',
  size INTEGER NOT NULL DEFAULT 0,
  allocated_size INTEGER,
  modified_at TEXT NOT NULL,
  identity TEXT NOT NULL DEFAULT '',
  recursive_files INTEGER NOT NULL DEFAULT 0,
  recursive_dirs INTEGER NOT NULL DEFAULT 0,
  recursive_size INTEGER NOT NULL DEFAULT 0,
  UNIQUE(run_id, path)
);
CREATE INDEX IF NOT EXISTS entries_scope_idx ON entries(root_id, run_id, path);
CREATE INDEX IF NOT EXISTS entries_parent_idx ON entries(root_id, run_id, parent_path, kind);
CREATE INDEX IF NOT EXISTS entries_category_idx ON entries(root_id, run_id, category, path);
CREATE INDEX IF NOT EXISTS entries_size_idx ON entries(root_id, run_id, size DESC);
CREATE INDEX IF NOT EXISTS scan_runs_root_time_idx ON scan_runs(root_id, completed_at DESC);
CREATE TABLE IF NOT EXISTS scan_errors (
  id INTEGER PRIMARY KEY,
  run_id INTEGER NOT NULL REFERENCES scan_runs(id) ON DELETE CASCADE,
  path TEXT NOT NULL,
  message TEXT NOT NULL
);
CREATE TABLE IF NOT EXISTS settings (
  id INTEGER PRIMARY KEY CHECK(id = 1),
  schedule_kind TEXT NOT NULL,
  schedule_time TEXT NOT NULL,
  schedule_day INTEGER NOT NULL,
  timezone TEXT NOT NULL,
  theme TEXT NOT NULL,
  language TEXT NOT NULL,
  exclude_json TEXT NOT NULL
);
INSERT OR IGNORE INTO settings(id, schedule_kind, schedule_time, schedule_day, timezone, theme, language, exclude_json)
VALUES(1, 'daily', '03:00', 1, 'Asia/Shanghai', 'tropical-coral', 'zh-CN', '["$RECYCLE.BIN","System Volume Information"]');`
	if _, err := s.db.ExecContext(ctx, schema); err != nil {
		return fmt.Errorf("migrate database: %w", err)
	}
	return nil
}

func (s *Store) SyncRoots(ctx context.Context, roots []model.RootConfig) error {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()
	if _, err := tx.ExecContext(ctx, "UPDATE roots SET enabled = 0"); err != nil {
		return err
	}
	for _, root := range roots {
		absolute, err := filepath.Abs(root.Path)
		if err != nil {
			return fmt.Errorf("resolve root %q: %w", root.Path, err)
		}
		clean := filepath.Clean(absolute)
		name := strings.TrimSpace(root.Name)
		if name == "" {
			name = filepath.Base(clean)
		}
		if _, err := tx.ExecContext(ctx, `
INSERT INTO roots(path, name, enabled, created_at) VALUES(?, ?, 1, ?)
ON CONFLICT(path) DO UPDATE SET name = excluded.name, enabled = 1`, clean, name, time.Now().UTC().Format(time.RFC3339Nano)); err != nil {
			return fmt.Errorf("sync root %q: %w", clean, err)
		}
	}
	return tx.Commit()
}

func (s *Store) RootConfigs(ctx context.Context) ([]model.RootConfig, error) {
	rows, err := s.db.QueryContext(ctx, "SELECT id, name, path, enabled FROM roots WHERE enabled = 1 ORDER BY name")
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var result []model.RootConfig
	for rows.Next() {
		var root model.RootConfig
		if err := rows.Scan(&root.ID, &root.Name, &root.Path, &root.Enabled); err != nil {
			return nil, err
		}
		result = append(result, root)
	}
	return result, rows.Err()
}

func (s *Store) Roots(ctx context.Context) ([]model.Root, error) {
	rows, err := s.db.QueryContext(ctx, `
SELECT r.id, r.name, r.enabled, r.current_scan_id, sr.completed_at,
       COALESCE(sr.file_count, 0), COALESCE(sr.directory_count, 0), COALESCE(sr.logical_size, 0)
FROM roots r LEFT JOIN scan_runs sr ON sr.id = r.current_scan_id
WHERE r.enabled = 1 ORDER BY r.name`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var result []model.Root
	for rows.Next() {
		var root model.Root
		var scanID sql.NullInt64
		var completed sql.NullString
		if err := rows.Scan(&root.ID, &root.Name, &root.Enabled, &scanID, &completed, &root.LastFileCount, &root.LastDirectoryCount, &root.LastLogicalSize); err != nil {
			return nil, err
		}
		if scanID.Valid {
			root.CurrentScanID = &scanID.Int64
		}
		if completed.Valid {
			parsed, err := time.Parse(time.RFC3339Nano, completed.String)
			if err == nil {
				root.LastScanAt = &parsed
			}
		}
		result = append(result, root)
	}
	return result, rows.Err()
}

func (s *Store) Settings(ctx context.Context) (model.Settings, error) {
	var result model.Settings
	var excludeJSON string
	err := s.db.QueryRowContext(ctx, `SELECT schedule_kind, schedule_time, schedule_day, timezone, theme, language, exclude_json FROM settings WHERE id = 1`).Scan(
		&result.ScheduleKind, &result.ScheduleTime, &result.ScheduleDay, &result.Timezone, &result.Theme, &result.Language, &excludeJSON)
	if err != nil {
		return result, err
	}
	if err := json.Unmarshal([]byte(excludeJSON), &result.Exclude); err != nil {
		return result, fmt.Errorf("decode exclusions: %w", err)
	}
	return result, nil
}

func (s *Store) UpdateSettings(ctx context.Context, settings model.Settings) error {
	excludeJSON, err := json.Marshal(settings.Exclude)
	if err != nil {
		return err
	}
	_, err = s.db.ExecContext(ctx, `UPDATE settings SET schedule_kind=?, schedule_time=?, schedule_day=?, timezone=?, theme=?, language=?, exclude_json=? WHERE id=1`,
		settings.ScheduleKind, settings.ScheduleTime, settings.ScheduleDay, settings.Timezone, settings.Theme, settings.Language, string(excludeJSON))
	return err
}

func (s *Store) StartRun(ctx context.Context, rootID int64) (int64, int64, error) {
	var previousCount int64
	_ = s.db.QueryRowContext(ctx, `SELECT COALESCE(sr.file_count, 0) FROM roots r LEFT JOIN scan_runs sr ON sr.id=r.current_scan_id WHERE r.id=?`, rootID).Scan(&previousCount)
	result, err := s.db.ExecContext(ctx, `INSERT INTO scan_runs(root_id, status, started_at) VALUES(?, ?, ?)`, rootID, model.ScanScanning, time.Now().UTC().Format(time.RFC3339Nano))
	if err != nil {
		return 0, 0, err
	}
	id, err := result.LastInsertId()
	return id, previousCount, err
}

func (s *Store) InsertEntries(ctx context.Context, rootID, runID int64, entries []Entry) error {
	if len(entries) == 0 {
		return nil
	}
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()
	stmt, err := tx.PrepareContext(ctx, `
INSERT INTO entries(run_id, root_id, path, parent_path, name, kind, category, size, allocated_size, modified_at, identity, recursive_files, recursive_dirs, recursive_size)
VALUES(?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`)
	if err != nil {
		return err
	}
	defer stmt.Close()
	for _, entry := range entries {
		if _, err := stmt.ExecContext(ctx, runID, rootID, entry.Path, entry.ParentPath, entry.Name, entry.Kind, entry.Category, entry.Size, entry.AllocatedSize,
			entry.ModifiedAt.UTC().Format(time.RFC3339Nano), entry.Identity, entry.RecursiveFiles, entry.RecursiveDirs, entry.RecursiveSize); err != nil {
			return err
		}
	}
	return tx.Commit()
}

func (s *Store) InsertScanErrors(ctx context.Context, runID int64, items map[string]string) error {
	if len(items) == 0 {
		return nil
	}
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()
	stmt, err := tx.PrepareContext(ctx, "INSERT INTO scan_errors(run_id, path, message) VALUES(?, ?, ?)")
	if err != nil {
		return err
	}
	defer stmt.Close()
	for path, message := range items {
		if _, err := stmt.ExecContext(ctx, runID, path, message); err != nil {
			return err
		}
	}
	return tx.Commit()
}

func (s *Store) FinishRun(ctx context.Context, rootID, runID int64, summary RunSummary) error {
	now := time.Now().UTC().Format(time.RFC3339Nano)
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()
	if _, err := tx.ExecContext(ctx, `UPDATE scan_runs SET status=?, completed_at=?, file_count=?, directory_count=?, logical_size=?, allocated_size=?, error_count=?, largest_name=?, largest_size=? WHERE id=? AND root_id=?`,
		model.ScanCompleted, now, summary.Files, summary.Directories, summary.LogicalSize, summary.AllocatedSize, summary.Errors, summary.LargestName, summary.LargestSize, runID, rootID); err != nil {
		return err
	}
	var oldScan sql.NullInt64
	if err := tx.QueryRowContext(ctx, "SELECT current_scan_id FROM roots WHERE id=?", rootID).Scan(&oldScan); err != nil {
		return err
	}
	if _, err := tx.ExecContext(ctx, "UPDATE roots SET current_scan_id=? WHERE id=?", runID, rootID); err != nil {
		return err
	}
	if oldScan.Valid && oldScan.Int64 != runID {
		if _, err := tx.ExecContext(ctx, "DELETE FROM entries WHERE run_id=?", oldScan.Int64); err != nil {
			return err
		}
	}
	if err := tx.Commit(); err != nil {
		return err
	}
	return s.compactHistory(ctx, rootID)
}

func (s *Store) FailRun(ctx context.Context, rootID, runID int64, status, message string) error {
	if status != model.ScanCancelled {
		status = model.ScanFailed
	}
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()
	if _, err := tx.ExecContext(ctx, "DELETE FROM entries WHERE run_id=?", runID); err != nil {
		return err
	}
	if _, err := tx.ExecContext(ctx, `UPDATE scan_runs SET status=?, completed_at=?, error_message=? WHERE id=? AND root_id=?`, status, time.Now().UTC().Format(time.RFC3339Nano), message, runID, rootID); err != nil {
		return err
	}
	return tx.Commit()
}

func (s *Store) compactHistory(ctx context.Context, rootID int64) error {
	rows, err := s.db.QueryContext(ctx, `SELECT id, completed_at FROM scan_runs WHERE root_id=? AND status=? AND completed_at IS NOT NULL ORDER BY completed_at DESC`, rootID, model.ScanCompleted)
	if err != nil {
		return err
	}
	var points []retention.Point
	for rows.Next() {
		var point retention.Point
		var raw string
		if err := rows.Scan(&point.ID, &raw); err != nil {
			rows.Close()
			return err
		}
		point.CompletedAt, err = time.Parse(time.RFC3339Nano, raw)
		if err != nil {
			rows.Close()
			return err
		}
		points = append(points, point)
	}
	if err := rows.Close(); err != nil {
		return err
	}
	keep := retention.KeepIDs(time.Now().UTC(), points)
	for _, point := range points {
		if _, ok := keep[point.ID]; ok {
			continue
		}
		if _, err := s.db.ExecContext(ctx, "DELETE FROM scan_runs WHERE id=? AND id NOT IN (SELECT current_scan_id FROM roots WHERE current_scan_id IS NOT NULL)", point.ID); err != nil {
			return err
		}
	}
	return nil
}

func (s *Store) DB() *sql.DB { return s.db }

func (s *Store) ScanErrors(ctx context.Context, rootID int64) ([]model.ScanError, error) {
	rows, err := s.db.QueryContext(ctx, `
SELECT se.path, se.message FROM scan_errors se
JOIN roots r ON r.current_scan_id=se.run_id
WHERE r.id=? ORDER BY se.id LIMIT 1000`, rootID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var result []model.ScanError
	for rows.Next() {
		var item model.ScanError
		if err := rows.Scan(&item.Path, &item.Message); err != nil {
			return nil, err
		}
		result = append(result, item)
	}
	return result, rows.Err()
}
