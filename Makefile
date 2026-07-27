# =============================================================================
# Xt-PMS 项目根 Makefile
# =============================================================================
# 前端: web/  后端: server/
# =============================================================================

# ANSI 颜色代码
C_GREEN  := \033[0;32m
C_BLUE   := \033[0;34m
C_CYAN   := \033[0;36m
C_YELLOW := \033[1;33m
C_BOLD   := \033[1m
C_RESET  := \033[0m

.PHONY: build build/linux build/server build/server/linux build/server/release build/web \
        run/server run/web \
        test test/server test/web test/cover \
        lint lint/server lint/web \
        clean clean/server clean/web clean/release \
        tidy docker/build docker/up docker/down \
        release release/check \
        setup help

# -------- 构建 --------
# make build       前后端（当前系统 + DevBuild）→ 本地开发
# make build/linux 前后端（Linux     + DevBuild）→ 交叉编译测试
# make release     前后端（Linux     生产模式） → 打包 tar.gz

build: build/server build/web ## 构建前后端（当前系统）
	@echo ""
	@echo "✓ 构建完成（当前系统，开发模式）"
	@echo "  启动：./server/bin/xt-pms"

build/linux: build/server/linux build/web ## 构建前后端（Linux amd64）
	@echo ""
	@echo "✓ 构建完成（Linux amd64，开发测试）"
	@echo "  后端：server/bin/xt-pms"
	@echo "  前端：web/dist/"

build/server: ## 编译后端（当前系统 + DevBuild）
	$(MAKE) -C server build

build/server/linux: ## 编译后端（Linux + DevBuild）
	$(MAKE) -C server build/linux

build/server/release: ## 编译后端（Linux 生产模式）
	$(MAKE) -C server build/release

build/web: ## 构建前端
	@cd web && VITE_APP_VERSION=$(VERSION) npm run build -- --logLevel warn
	@printf "$(C_GREEN)  ✓ 前端构建完成$(C_RESET)\n"

# -------- 本地运行 --------

run/server: ## 启动后端（go run）
	$(MAKE) -C server run

run/web: ## 启动前端（vite dev）
	@cd web && VITE_APP_VERSION=$(VERSION) npm run dev

# -------- 测试 --------

test: test/server test/web ## 运行所有测试

test/server: ## 运行后端测试
	$(MAKE) -C server test/short

test/web: ## 运行前端测试（待实现）
	@cd web && npm run test

test/cover: ## 后端测试覆盖率
	$(MAKE) -C server test/cover

# -------- 代码检查 --------

lint: lint/server lint/web ## 代码检查

lint/server: ## 后端代码检查
	$(MAKE) -C server lint

lint/web: ## 前端代码检查
	@cd web && npm run lint

# -------- 依赖管理 --------

tidy: ## 清理并同步依赖
	@cd server && go mod tidy
	@echo "✓ Go 依赖已同步"

# -------- Docker --------

docker/build: ## 构建 Docker 镜像
	docker compose -f deploy/docker-compose.yml build

docker/up: ## 启动 Docker 服务
	@test -f deploy/.env || { echo "未找到 deploy/.env，已从 .env.example 复制默认配置（请按需修改敏感值）"; cp deploy/.env.example deploy/.env; }
	docker compose -f deploy/docker-compose.yml up -d

docker/down: ## 停止 Docker 服务
	docker compose -f deploy/docker-compose.yml down

docker/logs: ## 查看 Docker 日志
	docker compose -f deploy/docker-compose.yml logs -f

# -------- 清理 --------

clean: clean/server clean/web ## 清理产物

clean/server: ## 清理后端产物
	$(MAKE) -C server clean

clean/web: ## 清理前端产物
	@rm -rf web/dist

# -------- 初始化 --------

setup: ## 项目初始化（安装依赖）
	@cd web && npm install
	@cd server && go mod tidy
	@echo "✓ 项目初始化完成"
	@echo ""
	@echo "下一步："
	@echo "  1. 复制 server/config/config.yaml.example -> server/config/config.yaml"
	@echo "  2. 修改数据库连接信息"
	@echo "  3. make run/server  # 启动后端"
	@echo "  4. make run/web     # 启动前端"

# -------- 发行版打包 --------

VERSION   ?= $(shell cat VERSION 2>/dev/null || echo "dev")
RELEASE   := xt-pms-$(VERSION)-linux-amd64
DIST_DIR  := dist

release/check: ## 打包前检查必要文件是否存在
	@if [ ! -f server/bin/xt-pms ]; then \
		echo "✗ 后端二进制不存在，请先执行 make build/server/linux"; \
		exit 1; \
	fi
	@if [ ! -d web/dist ]; then \
		echo "✗ 前端构建产物不存在，请先执行 make build/web"; \
		exit 1; \
	fi
	@if [ ! -f server/migrations/init_schema.sql ] || [ ! -f server/migrations/init_seed.sql ]; then \
		echo "✗ 数据库初始化脚本不存在"; \
		exit 1; \
	fi
	@if [ ! -f deploy/xt-pms.service ] || [ ! -f deploy/nginx.conf ]; then \
		echo "✗ deploy/ 运维模板不存在"; \
		exit 1; \
	fi
	@echo "✓ 所有文件检查通过"

release: ## 构建发行版 tar.gz（前端先构建，再嵌入后端）
	@echo ""
	@printf "$(C_CYAN)$(C_BOLD)╔══════════════════════════════════════════════════════╗$(C_RESET)\n"
	@printf "$(C_CYAN)$(C_BOLD)║$(C_RESET)  📦 构建发行版 $(C_BOLD)$(RELEASE).tar.gz$(C_RESET)\n"
	@printf "$(C_CYAN)$(C_BOLD)╚══════════════════════════════════════════════════════╝$(C_RESET)\n"
	@echo ""
	@printf "$(C_BLUE)[1/4]$(C_RESET) 🌐 构建前端（Vite）...\n"
	@$(MAKE) --no-print-directory build/web
	@printf "$(C_BLUE)[2/4]$(C_RESET) 🔧 编译后端（Go + 嵌入前端）...\n"
	@$(MAKE) --no-print-directory -C server build/release
	@printf "$(C_BLUE)[3/4]$(C_RESET) 🔍 检查文件完整性...\n"
	@$(MAKE) --no-print-directory release/check
	@printf "$(C_BLUE)[4/4]$(C_RESET) 📦 打包压缩...\n"
	@rm -rf $(DIST_DIR)/$(RELEASE) $(DIST_DIR)/$(RELEASE).tar.gz $(DIST_DIR)/$(RELEASE).sha256
	@mkdir -p $(DIST_DIR)/$(RELEASE)/{web,migrations,system-assets,deploy}
	@# 后端二进制
	@cp server/bin/xt-pms $(DIST_DIR)/$(RELEASE)/
	@# 版本文件
	@echo "$(VERSION)" > $(DIST_DIR)/$(RELEASE)/VERSION
	@# 前端静态文件
	@cp -r web/dist/* $(DIST_DIR)/$(RELEASE)/web/
	@# 数据库初始化脚本
	@cp server/migrations/init_schema.sql $(DIST_DIR)/$(RELEASE)/migrations/
	@cp server/migrations/init_seed.sql $(DIST_DIR)/$(RELEASE)/migrations/
	@# 增量迁移脚本（启动时自动执行）
	@[ -d server/migrations/updates ] && cp -r server/migrations/updates $(DIST_DIR)/$(RELEASE)/migrations/ || true
	@#[ -d server/migrations/data_seeds ] && cp -r server/migrations/data_seeds $(DIST_DIR)/$(RELEASE)/migrations/ || true
	@# 系统资源
	@cp server/system-assets/logo.svg $(DIST_DIR)/$(RELEASE)/system-assets/
	@# 运维模板（systemd + nginx + 一键部署脚本）
	@cp deploy/xt-pms.service $(DIST_DIR)/$(RELEASE)/deploy/
	@cp deploy/nginx.conf $(DIST_DIR)/$(RELEASE)/deploy/
	@cp deploy/deploy.sh $(DIST_DIR)/$(RELEASE)/
	@# ========================================
	@# 部署说明
	@# ========================================
	@echo '# Xt-PMS $(VERSION) 部署指南' > $(DIST_DIR)/$(RELEASE)/DEPLOY.md
	@echo '' >> $(DIST_DIR)/$(RELEASE)/DEPLOY.md
	@echo '---' >> $(DIST_DIR)/$(RELEASE)/DEPLOY.md
	@echo '' >> $(DIST_DIR)/$(RELEASE)/DEPLOY.md
	@echo '## 方式一：面板部署（1Panel / 宝塔）' >> $(DIST_DIR)/$(RELEASE)/DEPLOY.md
	@echo '' >> $(DIST_DIR)/$(RELEASE)/DEPLOY.md
	@echo '### 1. 创建数据库' >> $(DIST_DIR)/$(RELEASE)/DEPLOY.md
	@echo '在面板数据库管理中创建空库（utf8mb4）：' >> $(DIST_DIR)/$(RELEASE)/DEPLOY.md
	@echo '```sql' >> $(DIST_DIR)/$(RELEASE)/DEPLOY.md
	@echo 'CREATE DATABASE production_pms DEFAULT CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;' >> $(DIST_DIR)/$(RELEASE)/DEPLOY.md
	@echo '```' >> $(DIST_DIR)/$(RELEASE)/DEPLOY.md
	@echo '' >> $(DIST_DIR)/$(RELEASE)/DEPLOY.md
	@echo '### 2. 部署后端' >> $(DIST_DIR)/$(RELEASE)/DEPLOY.md
	@echo '```bash' >> $(DIST_DIR)/$(RELEASE)/DEPLOY.md
	@echo 'mkdir -p /opt/xt-pms' >> $(DIST_DIR)/$(RELEASE)/DEPLOY.md
	@echo 'tar -xzf $(RELEASE).tar.gz -C /opt/xt-pms --strip-components=1' >> $(DIST_DIR)/$(RELEASE)/DEPLOY.md
	@echo '```' >> $(DIST_DIR)/$(RELEASE)/DEPLOY.md
	@echo '面板中创建 Go 服务：' >> $(DIST_DIR)/$(RELEASE)/DEPLOY.md
	@echo '- 启动命令：`/opt/xt-pms/xt-pms`' >> $(DIST_DIR)/$(RELEASE)/DEPLOY.md
	@echo '- 端口映射：`9898 → 9898`' >> $(DIST_DIR)/$(RELEASE)/DEPLOY.md
	@echo '' >> $(DIST_DIR)/$(RELEASE)/DEPLOY.md
	@echo '### 3. 完成安装向导' >> $(DIST_DIR)/$(RELEASE)/DEPLOY.md
	@echo '启动后查看面板服务日志，找到安装地址（含 token）：' >> $(DIST_DIR)/$(RELEASE)/DEPLOY.md
	@echo '`http://<服务器IP>:9898?token=xxxxxxxx`' >> $(DIST_DIR)/$(RELEASE)/DEPLOY.md
	@echo '浏览器打开，完成 4 步引导后服务自动重启。' >> $(DIST_DIR)/$(RELEASE)/DEPLOY.md
	@echo '' >> $(DIST_DIR)/$(RELEASE)/DEPLOY.md
	@echo '### 4. 配置前端站点' >> $(DIST_DIR)/$(RELEASE)/DEPLOY.md
	@echo '面板中新建网站（PHP 类型即可，实际只用到 Nginx）：' >> $(DIST_DIR)/$(RELEASE)/DEPLOY.md
	@echo '- 网站目录：`/opt/xt-pms/web`' >> $(DIST_DIR)/$(RELEASE)/DEPLOY.md
	@echo '- 反向代理：`^/api/` → `http://127.0.0.1:9898`' >> $(DIST_DIR)/$(RELEASE)/DEPLOY.md
	@echo '' >> $(DIST_DIR)/$(RELEASE)/DEPLOY.md
	@echo '---' >> $(DIST_DIR)/$(RELEASE)/DEPLOY.md
	@echo '' >> $(DIST_DIR)/$(RELEASE)/DEPLOY.md
	@echo '## 方式二：命令行部署（systemd + Nginx）' >> $(DIST_DIR)/$(RELEASE)/DEPLOY.md
	@echo '' >> $(DIST_DIR)/$(RELEASE)/DEPLOY.md
	@echo '> 适用于 Ubuntu 20.04+ / Debian 11+ / CentOS 7+。' >> $(DIST_DIR)/$(RELEASE)/DEPLOY.md
	@echo '' >> $(DIST_DIR)/$(RELEASE)/DEPLOY.md
	@echo '### 1. 准备环境' >> $(DIST_DIR)/$(RELEASE)/DEPLOY.md
	@echo '```bash' >> $(DIST_DIR)/$(RELEASE)/DEPLOY.md
	@echo '# 安装 Nginx（如未安装）' >> $(DIST_DIR)/$(RELEASE)/DEPLOY.md
	@echo 'sudo apt install -y nginx   # Debian/Ubuntu' >> $(DIST_DIR)/$(RELEASE)/DEPLOY.md
	@echo '# sudo yum install -y nginx  # CentOS/RHEL' >> $(DIST_DIR)/$(RELEASE)/DEPLOY.md
	@echo '' >> $(DIST_DIR)/$(RELEASE)/DEPLOY.md
	@echo '# 连接 MySQL 创建数据库' >> $(DIST_DIR)/$(RELEASE)/DEPLOY.md
	@echo "CREATE DATABASE production_pms DEFAULT CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;" >> $(DIST_DIR)/$(RELEASE)/DEPLOY.md
	@echo '```' >> $(DIST_DIR)/$(RELEASE)/DEPLOY.md
	@echo '' >> $(DIST_DIR)/$(RELEASE)/DEPLOY.md
	@echo '### 2. 部署后端' >> $(DIST_DIR)/$(RELEASE)/DEPLOY.md
	@echo '```bash' >> $(DIST_DIR)/$(RELEASE)/DEPLOY.md
	@echo 'sudo mkdir -p /opt/xt-pms' >> $(DIST_DIR)/$(RELEASE)/DEPLOY.md
	@echo 'sudo tar -xzf $(RELEASE).tar.gz -C /opt/xt-pms --strip-components=1' >> $(DIST_DIR)/$(RELEASE)/DEPLOY.md
	@echo '' >> $(DIST_DIR)/$(RELEASE)/DEPLOY.md
	@echo '# 安装 systemd 服务' >> $(DIST_DIR)/$(RELEASE)/DEPLOY.md
	@echo 'sudo cp /opt/xt-pms/deploy/xt-pms.service /etc/systemd/system/' >> $(DIST_DIR)/$(RELEASE)/DEPLOY.md
	@echo 'sudo systemctl daemon-reload' >> $(DIST_DIR)/$(RELEASE)/DEPLOY.md
	@echo 'sudo systemctl enable xt-pms' >> $(DIST_DIR)/$(RELEASE)/DEPLOY.md
	@echo 'sudo systemctl start xt-pms' >> $(DIST_DIR)/$(RELEASE)/DEPLOY.md
	@echo '```' >> $(DIST_DIR)/$(RELEASE)/DEPLOY.md
	@echo '' >> $(DIST_DIR)/$(RELEASE)/DEPLOY.md
	@echo '### 3. 完成安装向导' >> $(DIST_DIR)/$(RELEASE)/DEPLOY.md
	@echo '```bash' >> $(DIST_DIR)/$(RELEASE)/DEPLOY.md
	@echo '# 查看服务日志，获取安装地址（含 token）' >> $(DIST_DIR)/$(RELEASE)/DEPLOY.md
	@echo 'sudo journalctl -u xt-pms -f' >> $(DIST_DIR)/$(RELEASE)/DEPLOY.md
	@echo '```' >> $(DIST_DIR)/$(RELEASE)/DEPLOY.md
	@echo '浏览器打开 `http://<服务器IP>:9898?token=xxxxxxxx`，完成 4 步引导。' >> $(DIST_DIR)/$(RELEASE)/DEPLOY.md
	@echo '安装完成后服务自动重启。' >> $(DIST_DIR)/$(RELEASE)/DEPLOY.md
	@echo '' >> $(DIST_DIR)/$(RELEASE)/DEPLOY.md
	@echo '### 4. 配置 Nginx' >> $(DIST_DIR)/$(RELEASE)/DEPLOY.md
	@echo '```bash' >> $(DIST_DIR)/$(RELEASE)/DEPLOY.md
	@echo '# 使用包内提供的 Nginx 配置' >> $(DIST_DIR)/$(RELEASE)/DEPLOY.md
	@echo 'sudo cp /opt/xt-pms/deploy/nginx.conf /etc/nginx/sites-available/xt-pms' >> $(DIST_DIR)/$(RELEASE)/DEPLOY.md
	@echo 'sudo ln -s /etc/nginx/sites-available/xt-pms /etc/nginx/sites-enabled/' >> $(DIST_DIR)/$(RELEASE)/DEPLOY.md
	@echo 'sudo nginx -t && sudo systemctl reload nginx' >> $(DIST_DIR)/$(RELEASE)/DEPLOY.md
	@echo '```' >> $(DIST_DIR)/$(RELEASE)/DEPLOY.md
	@echo '访问 `http://<服务器IP>` 即可使用系统。' >> $(DIST_DIR)/$(RELEASE)/DEPLOY.md
	@echo '' >> $(DIST_DIR)/$(RELEASE)/DEPLOY.md
	@echo '---' >> $(DIST_DIR)/$(RELEASE)/DEPLOY.md
	@echo '' >> $(DIST_DIR)/$(RELEASE)/DEPLOY.md
	@echo '## 更新升级' >> $(DIST_DIR)/$(RELEASE)/DEPLOY.md
	@echo '' >> $(DIST_DIR)/$(RELEASE)/DEPLOY.md
	@echo '### 面板用户' >> $(DIST_DIR)/$(RELEASE)/DEPLOY.md
	@echo '1. 面板中停止 Go 服务' >> $(DIST_DIR)/$(RELEASE)/DEPLOY.md
	@echo '2. 替换 `/opt/xt-pms/xt-pms` 二进制和 `/opt/xt-pms/web/` 目录' >> $(DIST_DIR)/$(RELEASE)/DEPLOY.md
	@echo '3. 重启 Go 服务' >> $(DIST_DIR)/$(RELEASE)/DEPLOY.md
	@echo '' >> $(DIST_DIR)/$(RELEASE)/DEPLOY.md
	@echo '### 命令行用户' >> $(DIST_DIR)/$(RELEASE)/DEPLOY.md
	@echo '```bash' >> $(DIST_DIR)/$(RELEASE)/DEPLOY.md
	@echo 'sudo systemctl stop xt-pms' >> $(DIST_DIR)/$(RELEASE)/DEPLOY.md
	@echo 'sudo tar -xzf $(RELEASE).tar.gz -C /opt/xt-pms --strip-components=1' >> $(DIST_DIR)/$(RELEASE)/DEPLOY.md
	@echo 'sudo systemctl start xt-pms' >> $(DIST_DIR)/$(RELEASE)/DEPLOY.md
	@echo '```' >> $(DIST_DIR)/$(RELEASE)/DEPLOY.md
	@# 清理 macOS 资源分支残留（._* / .DS_Store）
	@find $(DIST_DIR)/$(RELEASE) -name '._*' -type f -delete 2>/dev/null || true
	@find $(DIST_DIR)/$(RELEASE) -name '.DS_Store' -type f -delete 2>/dev/null || true
	@# 打包（COPYFILE_DISABLE=1 防止 macOS 再次写入资源分支）
	@cd $(DIST_DIR) && COPYFILE_DISABLE=1 tar -czf $(RELEASE).tar.gz $(RELEASE)/
	@# 生成 sha256 校验和
	@cd $(DIST_DIR) && shasum -a 256 $(RELEASE).tar.gz > $(RELEASE).sha256
	@rm -rf $(DIST_DIR)/$(RELEASE)
	@echo ""
	@printf "$(C_GREEN)$(C_BOLD)╔══════════════════════════════════════════════════════╗$(C_RESET)\n"
	@printf "$(C_GREEN)$(C_BOLD)║$(C_RESET)  $(C_BOLD)✓ 构建成功$(C_RESET)\n"
	@printf "$(C_GREEN)$(C_BOLD)║$(C_RESET)  📄 $(DIST_DIR)/$(RELEASE).tar.gz\n"
	@printf "$(C_GREEN)$(C_BOLD)║$(C_RESET)  🔑 sha256: $$(cat $(DIST_DIR)/$(RELEASE).sha256 | awk '{print $$1}')\n"
	@printf "$(C_GREEN)$(C_BOLD)║$(C_RESET)  📏 大小: $$(du -h $(DIST_DIR)/$(RELEASE).tar.gz | cut -f1)\n"
	@printf "$(C_GREEN)$(C_BOLD)╚══════════════════════════════════════════════════════╝$(C_RESET)\n"
	@echo ""

clean/release: ## 清理发行版产物
	@rm -rf $(DIST_DIR)/*.tar.gz $(DIST_DIR)/*.sha256
	@echo "✓ 发行版清理完成"

help: ## 显示帮助信息
	@grep -E '^[a-zA-Z_/]+:.*?## ' $(MAKEFILE_LIST) | sort | \
		awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-22s\033[0m %s\n", $$1, $$2}'
