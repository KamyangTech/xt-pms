#!/bin/sh
# =============================================================================
# Xt-PMS Docker 入口脚本
# 从环境变量自动生成配置文件，适用于容器化部署
# 与裸机部署（安装向导）完全独立，互不影响
# =============================================================================

set -e

APP_DIR="/app"
CONFIG_DIR="${APP_DIR}/config"
CONFIG_FILE="${CONFIG_DIR}/config.yaml"
ENV_FILE="${CONFIG_DIR}/.env"
INSTALLED_FILE="${CONFIG_DIR}/.installed"

# -------------------- 修复挂载 volume 目录权限 --------------------
# volume 挂载到 /app/{config,uploads,logs} 时，目录会被 root 拥有，
# 需在执行后续命令前将其调整为可写（运行用户固定为 root）。
mkdir -p "${CONFIG_DIR}" "${APP_DIR}/uploads" "${APP_DIR}/logs"
chown -R root:root "${APP_DIR}" 2>/dev/null || true
find "${APP_DIR}" -mindepth 1 -maxdepth 1 -exec chown -R root:root {} \; 2>/dev/null || true

# -------------------- 默认值 --------------------
SERVER_PORT="${PMS_SERVER_PORT:-9898}"
APP_MODE="${PMS_APP_MODE:-production}"
COOKIE_SECURE="${PMS_COOKIE_SECURE:-false}"
LOG_LEVEL="${PMS_LOG_LEVEL:-info}"

# -------------------- 仅在首次运行时生成配置 --------------------
generate_config() {
    echo "[docker-entrypoint] 检测到首次运行，从环境变量生成配置..."

    # 如果 JWT 密钥 / Pepper 为空或仍是示例值，自动随机生成
    if [ -z "${PMS_JWT_SECRET}" ] || [ "${PMS_JWT_SECRET}" = "your-jwt-secret-change-in-production" ]; then
        PMS_JWT_SECRET=$(tr -dc 'a-f0-9' < /dev/urandom | head -c 64)
        echo "[docker-entrypoint] 已自动生成 PMS_JWT_SECRET"
    fi
    if [ -z "${PMS_SECURITY_PEPPER}" ] || [ "${PMS_SECURITY_PEPPER}" = "your-pepper-change-in-production" ]; then
        PMS_SECURITY_PEPPER=$(tr -dc 'a-f0-9' < /dev/urandom | head -c 64)
        echo "[docker-entrypoint] 已自动生成 PMS_SECURITY_PEPPER"
    fi

    mkdir -p "${CONFIG_DIR}"

    # 生成 config.yaml
    cat > "${CONFIG_FILE}" << YAMLEOF
# Xt-PMS 配置文件（Docker 部署，由 entrypoint 自动生成）
server:
  port: ${SERVER_PORT}
  cookie_secure: ${COOKIE_SECURE}
database:
  host: ${PMS_DB_HOST:-mysql}
  port: ${PMS_DB_PORT:-3306}
  user: ${PMS_DB_USER:-production_pms}
  password: "${PMS_DB_PASSWORD:-}"
  dbname: ${PMS_DB_NAME:-production_pms}
jwt:
  secret: "${PMS_JWT_SECRET}"
  expire: 7200
security:
  pepper: "${PMS_SECURITY_PEPPER}"
app:
  mode: ${APP_MODE}
log:
  level: ${LOG_LEVEL}
  format: console
  output: both
  file:
    path: logs/app.log
    max_size: 100
    max_backups: 30
    max_age: 28
    compress: true
upload:
  base_path: uploads
  max_size: 10485760
  allowed_image_types: [".jpg", ".jpeg", ".png", ".gif", ".svg", ".ico", ".webp"]
  allowed_file_types: [".jpg", ".jpeg", ".png", ".gif", ".svg", ".ico", ".webp", ".pdf", ".doc", ".docx", ".xls", ".xlsx"]
  thumbnail_max_width: 200
  thumbnail_max_height: 200
cors:
  # 默认同源部署，无需配置跨域白名单
  # 如需前后端分离部署，添加前端域名，如: ["https://frontend.example.com"]
  allowed_origins: []
auth:
  super_admin_user_id: 1
  super_admin_role_type: 1
YAMLEOF

    # 生成 .env
    cat > "${ENV_FILE}" << ENVEOF
PMS_DB_HOST=${PMS_DB_HOST:-mysql}
PMS_DB_PORT=${PMS_DB_PORT:-3306}
PMS_DB_USER=${PMS_DB_USER:-production_pms}
PMS_DB_PASSWORD=${PMS_DB_PASSWORD:-}
PMS_DB_NAME=${PMS_DB_NAME:-production_pms}
PMS_JWT_SECRET=${PMS_JWT_SECRET}
PMS_SECURITY_PEPPER=${PMS_SECURITY_PEPPER}
PMS_APP_MODE=${APP_MODE}
PMS_SERVER_PORT=${SERVER_PORT}
PMS_COOKIE_SECURE=${COOKIE_SECURE}
ENVEOF
    chmod 600 "${ENV_FILE}"

    # 生成 .installed → 跳过 Web 安装向导
    INSTALLED_AT=$(date -u +"%Y-%m-%dT%H:%M:%SZ")
    cat > "${INSTALLED_FILE}" << JSONEOF
{
  "version": "docker",
  "build_time": "${INSTALLED_AT}",
  "git_commit": "docker",
  "go_version": "docker",
  "schema_version": 1,
  "installed_at": "${INSTALLED_AT}"
}
JSONEOF

    echo "[docker-entrypoint] 配置文件生成完成。"
}

