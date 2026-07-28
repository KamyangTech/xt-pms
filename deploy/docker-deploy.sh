#!/bin/sh
# =============================================================================
# Xt-PMS Docker 一键部署
# 使用方式：curl -fsSL https://gitee.com/kamyang-tech/xt-pms/raw/main/deploy/docker-deploy.sh | sh
#
# 自定义端口：
#   1) 交互模式：本地运行脚本时会询问，直接回车使用默认值 9898
#   2) 环境变量：SERVER_PORT=8888 curl ... | sh
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

DEFAULT_SERVER_PORT=9898

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

# ---- 检测端口是否被占用 ----
is_port_in_use() {
    _p=$1
    if ss -tlnp 2>/dev/null | grep -qE "[.:]$_p[[:space:]]"; then
        return 0
    fi
    if netstat -tlnp 2>/dev/null | grep -qE "[.:]$_p[[:space:]]"; then
        return 0
    fi
    if command -v lsof >/dev/null 2>&1 && lsof -i ":$_p" >/dev/null 2>&1; then
        return 0
    fi
    return 1
}

# ---- 校验端口合法性 ----
is_valid_port() {
    _p=$1
    case "$_p" in
        ''|*[!0-9]*) return 1 ;;
    esac
    if [ "$_p" -lt 1 ] || [ "$_p" -gt 65535 ]; then
        return 1
    fi
    return 0
}

# ---- 检测是否能从 /dev/tty 读取用户输入（兼容 curl | sh）----
# 返回 0 表示可读，1 表示不可读
can_read_tty() {
    # 优先尝试打开 /dev/tty
    if [ -r /dev/tty ] 2>/dev/null; then
        return 0
    fi
    # 否则检查 stdin 是否为终端
    if [ -t 0 ]; then
        return 0
    fi
    return 1
}

# ---- 交互式询问端口 ----
# 用法：prompt_port "提示文字" "默认值" "回填到 .env 的键名（如 SERVER_PORT）" -> 输出到 stdout
prompt_port() {
    _msg=$1
    _default=$2
    _key=$3
    _input_port=""

    while :; do
        printf "%s [%s]: " "$_msg" "$_default" 1>&2
        if [ -r /dev/tty ] 2>/dev/null; then
            if ! read -r _input_port </dev/tty; then
                _input_port=""
            fi
        else
            if ! read -r _input_port; then
                _input_port=""
            fi
        fi

        # 用户直接回车：使用默认值
        if [ -z "$_input_port" ]; then
            _input_port=$_default
        fi

        if ! is_valid_port "$_input_port"; then
            warn "无效端口（必须是 1-65535 之间的数字）"
            continue
        fi

        if [ "$_input_port" -lt 1024 ]; then
            warn "建议使用 1024 以上的非特权端口"
            printf "${YELLOW}      是否仍要使用 %s？[y/N]: ${NC}" "$_input_port" 1>&2
            if [ -r /dev/tty ] 2>/dev/null; then
                read -r _yn </dev/tty || _yn="n"
            else
                read -r _yn || _yn="n"
            fi
            case "$_yn" in
                y|Y|yes|YES) ;;
                *) continue ;;
            esac
        fi

        if is_port_in_use "$_input_port"; then
            warn "端口 $_input_port 已被占用，请换一个端口"
            continue
        fi

        # 端口合法可用，写回 .env
        if [ -n "$_key" ] && [ -f .env ]; then
            # 若键已存在则替换；否则追加
            if grep -qE "^${_key}=" .env; then
                # 使用 | 作为 sed 分隔符，避免与端口中的特殊字符冲突
                sed -i.bak "s|^${_key}=.*|${_key}=${_input_port}|" .env
                rm -f .env.bak
            else
                printf "%s=%s\n" "$_key" "$_input_port" >> .env
            fi
        fi

        printf "%s" "$_input_port"
        return 0
    done
}

# ---- 让用户确认/选择端口（仅在交互模式下触发）----
configure_ports_interactive() {
    # 已在 .env 中存在的配置优先
    if [ -f .env ]; then
        # shellcheck disable=SC1091
        set -a
        . ./.env
        set +a
    fi

    SERVER_PORT=${SERVER_PORT:-$DEFAULT_SERVER_PORT}

    if can_read_tty; then
        echo ""
        info "请确认服务端口配置（直接回车使用默认值）"
        echo ""
        SERVER_PORT=$(prompt_port "服务端口" "$SERVER_PORT" "SERVER_PORT")
    fi

    # 最终落库的环境变量
    export SERVER_PORT
    info "当前配置：SERVER_PORT=$SERVER_PORT"
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

# ---- 端口配置（交互式 + 环境变量回退）----
configure_ports_interactive

# ---- 最终端口校验（非交互模式兜底）----
if is_port_in_use "$SERVER_PORT"; then
    error "端口 $SERVER_PORT 已被占用"
    if can_read_tty; then
        error "已无可用交互端口，请检查系统环境"
    else
        error "当前为非交互模式，请通过环境变量指定其他端口，例如："
        error "  curl ... | SERVER_PORT=8888 sh"
    fi
    exit 1
fi

echo "[4/4] 拉取镜像并启动..."
$COMPOSE_CMD up -d

echo ""
printf "${GREEN}╔══════════════════════════════════════════════════════╗${NC}\n"
printf "${GREEN}║${NC}  部署完成！                                       ${GREEN}║${NC}\n"
printf "${GREEN}║${NC}  地址：http://localhost:${SERVER_PORT}                      ${GREEN}║${NC}\n"
printf "${GREEN}║${NC}  登录：admin / admin123                          ${GREEN}║${NC}\n"
printf "${GREEN}╚══════════════════════════════════════════════════════╝${NC}\n"
echo ""
echo "升级命令：$COMPOSE_CMD pull && $COMPOSE_CMD up -d"