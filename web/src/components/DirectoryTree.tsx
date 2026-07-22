import { useVirtualizer } from "@tanstack/react-virtual";
import { ChevronRight, Database, Folder, FolderOpen } from "lucide-react";
import { useCallback, useEffect, useMemo, useRef, useState } from "react";
import { api } from "../lib/api";
import { cn } from "../lib/cn";
import { formatBytes } from "../lib/format";
import type { Root, TreeNode } from "../types";

interface FlatNode extends TreeNode {
  depth: number;
}

export function DirectoryTree({
  roots,
  rootId,
  path,
  onRootChange,
  onPathChange,
}: {
  roots: Root[];
  rootId?: number;
  path: string;
  onRootChange: (id: number) => void;
  onPathChange: (path: string) => void;
}) {
  const [expanded, setExpanded] = useState<Set<string>>(new Set([""]));
  const [children, setChildren] = useState<Record<string, TreeNode[]>>({});
  const [loading, setLoading] = useState<Set<string>>(new Set());
  const parentRef = useRef<HTMLDivElement>(null);
  const activeRoot = roots.find((root) => root.id === rootId);

  const loadChildren = useCallback(
    async (nodePath: string) => {
      if (!rootId || loading.has(nodePath)) return;
      setLoading((current) => new Set(current).add(nodePath));
      try {
        const nodes = await api.tree(rootId, nodePath);
        setChildren((current) => ({ ...current, [nodePath]: nodes }));
      } finally {
        setLoading((current) => {
          const next = new Set(current);
          next.delete(nodePath);
          return next;
        });
      }
    },
    [rootId, loading],
  );

  useEffect(() => {
    if (!rootId || children[""] || loading.has("")) return;
    void loadChildren("");
  }, [rootId, children, loading, loadChildren]);

  const flatNodes = useMemo(() => {
    const result: FlatNode[] = [];
    const walk = (parentPath: string, depth: number) => {
      for (const node of children[parentPath] ?? []) {
        result.push({ ...node, depth });
        if (expanded.has(node.path)) walk(node.path, depth + 1);
      }
    };
    walk("", 1);
    return result;
  }, [children, expanded]);

  const virtualizer = useVirtualizer({
    count: flatNodes.length,
    getScrollElement: () => parentRef.current,
    estimateSize: () => 36,
    overscan: 10,
  });

  const toggle = (node: TreeNode) => {
    if (!node.hasChildren) return;
    setExpanded((current) => {
      const next = new Set(current);
      if (next.has(node.path)) next.delete(node.path);
      else next.add(node.path);
      return next;
    });
    if (!children[node.path]) void loadChildren(node.path);
  };

  return (
    <div className="directory-tree">
      <label className="sr-only" htmlFor="root-select">
        Scan root
      </label>
      <select
        id="root-select"
        value={rootId ?? ""}
        onChange={(event) => onRootChange(Number(event.target.value))}
      >
        {roots.map((root) => (
          <option key={root.id} value={root.id}>
            {root.name}
          </option>
        ))}
      </select>
      {activeRoot && (
        <button
          type="button"
          className={cn("tree-row root-row", path === "" && "selected")}
          onClick={() => onPathChange("")}
        >
          <span className="tree-expander">
            <Database size={16} />
          </span>
          <span className="tree-name">{activeRoot.name}</span>
          <span className="tree-size">
            {formatBytes(activeRoot.lastLogicalSize)}
          </span>
        </button>
      )}
      <div className="tree-scroll" ref={parentRef}>
        <div
          style={{
            height: `${virtualizer.getTotalSize()}px`,
            position: "relative",
          }}
        >
          {virtualizer.getVirtualItems().map((virtualRow) => {
            const node = flatNodes[virtualRow.index];
            const open = expanded.has(node.path);
            return (
              <div
                key={node.path}
                style={{
                  position: "absolute",
                  top: 0,
                  left: 0,
                  width: "100%",
                  height: `${virtualRow.size}px`,
                  transform: `translateY(${virtualRow.start}px)`,
                }}
              >
                <div
                  className={cn("tree-row", path === node.path && "selected")}
                  style={{ paddingLeft: `${node.depth * 14 + 8}px` }}
                >
                  <button
                    type="button"
                    aria-label={
                      open ? "Collapse directory" : "Expand directory"
                    }
                    className={cn("tree-expander", open && "open")}
                    onClick={(event) => {
                      event.stopPropagation();
                      toggle(node);
                    }}
                  >
                    {node.hasChildren ? <ChevronRight size={14} /> : null}
                  </button>
                  <button
                    type="button"
                    className="tree-select"
                    onClick={() => onPathChange(node.path)}
                  >
                    {open ? <FolderOpen size={16} /> : <Folder size={16} />}
                    <span className="tree-name">{node.name}</span>
                    <span className="tree-size">{formatBytes(node.size)}</span>
                  </button>
                </div>
              </div>
            );
          })}
        </div>
      </div>
    </div>
  );
}
