# Copilot Instructions - assistant-qisumi

## 项目概述

基于 AI 的**任务规划 & 备忘录系统**，用户通过自然语言与多 Agent 协作完成任务管理。

**技术栈**: Go 1.24+ (推荐 1.25) 后端 + React 18 (Vite + Zustand) 前端 + MySQL/SQLite 数据库

## 核心架构

### Multi-Agent 系统 (`internal/agent/`)

系统采用 Router + 多 Agent 分发架构：
- **Router** (`router.go` / `router_agent.go`): 
  - `SimpleRouter`: 基于关键词匹配路由（当前默认使用）。
  - `RouterAgent`: 基于 LLM 的智能路由（可选）。
- **ExecutorAgent**: 执行类操作（标记完成、修改任务属性）。
- **PlannerAgent**: 规划类操作（拆分步骤、重排日程）。
- **SummarizerAgent**: 单任务总结。
- **GlobalAgent**: 跨任务全局查询（如"今天要做什么"）。
- **TaskCreationAgent**: 从自然语言文本创建任务。

**关键模式**: 
- Agent 通过 LLM tool calling 返回 `TaskPatch` 结构，由 `service.go` 统一应用到数据库。
- `ChatCompletionsHandler` (`chat_completions.go`) 统一处理 LLM 交互与工具调用逻辑。

### TaskPatch 机制 (`internal/agent/patch.go`)

所有 Agent 对数据的修改必须通过 TaskPatch 声明式描述：
```go
// PatchKind: update_task, update_step, add_steps, add_dependencies, mark_tasks_focus_today, create_task
TaskPatch{Kind: PatchUpdateStep, UpdateStep: &UpdateStepPatch{...}}
```

### 数据流

```
用户输入 → Router.Route() → Agent.Handle() → LLM Chat (tool calling)
    → BuildPatchesFromToolCalls() → Service.applyTaskPatches() → 数据库更新
```

## 开发工作流

### 运行项目

```bash
# 后端（需要先配置 .env）
go run ./cmd/server/main.go

# 前端
cd frontend && npm install && npm run dev
```

### 环境配置 (.env)

必需配置项：
- `DB_TYPE`: mysql 或 sqlite
- `JWT_SECRET`: JWT 签名密钥
- `API_KEY_ENCRYPTION_KEY`: 用户 LLM API Key 加密密钥（32字节）

### 测试

```bash
# 运行所有测试（推荐禁用缓存以确保真实性）
go test -count=1 ./test/...

# 单独运行集成测试
go test ./test/ -run TestFullTaskWorkflow
```

测试配置在 `test/test_config.go`，使用 `SetupTestUser()` 初始化测试用户。目前后端集成测试已全部通过。

## 代码约定

### Agent 开发

1. 新建 Agent 需实现 `Agent` 接口：`Name()` + `Handle(AgentRequest) (*AgentResponse, error)`
2. Prompt 模板统一放在 `internal/agent/prompts.go`
3. 使用 `prompt_builder.go` 中的辅助函数构建消息
4. Tool schema 定义在 `internal/llm/tools.go`（提供 `CommonTools()`, `ExecutorTools()`, `PlannerTools()`, `GlobalTools()` 等分组函数）
5. Tool 参数解析结构体定义在 `internal/agent/tool_args.go`

### LLM 交互

- 使用 OpenAI-Compatible API（默认阿里云 DashScope）
- 用户 LLM 配置加密存储，通过 `LLMSettingService` 管理
- Tool calling 响应解析：`tool_to_patch.go` 中的 `BuildPatchesFromToolCalls()`

### 数据库模型

- Task/TaskStep/TaskDependency 定义在 `internal/task/models.go`
- 使用 GORM 自动迁移，迁移逻辑在 `internal/db/migrate.go`
- 更新字段使用指针类型实现 partial update（`UpdateTaskFields`/`UpdateStepFields`）

### API 路由

HTTP handler 在 `internal/http/` 下，路由注册在 `server.go`：
- `/api/auth/*` - 认证（register, login）
- `/api/tasks/*` - 任务 CRUD（POST /tasks, GET /tasks, GET /tasks/:id, PATCH /tasks/:id, POST /tasks/from-text）
- `/api/sessions/*` - 会话消息（POST /sessions/:id/messages）
- `/api/settings/*` - 用户 LLM 配置（GET/POST/DELETE /settings/llm）

### Session 机制 (`internal/session/`)

每个任务有独立的 Session 对话，支持两种类型：
- `task`: 绑定特定任务的对话（`TaskID` 非空）
- `global`: 跨任务的全局对话（`TaskID` 为空）

**消息流程**：
```
POST /api/sessions/:id/messages {content: "用户输入"}
  → SessionHandler.postMessage()
  → 获取用户 LLM 配置 (LLMSettingService)
  → AgentService.HandleUserMessage()
      → 加载 Session + 最近 20 条历史消息
      → Router 选择 Agent
      → Agent 构建 prompt + 调用 LLM
      → 解析 tool calls → TaskPatches
      → 事务写入数据库 + 保存 assistant 消息
  → 返回 {assistant_message, task_patches}
```

**历史消息加载**: `ListRecentMessages()` 返回按时间正序的最近 N 条消息，用于构建 LLM 对话上下文。

## 前端架构 (`frontend/`)

**技术栈**: React 18 + TypeScript + Vite + Zustand + React Router + **Ant Design 5**

### 页面结构 (`src/pages/`)

| 页面 | 路由 | 功能 |
|------|------|------|
| `Login.tsx` | `/login` | 用户登录 |
| `Tasks.tsx` | `/tasks`, `/` | 任务列表 |
| `TaskDetail.tsx` | `/tasks/:id` | 任务详情 + 对话 |
| `GlobalAssistant.tsx` | `/global-assistant` | 全局助手对话 |
| `CreateFromText.tsx` | `/create-from-text` | 从文本创建任务 |
| `Settings.tsx` | `/settings` | LLM 配置管理 |

### 开发约定

- **UI 组件库**: Ant Design 5，中文 locale 已在 `main.tsx` 配置
- **常用组件映射**:
  - 任务列表 → `List` + `Card` + `Tag`（状态/优先级）
  - 步骤展示 → `Steps` 或 `Checkbox` 列表
  - 对话界面 → `List` + 自定义消息气泡
  - 表单 → `Form` + `Input` / `DatePicker` / `Select`
  - 图标 → `@ant-design/icons`
- 状态管理使用 Zustand（store 文件已创建，如 authStore.ts）
- API 调用封装在独立的 service 层（src/api/）

## 常见任务示例

### 添加新的 Tool

1. 在 `internal/llm/tools.go` 添加 Tool 定义
2. 在 `internal/agent/patch.go` 添加对应 PatchKind
3. 在 `internal/agent/tool_to_patch.go` 添加解析逻辑
4. 在 `internal/agent/service.go` 的 `applyTaskPatches` 添加应用逻辑

### 添加新的 Agent

1. 创建 `internal/agent/xxx_agent.go` 实现 `Agent` 接口（参考 `executor_agent.go`）
2. 在 `prompts.go` 添加 System Prompt（或直接在 Agent 内联定义）
3. 在 `internal/http/server.go` 的 `setupRoutes()` 中注册到 agents 列表
4. 更新 `router.go` 中的 `Route()` 方法添加路由规则
