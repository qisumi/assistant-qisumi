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
cmd/server/main.go          # Entry point: config -> logger -> DB -> HTTP server
internal/
├── domain/                 # Core domain models (no dependencies on other internal packages)
│   ├── models.go           # User, Task, Session, Message models
│   ├── llm_config.go       # LLM configuration models
│   ├── task_creation.go    # Task creation request/response models
│   ├── time.go             # FlexibleTime type for dates
│   └── utils.go            # Domain utilities
├── agent/                  # AI agent system
│   ├── service.go          # Main orchestrator: handles messages, applies patches
│   ├── router.go           # Routes requests to appropriate agents
│   ├── router_agent.go     # Agent interface and router implementation
│   ├── chat_completions.go # LLM interaction handler
│   ├── global_agent.go     # General-purpose assistant
│   ├── executor_agent.go   # Task execution guidance
│   ├── planner_agent.go    # Task planning/rescheduling
│   ├── summarizer_agent.go # Conversation summarization
│   ├── task_creation_agent.go # Text-to-task creation
│   ├── tool_executors.go   # Tool execution implementations
│   ├── tool_args.go        # Tool argument schemas
│   ├── patch.go            # TaskPatch model and patch kinds
│   ├── prompt_builder.go   # Builds prompts with context
│   ├── models.go           # Agent request/response models
│   └── fallback_message.go # Fallback message generation
├── prompts/                # Centralized agent system prompts
├── http/                   # Gin HTTP handlers and routes
├── auth/                   # JWT authentication and user settings
├── config/                 # Environment-based configuration
├── db/                     # GORM database initialization and auto-migration
├── llm/                    # LLM client (OpenAI-compatible)
├── session/                # Chat session and message repository
├── task/                   # Task and step repository and service
├── dependency/             # Task-step dependency resolution
├── logger/                 # Zap logger wrapper
└── common/                 # Shared utilities
```

### Domain Layer Architecture

**Critical:** The `internal/domain/` package is the foundation layer. It contains core domain models and has **no dependencies** on other `internal/` packages. This prevents circular dependencies.

When adding new models or types:
- Place shared domain models in `internal/domain/`
- Other packages import from domain, not vice versa
- Recent refactoring extracted models from `task/`, `auth/`, `session/` into `domain/`

### Agent System Flow

The agent system follows a clear flow:

1. **Routing** (`router.go`): Determines which agent handles the request based on session type and keywords
2. **Context Building** (`service.go`): Loads task, session, messages, and dependencies
3. **Agent Execution** (`*agent.go`): Each agent implements `Handle(req) -> response`
4. **Patch Application** (`service.go`): Applies `TaskPatch` objects in a transaction
5. **Message Storage** (`service.go`): Saves user/assistant messages

**Agent Types:**
- **global**: Used when `session.Type == "global"` - handles any query
- **task_creation**: Creates tasks from natural language (router-independent, invoked separately)
- **summarizer**: Triggered by keywords "总结" or "overview"
- **planner**: Triggered by keywords "重排", "reschedule", "重新规划"
- **executor**: Default agent for task execution guidance

**Key Pattern - TaskPatches:**
Agents never modify tasks directly. They return `TaskPatch` objects with `PatchKind`:
- `PatchUpdateTask`: Update task fields
- `PatchUpdateStep`: Update step fields
- `PatchAddSteps`: Insert new steps
- `PatchAddDependencies`: Add dependency relationships
- `PatchMarkTasksFocusToday`: Update focus_today flags

The service layer applies patches in `applyTaskPatches()` within a GORM transaction.

### Dependency System

Dependencies are stored in `task_dependencies` with:
- **Predecessor**: The task/step that must complete
- **Successor**: The task/step affected by completion
- **Condition**: "task_done" or "step_done"
- **Action**: "unlock_step", "set_task_todo", or "notify_only"

When a task/step completes, `dependency.Service.OnTaskOrStepDone()` triggers:
- Unlocks locked steps
- Sets successor tasks to todo
- Sends notification messages

### Database Layer

- **Auto-migration**: Runs on startup via `db.AutoMigrate()`
- **Repository Pattern**: Each domain has a `*Repository` with `WithTx(tx)` for transaction support
- **Transaction Management**: Critical for multi-step updates (patches, dependencies)

### LLM Integration

The LLM client (`internal/llm/client.go`) uses OpenAI SDK with:
- **Config**: `domain.LLMConfig` contains API key, base URL, model
- **Deep Thinking**: Supports `thinking_type` (disabled/enabled/auto) and `reasoning_effort` (low/medium/high)
- **Tool Calling**: Agents define tools, LLM returns tool calls, executors implement them
- **Prompt Building**: `prompt_builder.go` constructs prompts with task context, history, and tools

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

1. **Repository Pattern**: Data access through `*Repository` types with `WithTx()` for transaction support
2. **Service Layer**: Business logic in `*Service` types (e.g., `agent.Service`, `dependency.Service`)
3. **Agent Patches**: Agents return patches to modify tasks, not direct mutations
4. **Transaction Management**: Use GORM transactions for multi-step updates, especially when applying patches
5. **Router-Based Agent Selection**: Agent router determines which agent handles a request based on session type and keywords
6. **Session Types**: Different agent behaviors based on session type ("global" vs "task")
7. **Domain Layer Independence**: `internal/domain/` has no dependencies on other internal packages
8. **Centralized Prompts**: All agent system prompts are in `internal/prompts/prompts.go`

## Important Implementation Details

### Task Status Auto-Updates

When a step completes, the service layer automatically updates task status:
- If any step completes: `todo` → `in_progress`
- If all steps complete: `in_progress` → `done`
- See `service.go:applyUpdateStepFields()` logic

### FlexibleTime Type

`domain.FlexibleTime` handles dates with three formats:
- Absolute dates: "2024-12-25"
- Relative dates: "today", "tomorrow", "next week"
- No date set: null

### API Key Encryption

User LLM API keys are encrypted using AES-256 before storage. See `API_KEY_ENCRYPTION_KEY` in config.

### Logging

Structured logging with Zap:
- Info level: API calls, agent routing
- Debug level: Detailed request/response content
- Error level: All errors with context

## Future Enhancements

Based on the README, planned features include:
1. Dark mode support
2. Streaming response support
3. Deadline bug fixes
4. Implicit dependency step completion automation
5. Loading process optimization