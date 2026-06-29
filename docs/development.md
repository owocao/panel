# biu-panel 开发说明

> 整理时间：2026-06-26
> 当前说明以本地直接运行前后端为主，Docker 用于交付验证。

## 1. 环境要求

需要：

- Go
- Node.js / npm
- Docker 和 Docker Compose（仅 Docker 验证或部署时需要）

本项目已有隔离安装的本地工具链：

```bash
export PATH=/project/panel/.tools/go/bin:/project/panel/.tools/node/bin:$PATH
```

当前已确认版本：

- Go：`go1.25.4 linux/amd64`
- Node.js：`v22.21.1`
- npm：`10.9.4`

## 2. 本地开发

后端：

```bash
cd /project/panel/backend
BIU_PANEL_DATA_DIR=/project/panel/data BIU_PANEL_PORT=55088 go run ./cmd/server
```

前端：

```bash
cd /project/panel/frontend
npm run dev
```

Vite 开发服务器会把 `/api` 和 `/uploads` 代理到后端。代理目标可通过 `VITE_BACKEND_TARGET` 覆盖。

## 3. 常用验证命令

后端测试：

```bash
cd /project/panel/backend
PATH=/project/panel/.tools/go/bin:/project/panel/.tools/node/bin:$PATH \
GOCACHE=/project/panel/.cache/go-build \
GOMODCACHE=/project/panel/.cache/go-mod \
go test ./...
```

后端构建检查：

```bash
cd /project/panel/backend
PATH=/project/panel/.tools/go/bin:/project/panel/.tools/node/bin:$PATH \
GOCACHE=/project/panel/.cache/go-build \
GOMODCACHE=/project/panel/.cache/go-mod \
go build ./...
```

前端构建：

```bash
cd /project/panel/frontend
PATH=/project/panel/.tools/go/bin:/project/panel/.tools/node/bin:$PATH npm run build
```

Docker 构建和启动：

```bash
cd /project/panel
docker compose up -d --build
```

健康检查：

```bash
curl -fsS http://127.0.0.1:55088/api/health
```

Smoke 脚本：

```bash
cd /project/panel
./scripts/smoke.sh
```

## 4. 当前 API 能力

后端已实现：

- 健康检查。
- 首次初始化管理员。
- 环境变量首次创建管理员。
- 登录、登出、当前用户。
- 登录失败锁定和登录日志。
- 过期 session 和旧登录日志清理。
- 导航分组和卡片 CRUD。
- 导航排序。
- 导航页备份和恢复。
- 收藏夹文件夹和书签 CRUD。
- 收藏夹搜索。
- 书签 HTML 导入导出。
- metadata 标题和 favicon 抓取。
- 图片上传。
- S3 连接测试和图片上传同步。
- 系统设置读写。
- 全局备份和恢复。

前端已实现：

- 登录页和初始化页。
- 首页导航分组和卡片。
- 卡片右键菜单和编辑弹窗。
- 系统设置弹窗。
- 设置草稿机制。
- 分组管理草稿机制。
- 搜索引擎管理。
- 个性化设置。
- 备份恢复入口。
- 收藏夹抽屉、文件夹树、书签列表、搜索、批量操作。
- 书签导入导出。
- 图片上传和 metadata 抓取入口。

## 5. 开发注意事项

- 开发前先读 `AGENTS.md` 和 `docs/current-direction.md`。
- 不要让首页首次加载完整收藏夹数据。
- 不要绕过设置页草稿机制。
- 不要恢复已取消的 Logo、独立导入导出菜单、自动内外网检测入口。
- 新增 UI 控件要复用现有按钮、输入框、弹窗、上传按钮样式。
- 修改后端接口时同步前端 API 封装和相关测试。
- 修改数据库结构前必须设计兼容旧库的迁移方式。
- 较大功能完成后同步 `docs/current-direction.md`、`docs/project-state.md`、`docs/tasks.md`。

## 6. 最近验证结果

2026-06-26 使用项目内隔离工具链执行：

```bash
cd /project/panel/backend
PATH=/project/panel/.tools/go/bin:/project/panel/.tools/node/bin:$PATH \
GOCACHE=/project/panel/.cache/go-build \
GOMODCACHE=/project/panel/.cache/go-mod \
go test ./...

cd /project/panel/frontend
PATH=/project/panel/.tools/go/bin:/project/panel/.tools/node/bin:$PATH npm run build
```

结果：

- 后端测试通过。
- 后端构建通过。
- 前端构建通过。
- 首次后端测试需要下载缺失 Go 模块；在受限沙箱中需要允许网络访问，之后依赖会进入 `/project/panel/.cache/go-mod`。
