# biu-panel 项目状态

> 整理时间：2026-06-26
> 项目路径：`/project/panel`
> 当前分支：`master`
> 当前最新提交：`f425aaa chore: refine docker deployment setup`
> 整理前工作区：干净

## 1. 项目概览

`biu-panel` 是一个单管理员个人使用的轻量导航面板与书签管理器。项目参考 Sun-Panel 的易用性，但不是 Sun-Panel 的二次开发。

当前状态：

- 已具备 Go 后端、Vue 前端、SQLite 数据库和 Docker 交付结构。
- 首页导航、系统设置、搜索引擎管理、收藏夹、书签导入导出、备份恢复等核心模块已有实现。
- 项目不再是静态原型，当前进入 V1 整理、稳定性修复和维护性提升阶段。
- 当前主要问题是文档曾经滞后、前端 `App.vue` 偏大、部分后端接口校验不足、部分关键流程缺少事务和测试。

## 2. 技术栈

后端：

- Go
- 标准库 `net/http`
- SQLite 驱动：`modernc.org/sqlite`
- 密码哈希：`golang.org/x/crypto/bcrypt`

前端：

- Vue 3
- Vite
- npm

数据和部署：

- SQLite 本地数据库
- 本地上传文件存储
- 可选 S3 兼容对象存储，用于图片上传同步
- Dockerfile + docker-compose.yml

## 3. 当前目录结构

```text
/project/panel/
  AGENTS.md
  Dockerfile
  README.md
  docker-compose.yml
  docs/
    current-direction.md
    project-state.md
    requirements.md
    tasks.md
    development.md
    release-checklist.md
    deployment.md
    release-notes-v0.1-stable.md
  backend/
    go.mod
    go.sum
    cmd/server/main.go
    internal/config/config.go
    internal/httpx/
      auth.go
      bookmarks.go
      bookmark_import.go
      health.go
      navigation.go
      server.go
      server_test.go
      settings.go
    internal/store/store.go
  frontend/
    package.json
    package-lock.json
    vite.config.js
    index.html
    public/
    src/
      App.vue
      main.js
      style.css
      lib/api.js
      components/
  scripts/
    smoke.sh
  data/
```

`data/`、`frontend/node_modules/`、`frontend/dist/` 是本地运行/依赖/构建产物，不应提交。

## 4. 后端现状

入口：

- `backend/cmd/server/main.go`

配置：

- `backend/internal/config/config.go`

HTTP 层：

- `backend/internal/httpx/server.go`：路由注册、静态文件、上传、S3、metadata、备份恢复、书签导入导出、通用工具。
- `backend/internal/httpx/auth.go`：初始化、登录、登出、当前用户、会话校验。
- `backend/internal/httpx/navigation.go`：导航分组和卡片基础 API。
- `backend/internal/httpx/bookmarks.go`：收藏夹文件夹、书签、搜索基础 API。
- `backend/internal/httpx/bookmark_import.go`：书签 HTML 导入解析逻辑。
- `backend/internal/httpx/settings.go`：系统设置读写。
- `backend/internal/httpx/health.go`：健康检查。
- `backend/internal/httpx/server_test.go`：现有后端测试。

存储层：

- `backend/internal/store/store.go`：SQLite 打开、建表迁移、基础 CRUD、搜索、导航替换等。

当前 API 大类：

- 健康检查：`GET /api/health`
- 初始化：`GET /api/setup/status`、`POST /api/setup`
- 认证：`POST /api/auth/login`、`POST /api/auth/logout`、`GET /api/auth/me`
- 导航：`GET /api/navigation`、`POST/PUT/DELETE /api/navigation/groups`、`POST/PUT/DELETE /api/navigation/items`
- 收藏夹：`GET/POST/PUT/DELETE /api/bookmark/folders`、`GET/POST/PUT/DELETE /api/bookmarks`、`GET /api/bookmark/search`
- 元数据：`GET /api/metadata`
- 设置：`GET /api/settings`、`PUT /api/settings`
- 上传：`POST /api/assets/upload`
- 备份：`GET /api/backup/download`、`POST /api/backup/restore`
- 导航备份：`GET /api/navigation/backup`、`POST /api/navigation/restore`
- 书签导入导出：`GET /api/bookmark/export`、`POST /api/bookmark/import`
- S3：`POST /api/s3/test`

