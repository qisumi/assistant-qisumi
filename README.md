# Assistant Qisumi

<div align="center">

![Assistant Qisumi](https://img.shields.io/badge/Assistant-Qisumi-blue)
![Go Version](https://img.shields.io/badge/Go-1.24.5-cyan)
![React Version](https://img.shields.io/badge/React-18-blue)
![License](https://img.shields.io/badge/License-MIT-green)

**智能任务管理助手** - 基于 AI 的任务规划与执行辅助系统

[功能特性](#功能特性) • [快速开始](#快速开始) • [部署指南](#部署指南) • [技术架构](#技术架构) • [API 文档](#api-文档) • [开发指南](#开发指南)

</div>

---

## 📋 项目简介

**Assistant Qisumi** 是一个功能强大的 AI 驱动任务管理应用，通过集成大语言模型（LLM）提供智能的任务规划、执行辅助和对话式任务管理功能。

### 核心能力

- 🤖 **智能对话助手** - 支持全局助手模式和任务专属助手
- 📝 **自然语言创建任务** - 将对话转换为可执行的任务
- 🔄 **智能任务重排** - AI 自动优化任务执行顺序
- 📊 **任务执行指导** - 基于任务上下文提供执行建议
- 📈 **进度追踪与总结** - 实时监控任务进度并生成总结
- 🔐 **安全可靠** - JWT 认证、API Key 加密存储
- 🎨 **现代化界面** - 基于 Ant Design 的响应式 UI

---

## ✨ 功能特性

### 智能代理系统

应用内置多种专业 AI 代理，根据场景自动路由：

| 代理类型 | 触发条件 | 功能描述 |
|---------|---------|---------|
| **全局助手** | 全局会话模式 | 处理各类通用问题 |
| **任务创建代理** | 自然语言输入 | 将文本转化为结构化任务 |
| **规划代理** | "重排"/"规划" 关键词 | 智能优化任务执行顺序 |
| **执行代理** | 任务对话场景 | 提供任务执行指导和建议 |
| **总结代理** | "总结"/"概览" 关键词 | 生成任务进度总结 |

### 任务管理

- ✅ 任务与步骤的层级管理
- 🔗 任务依赖关系支持（条件触发/解锁）
- 📅 状态自动更新（基于步骤完成情况）
- 🔄 依赖关系解析与验证

### 多用户支持

- 👤 用户注册与登录
- ⚙️ 个性化 LLM 配置（每个用户可配置自己的模型）
- 🔑 API Key 安全加密存储
- 🗣️ 自定义助手名称

---

## 🛠️ 技术架构

### 技术栈

#### 后端
```
Go 1.24.5
├── Gin         - HTTP 框架
├── GORM        - ORM 数据访问层
├── SQLite/MySQL - 数据库
├── Zap         - 结构化日志
├── JWT         - 身份认证
└── OpenAI SDK  - LLM 客户端
```

#### 前端
```
React 18 + TypeScript
├── Vite        - 构建工具
├── Ant Design  - UI 组件库
├── Zustand     - 状态管理
├── React Query - 服务端状态管理
└── React Router - 路由管理
```

### 项目结构

```
assistant-qisumi/
├── cmd/
│   └── server/
│       └── main.go              # 后端入口
├── internal/
│   ├── agent/                   # AI 代理系统
│   │   ├── service.go           # 代理编排器
│   │   ├── router.go            # 请求路由器
│   │   ├── global_agent.go      # 全局助手
│   │   ├── planner_agent.go     # 规划代理
│   │   ├── executor_agent.go    # 执行代理
│   │   ├── summarizer_agent.go  # 总结代理
│   │   ├── task_creation_agent.go # 任务创建
│   │   └── tool_executors.go    # 工具执行器
│   ├── http/                    # HTTP 处理器
│   │   ├── server.go            # 服务器设置
│   │   ├── auth_handler.go      # 认证接口
│   │   ├── task_handler.go      # 任务接口
│   │   ├── session_handler.go   # 会话接口
│   │   └── settings_handler.go  # 设置接口
│   ├── auth/                    # 认证模块
│   │   ├── service.go           # 认证服务
│   │   ├── jwt.go               # JWT 工具
│   │   └── models.go            # 用户模型
│   ├── task/                    # 任务模块
│   │   ├── service.go           # 任务服务
│   │   ├── repo.go              # 数据访问
│   │   └── models.go            # 任务模型
│   ├── session/                 # 会话模块
│   ├── dependency/              # 依赖解析
│   ├── config/                  # 配置管理
│   ├── db/                      # 数据库初始化
│   ├── llm/                     # LLM 客户端
│   └── logger/                  # 日志系统
├── frontend/
│   ├── src/
│   │   ├── api/                 # API 客户端
│   │   ├── components/          # React 组件
│   │   │   ├── chat/           # 聊天组件
│   │   │   ├── common/         # 通用组件
│   │   │   └── layout/         # 布局组件
│   │   ├── pages/              # 页面组件
│   │   ├── store/              # Zustand 状态
│   │   ├── App.tsx             # 路由配置
│   │   └── main.tsx            # 入口文件
│   └── package.json
├── .env.example                 # 配置示例
├── CLAUDE.md                    # 项目指南
└── README.md                    # 本文件
```

---

## 🚀 快速开始

### 环境要求

- **Go**: 1.24.5+
- **Node.js**: 18+
- **数据库**: SQLite (默认) 或 MySQL 5.7+

### 1. 克隆项目

```bash
git clone https://github.com/yourusername/assistant-qisumi.git
cd assistant-qisumi
```

### 2. 后端配置

```bash
# 安装 Go 依赖
go mod download

# 复制环境变量配置文件
cp .env.example .env

# 编辑 .env 文件，配置必要信息：
# - LLM_API_KEY: LLM 服务密钥（必填）
# - LLM_MODEL_NAME: 模型名称
# - LLM_API_BASE_URL: API 端点
# - JWT_SECRET: JWT 签名密钥
# - API_KEY_ENCRYPTION_KEY: 加密密钥
```

**获取 LLM API Key**:

- 阿里云通义千问: https://dashscope.console.aliyun.com/
- OpenAI: https://platform.openai.com/api-keys
- 火山引擎豆包: https://console.volcengine.com/ark

### 3. 启动后端服务

```bash
# 开发模式
go run ./cmd/server

# 生产构建
go build -o server ./cmd/server
./server
```

服务器默认监听: `http://0.0.0.0:4569`

### 4. 前端配置

```bash
cd frontend

# 安装依赖
npm install

# 复制环境变量配置（如需要代理或其他配置）
cp .env.example .env

# 启动开发服务器
npm run dev
```

前端开发服务器默认运行在: `http://localhost:5173`

### 5. 访问应用

打开浏览器访问 `http://localhost:5173`，注册新用户并开始使用！

---

## 🚢 部署指南

本项目提供了完整的 Docker 容器化部署方案，支持一键部署和快速更新。

### 部署方式对比

| 部署方式 | 适用场景 | 难度 | 更新便利性 |
|---------|---------|------|-----------|
| **Docker Compose** | 生产环境、单机部署 | ⭐⭐ | ⭐⭐⭐⭐⭐ |
| **手动部署** | 定制化需求、学习目的 | ⭐⭐⭐⭐ | ⭐⭐ |
| **Kubernetes** | 大规模集群、云原生 | ⭐⭐⭐⭐⭐ | ⭐⭐⭐⭐ |

### 方式一：Docker Compose 部署（推荐）

这是最简单快捷的部署方式，适合大多数场景。

#### 前置要求

- Docker 20.10+
- Docker Compose 2.0+

#### Linux/macOS 部署

```bash
# 1. 克隆项目
git clone https://github.com/qisumi/assistant-qisumi.git
cd assistant-qisumi

# 2. 配置环境变量
cp .env.example .env
vim .env  # 编辑配置，至少配置 LLM_API_KEY

# 3. 一键部署
chmod +x deploy.sh
./deploy.sh

# 或部署生产环境
./deploy.sh prod
```

#### Windows 部署

```powershell
# 1. 克隆项目
git clone https://github.com/qisumi/assistant-qisumi.git
cd assistant-qisumi

# 2. 配置环境变量
Copy-Item .env.example .env
notepad .env  # 编辑配置，至少配置 LLM_API_KEY

# 3. 一键部署
.\deploy.ps1

# 或部署生产环境
.\deploy.ps1 -Environment prod
```

#### 部署完成后

- **前端地址**: http://localhost
- **后端 API**: http://localhost:4569
- **查看日志**: `docker compose logs -f`
- **停止服务**: `docker compose down`
- **重启服务**: `docker compose restart`

### 方式二：使用 Docker Compose 命令

如果您更喜欢手动控制每个步骤：

```bash
# 构建镜像
docker compose build

# 启动服务（后台运行）
docker compose up -d

# 查看运行状态
docker compose ps

# 查看日志
docker compose logs -f

# 停止服务
docker compose down

# 停止并删除数据卷（谨慎使用）
docker compose down -v
```

### 方式三：手动部署

如果您需要更灵活的部署配置：

#### 后端部署

```bash
# 1. 构建后端
go build -o assistant-qisumi ./cmd/server

# 2. 配置环境变量
cp .env.example .env
vim .env

# 3. 运行
./assistant-qisumi
```

#### 前端部署

```bash
# 1. 构建前端
cd frontend
npm install
npm run build

# 2. 使用 Nginx 托管
# 将 dist 目录复制到 Nginx root 目录
sudo cp -r dist/* /var/www/html/
```

### 更新应用

#### 自动更新（推荐）

**Linux/macOS:**
```bash
# 更新到最新版本
./update.sh

# 更新到特定版本
./update.sh 0.2.0
```

**Windows:**
```powershell
# 更新到最新版本
.\update.ps1

# 更新到特定版本
.\update.ps1 -Version 0.2.0
```

#### 手动更新

```bash
# 1. 拉取最新代码
git pull origin master

# 2. 重新构建并启动
docker compose down
docker compose build --no-cache
docker compose up -d
```

### 数据备份

#### 自动备份

更新脚本会自动备份数据库到 `backups/` 目录。

#### 手动备份

```bash
# 创建备份目录
mkdir -p backups/$(date +%Y%m%d)

# 备份数据库
docker cp qisumi-backend:/app/data/assistant.db backups/$(date +%Y%m%d)/
```

### 数据恢复

```bash
# 停止服务
docker compose down

# 恢复数据库
docker cp backups/20241225/assistant.db qisumi-backend:/app/data/assistant.db

# 重启服务
docker compose up -d
```

### 生产环境优化建议

1. **使用 HTTPS**
   - 配置 Nginx SSL 证书
   - 使用 Let's Encrypt 免费证书

2. **数据库优化**
   - 生产环境建议使用 MySQL 替代 SQLite
   - 配置定期备份

3. **监控和日志**
   - 使用 Docker 日志驱动收集日志
   - 配置健康检查和监控告警

4. **安全加固**
   - 修改默认的 JWT_SECRET
   - 使用强密码策略
   - 配置防火墙规则

5. **性能优化**
   - 启用 Nginx Gzip 压缩
   - 配置 CDN 加速静态资源
   - 使用 Redis 缓存（如需要）

### 环境变量说明

生产环境必须配置以下环境变量：

| 变量名 | 说明 | 默认值 | 生产环境建议 |
|--------|------|--------|-------------|
| `JWT_SECRET` | JWT 签名密钥 | - | **必须修改**，使用强随机字符串 |
| `API_KEY_ENCRYPTION_KEY` | API Key 加密密钥 | - | **必须修改**，32字节随机字符串 |
| `LLM_API_KEY` | LLM API 密钥 | - | **必须配置** |
| `DB_TYPE` | 数据库类型 | sqlite | mysql（生产推荐） |
| `LOG_LEVEL` | 日志级别 | info | warn（生产推荐） |

生成安全密钥：
```bash
# JWT Secret
openssl rand -base64 32

# API Key Encryption Key（32字节）
openssl rand -hex 32
```

### Docker Compose 高级配置

#### 使用外部数据库

修改 `docker-compose.yml`：

```yaml
services:
  backend:
    environment:
      - DB_TYPE=mysql
      - DB_HOST=db.example.com
      - DB_PORT=3306
      - DB_USERNAME=qisumi
      - DB_PASSWORD=your-password
      - DB_DATABASE=assistant_qisumi
```

#### 自定义端口

```yaml
services:
  backend:
    ports:
      - "8080:4569"  # 将后端映射到 8080 端口

  frontend:
    ports:
      - "8081:80"  # 将前端映射到 8081 端口
```

---

## 📚 API 文档

### 认证相关

#### `POST /api/auth/login`
用户登录

**请求体**:
```json
{
  "username": "string",
  "password": "string"
}
```

**响应**:
```json
{
  "token": "jwt-token-string",
  "user": {
    "id": 1,
    "username": "string",
    "llm_settings": {...}
  }
}
```

#### `GET /api/auth/me`
获取当前用户信息

**请求头**:
```
Authorization: Bearer {token}
```

### 任务相关

#### `GET /api/tasks`
获取任务列表

**查询参数**:
- `status`: 过滤状态 (todo/in_progress/completed)
- `page`: 页码
- `page_size`: 每页数量

#### `POST /api/tasks`
创建新任务

**请求体**:
```json
{
  "title": "任务标题",
  "description": "任务描述",
  "priority": "high",
  "due_date": "2024-12-31T23:59:59Z"
}
```

#### `GET /api/tasks/:id`
获取任务详情

#### `PUT /api/tasks/:id`
更新任务

#### `DELETE /api/tasks/:id`
删除任务

### 会话相关

#### `GET /api/sessions`
获取会话列表

#### `POST /api/sessions`
创建新会话

**请求体**:
```json
{
  "name": "会话名称",
  "type": "global",  // 或 "task"
  "task_id": 1       // type 为 task 时必填
}
```

#### `POST /api/sessions/:id/chat`
发送消息到会话

**请求体**:
```json
{
  "content": "用户消息内容"
}
```

**响应** (流式):
```json
{
  "message": {
    "id": 1,
    "role": "assistant",
    "content": "AI 回复内容",
    "created_at": "2024-12-25T10:00:00Z"
  },
  "task_patches": [...]  // 任务补丁（如有）
}
```

### 设置相关

#### `GET /api/settings/llm`
获取 LLM 配置

#### `PUT /api/settings/llm`
更新 LLM 配置

**请求体**:
```json
{
  "api_key": "sk-...",
  "model_name": "qwen-plus",
  "api_base_url": "https://dashscope.aliyuncs.com/compatible-mode/v1",
  "thinking_type": "auto",
  "reasoning_effort": "medium"
}
```

---

## 💡 使用指南

### 创建任务

1. **通过界面创建** - 点击"新建任务"按钮
2. **自然语言创建** - 在全局助手中输入任务描述，AI 会自动创建结构化任务

示例对话:
```
你: 帮我创建一个学习 Go 语言的任务
助手: 好的，我已为您创建了任务"学习 Go 语言"...
```

### 任务重排

在任务会话中使用关键词触发:
```
你: 帮我重排这些任务
你: reschedule
你: 重新规划任务顺序
```

### 获取总结

在任务会话中使用关键词触发:
```
你: 总结当前进度
你: 给我一个概览
你: overview
```

### 配置 LLM

每个用户可以配置自己的 LLM 设置:

1. 进入"设置"页面
2. 配置以下选项:
   - **API Key**: LLM 服务密钥
   - **模型名称**: 如 qwen-plus、gpt-4 等
   - **API 端点**: OpenAI 兼容接口地址
   - **深度思考模式**: disabled/enabled/auto
   - **思考强度**: minimal/low/medium/high
   - **助手名称**: 自定义显示名称

---

## 🧪 测试

### 后端测试

```bash
# 运行所有测试
go test ./...

# 运行集成测试
go run test_integration.go

# 运行特定包测试
go test ./internal/agent
```

### 前端测试

```bash
cd frontend

# 代码检查
npm run lint
```

---

## ⚙️ 配置说明

### 环境变量

| 变量名 | 说明 | 默认值 |
|--------|------|--------|
| `DB_TYPE` | 数据库类型 (sqlite/mysql) | sqlite |
| `DB_FILE_PATH` | SQLite 文件路径 | assistant.db |
| `HTTP_HOST` | 服务器监听地址 | 0.0.0.0 |
| `HTTP_PORT` | 服务器端口 | 4569 |
| `JWT_SECRET` | JWT 签名密钥 | - |
| `JWT_EXPIRE_HOUR` | JWT 过期时间(小时) | 24 |
| `API_KEY_ENCRYPTION_KEY` | API Key 加密密钥 | - |
| `LOG_LEVEL` | 日志级别 | info |
| `LLM_API_KEY` | LLM API 密钥 | - |
| `LLM_MODEL_NAME` | 模型名称 | qwen-plus |
| `LLM_API_BASE_URL` | API 端点 | - |
| `LLM_THINKING_TYPE` | 深度思考模式 | auto |
| `LLM_REASONING_EFFORT` | 思考强度 | medium |
| `ASSISTANT_NAME` | 助手名称 | 小奇 |

### 数据库迁移

应用启动时会自动执行数据库迁移，创建所需表结构。

---

## 🔒 安全特性

- **JWT 认证**: 所有 API 需要有效令牌
- **密码加密**: 使用 bcrypt 加密存储
- **API Key 加密**: 使用 AES-256 加密存储用户 LLM API Key
- **CORS 配置**: 可配置跨域访问策略
- **SQL 注入防护**: GORM 参数化查询

---

## 🗺️ 开发路线图

- [x] 基础任务管理
- [x] AI 对话助手
- [x] 多用户系统
- [x] 任务依赖关系
- [x] 自定义 LLM 配置
- [x] 深度思考模式
- [ ] 暗色模式支持
- [ ] 流式响应支持
- [ ] 任务模板功能
- [ ] 文件附件支持
- [ ] 移动端适配

---

## 🤝 贡献指南

欢迎贡献代码！请遵循以下步骤:

1. Fork 本仓库
2. 创建特性分支 (`git checkout -b feature/AmazingFeature`)
3. 提交更改 (`git commit -m 'Add some AmazingFeature'`)
4. 推送到分支 (`git push origin feature/AmazingFeature`)
5. 创建 Pull Request

### 代码规范

- **Go**: 遵循 [Effective Go](https://go.dev/doc/effective_go) 规范
- **React**: 遵循 Airbnb JavaScript 规范
- **提交信息**: 使用清晰的提交消息

---

## 📄 许可证

本项目采用 MIT 许可证 - 详见 [LICENSE](LICENSE) 文件

---

## 📞 联系方式

- 项目主页: [https://github.com/yourusername/assistant-qisumi](https://github.com/yourusername/assistant-qisumi)
- 问题反馈: [Issues](https://github.com/yourusername/assistant-qisumi/issues)
- 邮箱: your-email@example.com

---

## 🙏 致谢

感谢以下开源项目:

- [Gin](https://github.com/gin-gonic/gin) - Go Web 框架
- [GORM](https://github.com/go-gorm/gorm) - Go ORM 库
- [Ant Design](https://ant.design/) - React UI 组件库
- [Zustand](https://github.com/pmndrs/zustand) - 状态管理
- [React Query](https://tanstack.com/query) - 服务端状态管理

---

<div align="center">

**⭐ 如果这个项目对你有帮助，请给一个 Star！**

Made with ❤️ by Assistant Qisumi Team

</div>
