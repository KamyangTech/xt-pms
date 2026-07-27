#!/bin/sh
# =============================================================================
# Xt-PMS Docker 一键部署
# 使用方式：curl -fsSL https://gitee.com/kamyang-tech/xt-pms/raw/develop/deploy/docker-deploy.sh | sh
# =============================================================================
set -e

# ---- 颜色输出 ----
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
CYAN='\033[0;36m'
NC='\033[0m' # No Color

info()    { printf "${GREEN}[INFO]${NC} %s\n" "$*"; }
warn()    { printf "${YELLOW}[WARN]${NC} %s\n" "$*"; }
error()   { printf "${RED}[ERROR]${NC} %s\n" "$*"; }

# ---- 检测系统发行版 ----
detect_os() {
    if [ -f /etc/os-release ]; then
        . /etc/os-release
        OS=$ID
        OS_VERSION=$VERSION_ID
    elif [ -f /etc/redhat-release ]; then
        OS="rhel"
    else
        OS="unknown"
    fi
}

# ---- 检查并安装 Docker ----
ensure_docker() {
    if command -v docker >/dev/null 2>&1 && docker info >/dev/null 2>&1; then
        info "Docker 已安装且正在运行，跳过安装"
        return 0
    fi

    # Docker 已安装但未启动
    if command -v docker >/dev/null 2>&1; then
        warn "Docker 已安装但未运行，尝试启动..."
        if command -v systemctl >/dev/null 2>&1; then
            sudo systemctl start docker 2>/dev/null || sudo service docker start 2>/dev/null
        else
            sudo service docker start 2>/dev/null
        fi
        sleep 2
        if docker info >/dev/null 2>&1; then
            info "Docker 已启动"
            return 0
        fi
        error "Docker 启动失败，请手动检查"
        exit 1
    fi

    info "Docker 未安装，开始自动安装..."

    case "$OS" in
        ubuntu|debian)
            info "检测到 Debian/Ubuntu 系统"
            sudo apt-get update -qq
            sudo apt-get install -y -qq ca-certificates curl
            sudo install -m 0755 -d /etc/apt/keyrings
            sudo curl -fsSL https://mirrors.tuna.tsinghua.edu.cn/docker-ce/linux/$OS/gpg -o /etc/apt/keyrings/docker.asc 2>/dev/null || \
                sudo curl -fsSL https://download.docker.com/linux/$OS/gpg -o /etc/apt/keyrings/docker.asc
            sudo chmod a+r /etc/apt/keyrings/docker.asc
            echo "deb [arch=$(dpkg --print-architecture) signed-by=/etc/apt/keyrings/docker.asc] https://mirrors.tuna.tsinghua.edu.cn/docker-ce/linux/$OS $(. /etc/os-release && echo "$VERSION_CODENAME") stable" | \
                sudo tee /etc/apt/sources.list.d/docker.list > /dev/null
            sudo apt-get update -qq
            sudo apt-get install -y -qq docker-ce docker-ce-cli containerd.io docker-buildx-plugin docker-compose-plugin
            ;;
        centos|rhel|fedora|rocky|almalinux)
            info "检测到 RHEL/CentOS 系列系统"
            if [ "$OS" = "fedora" ]; then
                sudo dnf -y install dnf-plugins-core
                sudo dnf config-manager --add-repo https://mirrors.tuna.tsinghua.edu.cn/docker-ce/linux/fedora/docker-ce.repo 2>/dev/null || \
                    sudo dnf config-manager --add-repo https://download.docker.com/linux/fedora/docker-ce.repo
                sudo dnf install -y docker-ce docker-ce-cli containerd.io docker-buildx-plugin docker-compose-plugin
            else
                sudo yum install -y yum-utils
                sudo yum-config-manager --add-repo https://mirrors.tuna.tsinghua.edu.cn/docker-ce/linux/centos/docker-ce.repo 2>/dev/null || \
                    sudo yum-config-manager --add-repo https://download.docker.com/linux/centos/docker-ce.repo
                sudo yum install -y docker-ce docker-ce-cli containerd.io docker-buildx-plugin docker-compose-plugin
            fi
            ;;
        *)
            error "不支持的系统：$OS，请手动安装 Docker 后重试"
            error "参考：https://docs.docker.com/engine/install/"
            exit 1
            ;;
    esac

    # 启动 Docker
    info "启动 Docker 服务..."
    if command -v systemctl >/dev/null 2>&1; then
        sudo systemctl start docker
        sudo systemctl enable docker
    else
        sudo service docker start
    fi

    # 将当前用户添加到 docker 组（避免每次 sudo）
    if ! groups | grep -q docker; then
        sudo usermod -aG docker "$(whoami)" 2>/dev/null || true
        warn "已将当前用户添加到 docker 组，部分系统需要重新登录才能生效"
    fi

    if docker info >/dev/null 2>&1; then
        info "Docker 安装成功"
    else
        error "Docker 安装或启动失败，请手动检查"
        exit 1
    fi
}

