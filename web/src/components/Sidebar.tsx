import {
  BarChart3,
  ChevronLeft,
  Clock3,
  Files,
  Menu,
  Radar,
  ScanSearch,
  Settings as SettingsIcon,
  X,
} from "lucide-react";
import { useEffect } from "react";
import { cn } from "../lib/cn";
import type { TranslationKey } from "../lib/i18n";
import type { Page, Root } from "../types";
import { DirectoryTree } from "./DirectoryTree";
import { Button, Tooltip } from "./ui";

const navigation: Array<{
  page: Page;
  label: TranslationKey;
  icon: typeof BarChart3;
}> = [
  { page: "overview", label: "overview", icon: BarChart3 },
  { page: "time", label: "time", icon: Clock3 },
  { page: "files", label: "files", icon: Files },
  { page: "scans", label: "scans", icon: ScanSearch },
  { page: "settings", label: "settings", icon: SettingsIcon },
];

export function Sidebar({
  open,
  collapsed,
  page,
  roots,
  rootId,
  path,
  t,
  onOpenChange,
  onCollapsedChange,
  onPageChange,
  onRootChange,
  onPathChange,
}: {
  open: boolean;
  collapsed: boolean;
  page: Page;
  roots: Root[];
  rootId?: number;
  path: string;
  t: (key: TranslationKey | string) => string;
  onOpenChange: (open: boolean) => void;
  onCollapsedChange: (collapsed: boolean) => void;
  onPageChange: (page: Page) => void;
  onRootChange: (id: number) => void;
  onPathChange: (path: string) => void;
}) {
  useEffect(() => {
    const close = (event: KeyboardEvent) =>
      event.key === "Escape" && onOpenChange(false);
    window.addEventListener("keydown", close);
    return () => window.removeEventListener("keydown", close);
  }, [onOpenChange]);

  return (
    <>
      <Button
        className="mobile-menu"
        variant="secondary"
        aria-label={t("menu")}
        onClick={() => onOpenChange(true)}
      >
        <Menu size={19} />
      </Button>
      <button
        type="button"
        aria-label="Close menu"
        className={cn("sidebar-backdrop", open && "visible")}
        onClick={() => onOpenChange(false)}
      />
      <aside
        className={cn(
          "sidebar",
          collapsed && "collapsed",
          open && "mobile-open",
        )}
      >
        <header className="brand">
          <div className="brand-mark">
            <Radar size={22} />
          </div>
          {!collapsed && (
            <div>
              <strong>{t("appName")}</strong>
              <span>{t("tagline")}</span>
            </div>
          )}
          <Button
            variant="ghost"
            className="mobile-close"
            aria-label="Close"
            onClick={() => onOpenChange(false)}
          >
            <X size={18} />
          </Button>
        </header>
        <nav className="primary-nav" aria-label="Primary">
          {navigation.map(({ page: itemPage, label, icon: Icon }) => {
            const button = (
              <button
                key={itemPage}
                type="button"
                className={cn(page === itemPage && "active")}
                onClick={() => {
                  onPageChange(itemPage);
                  onOpenChange(false);
                }}
              >
                <Icon size={18} />
                <span>{t(label)}</span>
              </button>
            );
            return collapsed ? (
              <Tooltip key={itemPage} content={t(label)}>
                {button}
              </Tooltip>
            ) : (
              button
            );
          })}
        </nav>
        {!collapsed && page !== "settings" && roots.length > 0 && (
          <DirectoryTree
            key={rootId}
            roots={roots}
            rootId={rootId}
            path={path}
            onRootChange={onRootChange}
            onPathChange={(next) => {
              onPathChange(next);
              onOpenChange(false);
            }}
          />
        )}
        <button
          type="button"
          className="collapse-button"
          onClick={() => onCollapsedChange(!collapsed)}
        >
          <ChevronLeft className={collapsed ? "rotated" : ""} size={16} />
          <span>{collapsed ? t("expand") : t("collapse")}</span>
        </button>
      </aside>
    </>
  );
}
