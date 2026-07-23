package scanner

import (
	"context"
	"errors"
	"fmt"
	"io/fs"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/ccawmiku/disk-insight/internal/classify"
	"github.com/ccawmiku/disk-insight/internal/model"
	"github.com/ccawmiku/disk-insight/internal/store"
)

var ErrAlreadyRunning = errors.New("scan already running")

type Manager struct {
	store      *store.Store
	mu         sync.RWMutex
	progress   map[int64]model.ScanProgress
	cancel     map[int64]context.CancelFunc
	onComplete func(int64)
}

type directoryRecord struct {
	entry     store.Entry
	aggregate aggregate
}

type aggregate struct {
	files int64
	dirs  int64
	size  int64
}

func New(managerStore *store.Store, onComplete func(int64)) *Manager {
	return &Manager{store: managerStore, progress: make(map[int64]model.ScanProgress), cancel: make(map[int64]context.CancelFunc), onComplete: onComplete}
}

func (m *Manager) Start(root model.RootConfig, exclusions []string) error {
	m.mu.Lock()
	if _, exists := m.cancel[root.ID]; exists {
		m.mu.Unlock()
		return ErrAlreadyRunning
	}
	ctx, cancel := context.WithCancel(context.Background())
	m.cancel[root.ID] = cancel
	m.mu.Unlock()
	go func() {
		_ = m.run(ctx, root, exclusions)
		m.mu.Lock()
		delete(m.cancel, root.ID)
		m.mu.Unlock()
	}()
	return nil
}

func (m *Manager) Run(ctx context.Context, root model.RootConfig, exclusions []string) error {
	return m.run(ctx, root, exclusions)
}

func (m *Manager) Cancel(rootID int64) bool {
	m.mu.RLock()
	cancel, exists := m.cancel[rootID]
	m.mu.RUnlock()
	if exists {
		cancel()
	}
	return exists
}

func (m *Manager) Progress() []model.ScanProgress {
	m.mu.RLock()
	defer m.mu.RUnlock()
	result := make([]model.ScanProgress, 0, len(m.progress))
	for _, progress := range m.progress {
		result = append(result, progress)
	}
	sort.Slice(result, func(i, j int) bool { return result[i].RootName < result[j].RootName })
	return result
}

func (m *Manager) run(ctx context.Context, root model.RootConfig, exclusions []string) (runErr error) {
	runID, previousCount, err := m.store.StartRun(ctx, root.ID)
	if err != nil {
		return err
	}
	started := time.Now()
	progress := model.ScanProgress{RootID: root.ID, RootName: root.Name, RunID: runID, Stage: model.ScanScanning, StartedAt: started.UTC()}
	lastProgressUpdate := time.Time{}
	publishProgress := func(force bool) {
		if !force && time.Since(lastProgressUpdate) < 350*time.Millisecond {
			return
		}
		m.setProgress(progress, previousCount)
		lastProgressUpdate = time.Now()
	}
	publishProgress(true)
	completed := false
	defer func() {
		if completed {
			return
		}
		status := model.ScanFailed
		message := "scan failed"
		if errors.Is(runErr, context.Canceled) || errors.Is(ctx.Err(), context.Canceled) {
			status = model.ScanCancelled
			message = "scan cancelled"
		} else if runErr != nil {
			message = runErr.Error()
		}
		_ = m.store.FailRun(context.Background(), root.ID, runID, status, message)
		finished := time.Now().UTC()
		progress.Stage = status
		progress.Error = message
		progress.FinishedAt = &finished
		publishProgress(true)
	}()

	rootInfo, err := osStat(root.Path)
	if err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			return fmt.Errorf("scan root %q is not mounted or does not exist: %w", root.Path, err)
		}
		if errors.Is(err, fs.ErrPermission) {
			return fmt.Errorf("scan root %q is not readable by the container user: %w", root.Path, err)
		}
		return fmt.Errorf("open scan root: %w", err)
	}
	if !rootInfo.IsDir() {
		return fmt.Errorf("scan root is not a directory")
	}

	directories := make(map[string]*directoryRecord, 1024)
	files := make([]store.Entry, 0, 512)
	scanErrors := make(map[string]string)
	var largestName string
	var largestSize int64

	flushFiles := func() error {
		if err := m.store.InsertEntries(ctx, root.ID, runID, files); err != nil {
			return err
		}
		files = files[:0]
		return nil
	}

	err = filepath.WalkDir(root.Path, func(path string, item fs.DirEntry, walkErr error) error {
		if ctx.Err() != nil {
			return ctx.Err()
		}
		relative, relErr := filepath.Rel(root.Path, path)
		if relErr != nil {
			return relErr
		}
		if relative == "." {
			relative = ""
		}
		relative = filepath.ToSlash(relative)
		progress.CurrentPath = relative
		if walkErr != nil {
			progress.Errors++
			if len(scanErrors) < 1000 {
				scanErrors[relative] = walkErr.Error()
			}
			publishProgress(false)
			if item != nil && item.IsDir() {
				return fs.SkipDir
			}
			return nil
		}
		if item.Type()&osModeSymlink() != 0 {
			if item.IsDir() {
				return fs.SkipDir
			}
			return nil
		}
		info, infoErr := item.Info()
		if infoErr != nil {
			progress.Errors++
			if len(scanErrors) < 1000 {
				scanErrors[relative] = infoErr.Error()
			}
			publishProgress(false)
			if item.IsDir() {
				return fs.SkipDir
			}
			return nil
		}
		if relative != "" && (platformHidden(path, item.Name(), info) || excluded(item.Name(), relative, exclusions)) {
			if item.IsDir() {
				return fs.SkipDir
			}
			return nil
		}
		parent := parentPath(relative)
		if item.IsDir() {
			record := ensureDirectory(directories, relative)
			record.entry = store.Entry{Path: relative, ParentPath: parent, Name: displayName(relative, root.Name), Kind: "directory", ModifiedAt: info.ModTime()}
			if relative != "" {
				progress.Directories++
				addDirectoryToAncestors(directories, parent)
			}
			publishProgress(false)
			return nil
		}
		if !info.Mode().IsRegular() {
			return nil
		}
		allocated, identity := platformMetadata(path, info)
		files = append(files, store.Entry{Path: relative, ParentPath: parent, Name: item.Name(), Kind: "file", Category: classify.Category(item.Name()), Size: info.Size(), AllocatedSize: allocated, ModifiedAt: info.ModTime(), Identity: identity, RecursiveFiles: 1, RecursiveSize: info.Size()})
		progress.Files++
		progress.LogicalBytes += info.Size()
		if info.Size() > largestSize {
			largestSize = info.Size()
			largestName = relative
		}
		addFileToAncestors(directories, parent, info.Size())
		if len(files) >= 512 {
			if err := flushFiles(); err != nil {
				return err
			}
			time.Sleep(2 * time.Millisecond)
			publishProgress(true)
		}
		publishProgress(false)
		return nil
	})
	if err != nil {
		return err
	}
	if err := flushFiles(); err != nil {
		return err
	}
	if err := m.store.InsertScanErrors(ctx, runID, scanErrors); err != nil {
		return err
	}

	progress.Stage = model.ScanIndexing
	publishProgress(true)
	directoryEntries := make([]store.Entry, 0, len(directories))
	for _, record := range directories {
		record.entry.RecursiveFiles = record.aggregate.files
		record.entry.RecursiveDirs = record.aggregate.dirs
		record.entry.RecursiveSize = record.aggregate.size
		directoryEntries = append(directoryEntries, record.entry)
		if len(directoryEntries) >= 512 {
			if err := m.store.InsertEntries(ctx, root.ID, runID, directoryEntries); err != nil {
				return err
			}
			directoryEntries = directoryEntries[:0]
		}
	}
	if err := m.store.InsertEntries(ctx, root.ID, runID, directoryEntries); err != nil {
		return err
	}

	progress.Stage = model.ScanFinalizing
	publishProgress(true)
	allocatedSummary, err := m.store.AllocatedSizeForRun(ctx, root.ID, runID)
	if err != nil {
		return err
	}
	if err := m.store.FinishRun(ctx, root.ID, runID, store.RunSummary{Files: progress.Files, Directories: progress.Directories, LogicalSize: progress.LogicalBytes, AllocatedSize: allocatedSummary, Errors: progress.Errors, LargestName: largestName, LargestSize: largestSize}); err != nil {
		return err
	}
	completed = true
	finished := time.Now().UTC()
	progress.Stage = model.ScanCompleted
	progress.CurrentPath = ""
	progress.FinishedAt = &finished
	percent := 100.0
	progress.EstimatedPercent = &percent
	progress.EstimatedSeconds = int64Pointer(0)
	publishProgress(true)
	if m.onComplete != nil {
		m.onComplete(root.ID)
	}
	return nil
}

