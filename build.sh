#!/bin/bash
# Assistant Qisumi 构建脚本
# 构建前端和后端，并将前端静态文件集成到后端

set -e

# 颜色输出
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

success() {
    echo -e "${GREEN}✓ $1${NC}"
}

error() {
    echo -e "${RED}✗ $1${NC}"
    exit 1
}

info() {
    echo -e "${BLUE}ℹ $1${NC}"
}

echo "========================================="
echo "  Assistant Qisumi - 完整构建"
echo "========================================="
echo ""

# 清理旧的构建产物
info "清理旧的构建产物..."
rm -rf static
rm -f assistant-qisumi
success "清理完成"

# 构建前端
info "构建前端..."
cd frontend
npm install
npm run build
cd ..
success "前端构建完成"

# 创建 static 目录
info "准备静态文件目录..."
mkdir -p static

# 复制前端构建产物到 static 目录
info "复制前端静态文件..."
cp -r frontend/dist/* static/
success "静态文件复制完成"

# 构建后端
info "构建后端..."
CGO_ENABLED=1 go build -o assistant-qisumi ./cmd/server
success "后端构建完成"

echo ""
success "========================================="
success "  构建完成！"
success "========================================="
echo ""
info "可执行文件: ./assistant-qisumi"
info "静态文件目录: ./static/"
echo ""
info "运行应用:"
echo "  ./assistant-qisumi"
echo ""
