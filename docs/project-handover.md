# biu-panel 项目交接文档

> 生成时间：2026-06-10  
> 依据来源：实际代码、当前 Docker 运行状态、SQLite 数据库结构、项目文档和本次验证命令。  
> 项目路径：`/project/panel`

## 1. 项目概述

### 项目名称

`biu-panel`

### 项目目标

`biu-panel` 是一个个人自用的轻量级 BS 端导航站 + 网页收藏夹系统。首页体验参考 Sun-Panel 的分组导航卡片模式，但项目不是 Sun-Panel 的二次开发；它是独立实现的 Go + Vue + SQLite 应用。

### 解决什么问题

- 将常用网站、家庭服务、NAS/内网服务、公网服务整理成一个快速打开的导航页。
- 在导航页之外提供网页收藏夹能力，支持目录结构、网址收藏、备注、模糊搜索。
- 面向单人自用，追求轻量、响应快、部署简单、备份方便。
- 收藏夹抽屉按需加载，避免首页初始化时加载大量收藏数据。

### 当前开发阶段

项目保留 Docker 交付能力，但当前调试以本地直接运行后端和前端为主，不强制依赖容器。

当前阶段更接近 **V1.0 功能迭代/打磨阶段**，不是纯原型阶段。核心功能已实现，但仍存在不少功能缺口、技术债务和体验问题，详见后文“当前未完成功能”和“已知问题”。

### 是否有版本管理

有 Git 仓库。

当前分支：`master`

最近 5 个提交：

```text
7f928d7 Add personalized search dashboard
f9a66af Redesign for efficient neutral dashboard
8bde5f1 Improve settings management interactions
a782796 Make settings menu switch content
3c8e7c0 Add settings center modal
```

当前工作区存在未提交修改：

```text
M backend/internal/httpx/server.go
M frontend/src/App.vue
M frontend/src/style.css
```

说明：以上修改是当前实际工作区状态的一部分。新的 AI 接手前应先阅读这些文件和 `git diff`，不要随意回滚。

## 2. 技术架构

### 前端技术栈

- Vue 3
- Vite
- npm
- 单文件主应用：`frontend/src/App.vue`
- 样式集中在：`frontend/src/style.css`
- API 封装：`frontend/src/lib/api.js`

`frontend/package.json`：

```json
{
  "dependencies": {
    "vue": "^3.5.35"
  },
  "devDependencies": {
    "@vitejs/plugin-vue": "^6.0.7",
    "vite": "^8.0.16"
  }
}
```

### 后端技术栈

- Go 1.25
- 标准库 `net/http` 路由，未使用 Gin/Echo 等 Web 框架
- SQLite 驱动：`modernc.org/sqlite`
- 密码哈希：`golang.org/x/crypto/bcrypt`
- 后端入口：`backend/cmd/server/main.go`
- HTTP 处理：`backend/internal/httpx/server.go`
- 数据访问和迁移：`backend/internal/store/store.go`
- 配置加载：`backend/internal/config/config.go`

### 数据库

- SQLite
- 当前 Docker 默认数据库路径：`/app/data/db/biu-panel.db`
- 宿主机实际路径：`/project/panel/data/db/biu-panel.db`
- 数据库由后端启动时自动创建并执行 `CREATE TABLE IF NOT EXISTS` 迁移。

当前实际数据库文件存在，大小约 `61440` 字节。

当前主要表及记录数（生成文档时）：

```text
assets rows=0
backup_records rows=0
bookmark_folders rows=0
bookmarks rows=0
login_logs rows=26
nav_groups rows=6
nav_items rows=30
sessions rows=26
settings rows=22
storage_configs rows=0
users rows=1
```

### Docker 部署结构

Docker 仍保留作为最终交付方式；本地调试时无需依赖容器。

项目提供：

- `Dockerfile`
- `docker-compose.yml`
- `.dockerignore`

当前运行容器：

```text
NAMES       IMAGE              STATUS                    PORTS
biu-panel   biu-panel:latest   Up 43 minutes (healthy)   0.0.0.0:55088->55088/tcp, [::]:55088->55088/tcp
```

