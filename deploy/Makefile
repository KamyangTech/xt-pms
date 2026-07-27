# =============================================================================
# Xt-PMS 部署 Makefile
# =============================================================================
# 使用方式：
#   make up       启动服务
#   make down     停止服务
#   make build    重新构建镜像
#   make logs     查看日志
#   make mysql    进入 MySQL 容器
#   make server   进入 Server 容器
#   make restart  重启服务
# =============================================================================

COMPOSE_FILE  := docker-compose.yml

.PHONY: up down build logs mysql server restart ps

up: ## 启动所有服务
	docker compose -f $(COMPOSE_FILE) up -d

down: ## 停止并移除所有容器
	docker compose -f $(COMPOSE_FILE) down

build: ## 重新构建镜像并启动
	docker compose -f $(COMPOSE_FILE) up -d --build

logs: ## 查看所有服务日志
	docker compose -f $(COMPOSE_FILE) logs -f

mysql: ## 进入 MySQL 容器
	docker compose -f $(COMPOSE_FILE) exec mysql mysql -u production_pms -p

server: ## 进入 Server 容器
	docker compose -f $(COMPOSE_FILE) exec server sh

nginx: ## 进入 Nginx 容器
	docker compose -f $(COMPOSE_FILE) exec nginx sh

restart: ## 重启所有服务
	docker compose -f $(COMPOSE_FILE) restart

ps: ## 查看服务状态
	docker compose -f $(COMPOSE_FILE) ps

help: ## 显示帮助信息
	@grep -E '^[a-zA-Z_/]+:.*?## ' $(MAKEFILE_LIST) | sort | \
		awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-20s\033[0m %s\n", $$1, $$2}'
