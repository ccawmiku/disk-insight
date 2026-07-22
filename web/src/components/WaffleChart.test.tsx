import { fireEvent, render, screen } from "@testing-library/react";
import { describe, expect, it, vi } from "vitest";
import { dictionary } from "../lib/i18n";
import { allocateCells, WaffleChart } from "./WaffleChart";

const data = [
  { category: "video", count: 75, bytes: 900 },
  { category: "document", count: 25, bytes: 100 },
];

describe("WaffleChart", () => {
  it("always allocates exactly 200 cells", () => {
    const cells = allocateCells(data, "bytes", 200);
    expect(cells).toHaveLength(200);
    expect(cells.filter((cell) => cell.category === "video")).toHaveLength(180);
  });

  it("exposes category filtering through the legend", () => {
    const onToggle = vi.fn();
    render(
      <WaffleChart
        data={data}
        selected={[]}
        onToggle={onToggle}
        t={dictionary("en-US")}
      />,
    );
    fireEvent.click(screen.getByRole("button", { name: /Video/ }));
    expect(onToggle).toHaveBeenCalledWith("video");
  });
});
