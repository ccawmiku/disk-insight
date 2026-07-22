# Disk Insight

Disk Insight 是一个面向本地磁盘、机械硬盘和 NAS 挂载目录的低影响空间分析仪表盘。它通过只读 Docker 挂载保留原始目录结构，使用温和的单路元数据扫描，并提供可下钻、可筛选、可联动的可视化。

Disk Insight is a low-impact, read-only storage analytics dashboard for local disks, HDDs, and mounted network storage. The interface is available in Simplified Chinese and English.

## 功能

- 多个只读挂载根目录，或一个共同父目录
- 左侧虚拟化目录树、面包屑和递归子目录统计
- 文件大小分布折线图，支持线性/对数轴、缩放和阈值锁定
- 累计文件数量与累计占用空间百分比曲线
- 悬停阈值左右两侧的数量、空间、占比和平均大小统计
- 20×10 Waffle Chart，按视频、音频、图片、文档等语义类别统计
- 文件类别多选筛选，并联动全部图表和文件排行
- 修改时间分布、容量历史、Treemap 和子目录占用排行
- 最大的 100 个文件，仅显示相对路径和元数据
- 首次自动扫描、手动扫描、每天或每周定时扫描
- 详细扫描进度：阶段、当前目录、数量、容量、速度、耗时、ETA 和错误
- SQLite 完整快照切换；扫描期间继续显示上一次完整结果
- 动态历史降采样：7 天内每日、31 天内每周、1 年内每月、更早每年
- 四套白色背景多彩主题，桌面优先并提供安全的移动端降级

Disk Insight 不读取文件内容、不分析压缩包内部、不计算哈希、不跟随符号链接，并默认跳过隐藏文件。文件大类通过文件名后缀映射，避免在 HDD 上产生额外内容读取。

## 快速部署

镜像只发布明确版本，不提供 `latest` 标签。

1. 编辑 [`compose.yaml`](compose.yaml)，把 `/path/to/your/data` 改为宿主机目录。
2. 启动：

   ```sh
   docker compose up -d
   ```

3. 打开 `http://localhost:8080`。

数据库保存在 `disk-insight-state` 卷中。扫描目录使用 `:ro` 挂载，容器根文件系统为只读，且默认丢弃全部 Linux capabilities。

### 多个独立目录

```yaml
services:
  disk-insight:
    image: ghcr.io/ccawmiku/disk-insight:v1.0.3
    environment:
      DISK_INSIGHT_ROOTS: "/scan/photos::Photos;/scan/backups::Backups"
    volumes:
      - /mnt/photos:/scan/photos:ro
      - /mnt/backups:/scan/backups:ro
      - disk-insight-state:/var/lib/disk-insight
```

新增宿主机挂载必须修改 Compose 并重新创建容器。网页只能调整已经挂入容器的扫描计划、配色、语言和排除规则；应用不会访问 Docker Socket。

### 一个共同父目录

将父目录映射到 `/data:ro`，并保持默认的 `DISK_INSIGHT_ROOTS=/data::Data`。选择 `/data` 下任意子目录时，所有图表会切换到该子树的递归统计。

## 配置

| 环境变量 | 默认值 | 说明 |
| --- | --- | --- |
| `DISK_INSIGHT_ADDRESS` | `:8080` | HTTP 监听地址 |
| `DISK_INSIGHT_DATABASE` | `/var/lib/disk-insight/disk-insight.db` | SQLite 数据库路径 |
| `DISK_INSIGHT_WEB` | `/opt/disk-insight/web` | 前端静态文件路径 |
| `DISK_INSIGHT_ROOTS` | `/data::Data` | `路径::显示名`，多个根用分号分隔 |

应用没有内置登录系统，设计目标是可信本机或局域网。不要直接暴露到公网；如需远程访问，请在前面配置带身份认证和 TLS 的反向代理。

## 扫描与性能

- 扫描器单路、流式遍历，不把完整文件列表一次性载入内存。
- 文件和目录以小批次写入 SQLite WAL，降低频繁同步写入。
- 容器进程默认以较低 CPU 调度优先级运行。
- 第一次扫描无法在不额外遍历磁盘的情况下知道总文件数，因此显示阶段进度、精确计数和速度；后续扫描根据上一完整快照提供明确标注的估算百分比和 ETA。
- 扫描成功后以事务切换 `current_scan_id`，失败或取消不会替换可见快照。
- 图表查询使用路径索引和有界内存缓存；目录树按需加载并虚拟化渲染。

## 原生开发与测试

要求 Go 1.26.5、Node.js 24.15.0 和 npm。

```sh
go mod verify
go vet ./...
go test ./...
go build ./cmd/disk-insight

cd web
npm ci
npm run typecheck
npm run lint
npm run test
npm run build
```

设置开发环境变量后运行后端，Vite 会把 `/api` 代理到 `127.0.0.1:8080`：

```sh
DISK_INSIGHT_ROOTS="/path/to/data::Data" go run ./cmd/disk-insight
cd web && npm run dev
```

## 构建与发布

本项目的容器只在 GitHub Actions 中构建和测试：

1. Ubuntu 原生运行 Go race tests、vet、前端类型检查、lint、单元测试和生产构建。
2. Chromium 在桌面和移动视口运行端到端测试。
3. 标签构建仅生成 `linux/amd64` 镜像。
4. 容器以只读根文件系统、无 capabilities 和只读数据挂载完成 smoke test。
5. 通过后发布到 `ghcr.io/ccawmiku/disk-insight:<version>`。

每个提交都有唯一的 annotated semantic-version tag；首个版本为 `v1.0.0`。

## Architecture

- **Backend:** Go standard HTTP server and low-impact filesystem scanner
- **Persistence:** SQLite in WAL mode with immutable completed-scan snapshots
- **Frontend:** React, TypeScript, shadcn-style accessible controls, Tailwind CSS, Apache ECharts
- **Visualization:** linked size distribution, cumulative curves, Waffle, age distribution, Treemap, ranking, and history
- **Delivery:** one non-root Docker image, GitHub Actions, public GHCR

## 安全边界

应用只返回扫描根目录内的相对路径，不返回宿主机绝对路径，也不提供文件预览、打开或下载接口。更多信息见 [`SECURITY.md`](SECURITY.md)。