func (m *Manager) setProgress(progress model.ScanProgress, previousCount int64) {
	elapsed := time.Since(progress.StartedAt)
	if elapsed > 0 {
		progress.FilesPerSecond = float64(progress.Files) / elapsed.Seconds()
	}
	if previousCount > 0 && progress.Stage != model.ScanCompleted {
		percent := float64(progress.Files) / float64(previousCount) * 100
		if percent > 99 {
			percent = 99
		}
		progress.EstimatedPercent = &percent
		if progress.FilesPerSecond > 0 && progress.Files < previousCount {
			seconds := int64(float64(previousCount-progress.Files) / progress.FilesPerSecond)
			progress.EstimatedSeconds = &seconds
		}
	}
	m.mu.Lock()
	m.progress[progress.RootID] = progress
	m.mu.Unlock()
}

func excluded(name, relative string, patterns []string) bool {
	for _, pattern := range patterns {
		pattern = strings.TrimSpace(filepath.ToSlash(pattern))
		if pattern == "" {
			continue
		}
		if strings.EqualFold(name, pattern) || strings.EqualFold(relative, pattern) {
			return true
		}
		if matched, _ := filepath.Match(pattern, name); matched {
			return true
		}
		if matched, _ := filepath.Match(pattern, relative); matched {
			return true
		}
	}
	return false
}

func addFileToAncestors(values map[string]*directoryRecord, parent string, size int64) {
	for {
		record := ensureDirectory(values, parent)
		record.aggregate.files++
		record.aggregate.size += size
		if parent == "" {
			return
		}
		parent = parentPath(parent)
	}
}

func addDirectoryToAncestors(values map[string]*directoryRecord, parent string) {
	for {
		ensureDirectory(values, parent).aggregate.dirs++
		if parent == "" {
			return
		}
		parent = parentPath(parent)
	}
}

func ensureDirectory(values map[string]*directoryRecord, path string) *directoryRecord {
	if values[path] == nil {
		values[path] = &directoryRecord{}
	}
	return values[path]
}

func parentPath(path string) string {
	if path == "" {
		return ""
	}
	parent := filepath.ToSlash(filepath.Dir(filepath.FromSlash(path)))
	if parent == "." {
		return ""
	}
	return parent
}

func displayName(path, rootName string) string {
	if path == "" {
		return rootName
	}
	return filepath.Base(filepath.FromSlash(path))
}

func int64Pointer(value int64) *int64 { return &value }
