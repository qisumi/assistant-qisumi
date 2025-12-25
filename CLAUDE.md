# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

**Assistant Qisumi** is an AI-powered task management application with a Go backend and React frontend. It provides intelligent task planning, execution assistance, and conversation-based task management using LLM integration.

### Tech Stack

**Backend:** Go 1.24.5, Gin (HTTP), GORM (ORM), SQLite/MySQL, Zap (logging), JWT auth
**Frontend:** React 18, Vite, Ant Design, Zustand, React Query, TypeScript
**LLM:** OpenAI-compatible API (defaults to Qwen-plus/Alibaba Cloud)

## Development Commands

### Backend
```bash
# Run development server
go run ./cmd/server

# Run tests
go test ./...

# Run integration test (verifies config, DB, LLM client initialization)
go run test_integration.go

# Build for production
go build -o server ./cmd/server
```

### Frontend
```bash
cd frontend

# Install dependencies
npm install

# Start development server
npm run dev

# Build for production
npm run build

# Preview production build
npm run preview

# Run linting
npm run lint
```

## Architecture

### Backend Structure

```
cmd/server/main.go          # Entry point: config -> DB -> HTTP server
internal/
├── agent/                  # AI agent system
│   ├── service.go          # Main agent orchestrator
│   ├── router.go           # Routes requests to appropriate agents
│   ├── executor_agent.go   # Task execution guidance
│   ├── planner_agent.go    # Task planning/rescheduling
│   ├── summarizer_agent.go # Conversation summarization
│   ├── global_agent.go     # General-purpose assistant
│   ├── task_creation_agent.go # Text-to-task creation
│   └── tool_executors.go   # Tool execution (task updates, etc.)
├── http/                   # Gin HTTP handlers and routes
├── auth/                   # JWT authentication
├── config/                 # Environment-based configuration
├── db/                     # GORM database layer with auto-migration
├── llm/                    # LLM client (OpenAI-compatible)
├── session/                # Chat session management
├── task/                   # Task and step CRUD
└── dependency/             # Task-step dependency resolution
```

### Agent System

The agent router (`internal/agent/router.go`) routes user requests to specialized agents:

- **Global Agent**: Used when `session.Type == "global"` - handles any query
- **Summarizer Agent**: Triggered by keywords "总结" or "overview"
- **Planner Agent**: Triggered by keywords "重排", "reschedule", "重新规划"
- **Executor Agent**: Default agent for task execution guidance
- **Task Creation Agent**: Creates tasks from natural language

Agents return `TaskPatch` objects to modify tasks rather than direct changes. Tool executors handle specific actions like task updates.

### Frontend Structure

```
frontend/src/
├── api/                    # Axios-based API clients
├── components/
│   ├── chat/               # Chat components (MessageList, InputBox)
│   ├── common/             # Shared components (TaskCard, StepList)
│   └── layout/             # App layout (Sider, Header)
├── pages/                  # Route pages (Tasks, TaskDetail, GlobalAssistant, etc.)
├── store/                  # Zustand stores (auth)
├── App.tsx                 # React Router setup
└── main.tsx                # Entry point
```

State management: Zustand for auth, React Query for server state, local component state for UI.

### Database Schema

Core tables: `users`, `user_llm_settings`, `tasks`, `task_steps`, `task_dependencies`, `sessions`, `messages`

Key relationships:
- Tasks have many steps
- Steps can depend on other tasks/steps (with conditions to unlock or trigger actions)
- Task status auto-updates based on step completion
- Dependencies are stored and resolved through `internal/dependency/`

### Configuration

Environment variables (`.env`):
- `DB_TYPE`: sqlite or mysql
- `DB_FILE_PATH`: SQLite file path (default: assistant.db)
- `HTTP_HOST`/`HTTP_PORT`: Server binding (default: 0.0.0.0:4569)
- `JWT_SECRET`/`JWT_EXPIRE_HOUR`: JWT authentication
- `API_KEY_ENCRYPTION_KEY`: 32-byte key for encrypting stored API keys
- `LOG_LEVEL`: Zap logging level
- `LLM_API_KEY`: LLM provider API key
- `LLM_MODEL_NAME`: Model name (default: qwen-plus)
- `LLM_API_BASE_URL`: API endpoint (default: Alibaba Cloud)

## API Endpoints

### Auth
- `POST /api/auth/login` - User login
- `GET /api/auth/me` - Get current user

### Tasks
- `GET /api/tasks` - List tasks
- `POST /api/tasks` - Create task
- `GET /api/tasks/:id` - Get task details
- `PUT /api/tasks/:id` - Update task
- `DELETE /api/tasks/:id` - Delete task

### Sessions
- `GET /api/sessions` - List sessions
- `POST /api/sessions` - Create session
- `POST /api/sessions/:id/chat` - Send message to session

### Settings
- `GET /api/settings/llm` - Get LLM settings
- `PUT /api/settings/llm` - Update LLM settings

## Key Patterns

1. **Repository Pattern**: Data access through `*Repository` types in each domain
2. **Service Layer**: Business logic in `*Service` types
3. **Agent Patches**: Agents return patches to modify tasks, not direct mutations
4. **Transaction Management**: Use GORM transactions for multi-step updates
5. **Router-Based Agent Selection**: Agent router determines which agent handles a request
6. **Session Types**: Different agent behaviors based on session type (global vs task-specific)

## Future Enhancements (TODO.md)

1. 暗色模式支持
2. 流式响应支持