## 5. 前端现状

入口：

- `frontend/src/main.js`
- `frontend/src/App.vue`

API 封装：

- `frontend/src/lib/api.js`

全局样式：

- `frontend/src/style.css`

已拆分组件：

- `BackupRestoreSection.vue`
- `BookmarkFolderTreeNode.vue`
- `BookmarkRow.vue`
- `ContextMenu.vue`
- `FloatingActions.vue`
- `HomeHero.vue`
- `MoveDialog.vue`
- `NavDragFloat.vue`
- `PersonalSettingsForm.vue`
- `SearchEngineManagerSection.vue`
- `SettingsMenu.vue`

`App.vue` 当前仍负责：

- 登录/初始化视图切换。
- 首页布局组合。
- 导航分组和卡片状态。
- 系统设置弹窗状态和草稿机制。
- 导航草稿批量保存。
- 收藏夹管理文件夹草稿保存。
- 收藏夹抽屉状态、树加载、书签列表和搜索。
- 右键菜单、编辑弹窗、拖拽排序。
- 上传、metadata 抓取、备份恢复动作编排。

## 6. 数据库现状

数据库使用 SQLite。建表逻辑在 `backend/internal/store/store.go`。

当前表：

- `users`：管理员账号。
- `sessions`：登录会话。
- `login_logs`：登录日志和失败锁定依据。
- `settings`：系统设置 Key-Value。
- `nav_groups`：首页导航分组。
- `nav_items`：首页导航卡片。
- `bookmark_folders`：收藏夹文件夹，支持层级。
- `bookmarks`：书签条目。
- `assets`：上传资源记录。
- `storage_configs`：外部存储配置预留表，当前实际 S3 配置保存在 `settings`。
- `backup_records`：备份记录预留表，当前备份流程未作为主路径使用。

迁移方式：

- 以 `CREATE TABLE IF NOT EXISTS` 为主。
- 当前有一条兼容迁移：旧 `nav_items.url_mode=''` 或 `auto` 会转为 `wan`。
- 尚未建立版本化 migration 体系。

## 7. 已实现功能

### 认证与安全

- 首次初始化管理员。
- 支持环境变量首次创建管理员。
- 登录、登出、当前用户。
- Cookie 会话。
- 默认会话随浏览器关闭失效。
- “记住登录”有效期 30 天。
- 登录成功/失败日志。
- 15 分钟内连续 5 次失败锁定。
- 启动时和每 24 小时清理过期 session 和旧登录日志。
- 主要读取类 API 已要求登录。

### 首页导航

- 分组和卡片展示。
- 真实导航数据为空时显示空状态，不再回填内置测试数据。
- 分组新增、编辑、删除、排序。
- 卡片新增、编辑、删除、排序。
- 卡片字段：标题、图标、内网地址、公网地址、打开模式、所属分组、排序。
- 分组名称限制 10 个字。
- 卡片标题限制 15 个字。
- 公网地址必填。
- `urlMode` 只允许 `lan`、`wan`。
- 未填写协议的 URL 打开前补全 `http://`。
- 右键菜单支持打开、编辑、删除等操作。
- 系统设置中的分组管理使用草稿机制。
- 导航页单独备份和恢复。

### 内外网切换

- 支持手动“内网优先 / 公网优先”切换。
- 点击卡片时按优先级探测地址，失败或超时后回退。
- `lanDetectTimeout` 用于优先地址探测超时。
- 自动检测入口已取消。

### 收藏夹

