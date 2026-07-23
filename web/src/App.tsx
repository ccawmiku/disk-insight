import { Database, LoaderCircle } from "lucide-react";
import { useCallback, useEffect, useMemo, useState } from "react";
import {
  AgeChart,
  DirectoryCharts,
  HistoryChart,
} from "./components/OtherCharts";
import { ScanManagement } from "./components/ScanManagement";
import { SettingsPage } from "./components/SettingsPage";
import { Sidebar } from "./components/Sidebar";
import { SizeAnalytics } from "./components/SizeAnalytics";
import { SummaryCards } from "./components/SummaryCards";
import { Topbar } from "./components/Topbar";
import { TopFiles } from "./components/TopFiles";
import { EmptyState, ErrorState, Skeleton } from "./components/ui";
import { WaffleChart } from "./components/WaffleChart";
import { api } from "./lib/api";
import { dictionary } from "./lib/i18n";
import type {
  CategoryStat,
  Dashboard,
  Page,
  Root,
  Scale,
  ScanProgress,
  Settings,
} from "./types";

const defaultSettings: Settings = {
  scheduleKind: "daily",
  scheduleTime: "03:00",
  scheduleDay: 1,
  timezone: "Asia/Shanghai",
  theme: "tropical-coral",
  language: "zh-CN",
  exclude: [],
};

