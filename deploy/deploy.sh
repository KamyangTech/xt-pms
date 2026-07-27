#!/usr/bin/env bash
# =============================================================================
# Xt-PMS 一键部署脚本
# =============================================================================
# 用法：
#   1. 上传 xt-pms-<version>-linux-amd64.tar.gz 到服务器
#   2. tar -xzf xt-pms-*.tar.gz
#   3. cd xt-pms-*
#   4. sudo bash deploy/deploy.sh
#
# 或一行命令：
#   curl -fsSL https://gitee.com/kamyang-tech/xt-pms/raw/main/deploy/deploy.sh | sudo bash
# =============================================================================

set -euo pipefail

# 禁用 dpkg 交互式弹窗（needrestart / kernel restart 提示等）
export DEBIAN_FRONTEND=noninteractive
export NEEDRESTART_MODE=a

# -------------------- 颜色输出 --------------------
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
CYAN='\033[0;36m'
BOLD='\033[1m'
NC='\033[0m'

log_info()  { printf "${GREEN}[INFO]${NC}  %s\n" "$*"; }
log_warn()  { printf "${YELLOW}[WARN]${NC}  %s\n" "$*"; }
log_error() { printf "${RED}[ERROR]${NC} %s\n" "$*"; }
log_step()  { printf "\n${CYAN}${BOLD}▶ %s${NC}\n" "$*"; }

# -------------------- 权限检查 --------------------
check_root() {
    if [ "$(id -u)" -ne 0 ]; then
        log_error "请使用 root 或 sudo 运行此脚本"
        echo "  sudo bash deploy/deploy.sh"
        exit 1
    fi
}

# -------------------- OS 检测 --------------------
APP_NAME="xt-pms"
INSTALL_DIR="${XT_PMS_HOME:-$PWD}"
SERVER_PORT="${SERVER_PORT:-9898}"
NGINX_ENABLED="${NGINX_ENABLED:-false}"

detect_os() {
    if [ -f /etc/os-release ]; then
        . /etc/os-release
        OS=$ID
        OS_VERSION=$VERSION_ID
    elif [ -f /etc/redhat-release ]; then
        OS="centos"
    else
        OS="unknown"
    fi
    log_info "检测到系统：${OS} ${OS_VERSION:-}"
}

# -------------------- 确定服务运行用户 --------------------
# www-data 为 Debian/Ubuntu 专有；CentOS/RHEL 等不存在该用户，
# 这里按系统自动选择/创建运行用户，避免 systemd 启动失败。
setup_service_user() {
    log_step "1.5/6  确定服务运行用户"

    if id -u www-data &>/dev/null; then
        SERVICE_USER="www-data"
        SERVICE_GROUP="www-data"
    elif id -u nginx &>/dev/null; then
        SERVICE_USER="nginx"
        SERVICE_GROUP="nginx"
    else
        SERVICE_USER="xtpms"
        SERVICE_GROUP="xtpms"
    fi

    # 若用户不存在则创建系统用户
    if ! id -u "${SERVICE_USER}" &>/dev/null; then
        log_info "创建服务用户 ${SERVICE_USER}..."
        if command -v useradd &>/dev/null; then
            useradd -r -s /usr/sbin/nologin "${SERVICE_USER}" 2>/dev/null \
                || useradd -r -s /sbin/nologin "${SERVICE_USER}" 2>/dev/null \
                || log_warn "创建用户失败，请手动确认 ${SERVICE_USER} 是否存在"
        else
            log_warn "未找到 useradd，请手动创建用户 ${SERVICE_USER}"
        fi
    fi

    log_info "服务将以 ${SERVICE_USER}:${SERVICE_GROUP} 运行"
}

# -------------------- 检查文件 --------------------
check_files() {
    log_step "1/6  检查安装包"

    if [ ! -f "./${APP_NAME}" ]; then
        log_error "找不到 ${APP_NAME} 二进制文件，请确保在解压后的目录中运行此脚本"
        exit 1
    fi

    if [ ! -d "./web" ]; then
        log_error "找不到 web 目录，安装包可能不完整"
        exit 1
    fi

    log_info "安装包文件检查通过"
}