当前镜像 ID：

```text
sha256:7b227e3cd958eaddf0f680917575a45c06b72cedd7e28486edab8e59d19988d2
```

当前挂载：

```text
/project/panel/data -> /app/data
```

当前环境变量：

```text
BIU_PANEL_DATA_DIR=/app/data
BIU_PANEL_STATIC_DIR=/app/public
BIU_PANEL_PORT=55088
```

`Dockerfile` 是多阶段构建：

1. `node:22-alpine` 构建前端。
2. `golang:1.25-alpine` 构建 Go 后端二进制。
3. `alpine:3.21` 作为最终运行镜像。

最终容器：

- 工作目录：`/app`
- 后端二进制：`/app/biu-panel`
- 前端静态文件：`/app/public`
- 数据目录：`/app/data`
- 健康检查：`GET http://127.0.0.1:55088/api/health`

### 第三方依赖

后端 `go.mod` 主要依赖：

- `modernc.org/sqlite`：纯 Go SQLite 驱动。
- `golang.org/x/crypto`：用于 bcrypt 密码哈希。
- 其他为 SQLite 驱动的间接依赖。

前端依赖：

- `vue`
- `vite`
- `@vitejs/plugin-vue`

运行时外部服务：

- 无强制外部服务。
- 可选 S3 兼容对象存储，用于图片上传和备份上传。

## 3. 项目目录结构

实际目录概览（已排除 `.git`、`node_modules`、`dist`、`data` 内部细节）：

```text
/project/panel/
  AGENTS.md
  Dockerfile
  README.md
  docker-compose.yml
  .dockerignore
  .gitignore
  .node-version
  .nvmrc
  backend/
    go.mod
    go.sum
    cmd/server/main.go
    internal/config/config.go
    internal/httpx/server.go
    internal/store/store.go
  frontend/
    index.html
    package.json
    package-lock.json
    vite.config.js
    src/App.vue
    src/main.js
    src/style.css
    src/lib/api.js
    public/favicon.svg
    public/icons.svg
  docs/
    development-v0.1.md
    release-checklist.md
    requirements-v0.1.md
    tasks-v1.0-v0.1.md
    project-handover.md
  scripts/
    smoke.sh
  data/
    db/biu-panel.db
    db/biu-panel.db.bak-1781005279
```

目录用途：

- `backend/`：Go 后端服务。
- `backend/cmd/server/`：服务启动入口。
- `backend/internal/config/`：环境变量配置加载。
- `backend/internal/httpx/`：HTTP 路由、认证、API、静态文件、上传、备份、S3、导入导出等处理逻辑。
- `backend/internal/store/`：SQLite 连接、建表迁移、基础 CRUD。
- `frontend/`：Vue 前端项目。
- `frontend/src/App.vue`：当前几乎所有前端状态、页面和交互逻辑都集中在这个文件。
- `frontend/src/style.css`：全局样式。
- `frontend/src/lib/api.js`：前端请求后端 API 的封装。
- `docs/`：需求、开发说明、发布清单和本交接文档。
- `scripts/`：辅助脚本，目前有 `smoke.sh`。
- `data/`：运行数据目录，包含 SQLite 数据库、上传文件、备份等。不要提交到 Git。
- `deploy/`：当前为空或未见关键内容，待确认是否后续用于部署材料。

## 4. 当前已实现功能

以下内容基于实际代码和当前运行状态确认。

### 认证与初始化

- 首次初始化管理员账号：`POST /api/setup`
- 初始化状态检查：`GET /api/setup/status`
- 登录：`POST /api/auth/login`
- 登出：`POST /api/auth/logout`
- 当前用户：`GET /api/auth/me`
- Cookie 会话：`biu_session`
- 默认浏览器会话；勾选“记住登录”后有效期 30 天。
- 登录日志：`login_logs`
- 失败锁定：15 分钟内失败次数达到 5 次后锁定。
- 支持环境变量首次自动创建管理员：`BIU_PANEL_ADMIN_USER`、`BIU_PANEL_ADMIN_PASSWORD`。

### 首页导航

