# AGENTS.md

## Project summary
- Go backend (Gin + GORM) and Vite/React frontend.
- HTTP API under /api (see internal/http).
- LLM integration via internal/llm and internal/agent.

## Key locations
- cmd/server/main.go: backend entrypoint
- internal/config: env loading and defaults
- internal/db: DB init and AutoMigrate
- internal/http: routing/handlers
- internal/task, internal/session, internal/auth: core domain models
- frontend/src: React UI (pages, components, api, store)
- docs/: architecture and API alignment docs (Chinese)

## Setup
Backend:
- Copy `.env.example` to `.env` and adjust as needed.
- Defaults: DB_TYPE=sqlite uses `assistant.db`, HTTP_PORT=4569.
- Run: `go run ./cmd/server`

Frontend:
- `cd frontend`
- Copy `.env.example` to `.env` (VITE_API_URL).
- Install deps: `npm install`
- Dev server: `npm run dev`
- Build: `npm run build`
- Lint: `npm run lint`

## Tests
- Go: `go test ./...`
- Some tests require MySQL at `localhost:3306` with user `root` and password `231966` (see `test/test_config.go`); they use database `assistant_qisumi`.
- `test/api_integration_test.go` uses in-memory sqlite and does not need MySQL.

## Notes for changes
- Add/modify models in `internal/*` and update `internal/db/AutoMigrate` if new tables are introduced.
- API routing is in `internal/http/server.go` and handler files.
- LLM defaults come from `LLM_API_KEY`, `LLM_MODEL_NAME`, `LLM_API_BASE_URL`.
- Generated artifacts: do not hand-edit `frontend/dist` or `frontend/node_modules`.
