#!/usr/bin/env bash
# =============================================================================
# Xt-PMS 快速安装脚本（自动获取最新发行版并部署）
# =============================================================================
# 用法：
#   普通用户（脚本内部会按需 sudo）：
#     bash <(curl -fsSL https://gitee.com/kamyang-tech/xt-pms/raw/main/deploy/quick-install.sh)
#
#   或下载后运行：
#     curl -fsSL <url> -o quick-install.sh
#     sudo bash quick-install.sh
#
# 可选环境变量：
#   XT_PMS_VERSION   指定版本标签（如 v0.1.0-beta.3），不填则自动取最新
#   GITEE_TOKEN      Gitee Access Token，用于提高 API 频率上限（可选）
#   NGINX_ENABLED    是否部署 Nginx 反向代理（true/false，默认 false）
#   SERVER_PORT      服务端口（默认 9898）
#   NGINX_PORT       Nginx 监听端口（默认 80，仅在 NGINX_ENABLED=true 时生效）
#
# 自定义端口：
#   1) 交互模式：脚本会询问，直接回车使用默认值
#   2) 环境变量：SERVER_PORT=8888 bash quick-install.sh
# =============================================================================

set -euo pipefail

# -------------------- 常量 --------------------
REPO="kamyang-tech/xt-pms"
APP_NAME="xt-pms"
API_BASE="https://gitee.com/api/v5/repos/${REPO}"
DL_BASE="https://gitee.com/${REPO}/releases/download"

# -------------------- 颜色输出 --------------------
RED='\033[0;31m'; GREEN='\033[0;32m'; YELLOW='\033[1;33m'
CYAN='\033[0;36m'; BOLD='\033[1m'; NC='\033[0m'
log_info()  { printf "${GREEN}[INFO]${NC}  %s\n" "$*" >&2; }
log_warn()  { printf "${YELLOW}[WARN]${NC}  %s\n" "$*" >&2; }
log_error() { printf "${RED}[ERROR]${NC} %s\n" "$*" >&2; }
log_step()  { printf "\n${CYAN}${BOLD}▶ %s${NC}\n" "$*" >&2; }

# -------------------- 权限检查 --------------------
require_root() {
    if [ "$(id -u)" -ne 0 ]; then
        log_error "请使用 root 或 sudo 运行此脚本"
        echo "  sudo bash quick-install.sh"
        exit 1
    fi
}

# -------------------- 获取最新版本 --------------------
fetch_latest_tag() {
    local url="${API_BASE}/releases/latest"
    [ -n "${GITEE_TOKEN:-}" ] && url="${url}?access_token=${GITEE_TOKEN}"

    log_info "查询最新发行版：${API_BASE}/releases/latest"
    local resp
    if ! resp=$(curl -fsSL "${url}" 2>/dev/null); then
        log_error "无法获取最新发行版信息（可能触发 Gitee API 频率限制）"
        log_info "请设置环境变量 GITEE_TOKEN，或指定 XT_PMS_VERSION 手动指定版本"
        exit 1
    fi

    local tag
    tag=$(printf '%s' "$resp" | grep -Po '"tag_name":\s*"\K[^"]+')
    [ -z "${tag:-}" ] && { log_error "解析最新版本 tag 失败"; exit 1; }
    printf '%s' "$tag"
}

# -------------------- 端口工具 --------------------
is_port_in_use() {
    local port=$1
    if ss -tlnp 2>/dev/null | grep -qE "[.:]${port}[[:space:]]"; then
        return 0
    fi
    if netstat -tlnp 2>/dev/null | grep -qE "[.:]${port}[[:space:]]"; then
        return 0
    fi
    if command -v lsof >/dev/null 2>&1 && lsof -i ":${port}" >/dev/null 2>&1; then
        return 0
    fi
    return 1
}

is_valid_port() {
    local port=$1
    case "$port" in
        ''|*[!0-9]*) return 1 ;;
    esac
    if [ "$port" -lt 1 ] || [ "$port" -gt 65535 ]; then
        return 1
    fi
    return 0
}

can_read_tty() {
    if [ -r /dev/tty ] 2>/dev/null; then
        return 0
    fi
    if [ -t 0 ]; then
        return 0
    fi
    return 1
}