- 首页分组展示。
- 导航卡片展示。
- 导航分组新增、编辑、删除。
- 导航卡片新增、编辑、删除。
- 分组排序、卡片排序：前端通过拖拽交换 sort 后保存。
- 卡片支持：标题、图标/文本、内网地址、公网地址、打开模式、所属分组、排序。
- 新增/编辑卡片弹窗近期已调整：
  - 登录页默认账号不再预填 `admin`。
  - “名称”改为“标题”。
  - “外网地址”改为“公网地址”。
  - 标题、公网地址、分组作为前端必填项。
  - 图标输入改为“文字 / 图片”互斥模式。
  - 文字模式限制 5 个字。
  - 图片模式支持输入图片 URL 或上传。
  - 自动抓取标题/图标按钮放在内网地址、公网地址右侧。
- 卡片支持当前标签页打开。
- 卡片右键菜单包含打开、编辑、删除等。
- 右下角网络切换按钮：当未开启自动检测时显示“内网优先 / 外网优先”。
- 自动检测开关存在于设置项中，但实际自动检测逻辑实现不足，详见“已知问题”。

### 收藏夹

- 左侧收藏夹抽屉。
- 抽屉打开时加载文件夹。
- 文件夹新增、编辑、删除。
- 收藏新增、编辑、删除。
- 收藏字段：标题、URL、favicon、备注、排序、所属文件夹。
- 收藏搜索：`GET /api/bookmark/search?q=...`，搜索标题、URL、备注。
- 收藏导入：`POST /api/bookmark/import`，解析 Netscape Bookmark HTML 格式。
- 收藏导出：`GET /api/bookmark/export`，导出 HTML。
- 收藏和文件夹拖拽排序相关代码存在。
- 收藏可跨文件夹拖拽移动到目标文件夹。

### 元数据抓取

- 后端接口：`GET /api/metadata?url=...`
- 支持抓取网页 `<title>`。
- 支持抓取 favicon / shortcut icon / apple-touch-icon，失败则回退为 `https://host/favicon.ico`。
- 前端可将抓取结果填入导航卡片、收藏等表单。

### 图片上传与静态文件

- 本地图片上传接口：`POST /api/assets/upload`
- 限制上传表单最大约 8MB。
- 仅允许 `image/*`。
- 本地文件存放在数据目录的 `uploads/` 下。
- `/uploads/...` 由后端从数据目录读取并返回。
- 若设置启用 S3，则上传本地后尝试同步到 S3，成功时返回 S3 public URL；失败时保留本地上传路径。
- 上传记录写入 `assets` 表。

### 设置中心

- 设置读取：`GET /api/settings`
- 设置保存：`PUT /api/settings`
- 当前允许保存的设置字段包括：
  - `siteTitle`
  - `logoUrl`
  - `showTitle`
  - `showLogo`
  - `showClock`
  - `showSeconds`
  - `showSearch`
  - `searchEngines`
  - `backgroundUrl`
  - `backgroundColor`
  - `lanDetectUrl`
  - `lanDetectTimeout`
  - `autoDetectLan`
  - `s3Endpoint`
  - `s3Region`
  - `s3Bucket`
  - `s3AccessKey`
  - `s3SecretKey`
  - `s3Prefix`
  - `s3PathStyle`
  - `s3Enabled`
  - `s3PublicBase`
- 设置页保存后显示“设置已保存”。
- 关闭设置页会清空保存提示，重新打开不会残留旧提示。

### 搜索

- 首页存在网页搜索框。
- 默认搜索引擎配置在前端 `settingsForm.searchEngines` 中，包含 Google、百度、Bing。
- 支持通过 `prompt()` 添加搜索引擎，但体验比较粗糙。

### 备份与恢复

- 下载备份：`GET /api/backup/download`
- 恢复备份：`POST /api/backup/restore`
- S3 备份：`POST /api/backup/s3`
- S3 测试：`POST /api/s3/test`
- 备份格式为 `.tar.gz`，包含 `data/` 下文件。
- 恢复会将备份包中 `data/` 前缀下的文件写回数据目录。

### Docker 与验证

已验证命令：

