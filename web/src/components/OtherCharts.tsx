import { memo, useMemo } from "react";
import { categoryColors } from "../lib/categories";
import { formatAge, formatBytes, formatNumber } from "../lib/format";
import type { TranslationKey } from "../lib/i18n";
import type { AgePoint, ChildUsage, HistoryPoint } from "../types";
import ReactECharts from "./EChart";
import { Card, CardHeader } from "./ui";

export const AgeChart = memo(function AgeChart({
  data,
  t,
}: {
  data: AgePoint[];
  t: (key: TranslationKey | string) => string;
}) {
  const option = useMemo(() => {
    const columns = 12;
    const rows = 5;
    const maximum = Math.max(1, ...data.map((point) => point.count));
    const cells = data.map((point, index) => {
      const lower = index === 0 ? 0 : data[index - 1].upperSeconds;
      return [
        index % columns,
        Math.floor(index / columns),
        point.count,
        point.bytes,
        lower,
        point.upperSeconds,
      ];
    });
    const rowLabels = Array.from({ length: rows }, (_, row) => {
      const start = row * columns;
      const end = Math.min((row + 1) * columns - 1, data.length - 1);
      if (!data[end]) return "";
      const lower = start === 0 ? 0 : data[start - 1].upperSeconds;
      return `${formatAge(lower)}–${formatAge(data[end].upperSeconds)}`;
    });
    return {
      animationDuration: 260,
      grid: { left: 96, right: 24, top: 28, bottom: 62 },
      tooltip: {
        formatter: (params: {
          value: [number, number, number, number, number, number];
        }) => {
          const value = params.value;
          return [
            `<strong>${formatAge(value[4])} – ${formatAge(value[5])}</strong>`,
            `${formatNumber(value[2])} ${t("files")}`,
            formatBytes(value[3]),
          ].join("<br/>");
        },
      },
      xAxis: {
        type: "category",
        data: Array.from({ length: columns }, (_, index) => index + 1),
        axisLabel: { color: "#7b8192", fontSize: 9, interval: 0 },
        axisTick: { show: false },
        axisLine: { show: false },
        splitArea: { show: false },
      },
      yAxis: {
        type: "category",
        inverse: true,
        data: rowLabels,
        axisLabel: { color: "#687083", fontSize: 10 },
        axisTick: { show: false },
        axisLine: { show: false },
      },
      visualMap: {
        min: 0,
        max: maximum,
        calculable: false,
        orient: "horizontal",
        left: "center",
        bottom: 4,
        itemWidth: 8,
        itemHeight: 120,
        dimension: 2,
        text: [t("dense"), t("quiet")],
        textStyle: { color: "#7b8192", fontSize: 10 },
        inRange: {
          color: ["#fff2ec", "#ffd2b8", "#ff9f68", "#ff5a5f", "#b72b6f"],
        },
      },
      series: [
        {
          type: "heatmap",
          data: cells,
          itemStyle: {
            borderColor: "rgba(255,255,255,.95)",
            borderWidth: 4,
            borderRadius: 7,
          },
          emphasis: {
            itemStyle: {
              borderColor: "#fff",
              shadowBlur: 14,
              shadowColor: "rgba(255,90,95,.28)",
            },
          },
        },
      ],
    };
  }, [data, t]);
  return (
    <Card className="heatmap-card">
      <CardHeader
        title={t("modifiedAge")}
        description={t("modifiedHeatmapHint")}
      />
      <ReactECharts option={option} className="chart-medium" />
    </Card>
  );
});

export const DirectoryCharts = memo(function DirectoryCharts({
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
});

export const HistoryChart = memo(function HistoryChart({
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
    legend: { data: [t("logicalSize"), t("fileCount")], top: 2 },
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
          inside: true,
          padding: [0, 4, 0, 0],
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
});