prompt_port() {
    local msg=$1
    local default_port=$2
    local input_port=""

    while :; do
        printf "%s [%s]: " "$msg" "$default_port" >&2
        if [ -r /dev/tty ] 2>/dev/null; then
            read -r input_port </dev/tty || input_port=""
        else
            read -r input_port || input_port=""
        fi

        if [ -z "$input_port" ]; then
            input_port=$default_port
        fi

        if ! is_valid_port "$input_port"; then
            log_warn "无效端口（必须是 1-65535 之间的数字）" >&2
            continue
        fi

        if [ "$input_port" -lt 1024 ]; then
            log_warn "建议使用 1024 以上的非特权端口" >&2
            printf "      是否仍要使用 %s？[y/N]: " "$input_port" >&2
            local yn
            if [ -r /dev/tty ] 2>/dev/null; then
                read -r yn </dev/tty || yn="n"
            else
                read -r yn || yn="n"
            fi
            case "$yn" in
                y|Y|yes|YES) ;;
                *) continue ;;
            esac
        fi

        if is_port_in_use "$input_port"; then
            log_warn "端口 $input_port 已被占用，请换一个端口" >&2
            continue
        fi

        printf "%s" "$input_port"
        return 0
    done
}

configure_ports() {
    log_step "配置服务端口"

    : "${SERVER_PORT:=9898}"
    : "${NGINX_PORT:=80}"

    if can_read_tty; then
        echo "" >&2
        log_info "请确认服务端口配置（直接回车使用默认值）" >&2
        echo "" >&2
        SERVER_PORT=$(prompt_port "服务端口 (PMS_SERVER_PORT)" "$SERVER_PORT")
        if [ "${NGINX_ENABLED:-false}" = "true" ]; then
            NGINX_PORT=$(prompt_port "Nginx 监听端口 (NGINX_PORT)" "$NGINX_PORT")
        fi
    else
        log_info "当前为非交互模式，使用环境变量/默认值：SERVER_PORT=$SERVER_PORT" >&2
        if is_port_in_use "$SERVER_PORT"; then
            log_error "端口 $SERVER_PORT 已被占用，请通过环境变量 SERVER_PORT 指定其他端口" >&2
            log_error "  示例：SERVER_PORT=8888 bash quick-install.sh" >&2
            exit 1
        fi
    fi

    export SERVER_PORT NGINX_PORT
    log_info "端口配置完成：SERVER_PORT=$SERVER_PORT${NGINX_ENABLED:+ NGINX_PORT=$NGINX_PORT}" >&2
}

# -------------------- 主流程 --------------------
require_root
configure_ports

# 记录原始工作目录（进入临时目录前保存）
XT_PMS_HOME="${PWD}"

# 1. 确定版本
if [ -n "${XT_PMS_VERSION:-}" ]; then
    TAG="${XT_PMS_VERSION}"
    log_info "使用指定的版本：${TAG}"
else
    TAG="$(fetch_latest_tag)"
    log_info "检测到最新版本：${TAG}"
fi

VERSION="${TAG#v}"
TARBALL="${APP_NAME}-${VERSION}-linux-amd64.tar.gz"

# 2. 准备临时目录
TMP_DIR="$(mktemp -d)"
trap 'rm -rf "${TMP_DIR}"' EXIT
cd "${TMP_DIR}"

# 3. 下载
DL_URL="${DL_BASE}/${TAG}/${TARBALL}"
SHA_URL="${DL_BASE}/${TAG}/${APP_NAME}-${VERSION}-linux-amd64.sha256"

log_step "下载 ${TARBALL}"
if ! curl -fL#O "${DL_URL}"; then
    log_error "下载失败：${DL_URL}"
    exit 1
fi

# 4. 校验 sha256（若 release 附带校验文件）
if curl -fsSL -o "${APP_NAME}-${VERSION}-linux-amd64.sha256" "${SHA_URL}" 2>/dev/null; then
    log_info "校验 sha256 ..."
    if grep "${TARBALL}" "${APP_NAME}-${VERSION}-linux-amd64.sha256" | sha256sum -c -; then
        log_info "校验通过"
    else
        log_error "sha256 校验失败，发行包可能被篡改"
        exit 1
    fi
else
    log_warn "未找到 sha256 校验文件，跳过校验"
fi

# 5. 解压并部署
log_step "解压 ${TARBALL}"
tar -xzf "${TARBALL}"
cd "${APP_NAME}-${VERSION}-linux-amd64"

log_step "执行部署脚本"
# 透传常用环境变量给 deploy.sh
sudo -E env \
    XT_PMS_HOME="${XT_PMS_HOME}" \
    NGINX_ENABLED="${NGINX_ENABLED:-false}" \
    NGINX_PORT="${NGINX_PORT:-80}" \
    SERVER_PORT="${SERVER_PORT:-9898}" \
    bash deploy.sh