```bash
cd /project/panel/frontend && npm run build
cd /project/panel/backend && go test ./...
cd /project/panel && docker compose build biu-panel
cd /project/panel && docker compose up -d --force-recreate biu-panel
curl -fsS http://127.0.0.1:55088/api/health
```

健康检查当前通过：

```json
{"data":{"app":"biu-panel","status":"ok"},"success":true}
```

## 5. 当前未完成功能

以下是根据需求文档、AGENTS.md 和实际代码对照得出的未完成项，不应当声称已经完成。

### 首页导航相关

- 分组折叠/展开字段存在于数据库 `nav_groups.collapsed`，但前端未看到完整折叠/展开交互落地。
- 自动判断内外网：设置字段存在，但未看到实际对 `lanDetectUrl` 执行探测并自动切换网络模式的完整实现。
- 单卡片独立检测地址未实现；当前 NavItem 只有 `lanUrl`、`wanUrl`、`urlMode`。
- 后端未强制校验导航卡片公网地址必填；目前主要是前端校验。
- 主页卡片“从收藏夹设为首页卡片”未完整实现。

### 收藏夹相关

- 无限层级数据结构已支持 `parent_id`，但当前前端目录树只展示顶层和基础状态，按需展开子目录体验不完整。
- 收藏夹批量删除未实现。
- 收藏夹批量移动未实现。
- 收藏夹菜单需求中的“复制链接”“移动到文件夹”“设为首页卡片 / 取消首页卡片”“批量选择入口”未完整实现。
- 搜索结果当前 `Path` 仅取直接文件夹名，未构造完整路径。
- 导入 HTML 未做重复检测和跳过逻辑。
- 导入结果只返回 folders/bookmarks 数量，没有跳过/失败明细。
- 导入书签备注 `<DD>` 解析不完整；代码中 `lastBookmarkID` 未实际用于写入备注。
- Safari/Chrome 兼容程度需要真实样例进一步验证。

### 设置和资源

- `storage_configs` 表存在，但实际 S3 配置保存在 `settings` 表中，`storage_configs` 未使用。
- `backup_records` 表存在，但备份下载/S3 上传未明显写入备份记录，当前记录数为 0。
- 上传限制文档提到普通图标/Logo 5MB，实际代码是 8MB，存在不一致。
- 背景图、Logo 上传入口是否完整好用需要进一步人工验证；代码中复用图片字段，但 UI 体验待确认。

### 测试

- Go 包当前无测试文件，`go test ./...` 只是确认能编译。
- 没有前端自动化测试。
- 没有书签导入导出的样例测试。
- 没有备份恢复自动化测试。
- 没有内外网 URL 选择逻辑测试。

### 文档/发布

- README 已有基本说明，但与实际状态和部分需求存在差异。
- Docker volume 目前是 `./data:/app/data` 单挂载；AGENTS.md 中曾规划拆分为 `./data/db`、`./data/uploads`、`./data/backups` 三个挂载，但实际未采用。
- `docs/development-v0.1.md` 部分内容仍停留在早期“静态原型/API 初版”描述，已经落后于代码现状。

## 6. 已知问题

### 代码结构问题

- `frontend/src/App.vue` 过大，当前接近 900 行，包含大量状态、业务逻辑和模板，维护成本较高。
- `backend/internal/httpx/server.go` 过大，当前约 1180 行，认证、导航、收藏、设置、上传、S3、备份、导入导出全部集中在一个文件。
- 后端 store 层只是基础 SQL 封装，没有领域边界和事务封装。
- 前端仍有多个 `prompt()` 流程，例如新增搜索引擎、部分新建分组/文件夹/收藏逻辑，体验不统一。

### 安全与权限问题

- `GET /api/navigation`、`GET /api/bookmark/folders`、`GET /api/bookmarks`、`GET /api/bookmark/search` 等读取接口中，部分未强制登录校验。对于个人内网使用可能可接受，但如果公网暴露，应重新评估。
- S3 Secret Key 明文存储在 SQLite `settings` 表中；这是 V1 设计选择，但要注意备份和数据库文件权限。
- CORS 固定允许 `http://localhost:5173`，适合本地开发；生产主要走同源静态服务。若未来前后端分离部署需调整。
- 上传文件通过 MIME 判断 `image/*`，未做更严格的文件内容校验。

