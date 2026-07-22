import { ChevronRight, Filter, Play, RefreshCw } from "lucide-react";
import type { TranslationKey } from "../lib/i18n";
import type { Root, ScanProgress } from "../types";
import { Button } from "./ui";

export function Topbar({
  root,
  path,
  categories,
  progress,
  t,
  onPathChange,
  onClearCategories,
  onScan,
  onRefresh,
}: {
  root?: Root;
  path: string;
  categories: string[];
  progress?: ScanProgress;
  t: (key: TranslationKey | string) => string;
  onPathChange: (path: string) => void;
  onClearCategories: () => void;
  onScan: () => void;
  onRefresh: () => void;
}) {
  const segments = path ? path.split("/") : [];
  return (
    <header className="topbar glass-panel">
      <nav className="breadcrumbs" aria-label="Breadcrumb">
        <button type="button" onClick={() => onPathChange("")}>
          {root?.name ?? t("allFiles")}
        </button>
        {segments.map((segment, index) => {
          const segmentPath = segments.slice(0, index + 1).join("/");
          return (
            <span key={segmentPath}>
              <ChevronRight size={14} />
              <button type="button" onClick={() => onPathChange(segmentPath)}>
                {segment}
              </button>
            </span>
          );
        })}
      </nav>
      <div className="topbar-actions">
        {categories.length > 0 && (
          <button
            className="filter-pill"
            type="button"
            onClick={onClearCategories}
          >
            <Filter size={14} />
            {categories.length} · {t("clearFilters")}
          </button>
        )}
        {progress &&
          ["scanning", "indexing", "finalizing"].includes(progress.stage) && (
            <span className="scan-live">
              <i />
              {t("scanning")} · {Math.round(progress.estimatedPercent ?? 0)}%
            </span>
          )}
        <Button variant="ghost" aria-label={t("refresh")} onClick={onRefresh}>
          <RefreshCw size={17} />
        </Button>
        <Button
          onClick={onScan}
          disabled={
            !root ||
            Boolean(
              progress &&
                ["scanning", "indexing", "finalizing"].includes(progress.stage),
            )
          }
        >
          <Play size={16} />
          {t("scanNow")}
        </Button>
      </div>
    </header>
  );
}