# -------------------- 等待 MySQL 就绪 --------------------
wait_for_mysql() {
    DB_HOST="${PMS_DB_HOST:-mysql}"
    DB_PORT="${PMS_DB_PORT:-3306}"

    echo "[docker-entrypoint] 等待 MySQL (${DB_HOST}:${DB_PORT}) 就绪..."

    RETRIES=30
    while [ "${RETRIES}" -gt 0 ]; do
        if mysqladmin ping -h"${DB_HOST}" -P"${DB_PORT}" -u"${PMS_DB_USER:-production_pms}" -p"${PMS_DB_PASSWORD:-}" --silent 2>/dev/null; then
            echo "[docker-entrypoint] MySQL 已就绪。"
            return 0
        fi
        RETRIES=$((RETRIES - 1))
        echo "[docker-entrypoint] 等待中... (${RETRIES}/30)"
        sleep 2
    done

    echo "[docker-entrypoint] 警告：MySQL 连接超时，将继续启动。"
}

# -------------------- 创建管理员（可选） --------------------
create_admin() {
    if [ -z "${PMS_ADMIN_USERNAME:-}" ] || [ -z "${PMS_ADMIN_PASSWORD:-}" ]; then
        echo "[docker-entrypoint] 未设置 PMS_ADMIN_USERNAME/PMS_ADMIN_PASSWORD，跳过管理员创建"
        return
    fi

    echo "[docker-entrypoint] 创建管理员 ${PMS_ADMIN_USERNAME}..."

    RETRIES=10
    while [ "${RETRIES}" -gt 0 ]; do
        if ./pms-server setup-admin; then
            echo "[docker-entrypoint] 管理员创建完成"
            return 0
        fi
        RETRIES=$((RETRIES - 1))
        if [ "${RETRIES}" -gt 0 ]; then
            echo "[docker-entrypoint] 数据库尚未就绪，重试中... (${RETRIES}/10)"
            sleep 3
        fi
    done

    echo "[docker-entrypoint] 管理员创建失败，请手动执行："
    echo "  docker exec xt-pms-server ./pms-server setup-admin"
}

# -------------------- 主流程 --------------------
main() {
    echo "[docker-entrypoint] Xt-PMS Docker 容器启动中..."

    # 首次运行：从环境变量生成配置
    if [ ! -f "${INSTALLED_FILE}" ] || [ ! -f "${CONFIG_FILE}" ]; then
        generate_config
    fi

    # 等待数据库
    wait_for_mysql

    # 创建管理员（如果配置了环境变量）
    create_admin

    echo "[docker-entrypoint] 启动 Xt-PMS 服务..."
    exec "$@"
}

main "$@"