- 左侧抽屉打开后加载。
- 文件夹树支持按需加载子节点。
- 文件夹新增、编辑、删除。
- 书签新增、编辑、删除。
- 书签搜索。
- 文件夹和书签拖拽排序。
- 书签跨文件夹移动。
- 收藏夹可在设置页或右键菜单中移动到根目录或其他收藏夹下。
- 设置页收藏夹管理支持上移、下移。
- 批量选择、批量移动、批量删除。
- 复制链接。
- 收藏夹条目设为首页卡片。
- 浏览器书签 HTML 导入。
- 浏览器书签 HTML 导出。

### 设置中心

- 个性化设置。
- 搜索引擎管理。
- S3 设置和连接测试。
- 分组管理。
- 备份恢复。
- 设置保存使用草稿机制。
- 保存成功后关闭设置页。

### 上传与 S3

- 本地图片上传。
- 上传文件限制为图片。
- 本地文件存放在 `data/uploads`。
- 若启用 S3，上传后尝试同步到 S3。
- S3 secret 明文保存在 SQLite，符合 V1 设计。

### 备份恢复

- 全局 `.tar.gz` 备份下载。
- 全局备份恢复。
- 导航页 JSON 备份下载。
- 导航页 JSON 恢复。
- 收藏夹 HTML 导入导出。

## 8. 当前风险和技术债

高优先级：

- 系统设置保存和导航草稿保存不是后端单事务，可能部分成功。
- 收藏夹接口的存在性校验、字段校验、错误码一致性弱于导航接口。
- 全局备份恢复缺少版本信息校验和细粒度恢复策略。
- 前端仍有原生 `confirm()` 和 `alert()`，需要逐步替换为统一弹窗。

中优先级：

- `frontend/src/App.vue` 偏大，后续维护成本高。
- `backend/internal/httpx/server.go` 仍包含多个模块的逻辑。
- Store 层缺少领域服务和事务封装。
- 数据库迁移缺少版本化机制。
- 书签导入导出需要真实 Chrome/Safari 样例持续验证。

低优先级：

- `storage_configs`、`backup_records` 暂未作为主路径使用。
- README、部署文档、发布清单也可能需要后续继续同步。

## 9. Docker 和运行方式

Docker 交付文件：

- `Dockerfile`
- `docker-compose.yml`

当前 compose 主要配置：

- 服务名：`biu-panel`
- 容器端口：`55088`
- 数据目录：`./data:/app/data`
- 环境变量：
  - `BIU_PANEL_PORT=55088`
  - `BIU_PANEL_DATA_DIR=/app/data`
  - `BIU_PANEL_STATIC_DIR=/app/public`

AGENTS.md 曾规划最终拆分挂载：

- `./data/db:/app/data/db`
- `./data/uploads:/app/data/uploads`
- `./data/backups:/app/data/backups`

但当前实际 compose 仍是 `./data:/app/data` 单挂载。未得到用户确认前不要擅自修改部署方式。

## 10. 验证状态

按 `docs/development.md`，应执行：

- 后端：使用项目内 `.tools/go/bin/go` 执行 `go test ./...`
- 前端：使用项目内 `.tools/node/bin/npm` 执行 `npm run build`

本次 2026-06-26 文档整理后已确认：

- Go：`go1.25.4 linux/amd64`
- Node.js：`v22.21.1`
- npm：`10.9.4`
- 后端 `go test ./...` 通过。
- 后端 `go build ./...` 通过。
- 前端 `npm run build` 通过。

注意：Go 依赖下载需要网络访问。首次在沙箱中运行测试时，需要允许访问 Go module proxy；依赖会写入 `/project/panel/.cache/go-mod`。

## 11. 后续维护原则

- 以 `docs/current-direction.md` 为产品方向基线。
- 以本文档为代码事实基线。
- 大功能完成后同步更新本文档。
- 不要把临时任务、阶段状态写入 `AGENTS.md`。
- 对数据库、部署、备份恢复、导入导出等高风险修改，必须先说明影响范围和回滚方式。
