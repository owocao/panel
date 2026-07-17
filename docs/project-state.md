# biu-panel 项目状态

> 更新日期：2026-07-17
> 项目路径：`/project/panel`
> 当前分支：`master`
> 当前最新提交：`d531f9d fix: add mobile bookmark drawer close action`
> 当前稳定版本：`v0.3.3`
> 当前工作区：更新前干净

## 1. 当前定位

`biu-panel` 是自托管、单管理员使用的个人导航页和分层收藏夹管理器。它面向手工维护常用服务入口、网页收藏和个人设置，使用本地 SQLite 保存数据，并通过 Docker 部署。

它不是浏览器收藏夹同步工具：不提供浏览器扩展、实时双向同步或自动接管浏览器收藏夹。导航和收藏夹是两套独立数据体系。

当前版本 `v0.3.3` 已发布为 GitHub Latest Release。Docker 交付方式保持可用，Compose 通过 `./data:/app/data` 持久化运行数据；首次空数据部署会进入管理员初始化，不会创建演示数据。

## 2. 已实现功能

### 认证和初始化

- 首次初始化管理员，也支持环境变量首次创建管理员。
- 登录、登出、Cookie 会话和“记住登录”。
- bcrypt 密码哈希；15 分钟内连续 5 次登录失败会锁定。
- 启动时和每 24 小时清理过期会话及旧登录日志。

### 首页导航

- 导航分组和卡片的新增、编辑、删除、排序及右键菜单。
- 卡片支持图片图标和最多 10 个字符的文字图标。
- 卡片支持公网和内网地址；全局模式可选优先内网或优先公网。
- 网络模式只按配置优先级直接选择地址，优先地址为空时才回退；不探测网络可达性。
- 支持导航 JSON 备份和恢复。

### 收藏夹

- 无限层级文件夹、书签 CRUD、跨目录移动和排序。
- 抽屉按需加载文件夹及书签；支持搜索、批量移动、批量删除、右键菜单和拖拽排序。
- 书签支持标题、URL、备注和 favicon；可手动抓取 metadata。
- 点击缺少真实 favicon 的书签时，会后台尝试刷新 favicon，失败不影响打开链接。
- 支持浏览器书签 HTML 导入和导出。
- 抽屉在桌面端和移动端均有明确的关闭入口。

### 设置、搜索和存储

- 设置中心包含个性化、导航分组、收藏夹管理、搜索引擎、S3 和备份恢复。
- 设置、导航管理和设置页收藏夹管理使用草稿后保存的方式。
- 首页支持维护多个外部搜索引擎，不提供本地导航或书签的统一搜索。
- 支持本地图片上传；可选 S3 兼容对象存储用于上传同步和连接测试。
- 支持全局 `.tar.gz` 数据备份/恢复和导航单独备份/恢复。

### 部署和适配

- Go 后端、Vue 前端、SQLite 本地数据库和 Docker Compose 交付。
- 数据目录为 `data/`，默认数据库路径为 `data/db/biu-panel.db`，上传文件位于 `data/uploads/`。
- 提供桌面端完整界面，以及设置页和收藏夹抽屉的移动端基础可用布局。

## 3. 当前代码结构

### 前端

- `frontend/src/App.vue`：顶层状态、初始化/登录视图、页面组件组合、生命周期和少量导航草稿编排。
- `frontend/src/components/`：首页、收藏夹抽屉、设置页、编辑弹窗、移动弹窗、右键菜单和浮动操作入口。
- `frontend/src/components/settings/`：设置面板和收藏夹管理展示。
- `frontend/src/composables/`：业务状态和动作。
  - `useNavigation.js`：导航数据、地址选择、搜索引擎和打开链接。
  - `useBookmarks.js`：文件夹、书签、搜索、选择态和按需加载。
  - `useBookmarkActions.js`：书签/文件夹编辑、移动、删除、批量操作和 favicon 刷新触发。
  - `useDragSort.js`：导航、文件夹和书签的拖拽排序。
  - `useEditSave.js`：编辑弹窗保存、上传和 metadata 处理。
  - `useSettings.js`、`useFolderDrafts.js`：设置与收藏夹管理草稿。
  - `useBackupRestore.js`：全局/导航备份恢复及 HTML 导入导出。
- `frontend/src/style.css`：全局玻璃主题、首页、设置、弹窗和收藏夹样式。
- `frontend/src/lib/api.js`：前端 API 封装。

