import { LockKeyhole, UnlockKeyhole } from "lucide-react";
import { useMemo, useState } from "react";
import { formatBytes, formatNumber } from "../lib/format";
import type { TranslationKey } from "../lib/i18n";
import type { Scale, SizePoint, Summary } from "../types";
import ReactECharts from "./EChart";
import { Button, Card, CardHeader, Segmented } from "./ui";

interface ZoomRange {
  start: number;
  end: number;
}

export function SizeAnalytics({
  points,
  summary,
  scale,
  onScaleChange,
  t,
}: {
  points: SizePoint[];
  summary: Summary;
  scale: Scale;
  onScaleChange: (scale: Scale) => void;
  t: (key: TranslationKey | string) => string;
}) {
  const [hovered, setHovered] = useState<number | null>(null);
  const [locked, setLocked] = useState(false);
  const [zoom, setZoom] = useState<ZoomRange>({ start: 0, end: 100 });
  const thresholdIndex =
    hovered ??
    Math.max(
      0,
      points.findIndex((point) => point.cumulativeCount >= 50),
    );
  const threshold = points[thresholdIndex]?.upper ?? 0;
  const split = useMemo(() => {
    const left = points.slice(0, thresholdIndex + 1).reduce(
      (value, point) => ({
        count: value.count + point.count,
        bytes: value.bytes + point.bytes,
      }),
      { count: 0, bytes: 0 },
    );
    return {
      left,
      right: {
        count: summary.fileCount - left.count,
        bytes: summary.logicalSize - left.bytes,
      },
    };
  }, [points, thresholdIndex, summary]);

  const updateHover = (params: { dataIndex?: number }) => {
    if (!locked && typeof params.dataIndex === "number")
      setHovered(params.dataIndex);
  };
  const lockAt = (params: { dataIndex?: number }) => {
    if (typeof params.dataIndex === "number") setHovered(params.dataIndex);
    setLocked(true);
  };
  const updateZoom = (params: {
    start?: number;
    end?: number;
    batch?: Array<{ start: number; end: number }>;
  }) => {
    const next = params.batch?.[0] ?? params;
    if (typeof next.start === "number" && typeof next.end === "number")
      setZoom({ start: next.start, end: next.end });
  };
  const commonEvents = {
    mouseover: updateHover,
    click: lockAt,
    datazoom: updateZoom,
  };

  return (
    <div className="size-analytics">
      <div className="threshold-panel glass-panel">
        <SplitStat
          label={t("leftOfThreshold")}
          threshold={`≤ ${formatBytes(threshold)}`}
          count={split.left.count}
          bytes={split.left.bytes}
          totalCount={summary.fileCount}
          totalBytes={summary.logicalSize}
          tone="left"
          t={t}
        />
        <div className="threshold-center">
          <span>{t("lockThreshold")}</span>
          {locked && (
            <Button variant="ghost" onClick={() => setLocked(false)}>
              <UnlockKeyhole size={15} />
              {t("unlock")}
            </Button>
          )}
        </div>
        <SplitStat
          label={t("rightOfThreshold")}
          threshold={`> ${formatBytes(threshold)}`}
          count={split.right.count}
          bytes={split.right.bytes}
          totalCount={summary.fileCount}
          totalBytes={summary.logicalSize}
          tone="right"
          t={t}
        />
      </div>
      {locked && (
        <label className="threshold-slider glass-panel">
          <LockKeyhole size={15} />
          <span>{formatBytes(threshold)}</span>
          <input
            aria-label="Size threshold"
            type="range"
            min={0}
            max={Math.max(points.length - 1, 0)}
            value={thresholdIndex}
            onChange={(event) => setHovered(Number(event.target.value))}
          />
        </label>
      )}
      <Card>
        <CardHeader
          title={t("sizeDistribution")}
          description={t("sizeDistributionHint")}
          action={
            <Segmented
              value={scale}
              onChange={onScaleChange}
              label="Size scale"
              options={[
                { value: "linear", label: t("linear") },
                { value: "log", label: t("logarithmic") },
              ]}
            />
          }
        />
        <ReactECharts
          option={distributionOption(points, thresholdIndex, zoom, t)}
          onEvents={commonEvents}
          notMerge
          className="chart-large"
        />
      </Card>
      <Card>
        <CardHeader title={t("cumulative")} description={t("cumulativeHint")} />
        <ReactECharts
          option={cumulativeOption(points, thresholdIndex, zoom, t)}
          onEvents={commonEvents}
          notMerge
          className="chart-large"
        />
      </Card>
    </div>
  );
}

function SplitStat({
  label,
  threshold,
  count,
  bytes,
  totalCount,
  totalBytes,
  tone,
  t,
}: {
  label: string;
  threshold: string;
  count: number;
  bytes: number;
  totalCount: number;
  totalBytes: number;
  tone: string;
  t: (key: TranslationKey | string) => string;
}) {
  return (
    <div className={`split-stat ${tone}`}>
      <span>
        {label}
        <b>{threshold}</b>
      </span>
      <strong>
        {formatNumber(count)}{" "}
        <small>
          · {totalCount ? ((count / totalCount) * 100).toFixed(1) : 0}%
        </small>
      </strong>
      <p>
        {formatBytes(bytes)} ·{" "}
        {totalBytes ? ((bytes / totalBytes) * 100).toFixed(1) : 0}% ·{" "}
        {t("averageSize")} {formatBytes(count ? bytes / count : 0)}
      </p>
    </div>
  );
}

