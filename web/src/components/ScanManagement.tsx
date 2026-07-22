import {
  AlertTriangle,
  Ban,
  CheckCircle2,
  FileSearch,
  FolderSearch,
  Gauge,
  Timer,
} from "lucide-react";
import { useEffect, useState } from "react";
import { api } from "../lib/api";
import { formatBytes, formatDuration, formatNumber } from "../lib/format";
import type { TranslationKey } from "../lib/i18n";
import type { Root, ScanError, ScanProgress } from "../types";
import { Button, Card, CardHeader } from "./ui";

export function ScanManagement({
  roots,
  progress,
  selectedRootId,
  t,
  onCancel,
  onScan,
}: {
  roots: Root[];
  progress: ScanProgress[];
  selectedRootId?: number;
  t: (key: TranslationKey | string) => string;
  onCancel: (rootId: number) => void;
  onScan: (rootId: number) => void;
}) {
  const [errors, setErrors] = useState<ScanError[]>([]);
  const selectedStage = progress.find(
    (item) => item.rootId === selectedRootId,
  )?.stage;
  useEffect(() => {
    if (!selectedRootId) return;
    void selectedStage;
    api
      .errors(selectedRootId)
      .then(setErrors)
      .catch(() => setErrors([]));
  }, [selectedRootId, selectedStage]);
  return (
    <div className="page-stack">
      <div className="page-heading">
        <div>
          <span className="eyebrow">Monitor</span>
          <h1>{t("progress")}</h1>
        </div>
      </div>
      <div className="scan-grid">
        {roots.map((root) => {
          const item = progress.find(
            (candidate) => candidate.rootId === root.id,
          );
          return item ? (
            <ProgressCard
              key={root.id}
              progress={item}
              t={t}
              onCancel={() => onCancel(root.id)}
            />
          ) : (
            <Card key={root.id} className="idle-scan-card">
              <CardHeader
                title={root.name}
                description={
                  root.lastScanAt
                    ? new Date(root.lastScanAt).toLocaleString()
                    : t("noData")
                }
              />
              <Button onClick={() => onScan(root.id)}>{t("scanNow")}</Button>
            </Card>
          );
        })}
      </div>
      <Card>
        <CardHeader
          title={t("scanReport")}
          description={
            errors.length ? `${errors.length} ${t("errors")}` : t("noErrors")
          }
        />
        {errors.length > 0 && (
          <div className="error-list">
            {errors.map((error) => (
              <div key={`${error.path}:${error.message}`}>
                <AlertTriangle size={15} />
                <span>
                  <strong>{error.path || "/"}</strong>
                  <small>{error.message}</small>
                </span>
              </div>
            ))}
          </div>
        )}
      </Card>
    </div>
  );
}

function ProgressCard({
  progress,
  t,
  onCancel,
}: {
  progress: ScanProgress;
  t: (key: TranslationKey | string) => string;
  onCancel: () => void;
}) {
  const active = ["scanning", "indexing", "finalizing", "queued"].includes(
    progress.stage,
  );
  const percent =
    progress.stage === "completed" ? 100 : progress.estimatedPercent;
  const elapsed = (Date.now() - new Date(progress.startedAt).getTime()) / 1000;
  const stageLabel = t(progress.stage);
  return (
    <Card className="progress-card">
      <CardHeader
        title={progress.rootName}
        description={stageLabel}
        action={
          progress.stage === "completed" ? (
            <CheckCircle2 className="success" />
          ) : progress.stage === "failed" ? (
            <AlertTriangle className="danger" />
          ) : (
            <span className="pulse-ring" />
          )
        }
      />
      <div
        className="progress-track"
        aria-label={`${stageLabel} ${percent ?? ""}`}
        role="progressbar"
        aria-valuemin={0}
        aria-valuemax={100}
        aria-valuenow={percent ? Math.round(percent) : undefined}
      >
        <span
          className={percent === undefined ? "indeterminate" : ""}
          style={percent === undefined ? undefined : { width: `${percent}%` }}
        />
      </div>
      <div className="progress-caption">
        <strong>
          {percent === undefined ? stageLabel : `${percent.toFixed(1)}%`}
        </strong>
        <span>{progress.currentPath || stageLabel}</span>
      </div>
      <div className="progress-metrics">
        <Metric
          icon={FileSearch}
          label={t("files")}
          value={formatNumber(progress.files)}
        />
        <Metric
          icon={FolderSearch}
          label={t("directoryCount")}
          value={formatNumber(progress.directories)}
        />
        <Metric
          icon={Gauge}
          label={t("speed")}
          value={`${formatNumber(progress.filesPerSecond)}/s`}
        />
        <Metric
          icon={Timer}
          label={t("elapsed")}
          value={formatDuration(elapsed)}
        />
        <Metric
          icon={Timer}
          label={t("remaining")}
          value={
            progress.estimatedSeconds === undefined
              ? "—"
              : formatDuration(progress.estimatedSeconds)
          }
        />
        <Metric
          icon={AlertTriangle}
          label={t("errors")}
          value={formatNumber(progress.errors)}
        />
      </div>
      <div className="scan-bytes">
        {formatBytes(progress.logicalBytes)} ·{" "}
        <span title={progress.currentPath}>
          {t("currentPath")}: {progress.currentPath || "/"}
        </span>
      </div>
      {active && (
        <Button variant="danger" onClick={onCancel}>
          <Ban size={15} />
          {t("cancel")}
        </Button>
      )}
    </Card>
  );
}

function Metric({
  icon: Icon,
  label,
  value,
}: {
  icon: typeof Timer;
  label: string;
  value: string;
}) {
  return (
    <div>
      <Icon size={15} />
      <span>{label}</span>
      <strong>{value}</strong>
    </div>
  );
}
