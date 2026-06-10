# biu-panel

个人轻量导航站 + 网页收藏夹，面向单机本地调试优先，最终交付再用 Docker 部署，SQLite 本地存储。

## 已支持功能

- 管理员首次初始化、登录、登出、登录失败锁定
- 首页导航分组 / 卡片增删改、排序、内外网地址和打开模式
- 收藏夹左侧抽屉按需加载、文件夹 / 网址收藏增删改、备注、搜索、排序
- 网页标题和 favicon 自动抓取
- 本地图片上传，文件保存到数据目录
- Chrome / Safari 通用书签 HTML 导入导出
- 手动下载 `.tar.gz` 备份、上传备份恢复
- 站点标题、Logo、背景图、背景色、内网检测配置保存

## 本地调试

先启动后端：

```bash
cd backend
BIU_PANEL_DATA_DIR=$(pwd)/../data BIU_PANEL_PORT=55088 go run ./cmd/server
```

再启动前端：

```bash
cd frontend
npm run dev
```

前端开发服务器已配置 `/api` 与 `/uploads` 代理到本机后端，默认可直接访问：

```text
http://localhost:5173
```

如果需要改后端地址，可设置 `VITE_BACKEND_TARGET`。

## 最终 Docker 部署

```bash
docker compose up -d --build
```

默认访问地址：

```text
http://服务器IP:55088
```

数据目录：

```text
./data -> /app/data
```

SQLite 数据库、上传图片、备份相关文件都会保存在 `./data` 下，备份这个目录即可迁移。

## 首次管理员

可以直接打开网页初始化管理员账号；也可以首次启动时用环境变量自动创建：

```yaml
environment:
  BIU_PANEL_ADMIN_USER: admin
  BIU_PANEL_ADMIN_PASSWORD: change-this-password
```

账号创建完成后，建议删除或注释这两个环境变量再重启容器。

## 开发运行

后端：

```bash
cd backend
BIU_PANEL_DATA_DIR=$(pwd)/../data BIU_PANEL_PORT=55088 go run ./cmd/server
```

前端：

```bash
cd frontend
npm run dev
```

默认后端端口：`55088`。前端开发模式会通过 Vite 代理把 `/api` 和 `/uploads` 转到后端。

## 环境变量

- `BIU_PANEL_PORT`：服务端口，默认 `55088`
- `BIU_PANEL_DATA_DIR`：数据目录，Docker 默认 `/app/data`
- `BIU_PANEL_DB`：SQLite 数据库路径，默认 `$BIU_PANEL_DATA_DIR/db/biu-panel.db`
- `BIU_PANEL_STATIC_DIR`：前端静态文件目录，Docker 默认 `/app/public`
- `BIU_PANEL_ADMIN_USER`：可选，首次启动自动创建管理员账号
- `BIU_PANEL_ADMIN_PASSWORD`：可选，首次启动自动创建管理员密码

## 验证命令

```bash
cd backend && go test ./...
cd frontend && npm run build
```

Docker 镜像构建命令：

```bash
docker build -t biu-panel:latest .
```

本项目已验证 `docker build`、容器启动、健康检查、登录、创建导航分组和读取导航数据。

## S3 / OSS 配置

在设置页填写 S3 兼容存储配置后，可以使用“测试 S3”按钮验证上传链路。

hi168 OSS 示例：

- Endpoint：`https://s3.hi168.com`
- Region：`hi168`
- Bucket：填写控制台里的真实桶名，例如 `hi168-24449-xxxxxx`
- Path-style：建议开启
- 上传前缀：`biu-panel/`
- 公开访问域名：没有自定义域名时可以留空

不要把 Access Key / Secret Key 写入 `docker-compose.yml`、README 或公开仓库；在网页设置页填写即可。