### 数据和迁移问题

- 迁移方式只有 `CREATE TABLE IF NOT EXISTS`，没有版本化 migration。后续修改表结构要谨慎，需手动兼容旧库。
- `sessions` 表已有 26 条记录，未见会话清理任务，长期运行可能积累历史会话。
- `login_logs` 也会持续增长，未见清理策略。

### 功能缺陷/设计债

- 自动内网检测设置存在，但实际自动探测逻辑不足。
- 收藏夹无限层级的后端结构存在，前端树形交互不足。
- 书签导入解析使用正则，面对复杂浏览器导出 HTML 可能不够稳健。
- 备份恢复会直接覆盖数据目录文件，前端是否有二次确认待确认；后端没有版本兼容检查。
- 后端导航卡片接口报错文案仍使用“卡片名称”，前端已改为“标题”，文案不一致。
- 前端新增/编辑卡片要求公网地址必填，但后端没有同步强约束，API 直接调用可绕过。

## 7. 数据库结构

数据库文件：`/project/panel/data/db/biu-panel.db`

后端迁移定义位置：`backend/internal/store/store.go`

### `users`

用途：管理员账号。

字段：

- `id INTEGER PRIMARY KEY`
- `username TEXT NOT NULL UNIQUE`
- `password_hash TEXT NOT NULL`
- `created_at TEXT NOT NULL`

### `sessions`

用途：登录会话。

字段：

- `token TEXT PRIMARY KEY`
- `user_id INTEGER NOT NULL`
- `expires_at TEXT`
- `remember INTEGER NOT NULL DEFAULT 0`
- `created_at TEXT NOT NULL`

### `login_logs`

用途：记录登录成功/失败、锁定判断依据。

字段：

- `id INTEGER PRIMARY KEY`
- `username TEXT NOT NULL`
- `success INTEGER NOT NULL`
- `ip TEXT`
- `message TEXT`
- `created_at TEXT NOT NULL`

### `settings`

用途：系统设置 Key-Value 存储。

字段：

- `key TEXT PRIMARY KEY`
- `value TEXT NOT NULL`

### `nav_groups`

用途：首页导航分组。

字段：

- `id INTEGER PRIMARY KEY`
- `name TEXT NOT NULL`
- `sort INTEGER NOT NULL DEFAULT 0`
- `collapsed INTEGER NOT NULL DEFAULT 0`

### `nav_items`

用途：首页导航卡片。

字段：

- `id INTEGER PRIMARY KEY`
- `group_id INTEGER NOT NULL`
- `name TEXT NOT NULL`
- `icon TEXT`
- `lan_url TEXT`
- `wan_url TEXT`
- `url_mode TEXT NOT NULL DEFAULT 'auto'`
- `sort INTEGER NOT NULL DEFAULT 0`
- 外键：`group_id` -> `nav_groups(id)`，删除分组时级联删除卡片。

### `bookmark_folders`

用途：收藏夹文件夹，支持无限层级。

字段：

- `id INTEGER PRIMARY KEY`
- `parent_id INTEGER`
- `name TEXT NOT NULL`
- `sort INTEGER NOT NULL DEFAULT 0`
- 外键：`parent_id` -> `bookmark_folders(id)`，删除父文件夹时级联删除子文件夹。

### `bookmarks`

用途：收藏夹网址。

字段：

- `id INTEGER PRIMARY KEY`
- `folder_id INTEGER NOT NULL`
- `title TEXT NOT NULL`
- `url TEXT NOT NULL`
- `favicon TEXT`
- `note TEXT`
- `sort INTEGER NOT NULL DEFAULT 0`
- 外键：`folder_id` -> `bookmark_folders(id)`，删除文件夹时级联删除收藏。

### `assets`

用途：记录上传的图片资源。

字段：

- `id INTEGER PRIMARY KEY`
- `name TEXT`
- `source TEXT NOT NULL`
- `path TEXT NOT NULL`
- `mime TEXT`
- `size INTEGER`
- `created_at TEXT NOT NULL`