# -------------------- 安装依赖 --------------------
install_deps() {
    log_step "2/6  检查系统依赖"

    # Go 二进制自带 HTTP 服务，数据库由安装向导页面配置，无需额外依赖
    # 仅当用户显式启用 Nginx 时才安装
    if [ "$NGINX_ENABLED" != "true" ]; then
        log_info "无需额外系统依赖，跳过"
        return
    fi

    if ! command -v nginx &>/dev/null; then
        case $OS in
            ubuntu|debian)
                log_info "安装 Nginx..."
                apt-get update -qq && apt-get install -y -qq nginx
                ;;
            centos|rhel|rocky|alma)
                log_info "安装 Nginx..."
                yum install -y epel-release && yum install -y nginx
                ;;
            *)
                log_warn "未知系统，请手动安装 nginx"
                ;;
        esac
    else
        log_info "Nginx 已安装，跳过"
    fi
}

# -------------------- 拷贝文件 --------------------
copy_files() {
    log_step "3/6  部署文件到 ${INSTALL_DIR}"

    # 如果已存在旧版本，先备份
    if [ -f "${INSTALL_DIR}/${APP_NAME}" ]; then
        log_info "检测到已有安装，备份到 ${INSTALL_DIR}.bak"
        rm -rf "${INSTALL_DIR}.bak"
        cp -a "${INSTALL_DIR}" "${INSTALL_DIR}.bak"
    fi

    mkdir -p "${INSTALL_DIR}/config"
    mkdir -p "${INSTALL_DIR}/logs"
    mkdir -p "${INSTALL_DIR}/uploads"

    # 拷贝二进制
    cp -f "./${APP_NAME}" "${INSTALL_DIR}/${APP_NAME}"
    chmod +x "${INSTALL_DIR}/${APP_NAME}"

    # 拷贝前端
    rm -rf "${INSTALL_DIR}/web"
    cp -r ./web "${INSTALL_DIR}/web"

    # 拷贝迁移脚本
    [ -d ./migrations ] && cp -r ./migrations "${INSTALL_DIR}/migrations"

    # 拷贝系统资源
    [ -d ./system-assets ] && cp -r ./system-assets "${INSTALL_DIR}/system-assets"

    # 版本文件
    [ -f ./VERSION ] && cp -f ./VERSION "${INSTALL_DIR}/VERSION"

    # 将安装目录授权给服务运行用户（否则 systemd 以非 root 用户启动后
    # 安装向导无法写入 config/logs/uploads 等目录）
    chown -R "${SERVICE_USER}:${SERVICE_GROUP}" "${INSTALL_DIR}"

    log_info "文件部署完成（属主 ${SERVICE_USER}:${SERVICE_GROUP}）"
}

# -------------------- 数据库 --------------------
setup_database() {
    log_step "4/6  数据库准备"

    echo ""
    log_info "数据库将通过安装向导页面进行配置，无需手动操作"
    log_info "请提前确保 MySQL 服务已启动并可连接"
    echo ""
    log_info "安装向导将引导你完成："
    echo "  ① 数据库连接 → ② 安全配置 → ③ 创建管理员 → ④ 确认安装"
    echo ""
}

