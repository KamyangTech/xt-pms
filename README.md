<div align="center">
  <h1>Xt-PMS</h1>
  <p><strong>专为小微模具制造与注塑企业打造的生产管理系统</strong></p>

  <p>
    <a href="https://gitee.com/kamyang-tech/xt-pms/blob/main/LICENSE"><img src="https://img.shields.io/badge/License-AGPL--3.0-blue" alt="License"></a>
    <a href="https://gitee.com/kamyang-tech/xt-pms/releases"><img src="https://img.shields.io/github/v/release/KamyangTech/xt-pms?include_prereleases" alt="Version"></a>
    <a href="https://go.dev/"><img src="https://img.shields.io/badge/Go-1.26-00ADD8?logo=go" alt="Go"></a>
    <a href="https://vuejs.org/"><img src="https://img.shields.io/badge/Vue-3.5-4FC08D?logo=vue.js" alt="Vue"></a>
    <a href="https://tdesign.tencent.com/"><img src="https://img.shields.io/badge/TDesign-1.13-0052D9" alt="TDesign"></a>
    <a href="https://gitee.com/kamyang-tech/xt-pms/pulls"><img src="https://img.shields.io/badge/PRs-welcome-brightgreen" alt="PRs Welcome"></a>
  </p>
</div>

---

## 目录