### `storage_configs`

用途：设计上用于存储外部存储配置。

当前状态：表存在，但实际 S3 配置保存在 `settings` 表，此表未使用。

字段：

- `id INTEGER PRIMARY KEY`
- `kind TEXT NOT NULL`
- `name TEXT NOT NULL`
- `config_json TEXT NOT NULL`
- `active INTEGER NOT NULL DEFAULT 0`

### `backup_records`

用途：设计上用于记录备份。

当前状态：表存在，但当前备份逻辑未明显写入记录。

字段：

- `id INTEGER PRIMARY KEY`
- `file_name TEXT NOT NULL`
- `target TEXT NOT NULL`
- `status TEXT NOT NULL`
- `created_at TEXT NOT NULL`

## 8. 配置文件说明

### `backend/internal/config/config.go`

负责读取环境变量。

支持变量：

| 变量 | 默认值 | 作用 |
| --- | --- | --- |
| `BIU_PANEL_PORT` | `55088` | HTTP 服务端口 |
| `BIU_PANEL_DATA_DIR` | `../data` | 数据目录 |
| `BIU_PANEL_DB` | `$BIU_PANEL_DATA_DIR/db/biu-panel.db` | SQLite 数据库路径 |
| `BIU_PANEL_STATIC_DIR` | `./public` | 前端静态文件目录 |
| `BIU_PANEL_ADMIN_USER` | 空 | 首次启动自动创建管理员账号 |
| `BIU_PANEL_ADMIN_PASSWORD` | 空 | 首次启动自动创建管理员密码 |

### `docker-compose.yml`

当前 Compose 服务：`biu-panel`

关键配置：

```yaml
services:
  biu-panel:
    build: .
    image: biu-panel:latest
    container_name: biu-panel
    restart: unless-stopped
    ports:
      - "55088:55088"
    environment:
      BIU_PANEL_PORT: "55088"
      BIU_PANEL_DATA_DIR: /app/data
      BIU_PANEL_STATIC_DIR: /app/public
    volumes:
      - ./data:/app/data
```

### `.dockerignore`

排除了：

- `data`
- `frontend/node_modules`
- `frontend/dist`
- `backend/tmp`
- `.git`
- `.gitignore`
- `Dockerfile`
- `README.md`
- `*.log`

注意：`.dockerignore` 排除了 `Dockerfile` 自身通常不影响构建，但该配置比较少见。若后续构建上下文需要 README 或其他文档，应调整。

### 前端环境变量

`frontend/src/lib/api.js` 使用：

```js
const API_BASE = import.meta.env.VITE_API_BASE || ''
```

默认同源调用后端 API。本地前后端分离开发时可设置 `VITE_API_BASE`。

### 设置项

系统设置保存在 SQLite 的 `settings` 表中，通过 `/api/settings` 读写。S3 密钥、站点标题、Logo、背景、搜索引擎、内网检测配置等都在这里。

## 9. 部署说明

### 当前访问地址

服务器外部访问：

```text
http://111.119.213.77:55088/
```

本机验证：

```text
http://127.0.0.1:55088/
```

### 启动

推荐：

```bash
cd /project/panel
docker compose up -d --build
```

若镜像已构建：

```bash
cd /project/panel
docker compose up -d
```

### 停止

```bash
cd /project/panel
docker compose stop biu-panel
```

或停止并移除容器：

```bash
cd /project/panel
docker compose down
```

### 更新/重新部署

推荐流程：

```bash
cd /project/panel
cd frontend && npm run build
cd /project/panel/backend && go test ./...
cd /project/panel
docker compose build biu-panel
docker compose up -d --force-recreate biu-panel
curl -fsS http://127.0.0.1:55088/api/health
```

注意：本工具环境会将 `docker compose up -d` 识别为可能长驻命令，使用自动化工具执行时可能需要后台运行并等待完成。

### 查看状态

```bash
docker ps --filter name=biu-panel
curl -fsS http://127.0.0.1:55088/api/health
```

### 查看日志

