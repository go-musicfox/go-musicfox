# go-musicfox v2 主应用容器
# 多阶段构建，支持微内核插件架构

# 构建阶段
FROM golang:1.21-alpine AS builder

# 安装构建依赖
RUN apk add --no-cache \
    git \
    make \
    gcc \
    musl-dev \
    pkgconfig \
    alsa-lib-dev \
    portaudio-dev \
    flac-dev \
    taglib-dev

# 设置工作目录
WORKDIR /build

# 复制go模块文件
COPY go.mod go.sum ./

# 下载依赖
RUN go mod download

# 复制源代码
COPY . .

# 构建参数
ARG VERSION=dev
ARG COMMIT=unknown
ARG BUILD_TIME
ARG GOOS=linux
ARG GOARCH=amd64

# 设置构建环境
ENV CGO_ENABLED=1
ENV GOOS=${GOOS}
ENV GOARCH=${GOARCH}

# 构建微内核和插件
RUN make build-all VERSION=${VERSION} COMMIT=${COMMIT} BUILD_TIME="${BUILD_TIME}"

# 运行时阶段
FROM alpine:3.18

# 安装运行时依赖
RUN apk add --no-cache \
    ca-certificates \
    alsa-lib \
    portaudio \
    flac \
    taglib \
    tzdata \
    && rm -rf /var/cache/apk/*

# 创建非root用户
RUN addgroup -g 1000 musicfox && \
    adduser -D -s /bin/sh -u 1000 -G musicfox musicfox

# 设置工作目录
WORKDIR /app

# 创建必要目录
RUN mkdir -p \
    /app/bin \
    /app/plugins \
    /app/config \
    /app/data \
    /app/logs \
    && chown -R musicfox:musicfox /app

# 从构建阶段复制文件
COPY --from=builder --chown=musicfox:musicfox /build/build/go-musicfox-linux-amd64 /app/bin/go-musicfox
COPY --from=builder --chown=musicfox:musicfox /build/build/plugins/ /app/plugins/
COPY --from=builder --chown=musicfox:musicfox /build/configs/ /app/config/

# 复制启动脚本
COPY --chown=musicfox:musicfox docker/entrypoint.sh /app/entrypoint.sh
RUN chmod +x /app/entrypoint.sh /app/bin/go-musicfox

# 设置环境变量
ENV MUSICFOX_HOME=/app
ENV MUSICFOX_CONFIG_DIR=/app/config
ENV MUSICFOX_PLUGIN_DIR=/app/plugins
ENV MUSICFOX_DATA_DIR=/app/data
ENV MUSICFOX_LOG_DIR=/app/logs
ENV MUSICFOX_LOG_LEVEL=info

# 暴露端口（用于插件RPC通信）
EXPOSE 8080 8081 8082

# 健康检查
HEALTHCHECK --interval=30s --timeout=10s --start-period=5s --retries=3 \
    CMD /app/bin/go-musicfox --health-check || exit 1

# 数据卷
VOLUME ["/app/data", "/app/logs", "/app/config"]

# 切换到非root用户
USER musicfox

# 设置入口点
ENTRYPOINT ["/app/entrypoint.sh"]
CMD ["--daemon"]