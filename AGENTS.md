# Agent Instructions for Counterspell

## Development Principles

1. **Clarity > Cleverness:** Optimize for maintainability. Use existing patterns over new ones.
2. **Atomic Changes:** Make the smallest change necessary to solve the problem.
3. **Don't Patch, Fix Root Cause:** Find and fix the underlying issue.
4. **Simplicity** is the prerequisite of reliability.
5. **Instruction Format:** [Rationale] → [Action]. Max 1 line rationale explaining 'why'.

### Change Size Heuristic

| Size | Scope | Action |
|------|-------|--------|
| Small | ≤1 package, no API change | Proceed |
| Medium | 2–3 packages or API change | Ask first |
| Large | >3 packages, new dependency, or schema change | Design review required |

### Constraints

Agents **MUST NOT**:
- Refactor unrelated code
- Start another server (Backend + UI server is always running)
- Change test outputs without approval
- Add dependencies without approval

**Definition of Done:**
- Code compiles and all tests pass
- New behavior is covered by tests
- No unrelated changes included

### Size Guidelines
- Functions: < 50 lines preferred
- Files: < 500 lines preferred

---

## Project Overview

Task orchestration platform that runs AI agents on user codebases.

| Layer | Technology |
|-------|------------|
| Backend | Go + Chi router |
| Frontend | SvelteKit SPA (embedded via `go:embed`) |
| Database | PostgreSQL (row-level user isolation via `user_id`) |
| Auth | Supabase (optional - falls back to single-user mode) |

### Architecture

```
┌─────────────┐     ┌─────────────┐     ┌─────────────┐
│   Browser   │────▶│   Handlers  │────▶│  Services   │
│  SvelteKit  │◀────│   (Chi)     │◀────│             │
└─────────────┘     └─────────────┘     └─────────────┘
       │                   │                   │
       │ SSE + JSON API    │                   │
       ▼                   ▼                   ▼
┌─────────────┐     ┌─────────────┐     ┌─────────────┐
│  EventBus   │     │ Orchestrator│     │ PostgreSQL  │
│  (pub/sub)  │     │ (worker)    │     │ (shared DB) │
└─────────────┘     └─────────────┘     └─────────────┘
```

---

## Directory Structure

```
cmd/app/              # Entry point
internal/
  agent/              # AI agent backends (native, claude-code)
  auth/               # JWT validation, Supabase auth
  config/             # Configuration management
  db/
    schema.sql        # PostgreSQL schema
    queries/          # SQL queries for sqlc
    sqlc/             # Generated code (don't edit)
  handlers/           # HTTP handlers (JSON API)
  models/             # Domain models
  services/           # Business logic
ui/                   # SvelteKit frontend
  src/
    lib/
      components/     # Svelte components
      stores/         # Svelte stores (*.svelte.ts)
      types/          # TypeScript types
      api.ts          # API client
    routes/           # SvelteKit routes
  dist/               # Built output (embedded in Go binary)
data/                 # Runtime data (repos, worktrees)
```

---

## Commands

```bash
cd ui && npm run build           # Build frontend
sqlc generate                    # Regenerate after .sql changes
go build ./...                   # Build Go backend
LOG_LEVEL=debug go run ./cmd/app # Run with debug logging
```

---

## Code Style

### Go Conventions

```go
// Always propagate context; never ignore cancellation
func (s *Service) DoWork(ctx context.Context) error {
    select {
    case <-ctx.Done():
        return ctx.Err()
    default:
        // work
    }
}

// Always wrap errors with context
if err != nil {
    return fmt.Errorf("failed to process task %s: %w", id, err)
}

// Use structured logging with key-value pairs
slog.Info("Starting task", "task_id", taskID, "user_id", userID)
slog.Error("Operation failed", "error", err, "task_id", taskID)

// Service methods take userID as first user-scoped parameter
func (s *TaskService) Get(ctx context.Context, userID, taskID string) (*models.Task, error)
```

### Handler Pattern

```go
func (h *Handlers) HandleAPITask(w http.ResponseWriter, r *http.Request) {
    ctx := r.Context()
    userID := auth.UserIDFromContext(ctx)  // Always extract userID
    taskID := chi.URLParam(r, "id")

    task, err := h.taskService.Get(ctx, userID, taskID)
    if err != nil {
        http.Error(w, "Task not found", http.StatusNotFound)
        return
    }

    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(task)
}
```

---

## Svelte 5 Conventions

### Runes Pattern (Svelte 5)

```typescript
// lib/stores/tasks.svelte.ts - Class-based store with runes
class TaskStore {
    tasks = $state<Task[]>([]);
    
    // Derived state
    activeTasks = $derived(this.tasks.filter(t => t.status === 'active'));
    
    async fetchTasks() {
        const response = await api.get('/api/tasks');
        this.tasks = response.active;
    }
}

export const taskStore = new TaskStore();
```