```bash
docker logs --tail=200 biu-panel
```

### 备份

最直接备份方式：备份整个数据目录。

```bash
cd /project/panel
tar -czf /root/biu-panel-data-$(date +%Y%m%d-%H%M%S).tar.gz data
```

应用内也支持登录后通过接口下载备份：

```text
GET /api/backup/download
```

但该接口需要登录 Cookie。

### 恢复

手动恢复数据目录前建议：

1. 停止容器。
2. 备份当前 `data/` 目录。
3. 解压旧备份覆盖 `data/`。
4. 重启容器。
5. 检查健康状态。

示例：

```bash
cd /project/panel
docker compose stop biu-panel
tar -czf /root/biu-panel-data-before-restore-$(date +%Y%m%d-%H%M%S).tar.gz data
# 解压目标备份到 /project/panel/data，具体命令取决于备份包结构
docker compose up -d biu-panel
curl -fsS http://127.0.0.1:55088/api/health
```

应用内恢复接口：

```text
POST /api/backup/restore
```

该接口会写入数据目录，使用前必须确认备份来源可信。

## 10. 最近开发记录

最近实际代码修改和部署包括：

- 登录页账号输入框默认值从 `admin` 改为空。
- 新增/编辑卡片弹窗调整：
  - “名称”改为“标题”。
  - “外网地址”改为“公网地址”。
  - 标题、公网地址、分组增加前端必填校验和红色星号。
  - 增加分组下拉框，默认使用点击新增时所在分组，也允许手动选择。
  - 图标输入改为“文字 / 图片”互斥模式。
  - 文字模式提示“请输入文本内容”，限制 5 字。
  - 图片模式提示“输入图标地址或上传”，右侧显示上传按钮。
  - 自动抓取标题/图标按钮移动到内网地址和公网地址右侧。
- 卡片背景相关调整：外层卡片和图片图标容器背景改为透明。
- 设置页保存提示调整：关闭设置页时清空“设置已保存”，重新打开不残留。
- 右下角网络切换按钮调整：在自动检测开启时隐藏手动切换按钮；关闭自动检测时显示内网/公网图标切换。
- 首页分组名右侧新增/编辑按钮布局调整。
- 新增卡片从 prompt 流程改为弹窗流程。
- 编辑分组状态下卡片可点击编辑，并支持拖拽排序。
- 手机端卡片布局保持一行 5 个。

最近验证：

```text
frontend npm run build: 通过
go test ./...: 通过（无测试文件，仅编译）
docker compose build biu-panel: 通过
docker compose up -d --force-recreate biu-panel: 已完成
/api/health: 通过
```

## 11. 后续开发建议

建议按以下优先级推进。

### P0：先做稳定性和一致性

1. 为后端接口补齐与前端一致的校验：标题、公网地址、分组必填。
2. 修正后端错误文案中的“卡片名称”为“标题”。
3. 补一个最小后端测试集：认证、导航 CRUD、收藏 CRUD、设置保存、metadata URL 校验。
4. 建立数据库迁移版本机制，避免后续表结构变更困难。
5. 清理过期 sessions 和历史 login_logs 的策略。

### P1：拆分大文件

1. 将 `frontend/src/App.vue` 拆分为组件：登录页、首页、导航分组、卡片弹窗、收藏夹抽屉、设置中心。
2. 将 `backend/internal/httpx/server.go` 拆分为 auth、navigation、bookmarks、settings、assets、backup、s3、metadata 等文件或 handler。
3. 将前端业务逻辑抽到 composables 或简单模块中，降低单文件复杂度。

### P2：补齐收藏夹核心

1. 完善多级目录树展开/折叠和按需加载。
2. 实现收藏夹批量删除、批量移动。
3. 实现复制链接、移动到文件夹、设为首页卡片。
4. 搜索结果显示完整路径，而不是仅直接文件夹名。
5. 改进书签导入：支持 `<DD>` 备注、重复检测、跳过统计、失败统计。

### P3：补齐内外网自动判断

