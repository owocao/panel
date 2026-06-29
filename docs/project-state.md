# biu-panel 项目状态

> 整理时间：2026-06-26
> 项目路径：`/project/panel`
> 当前分支：`master`
> 当前最新提交：`1df48e9 refactor: split httpx bookmark transfer and navigation backup`
> 整理前工作区：干净

## 1. 项目概览

`biu-panel` 是一个单管理员个人使用的轻量导航面板与书签管理器。项目参考 Sun-Panel 的易用性，但不是 Sun-Panel 的二次开发。

当前状态：

- 已具备 Go 后端、Vue 前端、SQLite 数据库和 Docker 交付结构。
- 首页导航、系统设置、搜索引擎管理、收藏夹、书签导入导出、备份恢复等核心模块已有实现。
- 项目不再是静态原型，当前进入 V1 整理、稳定性修复和维护性提升阶段。
- 当前主要问题是部分后端接口校验不足、部分关键流程缺少事务和测试；前端已完成一轮模块化拆分，后续需继续保持模块边界。

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
      assets.go
      auth.go
      bookmark_import.go
      bookmark_transfer.go
      bookmarks.go
      health.go
      metadata.go
      navigation.go
      navigation_backup.go
      s3.go
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

- `backend/internal/httpx/server.go`：`Server` 结构、路由注册、静态文件服务、初始化管理员、定时清理任务、全局备份下载/恢复、通用 HTTP 工具函数。当前不再承载 metadata、上传、S3、书签导入导出和导航备份恢复逻辑。
- `backend/internal/httpx/auth.go`：初始化、登录、登出、当前用户、会话校验。
- `backend/internal/httpx/navigation.go`：导航分组和卡片基础 API。
- `backend/internal/httpx/navigation_backup.go`：导航页 JSON 备份下载、导航页 JSON 恢复、导航备份文件结构和恢复校验。
- `backend/internal/httpx/bookmarks.go`：收藏夹文件夹、书签、搜索基础 API。
- `backend/internal/httpx/bookmark_import.go`：书签 HTML 导入解析逻辑。
- `backend/internal/httpx/bookmark_transfer.go`：浏览器书签 HTML 导出 HTTP 入口、书签 HTML 导入 HTTP 入口；解析逻辑仍在 `bookmark_import.go`。
- `backend/internal/httpx/metadata.go`：网页 metadata 抓取，包括标题提取、favicon 解析和 HTML 实体还原。
- `backend/internal/httpx/assets.go`：图片上传 HTTP 入口、本地上传文件保存、上传资源记录、上传后可选 S3 同步触发。
- `backend/internal/httpx/s3.go`：S3 连接测试、Public URL 生成、S3 PUT Object、AWS V4 签名相关辅助函数。
- `backend/internal/httpx/settings.go`：系统设置读写。
- `backend/internal/httpx/health.go`：健康检查。
- `backend/internal/httpx/server_test.go`：现有后端测试。

存储层：

- `backend/internal/store/store.go`：SQLite 打开、建表迁移、基础 CRUD、搜索、导航替换等。

后续后端问题排查优先读取：

- 路由是否注册、静态文件、全局备份恢复、通用响应/解析工具：先读 `server.go`。
- 初始化、登录、登出、会话、登录失败锁定：先读 `auth.go`，再读 `store.go` 的用户、session、登录日志相关方法。
- 首页导航分组和卡片 CRUD：先读 `navigation.go`，再读 `store.go` 的导航相关方法。
- 导航页备份下载和恢复：先读 `navigation_backup.go`，再读 `store.go` 的 `ListNavigation`、`ReplaceNavigation`。
- 收藏夹文件夹、书签 CRUD 和搜索：先读 `bookmarks.go`，再读 `store.go` 的收藏夹和书签相关方法。
- 浏览器书签导入解析异常：先读 `bookmark_import.go`；导入/导出 HTTP 入口异常：先读 `bookmark_transfer.go`。
- metadata 标题/favicon 抓取：先读 `metadata.go`。
- 图片上传、本地上传路径、上传资源记录：先读 `assets.go`；S3 同步和连接测试：先读 `s3.go`。
- 系统设置读写和 S3 配置保存：先读 `settings.go`，再读 `store.go` 的 settings 方法。

维护约束：

- `server.go` 当前仍保留路由、静态文件、全局备份恢复和通用工具，暂时不要继续为了行数拆分路由注册或通用工具。
- 全局备份恢复仍在 `server.go`，涉及文件覆盖和路径安全；在备份 manifest、版本校验和恢复策略明确前，暂时不要仅为拆分而移动或重写。
- `auth.go`、`navigation.go`、`bookmarks.go`、`settings.go` 当前边界清晰，除非处理对应领域问题，否则不要继续拆。
- `metadata.go`、`assets.go`、`s3.go`、`bookmark_transfer.go`、`navigation_backup.go` 是同包保守拆分结果，后续维护应保持 API 行为不变。
- `backend/internal/store/store.go` 当前仍未拆分，继续集中承载连接、迁移、模型和各领域 SQL。后续应保守处理：优先补必要查询/校验方法和测试，不要直接引入 repository/interface 或大规模重构；如需拆文件，也应保持同包、只移动代码、不改变调用方和事务行为。

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

