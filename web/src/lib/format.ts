const units = ["B", "KB", "MB", "GB", "TB", "PB"];

export function formatBytes(value?: number, digits = 1): string {
  if (value === undefined || Number.isNaN(value)) return "—";
  if (value === 0) return "0 B";
  const index = Math.min(
    Math.floor(Math.log(Math.abs(value)) / Math.log(1024)),
    units.length - 1,
  );
  const scaled = value / 1024 ** index;
  return `${new Intl.NumberFormat(undefined, { maximumFractionDigits: digits }).format(scaled)} ${units[index]}`;
}

export function formatNumber(value: number): string {
  return new Intl.NumberFormat().format(Math.round(value));
}

export function formatDuration(seconds: number): string {
  if (seconds < 60) return `${Math.max(0, Math.round(seconds))}s`;
  if (seconds < 3600)
    return `${Math.floor(seconds / 60)}m ${Math.round(seconds % 60)}s`;
  return `${Math.floor(seconds / 3600)}h ${Math.round((seconds % 3600) / 60)}m`;
}

export function formatAge(seconds: number): string {
  if (seconds < 0) return "Future";
  const day = 86400;
  if (seconds < day) return `${Math.max(1, Math.round(seconds / 3600))}h`;
  if (seconds < day * 30) return `${Math.round(seconds / day)}d`;
  if (seconds < day * 365) return `${Math.round(seconds / (day * 30))}mo`;
  return `${(seconds / (day * 365)).toFixed(1)}y`;
}

export function formatDate(value?: string): string {
  if (!value) return "—";
  return new Intl.DateTimeFormat(undefined, {
    dateStyle: "medium",
    timeStyle: "short",
  }).format(new Date(value));
}
