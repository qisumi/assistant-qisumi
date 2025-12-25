# 部署方式说明

本项目支持两种部署方式：

## 方式一：统一部署（推荐）

**优点**：
- 简单快捷，单一可执行文件
- 前端静态文件内嵌在二进制中
- 部署和更新非常方便
- 适合大多数场景

### 本地开发构建

#### Linux/macOS

```bash
# 使用构建脚本（推荐）
chmod +x build.sh
./build.sh

# 或使用 Makefile
make build

# 运行应用
./assistant-qisumi
```

#### Windows

```powershell
# 使用 PowerShell 脚本
.\build.ps1

# 运行应用
.\assistant-qisumi.exe
```

### Docker 部署

```bash
# 使用统一 Docker Compose 配置
docker compose -f docker-compose.standalone.yml up -d

# 查看日志
docker compose -f docker-compose.standalone.yml logs -f

# 停止服务
docker compose -f docker-compose.standalone.yml down
```

### 访问应用

- **应用地址**: http://localhost:4569
- **API 地址**: http://localhost:4569/api/
- **健康检查**: http://localhost:4569/api/health

## 方式二：分离部署

**优点**：
- 前后端独立部署和扩展
- 适合需要独立优化的场景
- 可以使用 CDN 加速前端

### Docker 部署

```bash
# 使用分离式 Docker Compose 配置
docker compose up -d

# 查看日志
docker compose logs -f

# 停止服务
docker compose down
```

### 访问应用

- **前端地址**: http://localhost
- **后端 API**: http://localhost:4569/api/

## 构建产物说明

### 统一部署模式

```
assistant-qisumi/
├── assistant-qisumi      # 包含静态文件的可执行文件
└── static/               # 静态文件目录（仅在构建时存在，会被嵌入到二进制中）
    ├── index.html
    └── assets/
```

### 分离部署模式

```
assistant-qisumi/
├── backend/
│   └── assistant-qisumi  # 后端可执行文件
└── frontend/
    └── dist/             # 前端构建产物
        ├── index.html
        └── assets/
```

## 开发模式

### 前后端分离开发

```bash
# 终端 1: 启动后端
go run ./cmd/server

# 终端 2: 启动前端
cd frontend
npm run dev
```

前端开发服务器默认运行在 http://localhost:5173

### 快速构建开发

使用 Makefile 可以快速构建：

```bash
# 仅构建前端
make frontend

# 仅构建后端
make backend

# 同时构建前后端
make build

# 清理构建产物
make clean

# 运行应用
make run
```

## 生产环境建议

1. **推荐使用统一部署模式**
   - 部署简单，维护方便
   - 单一服务，资源占用少
   - 更新时只需要替换一个文件

2. **如果需要分离部署**
   - 使用 Nginx 托管前端静态文件
   - 配置反向代理到后端 API
   - 使用 HTTPS 证书

3. **数据库选择**
   - 开发/小规模部署：SQLite（默认）
   - 生产环境：MySQL（推荐）

4. **安全配置**
   - 修改 JWT_SECRET
   - 修改 API_KEY_ENCRYPTION_KEY
   - 使用强密码策略
   - 配置防火墙规则

## 故障排查

### 静态文件无法加载

1. 检查 static 目录是否存在：
   ```bash
   ls -la static/
   ```

2. 重新构建：
   ```bash
   make clean
   make build
   ```

### API 请求失败

1. 检查后端是否正常运行：
   ```bash
   curl http://localhost:4569/api/health
   ```

2. 查看日志：
   ```bash
   docker compose logs -f
   ```

### Docker 构建失败

1. 清理 Docker 缓存：
   ```bash
   docker system prune -a
   ```

2. 重新构建：
   ```bash
   docker compose build --no-cache
   ```

## 性能优化

### 统一部署模式

- 二进制文件已包含所有静态资源
- 无需额外配置 CDN
- 适合小型到中型应用

### 分离部署模式

- 前端可以使用 CDN 加速
- Nginx 提供 Gzip 压缩
- 静态资源缓存策略
- 适合大型应用

## 更新应用

### 统一部署模式

```bash
# 停止服务
docker compose -f docker-compose.standalone.yml down

# 拉取最新代码
git pull origin master

# 重新构建并启动
docker compose -f docker-compose.standalone.yml up -d --build
```

### 分离部署模式

```bash
# 使用更新脚本
./update.sh
```

## 监控和日志

### 查看应用日志

```bash
# Docker 部署
docker compose logs -f

# 直接运行
./assistant-qisumi 2>&1 | tee app.log
```

### 健康检查

```bash
# 检查 API 健康状态
curl http://localhost:4569/api/health

# 检查容器健康状态
docker compose ps
```

## 备份和恢复

### 数据备份

```bash
# 备份 SQLite 数据库
docker cp qisumi-app:/app/data/assistant.db backups/$(date +%Y%m%d)/
```

### 数据恢复

```bash
# 恢复数据库
docker cp backups/20241225/assistant.db qisumi-app:/app/data/assistant.db

# 重启服务
docker compose restart
```