1. 实现 `lanDetectUrl` + `lanDetectTimeout` 的实际探测逻辑。
2. 明确自动检测应在前端执行还是后端执行。个人浏览器环境下，前端探测内网地址更接近实际访问场景。
3. 若要单卡片覆盖检测地址，需要扩展 `nav_items` 表。

### P4：发布质量

1. 编写真实的 release checklist，替换过时文档。
2. 增加 Playwright 或最小浏览器自动化测试覆盖登录、导航卡片弹窗、设置保存。
3. 明确 Docker volume 是否继续单挂载 `./data:/app/data`，还是改为分目录挂载。
4. 增加版本号展示和备份包版本 metadata。

## 12. 风险与注意事项

### 不要随意修改/删除的数据

- 不要删除 `/project/panel/data/`。
- 不要删除 `/project/panel/data/db/biu-panel.db`。
- 不要直接编辑 SQLite 数据库，除非先备份。
- 不要提交数据库、上传文件、备份文件、S3 密钥。

### 接手开发前必须检查

```bash
cd /project/panel
git status --short
git diff -- backend/internal/httpx/server.go frontend/src/App.vue frontend/src/style.css
```

当前存在未提交修改，新的 AI 不应回滚这些改动，除非用户明确要求。

### 容易出问题的模块

- `frontend/src/App.vue`：逻辑高度集中，改一个弹窗可能影响多个状态。
- `backend/internal/httpx/server.go`：功能过于集中，改导入/备份/S3 时可能影响静态服务或认证。
- 书签导入解析：正则解析 HTML 脆弱，需要真实样例验证。
- 备份恢复：会写数据目录，操作前必须备份。
- S3 上传：手写 AWS Signature V4，改动时必须用真实 S3/OSS 测试。
- Docker 数据挂载：当前容器以 root 运行，简化了 bind mount 权限；改非 root 用户会涉及数据目录权限。

### 兼容性要求

- UI 文案默认中文。
- 面向单管理员，不要引入复杂多用户权限系统，除非用户明确要求。
- 首页必须保持轻量，不应在初始加载时拉取完整收藏夹树和所有收藏。
- Docker 默认端口保持 `55088`。
- HTTPS 由反向代理负责，应用本身只提供 HTTP。
- 不优先支持子路径部署。
- 个人自用优先，避免引入过重框架和复杂服务。

### 待确认事项

- 是否需要将当前未提交修改做一次 Git commit。
- 是否继续采用单目录挂载 `./data:/app/data`，还是按早期规划拆分为 `db/uploads/backups` 三个 volume。
- 是否需要公网访问时强制所有读取接口登录。
- 是否需要保留首页网页搜索功能；早期需求文档曾写“不做首页搜索”，但当前代码已实现网页搜索。
- 是否需要为当前项目写正式版本号和升级策略。

## 13. 快速接手命令清单

```bash
# 进入项目
cd /project/panel

# 查看当前改动
git status --short
git diff

# 后端编译检查
cd /project/panel/backend
go test ./...

# 前端构建
cd /project/panel/frontend
npm run build

# Docker 重建部署
cd /project/panel
docker compose build biu-panel
docker compose up -d --force-recreate biu-panel

# 健康检查
curl -fsS http://127.0.0.1:55088/api/health

# 查看运行状态
docker ps --filter name=biu-panel

# 查看日志
docker logs --tail=200 biu-panel
```

## 14. 验证记录

本交接文档生成前执行过以下实际检查：

```text
读取 Dockerfile、docker-compose.yml、README、AGENTS.md、docs、前后端源码。
检查 git status 和最近提交。
检查当前 Docker 容器状态、镜像、挂载、环境变量。
检查 SQLite 实际表结构和表记录数。
执行 go test ./...：通过，但无测试文件。
执行 curl /api/health：通过。
确认静态前端首页可返回 HTML。
```

注意：`pygount` 未安装，因此代码规模统计使用 Python 脚本粗略统计；统计排除了 `.git`、`node_modules`、`dist`、`data`。

粗略代码规模：

```text
.go    files=4  lines=1640
.vue   files=1  lines=899
.css   files=1  lines=225
.js    files=3  lines=81
.md    files=6  lines=1041
```
