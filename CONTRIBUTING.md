# 贡献指南

感谢你对 Xt-PMS 的关注！我们欢迎所有形式的贡献，包括但不限于：

- 提交 Bug 报告
- 功能建议
- 代码贡献
- 文档改进
- 测试用例
- 部署与运维经验分享

## 行为准则

请阅读并遵守 [CODE_OF_CONDUCT.md](./CODE_OF_CONDUCT.md)。

## 如何贡献

### 1. 提交 Issue

在提交 Issue 前，请先检查是否已存在相关 Issue。

- **Bug 报告**：请详细描述问题现象、复现步骤、运行环境（OS、浏览器、数据库版本等）
- **功能建议**：请说明使用场景和期望效果

Issue 可在以下任一平台提交：
- Gitee：https://gitee.com/kamyang-tech/xt-pms/issues
- GitHub：https://github.com/KamyangTech/xt-pms/issues

### 2. 提交 Pull Request

#### 前置准备

```bash
# Fork 并克隆仓库
git clone https://gitee.com/your-username/xt-pms.git
cd xt-pms

# 安装依赖
make setup

# 安装 golangci-lint（用于后端代码检查）
# macOS:
brew install golangci-lint
# Linux: 参考 https://golangci-lint.run/usage/install/
```

#### 开发流程

1. 从最新的 `develop` 分支创建功能分支：
   ```bash
   git checkout -b feat/your-feature-name develop
   ```

2. 遵循项目编码规范：
   - 后端：采用 Controller → Service → DAO 三层架构，遵循 `server/README.md` 中的开发规范
   - 前端：项目已配置 ESLint + Prettier + Stylelint + Commitlint，提交时自动执行检查

3. 确保代码通过全部检查：
   ```bash
   make lint         # 代码检查（后端 + 前端）
   make test         # 运行单元测试
   make test/cover   # 查看测试覆盖率（后端）
   ```

4. Commit 信息遵循 [Conventional Commits](https://www.conventionalcommits.org/) 规范：
   ```
   <type>(<scope>): <description>

   [optional body]

   [optional footer]
   ```

   常用 type：
   - `feat` — 新功能
   - `fix` — Bug 修复
   - `docs` — 文档
   - `style` — 代码风格（不影响功能）
   - `refactor` — 重构
   - `perf` — 性能优化
   - `test` — 测试
   - `chore` — 构建/CI/工具链

   scope 示例：`server`, `web`, `config`, `deploy`, `docs`, `mold`, `production`, `finance`

5. 提交 PR 到 `develop` 分支，填写 [PR 模板](./.github/PULL_REQUEST_TEMPLATE.md)

### 3. 代码审查

所有 PR 需要至少一位维护者 review 后方可合并。审查标准：

- 代码质量与风格
- 测试覆盖率
- 向后兼容性
- 文档更新
- 安全性（无敏感信息泄露、无注入漏洞等）

## 开发环境

### 后端

```bash
cd server
cp config/config.yaml.example config/config.yaml
# 编辑 config.yaml 配置数据库连接信息
make run        # go run cmd/main.go
```

### 前端

```bash
cd web
npm install
npm run dev
# 或使用根目录命令：
make run/web
```

### Docker（可选）

```bash
cp deploy/.env.example deploy/.env
# 编辑 deploy/.env 修改敏感配置
make docker/up
# 访问 http://localhost:9898 完成安装向导
```

## 数据库迁移

如果你添加了新的数据表或修改了表结构：

1. 在 `server/migrations/updates/` 目录下创建增量 SQL 脚本
2. 命名格式：`YYYYMMDD_description.sql`（如 `20260722_add_new_table.sql`）
3. 同时更新 `server/migrations/init_schema.sql` 和 `init_seed.sql`
4. 确保脚本可重复执行（使用 `IF NOT EXISTS` / `IF EXISTS`）

## 项目路线图

参见 [CHANGELOG.md](./CHANGELOG.md) 了解版本记录和规划。

## 问题？

- 在 [Gitee Issues](https://gitee.com/kamyang-tech/xt-pms/issues) 提问
- 或直接联系维护者：**pms@xintest.cn**

---

**感谢你的每一份贡献！**
