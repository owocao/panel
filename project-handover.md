# Biu-Panel 项目交接文档

## 0. 新服务器迁移与环境初始化指南

如果需要将项目迁移到全新的服务器，请在进行任何开发前完成以下环境依赖的安装与配置：

### 0.1 安装系统基础依赖
```bash
# 更新软件包列表
sudo apt update

# 安装基础构建工具和版本控制
sudo apt install -y build-essential git curl wget systemd
```

### 0.2 安装 Go 运行环境 (后端)
本项目后端由 Go 语言编写，需要安装 Go：
```bash
# 可以通过 apt 安装或者从官方下载最新版本
sudo apt install -y golang-go

# 验证安装
go version
```

### 0.3 安装 Node.js 与 npm (前端)
本项目前端使用 Vue 3 + Vite 构建，使用 npm 管理包：
```bash
# 推荐使用 Node.js 18.x 或 20.x 版本
curl -fsSL https://deb.nodesource.com/setup_20.x | sudo -E bash -
sudo apt install -y nodejs

# 验证安装
node -v
npm -v
```

### 0.4 拉取代码与编译
```bash
# 克隆仓库包 (替换为实际的仓库地址)
git clone <repository_url> /project/panel
cd /project/panel

# 编译后端
cd backend
go build -o bin/biu-panel ./cmd/server

# 编译前端
cd ../frontend
npm install
npm run build
```

### 0.5 配置 systemd 守护进程
新服务器上需要重新配置 systemd 服务以便后台运行：
```bash
# 将项目 deploy 目录下的 service 文件软链接或复制到 systemd 目录
# 示例：
sudo cp /project/panel/deploy/biu-panel-dev.service /etc/systemd/system/
sudo systemctl daemon-reload
sudo systemctl enable biu-panel-dev.service
sudo systemctl start biu-panel-dev.service
```
----------------------------------------------------------------------


## 1. 项目基本信息
- **项目名称**: `biu-panel`
- **定位**: 个人轻量级 Web 导航面板 + 收藏夹管理器
- **技术栈**: 后端 Go (SQLite), 前端 Vue 3 + Vite, 包管理 npm
- **部署环境**: 当前通过 `systemd` 本地运行在云主机，最终目标是 Docker。
- **访问地址**: `http://111.119.213.77:55088/`
- **当前代码状态**: `master` 分支，工作区已清理并全部提交。

## 2. 当前开发进度总结

### ✅ 已完成的核心功能
1. **首页导航卡片**
   - 卡片分组展示与组内编辑模式。
   - 分组处于编辑模式时，卡片支持 **自定义 Pointer 拖拽排序**（彻底抛弃 HTML5 原生拖拽，解决了浮层难看、重排闪烁、卡顿等问题）。
   - 卡片右键菜单（支持新标签/窗口打开、编辑、删除，图标与文字留白已精准对齐）。
   - 拖拽卡片时有顺滑的挤占动画。
2. **全局内外网访问逻辑**
   - 完全摒弃了单卡片维度的打开模式配置，统一走全局策略。
   - 界面右下角提供全局切换按钮，仅保留 **「优先内网」** 和 **「优先公网」** 两种模式。
   - 访问检测与回退闭环逻辑：
     - **优先内网**：内网地址不为空时先探测内网，超时或不可达则退至公网；内网地址为空则直达公网。
     - **优先公网**：公网地址不为空时先探测公网，超时或不可达则退至内网；公网地址为空则直达内网。
   - 移除不必要的设置，仅保留探测「超时时间（ms）」。

### ⏳ 部分完成 / 待优化的功能
1. **导航卡片编辑弹窗**
   - 弹窗内已移除无效的“打开模式”选项。
   - 弹窗体验可进一步打磨（名称、图标、分组、内外网地址的基础录入已跑通）。
2. **图标上传 / 本地图片选择**
   - 当前允许将图片链接作为图标。本地上传功能需限制 5MB，需完善预览等细节，目前优先级稍后。

### ❌ 未完成 / 待开发的优先级清单
交接后，建议按以下优先级顺序继续推进：

1. **移动端导航页适配**
   - 移动端普通模式点击卡片正常打开。
   - 导航页编辑模式下，要求**长按卡片**进入拖动排序（避免普通点击被误触发）。
2. **全面开发收藏夹模块（左侧抽屉）**
   - 导航页功能稳定后，集中精力做收藏夹。
   - 实现无限级文件夹、树状/列表拖拽排序（含跨文件夹移动）。
   - 浏览器书签导入/导出（保留原有文件夹层级，多次导入时不强制内部去重）。
   - 实现仅针对收藏夹的模糊搜索（与首页搜索引擎搜索分离）。
3. **S3 云存储与系统备份**
   - 实现本地上传文件的 S3 远端同步。
   - 完成系统整体备份为 `.tar.gz` 格式（包含 SQLite 数据、本地配置和版本信息）并支持手动恢复。

## 3. 运行与调试环境交接
- **重新构建前端**
  ```bash
  cd /project/panel/frontend
  npm run build
  ```
- **重启后端服务（应用更改）**
  ```bash
  systemctl restart biu-panel-dev.service
  ```
- **健康检查与状态确认**
  ```bash
  systemctl status biu-panel-dev.service --no-pager
  curl -fsS http://127.0.0.1:55088/api/health
  ```

## 4. 产品与代码核心约定
- **禁止过度封装**：个人项目，代码追求直接、好维护，不引入重型框架。
- **数据库**：仅使用 SQLite 本地单文件（`./data/db`）。
- **拖拽方案**：导航卡片的拖拽已采用 Pointer Event + 实时计算互换。切勿回退到 HTML5 的 `draggable="true"`。