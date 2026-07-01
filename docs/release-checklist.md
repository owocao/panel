# biu-panel 发布检查清单

> 整理时间：2026-06-26
> 用途：发布前检查、回归验证和风险确认
> 注意：本清单不表示当前环境已经执行通过；每次发布前应重新勾选并记录结果。

## 1. 构建检查

- [ ] 后端测试：`cd /project/panel/backend && go test ./...`
- [ ] 后端构建：`cd /project/panel/backend && go build ./...`
- [ ] 前端构建：`cd /project/panel/frontend && npm run build`
- [ ] Docker 构建：`cd /project/panel && docker compose build biu-panel`
- [ ] Docker 启动：`cd /project/panel && docker compose up -d --force-recreate biu-panel`
- [ ] 健康检查：`curl -fsS http://127.0.0.1:55088/api/health`

## 2. 基础冒烟

- [ ] 容器可以使用 `./data:/app/data` 挂载启动。
- [ ] `/api/health` 返回 `status=ok`。
- [ ] 后端可以正确服务前端静态文件。
- [ ] 未初始化时可以进入初始化流程。
- [ ] 使用环境变量可首次创建管理员。
- [ ] 登录成功后可以进入首页。
- [ ] 登出后私有 API 返回 401。

## 3. 导航页回归

- [ ] 首页首次加载不请求完整收藏夹数据。
- [ ] 可以创建、编辑、删除导航分组。
- [ ] 可以创建、编辑、删除导航卡片。
- [ ] 分组名称最多 10 个字。
- [ ] 卡片标题最多 15 个字。
- [ ] 卡片公网地址必填。
- [ ] `urlMode` 只允许 `lan` / `wan`。
- [ ] 卡片默认当前标签页打开。
- [ ] 未填写协议的 URL 打开前会补全协议。
- [ ] 分组和卡片排序刷新后保持。
- [ ] 系统设置内分组管理未保存前不影响首页。
- [ ] 设置保存成功后关闭设置页。

## 4. 收藏夹回归

- [ ] 打开收藏夹抽屉后才加载根目录。
- [ ] 展开文件夹后才加载子节点。
- [ ] 可以创建、编辑、删除文件夹。
- [ ] 可以创建、编辑、删除书签。
- [ ] 删除文件夹和书签有二次确认。
- [ ] 文件夹和书签拖拽排序后保存。
- [ ] 书签可以跨文件夹移动。
- [ ] 批量选择、批量移动、批量删除可用。
- [ ] 收藏夹搜索只返回书签结果，并显示路径。
- [ ] 书签设为首页卡片后与原书签互不影响。

## 5. 导入导出和备份

- [ ] Chrome 书签 HTML 可导入并保留结构。
- [ ] Safari 书签 HTML 可导入并保留结构。
- [ ] 后续导入重复判断符合 URL、标题、备注、所属目录规则。
- [ ] 收藏夹导出 HTML 可被浏览器识别。
- [ ] 导航页 JSON 备份可下载。
- [ ] 导航页 JSON 恢复前有二次确认。
- [ ] 导航页恢复后刷新首页和设置页草稿。
- [ ] 全局 `.tar.gz` 备份可下载。
- [ ] 全局恢复前有二次确认。
- [ ] 全局恢复不会写出 `data/` 目录之外。

## 6. 设置与资源

- [ ] 搜索引擎管理在系统设置独立菜单中。
- [ ] 搜索引擎修改未保存前不影响首页。
- [ ] 个性化设置保存后刷新仍生效。
- [ ] Logo 设置未暴露，`showLogo` / `logoUrl` 不被保存。
- [ ] `lanDetectUrl` / `autoDetectLan` 不暴露。
- [ ] `lanDetectTimeout` 可保存。
- [ ] 本地图片上传可用。
- [ ] 非图片上传被拒绝。
- [ ] S3 设置可保存。
- [ ] S3 测试接口可用。
- [ ] 备份恢复页不出现“备份到 S3”入口。

## 7. 安全和数据

- [ ] 登录失败 5 次后 15 分钟锁定。
- [ ] 登录成功和失败写入日志。
- [ ] 默认会话随浏览器关闭失效。
- [ ] “记住登录”会话可长期保持。
- [ ] 过期 session 会被清理。
- [ ] 不提交 `.env`、token、key、数据库、上传文件、备份文件。
- [ ] 备份前确认 SQLite 和上传文件目录可读。
- [ ] 恢复前确认已有数据已经另行备份。

## 8. 文档发布

