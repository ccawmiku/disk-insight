import { useMemo, useState } from "react";
import { categoryColors } from "../lib/categories";
import { formatBytes, formatNumber } from "../lib/format";
import type { TranslationKey } from "../lib/i18n";
import type { CategoryStat } from "../types";
import { Card, CardHeader, Segmented } from "./ui";

type Metric = "count" | "bytes";

export function WaffleChart({
  data,
  selected,
  onToggle,
  t,
}: {
  data: CategoryStat[];
  selected: string[];
  onToggle: (category: string) => void;
  t: (key: TranslationKey | string) => string;
}) {
  const [metric, setMetric] = useState<Metric>("bytes");
  const cells = useMemo(() => allocateCells(data, metric, 200), [data, metric]);
  const total = data.reduce((sum, item) => sum + item[metric], 0);
  return (
    <Card className="waffle-card">
      <CardHeader
        title={t("fileTypes")}
        action={
          <Segmented
            value={metric}
            onChange={setMetric}
            label="Waffle metric"
            options={[
              { value: "count", label: t("byCount") },
              { value: "bytes", label: t("bySize") },
            ]}
          />
        }
      />
      <div className="waffle-layout">
        <div
          className="waffle"
          role="img"
          aria-label={`${t("fileTypes")} ${metric}`}
        >
          {cells.map((cell) => (
            <span
              key={cell.id}
              className={
                selected.length === 0 || selected.includes(cell.category)
                  ? ""
                  : "muted"
              }
              style={{ backgroundColor: categoryColors[cell.category] }}
              title={t(cell.category)}
            />
          ))}
        </div>
        <div className="waffle-legend">
          {data.map((item) => {
            const value = item[metric];
            return (
              <button
                key={item.category}
                type="button"
                aria-pressed={selected.includes(item.category)}
                className={selected.includes(item.category) ? "selected" : ""}
                onClick={() => onToggle(item.category)}
              >
                <i style={{ background: categoryColors[item.category] }} />
                <span>
                  <strong>{t(item.category)}</strong>
                  <small>
                    {metric === "bytes"
                      ? formatBytes(value)
                      : formatNumber(value)}{" "}
                    · {total ? ((value / total) * 100).toFixed(1) : 0}%
                  </small>
                </span>
              </button>
            );
          })}
        </div>
      </div>
    </Card>
  );
}

export function allocateCells(
  data: CategoryStat[],
  metric: Metric,
  count: number,
): Array<{ id: string; category: string }> {
  const nonzero = data.filter((item) => item[metric] > 0);
  const total = nonzero.reduce((sum, item) => sum + item[metric], 0);
  if (!total) return [];
  const allocations = nonzero.map((item) => ({
    category: item.category,
    exact: (item[metric] / total) * count,
    cells: Math.max(1, Math.floor((item[metric] / total) * count)),
  }));
  let used = allocations.reduce((sum, item) => sum + item.cells, 0);
  while (used > count) {
    const candidate = allocations
      .filter((item) => item.cells > 1)
      .sort((a, b) => a.exact - a.cells - (b.exact - b.cells))[0];
    if (!candidate) break;
    candidate.cells--;
    used--;
  }
  while (used < count) {
    allocations.sort((a, b) => b.exact - b.cells - (a.exact - a.cells));
    allocations[0].cells++;
    used++;
  }
  return allocations.flatMap((item) =>
    Array.from({ length: item.cells }, (_, cellIndex) => ({
      id: `${item.category}-${cellIndex}`,
      category: item.category,
    })),
  );
}