当前前端模块边界：

- `frontend/src/App.vue`：前端顶层入口，负责登录/初始化视图切换、页面级状态、顶层组件组合、生命周期、全局快捷键、模块依赖注入和少量尚未拆分的导航草稿保存逻辑。
- `frontend/src/components/`：负责页面和弹窗展示层，包括首页、收藏夹抽屉、编辑弹窗、右键菜单、移动弹窗、浮动操作按钮、设置页分区等。
- `frontend/src/components/settings/`：负责设置页整体面板和收藏夹管理展示。
- `frontend/src/composables/`：负责业务状态、业务动作和跨组件交互编排。
- `frontend/src/utils/`：负责不依赖 Vue 状态、不调用 API 的纯工具函数。

主要组件：

- `HomeHero.vue`：首页导航展示。
- `FloatingActions.vue`：首页浮动操作入口。
- `BookmarkDrawer.vue`、`BookmarkFolderTreeNode.vue`、`BookmarkRow.vue`：收藏夹抽屉、文件夹树和书签行展示。
- `SettingsPanel.vue`、`BookmarkManager.vue`、`PersonalSettingsForm.vue`、`SearchEngineManagerSection.vue`、`BackupRestoreSection.vue`、`SettingsMenu.vue`：设置页布局和分区展示。
- `EditDialog.vue`、`MoveDialog.vue`、`ContextMenu.vue`、`NavDragFloat.vue`：编辑、移动、右键菜单和拖拽浮层展示。

Composable 职责：

- `useNavigation.js`：首页导航数据、网络模式、搜索引擎、搜索执行和导航链接打开逻辑。
- `useBookmarks.js`：收藏夹基础状态、当前文件夹、书签列表、搜索、防抖和选择态。
- `useBookmarkActions.js`：书签/收藏夹移动、删除、批量操作、编辑入口和打开书签。
- `useFolderDrafts.js`：设置页收藏夹草稿、收藏夹管理树、草稿同步、草稿保存、收藏夹移动和删除。
- `useDragSort.js`：首页导航、收藏夹和书签的拖拽排序状态、悬停排序、落点保存和失败刷新。
- `useEditDialog.js`：编辑弹窗 UI 状态、关闭、字段限制、分组选择和图标模式切换。
- `useEditSave.js`：编辑弹窗保存、删除导航卡片、上传图标、metadata 抓取，以及新增分组、卡片、收藏夹、书签。
- `useSettings.js`：设置页打开关闭、菜单切换、设置加载、S3 测试和设置保存。
- `useContextMenu.js`：右键菜单状态、位置、菜单项生成和菜单动作触发。
- `useBackupRestore.js`：整站备份恢复、导航备份导入导出、收藏夹导入导出。

Utils 职责：

- `bookmarkTree.js`：收藏夹树标准化、扁平化、查找和克隆。
- `display.js`：图标值判断、卡片文本截断、图标 URL、时间日期和网络模式展示文案。
- `navigation.js`：网络模式标准化、URL 补协议、内外网候选 URL 和最终打开 URL 解析。

后续排查优先读取：

- 首页导航、内外网打开、搜索引擎：先读 `useNavigation.js` 和 `utils/navigation.js`，再读 `HomeHero.vue`。
- 收藏夹加载、文件夹选择、书签搜索和选择态：先读 `useBookmarks.js`，再读 `BookmarkDrawer.vue`。
- 收藏夹/书签移动、删除和批量操作：先读 `useBookmarkActions.js`。
- 设置页收藏夹管理、草稿移动删除和保存：先读 `useFolderDrafts.js`，再读 `useSettings.js` 和 `BookmarkManager.vue`。
- 编辑弹窗字段、保存、上传图标和 metadata 抓取：先读 `EditDialog.vue`、`useEditDialog.js` 和 `useEditSave.js`。
- 拖拽排序：先读 `useDragSort.js`，再读相关展示组件。
- 右键菜单：先读 `useContextMenu.js` 和 `ContextMenu.vue`。
- 备份恢复、导入导出：先读 `useBackupRestore.js` 和 `BackupRestoreSection.vue`。
- 纯展示异常：先读 `utils/display.js`。

维护约束：

- `App.vue` 已降到 1000 行以下，但仍是前端顶层入口。后续不要继续堆业务流程，新逻辑优先放入对应 composable 或组件。
- `useDragSort.js` 已承载复杂拖拽状态和保存逻辑，只应继续维护拖拽相关逻辑，不要混入非拖拽业务。
- `useEditSave.js` 保存分支较多，新增编辑类型时应评估是否继续拆分为更小的保存模块。
- `useBookmarkActions.js` 已包含移动、删除和批量操作，后续批量功能应避免继续无边界扩张。
- `SettingsPanel.vue` 和 `EditDialog.vue` 仍需保持展示层职责；如果设置页或编辑表单继续扩展，应优先拆子组件。

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

- 前端部分 composable 已接近变大风险，后续新增逻辑需要按模块边界拆分。
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