### 后端

- `backend/cmd/server/main.go`：加载配置、打开数据库、可选初始化管理员、启动清理任务和 HTTP 服务。
- `backend/internal/httpx/`：按认证、导航、收藏夹、导入导出、metadata、资源上传、S3、设置和备份恢复拆分的 HTTP 领域文件。
- `backend/internal/store/`：已按模型、认证、导航、收藏夹和设置拆分；`store.go` 只保留 SQLite 打开、WAL 配置和建表迁移。
- SQLite 使用 `modernc.org/sqlite`，打开时启用 foreign keys、busy timeout、WAL 和 `synchronous=NORMAL`。

### 数据与运行

- Docker 入口：`Dockerfile`、`docker-compose.yml`。
- Compose 数据挂载：`./data:/app/data`。
- 常用本地启动、构建和验证命令见 `docs/development.md`。
- 前端构建：

```bash
cd /project/panel/frontend
PATH=/project/panel/.tools/node/bin:$PATH npm run build
```

- 后端验证：

```bash
cd /project/panel/backend
PATH=/project/panel/.tools/go/bin:$PATH GOCACHE=/tmp/panel-go-build go test ./...
PATH=/project/panel/.tools/go/bin:$PATH GOCACHE=/tmp/panel-go-build go vet ./...
```

`v0.3.3` 合并前已通过前端构建、`go test ./...` 和 `go vet ./...`。

## 4. 最近版本变化

### v0.3.1

- 清理确认无用的代码和 CSS。
- 归拢设置、弹窗和收藏夹抽屉样式。
- 后端 store 按领域拆分。
- 提取文件夹展示、首页选择保护和首页时间显示辅助逻辑。
- 补充 Docker 空数据部署说明。

### v0.3.2

- 修复优先内网/优先公网未按配置直接打开的问题。
- 修复设置页收藏夹父级选择器的字体、层级和裁剪问题。
- 首页文字图标上限由 5 个字符改为 10 个字符。
- 首页卡片图标填满图标框。

### v0.3.3

- 收藏夹抽屉右上角增加“关闭”按钮。
- 关闭按钮桌面端和移动端均可见，并复用系统设置的按钮风格。
- 保留遮罩、键盘、搜索、批量操作、右键菜单和拖拽的既有行为。

## 5. 已知限制和维护注意事项

- `App.vue` 仍有约 969 行，`style.css` 约 898 行；两者已整理过，但继续为减少行数而强拆的风险较高。
- `useDragSort.js` 负责多种拖拽状态和排序保存，属于高风险区域；没有明确问题时不要重构。
- `useEditSave.js` 保存分支较多，涉及导航、书签、上传和 metadata；新增编辑类型前应先评估边界。
- 导航与收藏夹没有统一模型或统一工作流。
- 没有浏览器同步、浏览器扩展、自动采集、跨设备同步、离线支持或多用户权限模型。
- 网络模式不做 ping、fetch 或异步可达性探测；它只按用户当前选择的优先级打开 URL。
- 全局备份恢复会写回数据目录，缺少版本化迁移、manifest 校验和细粒度恢复策略；操作前应先备份当前 `data/`。
- S3 仅用于图片上传同步和测试，不用于备份；S3 密钥按当前 V1 设计存于本地 SQLite，未额外加密。
- 收藏夹导入的实际去重行为以目标目录和 URL 为准；导入前应保留原始 HTML 备份。
- 没有前端自动化测试；现有后端测试主要覆盖认证读取保护、导航校验和导航恢复校验。
- 页面仍有原生 `confirm()` / `alert()`，高风险操作的体验未完全统一。

## 6. 暂停期间的接手入口

未来重新启动项目时，按以下顺序阅读和确认：

1. `README.md`：当前对外说明和本地/Docker 入口。
2. `docs/project-state.md`：当前代码、版本和已知限制。
3. `docs/pause-notes.md`：暂停原因、未解决的产品问题和重启原则。
4. `AGENTS.md`：长期工程约束和产品边界。
5. `docs/deployment.md`：Docker 数据目录、部署、备份和回滚操作。
6. GitHub 的 `v0.3.1`、`v0.3.2`、`v0.3.3` Release notes 与 `git log`：确认后续是否已有必要修复。

恢复开发前，不要直接重构或增加功能。先实际部署 `v0.3.3` 使用一段时间，并重新明确单一核心痛点和最小目标。
