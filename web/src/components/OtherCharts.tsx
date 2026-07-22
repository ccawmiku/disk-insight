import { categoryColors } from "../lib/categories";
import { formatAge, formatBytes, formatNumber } from "../lib/format";
import type { TranslationKey } from "../lib/i18n";
import type { AgePoint, ChildUsage, HistoryPoint, Scale } from "../types";
import ReactECharts from "./EChart";
import { Card, CardHeader, Segmented } from "./ui";

export function AgeChart({
  data,
  scale,
  onScaleChange,
  t,
}: {
  data: AgePoint[];
  scale: Scale;
  onScaleChange: (scale: Scale) => void;
  t: (key: TranslationKey | string) => string;
}) {
  const option = {
    animationDuration: 280,
    grid: { left: 58, right: 20, top: 26, bottom: 55 },
    tooltip: {
      trigger: "axis",
      formatter: (params: Array<{ name: string; value: number }>) =>
        `${params[0]?.name}<br/><strong>${formatNumber(params[0]?.value ?? 0)}</strong> ${t("files")}`,
    },
    xAxis: {
      type: "category",
      boundaryGap: false,
      data: data.map((point) => formatAge(point.upperSeconds)),
      axisLabel: { color: "#687083", hideOverlap: true },
      axisLine: { lineStyle: { color: "#dce2ea" } },
    },
    yAxis: {
      type: "value",
      axisLabel: { color: "#687083" },
      splitLine: { lineStyle: { color: "#edf0f5" } },
    },
    dataZoom: [{ type: "inside" }],
    series: [
      {
        type: "line",
        smooth: 0.45,
        smoothMonotone: "x",
        showSymbol: false,
        data: data.map((point) => point.count),
        lineStyle: { color: "#ff8a00", width: 3 },
        areaStyle: {
          color: {
            type: "linear",
            x: 0,
            y: 0,
            x2: 0,
            y2: 1,
            colorStops: [
              { offset: 0, color: "rgba(255,138,0,.45)" },
              { offset: 0.55, color: "rgba(255,193,69,.16)" },
              { offset: 1, color: "rgba(255,255,255,0)" },
            ],
          },
        },
      },
    ],
  };
  return (
    <Card>
      <CardHeader
        title={t("modifiedAge")}
        action={
          <Segmented
            value={scale}
            onChange={onScaleChange}
            label="Age scale"
            options={[
              { value: "linear", label: t("linear") },
              { value: "log", label: t("logarithmic") },
            ]}
          />
        }
      />
      <ReactECharts option={option} className="chart-medium" />
    </Card>
  );
}

export function DirectoryCharts({
  data,
  onNavigate,
  t,
}: {
  data: ChildUsage[];
  onNavigate: (path: string) => void;
  t: (key: TranslationKey | string) => string;
}) {
  const limited = data.slice(0, 30);
  const treemapOption = {
    animationDuration: 280,
    tooltip: {
      formatter: (params: {
        name: string;
        value: number;
        data: { fileCount: number };
      }) =>
        `<strong>${params.name}</strong><br/>${formatBytes(params.value)} · ${formatNumber(params.data.fileCount)} ${t("files")}`,
    },
    series: [
      {
        type: "treemap",
        roam: false,
        nodeClick: false,
        breadcrumb: { show: false },
        visibleMin: 300,
        upperLabel: { show: false },
        label: { color: "#fff", fontWeight: 700, formatter: "{b}" },
        itemStyle: {
          borderColor: "#fff",
          borderWidth: 3,
          gapWidth: 2,
          borderRadius: 8,
        },
        color: Object.values(categoryColors).slice(0, 10),
        data: limited.map((item) => ({
          name: item.name,
          value: item.size,
          fileCount: item.fileCount,
          path: item.path,
        })),
      },
    ],
  };
  const barData = data.slice(0, 12);
  const barOption = {
    animationDuration: 280,
    grid: { left: 110, right: 24, top: 10, bottom: 28 },
    tooltip: {
      trigger: "axis",
      axisPointer: { type: "shadow" },
      formatter: (params: Array<{ name: string; value: number }>) =>
        `<strong>${params[0]?.name}</strong><br/>${formatBytes(params[0]?.value ?? 0)}`,
    },
    xAxis: {
      type: "value",
      axisLabel: {
        formatter: (value: number) => formatBytes(value, 0),
        color: "#687083",
      },
      splitLine: { lineStyle: { color: "#edf0f5" } },
    },
    yAxis: {
      type: "category",
      inverse: true,
      data: barData.map((item) => item.name),
      axisLabel: { color: "#444b5d", width: 90, overflow: "truncate" },
      axisTick: { show: false },
      axisLine: { show: false },
    },
    series: [
      {
        type: "bar",
        data: barData.map((item, index) => ({
          value: item.size,
          itemStyle: {
            color: Object.values(categoryColors)[index % 11],
            borderRadius: [0, 6, 6, 0],
          },
        })),
        barMaxWidth: 16,
      },
    ],
  };
  return (
    <div className="two-column-charts">
      <Card>
        <CardHeader title={t("directoryMap")} />
        <ReactECharts
          option={treemapOption}
          className="chart-medium"
          onEvents={{
            click: (params: { data?: { path?: string } }) =>
              params.data?.path && onNavigate(params.data.path),
          }}
        />
      </Card>
      <Card>
        <CardHeader title={t("childRanking")} />
        <ReactECharts
          option={barOption}
          className="chart-medium"
          onEvents={{
            click: (params: { dataIndex?: number }) =>
              typeof params.dataIndex === "number" &&
              onNavigate(barData[params.dataIndex].path),
          }}
        />
      </Card>
    </div>
  );
}

export function HistoryChart({
  data,
  t,
}: {
  data: HistoryPoint[];
  t: (key: TranslationKey | string) => string;
}) {
  const option = {
    animationDuration: 280,
    grid: { left: 68, right: 68, top: 38, bottom: 48 },
    tooltip: { trigger: "axis" },
    legend: { data: [t("logicalSize"), t("fileCount")] },
    xAxis: {
      type: "category",
      data: data.map((item) => new Date(item.completedAt).toLocaleDateString()),
      axisLabel: { color: "#687083", hideOverlap: true },
    },
    yAxis: [
      {
        type: "value",
        axisLabel: {
          formatter: (value: number) => formatBytes(value, 0),
          color: "#687083",
        },
        splitLine: { lineStyle: { color: "#edf0f5" } },
      },
      {
        type: "value",
        axisLabel: {
          formatter: (value: number) => formatNumber(value),
          color: "#687083",
        },
        splitLine: { show: false },
      },
    ],
    series: [
      {
        name: t("logicalSize"),
        type: "line",
        smooth: 0.42,
        smoothMonotone: "x",
        yAxisIndex: 0,
        showSymbol: data.length < 20,
        data: data.map((item) => item.logicalSize),
        lineStyle: { width: 3, color: "#e848a0" },
        areaStyle: { color: "rgba(232,72,160,.12)" },
      },
      {
        name: t("fileCount"),
        type: "line",
        smooth: 0.42,
        smoothMonotone: "x",
        yAxisIndex: 1,
        showSymbol: data.length < 20,
        data: data.map((item) => item.fileCount),
        lineStyle: { width: 3, color: "#00a8e8" },
      },
    ],
  };
  return (
    <Card>
      <CardHeader title={t("history")} description={t("rootHistoryHint")} />
      <ReactECharts option={option} className="chart-large" />
    </Card>
  );
}