- [ ] `docs/current-direction.md` 与当前方向一致。
- [ ] `docs/project-state.md` 与代码状态一致。
- [ ] `docs/tasks.md` 已更新剩余任务。
- [ ] `docs/development.md` 的命令仍可用。
- [ ] `README.md` 与实际部署方式一致。
- [ ] `docs/deployment.md` 与当前 Docker 配置一致。
- [ ] 记录本次实际执行的测试、构建、发布命令和结果。

## 9. 最近验证记录

2026-06-26 文档整理后，已找到项目内隔离工具链：

- Go：`/project/panel/.tools/go/bin/go`
- Node：`/project/panel/.tools/node/bin/node`
- npm：`/project/panel/.tools/node/bin/npm`

已执行并通过：

- `cd /project/panel/backend && PATH=/project/panel/.tools/go/bin:/project/panel/.tools/node/bin:$PATH GOCACHE=/project/panel/.cache/go-build GOMODCACHE=/project/panel/.cache/go-mod go test ./...`
- `cd /project/panel/backend && PATH=/project/panel/.tools/go/bin:/project/panel/.tools/node/bin:$PATH GOCACHE=/project/panel/.cache/go-build GOMODCACHE=/project/panel/.cache/go-mod go build ./...`
- `cd /project/panel/frontend && PATH=/project/panel/.tools/go/bin:/project/panel/.tools/node/bin:$PATH npm run build`

未在本次补跑：

- Docker 构建和容器冒烟

## 10. 人工回归测试清单

> 适用范围：当前 V1 开发阶段，前端模块化与 httpx 拆分后的人工回归验证
> 说明：本清单用于人工验证，不包含自动测试命令。

### 10.1 初始化与登录

| 功能 | 操作步骤 | 预期结果 | 异常优先排查文件 |
|---|---|---|---|
| 首次初始化 | 清空测试数据后首次打开页面，填写管理员账号密码并提交 | 初始化成功，进入登录或首页；刷新后不再显示初始化页 | `backend/internal/httpx/auth.go`、`backend/internal/store/store.go`、`frontend/src/App.vue` |
| 登录 | 输入正确账号密码登录 | 成功进入首页，顶部/页面状态为已登录 | `backend/internal/httpx/auth.go`、`frontend/src/App.vue` |
| 登录失败锁定 | 连续输入错误密码 5 次 | 返回中文错误提示，并锁定约 15 分钟 | `backend/internal/httpx/auth.go`、`backend/internal/store/store.go` |
| 记住登录 | 勾选“记住登录”后登录，关闭并重新打开浏览器 | 会话保持登录 | `backend/internal/httpx/auth.go`、`frontend/src/App.vue` |
| 退出登录 | 点击退出登录 | 会话清除，回到登录页 | `backend/internal/httpx/auth.go`、`frontend/src/App.vue` |

### 10.2 首页导航

| 功能 | 操作步骤 | 预期结果 | 异常优先排查文件 |
|---|---|---|---|
| 首页空状态 | 删除全部导航分组并保存后刷新 | 显示空状态，不出现测试分组/测试卡片 | `frontend/src/composables/useNavigation.js`、`frontend/src/components/HomeHero.vue`、`backend/internal/httpx/navigation.go` |
| 分组新增/编辑/删除 | 打开设置，新增、重命名、删除分组并保存 | 保存后首页与刷新后数据一致 | `frontend/src/composables/useSettings.js`、`frontend/src/composables/useNavigation.js`、`backend/internal/httpx/navigation.go` |
| 卡片新增/编辑/删除 | 新增卡片，填写标题、公网地址、图标等；编辑后保存；再删除 | 卡片展示、编辑、删除均生效，刷新后保持 | `frontend/src/composables/useEditDialog.js`、`frontend/src/composables/useEditSave.js`、`frontend/src/components/EditDialog.vue`、`backend/internal/httpx/navigation.go` |
| 卡片打开 | 点击卡片 | 默认当前标签页打开目标地址 | `frontend/src/composables/useNavigation.js`、`frontend/src/utils/navigation.js` |
| 内外网优先切换 | 切换内网优先/公网优先后点击卡片 | 按优先级探测，失败后回退另一个地址 | `frontend/src/composables/useNavigation.js`、`frontend/src/utils/navigation.js` |
| 拖拽排序 | 拖动分组或卡片调整顺序并保存 | 顺序更新，刷新后保持 | `frontend/src/composables/useDragSort.js`、`backend/internal/httpx/navigation.go` |
| 首页禁止拖选 | 在首页空白处按住鼠标拖动 | 不出现页面内容选中阴影，卡片和标题不被拖选 | `frontend/src/App.vue`、`frontend/src/style.css` |
| 首页 Ctrl+A | 在首页非输入区域按 Ctrl+A / Cmd+A | 不全选首页内容；搜索框内 Ctrl+A 仍可选中输入内容 | `frontend/src/App.vue`、`frontend/src/components/HomeHero.vue` |
| 首页搜索框 focus | 点击搜索框并输入内容 | 搜索框保持单层视觉，不出现内层圆角输入框 | `frontend/src/components/HomeHero.vue`、`frontend/src/style.css` |

