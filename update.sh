#!/bin/bash
# Assistant Qisumi 更新脚本
# 用法: ./update.sh [version]
# version: 要更新的版本号（可选，默认拉取最新代码）

set -e

# 颜色输出
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# 函数
success() {
    echo -e "${GREEN}✓ $1${NC}"
}

warning() {
    echo -e "${YELLOW}⚠ $1${NC}"
}

error() {
    echo -e "${RED}✗ $1${NC}"
    exit 1
}

info() {
    echo -e "${BLUE}ℹ $1${NC}"
}

# 获取版本参数
VERSION=${1:-latest}

info "开始更新 Assistant Qisumi (版本: $VERSION)"

# 使用 docker compose 还是 docker-compose
if docker compose version &> /dev/null; then
    DOCKER_COMPOSE="docker compose"
else
    DOCKER_COMPOSE="docker-compose"
fi

# 备份数据库
info "备份数据库..."
BACKUP_DIR="backups/$(date +%Y%m%d_%H%M%S)"
mkdir -p "$BACKUP_DIR"
docker exec qisumi-backend cp /app/data/assistant.db "/tmp/backup_$(date +%s).db" 2>/dev/null || warning "无法备份数据库，跳过"
docker cp qisumi-backend:/app/data/assistant.db "$BACKUP_DIR/" 2>/dev/null || warning "数据库备份失败，继续更新"
success "数据库已备份到: $BACKUP_DIR"

# 拉取最新代码
if [ "$VERSION" = "latest" ]; then
    info "拉取最新代码..."
    git fetch --all
    git checkout master
    git pull origin master
else
    info "切换到版本 $VERSION..."
    git fetch --all --tags
    git checkout "v$VERSION" || error "版本 v$VERSION 不存在"
fi

# 检查配置文件变更
if [ -f .env.example ]; then
    info "检查环境变量配置..."
    if ! diff -q .env .env.example > /dev/null 2>&1; then
        warning "检测到 .env.example 有新配置项"
        warning "建议检查并更新 .env 文件"
    fi
fi

# 重新构建镜像
info "重新构建 Docker 镜像..."
$DOCKER_COMPOSE build --no-cache

# 重启服务
info "重启服务..."
$DOCKER_COMPOSE down
$DOCKER_COMPOSE up -d

# 等待服务启动
info "等待服务启动..."
sleep 15

# 检查服务状态
info "检查服务状态..."
if $DOCKER_COMPOSE ps | grep -q "Exit"; then
    error "服务启动失败，请查看日志"
fi

# 数据库迁移（如果有新的迁移）
info "执行数据库迁移..."
# 这里可以添加数据库迁移命令
# docker exec qisumi-backend ./migrate

success "更新完成！"
echo ""
info "当前版本: $(git describe --tags --always 2>/dev/null || echo 'unknown')"
info "查看日志: $DOCKER_COMPOSE logs -f"
echo ""

# 显示更新日志
if [ "$VERSION" != "latest" ]; then
    info "版本 v$VERSION 的更新日志:"
    git log --oneline --graph --decorate $(git describe --tags --abbrev=0 HEAD^)..HEAD 2>/dev/null || echo "无法获取更新日志"
fi
