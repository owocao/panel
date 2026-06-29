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