export default function App() {
  const initial = useMemo(
    () => new URLSearchParams(window.location.search),
    [],
  );
  const [roots, setRoots] = useState<Root[]>([]);
  const [settings, setSettings] = useState<Settings>(defaultSettings);
  const [progress, setProgress] = useState<ScanProgress[]>([]);
  const [rootId, setRootId] = useState<number | undefined>(
    () => Number(initial.get("root")) || undefined,
  );
  const [path, setPath] = useState(() => initial.get("path") ?? "");
  const [page, setPage] = useState<Page>(
    () => (initial.get("page") as Page) || "overview",
  );
  const [categories, setCategories] = useState<string[]>(
    () => initial.get("categories")?.split(",").filter(Boolean) ?? [],
  );
  const [sizeScale, setSizeScale] = useState<Scale>("linear");
  const [dashboard, setDashboard] = useState<Dashboard | null>(null);
  const [catalog, setCatalog] = useState<CategoryStat[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<Error | null>(null);
  const [revision, setRevision] = useState(0);
  const [sidebarOpen, setSidebarOpen] = useState(false);
  const [sidebarCollapsed, setSidebarCollapsed] = useState(false);
  const t = useMemo(() => dictionary(settings.language), [settings.language]);
  const activeRoot = roots.find((root) => root.id === rootId);
  const activeProgress = progress.find((item) => item.rootId === rootId);

  const loadInitial = useCallback(async () => {
    try {
      const [nextRoots, nextSettings, nextProgress] = await Promise.all([
        api.roots(),
        api.settings(),
        api.progress(),
      ]);
      setRoots(nextRoots);
      setSettings(nextSettings);
      setProgress(nextProgress);
      setRootId((current) =>
        current && nextRoots.some((root) => root.id === current)
          ? current
          : nextRoots[0]?.id,
      );
    } catch (loadError) {
      setError(loadError as Error);
    }
  }, []);

  useEffect(() => {
    void loadInitial();
  }, [loadInitial]);

  const refreshProgress = useCallback(async () => {
    try {
      const next = await api.progress();
      setProgress((current) =>
        JSON.stringify(current) === JSON.stringify(next) ? current : next,
      );
    } catch (loadError) {
      setError(loadError as Error);
    }
  }, []);

  const refreshRoots = useCallback(async () => {
    try {
      const next = await api.roots();
      setRoots((current) =>
        JSON.stringify(current) === JSON.stringify(next) ? current : next,
      );
    } catch (loadError) {
      setError(loadError as Error);
    }
  }, []);

  useEffect(() => {
    const timer = window.setInterval(() => void refreshProgress(), 3000);
    return () => window.clearInterval(timer);
  }, [refreshProgress]);

  useEffect(() => {
    document.documentElement.dataset.theme = settings.theme;
    document.documentElement.lang = settings.language;
  }, [settings.theme, settings.language]);

  useEffect(() => {
    const params = new URLSearchParams();
    if (rootId) params.set("root", String(rootId));
    if (path) params.set("path", path);
    if (page !== "overview") params.set("page", page);
    if (categories.length) params.set("categories", categories.join(","));
    window.history.replaceState(
      null,
      "",
      `${window.location.pathname}${params.size ? `?${params}` : ""}`,
    );
  }, [rootId, path, page, categories]);

  const loadDashboard = useCallback(async () => {
    void revision;
    if (!rootId || page === "settings" || page === "scans") {
      setLoading(false);
      return;
    }
    setLoading(true);
    setError(null);
    try {
      const next = await api.dashboard(
        rootId,
        path,
        categories,
        sizeScale,
        "linear",
      );
      setDashboard(next);
      if (categories.length === 0) setCatalog(next.categories);
    } catch (loadError) {
      const nextError = loadError as Error;
      if (nextError.message.includes("no completed scan")) setDashboard(null);
      else setError(nextError);
    } finally {
      setLoading(false);
    }
  }, [rootId, path, categories, sizeScale, page, revision]);

  useEffect(() => {
    void loadDashboard();
  }, [loadDashboard]);
  const completedRunId =
    activeProgress?.stage === "completed" ? activeProgress.runId : undefined;
  useEffect(() => {
    if (completedRunId) {
      setRevision((value) => value + 1);
      void refreshRoots();
    }
  }, [completedRunId, refreshRoots]);

  const changeRoot = (id: number) => {
    setRootId(id);
    setPath("");
    setCategories([]);
    setCatalog([]);
  };
  const toggleCategory = useCallback(
    (category: string) =>
      setCategories((current) =>
        current.includes(category)
          ? current.filter((item) => item !== category)
          : [...current, category],
      ),
    [],
  );
  const startScan = async (ids = rootId ? [rootId] : []) => {
    await api.startScan(ids);
    await refreshProgress();
    setPage("scans");
  };
  const saveSettings = async (next: Settings) => {
    const saved = await api.updateSettings(next);
    setSettings(saved);
  };

  return (
    <div
      className={`app-shell ${sidebarCollapsed ? "sidebar-is-collapsed" : ""}`}
    >
      <div className="ambient ambient-one" />
      <div className="ambient ambient-two" />
      <div className="ambient ambient-three" />
      <Sidebar
        open={sidebarOpen}
        collapsed={sidebarCollapsed}
        page={page}
        roots={roots}
        rootId={rootId}
        path={path}
        t={t}
        onOpenChange={setSidebarOpen}
        onCollapsedChange={setSidebarCollapsed}
        onPageChange={setPage}
        onRootChange={changeRoot}
        onPathChange={setPath}
      />
      <main>
        {page !== "settings" && page !== "scans" && (
          <Topbar
            root={activeRoot}
            path={path}
            categories={categories}
            progress={activeProgress}
            t={t}
            onPathChange={setPath}
            onClearCategories={() => setCategories([])}
            onScan={() => void startScan()}
            onRefresh={() => setRevision((value) => value + 1)}
          />
        )}
        <div className="content">
          {page === "settings" ? (
            <SettingsPage settings={settings} t={t} onSave={saveSettings} />
          ) : page === "scans" ? (
            <ScanManagement
              roots={roots}
              progress={progress}
              selectedRootId={rootId}
              t={t}
              onCancel={(id) => void api.cancelScan(id).then(refreshProgress)}
              onScan={(id) => void startScan([id])}
            />
          ) : error ? (
            <ErrorState
              error={error}
              onRetry={() => setRevision((value) => value + 1)}
              retryLabel={t("retry")}
            />
          ) : loading ? (
            <DashboardSkeleton />
          ) : !activeRoot ? (
            <EmptyState
              title={t("noData")}
              description="Configure DISK_INSIGHT_ROOTS and recreate the container."
              action={<Database size={24} />}
            />
          ) : !dashboard ? (
            <EmptyState
              title={t("noData")}
              description={t("firstScanHint")}
              action={
                activeProgress ? <LoaderCircle className="spin" /> : undefined
              }
            />
          ) : dashboard.summary.fileCount === 0 ? (
            <EmptyState
              title={t("directoryEmpty")}
              description={t("firstScanHint")}
            />
          ) : (
            <>
              {page === "overview" && (
                <div className="page-stack">
                  <SummaryCards
                    summary={dashboard.summary}
                    root={activeRoot}
                    t={t}
                  />
                  <SizeAnalytics
                    points={dashboard.sizeDistribution}
                    summary={dashboard.summary}
                    scale={sizeScale}
                    onScaleChange={setSizeScale}
                    t={t}
                  />
                  <div className="two-column-charts">
                    <WaffleChart
                      data={catalog.length ? catalog : dashboard.categories}
                      selected={categories}
                      onToggle={toggleCategory}
                      t={t}
                    />
                    <AgeChart data={dashboard.ageDistribution} t={t} />
                  </div>
                  <DirectoryCharts
                    data={dashboard.children}
                    onNavigate={setPath}
                    t={t}
                  />
                </div>
              )}
              {page === "time" && (
                <div className="page-stack">
                  <div className="page-heading">
                    <div>
                      <span className="eyebrow">Timeline</span>
                      <h1>{t("time")}</h1>
                    </div>
                  </div>
                  <AgeChart data={dashboard.ageDistribution} t={t} />
                  <HistoryChart data={dashboard.history} t={t} />
                </div>
              )}
              {page === "files" && (
                <div className="page-stack">
                  <div className="page-heading">
                    <div>
                      <span className="eyebrow">Metadata only</span>
                      <h1>{t("files")}</h1>
                    </div>
                  </div>
                  <TopFiles
                    files={dashboard.topFiles}
                    t={t}
                    onNavigate={(next) => {
                      setPath(next);
                      setPage("overview");
                    }}
                  />
                </div>
              )}
            </>
          )}
        </div>
      </main>
    </div>
  );
}

function DashboardSkeleton() {
  return (
    <div className="page-stack">
      <div className="summary-grid">
        {[
          "logical",
          "allocated",
          "files",
          "directories",
          "largest",
          "scan",
        ].map((item) => (
          <Skeleton className="summary-skeleton" key={item} />
        ))}
      </div>
      <Skeleton className="chart-skeleton" />
      <Skeleton className="chart-skeleton" />
    </div>
  );
}
