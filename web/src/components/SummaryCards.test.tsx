import { render, screen } from "@testing-library/react";
import { describe, expect, it } from "vitest";
import { dictionary } from "../lib/i18n";
import { SummaryCards } from "./SummaryCards";

describe("SummaryCards", () => {
  it("renders all six agreed summary metrics", () => {
    render(
      <SummaryCards
        t={dictionary("en-US")}
        summary={{
          logicalSize: 2048,
          allocatedSize: 4096,
          fileCount: 2,
          directoryCount: 1,
          largestFileName: "big.bin",
          largestFileSize: 1024,
          lastScanDurationMs: 1200,
          scanErrors: 0,
        }}
      />,
    );
    expect(screen.getByText("Logical size")).toBeInTheDocument();
    expect(screen.getByText("Allocated space")).toBeInTheDocument();
    expect(screen.getByText("Files")).toBeInTheDocument();
    expect(screen.getByText("Directories")).toBeInTheDocument();
    expect(screen.getByText("Last scan")).toBeInTheDocument();
  });
});
