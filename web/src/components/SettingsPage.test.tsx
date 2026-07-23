import { fireEvent, render, screen, waitFor } from "@testing-library/react";
import { describe, expect, it, vi } from "vitest";
import { dictionary } from "../lib/i18n";
import type { Settings } from "../types";
import { SettingsPage } from "./SettingsPage";

const settings: Settings = {
  scheduleKind: "daily",
  scheduleTime: "03:00",
  scheduleDay: 1,
  timezone: "Asia/Shanghai",
  theme: "tropical-coral",
  language: "zh-CN",
  exclude: ["node_modules"],
};

describe("SettingsPage", () => {
  it("preserves a multiline unsaved draft across background prop refreshes", async () => {
    const onSave = vi.fn(async () => undefined);
    const view = render(
      <SettingsPage
        settings={settings}
        t={dictionary("zh-CN")}
        onSave={onSave}
      />,
    );
    const textarea = screen.getByLabelText(dictionary("zh-CN")("exclusions"));
    fireEvent.change(textarea, {
      target: { value: "node_modules\n.cache\n\narchive-*" },
    });
    fireEvent.click(screen.getByRole("button", { name: "海洋流光" }));

    view.rerender(
      <SettingsPage
        settings={{ ...settings }}
        t={dictionary("zh-CN")}
        onSave={onSave}
      />,
    );

    expect(textarea).toHaveValue("node_modules\n.cache\n\narchive-*");
    expect(screen.getByText("有未保存修改")).toBeVisible();
    fireEvent.click(screen.getByRole("button", { name: "保存设置" }));
    await waitFor(() =>
      expect(onSave).toHaveBeenCalledWith({
        ...settings,
        theme: "ocean-glow",
        exclude: ["node_modules", ".cache", "archive-*"],
      }),
    );
  });
});
