# biu-panel 开发说明 v0.1

## 本机开发

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

前端开发服务器会通过 Vite 代理把 `/api` 和 `/uploads` 转到本机后端；如需切换后端目标，可通过 `VITE_BACKEND_TARGET` 覆盖。

## 后端当前能力

已实现：

- 配置加载。
- SQLite 数据库初始化。
- 数据表迁移。
- 健康检查：`GET /api/health`。
- 初始化状态：`GET /api/setup/status`。
- 首次初始化管理员：`POST /api/setup`。
- 登录：`POST /api/auth/login`。
- 登出：`POST /api/auth/logout`。
- 当前用户：`GET /api/auth/me`。
- 登录成功/失败日志。
- 连续失败 5 次锁定 15 分钟。
- 导航分组和卡片的创建/列表 API 初版。
- 收藏夹文件夹、网址、搜索 API 初版。

## 前端当前能力

已实现：

- 首页静态原型。
- 收藏夹抽屉静态原型。
- 登录页静态原型。
- 初始化页静态原型。
- 设置页静态原型。
- 右键菜单原型。
- API 客户端基础封装：`frontend/src/lib/api.js`。

## 验证命令

后端：

```bash
cd /project/panel/backend
go test ./...
```

前端：

```bash
cd /project/panel/frontend
npm run build
```