function baseOption(
  points: SizePoint[],
  thresholdIndex: number,
  zoom: ZoomRange,
) {
  const labels = points.map((point) => formatBytes(point.upper));
  const thresholdLabel = labels[thresholdIndex];
  return {
    animationDuration: 280,
    animationEasing: "cubicOut",
    grid: { left: 62, right: 26, top: 28, bottom: 70 },
    tooltip: {
      trigger: "axis",
      backgroundColor: "rgba(20, 24, 36, .92)",
      borderWidth: 0,
      textStyle: { color: "#fff" },
      axisPointer: { type: "line", lineStyle: { color: "#ff5a5f", width: 2 } },
    },
    xAxis: {
      type: "category",
      boundaryGap: false,
      data: labels,
      axisLine: { lineStyle: { color: "#dce2ea" } },
      axisLabel: { color: "#687083", hideOverlap: true },
      splitLine: { show: false },
    },
    yAxis: {
      type: "value",
      axisLabel: { color: "#687083" },
      splitLine: { lineStyle: { color: "#edf0f5" } },
    },
    dataZoom: [
      { type: "inside", start: zoom.start, end: zoom.end },
      {
        type: "slider",
        start: zoom.start,
        end: zoom.end,
        height: 20,
        bottom: 16,
        borderColor: "transparent",
        backgroundColor: "#f2f4f8",
        fillerColor: "rgba(255,90,95,.18)",
        handleStyle: { color: "#ff5a5f" },
      },
    ],
    markLine: {
      silent: true,
      symbol: "none",
      data: [{ xAxis: thresholdLabel }],
      lineStyle: { color: "#ff5a5f", width: 2 },
      label: {
        formatter: thresholdLabel,
        color: "#b4233f",
        backgroundColor: "#fff1f2",
        padding: [4, 7],
        borderRadius: 5,
      },
    },
    markArea: {
      silent: true,
      data: [
        [
          { xAxis: labels[0], itemStyle: { color: "rgba(0,168,232,.045)" } },
          { xAxis: thresholdLabel },
        ],
        [
          {
            xAxis: thresholdLabel,
            itemStyle: { color: "rgba(255,90,95,.045)" },
          },
          { xAxis: labels.at(-1) },
        ],
      ],
    },
  };
}

function distributionOption(
  points: SizePoint[],
  thresholdIndex: number,
  zoom: ZoomRange,
  t: (key: TranslationKey | string) => string,
) {
  const base = baseOption(points, thresholdIndex, zoom);
  return {
    ...base,
    series: [
      {
        name: t("fileCount"),
        type: "line",
        smooth: 0.42,
        smoothMonotone: "x",
        showSymbol: false,
        lineStyle: {
          width: 3,
          color: "#ff5a5f",
          shadowBlur: 12,
          shadowColor: "rgba(255,90,95,.25)",
        },
        itemStyle: { color: "#ff5a5f" },
        areaStyle: {
          opacity: 1,
          color: {
            type: "linear",
            x: 0,
            y: 0,
            x2: 0,
            y2: 1,
            colorStops: [
              { offset: 0, color: "rgba(255,90,95,.48)" },
              { offset: 0.55, color: "rgba(255,193,69,.18)" },
              { offset: 1, color: "rgba(255,255,255,0)" },
            ],
          },
        },
        data: points.map((point) => point.count),
        markLine: base.markLine,
        markArea: base.markArea,
      },
    ],
  };
}

function cumulativeOption(
  points: SizePoint[],
  thresholdIndex: number,
  zoom: ZoomRange,
  t: (key: TranslationKey | string) => string,
) {
  const base = baseOption(points, thresholdIndex, zoom);
  return {
    ...base,
    yAxis: {
      ...base.yAxis,
      min: 0,
      max: 100,
      axisLabel: { formatter: "{value}%", color: "#687083" },
    },
    legend: {
      data: [t("cumulativeCount"), t("cumulativeBytes")],
      top: 0,
      right: 24,
    },
    series: [
      {
        name: t("cumulativeCount"),
        type: "line",
        smooth: 0.42,
        smoothMonotone: "x",
        showSymbol: false,
        lineStyle: { width: 3, color: "#7c3aed" },
        areaStyle: {
          color: {
            type: "linear",
            x: 0,
            y: 0,
            x2: 0,
            y2: 1,
            colorStops: [
              { offset: 0, color: "rgba(124,58,237,.30)" },
              { offset: 1, color: "rgba(124,58,237,0)" },
            ],
          },
        },
        data: points.map((point) => point.cumulativeCount),
        markLine: base.markLine,
        markArea: base.markArea,
      },
      {
        name: t("cumulativeBytes"),
        type: "line",
        smooth: 0.42,
        smoothMonotone: "x",
        showSymbol: false,
        lineStyle: { width: 3, color: "#00a8e8" },
        areaStyle: {
          color: {
            type: "linear",
            x: 0,
            y: 0,
            x2: 0,
            y2: 1,
            colorStops: [
              { offset: 0, color: "rgba(0,168,232,.24)" },
              { offset: 1, color: "rgba(0,168,232,0)" },
            ],
          },
        },
        data: points.map((point) => point.cumulativeBytes),
      },
    ],
  };
}
