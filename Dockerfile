# Build stage
FROM golang:1.21-alpine AS builder

# 安装必要的工具
RUN apk add --no-cache git ca-certificates tzdata

# 设置工作目录
WORKDIR /app

# 复制go mod文件
COPY go.mod go.sum ./

# 下载依赖
RUN go mod download

# 复制源代码
COPY . .

# 构建参数
ARG VERSION=dev
ARG COMMIT=unknown
ARG BUILD_DATE=unknown

# 构建应用
RUN CGO_ENABLED=0 GOOS=linux go build \
    -ldflags "-s -w -X github.com/harpoon/hpn/internal/version.Version=${VERSION} -X github.com/harpoon/hpn/internal/version.GitCommit=${COMMIT} -X github.com/harpoon/hpn/internal/version.BuildDate=${BUILD_DATE}" \
    -o hpn ./cmd/hpn

# Runtime stage
FROM alpine:3.18

# 安装运行时依赖
RUN apk add --no-cache \
    ca-certificates \
    tzdata \
    curl \
    docker-cli \
    && rm -rf /var/cache/apk/*

# 创建非root用户
RUN addgroup -g 1001 -S harpoon && \
    adduser -u 1001 -S harpoon -G harpoon

# 设置工作目录
WORKDIR /app

# 从构建阶段复制二进制文件
COPY --from=builder /app/hpn /usr/local/bin/hpn

# 复制配置文件模板
COPY config.yaml.example /app/config.yaml.example

# 创建必要的目录
RUN mkdir -p /app/data /app/logs && \
    chown -R harpoon:harpoon /app

# 切换到非root用户
USER harpoon

# 健康检查
HEALTHCHECK --interval=30s --timeout=10s --start-period=5s --retries=3 \
    CMD curl -f http://localhost:8080/health || exit 1

# 暴露端口
EXPOSE 8080

# 设置环境变量
ENV HPN_CONFIG_FILE=/app/config.yaml
ENV HPN_LOG_LEVEL=info
ENV HPN_HTTP_PORT=8080

# 启动命令
ENTRYPOINT ["hpn"]
CMD ["server", "--config", "/app/config.yaml"]