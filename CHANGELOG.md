# 更新日志

所有显著的变更都将记录在此文件中。

格式基于 [Keep a Changelog](https://keepachangelog.com/zh-CN/1.0.0/)，
并遵循 [Semantic Versioning](https://semver.org/lang/zh-CN/)。

## [0.1.0-beta.1] - 2026-05-18

### Added

- 用户管理（登录、JWT 认证、CRUD）
- 角色管理
- 部门管理
- 菜单权限管理
- 客户管理
- 物料管理
- 设备管理
- 仓库管理
- 工序模板管理
- 供应商管理
- 模具档案管理
- 模具订单管理
- 模具 BOM 管理
- 模具工单管理（多工序流转）
- 模具试模管理
- 模具库存/交付管理
- 产品档案管理
- 产品 BOM 管理（自动计算总成本）
- 销售订单管理
- 生产工单管理（自动进度更新）
- 生产报工管理（产量自动累加）
- 库存管理（智能入库、出库校验、安全库存预警）
- 模具统计报表
- 生产统计报表
- Gin 路由框架 + 分层架构（Controller/Service/DAO）
- GORM 数据库 ORM + 连接池配置
- JWT 认证中间件
- CORS 跨域中间件
- Zap 日志系统（分级日志 + 文件轮转）
- 环境变量配置覆盖（支持 K8s/容器部署）
- 定时任务（文件清理、操作日志清理）
- 优雅关机（HTTP + WebSocket + 定时任务）
- pprof 性能分析（开发模式）
- ESLint v9 + Prettier 前端代码规范
- Stylelint CSS/Less 校验
- Commitlint + Husky + Lint-staged 提交规范
- golangci-lint 后端静态检查
- EditorConfig 跨编辑器风格统一
- Docker 多阶段构建
- Docker Compose（MySQL + Server + Nginx）
- GitHub Actions CI 工作流
- Nginx 反向代理配置
- 环境变量模板（.env.example）
- 贡献指南
- 行为准则
- 安全策略
- PR 模板
- Node 版本锁定 (.nvmrc)

### Fixed

- 无（首次发布）

[0.1.0-beta.1]: https://github.com/KamyangTech/xt-pms/releases/tag/v0.1.0-beta.1
