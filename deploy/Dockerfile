# =============================================================================
# Xt-PMS 多阶段构建 Dockerfile
# =============================================================================

# ---- 通用构建参数（可在 `docker build --build-arg` 覆盖） ----
ARG APP_VERSION=dev
ARG BUILD_TIME=unknown
ARG GIT_COMMIT=unknown

# ---- Stage 1: 构建前端 ----
FROM node:18-alpine AS frontend-builder

# 允许在构建时覆盖 npm 镜像源
ARG NPM_REGISTRY=https://registry.npmmirror.com
# 把后端传入的版本号透传给前端构建（vite.config.ts 读取 VITE_APP_VERSION）
ARG APP_VERSION
ARG BUILD_TIME
ARG GIT_COMMIT

WORKDIR /build
COPY web/package.json web/package-lock.json ./
RUN npm config set registry ${NPM_REGISTRY} && \
    npm ci

COPY web/ .
# 注入版本号，让前端 __APP_VERSION__ 渲染正确的版本
RUN VITE_APP_VERSION=${APP_VERSION} \
    VITE_BUILD_TIME=${BUILD_TIME} \
    VITE_GIT_COMMIT=${GIT_COMMIT} \
    npm run build

# ---- Stage 2: 构建后端 ----
FROM golang:1.26-alpine AS backend-builder

# 允许在构建时覆盖 Go 模块代理（默认走国内镜像，失败自动回退到其它源）
ARG GOPROXY=https://goproxy.cn,https://goproxy.io,direct
ARG APP_VERSION
ARG BUILD_TIME
ARG GIT_COMMIT

# 切换到国内镜像源（默认 dl-cdn.alpinelinux.org 在国内网络下经常超时/I/O 错误）
RUN sed -i 's|dl-cdn.alpinelinux.org|mirrors.tuna.tsinghua.edu.cn|g' /etc/apk/repositories && \
    apk add --no-cache gcc musl-dev git

WORKDIR /build

# 先复制依赖文件，利用 Docker 缓存层
COPY server/go.mod server/go.sum ./
RUN go env -w GOPROXY=${GOPROXY} GOSUMDB=off && \
    go mod download

# 复制后端源码
COPY server/ .

# 从 Stage 1 复制前端构建产物到 cmd/web-dist（供 -tags release embed 使用）
COPY --from=frontend-builder /build/dist ./cmd/web-dist

# 编译（含前端嵌入），通过 -X 注入版本号 / 构建时间 / Git 提交到 main 包变量
RUN CGO_ENABLED=0 GOOS=linux go build -tags release \
    -ldflags="-s -w \
        -X 'main.Version=${APP_VERSION}' \
        -X 'main.BuildTime=${BUILD_TIME}' \
        -X 'main.GitCommit=${GIT_COMMIT}'" \
    -o bin/pms-server ./cmd

# ---- Stage 3: 运行时镜像 ----
# 后端自带前端静态文件服务，无需额外 Nginx 即可独立运行。
FROM alpine:3.20

LABEL maintainer="KamyangTech <pms@xintest.cn>"
LABEL description="Xt-PMS - 小微模具制造与注塑企业生产管理系统"

# 安装运行时依赖（时区数据、ca-certificates、健康检查工具、mysql 客户端）
# 切换到国内镜像源（默认 dl-cdn.alpinelinux.org 在国内网络下经常超时/I/O 错误）
RUN sed -i 's|dl-cdn.alpinelinux.org|mirrors.tuna.tsinghua.edu.cn|g' /etc/apk/repositories && \
    apk add --no-cache tzdata ca-certificates wget mysql-client && \
    cp /usr/share/zoneinfo/Asia/Shanghai /etc/localtime && \
    echo "Asia/Shanghai" > /etc/timezone && \
    rm -rf /var/cache/apk/*

WORKDIR /app

# 从构建阶段复制产物
COPY --from=backend-builder /build/bin/pms-server .
COPY --from=backend-builder /build/migrations ./migrations
COPY --from=backend-builder /build/system-assets ./system-assets

# 入口脚本
COPY deploy/docker-entrypoint.sh /docker-entrypoint.sh
RUN chmod +x /docker-entrypoint.sh

# 创建运行时目录
# 注意：容器以 root 运行（因为挂载的 volume 默认由 root 创建，
# 需要 root 权限才能在容器启动时 chown /app/{config,uploads,logs}）。
RUN mkdir -p /app/uploads /app/logs /app/config

EXPOSE 9898

# 健康检查
HEALTHCHECK --interval=30s --timeout=3s --start-period=10s --retries=3 \
    CMD wget --no-verbose --tries=1 --spider http://localhost:9898/api/ping || exit 1

# 入口脚本处理首次配置，然后启动服务
ENTRYPOINT ["/docker-entrypoint.sh"]
CMD ["./pms-server"]
