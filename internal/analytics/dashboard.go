package analytics

import (
	"container/heap"
	"context"
	"database/sql"
	"errors"
	"fmt"
	"math"
	"path"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/ccawmiku/disk-insight/internal/model"
	"github.com/ccawmiku/disk-insight/internal/store"
)

var ErrNoScan = errors.New("no completed scan is available")

type DashboardService struct {
	store *store.Store
	mu    sync.RWMutex
	cache map[string]model.Dashboard
	order []string
}

type fileSample struct {
	size int64
	age  int64
}

func NewDashboardService(dataStore *store.Store) *DashboardService {
	return &DashboardService{store: dataStore, cache: make(map[string]model.Dashboard)}
}

func (s *DashboardService) Invalidate(rootID int64) {
	prefix := fmt.Sprintf("%d:", rootID)
	s.mu.Lock()
	defer s.mu.Unlock()
	for key := range s.cache {
		if strings.HasPrefix(key, prefix) {
			delete(s.cache, key)
		}
	}
	filtered := s.order[:0]
	for _, key := range s.order {
		if !strings.HasPrefix(key, prefix) {
			filtered = append(filtered, key)
		}
	}
	s.order = filtered
}

func (s *DashboardService) Build(ctx context.Context, rootID int64, selectedPath string, categories []string, sizeScale, ageScale string) (model.Dashboard, error) {
	selectedPath, err := cleanRelativePath(selectedPath)
	if err != nil {
		return model.Dashboard{}, err
	}
	if sizeScale != "log" {
		sizeScale = "linear"
	}
	if ageScale != "log" {
		ageScale = "linear"
	}
	sort.Strings(categories)
	var runID int64
	if err := s.store.DB().QueryRowContext(ctx, "SELECT current_scan_id FROM roots WHERE id=? AND enabled=1", rootID).Scan(&runID); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return model.Dashboard{}, ErrNoScan
		}
		return model.Dashboard{}, err
	}
	cacheKey := fmt.Sprintf("%d:%d:%s:%s:%s:%s", rootID, runID, selectedPath, strings.Join(categories, ","), sizeScale, ageScale)
	s.mu.RLock()
	if cached, ok := s.cache[cacheKey]; ok {
		s.mu.RUnlock()
		return cached, nil
	}
	s.mu.RUnlock()

	query, args := fileQuery(rootID, runID, selectedPath, categories)
	rows, err := s.store.DB().QueryContext(ctx, query, args...)
	if err != nil {
		return model.Dashboard{}, err
	}
	defer rows.Close()

	now := time.Now().UTC()
	result := model.Dashboard{RootID: rootID, Path: selectedPath, GeneratedAt: now, TopFiles: make([]model.FileItem, 0, 100)}
	categoryTotals := make(map[string]*model.CategoryStat)
	childTotals := make(map[string]*model.ChildUsage)
	directories := make(map[string]struct{})
	physicalSeen := make(map[string]struct{})
	var allocatedTotal int64
	var hasAllocated bool
	var samples []fileSample
	top := &fileHeap{}
	heap.Init(top)
	var maxSize int64
	for rows.Next() {
		var name, filePath, category, modifiedRaw, identity string
		var size int64
		var allocated sql.NullInt64
		if err := rows.Scan(&name, &filePath, &category, &size, &allocated, &modifiedRaw, &identity); err != nil {
			return model.Dashboard{}, err
		}
		modified, err := time.Parse(time.RFC3339Nano, modifiedRaw)
		if err != nil {
			return model.Dashboard{}, err
		}
		age := int64(now.Sub(modified).Seconds())
		result.Summary.FileCount++
		result.Summary.LogicalSize += size
		if size > maxSize {
			maxSize = size
			result.Summary.LargestFileName = filePath
			result.Summary.LargestFileSize = size
		}
		if allocated.Valid {
			hasAllocated = true
			if identity == "" {
				allocatedTotal += allocated.Int64
			} else if _, exists := physicalSeen[identity]; !exists {
				physicalSeen[identity] = struct{}{}
				allocatedTotal += allocated.Int64
			}
		}
		samples = append(samples, fileSample{size: size, age: age})
		stat := categoryTotals[category]
		if stat == nil {
			stat = &model.CategoryStat{Category: category}
			categoryTotals[category] = stat
		}
		stat.Count++
		stat.Bytes += size
		addDirectoryPaths(directories, selectedPath, filePath)
		addChildUsage(childTotals, selectedPath, filePath, size)
		item := model.FileItem{Name: name, Path: filePath, Category: category, Size: size, ModifiedAt: modified}
		if top.Len() < 100 {
			heap.Push(top, item)
		} else if (*top)[0].Size < size {
			heap.Pop(top)
			heap.Push(top, item)
		}
	}
	if err := rows.Err(); err != nil {
		return model.Dashboard{}, err
	}
	if hasAllocated {
		result.Summary.AllocatedSize = &allocatedTotal
	}
	result.Summary.DirectoryCount = int64(len(directories))
	result.AxisMax = NiceAxisMax(maxSize)
	result.Size = buildSizePoints(samples, result.AxisMax, sizeScale, result.Summary.FileCount, result.Summary.LogicalSize)
	result.Age = buildAgePoints(samples, ageScale)
	for _, stat := range categoryTotals {
		result.Categories = append(result.Categories, *stat)
	}
	sort.Slice(result.Categories, func(i, j int) bool { return result.Categories[i].Bytes > result.Categories[j].Bytes })
	for _, child := range childTotals {
		result.Children = append(result.Children, *child)
	}
	sort.Slice(result.Children, func(i, j int) bool { return result.Children[i].Size > result.Children[j].Size })
	for top.Len() > 0 {
		result.TopFiles = append(result.TopFiles, heap.Pop(top).(model.FileItem))
	}
	sort.Slice(result.TopFiles, func(i, j int) bool { return result.TopFiles[i].Size > result.TopFiles[j].Size })
	if err := s.addRunMetadata(ctx, &result, rootID, runID); err != nil {
		return model.Dashboard{}, err
	}
	if err := s.addHistory(ctx, &result, rootID); err != nil {
		return model.Dashboard{}, err
	}
	s.putCache(cacheKey, result)
	return result, nil
}

