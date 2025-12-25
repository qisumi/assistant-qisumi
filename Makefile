.PHONY: all build frontend backend clean run test help

# 默认目标
all: build

# 构建前端和后端
build: frontend backend
	@echo "✓ 构建完成"

# 仅构建前端
frontend:
	@echo "构建前端..."
	@cd frontend && npm install && npm run build
	@mkdir -p static
	@cp -r frontend/dist/* static/
	@echo "✓ 前端构建完成"

# 仅构建后端
backend:
	@echo "构建后端..."
	@CGO_ENABLED=1 go build -o assistant-qisumi ./cmd/server
	@echo "✓ 后端构建完成"

# 清理构建产物
clean:
	@echo "清理构建产物..."
	@rm -rf static
	@rm -f assistant-qisumi
	@rm -rf frontend/dist
	@rm -rf frontend/node_modules
	@echo "✓ 清理完成"

# 运行应用
run: build
	@echo "启动应用..."
	@./assistant-qisumi

# 开发模式（前端热更新 + 后端热重载）
dev:
	@echo "启动开发模式..."
	@echo "后端: go run ./cmd/server"
	@echo "前端: cd frontend && npm run dev"
	@make -j2 dev-backend dev-frontend

dev-backend:
	@go run ./cmd/server

dev-frontend:
	@cd frontend && npm run dev

# 运行测试
test:
	@echo "运行测试..."
	@go test ./...
	@echo "✓ 测试完成"

# 构建 Docker 镜像
docker-build:
	@echo "构建 Docker 镜像..."
	@docker build -t qisumi/assistant-qisumi:latest .
	@echo "✓ Docker 镜像构建完成"

# 运行 Docker 容器
docker-run:
	@echo "运行 Docker 容器..."
	@docker compose up -d

# 停止 Docker 容器
docker-stop:
	@echo "停止 Docker 容器..."
	@docker compose down

# 查看日志
logs:
	@docker compose logs -f

# 帮助信息
help:
	@echo "Assistant Qisumi - 构建命令"
	@echo ""
	@echo "使用方法: make [target]"
	@echo ""
	@echo "可用目标:"
	@echo "  all           构建前端和后端（默认）"
	@echo "  build         同 all"
	@echo "  frontend      仅构建前端"
	@echo "  backend       仅构建后端"
	@echo "  clean         清理构建产物"
	@echo "  run           构建并运行应用"
	@echo "  dev           开发模式（前后端同时运行）"
	@echo "  dev-backend   仅运行后端开发服务器"
	@echo "  dev-frontend  仅运行前端开发服务器"
	@echo "  test          运行测试"
	@echo "  docker-build  构建 Docker 镜像"
	@echo "  docker-run    运行 Docker 容器"
	@echo "  docker-stop   停止 Docker 容器"
	@echo "  logs          查看 Docker 日志"
	@echo "  help          显示此帮助信息"
	@echo ""
