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

截至 2026-06-15，导航页相关功能已进入用户确认可用状态：导航分组和卡片的新增、编辑、删除、排序、系统设置草稿保存、内外网手动优先级切换、右键菜单打开、导航页单独备份和恢复均已完成阶段性验证。

### 是否有版本管理

有 Git 仓库。

当前分支：`master`

最近 5 个提交：

```text
47de4cf Split frontend display components
30fa80e Stabilize navigation persistence flow
551bc22 Refine navigation and settings UI
7f928d7 Add personalized search dashboard
f9a66af Redesign for efficient neutral dashboard
```

当前工作区存在导航页备份恢复、设置页样式和右键打开修复相关的未提交修改。新的 AI 接手前应先阅读 `git diff`，不要随意回滚。

## 2. 技术架构

### 前端技术栈

- Vue 3
- Vite
- npm
- 主应用入口：`frontend/src/App.vue`
- 展示型组件目录：`frontend/src/components/`
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
- 可选 S3 兼容对象存储，用于图片上传；备份上传入口已按当前方向移除。

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
    internal/httpx/auth.go
    internal/httpx/bookmarks.go
    internal/httpx/health.go
    internal/httpx/navigation.go
    internal/httpx/server.go
    internal/httpx/settings.go
    internal/store/store.go
  frontend/
    index.html
    package.json
    package-lock.json
    vite.config.js
    src/App.vue
    src/components/BackupRestoreSection.vue
    src/components/BookmarkFolderTreeNode.vue
    src/components/BookmarkRow.vue
    src/components/ContextMenu.vue
    src/components/FloatingActions.vue
    src/components/HomeHero.vue
    src/components/MoveDialog.vue
    src/components/NavDragFloat.vue
    src/components/PersonalSettingsForm.vue
    src/components/SearchEngineManagerSection.vue
    src/components/SettingsMenu.vue
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
- `backend/internal/httpx/`：HTTP 路由、认证、API、静态文件、上传、备份、S3、导入导出等处理逻辑。已完成后端 HTTP handler 保守拆分整理：`auth.go` 处理认证与会话，`navigation.go` 处理导航页基础接口，`bookmarks.go` 处理收藏夹基础接口，`health.go` 处理健康检查，`settings.go` 处理设置相关接口；`server.go` 继续保留路由注册、通用工具、备份恢复、S3、静态资源上传、metadata、书签导入导出等尚未拆分逻辑。
- `backend/internal/store/`：SQLite 连接、建表迁移、基础 CRUD。
- `frontend/`：Vue 前端项目。
- `frontend/src/App.vue`：当前保留全局布局、主要状态、组件组合、顶层事件处理和核心业务逻辑。
- `frontend/src/components/`：已承载部分展示型组件和设置页区块，例如首页 Hero、右键菜单、移动弹窗、收藏行、设置菜单、拖拽浮层、备份恢复、搜索引擎管理、个性化设置表单等。
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
- **定期清理过期 Session**：服务启动时和每 24 小时自动清理已过期的 sessions。
- 登录日志：`login_logs`
- 失败锁定：15 分钟内失败次数达到 5 次后锁定。
- **定期清理陈旧登录日志**：服务启动时和每 24 小时自动清理 30 天前的 `login_logs` 记录（不影响 15 分钟内的防暴力破解锁定机制）。
- 支持环境变量首次自动创建管理员：`BIU_PANEL_ADMIN_USER`、`BIU_PANEL_ADMIN_PASSWORD`。

### 首页导航