func (s *DashboardService) Tree(ctx context.Context, rootID int64, selectedPath string) ([]model.TreeNode, error) {
	selectedPath, err := cleanRelativePath(selectedPath)
	if err != nil {
		return nil, err
	}
	rows, err := s.store.DB().QueryContext(ctx, `
SELECT e.name, e.path, e.recursive_files, e.recursive_size,
       EXISTS(SELECT 1 FROM entries c WHERE c.root_id=e.root_id AND c.run_id=e.run_id AND c.parent_path=e.path AND c.kind='directory')
FROM entries e JOIN roots r ON r.id=e.root_id AND r.current_scan_id=e.run_id
WHERE e.root_id=? AND e.parent_path=? AND e.path<>? AND e.kind='directory'
ORDER BY e.name COLLATE NOCASE`, rootID, selectedPath, selectedPath)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var result []model.TreeNode
	for rows.Next() {
		var node model.TreeNode
		if err := rows.Scan(&node.Name, &node.Path, &node.FileCount, &node.Size, &node.HasChildren); err != nil {
			return nil, err
		}
		result = append(result, node)
	}
	return result, rows.Err()
}

func (s *DashboardService) addRunMetadata(ctx context.Context, result *model.Dashboard, rootID, runID int64) error {
	var startedRaw, completedRaw string
	if err := s.store.DB().QueryRowContext(ctx, `SELECT started_at, completed_at, error_count FROM scan_runs WHERE id=? AND root_id=?`, runID, rootID).Scan(&startedRaw, &completedRaw, &result.Summary.ScanErrors); err != nil {
		return err
	}
	started, err := time.Parse(time.RFC3339Nano, startedRaw)
	if err != nil {
		return err
	}
	completed, err := time.Parse(time.RFC3339Nano, completedRaw)
	if err != nil {
		return err
	}
	result.Summary.LastScanDuration = completed.Sub(started).Milliseconds()
	return nil
}

func (s *DashboardService) addHistory(ctx context.Context, result *model.Dashboard, rootID int64) error {
	rows, err := s.store.DB().QueryContext(ctx, `SELECT completed_at, file_count, logical_size FROM scan_runs WHERE root_id=? AND status=? AND completed_at IS NOT NULL ORDER BY completed_at`, rootID, model.ScanCompleted)
	if err != nil {
		return err
	}
	defer rows.Close()
	for rows.Next() {
		var point model.HistoryPoint
		var raw string
		if err := rows.Scan(&raw, &point.FileCount, &point.LogicalSize); err != nil {
			return err
		}
		point.CompletedAt, err = time.Parse(time.RFC3339Nano, raw)
		if err != nil {
			return err
		}
		result.History = append(result.History, point)
	}
	return rows.Err()
}

func (s *DashboardService) putCache(key string, value model.Dashboard) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if len(s.order) >= 128 {
		delete(s.cache, s.order[0])
		s.order = s.order[1:]
	}
	s.cache[key] = value
	s.order = append(s.order, key)
}

func fileQuery(rootID, runID int64, selectedPath string, categories []string) (string, []any) {
	query := `SELECT name, path, category, size, allocated_size, modified_at, identity FROM entries WHERE root_id=? AND run_id=? AND kind='file'`
	args := []any{rootID, runID}
	if selectedPath != "" {
		query += " AND path LIKE ? ESCAPE '\\'"
		args = append(args, escapeLike(selectedPath)+"/%")
	}
	if len(categories) > 0 {
		query += " AND category IN (" + strings.TrimSuffix(strings.Repeat("?,", len(categories)), ",") + ")"
		for _, category := range categories {
			args = append(args, category)
		}
	}
	return query, args
}