# -------------------- systemd 服务 --------------------
setup_systemd() {
    log_step "5/6  注册系统服务"

    local service_file="/etc/systemd/system/${APP_NAME}.service"

    # 优先使用包内的 service 文件
    if [ -f "./deploy/xt-pms.service" ]; then
        cp -f ./deploy/xt-pms.service "${service_file}"
    else
        cat > "${service_file}" << EOF
[Unit]
Description=Xt-PMS 生产管理系统
After=network.target mysql.service

[Service]
Type=simple
User=root
Group=root
WorkingDirectory=${INSTALL_DIR}
ExecStart=${INSTALL_DIR}/${APP_NAME}
ExecReload=/bin/kill -HUP \$MAINPID
Restart=on-failure
RestartSec=5

NoNewPrivileges=yes
PrivateTmp=yes
ProtectSystem=strict
ProtectHome=yes
ReadWritePaths=${INSTALL_DIR}/config ${INSTALL_DIR}/uploads ${INSTALL_DIR}/logs ${INSTALL_DIR}/system-assets

Environment=TZ=Asia/Shanghai

[Install]
WantedBy=multi-user.target
EOF
    fi

    # 统一 systemd 运行用户（兼容 Debian www-data / CentOS 无 www-data 的场景）
    sed -i "s/^User=.*/User=${SERVICE_USER}/; s/^Group=.*/Group=${SERVICE_GROUP}/" "${service_file}"

    # 如果 Nginx 启用且包内有 nginx 配置，部署之
    if [ "$NGINX_ENABLED" = "true" ] && [ -f "./deploy/nginx.conf" ]; then
        if [ -d "/etc/nginx/sites-available" ]; then
            cp -f ./deploy/nginx.conf /etc/nginx/sites-available/${APP_NAME}
            ln -sf /etc/nginx/sites-available/${APP_NAME} /etc/nginx/sites-enabled/${APP_NAME}
        elif [ -d "/etc/nginx/conf.d" ]; then
            cp -f ./deploy/nginx.conf /etc/nginx/conf.d/${APP_NAME}.conf
        fi
        nginx -t 2>/dev/null && systemctl reload nginx && log_info "Nginx 配置已生效" || log_warn "Nginx 配置测试失败，请检查后手动重载"
    fi

    systemctl daemon-reload
    systemctl enable "${APP_NAME}" &>/dev/null
    systemctl start "${APP_NAME}"

    sleep 2

    if systemctl is-active --quiet "${APP_NAME}"; then
        log_info "${APP_NAME} 服务启动成功"
    else
        log_error "${APP_NAME} 服务启动失败，查看日志：journalctl -u ${APP_NAME} -n 50"
        exit 1
    fi
}

# -------------------- 安装向导 --------------------
show_installer_url() {
    log_step "6/6  完成安装"

    echo ""
    printf "${GREEN}${BOLD}╔══════════════════════════════════════════════════════════╗${NC}\n"
    printf "${GREEN}${BOLD}║${NC}  ${BOLD}部署完成！${NC}\n"
    printf "${GREEN}${BOLD}╚══════════════════════════════════════════════════════════╝${NC}\n"
    echo ""

    # 等待安装向导启动并获取 URL
    echo "正在等待安装向导启动..."
    for i in $(seq 1 10); do
        local token_line
        token_line=$(journalctl -u "${APP_NAME}" --no-pager -n 20 2>/dev/null | grep -oP 'http://[^\s]+\?token=[a-f0-9]+' | tail -1 || true)
        if [ -n "${token_line}" ]; then
            printf "\n  ${CYAN}${BOLD}👉 安装向导地址：${NC}"
            printf "${BOLD}%s${NC}\n\n" "$token_line"
            break
        fi
        sleep 2
    done

    if [ -z "${token_line:-}" ]; then
        log_warn "未能自动获取安装向导地址，请手动查看："
        echo "  journalctl -u ${APP_NAME} -f"
    fi

    echo "请在浏览器打开上方地址，完成 4 步引导："
    echo "  ① 数据库连接 → ② 安全配置 → ③ 创建管理员 → ④ 确认安装"
    echo ""
    if [ "$NGINX_ENABLED" = "true" ]; then
        local server_ip
        server_ip=$(hostname -I 2>/dev/null | awk '{print $1}')
        echo "安装完成后，访问 http://${server_ip:-<服务器IP>} 即可使用系统"
    else
        echo "安装完成后，访问 http://<服务器IP>:${SERVER_PORT} 即可使用系统"
    fi
    echo ""
}

# -------------------- 主流程 --------------------
main() {
    echo ""
    printf "${CYAN}${BOLD}╔══════════════════════════════════════════════════════╗${NC}\n"
    printf "${CYAN}${BOLD}║${NC}        ${BOLD}Xt-PMS 一键部署脚本${NC}                         ${CYAN}${BOLD}║${NC}\n"
    printf "${CYAN}${BOLD}╚══════════════════════════════════════════════════════╝${NC}\n"
    echo ""

    check_root
    detect_os
    setup_service_user
    check_files
    install_deps
    copy_files
    setup_database
    setup_systemd
    show_installer_url
}

main "$@"