### 10.3 系统设置

| 功能 | 操作步骤 | 预期结果 | 异常优先排查文件 |
|---|---|---|---|
| 设置入口 | 从首页打开系统设置 | 默认进入“个性化”，菜单和内容正常显示 | `frontend/src/composables/useSettings.js`、`frontend/src/components/settings/SettingsPanel.vue` |
| 草稿机制 | 修改标题/背景/搜索引擎但不保存，关闭设置页 | 首页不受未保存修改影响 | `frontend/src/composables/useSettings.js`、`frontend/src/components/settings/PersonalSettingsForm.vue` |
| 保存设置 | 修改个性化设置后点击保存 | 保存成功并关闭设置页，刷新后保持 | `frontend/src/composables/useSettings.js`、`backend/internal/httpx/settings.go` |
| 搜索引擎管理 | 新增/编辑/删除搜索引擎并保存，在首页搜索 | 首页搜索跳转到选定外部搜索引擎 | `frontend/src/components/settings/SearchEngineManagerSection.vue`、`frontend/src/composables/useNavigation.js`、`backend/internal/httpx/settings.go` |
| 设置页滚动 | 在设置弹窗内滚动 | 只滚动弹窗内容，不穿透到首页 | `frontend/src/components/settings/SettingsPanel.vue`、`frontend/src/style.css` |
| 设置页深色主题 | 打开个性化、收藏夹、S3、备份恢复等菜单 | 深色背景图上文字、按钮、输入框和卡片可读 | `frontend/src/components/settings/SettingsPanel.vue`、`frontend/src/style.css` |

### 10.4 收藏夹抽屉

| 功能 | 操作步骤 | 预期结果 | 异常优先排查文件 |
|---|---|---|---|
| 按需加载 | 首次打开首页不点收藏夹；再点击收藏夹按钮 | 首页首次不加载完整收藏夹；打开抽屉后加载根目录 | `frontend/src/composables/useBookmarks.js`、`frontend/src/components/BookmarkDrawer.vue`、`backend/internal/httpx/bookmarks.go` |
| 文件夹树展开 | 展开某个文件夹 | 按需加载子目录，树结构正确 | `frontend/src/composables/useBookmarks.js`、`frontend/src/components/BookmarkFolderTreeNode.vue`、`backend/internal/httpx/bookmarks.go` |
| 书签列表 | 选择某个文件夹 | 右侧显示该文件夹下书签 | `frontend/src/composables/useBookmarks.js`、`frontend/src/components/BookmarkRow.vue`、`backend/internal/httpx/bookmarks.go` |
| 新增/编辑文件夹 | 在抽屉或设置中新增、编辑文件夹 | 名称和层级正确，刷新后保持 | `frontend/src/composables/useEditSave.js`、`frontend/src/composables/useBookmarkActions.js`、`backend/internal/httpx/bookmarks.go` |
| 新增/编辑书签 | 新增书签，填写 URL、标题、备注、favicon | 保存成功，列表展示正确 | `frontend/src/composables/useEditDialog.js`、`frontend/src/composables/useEditSave.js`、`backend/internal/httpx/bookmarks.go` |
| 访问后 favicon 自动补全 | 点击没有真实 favicon 的书签，返回应用并刷新收藏夹 | 不影响打开书签；刷新后该书签显示 favicon | `frontend/src/composables/useBookmarkActions.js`、`frontend/src/lib/api.js`、`backend/internal/httpx/bookmarks.go`、`backend/internal/httpx/metadata.go` |
| 删除确认 | 删除文件夹或书签 | 出现二次确认，确认后永久删除 | `frontend/src/composables/useBookmarkActions.js`、`frontend/src/components/ContextMenu.vue`、`backend/internal/httpx/bookmarks.go` |
| 批量操作 | 进入批量选择，选择多条书签后移动/删除 | 批量操作成功，列表刷新正确 | `frontend/src/composables/useBookmarkActions.js`、`frontend/src/components/MoveDialog.vue`、`backend/internal/httpx/bookmarks.go` |
| 收藏夹搜索 | 在收藏夹抽屉搜索关键词 | 只显示匹配书签，包含路径信息 | `frontend/src/composables/useBookmarks.js`、`backend/internal/httpx/bookmarks.go`、`backend/internal/store/store.go` |
| 收藏夹抽屉视觉 | 打开收藏夹抽屉并浏览树和书签列表 | 按钮风格统一，书签 favicon 尺寸合理，深色主题可读 | `frontend/src/components/BookmarkDrawer.vue`、`frontend/src/components/BookmarkRow.vue`、`frontend/src/style.css` |