### Component Conventions

```svelte
<!-- Use Svelte 5 runes syntax -->
<script lang="ts">
    import { taskStore } from '$lib/stores/tasks.svelte';
    
    // Props with defaults
    let { taskId, onClose = () => {} }: { taskId: string; onClose?: () => void } = $props();
    
    // Local state
    let isLoading = $state(false);
    
    // Derived values
    let task = $derived(taskStore.tasks.find(t => t.id === taskId));
    
    // Effects for side effects only (not render logic)
    $effect(() => {
        if (taskId) {
            // Subscribe to SSE, etc.
        }
    });
</script>
```

### UI Rules

| Rule | Guidance |
|------|----------|
| Accessibility | `aria-label` on icon-only buttons, keyboard navigation |
| Loading | Use skeleton screens matching content structure |
| Errors | Show next to where the action happens |
| Height | Use `h-dvh` not `h-screen` |
| Animation | Only when explicitly requested; use `transform`/`opacity` only |
| Forms | Enter key submits, labels on all fields, no paste blocking |

### SSE Subscription

```typescript
// lib/utils/sse.ts
export function subscribeToSSE(taskId: string, onEvent: (event: Event) => void) {
    const eventSource = new EventSource(`/events?task_id=${taskId}`);
    eventSource.onmessage = (e) => onEvent(JSON.parse(e.data));
    return () => eventSource.close();  // Return cleanup function
}
```

---

## Testing Policy

### Logic vs. Wiring

Before writing a test, classify the function:

**TYPE A: WIRING (Do NOT Test)**
- Pass-throughs that call another function
- Constructors (`func New...`) that assign fields
- Standard error bubbling (`if err != nil { return err }`)
- Getters/Setters

**TYPE B: LOGIC (MUST Test)**
- Math/calculations
- Conditionals checking business values
- Loops with logic
- Transformation (parsing, mapping, filtering)

### Table-Driven Tests

```go
tests := []struct {
    name    string
    input   string
    want    string
    wantErr bool
}{
    {"valid input", "foo", "FOO", false},
    {"empty input", "", "", true},
}
for _, tt := range tests {
    t.Run(tt.name, func(t *testing.T) {
        t.Parallel()
        got, err := Function(tt.input)
        if tt.wantErr {
            require.Error(t, err)
            return
        }
        require.NoError(t, err)
        assert.Equal(t, tt.want, got)
    })
}
```

---

## Common Pitfalls

### 1. Forgetting userID Parameter

```go
// WRONG
h.taskService.Get(ctx, taskID)

// CORRECT
userID := auth.UserIDFromContext(ctx)
h.taskService.Get(ctx, userID, taskID)
```

### 2. Frontend Build Required

After Svelte changes, rebuild before testing:
```bash
cd ui && npm run build
```
The Go binary embeds `ui/dist/` - changes won't appear until rebuilt.

### 3. sqlc Regeneration

After modifying `.sql` files:
```bash
sqlc generate
go build ./...  # Verify it compiles
```

### 4. Svelte 5 Pitfalls

```typescript
// WRONG: useEffect for render logic
$effect(() => {
    filteredTasks = tasks.filter(t => t.status === 'active');
});

// CORRECT: Use $derived
let filteredTasks = $derived(tasks.filter(t => t.status === 'active'));
```

---

## Key Files Reference

| File | Purpose |
|------|---------|
| `cmd/app/main.go` | Entry point, server setup |
| `internal/handlers/handlers_registration.go` | Handler struct, service injection |
| `internal/handlers/sse.go` | SSE streaming endpoint |
| `internal/services/orchestrator.go` | Task execution, agent coordination |
| `internal/db/schema.sql` | PostgreSQL schema |
| `ui/src/lib/stores/tasks.svelte.ts` | Task state management |
| `ui/src/lib/api.ts` | API client |

---

## Environment Variables

```bash
# Required
DATABASE_URL=postgres://user:pass@localhost:5432/counterspell?sslmode=disable
GITHUB_CLIENT_ID=xxx
GITHUB_CLIENT_SECRET=xxx
GITHUB_REDIRECT_URI=http://localhost:8710/github/callback

# Optional: Supabase auth (for multi-user)
SUPABASE_URL=https://xxx.supabase.co
SUPABASE_ANON_KEY=xxx
SUPABASE_JWT_SECRET=xxx

# Optional
LOG_LEVEL=debug
DATA_DIR=./data
```

Without Supabase env vars, all requests use `userID = "default"` (single-user mode).