- 首页分组展示。
- 导航卡片展示。
- 导航分组新增、编辑、删除。
- 导航卡片新增、编辑、删除。
- 分组排序、卡片排序：前端通过拖拽或设置页操作更新 sort 后保存。
- 卡片支持：标题、图标/文本、内网地址、公网地址、打开模式（`lan`/`wan`，默认 `wan`）、所属分组、排序。
- 系统设置中的导航管理已经统一为草稿模式：分组和卡片的新增、编辑、删除、排序都先写入草稿，点击系统设置保存后才批量写入后端。
- 导航真实接口闭环已验证：分组和卡片的新增、编辑、删除、排序刷新后保持。
- 导航页整体已由用户确认处于可用状态。
- 导航页单独备份和恢复已接入真实接口并由用户验证可用。
- 右键菜单“新标签页打开 / 新窗口打开”已修复，不再额外打开 `about:blank` 或错误打开前端站点地址。
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
- 右下角网络切换按钮显示“内网优先 / 外网优先”。
- 内外网前端切换使用右下角“内网优先 / 外网优先”按钮，点击卡片时按优先级选择内网或公网地址，并在地址未填写协议时自动补全 `http://`。
- 当前保留优先地址超时回退机制：内网优先时探测内网地址，超时后回退公网；公网优先时探测公网地址，超时后回退内网。
- 自动检测地址和自动检测开关方向已取消，不再恢复 `lanDetectUrl` / `autoDetectLan` 配置入口。

### 收藏夹

- 左侧收藏夹抽屉已重构为现代资源管理器式两栏布局：左侧收藏夹树，右侧当前收藏夹书签列表/搜索结果。
- 抽屉打开时加载收藏夹根节点，子收藏夹按需加载；点击收藏夹标题可全展开/全收起。
- 收藏夹新增、编辑、删除已使用统一弹窗；新增收藏夹可选择上级收藏夹或根目录，前端限制最多 4 级。
- 书签新增、编辑、删除已使用统一弹窗；新增/编辑书签弹窗不再暴露上传图标和图标手填字段，图标由自动抓取/默认逻辑处理。
- 书签字段：标题、URL、favicon、备注、排序、所属收藏夹。
- 收藏搜索：`GET /api/bookmark/search?q=...`，搜索标题、URL、备注。
- 收藏导入：`POST /api/bookmark/import`，解析 Netscape Bookmark HTML 格式。
- 收藏导出：`GET /api/bookmark/export`，导出 HTML。
- 收藏夹和书签右键菜单已移除“上移/下移”，排序方向统一为拖拽排序。
- 书签支持批量操作模式，当前保留批量移动、批量删除入口；按 `Esc` 可退出批量操作。
- 新增/编辑收藏夹、新增/编辑书签弹窗支持 `Esc` 关闭，且打开弹窗时不应自动收回收藏夹抽屉。

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
  - `showTitle`
  - `showClock`
  - `showSeconds`
  - `showSearch`
  - `searchEngines`
  - `backgroundUrl`
  - `backgroundColor`
  - `lanDetectTimeout`
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
- `lanDetectTimeout` 当前用于内外网优先地址探测的超时回退配置；`lanDetectUrl` / `autoDetectLan` 属于已取消的自动检测方向，不应再作为页面配置入口恢复。

### 搜索

- 首页存在网页搜索框。
- 默认搜索引擎配置在前端 `settingsForm.searchEngines` 中，包含 Google、百度、Bing。
- 搜索引擎管理已经统一到系统设置菜单，使用自定义弹窗新增和编辑，不再使用原生 `prompt()`。

### 备份与恢复

- 下载备份：`GET /api/backup/download`
- 恢复备份：`POST /api/backup/restore`
- 导航页备份：`GET /api/navigation/backup`
- 导航页恢复：`POST /api/navigation/restore`
- S3 测试：`POST /api/s3/test`
- “备份到 S3”入口和 `POST /api/backup/s3` 专用接口已按当前方向清理；S3 测试和图片上传同步 S3 能力保留。
- 备份格式为 `.tar.gz`，包含 `data/` 下文件。
- 恢复会将备份包中 `data/` 前缀下的文件写回数据目录。
- 导航页备份格式为 JSON，包含 `version`、`createdAt`、`groups`、`items`，只导出真实后端导航数据，不包含前端演示数据。
- 导航页恢复会替换当前全部导航分组和卡片，前端恢复前有二次确认，恢复后刷新首页导航和设置页导航草稿。

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