### 10.5 收藏夹管理与拖拽

| 功能 | 操作步骤 | 预期结果 | 异常优先排查文件 |
|---|---|---|---|
| 设置页收藏夹草稿 | 在设置页移动、上移、下移、删除文件夹但不保存 | 抽屉和后端数据不立即变化 | `frontend/src/composables/useFolderDrafts.js`、`frontend/src/components/settings/BookmarkManager.vue` |
| 保存收藏夹草稿 | 修改收藏夹结构后点击保存 | 保存后抽屉结构更新，刷新后保持 | `frontend/src/composables/useFolderDrafts.js`、`backend/internal/httpx/bookmarks.go` |
| 防循环移动 | 尝试把文件夹移动到自身或子文件夹下 | 操作被阻止或提示错误 | `frontend/src/composables/useFolderDrafts.js`、`frontend/src/composables/useBookmarkActions.js`、`backend/internal/httpx/bookmarks.go` |
| 书签拖拽排序 | 在同一文件夹内拖动书签排序 | 顺序保存，刷新后保持 | `frontend/src/composables/useDragSort.js`、`backend/internal/httpx/bookmarks.go` |
| 跨文件夹移动 | 将书签移动到其他文件夹 | 目标文件夹显示该书签，原文件夹移除 | `frontend/src/composables/useBookmarkActions.js`、`frontend/src/components/MoveDialog.vue`、`backend/internal/httpx/bookmarks.go` |

### 10.6 导入导出

| 功能 | 操作步骤 | 预期结果 | 异常优先排查文件 |
|---|---|---|---|
| 书签 HTML 导入 | 在备份恢复中上传浏览器导出的书签 HTML | 保留原始目录结构和顺序，导入成功 | `frontend/src/composables/useBackupRestore.js`、`backend/internal/httpx/bookmark_transfer.go`、`backend/internal/httpx/bookmark_import.go` |
| 重复导入 | 再次导入相同书签文件 | 仅按 URL、标题、备注、所属目录判断重复 | `backend/internal/httpx/bookmark_import.go`、`backend/internal/store/store.go` |
| 书签 HTML 导出 | 点击收藏夹导出 | 下载 `bookmarks.html`，浏览器可识别 | `frontend/src/composables/useBackupRestore.js`、`backend/internal/httpx/bookmark_transfer.go` |
| 导航备份导出 | 点击导航备份导出 | 下载 JSON，包含 `version`、`createdAt`、`groups`、`items` | `frontend/src/composables/useBackupRestore.js`、`backend/internal/httpx/navigation_backup.go` |
| 导航备份恢复 | 上传刚导出的导航 JSON | 导航分组和卡片恢复成功 | `backend/internal/httpx/navigation_backup.go`、`backend/internal/store/store.go` |

### 10.7 备份恢复

| 功能 | 操作步骤 | 预期结果 | 异常优先排查文件 |
|---|---|---|---|
| 全局备份下载 | 点击全局备份下载 | 下载 `.tar.gz` 文件 | `frontend/src/composables/useBackupRestore.js`、`backend/internal/httpx/server.go` |
| 全局恢复确认 | 上传全局备份文件并恢复 | 出现覆盖风险确认，确认后恢复 | `frontend/src/composables/useBackupRestore.js`、`frontend/src/components/settings/BackupRestoreSection.vue`、`backend/internal/httpx/server.go` |
| 非法备份文件 | 上传非 `.tar.gz` 或损坏文件 | 返回中文错误，不破坏现有数据 | `backend/internal/httpx/server.go` |

### 10.8 上传、Metadata 与 S3

