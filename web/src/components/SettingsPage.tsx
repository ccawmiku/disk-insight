import { Check, Palette, RotateCcw, Save, TimerReset } from "lucide-react";
import { useEffect, useState } from "react";
import type { TranslationKey } from "../lib/i18n";
import type { Settings, ThemeName } from "../types";
import { Button, Card, CardHeader } from "./ui";

const themes: Array<{
  value: ThemeName;
  key: TranslationKey;
  colors: string[];
}> = [
  {
    value: "tropical-coral",
    key: "tropicalCoral",
    colors: ["#ff5a5f", "#ff8a00", "#ffc145", "#e848a0", "#00a8e8"],
  },
  {
    value: "citrus-sunset",
    key: "citrusSunset",
    colors: ["#f94144", "#f3722c", "#f8961e", "#f9c74f", "#9b5de5"],
  },
  {
    value: "ocean-glow",
    key: "oceanGlow",
    colors: ["#00b4d8", "#0077b6", "#5a4fcf", "#00c2a8", "#90e0ef"],
  },
  {
    value: "berry-candy",
    key: "berryCandy",
    colors: ["#ff4d8d", "#c44dff", "#7950f2", "#ff7b54", "#4dabf7"],
  },
];

export function SettingsPage({
  settings,
  t,
  onSave,
}: {
  settings: Settings;
  t: (key: TranslationKey | string) => string;
  onSave: (settings: Settings) => Promise<void>;
}) {
  const [draft, setDraft] = useState(settings);
  const [excludeText, setExcludeText] = useState(() =>
    settings.exclude.join("\n"),
  );
  const [dirty, setDirty] = useState(false);
  const [saved, setSaved] = useState(false);
  useEffect(() => {
    if (dirty) return;
    setDraft(settings);
    setExcludeText(settings.exclude.join("\n"));
  }, [settings, dirty]);
  useEffect(() => {
    document.documentElement.dataset.theme = draft.theme;
    return () => {
      document.documentElement.dataset.theme = settings.theme;
    };
  }, [draft.theme, settings.theme]);
  const updateDraft = (next: Partial<Settings>) => {
    setDraft((current) => ({ ...current, ...next }));
    setDirty(true);
  };
  const reset = () => {
    setDraft(settings);
    setExcludeText(settings.exclude.join("\n"));
    setDirty(false);
  };
  const save = async () => {
    const next = {
      ...draft,
      exclude: Array.from(
        new Set(
          excludeText
            .split(/\r?\n/)
            .map((line) => line.trim())
            .filter(Boolean),
        ),
      ),
    };
    await onSave(next);
    setDraft(next);
    setExcludeText(next.exclude.join("\n"));
    setDirty(false);
    setSaved(true);
    window.setTimeout(() => setSaved(false), 1800);
  };
  return (
    <div className="page-stack settings-page">
      <div className="page-heading">
        <div>
          <span className="eyebrow">Preferences</span>
          <h1>{t("settings")}</h1>
        </div>
        <div className="settings-actions">
          {dirty && <span className="unsaved-badge">{t("unsaved")}</span>}
          <Button variant="ghost" onClick={reset} disabled={!dirty}>
            <RotateCcw size={16} />
            {t("discard")}
          </Button>
          <Button onClick={() => void save()} disabled={!dirty}>
            <Save size={16} />
            {saved ? t("saved") : t("save")}
          </Button>
        </div>
      </div>
      <Card>
        <CardHeader title={t("schedule")} action={<TimerReset size={20} />} />
        <div className="form-grid">
          <label>
            <span>{t("schedule")}</span>
            <select
              value={draft.scheduleKind}
              onChange={(event) =>
                updateDraft({
                  scheduleKind: event.target.value as Settings["scheduleKind"],
                })
              }
            >
              <option value="daily">{t("daily")}</option>
              <option value="weekly">{t("weekly")}</option>
              <option value="off">{t("off")}</option>
            </select>
          </label>
          <label>
            <span>{t("scheduleTime")}</span>
            <input
              type="time"
              value={draft.scheduleTime}
              onChange={(event) =>
                updateDraft({ scheduleTime: event.target.value })
              }
            />
          </label>
          {draft.scheduleKind === "weekly" && (
            <label>
              <span>{t("weekday")}</span>
              <select
                value={draft.scheduleDay}
                onChange={(event) =>
                  updateDraft({
                    scheduleDay: Number(event.target.value),
                  })
                }
              >
                {[1, 2, 3, 4, 5, 6, 7].map((day) => (
                  <option value={day} key={day}>
                    {day}
                  </option>
                ))}
              </select>
            </label>
          )}
          <label>
            <span>{t("timezone")}</span>
            <input
              value={draft.timezone}
              onChange={(event) =>
                updateDraft({ timezone: event.target.value })
              }
            />
          </label>
        </div>
      </Card>
      <Card>
        <CardHeader title={t("appearance")} action={<Palette size={20} />} />
        <div className="theme-grid">
          {themes.map((theme) => (
            <button
              type="button"
              key={theme.value}
              className={draft.theme === theme.value ? "selected" : ""}
              onClick={() => updateDraft({ theme: theme.value })}
            >
              <span className="theme-swatches">
                {theme.colors.map((color) => (
                  <i key={color} style={{ background: color }} />
                ))}
              </span>
              <strong>{t(theme.key)}</strong>
              {draft.theme === theme.value && <Check size={17} />}
            </button>
          ))}
        </div>
        <div className="form-grid language-row">
          <label>
            <span>{t("language")}</span>
            <select
              value={draft.language}
              onChange={(event) =>
                updateDraft({
                  language: event.target.value as Settings["language"],
                })
              }
            >
              <option value="zh-CN">简体中文</option>
              <option value="en-US">English</option>
            </select>
          </label>
        </div>
      </Card>
      <Card>
        <CardHeader title={t("exclusions")} description={t("exclusionsHint")} />
        <textarea
          aria-label={t("exclusions")}
          rows={8}
          value={excludeText}
          onChange={(event) => {
            setExcludeText(event.target.value);
            setDirty(true);
          }}
        />
      </Card>
    </div>
  );
}