# ---- 检测 docker compose 命令 ----
detect_compose_cmd() {
    # 优先使用 docker compose（插件模式）
    if docker compose version >/dev/null 2>&1; then
        COMPOSE_CMD="docker compose"
        return 0
    fi
    # 兼容独立 docker-compose
    if command -v docker-compose >/dev/null 2>&1; then
        COMPOSE_CMD="docker-compose"
        return 0
    fi
    warn "docker compose 未安装，尝试安装 docker-compose-plugin..."
    case "$OS" in
        ubuntu|debian)
            sudo apt-get install -y -qq docker-compose-plugin 2>/dev/null || true
            ;;
        centos|rhel|fedora|rocky|almalinux)
            sudo yum install -y docker-compose-plugin 2>/dev/null || \
                sudo dnf install -y docker-compose-plugin 2>/dev/null || true
            ;;
    esac
    if docker compose version >/dev/null 2>&1; then
        COMPOSE_CMD="docker compose"
    else
        error "docker compose 安装失败，请手动安装"
        exit 1
    fi
}

# ---- 检查端口是否被占用 ----
check_ports() {
    PORT=${SERVER_PORT:-9898}
    MYSQL_PORT=${MYSQL_PORT:-}
    HAS_CONFLICT=0

    for p in $PORT ${MYSQL_PORT:+$MYSQL_PORT}; do
        if ss -tlnp 2>/dev/null | grep -q ":$p " || netstat -tlnp 2>/dev/null | grep -q ":$p "; then
            warn "端口 $p 已被占用，请修改 .env 中的端口配置"
            HAS_CONFLICT=1
        fi
    done
    if [ "$HAS_CONFLICT" -eq 1 ]; then
        error "端口冲突，请调整后重试"
        exit 1
    fi
}

# =============================================================================
# 主流程
# =============================================================================

BASE_URL="https://gitee.com/kamyang-tech/xt-pms/raw/develop"

echo ""
printf "${CYAN}╔══════════════════════════════════════════════════════╗${NC}\n"
printf "${CYAN}║${NC}        Xt-PMS Docker 一键部署                     ${CYAN}║${NC}\n"
printf "${CYAN}╚══════════════════════════════════════════════════════╝${NC}\n"
echo ""

# ---- Step 0: 环境检测 ----
detect_os
info "检测到系统：$OS"

ensure_docker
detect_compose_cmd

echo ""

# ---- Step 1: 下载必要文件 ----
echo "[1/4] 下载 docker-compose.yml..."
curl -fsSL "$BASE_URL/deploy/docker-compose.yml" -O

echo "[2/4] 下载数据库初始化脚本..."
curl -fsSL "$BASE_URL/server/migrations/init_schema.sql" -O
curl -fsSL "$BASE_URL/server/migrations/init_seed.sql" -O

echo "[3/4] 生成 .env（如需外连 MySQL 请编辑此文件）..."
if [ ! -f .env ]; then
    curl -fsSL "$BASE_URL/deploy/.env.example" -o .env
fi

check_ports

echo "[4/4] 拉取镜像并启动..."
$COMPOSE_CMD up -d

echo ""
printf "${GREEN}╔══════════════════════════════════════════════════════╗${NC}\n"
printf "${GREEN}║${NC}  部署完成！                                       ${GREEN}║${NC}\n"
printf "${GREEN}║${NC}  地址：http://localhost:${SERVER_PORT:-9898}                      ${GREEN}║${NC}\n"
printf "${GREEN}║${NC}  登录：admin / admin123                          ${GREEN}║${NC}\n"
printf "${GREEN}╚══════════════════════════════════════════════════════╝${NC}\n"
echo ""
echo "升级命令：$COMPOSE_CMD pull && $COMPOSE_CMD up -d"