| 功能 | 操作步骤 | 预期结果 | 异常优先排查文件 |
|---|---|---|---|
| 图片上传 | 在卡片图标或背景图处上传图片 | 文件保存到 `uploads`，前端可预览 | `backend/internal/httpx/assets.go`、`frontend/src/composables/useEditSave.js`、`frontend/src/components/EditDialog.vue` |
| 非图片上传 | 上传文本或其它非图片文件 | 返回“仅支持图片文件”类错误 | `backend/internal/httpx/assets.go` |
| metadata 抓取 | 输入 URL 后点击抓取标题/favicon | 自动填充标题和 favicon，失败不影响手动保存 | `backend/internal/httpx/metadata.go`、`frontend/src/composables/useEditSave.js` |
| S3 连接测试 | 填写 S3 配置后点击测试 | 成功返回 `key/url/size`，失败显示错误 | `backend/internal/httpx/s3.go`、`backend/internal/httpx/settings.go`、相关设置组件 |
| 上传后 S3 同步 | 启用 S3 后上传图片 | 本地保存成功；S3 成功时返回公开 URL，失败时不影响本地上传 | `backend/internal/httpx/assets.go`、`backend/internal/httpx/s3.go` |

### 10.9 响应式与交互一致性

| 功能 | 操作步骤 | 预期结果 | 异常优先排查文件 |
|---|---|---|---|
| 桌面端布局 | 在桌面宽度浏览首页、设置、收藏夹 | 卡片、抽屉、弹窗布局稳定，无重叠 | `frontend/src/components/HomeHero.vue`、`frontend/src/components/settings/SettingsPanel.vue`、`frontend/src/components/BookmarkDrawer.vue`、`frontend/src/style.css` |
| 手机端布局 | 使用移动端宽度打开首页、设置、收藏夹 | 主要按钮可点击，文本不溢出，弹窗可操作 | 同上 |
| 右键菜单 | 右键卡片/书签/文件夹 | 菜单位置正确，点击外部关闭 | `frontend/src/composables/useContextMenu.js`、`frontend/src/components/ContextMenu.vue` |
| 弹窗层级 | 同时打开设置、编辑、移动或确认类弹窗 | 层级正确，滚动不穿透 | `frontend/src/components/settings/SettingsPanel.vue`、`frontend/src/components/EditDialog.vue`、`frontend/src/components/MoveDialog.vue`、`frontend/src/style.css` |
| 深色弹窗可读性 | 打开编辑弹窗、移动弹窗、右键菜单 | 深色半透明面板内标题、表单、按钮和危险操作状态清楚 | `frontend/src/components/EditDialog.vue`、`frontend/src/components/MoveDialog.vue`、`frontend/src/components/ContextMenu.vue`、`frontend/src/style.css` |
| 隐藏滚动条后滚动 | 首页长内容、设置页长内容、收藏夹抽屉左右栏、编辑弹窗长内容分别滚动 | 原生滚动条不可见，但鼠标滚轮、触控板和键盘滚动仍可用 | `frontend/src/style.css` |

### 10.10 后端拆分后接口冒烟

| 接口 | 操作步骤 | 预期结果 | 异常优先排查文件 |
|---|---|---|---|
| `/api/metadata` | 前端触发 metadata 抓取 | 返回 `title/favicon` 或明确错误 | `backend/internal/httpx/metadata.go` |
| `/api/assets/upload` | 前端上传图片 | 返回 `url/name/size/mime/source` | `backend/internal/httpx/assets.go`、`backend/internal/httpx/s3.go` |
| `/api/s3/test` | 前端测试 S3 | 返回 `key/url/size` 或 `502` 错误 | `backend/internal/httpx/s3.go` |
| `/api/bookmark/export` | 点击书签导出 | 返回 HTML 下载 | `backend/internal/httpx/bookmark_transfer.go` |
| `/api/bookmark/import` | 上传书签 HTML | 返回导入结果 | `backend/internal/httpx/bookmark_transfer.go`、`backend/internal/httpx/bookmark_import.go` |
| `/api/bookmarks/favicon/refresh` | 点击无真实 favicon 的书签后触发刷新 | 接口快速返回；后端可为该书签写入 favicon，失败不影响打开书签 | `backend/internal/httpx/bookmarks.go`、`backend/internal/httpx/metadata.go`、`backend/internal/store/store.go` |
| `/api/navigation/backup` | 点击导航备份 | 返回 JSON 下载 | `backend/internal/httpx/navigation_backup.go` |
| `/api/navigation/restore` | 上传导航备份 JSON | 返回 `groups/items` 数量 | `backend/internal/httpx/navigation_backup.go` |
| `/api/backup/download` | 点击全局备份 | 返回 `tar.gz` 下载 | `backend/internal/httpx/server.go` |
| `/api/backup/restore` | 上传全局备份 | 恢复文件或返回明确错误 | `backend/internal/httpx/server.go` |
