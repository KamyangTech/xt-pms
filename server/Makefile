# =============================================================================
# xt-pms 后端 Makefile
# =============================================================================
# 使用方式：
#   make build       编译二进制
#   make run         启动服务
#   make test        运行测试
#   make test/cover  测试覆盖率
#   make lint        代码静态检查（需安装 golangci-lint）
#   make clean       清理产物
# =============================================================================

APP_NAME    ?= xt-pms
BUILD_DIR   ?= ./bin
MAIN_PATH   := ./cmd

VERSION     ?= $(shell cat ../VERSION 2>/dev/null || echo "dev")
BUILD_TIME  ?= $(shell date -u +"%Y-%m-%dT%H:%M:%SZ")
GIT_COMMIT  ?= $(shell git rev-parse --short HEAD 2>/dev/null || echo "unknown")

LD_FLAGS    := -s -w \
	-X 'main.Version=$(VERSION)' \
	-X 'main.BuildTime=$(BUILD_TIME)' \
	-X 'main.GitCommit=$(GIT_COMMIT)'

# 本地开发编译额外标记（跳过安装向导）
LD_FLAGS_DEV := $(LD_FLAGS) -X 'main.DevBuild=true'

.PHONY: build build/linux build/release build/race run run/prod test test/short test/cover test/pkg lint lint/fix vet tidy clean help

build: ## 编译二进制（当前系统，跳过安装）
	@mkdir -p $(BUILD_DIR)
	go build -ldflags "$(LD_FLAGS_DEV)" -o $(BUILD_DIR)/$(APP_NAME) $(MAIN_PATH)
	@echo "✓ 编译完成：$(BUILD_DIR)/$(APP_NAME)  version=$(VERSION)  commit=$(GIT_COMMIT)"

build/linux: ## 编译为 Linux amd64（开发测试，跳过安装）
	@mkdir -p $(BUILD_DIR)
	GOOS=linux GOARCH=amd64 go build -ldflags "$(LD_FLAGS_DEV)" -o $(BUILD_DIR)/$(APP_NAME) $(MAIN_PATH)
	@echo "✓ 编译完成（Linux amd64）：$(BUILD_DIR)/$(APP_NAME)  version=$(VERSION)  commit=$(GIT_COMMIT)"

build/release: ## 编译为 Linux amd64（生产发布，含前端嵌入）
	@mkdir -p $(BUILD_DIR)
	@rm -rf ./cmd/web-dist
	@cp -r ../web/dist ./cmd/web-dist
	@GOOS=linux GOARCH=amd64 go build -tags release -ldflags "$(LD_FLAGS)" -o $(BUILD_DIR)/$(APP_NAME) $(MAIN_PATH)
	@rm -rf ./cmd/web-dist
	@printf "\033[0;32m  ✓ 编译完成 (Linux amd64, 含前端嵌入)  ver=%s  commit=%s\033[0m\n" $(VERSION) $(GIT_COMMIT)

build/race: ## 编译（启用竞态检测）
	@mkdir -p $(BUILD_DIR)
	go build -race -ldflags "$(LD_FLAGS)" -o $(BUILD_DIR)/$(APP_NAME) $(MAIN_PATH)
	@echo "✓ 编译完成（race）：$(BUILD_DIR)/$(APP_NAME)  version=$(VERSION)  commit=$(GIT_COMMIT)"

run: ## 启动服务（development 模式）
	go run $(MAIN_PATH)

run/prod: build ## 以 production 模式启动
	PMS_APP_MODE=production $(BUILD_DIR)/$(APP_NAME)

test: ## 运行所有测试
	go test -count=1 -v ./...

test/short: ## 运行测试（跳过集成测试）
	go test -count=1 -short -v ./...

test/cover: ## 测试覆盖率
	@mkdir -p $(BUILD_DIR)
	go test -count=1 -coverprofile=$(BUILD_DIR)/cover.out ./...
	go tool cover -func=$(BUILD_DIR)/cover.out
	@echo ""
	@echo "--- HTML 覆盖率报告：$(BUILD_DIR)/cover.html ---"
	go tool cover -html=$(BUILD_DIR)/cover.out -o $(BUILD_DIR)/cover.html

test/pkg: ## 运行指定包的测试（用法：make test/pkg PKG=./internal/service）
	go test -count=1 -v $(PKG)

lint: ## 代码静态检查（需安装 golangci-lint）
	@command -v golangci-lint >/dev/null 2>&1 || { \
		echo "✗ 请先安装 golangci-lint："; \
		echo "  brew install golangci-lint"; \
		exit 1; \
	}
	golangci-lint run -c .golangci.yml ./... --timeout=5m

lint/fix: ## 自动修复 lint 问题
	@command -v golangci-lint >/dev/null 2>&1 || { \
		echo "✗ 请先安装 golangci-lint："; \
		echo "  brew install golangci-lint"; \
		exit 1; \
	}
	golangci-lint run -c .golangci.yml ./... --timeout=5m --fix

vet: ## 运行 go vet
	go vet ./...

tidy: ## 清理 Go 依赖
	go mod tidy
	@echo "✓ go mod tidy 完成"

clean: ## 清理编译产物
	@rm -rf $(BUILD_DIR)
	@echo "✓ 清理完成"

help: ## 显示帮助信息
	@grep -E '^[a-zA-Z_/]+:.*?## ' $(MAKEFILE_LIST) | sort | \
		awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-20s\033[0m %s\n", $$1, $$2}'
