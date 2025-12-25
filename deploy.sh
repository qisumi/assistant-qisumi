#!/bin/bash
# Assistant Qisumi 部署脚本
# 用法: ./deploy.sh [environment]
# environment: dev (默认) | prod

set -e

# 颜色输出
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# 函数：打印成功消息
success() {
    echo -e "${GREEN}✓ $1${NC}"
}

# 函数：打印警告消息
warning() {
    echo -e "${YELLOW}⚠ $1${NC}"
}

# 函数：打印错误消息
error() {
    echo -e "${RED}✗ $1${NC}"
    exit 1
}

# 函数：打印信息
info() {
    echo -e "ℹ $1"
}

# 获取环境参数
ENVIRONMENT=${1:-dev}

info "开始部署 Assistant Qisumi (环境: $ENVIRONMENT)"

# 检查 Docker 是否安装
if ! command -v docker &> /dev/null; then
    error "Docker 未安装，请先安装 Docker"
fi

# 检查 Docker Compose 是否安装
if ! command -v docker-compose &> /dev/null && ! docker compose version &> /dev/null; then
    error "Docker Compose 未安装，请先安装 Docker Compose"
fi

# 使用 docker compose 还是 docker-compose
if docker compose version &> /dev/null; then
    DOCKER_COMPOSE="docker compose"
else
    DOCKER_COMPOSE="docker-compose"
fi

# 检查 .env 文件是否存在
if [ ! -f .env ]; then
    warning ".env 文件不存在，从 .env.example 复制"
    if [ -f .env.example ]; then
        cp .env.example .env
        warning "请编辑 .env 文件配置必要的环境变量"
        read -p "按 Enter 继续（或 Ctrl+C 退出）..."
    else
        error ".env.example 文件不存在"
    fi
fi

# 拉取最新代码（如果是在 git 仓库中）
if [ -d .git ]; then
    info "拉取最新代码..."
    git pull origin master || warning "无法拉取代码，继续使用本地代码"
fi

# 停止现有容器
info "停止现有容器..."
$DOCKER_COMPOSE down 2>/dev/null || true

# 构建镜像
info "构建 Docker 镜像..."
if [ "$ENVIRONMENT" = "prod" ]; then
    $DOCKER_COMPOSE build --no-cache
else
    $DOCKER_COMPOSE build
fi

# 启动服务
info "启动服务..."
if [ "$ENVIRONMENT" = "prod" ]; then
    $DOCKER_COMPOSE up -d --profile production
else
    $DOCKER_COMPOSE up -d
fi

# 等待服务启动
info "等待服务启动..."
sleep 10

# 检查服务状态
info "检查服务状态..."
$DOCKER_COMPOSE ps

# 显示日志
echo ""
success "部署完成！"
echo ""
info "服务访问地址:"
if [ "$ENVIRONMENT" = "prod" ]; then
    echo "  前端: http://localhost"
    echo "  后端 API: http://localhost:4569"
    echo "  HTTPS: https://localhost (需要配置 SSL 证书)"
else
    echo "  前端: http://localhost"
    echo "  后端 API: http://localhost:4569"
fi
echo ""
info "查看日志: $DOCKER_COMPOSE logs -f"
info "停止服务: $DOCKER_COMPOSE down"
info "重启服务: $DOCKER_COMPOSE restart"
echo ""

# 显示初始日志
info "显示最近日志..."
$DOCKER_COMPOSE logs --tail=20