- [项目简介](#项目简介)
- [核心亮点](#核心亮点)
- [功能特性](#功能特性)
- [项目结构](#项目结构)
- [技术栈](#技术栈)
- [环境要求](#环境要求)
- [快速开始](#快速开始)
- [部署说明](#部署说明)
- [开发指南](#开发指南)
- [更新日志](#更新日志)
- [贡献指南](#贡献指南)
- [许可证](#许可证)

---

## 项目简介

**Xt-PMS**（Production Management System）是一款面向国内小微模具制造与注塑企业的生产管理系统。系统紧贴行业实际需求，覆盖**订单录入 → 模具档案 → 工序管理 → 生产报工 → 产品交付**全流程环节，助力企业实现生产流程的数字化、透明化与规范化管理。

### 业务闭环

| 阶段 | 销售订单 | 生产工单 | 生产报工 | 库存管理 |
|------|---------|---------|---------|---------|
| 流转 | 客户需求录入 | 生产计划排期 | 进度跟踪 | 出入库管理 |
| 关联 | 客户 / 模具订单 | 模具工单 / 生产订单 | 报工记录 | 库存预警 |

### 系统架构

```
┌─────────────────────────────────────────────────────┐
│                    浏览器 / 客户端                    │
├─────────────────────────────────────────────────────┤
│  Nginx (反向代理 / 静态资源)                          │
├─────────────────────────────────────────────────────┤
│  Vue 3 SPA (TDesign / ECharts / Pinia)              │
├─────────────────────────────────────────────────────┤
│  Go + Gin API Server (:9898)                        │
│  ┌──────────┬──────────┬──────────┬──────────┐      │
│  │Controller│ Service  │   DAO    │   Job    │      │
│  │ JWT Auth │ Validator│  GORM    │  Cron    │      │
│  └──────────┴──────────┴──────────┴──────────┘      │
│  ┌──────────────────────────────────────────┐       │
│  │     WebSocket (实时通知)                   │       │
│  └──────────────────────────────────────────┘       │
├─────────────────────────────────────────────────────┤
│  MySQL 8.0 (60+ 业务表)                             │
└─────────────────────────────────────────────────────┘
```

---

## 核心亮点

- **安装向导**：首次部署自动引导完成数据库配置与管理员初始化，4 步即可上线
- **RBAC 权限**：细粒度菜单/按钮权限控制，支持角色/部门/用户三级管理
- **实时通知**：WebSocket 推送，任务进度、库存预警实时触达
- **安全加固**：JWT 无状态认证 + bcrypt 密码加密 + IP 登录限流 + 验证码 + XSS 过滤 + 安全响应头
- **自动化任务**：定时清理孤立文件与操作日志，维护系统健康
- **在线更新**：支持从 Gitee Release 检查更新并一键升级
- **生产模式**：后端内嵌前端 SPA，单个二进制即可独立运行
- **外发协同**：加工商可通过分享链接免登录报工
- **打印模板**：支持送货单、通知单等自定义打印配置

---

## 功能特性

### 系统管理
- 用户管理（登录 / 注册 / 密码重置 / 账户锁定）
- 角色管理与菜单/按钮权限分配
- 部门组织架构
- 数据字典（系统字典 + 业务字典）
- 公司信息配置
- 操作日志与安全设置
- 系统监控面板（运行时指标 / 连接池状态）

### 基础资料
- 客户管理 / 物料管理
- 设备管理 / 仓库管理
- 工序模板 / 供应商管理 / 加工商管理

### 模具管理 ⭐
| 模块 | 功能描述 |
|------|---------|
| 模具档案 | 模具基本信息、编码管理、统计看板 |
| 模具订单 | 客户订单创建、全生命周期状态流转 |
| 模具工单 | 生产计划排期、多工序流转管理 |
| 工单工序 | 工序勾选完成、进度实时更新 |
| 模具试模 | 试模记录与结果追踪 |
| 模具交付 | 交付记录与验收管理 |
| 模具库存 | 入库 / 出库 / 库存预警 |
| 模具图片 | 附件上传、主图管理 |
| 模具订单款项 | 收款记录、按工单统计 |

### 产品管理 ⭐
| 模块 | 功能描述 |
|------|---------|
| 产品档案 | 产品基本信息、编码自动生成 |
| 产品 BOM | 物料清单、**自动计算总成本** |
| 产品库存 | 入库 / 出库、**安全库存预警** |

### 生产管理 ⭐
| 模块 | 功能描述 |
|------|---------|
| 生产订单 | 全生命周期（创建 → 启动 → 完成 → 入库 → 发货 → 完结 → 补发） |
| 生产报工 | 报工提交、**产量自动累加** |
| 外发管理 | 外发记录、发送 / 接收 / 报工，支持加工商免登录协同 |
| 模具使用记录 | 使用历史与统计 |
| 库存管理 | 智能入库 / 出库校验、库存趋势、快递单号识别 |
| 外来模具档案 | 外部模具编码管理 |

### 综合订单
- 模具 + 产品合并下单，统一管理

### 财务管理
- 应收账款（收款记录、核销）
- 应付账款（付款记录）
- 费用支出管理
- 收款账户管理

### 数据看板
- 模具统计 / 生产统计 / 财务概览（ECharts 图表）

### 其他
- 全局搜索（跨模块全文检索）
- 文件上传（通用文件上传、缩略图自动生成、引用计数清理）

---

## 项目结构

```
xt-pms/
├── server/                          # 后端项目（Go + Gin）
│   ├── cmd/                         # 应用入口（启动、优雅关机、安装向导、SPA 嵌入）
│   ├── config/                      # 配置管理（YAML + 环境变量覆盖）
│   ├── internal/
│   │   ├── controller/              # 控制器层（30+ 模块）
│   │   ├── service/                 # 业务逻辑层
│   │   ├── dao/                     # 数据访问层（GORM）
│   │   ├── model/                   # 数据模型（Entity + DTO）
│   │   ├── router/                  # 模块化路由注册（11 个路由文件）
│   │   ├── common/
│   │   │   ├── middleware/          # 中间件（JWT / CORS / 权限 / 限流 / XSS / 日志）
│   │   │   ├── response/           # 统一响应格式
│   │   │   ├── monitor/            # 运行时监控
│   │   │   └── error/              # 自定义错误
│   │   ├── installer/               # Web 安装向导（4 步引导）
│   │   ├── job/                     # 定时任务（文件清理、日志清理）
│   │   └── utils/                   # 工具函数
│   ├── pkg/logger/                  # 日志封装（Zap + Lumberjack）
│   └── migrations/                  # 数据库初始化、种子数据、增量迁移
│
├── web/                             # 前端项目（Vue 3 + TDesign）
│   └── src/
│       ├── api/                     # API 接口封装
│       ├── pages/                   # 页面组件（130+ .vue 文件）
│       │   ├── login/               # 登录/注册
│       │   ├── dashboard/           # 数据看板
│       │   ├── mold/                # 模具管理
│       │   ├── production/          # 生产管理
│       │   ├── order-management/    # 综合订单
│       │   ├── outsourcing/         # 外发管理
│       │   ├── finance/             # 财务管理
│       │   ├── basic/               # 基础资料
│       │   └── system/              # 系统设置
│       ├── router/                  # 路由配置（模块化拆分）
│       ├── store/                   # Pinia 状态管理 + 持久化
│       ├── components/              # 公共组件
│       ├── layouts/                 # 布局组件
│       ├── locales/                 # 国际化（vue-i18n）
│       └── utils/                   # 工具函数
│
├── deploy/                          # 部署配置
│   ├── Dockerfile                   # 多阶段构建（Go 编译 + Alpine 运行）
│   ├── docker-compose.yml           # Docker Compose 编排（MySQL + Server + Nginx）
│   ├── docker-compose.dev.yml       # 本地开发编排
│   ├── docker-entrypoint.sh         # 容器入口脚本
│   ├── docker-deploy.sh             # Docker 一键部署脚本
│   ├── deploy.sh                    # 裸机一键部署脚本
│   ├── quick-install.sh             # 快速安装脚本
│   ├── nginx.conf                   # Nginx 反向代理配置
│   ├── xt-pms.service               # systemd 服务单元
│   └── .env.example                 # 环境变量模板
│
├── scripts/                         # 辅助脚本
├── CONTRIBUTING.md                  # 贡献指南
├── CODE_OF_CONDUCT.md               # 行为准则
├── SECURITY.md                      # 安全策略
├── CHANGELOG.md                     # 更新日志
├── Makefile                         # 统一构建入口
└── VERSION                          # 版本号
```

---

## 技术栈

| 层次 | 技术 | 版本 |
|------|------|------|
| 前端框架 | Vue 3 (Composition API) | 3.5.13 |
| 构建工具 | Vite | 6.x |
| UI 组件库 | TDesign Vue Next | 1.13.x |
| 语言 | TypeScript | ~5.4 |
| 状态管理 | Pinia + persistedstate | 2.1.7 |
| 路由 | Vue Router | ~4.3 |
| HTTP 客户端 | Axios | 1.8.x |
| 图表 | ECharts | 5.4.3 |
| 国际化 | vue-i18n | v9 |
| 拖拽 | SortableJS / vuedraggable | — |
| 二维码 | qrcode.vue | ^3.4 |
| CSS 预处理 | Less | — |
| 后端框架 | Go + Gin | 1.26.1 / 1.12.0 |
| ORM | GORM + MySQL Driver | 1.31.1 |
| 数据库 | MySQL | 8.0 |
| 认证 | JWT (golang-jwt v5) | — |
| 密码加密 | bcrypt + Pepper | — |
| 日志 | Zap + Lumberjack | 1.27.1 |
| 定时任务 | robfig/cron v3 | — |
| WebSocket | gorilla/websocket | 1.5.3 |
| 验证码 | base64Captcha | — |
| UUID | google/uuid | 1.6.0 |
| 图片处理 | golang.org/x/image | 0.13.0 |
| 容器化 | Docker + Docker Compose | — |
| CI/CD | GitHub Actions | — |

### 部署架构

系统支持三种部署方式，灵活适配不同场景：

| 方式 | 适用场景 | 说明 |
|------|---------|------|
| **Docker Compose** | 快速体验 / 中小规模 | 一键部署 MySQL + Server + Nginx |
| **裸机 systemd** | 生产环境 | 二进制 + systemd + Nginx 反向代理 |
| **面板部署** | 1Panel / 宝塔 | 创建 Go 服务 + 网站反向代理 |

---

## 环境要求

| 依赖 | 版本要求 |
|------|---------|
| Node.js | >= 18.18.0（参考 [.nvmrc](./.nvmrc)） |
| Go | >= 1.26 |
| MySQL | >= 8.0 |

---

## 快速开始

### 环境准备

确保已安装以下依赖：

| 依赖 | 版本要求 | 说明 |
|------|---------|------|
| Node.js | >= 18.18.0 | 前端构建 |
| Go | >= 1.26 | 后端编译 |
| MySQL | >= 8.0 | 数据存储 |

### 本地开发

```bash
# 1. 克隆项目
git clone https://gitee.com/kamyang-tech/xt-pms.git
cd xt-pms

# 2. 一键初始化（安装前后端依赖）
make setup

# 3. 配置数据库连接
cp server/config/config.yaml.example server/config/config.yaml
# 编辑 config.yaml，修改数据库连接信息

# 4. 创建数据库
mysql -u root -p -e "CREATE DATABASE IF NOT EXISTS production_pms DEFAULT CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;"

# 5. 启动后端（首次启动会自动进入安装向导）
make run/server
# 访问 http://localhost:9898?token=xxxxxxxx 完成安装向导

# 6. 启动前端（新终端）
make run/web
# 访问 http://localhost:5173
```

### Docker 快速体验

```bash
# 一键部署
curl -fsSL https://gitee.com/kamyang-tech/xt-pms/raw/main/deploy/docker-deploy.sh | sh

# 或手动部署
cp deploy/.env.example deploy/.env
make docker/up

# 访问 http://localhost:9898 完成安装向导
```

---

## 部署说明

### 方式一：Docker Compose 部署

```bash
# 1. 准备环境变量
cp deploy/.env.example deploy/.env
# 编辑 deploy/.env，修改数据库密码、JWT 密钥等

# 2. 启动服务
make docker/up

# 3. 查看日志
make docker/logs

# 4. 停止服务
make docker/down
```

服务启动后访问 `http://<服务器IP>:9898`，按安装向导完成初始化。

### 方式二：裸机部署（systemd + Nginx）

```bash
# 1. 构建发行版
make release
# 生成 dist/xt-pms-0.1.0-beta.1-linux-amd64.tar.gz

# 2. 上传到服务器并解压
sudo mkdir -p /opt/xt-pms
sudo tar -xzf xt-pms-*.tar.gz -C /opt/xt-pms --strip-components=1

# 3. 安装 systemd 服务
sudo cp /opt/xt-pms/deploy/xt-pms.service /etc/systemd/system/
sudo systemctl daemon-reload
sudo systemctl enable xt-pms
sudo systemctl start xt-pms

# 4. 完成安装向导
sudo journalctl -u xt-pms -f
# 找到安装地址，浏览器打开完成配置

# 5. 配置 Nginx（可选）
sudo cp /opt/xt-pms/deploy/nginx.conf /etc/nginx/sites-available/xt-pms
sudo ln -s /etc/nginx/sites-available/xt-pms /etc/nginx/sites-enabled/
sudo nginx -t && sudo systemctl reload nginx
```

### 方式三：面板部署（1Panel / 宝塔）

1. 面板数据库管理中创建空库 `production_pms`（utf8mb4）
2. 上传发行版并解压到 `/opt/xt-pms`
3. 创建 Go 服务：启动命令 `/opt/xt-pms/xt-pms`，端口 `9898`
4. 创建网站：目录 `/opt/xt-pms/web`，反向代理 `^/api/` → `http://127.0.0.1:9898`

详细的部署指引请参考 [Xt-PMS 部署文档](https://doc.pms.kamyang.com/guide/quick-start.html)。

---

## 开发指南

### 常用命令

| 命令 | 说明 |
|------|------|
| `make setup` | 一键初始化项目依赖 |
| `make run/server` | 启动后端开发服务 |
| `make run/web` | 启动前端开发服务（Vite HMR） |
| `make build` | 构建前后端（当前系统，开发模式） |
| `make build/linux` | 交叉编译 Linux 版本 |
| `make release` | 构建生产发行版 tar.gz |
| `make lint` | 全量代码检查 |
| `make lint/server` | 后端代码检查（golangci-lint） |
| `make lint/web` | 前端代码检查（ESLint + Stylelint） |
| `make test` | 运行全部测试 |
| `make test/server` | 运行后端测试 |
| `make test/cover` | 后端测试覆盖率报告 |
| `make docker/build` | 构建 Docker 镜像 |
| `make docker/up` | 启动 Docker 服务 |
| `make docker/down` | 停止 Docker 服务 |
| `make docker/logs` | 查看 Docker 日志 |
| `make clean` | 清理构建产物 |
| `make tidy` | 同步 Go 依赖 |
| `make help` | 查看所有命令 |

### 代码提交规范

项目配置了 Git 提交规范工具链（Commitlint + Husky）：

```
<type>(<scope>): <description>
```

| type | 说明 |
|------|------|
| feat | 新功能 |
| fix | Bug 修复 |
| docs | 文档变更 |
| refactor | 代码重构 |
| perf | 性能优化 |
| test | 测试相关 |
| chore | 构建 / CI / 工具链 |

### 后端架构

项目采用经典的 **Controller → Service → DAO** 三层架构：

- **Controller**：处理 HTTP 请求，参数校验，调用 Service
- **Service**：核心业务逻辑，事务管理
- **DAO**：GORM 数据访问，连接池管理与慢查询日志

统一的 `HttpError` 错误模型和 `response.Success/Error` 响应格式贯穿所有 API。

路由采用模块化注册，每个业务模块独立管理路由：
- `SetupSystemRoutes` — 系统管理（用户/角色/菜单/登录等）
- `SetupBasicRoutes` — 基础资料（客户/物料/仓库等）
- `SetupMoldRoutes` — 模具管理
- `SetupProductRoutes` — 产品管理
- `SetupProductionRoutes` — 生产管理
- `SetupOrderRoutes` — 综合订单
- `SetupFinanceRoutes` — 财务管理
- `SetupDashboardRoutes` — 数据看板
- `SetupUploadRoutes` — 文件上传
- `SetupReferenceRoutes` — 公共引用数据
- `SetupOutsourcingRoutes` — 外发管理

### 数据库

| 项目 | 说明 |
|------|------|
| 数据库名 | `production_pms` |
| 字符集 | `utf8mb4` |
| 排序规则 | `utf8mb4_unicode_ci` |
| 表数量 | 60+ 张业务表 |
| 初始化脚本 | `server/migrations/init_schema.sql` |
| 种子数据 | `server/migrations/init_seed.sql` |
| 增量迁移 | `server/migrations/updates/` |

### 配置优先级

```
OS 环境变量 > .env 文件 > config.yaml
```

关键环境变量：

| 变量 | 说明 |
|------|------|
| `PMS_DB_HOST` | 数据库主机地址 |
| `PMS_DB_PORT` | 数据库端口 |
| `PMS_DB_USER` | 数据库用户名 |
| `PMS_DB_PASSWORD` | 数据库密码 |
| `PMS_DB_NAME` | 数据库名 |
| `PMS_JWT_SECRET` | JWT 签名密钥 |
| `PMS_SECURITY_PEPPER` | 密码加密 Pepper |
| `PMS_SERVER_PORT` | 服务端口（默认 9898） |
| `PMS_APP_MODE` | 运行模式（development / production） |

### API 端点概览

| 路由前缀 | 说明 | 认证 |
|---------|------|------|
| `/api/ping` | 健康检查 | 公开 |
| `/health` | 负载均衡器健康检查 | 公开 |
| `/api/ws` | WebSocket 实时通知 | 公开（Token 参数验证） |
| `/api/public/outsourcing/` | 加工商免登录报工 | 公开（Token 验证） |
| `/api/debug/metrics` | 运行时指标 | JWT |
| `/api/system/` | 系统管理 | JWT + RBAC |
| `/api/basic/` | 基础资料 | JWT + RBAC |
| `/api/mold/` | 模具管理 | JWT + RBAC |
| `/api/product/` | 产品管理 | JWT + RBAC |
| `/api/production/` | 生产管理 | JWT + RBAC |
| `/api/master-orders/` | 综合订单 | JWT + RBAC |
| `/api/finance/` | 财务管理 | JWT + RBAC |
| `/api/dashboard/` | 数据看板 | JWT + RBAC |
| `/api/upload/` | 文件上传 | JWT + RBAC |
| `/api/reference/` | 公共引用数据 | JWT |
| `/api/search` | 全局搜索 | JWT |

---

## 更新日志

详见 [CHANGELOG.md](./CHANGELOG.md)

---

## 贡献指南

我们欢迎任何形式的贡献！请阅读 [CONTRIBUTING.md](./CONTRIBUTING.md) 了解提交流程与规范。

参与本项目意味着您同意遵守 [CODE_OF_CONDUCT.md](./CODE_OF_CONDUCT.md) 行为准则。

如有安全漏洞，请参考 [SECURITY.md](./SECURITY.md) 中的报告流程。

---

## 许可证

本项目基于 [AGPL-3.0](./LICENSE) 协议开源。

Copyright (c) 2026 [KamyangTech](https://gitee.com/kamyang-tech)
