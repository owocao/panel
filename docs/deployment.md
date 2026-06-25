# Docker 部署操作指南

本文说明如何在另一台服务器上使用 Docker 快速部署 `biu-panel`。

## 1. 前置要求

服务器需要提前安装：

- Docker
- Docker Compose 插件，即可使用 `docker compose` 命令
- Git

建议服务器开放端口：

- 默认：`55088/tcp`
- 如果使用反向代理，只需要让反向代理访问容器端口即可，公网可只开放 `80/443`

## 2. 拉取项目

```bash
git clone https://github.com/owocao/panel.git
cd panel
```

如需部署指定稳定版本：

```bash
git checkout v0.1-stable
```

## 3. 配置环境变量

复制示例环境变量文件：

```bash
cp .env.example .env
```

默认配置可直接使用：

```env
BIU_PANEL_HOST_PORT=55088
TZ=Asia/Shanghai
```

如果服务器 `55088` 已被占用，只修改宿主机端口即可，例如：

```env
BIU_PANEL_HOST_PORT=18080
TZ=Asia/Shanghai
```

容器内部端口仍固定为 `55088`，不建议修改。

## 4. 启动服务

首次构建并启动：

```bash
docker compose up -d --build
```

查看容器状态：

```bash
docker compose ps
```

健康检查：

```bash
curl -fsS http://127.0.0.1:${BIU_PANEL_HOST_PORT:-55088}/api/health
```

正常会返回：

```json
{"status":"ok"}
```

浏览器访问：

```text
http://服务器IP:55088
```

如果 `.env` 中改过端口，请使用对应端口。

## 5. 首次初始化

推荐方式：打开网页后按页面提示创建管理员账号。

也可以首次启动时使用环境变量自动创建管理员账号。编辑 `.env`：

```env
BIU_PANEL_ADMIN_USER=your-admin-name
BIU_PANEL_ADMIN_PASSWORD=change-this-password-before-first-start
```

然后启动：

```bash
docker compose up -d --build
```

账号创建完成后，建议删除或注释 `.env` 中的管理员初始化变量，并重启容器：

```bash
docker compose up -d
```

## 6. 数据目录

当前 Docker 部署采用单目录挂载：

```text
./data -> /app/data
```

该目录包含：

- SQLite 数据库：`data/db/biu-panel.db`
- 上传文件：`data/uploads/`
- 备份和其他运行数据

迁移服务器时，保留并迁移整个 `data/` 目录即可。

不要把 `data/` 提交到 Git 仓库。

## 7. 停止和重启

停止服务：

```bash
docker compose stop biu-panel
```

重启服务：

```bash
docker compose restart biu-panel
```

停止并移除容器：

```bash
docker compose down
```

`docker compose down` 不会删除 `./data` 目录。

## 8. 更新版本

更新前建议先备份数据目录或使用应用内备份。

```bash
cd panel
git pull
docker compose build biu-panel
docker compose up -d --force-recreate biu-panel
curl -fsS http://127.0.0.1:${BIU_PANEL_HOST_PORT:-55088}/api/health
```

如需更新到指定标签：

```bash
git fetch --tags
git checkout v0.1-stable
docker compose up -d --build
```

## 9. 备份和恢复

优先推荐登录系统后使用应用内备份功能。

如需手动备份数据目录，建议先停止容器，避免运行中的 SQLite 文件出现一致性风险：

```bash
docker compose stop biu-panel
tar -czf /root/biu-panel-data-$(date +%Y%m%d-%H%M%S).tar.gz data
docker compose up -d biu-panel
```

手动恢复流程：

```bash
docker compose stop biu-panel
tar -czf /root/biu-panel-data-before-restore-$(date +%Y%m%d-%H%M%S).tar.gz data
rm -rf data
tar -xzf /path/to/biu-panel-data.tar.gz
docker compose up -d biu-panel
curl -fsS http://127.0.0.1:${BIU_PANEL_HOST_PORT:-55088}/api/health
```

恢复前请确认备份包来源可信。

## 10. 查看日志

```bash
docker logs --tail=200 biu-panel
```

持续查看：

```bash
docker logs -f --tail=200 biu-panel
```

Compose 已配置日志轮转：单文件最大 `10m`，最多保留 `3` 个文件。

## 11. docker run 示例

如果不使用 Docker Compose，也可以直接运行：

```bash
docker build -t biu-panel:latest .
docker run -d \
  --name biu-panel \
  --restart unless-stopped \
  -p 55088:55088 \
  -e BIU_PANEL_PORT=55088 \
  -e BIU_PANEL_DATA_DIR=/app/data \
  -e BIU_PANEL_STATIC_DIR=/app/public \
  -e TZ=Asia/Shanghai \
  -v $(pwd)/data:/app/data \
  biu-panel:latest
```

查看健康状态：

```bash
curl -fsS http://127.0.0.1:55088/api/health
```

## 12. 反向代理建议

项目本身只提供 HTTP。公网部署建议通过 Nginx Proxy Manager、Nginx 或 Caddy 提供 HTTPS。

### Nginx Proxy Manager

新增 Proxy Host：

- Domain Names：你的域名，例如 `panel.example.com`
- Scheme：`http`
- Forward Hostname / IP：服务器内网 IP 或 `127.0.0.1`
- Forward Port：`55088`
- 开启 Websockets Support 可选
- SSL 页签申请 Let's Encrypt 证书并开启 Force SSL

### Nginx

```nginx
server {
    listen 80;
    server_name panel.example.com;

    location / {
        proxy_pass http://127.0.0.1:55088;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
    }
}
```

### Caddy

```caddyfile
panel.example.com {
    reverse_proxy 127.0.0.1:55088
}
```

当前版本不支持子路径部署，例如 `https://example.com/panel/`。请使用独立域名或子域名。

## 13. 安全提示

- 本项目定位为私人使用，不提供公开浏览模式。
- 除 `/api/health`、初始化和登录接口外，业务 API 需要登录后访问。
- `/uploads/` 当前作为静态资源访问，暂未纳入登录鉴权。
- 请使用强密码。
- 不要把 `.env`、`data/`、数据库、备份包、S3 密钥提交到 Git 仓库。
- 公网部署建议启用 HTTPS 反向代理。
