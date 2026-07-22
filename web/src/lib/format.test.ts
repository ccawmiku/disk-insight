import { describe, expect, it } from "vitest";
import { formatAge, formatBytes, formatDuration } from "./format";

describe("format utilities", () => {
  it("formats binary storage units", () => {
    expect(formatBytes(1024)).toBe("1 KB");
    expect(formatBytes(3.5 * 1024 ** 3)).toBe("3.5 GB");
  });

  it("formats durations and ages", () => {
    expect(formatDuration(3665)).toBe("1h 1m");
    expect(formatAge(86400 * 30)).toBe("1mo");
    expect(formatAge(-1)).toBe("Future");
  });
});
