# Agent Instructions for Counterspell

## Project Overview
Task orchestration platform running AI agents on user codebases.

**Stack:** Go (Chi) + SvelteKit 5 + SQLite/PostgreSQL + Supabase auth

## Control Plane + Tunnel Context (2026-01)

- **Invoker control plane:** `counterspell.io` (auth, machine registry, tunnel provisioning)
- **Data plane / tunnel domain:** `*.counterspell.app` (public URL points at local machine via Cloudflare tunnel)
- **Auth flow (CLI startup):** browser OAuth by default; device-code flow when `HEADLESS=true` or `FORCE_DEVICE_CODE=true`
- **Device approval UI:** `https://counterspell.io/device` (enters `user_code`, calls `/api/v1/auth/device/approve`)

## Quick Commands

```bash
make dev          # Start backend (:8710)
make ui           # Start frontend dev server (:5173)
make test         # Run Go tests
make test-e2e     # Run Playwright E2E tests
make verify       # Change-aware verification (sqlc/build/tests/e2e)
make sqlc         # Regenerate DB code after schema changes
make format       # Format all code (Go + Svelte)
make check-all    # Run all linters and type checks
cd ui && npm run build  # Build frontend (required before testing Go changes)
```

**Logs:** `tail -f server.log` (root directory)

## Verification Matrix

Use `make verify` (runs `scripts/verify.sh`) to automatically select checks based on local changes:

| Change Type | Commands |
|-------------|----------|
| Go only | `cd ui && npm run build` → `make test` |
| UI only | `cd ui && npm run build` → `make test-e2e` |
| SQL / schema | `make sqlc` → `cd ui && npm run build` → `make test` |
| Go + UI | `cd ui && npm run build` → `make test` → `make test-e2e` |
| Docs only | No tests |

## Development Principles

| Principle | What It Means |
|-----------|---------------|
| Clarity > Cleverness | Use existing patterns over new ones |
| Atomic Changes | Smallest change that solves the problem |
| Don't Patch, Fix Root Cause | Find and fix underlying issues |
| Always Propagate Context | Never ignore context cancellation |
| Wrap Errors | `fmt.Errorf("context: %w", err)` |
| Structured Logging | `slog.Info("msg", "key", value)` |
| Use sqlc Queries | Always use sqlc-generated queries, not raw SQL or JSON operations |

## Local Auth Storage (Counterspell)

- **SQLite (local):** `settings.machine_jwt` and `settings.machine_id`
- **Machine metadata:** `machine_identity` (subdomain, tunnel_token, etc.)
- **Do not expose** `machine_jwt` via public settings APIs

## Auth + Tunnel Environment Variables

```bash
# Control plane base URL
INVOKER_BASE_URL=https://counterspell.io

# Auth flow
HEADLESS=false
FORCE_DEVICE_CODE=false
INVOKER_OAUTH_PROVIDER=google

# OAuth callback port (local callback server)
OAUTH_CALLBACK_PORT=8711
```

## Change Size Heuristic

| Size | Scope | Action |
|------|-------|--------|
| Small | ≤1 package, no API change | Proceed |
| Medium | 2–3 packages or API change | Ask first |
| Large | >3 packages, new dep, or schema change | Design review required |

## Constraints

**MUST NOT:**
- Refactor unrelated code
- Start another server (servers already running)
- Change test outputs without approval
- Add dependencies without approval
- Make UI changes without running E2E tests

**Definition of Done:** Code compiles, tests pass, new behavior covered by tests, no unrelated changes

**UI Changes:** After any Svelte/component changes, run `make test-e2e` to verify

## Go Code Style

```go
// Service methods: userID as first scoped parameter
func (s *TaskService) Get(ctx context.Context, userID, taskID string) (*models.Task, error)

// Handlers: Always extract userID from context
func (h *Handlers) HandleAPITask(w http.ResponseWriter, r *http.Request) {
    ctx := r.Context()
    userID := auth.UserIDFromContext(ctx)  // ALWAYS
    taskID := chi.URLParam(r, "id")
    // ...
}
```

## Svelte 5 Code Style

```typescript
// Use $derived for computed values, NOT $effect
let filteredTasks = $derived(tasks.filter(t => t.status === 'active'));

// $effect only for side effects (subscriptions, API calls)
$effect(() => {
    // Subscribe to SSE, etc.
});

// Props with defaults
let { taskId, onClose = () => {} }: { taskId: string; onClose?: () => void } = $props();
```

## UI Rules

| Rule | Requirement |
|------|-------------|
| Accessibility | `aria-label` on icon-only buttons |
| Loading | Skeleton screens matching content |
| Errors | Show next to action |
| Height | Use `h-dvh` not `h-screen` |
| Animation | Only when requested; `transform`/`opacity` only |
| Forms | Enter key submits, labels required |

## Testing Policy

**DO NOT TEST (Wiring):** Pass-throughs, constructors, error bubbling, getters/setters

**MUST TEST (Logic):** Math/calculations, business conditionals, loops with logic, transformations (parsing, mapping, filtering)

## Common Pitfalls

1. **Forgotten userID** → Always extract via `auth.UserIDFromContext(ctx)`
2. **UI changes not tested** → After Svelte changes, run `make test-e2e`
3. **Frontend not rebuilt** → `cd ui && npm run build` after Svelte changes
4. **sqlc not regenerated** → `make sqlc` after `.sql` changes
5. **Using $effect for derived state** → Use `$derived` instead

## Key Files

| File | Purpose |
|------|---------|
| `cmd/app/main.go` | Entry point |
| `internal/handlers/handlers_registration.go` | Handler struct, service injection |
| `internal/handlers/sse.go` | SSE streaming |
| `internal/services/orchestrator.go` | Task execution, agent coordination |
| `internal/services/oauth.go` | Invoker auth + device flow + machine registration |
| `internal/tunnel/cloudflare.go` | Cloudflare tunnel runner |
| `internal/db/schema.sql` | Database schema |
| `ui/src/lib/stores/tasks.svelte.ts` | Task state management |
| `ui/src/lib/api.ts` | API client |

## Environment Variables

```bash
DATABASE_URL=postgres://user:pass@localhost:5432/counterspell?sslmode=disable
GITHUB_CLIENT_ID=xxx
GITHUB_CLIENT_SECRET=xxx
GITHUB_REDIRECT_URI=http://localhost:8710/github/callback

# Supabase (multi-user mode)
SUPABASE_URL=https://xxx.supabase.co
SUPABASE_ANON_KEY=xxx
SUPABASE_JWT_SECRET=xxx
VITE_SUPABASE_URL=https://xxx.supabase.co
VITE_SUPABASE_ANON_KEY=xxx

LOG_LEVEL=debug
DATA_DIR=./data
```