- 分组折叠/展开功能已明确不需要，前端 payload、后端模型和新建数据库结构已移除导航分组 `collapsed` 残留。
- 自动判断内外网的 `lanDetectUrl` 和 `autoDetectLan` 方向已经取消，当前保留手动切换“内网优先 / 外网优先”，以及点击卡片时的优先地址探测超时回退机制，该前端页面已验证可用。
- 单卡片独立检测地址未实现；当前 NavItem 只有 `lanUrl`、`wanUrl`、`urlMode`。
- 后端未强制校验导航卡片公网地址必填；目前主要是前端校验。
- 收藏夹部分近期暂不作为主线，主页卡片“从收藏夹设为首页卡片”后续再继续完善。

### 收藏夹相关

- 书签拖拽排序和收藏夹拖拽排序已完成联调和修复：解决拖拽挤占排序抖动问题；展开的收藏夹在拖拽时自动收起。
- 收藏夹菜单需求中的“复制链接”“移动到文件夹”“设为首页卡片 / 取消首页卡片”“批量选择入口”已全部对照验收和修复。
- 书签移动、首页卡片弹窗UI细节已修复（调整了滚动条和下拉框响应及子节点展示）。
- 新增/编辑书签时的URL自动抓取：增加了超时控制（3秒），增加了兜底失败反馈（alert），修复了抓取时不覆盖已有标题的逻辑，兼容了缺省HTTP协议的输入。
- 搜索结果当前 `Path` 仅取直接文件夹名，未构造完整路径。
- 导入 HTML 未做重复检测和跳过逻辑。
- 导入结果只返回 folders/bookmarks 数量，没有跳过/失败明细。
- 导入书签备注 `<DD>` 解析不完整；代码中 `lastBookmarkID` 未实际用于写入备注。
- Safari/Chrome 兼容程度需要真实样例进一步验证。

### 设置和资源

- `storage_configs` 表存在，但实际 S3 配置保存在 `settings` 表中，`storage_configs` 未使用。
- `backup_records` 表存在，但备份下载未明显写入备份记录，当前记录数为 0。
- Logo 配置已明确取消，前端默认设置和后端设置白名单已移除 `logoUrl` / `showLogo`。
- 背景图、背景色能力保留给后续使用，当前 UI 体验待进一步补齐。

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

### 近期后端拆分整理

- 后端 HTTP handler 已完成多轮保守拆分，并已整理重复定义问题，当前 `go fmt ./...`、`go test ./...`、`go build ./...` 均通过。
- 已从 `server.go` 拆出：认证相关 handler、导航页基础 handler、收藏夹基础 handler、设置 handler、健康检查 handler。
- `Routes()` 行为未修改，API 路径、HTTP 方法、请求参数和响应 JSON 保持不变。
- 本轮未拆分备份恢复、S3、静态资源上传、metadata、书签 HTML 导入导出相关逻辑。
- `server.go` 当前职责收敛为：路由注册、通用工具、备份恢复、S3、静态资源上传、metadata、书签导入导出等剩余 HTTP 逻辑。

### 当前架构说明

