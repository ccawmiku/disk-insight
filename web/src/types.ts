export type Page = "overview" | "time" | "files" | "scans" | "settings";
export type Scale = "linear" | "log";

export interface Root {
  id: number;
  name: string;
  enabled: boolean;
  currentScanId?: number;
  lastScanAt?: string;
  lastFileCount: number;
  lastDirectoryCount: number;
  lastLogicalSize: number;
}

export interface Settings {
  scheduleKind: "daily" | "weekly" | "off";
  scheduleTime: string;
  scheduleDay: number;
  timezone: string;
  theme: ThemeName;
  language: Language;
  exclude: string[];
}

export type ThemeName =
  | "tropical-coral"
  | "citrus-sunset"
  | "ocean-glow"
  | "berry-candy";
export type Language = "zh-CN" | "en-US";

export interface ScanProgress {
  rootId: number;
  rootName: string;
  runId: number;
  stage: string;
  currentPath: string;
  files: number;
  directories: number;
  logicalBytes: number;
  errors: number;
  startedAt: string;
  finishedAt?: string;
  filesPerSecond: number;
  estimatedPercent?: number;
  estimatedSeconds?: number;
  error?: string;
}

export interface Dashboard {
  rootId: number;
  path: string;
  generatedAt: string;
  axisMax: number;
  summary: Summary;
  sizeDistribution: SizePoint[];
  ageDistribution: AgePoint[];
  categories: CategoryStat[];
  topFiles: FileItem[];
  children: ChildUsage[];
  history: HistoryPoint[];
}

export interface Summary {
  logicalSize: number;
  allocatedSize?: number;
  fileCount: number;
  directoryCount: number;
  largestFileName?: string;
  largestFileSize: number;
  lastScanDurationMs: number;
  scanErrors: number;
}

export interface SizePoint {
  upper: number;
  count: number;
  bytes: number;
  cumulativeCount: number;
  cumulativeBytes: number;
}

export interface AgePoint {
  upperSeconds: number;
  count: number;
  bytes: number;
}

export interface CategoryStat {
  category: string;
  count: number;
  bytes: number;
}

export interface FileItem {
  name: string;
  path: string;
  category: string;
  size: number;
  modifiedAt: string;
}

export interface ChildUsage {
  name: string;
  path: string;
  kind: "file" | "directory";
  size: number;
  fileCount: number;
}

export interface HistoryPoint {
  completedAt: string;
  fileCount: number;
  logicalSize: number;
}

export interface TreeNode {
  name: string;
  path: string;
  fileCount: number;
  size: number;
  hasChildren: boolean;
}

export interface ScanError {
  path: string;
  message: string;
}
