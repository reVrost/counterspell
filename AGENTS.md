# AGENTS.md - cs-platform (monorepo)

This repo contains the Counterspell app and Invoker control plane.

## Repo Layout
- `counterspell/` - local-first task orchestration app (Go + SvelteKit + SQLite)
- `invoker/` - control plane (Go + SvelteKit + Postgres + Supabase + Fly)

## Control Plane Context
- Control plane domain: `counterspell.io`
- Data plane domain: `*.counterspell.app` (public URL to user machine via tunnel)
- Auth flow: browser OAuth only (device flow removed)

## Quick Commands

### Counterspell
```bash
cd counterspell
make dev          # Start backend (:8710)
make ui           # Start frontend dev server (:5173)
make test         # Run Go tests
make test-e2e     # Run Playwright E2E tests
make verify       # Change-aware verification
make sqlc         # Regenerate DB code after schema changes
```

### Invoker
```bash
cd invoker
make build        # Build Go binary
make dev          # Build and run
make migrate-up   # Run migrations
make sqlc         # Regenerate sqlc code
make mock-gen     # Regenerate mocks
```

## Engineering Principles (shared)
- Use `context.Context` first in method signatures and always propagate it.
- Wrap errors: `fmt.Errorf("context: %w", err)`.
- Use structured logging (`slog.Info("msg", "key", val)`).
- Prefer sqlc queries over raw SQL.
- Keep changes scoped; avoid unrelated refactors.

## Agent workflow (jj workspaces, unified merge, location-aware)

- This repo uses Jujutsu (jj), there is one main workspace called `main` exist at original repo
- We use jj workspaces for parallel work
- Agents MUST NOT edit in `main` and work only in their own workspace
- Each agent workspace is temporary and deleted after merge

- The project root is the directory that contains `.jj/` don't run jj commands if `.jj` doesnt exist
- NEVER assume the current directory is the project root; We may operate either inside the project root or its parent

- Agent workspaces live in the parent of the project root and MUST NOT be created inside the project root
- Workspace naming is `jjws-<reponame>-<task>`
- Create with `jj workspace add ../jjws-<reponame>-<task>` after creation confirm with `jj workspace list`

- In an agent workspace, all work for a task MUST be in a single jj change
- If multiple changes exist, squash/fold into ONE final change with clear description before merge

- When finished, the agent requests merge back to `main` as ONE unified change

- `main` MUST end up with exactly ONE new change for the task
- Agent workspaces are deleted only after successful merge and approved by admin
- Only delete directories matching `jjws-*`
- Use feat, fix, refactor, test, chore