- 前端当前以 `frontend/src/App.vue` 为顶层容器，配合 `frontend/src/components/` 下的展示型组件和设置页区块组件组织页面。
- `App.vue` 当前仍负责全局状态、登录/初始化视图、首页布局组合、设置弹窗组合、导航草稿保存、收藏夹抽屉状态、上下文菜单事件、上传/备份/恢复等顶层业务编排。
- 前端已完成拆分的模块包括：首页 Hero、右侧浮动按钮、收藏行、移动弹窗、右键菜单、导航拖拽浮层、系统设置菜单、备份恢复区域、搜索引擎管理列表、个性化设置表单。
- 前端仍留在 `App.vue` 的较大职责包括：导航分组和卡片编辑弹窗、分组管理草稿逻辑、收藏夹抽屉整体编排、登录/初始化页面、设置弹窗整体状态、备份恢复动作编排、上传动作编排。
- 后端当前以 `backend/internal/httpx/server.go` 负责 `Routes()` 路由注册和剩余未拆分 HTTP 逻辑，已拆出的 handler 文件与 `server.go` 同属 `httpx` 包，不改变路由调用方式。
- 后端已完成拆分的模块包括：`auth.go` 认证与会话、`navigation.go` 导航页基础接口、`bookmarks.go` 收藏夹基础接口、`settings.go` 设置接口、`health.go` 健康检查。
- 后端仍保留在 `server.go` 中的模块包括：路由注册、通用 JSON/参数工具、静态文件服务、metadata 抓取、书签 HTML 导入导出、全局备份恢复、导航备份恢复、S3 测试、静态资源上传、签名和上传辅助函数。
- 后续如继续拆分，建议顺序为：先拆 metadata 抓取，再拆静态资源上传，再拆书签 HTML 导入导出，再拆导航备份恢复，再拆全局备份恢复，最后拆 S3；每轮只移动一个边界清晰模块并立即执行 `go fmt ./...`、`go test ./...`、`go build ./...`。

### 代码结构问题

- `frontend/src/App.vue` 仍然偏大，当前约 1500 行；已完成两轮保守组件拆分（抽离了右键菜单、拖拽浮层、移动弹窗、设置页菜单、备份恢复、搜索引擎列表、个性化设置等），但编辑弹窗、导航分组网格、收藏夹抽屉等复杂区域仍在 App.vue 中。
- `backend/internal/httpx/server.go` 已完成首批保守拆分，不再集中承载认证、导航、收藏夹基础、设置、健康检查 handler；备份恢复、S3、静态资源上传、metadata、书签 HTML 导入导出等仍保留在 `server.go`，后续如继续拆分应保持小步、只移动、不改行为。
- 后端 store 层只是基础 SQL 封装，没有领域边界和事务封装。
- 前端已经移除了之前多处使用原生 `prompt()` 的地方（例如新增搜索引擎、新增分组等），统一改为了更加美观和体验一致的自定义弹窗组件，但仍需注意未来新功能的表单交互体验。

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

- 自动内网检测设置方向已取消，不应再恢复 `lanDetectUrl` 和 `autoDetectLan` 配置入口；`lanDetectTimeout` 可继续作为优先地址探测超时回退配置。
- 收藏夹无限层级的后端结构存在，前端树形交互不足。
- 书签导入解析使用正则，面对复杂浏览器导出 HTML 可能不够稳健。
- 全局备份恢复会直接覆盖数据目录文件，后端版本兼容检查仍需继续完善；导航页单独恢复已加入前端二次确认和备份版本校验。
- 后端导航卡片接口报错文案已统一为“标题 / 公网地址 / 分组”，前后端文案一致。
- 前端新增/编辑卡片要求公网地址必填，后端已补齐同步强约束，API 直接调用不可绕过。
- 后端删除不存在分组/卡片时返回明确的 404 错误。
- `urlMode` 前后端已统一为只允许 `lan`、`wan`（不再使用 `auto`），默认值统一为 `wan`；旧数据已通过迁移兼容转换。

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

### `nav_items`

用途：首页导航卡片。

字段：

- `id INTEGER PRIMARY KEY`
- `group_id INTEGER NOT NULL`
- `name TEXT NOT NULL`
- `icon TEXT`
- `lan_url TEXT`
- `wan_url TEXT`
- `url_mode TEXT NOT NULL DEFAULT 'wan'`
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

系统设置保存在 SQLite 的 `settings` 表中，通过 `/api/settings` 读写。S3 密钥、站点标题、背景、搜索引擎、内外网优先地址探测超时配置等都在这里。

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
- 新增卡片、新增分组、搜索引擎新增/编辑、收藏夹新增文件夹、新增收藏和收藏转首页卡片都已从原生 `prompt()` 流程改为统一弹窗流程。
- 编辑分组状态下卡片可点击编辑，并支持拖拽排序。
- 手机端卡片布局保持一行 5 个。
- 系统设置内导航管理已全面草稿化，分组和卡片的新增、编辑、删除、排序点击保存后才批量应用。
- 首页卡片保存按钮事件已修复，演示数据模式下分组和卡片的新增、编辑、删除也已补齐本地状态更新。
- 导航地址打开逻辑已增加协议补全，`www.baidu.com` 会按 `http://www.baidu.com` 打开，不会被浏览器当作当前站点相对路径。

