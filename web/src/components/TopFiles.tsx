import { FileText } from "lucide-react";
import { categoryColors } from "../lib/categories";
import { formatBytes, formatDate } from "../lib/format";
import type { TranslationKey } from "../lib/i18n";
import type { FileItem } from "../types";
import { Card, CardHeader } from "./ui";

export function TopFiles({
  files,
  t,
  onNavigate,
}: {
  files: FileItem[];
  t: (key: TranslationKey | string) => string;
  onNavigate: (path: string) => void;
}) {
  return (
    <Card className="file-table-card">
      <CardHeader title={t("topFiles")} />
      <div className="table-scroll">
        <table>
          <thead>
            <tr>
              <th>{t("name")}</th>
              <th>{t("path")}</th>
              <th>{t("category")}</th>
              <th>{t("size")}</th>
              <th>{t("modified")}</th>
            </tr>
          </thead>
          <tbody>
            {files.map((file) => {
              const parent = file.path.includes("/")
                ? file.path.slice(0, file.path.lastIndexOf("/"))
                : "";
              return (
                <tr key={file.path}>
                  <td>
                    <span className="file-name">
                      <FileText size={15} />
                      {file.name}
                    </span>
                  </td>
                  <td>
                    <button
                      type="button"
                      className="path-link"
                      onClick={() => onNavigate(parent)}
                      title={file.path}
                    >
                      {file.path}
                    </button>
                  </td>
                  <td>
                    <span className="category-badge">
                      <i
                        style={{
                          backgroundColor: categoryColors[file.category],
                        }}
                      />
                      {t(file.category)}
                    </span>
                  </td>
                  <td className="numeric">{formatBytes(file.size)}</td>
                  <td>{formatDate(file.modifiedAt)}</td>
                </tr>
              );
            })}
          </tbody>
        </table>
      </div>
    </Card>
  );
}