func escapeLike(value string) string {
	value = strings.ReplaceAll(value, "\\", "\\\\")
	value = strings.ReplaceAll(value, "%", "\\%")
	return strings.ReplaceAll(value, "_", "\\_")
}

func cleanRelativePath(value string) (string, error) {
	value = strings.Trim(strings.ReplaceAll(value, "\\", "/"), "/")
	if value == "" {
		return "", nil
	}
	cleaned := path.Clean(value)
	if cleaned == "." || cleaned == ".." || strings.HasPrefix(cleaned, "../") || strings.HasPrefix(cleaned, "/") {
		return "", fmt.Errorf("invalid relative path")
	}
	return cleaned, nil
}

func addDirectoryPaths(values map[string]struct{}, selectedPath, filePath string) {
	parent := path.Dir(filePath)
	if parent == "." {
		parent = ""
	}
	for parent != "" && parent != selectedPath {
		values[parent] = struct{}{}
		parent = path.Dir(parent)
		if parent == "." {
			parent = ""
		}
	}
}

func addChildUsage(values map[string]*model.ChildUsage, selectedPath, filePath string, size int64) {
	relative := filePath
	if selectedPath != "" {
		relative = strings.TrimPrefix(filePath, selectedPath+"/")
	}
	parts := strings.Split(relative, "/")
	if len(parts) == 0 || parts[0] == "" {
		return
	}
	childPath := parts[0]
	kind := "file"
	if selectedPath != "" {
		childPath = selectedPath + "/" + childPath
	}
	if len(parts) > 1 {
		kind = "directory"
	}
	item := values[childPath]
	if item == nil {
		item = &model.ChildUsage{Name: parts[0], Path: childPath, Kind: kind}
		values[childPath] = item
	}
	item.Size += size
	item.FileCount++
}

func buildSizePoints(samples []fileSample, axisMax int64, scale string, totalCount, totalBytes int64) []model.SizePoint {
	const bins = 80
	result := make([]model.SizePoint, bins)
	for index := range result {
		result[index].Upper = binUpper(index, bins, axisMax, scale)
	}
	for _, sample := range samples {
		index := binIndex(sample.size, bins, axisMax, scale)
		result[index].Count++
		result[index].Bytes += sample.size
	}
	var count, bytes int64
	for index := range result {
		count += result[index].Count
		bytes += result[index].Bytes
		if totalCount > 0 {
			result[index].CumulativeCount = float64(count) / float64(totalCount) * 100
		}
		if totalBytes > 0 {
			result[index].CumulativeBytes = float64(bytes) / float64(totalBytes) * 100
		}
	}
	return result
}

func buildAgePoints(samples []fileSample, scale string) []model.AgePoint {
	const bins = 60
	var maximum int64 = 1
	var future int64
	for _, sample := range samples {
		if sample.age < 0 {
			future++
		} else if sample.age > maximum {
			maximum = sample.age
		}
	}
	result := make([]model.AgePoint, bins, bins+1)
	for index := range result {
		result[index].UpperSeconds = binUpper(index, bins, maximum, scale)
	}
	for _, sample := range samples {
		if sample.age < 0 {
			continue
		}
		result[binIndex(sample.age, bins, maximum, scale)].Count++
	}
	if future > 0 {
		result = append([]model.AgePoint{{UpperSeconds: -1, Count: future}}, result...)
	}
	return result
}

func binIndex(value int64, bins int, maximum int64, scale string) int {
	if maximum <= 0 || value <= 0 {
		return 0
	}
	var position float64
	if scale == "log" {
		position = math.Log1p(float64(value)) / math.Log1p(float64(maximum))
	} else {
		position = float64(value) / float64(maximum)
	}
	index := int(position * float64(bins))
	if index >= bins {
		return bins - 1
	}
	if index < 0 {
		return 0
	}
	return index
}

func binUpper(index, bins int, maximum int64, scale string) int64 {
	fraction := float64(index+1) / float64(bins)
	if scale == "log" {
		return int64(math.Expm1(fraction * math.Log1p(float64(maximum))))
	}
	return int64(math.Ceil(fraction * float64(maximum)))
}

type fileHeap []model.FileItem

func (h fileHeap) Len() int           { return len(h) }
func (h fileHeap) Less(i, j int) bool { return h[i].Size < h[j].Size }
func (h fileHeap) Swap(i, j int)      { h[i], h[j] = h[j], h[i] }
func (h *fileHeap) Push(value any)    { *h = append(*h, value.(model.FileItem)) }
func (h *fileHeap) Pop() any {
	old := *h
	last := old[len(old)-1]
	*h = old[:len(old)-1]
	return last
}
