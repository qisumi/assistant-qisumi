# 多阶段构建 - 后端
# Stage 1: 构建阶段
FROM golang:1.24.5-alpine AS builder

# 安装必要的工具
RUN apk add --no-cache git ca-certificates tzdata

# 设置工作目录
WORKDIR /build

# 复制 go.mod 和 go.sum（利用 Docker 缓存）
COPY go.mod go.sum ./

# 下载依赖
RUN go mod download

# 复制源代码
COPY . .

# 构建应用
RUN CGO_ENABLED=1 GOOS=linux go build -a -installsuffix cgo -ldflags="-w -s" -o assistant-qisumi ./cmd/server

# Stage 2: 运行阶段
FROM alpine:3.19

# 安装运行时依赖
RUN apk --no-cache add ca-certificates tzdata sqlite-libs

# 创建非 root 用户
RUN addgroup -g 1000 qisumi && \
    adduser -D -u 1000 -G qisumi qisumi

# 设置工作目录
WORKDIR /app

# 从构建阶段复制二进制文件
COPY --from=builder /build/assistant-qisumi .

# 创建数据目录
RUN mkdir -p /app/data && chown -R qisumi:qisumi /app

# 切换到非 root 用户
USER qisumi

# 暴露端口
EXPOSE 4569

# 健康检查
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
    CMD wget --no-verbose --tries=1 --spider http://localhost:4569/health || exit 1

# 启动应用
CMD ["./assistant-qisumi"]