最近验证：

```text
frontend npm run build: 通过
go test ./...: 通过（新增导航后端校验测试）
docker compose build biu-panel: 通过
docker compose up -d --force-recreate biu-panel: 已完成
/api/health: 通过
```

近期规则统一与后端加固：

- 系统设置菜单“导航管理”已更名为“分组管理”。
- 分组名称前后端统一限制为最多 10 个字。
- 卡片标题前后端统一限制为最多 15 个字。
- `urlMode` 前后端统一只允许 `lan`、`wan`，默认值统一为 `wan`。
- 后端新增/编辑导航接口已补齐与前端一致的必填校验（标题、公网地址、分组）。
- 后端删除不存在分组/卡片时返回明确的 404 错误。
- 后端导航错误文案已统一为“标题 / 公网地址 / 分组”。
- 旧数据 `urlMode='auto'` 已通过数据库迁移兼容转换为 `wan`；导航备份恢复也已兼容旧 `auto` 的导入转换。
- 新增后端导航校验测试文件 `backend/internal/httpx/server_test.go`。

后端第一轮拆分完成（2026-06-15）：

- 新增 `backend/internal/httpx/settings.go`，从 `server.go` 中拆分出设置相关接口。
- 移动的函数：`getSettings`、`saveSettings`。
- API 行为未变化：`GET /api/settings`、`PUT /api/settings` 路径、方法、参数、响应格式均保持不变。
- 路由注册仍在 `server.go` 的 `Routes()` 中，未改动。
- 已通过验证：`go test ./...`、`go build ./...`、前端设置页保存功能正常。

## 11. 后续开发建议

建议按以下优先级推进。

### P0：先做稳定性和一致性

1. 为后端接口补齐与前端一致的校验：标题、公网地址、分组必填。
2. 修正后端错误文案中的“卡片名称”为“标题”。
3. 补一个最小后端测试集：认证、导航 CRUD、收藏 CRUD、设置保存、metadata URL 校验；导航 CRUD 已完成一次人工真实接口闭环验证，但还未固化为自动化测试。
4. 建立数据库迁移版本机制，避免后续表结构变更困难。
5. 清理过期 sessions 和历史 login_logs 的策略。

### P1：拆分大文件

1. 继续保守拆分 `frontend/src/App.vue`：优先拆 `BackupRestoreSection`、`SearchEngineManagerSection`、`PersonalSettingsForm`、`NavGroupSection`，暂不强拆整个设置面板或编辑弹窗。
2. 将 `backend/internal/httpx/server.go` 拆分为 auth、navigation、bookmarks、settings、assets、backup、s3、metadata 等文件或 handler。
3. 等展示组件稳定后，再考虑将前端业务逻辑抽到 composables 或简单模块中，降低单文件复杂度。

### P2：补齐收藏夹核心

1. 完善多级目录树展开/折叠和按需加载。
2. 实现收藏夹批量删除、批量移动。
3. 实现复制链接、移动到文件夹、设为首页卡片。
4. 搜索结果显示完整路径，而不是仅直接文件夹名。
5. 改进书签导入：支持 `<DD>` 备注、重复检测、跳过统计、失败统计。

### P3：完善内外网优先地址回退

1. 保留当前“内网优先 / 外网优先”手动切换方向，不恢复 `lanDetectUrl` / `autoDetectLan` 自动检测配置入口。
2. 明确并验证 `lanDetectTimeout` 是否只用于优先地址探测超时回退：内网优先时内网超时回退公网，公网优先时公网超时回退内网。
3. 若未来需要单卡片独立检测地址，需要重新确认产品范围并扩展 `nav_items` 表。

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
