import {
  BarChart,
  HeatmapChart,
  LineChart,
  TreemapChart,
} from "echarts/charts";
import {
  DataZoomComponent,
  GridComponent,
  LegendComponent,
  MarkAreaComponent,
  MarkLineComponent,
  TooltipComponent,
  VisualMapComponent,
} from "echarts/components";
import { type EChartsType, init, use } from "echarts/core";
import { CanvasRenderer } from "echarts/renderers";
import { useEffect, useRef } from "react";

use([
  BarChart,
  LineChart,
  TreemapChart,
  HeatmapChart,
  DataZoomComponent,
  GridComponent,
  LegendComponent,
  MarkAreaComponent,
  MarkLineComponent,
  TooltipComponent,
  VisualMapComponent,
  CanvasRenderer,
]);

type ChartOption = Record<string, unknown>;
type EventHandler = (params: never) => void;
type EventBinder = (eventName: string, handler: EventHandler) => void;

export default function EChart({
  option,
  onEvents,
  notMerge = false,
  className,
}: {
  option: ChartOption;
  onEvents?: Record<string, EventHandler>;
  notMerge?: boolean;
  className?: string;
}) {
  const elementRef = useRef<HTMLDivElement>(null);
  const chartRef = useRef<EChartsType | null>(null);

  useEffect(() => {
    if (!elementRef.current) return;
    const chart = init(elementRef.current, undefined, { renderer: "canvas" });
    chartRef.current = chart;
    const observer = new ResizeObserver(() => chart.resize());
    observer.observe(elementRef.current);
    return () => {
      observer.disconnect();
      chart.dispose();
      chartRef.current = null;
    };
  }, []);

  useEffect(() => {
    chartRef.current?.setOption(
      option as Parameters<EChartsType["setOption"]>[0],
      { notMerge, lazyUpdate: true },
    );
  }, [option, notMerge]);

  useEffect(() => {
    const chart = chartRef.current;
    if (!chart || !onEvents) return;
    const on = chart.on.bind(chart) as unknown as EventBinder;
    const off = chart.off.bind(chart) as unknown as EventBinder;
    for (const [eventName, handler] of Object.entries(onEvents))
      on(eventName, handler);
    return () => {
      for (const [eventName, handler] of Object.entries(onEvents))
        off(eventName, handler);
    };
  }, [onEvents]);

  return <div ref={elementRef} className={className} />;
}
