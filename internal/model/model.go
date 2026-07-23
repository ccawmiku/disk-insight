package model

import "time"

const (
	ScanQueued     = "queued"
	ScanScanning   = "scanning"
	ScanIndexing   = "indexing"
	ScanFinalizing = "finalizing"
	ScanCompleted  = "completed"
	ScanFailed     = "failed"
	ScanCancelled  = "cancelled"
)

type RootConfig struct {
	ID      int64  `json:"id"`
	Name    string `json:"name"`
	Path    string `json:"-"`
	Enabled bool   `json:"enabled"`
}

type Root struct {
	ID                 int64      `json:"id"`
	Name               string     `json:"name"`
	Enabled            bool       `json:"enabled"`
	CurrentScanID      *int64     `json:"currentScanId,omitempty"`
	LastScanAt         *time.Time `json:"lastScanAt,omitempty"`
	LastFileCount      int64      `json:"lastFileCount"`
	LastDirectoryCount int64      `json:"lastDirectoryCount"`
	LastLogicalSize    int64      `json:"lastLogicalSize"`
}

type Settings struct {
	ScheduleKind string   `json:"scheduleKind"`
	ScheduleTime string   `json:"scheduleTime"`
	ScheduleDay  int      `json:"scheduleDay"`
	Timezone     string   `json:"timezone"`
	Theme        string   `json:"theme"`
	Language     string   `json:"language"`
	Exclude      []string `json:"exclude"`
}

type ScanProgress struct {
	RootID           int64      `json:"rootId"`
	RootName         string     `json:"rootName"`
	RunID            int64      `json:"runId"`
	Stage            string     `json:"stage"`
	CurrentPath      string     `json:"currentPath"`
	Files            int64      `json:"files"`
	Directories      int64      `json:"directories"`
	LogicalBytes     int64      `json:"logicalBytes"`
	Errors           int64      `json:"errors"`
	StartedAt        time.Time  `json:"startedAt"`
	FinishedAt       *time.Time `json:"finishedAt,omitempty"`
	FilesPerSecond   float64    `json:"filesPerSecond"`
	EstimatedPercent *float64   `json:"estimatedPercent,omitempty"`
	EstimatedSeconds *int64     `json:"estimatedSeconds,omitempty"`
	Error            string     `json:"error,omitempty"`
}

type Summary struct {
	LogicalSize      int64  `json:"logicalSize"`
	AllocatedSize    *int64 `json:"allocatedSize,omitempty"`
	FileCount        int64  `json:"fileCount"`
	DirectoryCount   int64  `json:"directoryCount"`
	LargestFileName  string `json:"largestFileName,omitempty"`
	LargestFileSize  int64  `json:"largestFileSize"`
	LastScanDuration int64  `json:"lastScanDurationMs"`
	ScanErrors       int64  `json:"scanErrors"`
}

type SizePoint struct {
	Upper           int64   `json:"upper"`
	Count           int64   `json:"count"`
	Bytes           int64   `json:"bytes"`
	CumulativeCount float64 `json:"cumulativeCount"`
	CumulativeBytes float64 `json:"cumulativeBytes"`
}

type AgePoint struct {
	UpperSeconds int64 `json:"upperSeconds"`
	Count        int64 `json:"count"`
	Bytes        int64 `json:"bytes"`
}

type CategoryStat struct {
	Category string `json:"category"`
	Count    int64  `json:"count"`
	Bytes    int64  `json:"bytes"`
}

type FileItem struct {
	Name       string    `json:"name"`
	Path       string    `json:"path"`
	Category   string    `json:"category"`
	Size       int64     `json:"size"`
	ModifiedAt time.Time `json:"modifiedAt"`
}

type ChildUsage struct {
	Name      string `json:"name"`
	Path      string `json:"path"`
	Kind      string `json:"kind"`
	Size      int64  `json:"size"`
	FileCount int64  `json:"fileCount"`
}

type HistoryPoint struct {
	CompletedAt time.Time `json:"completedAt"`
	FileCount   int64     `json:"fileCount"`
	LogicalSize int64     `json:"logicalSize"`
}

type Dashboard struct {
	RootID      int64          `json:"rootId"`
	Path        string         `json:"path"`
	GeneratedAt time.Time      `json:"generatedAt"`
	AxisMax     int64          `json:"axisMax"`
	Summary     Summary        `json:"summary"`
	Size        []SizePoint    `json:"sizeDistribution"`
	Age         []AgePoint     `json:"ageDistribution"`
	Categories  []CategoryStat `json:"categories"`
	TopFiles    []FileItem     `json:"topFiles"`
	Children    []ChildUsage   `json:"children"`
	History     []HistoryPoint `json:"history"`
}

type TreeNode struct {
	Name        string `json:"name"`
	Path        string `json:"path"`
	FileCount   int64  `json:"fileCount"`
	Size        int64  `json:"size"`
	HasChildren bool   `json:"hasChildren"`
}

type ScanError struct {
	Path    string `json:"path"`
	Message string `json:"message"`
}
