import type {
  Dashboard,
  Root,
  Scale,
  ScanError,
  ScanProgress,
  Settings,
  TreeNode,
} from "../types";

async function request<T>(url: string, init?: RequestInit): Promise<T> {
  const response = await fetch(url, {
    ...init,
    headers: { "Content-Type": "application/json", ...init?.headers },
  });
  if (!response.ok) {
    const payload = (await response.json().catch(() => null)) as {
      error?: string;
    } | null;
    throw new Error(
      payload?.error ?? `${response.status} ${response.statusText}`,
    );
  }
  if (response.status === 204) return undefined as T;
  return response.json() as Promise<T>;
}

export const api = {
  roots: () => request<Root[]>("/api/v1/roots"),
  settings: () => request<Settings>("/api/v1/settings"),
  updateSettings: (settings: Settings) =>
    request<Settings>("/api/v1/settings", {
      method: "PUT",
      body: JSON.stringify(settings),
    }),
  progress: () => request<ScanProgress[]>("/api/v1/scans/progress"),
  startScan: (rootIds: number[]) =>
    request<{ startedRootIds: number[] }>("/api/v1/scans", {
      method: "POST",
      body: JSON.stringify({ rootIds }),
    }),
  cancelScan: (rootId: number) =>
    request<void>(`/api/v1/scans/${rootId}`, { method: "DELETE" }),
  errors: (rootId: number) =>
    request<ScanError[]>(`/api/v1/scan-errors?rootId=${rootId}`),
  tree: (rootId: number, path: string) =>
    request<TreeNode[]>(
      `/api/v1/tree?rootId=${rootId}&path=${encodeURIComponent(path)}`,
    ),
  dashboard: (
    rootId: number,
    path: string,
    categories: string[],
    sizeScale: Scale,
    ageScale: Scale,
  ) => {
    const params = new URLSearchParams({
      rootId: String(rootId),
      path,
      categories: categories.join(","),
      sizeScale,
      ageScale,
    });
    return request<Dashboard>(`/api/v1/dashboard?${params}`);
  },
};
