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

# -------------------- 主流程 --------------------
require_root

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
    SERVER_PORT="${SERVER_PORT:-9898}" \
    bash deploy.sh
