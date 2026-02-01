# Codebase Quick Reference

## Where to Find...

### Task-Related Code
- **Task model:** `internal/models/task.go`
- **Task service:** `internal/services/task.go`
- **Task handlers:** `internal/handlers/task.go`
- **Task store (frontend):** `ui/src/lib/stores/tasks.svelte.ts`

### Agent Execution
- **Orchestrator:** `internal/services/orchestrator.go` - Task execution engine
- **Agent backends:** `internal/agent/` - Native, Claude Code implementations
- **Event system:** `internal/services/events.go` - Pub/sub for agent events

### Database
- **Schema:** `internal/db/schema.sql`
- **Queries:** `internal/db/queries/*.sql`
- **Generated code:** `internal/db/sqlc/` (DO NOT EDIT - regenerate with `make sqlc`)

### Authentication
- **JWT validation:** `internal/auth/jwt.go`
- **Supabase integration:** `internal/auth/supabase.go`
- **Auth middleware:** `internal/auth/middleware.go`

### API Routes
- **Registration:** `internal/handlers/handlers_registration.go` - All routes and handlers
- **SSE endpoint:** `internal/handlers/sse.go` - Real-time events
- **Auth callback:** `internal/handlers/auth.go` - GitHub OAuth flow

### Frontend Components
- **Feed:** `ui/src/lib/components/Feed.svelte`
- **Task:** `ui/src/lib/components/Task.svelte`
- **Dashboard layout:** `ui/src/routes/dashboard/+layout.svelte`

## Architecture Flow

```
User Request
    ↓
Handler (internal/handlers/*.go)
    ↓
Service Layer (internal/services/*.go)
    ↓
Database (SQLite/PostgreSQL via sqlc)
    ↓
Event Bus (internal/services/events.go)
    ↓
SSE Stream (internal/handlers/sse.go)
    ↓
Frontend (ui/src/lib/stores/*.svelte.ts)
```

## Service Dependencies

```
Orchestrator
    ├─→ TaskService
    ├─→ RepoService (git operations)
    ├─→ FileService (file system)
    ├─→ GitHubService (GitHub API)
    └─→ Agent (internal/agent/*.go)

TaskService
    └─→ DB (via sqlc-generated queries)

AuthService
    └─→ Supabase or local JWT
```

## Error Patterns

**Common error locations:**
- Database errors → Check `internal/db/queries/*.sql` and generated code
- Agent execution → Check `internal/services/orchestrator.go` and agent logs in `server.log`
- Auth failures → Check `internal/auth/*.go` and Supabase config
- Frontend state → Check `ui/src/lib/stores/*.svelte.ts` and SSE connection

**Debug with:** `tail -f server.log | grep -E "error|Error|ERROR"`

## Adding New Features

1. **New model:** Add to `internal/models/`, update `schema.sql`, run `make sqlc`
2. **New API route:** Add handler in `internal/handlers/*.go`, register in `handlers_registration.go`
3. **New service:** Create in `internal/services/*.go`, inject via dependency injection
4. **New UI component:** Add to `ui/src/lib/components/`, import in routes
5. **New page:** Add to `ui/src/routes/`

**After UI changes:** Run `make test-e2e` to verify no regressions

**Before testing UI:** Run `cd ui && npm run build` to embed changes in Go binary

## File Naming Conventions

- **Go handlers:** `{resource}.go` (e.g., `task.go`, `auth.go`)
- **Go services:** `{service}.go` (e.g., `task.go`, `orchestrator.go`)
- **Svelte components:** PascalCase (e.g., `Task.svelte`, `Feed.svelte`)
- **Svelte stores:** `{resource}.svelte.ts` (e.g., `tasks.svelte.ts`)
- **TypeScript types:** `{domain}.ts` (e.g., `task.ts`, `agent.ts`)

## Testing Locations

- **Go tests:** Place in `*_test.go` next to implementation
- **Frontend E2E tests:** `ui/tests/e2e/*.spec.ts` (Playwright)
- **Frontend unit tests:** `ui/src/*.test.ts` (not yet implemented)

## Testing Commands

```bash
make test        # Run Go tests
make test-e2e    # Run Playwright E2E tests (starts dev server)
cd ui && npm run test:e2e:ui  # Run E2E tests with interactive UI
cd ui && npm run test:e2e:report  # View HTML report
```

## Configuration Files

- **Go modules:** `go.mod`, `go.sum`
- **Node deps:** `ui/package.json`, `ui/package-lock.json`
- **Docker:** `Dockerfile`, `docker-compose.yml`
- **Makefile:** All build/dev commands

## Environment Setup

Required tools:
- Go 1.25+
- Node.js 20+
- sqlc (for DB code generation)
- Docker (optional, for containerized runs)

First-time setup:
```bash
make preprod  # Build everything
make dev      # Start dev server
```
