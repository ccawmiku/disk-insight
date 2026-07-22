import {
  Clock,
  DatabaseZap,
  File,
  Files,
  FolderTree,
  HardDrive,
} from "lucide-react";
import { formatBytes, formatDuration, formatNumber } from "../lib/format";
import type { TranslationKey } from "../lib/i18n";
import type { Root, Summary } from "../types";

export function SummaryCards({
  summary,
  root,
  t,
}: {
  summary: Summary;
  root?: Root;
  t: (key: TranslationKey | string) => string;
}) {
  const items = [
    {
      label: t("logicalSize"),
      value: formatBytes(summary.logicalSize),
      detail: summary.largestFileName
        ? `${t("largestFile")}: ${formatBytes(summary.largestFileSize)}`
        : "",
      icon: HardDrive,
      accent: "coral",
    },
    {
      label: t("allocatedSize"),
      value:
        summary.allocatedSize === undefined
          ? "—"
          : formatBytes(summary.allocatedSize),
      detail:
        summary.allocatedSize === undefined
          ? "Not reported by filesystem"
          : `${Math.round((summary.allocatedSize / Math.max(summary.logicalSize, 1)) * 100)}% logical`,
      icon: DatabaseZap,
      accent: "orange",
    },
    {
      label: t("fileCount"),
      value: formatNumber(summary.fileCount),
      detail: `${formatBytes(summary.fileCount ? summary.logicalSize / summary.fileCount : 0)} ${t("averageSize")}`,
      icon: Files,
      accent: "yellow",
    },
    {
      label: t("directoryCount"),
      value: formatNumber(summary.directoryCount),
      detail: root?.name ?? "",
      icon: FolderTree,
      accent: "pink",
    },
    {
      label: t("largestFile"),
      value: formatBytes(summary.largestFileSize),
      detail: summary.largestFileName ?? "—",
      icon: File,
      accent: "purple",
    },
    {
      label: t("lastScan"),
      value: formatDuration(summary.lastScanDurationMs / 1000),
      detail: summary.scanErrors
        ? `${summary.scanErrors} ${t("errors")}`
        : root?.lastScanAt
          ? new Date(root.lastScanAt).toLocaleString()
          : "—",
      icon: Clock,
      accent: "blue",
    },
  ];
  return (
    <section className="summary-grid" aria-label="Summary">
      {items.map(({ label, value, detail, icon: Icon, accent }) => (
        <article
          className={`summary-card glass-panel accent-${accent}`}
          key={label}
        >
          <div className="summary-icon">
            <Icon size={19} />
          </div>
          <div>
            <span>{label}</span>
            <strong>{value}</strong>
            <small title={detail}>{detail}</small>
          </div>
        </article>
      ))}
    </section>
  );
}
