# Counterspell

A task orchestration platform for running AI agents on your codebases. Fire-and-forget task execution with real-time feedback and secure sandboxed code execution.

> **Philosophy:** "Persistent Brain (Go), Ephemeral Hands (Bubblewrap), Shared Memory (PostgreSQL)"

## Features

- **Self-Hostable**: Single PostgreSQL database, optional Supabase auth for multi-user
- **Single-User Mode**: Run locally without any auth - just set `DATABASE_URL`
- **Bubblewrap Sandboxing**: Secure code execution with Linux kernel namespaces (<5ms startup)
- **GitHub Integration**: OAuth-based repo access, shared bare repos, user-isolated worktrees
- **Mobile-First PWA**: Native-feeling interface with offline support
- **Real-Time Logs**: SSE streaming of agent execution (the "Matrix Rain" console)
- **Task State Machine**: Backlog → In Progress → Review → Done

## Architecture

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

### User Isolation

All users share one PostgreSQL database. User isolation is achieved via `user_id` columns:

```sql
-- Every user-scoped table has user_id
SELECT * FROM tasks WHERE user_id = $1 AND id = $2;
```

### Directory Structure

```
data/
├── repos/                  # Shared bare git repos (deduplicated)
│   └── {owner}/{repo}.git
└── workspaces/
    └── {user_id}/
        └── worktrees/
            └── {repo}_{task_id}/   # Isolated agent workspace
```

## Tech Stack

| Component | Technology |
|-----------|------------|
| Language | Go 1.23+ |
| Frontend | SvelteKit (embedded SPA) |
| Database | PostgreSQL (single shared DB) |
| Auth | Supabase (optional - for multi-user) |
| Isolation | Bubblewrap (bwrap) |
| Agent Engine | Native Go + Claude Code backends |

## Getting Started

### Prerequisites

- Go 1.23+
- PostgreSQL 14+
- Node.js 20+ (for building frontend)
- Git
- Bubblewrap (`bwrap`) - for sandboxed execution (Linux only)

### Installation

```bash
git clone https://github.com/revrost/counterspell.git
cd counterspell

# Build frontend
cd ui && npm install && npm run build && cd ..

# Build backend
go build -o counterspell ./cmd/app

# Run
./counterspell
```

### Environment Variables

```bash
# Required
DATABASE_URL=postgres://user:pass@localhost:5432/counterspell?sslmode=disable

# Supabase Auth (required for multi-user OAuth)
SUPABASE_URL=https://xxx.supabase.co
SUPABASE_ANON_KEY=eyJ...
SUPABASE_JWT_SECRET=your-jwt-secret

# Frontend Supabase config (required for OAuth)
VITE_SUPABASE_URL=https://xxx.supabase.co
VITE_SUPABASE_ANON_KEY=eyJ...

# GitHub OAuth (for repo access)
GITHUB_CLIENT_ID=xxx
GITHUB_CLIENT_SECRET=xxx
GITHUB_REDIRECT_URI=http://localhost:8710/github/callback

# Optional: AI provider keys
OPENROUTER_API_KEY=sk-or-...
ANTHROPIC_API_KEY=sk-ant-...

# Optional: Server config
PORT=8710
LOG_LEVEL=info
DATA_DIR=./data
```

### Quick Start with Docker

```bash
# Start PostgreSQL
docker run -d --name counterspell-db \
  -p 5432:5432 \
  -e POSTGRES_DB=counterspell \
  -e POSTGRES_PASSWORD=dev \
  postgres:16

# Run app
export DATABASE_URL="postgres://postgres:dev@localhost:5432/counterspell?sslmode=disable"
./counterspell
```

## Usage

### 1. Connect GitHub

Access **http://localhost:8710** and authenticate with GitHub OAuth. This grants Counterspell access to your repositories.

### 2. Select a Project

Choose a repository from your GitHub account to start creating tasks.

### 3. Create & Execute Tasks

- Create a task with a title and intent description
- Drag to "In Progress" to trigger the AI agent
- Watch real-time logs as the agent works
- Review changes and approve/reject

### Agent Workflow

1. **Planning**: Agent analyzes task and creates execution plan
2. **Worktree Creation**: Isolated git worktree in user's workspace
3. **Sandboxed Execution**: Code runs inside Bubblewrap container
4. **Review**: Task moves to review with diffs
5. **Approval**: Human approves or rejects changes

## Self-Hosting Modes

### Single-User Mode (No Supabase)

Just set `DATABASE_URL` and GitHub OAuth - no Supabase needed:

```bash
export DATABASE_URL="postgres://postgres:dev@localhost:5432/counterspell?sslmode=disable"
export GITHUB_CLIENT_ID="..."
export GITHUB_CLIENT_SECRET="..."
./counterspell
```

Without Supabase env vars, all requests use `userID = "default"`.

### Multi-User Mode (With Supabase)

Add Supabase env vars for JWT-based auth with per-user isolation:

```bash
export SUPABASE_URL="https://xxx.supabase.co"
export SUPABASE_ANON_KEY="eyJ..."
export SUPABASE_JWT_SECRET="..."
```

## Deployment

### Single Server (Recommended)

Deploy to a CPU-optimized server (DigitalOcean, EC2, etc.):

```bash
# Fly.io
flyctl deploy

# Or any VPS with Docker
docker-compose up -d
```

**Deployment docs:**
- [FLY_DEPLOYMENT.md](FLY_DEPLOYMENT.md) - Fly.io setup
- [FLY_CHECKLIST.md](FLY_CHECKLIST.md) - Pre-deploy checklist

### Authentication Setup

See [SUPABASE_SETUP.md](SUPABASE_SETUP.md) for Supabase configuration.

**Note:** Only GitHub OAuth is supported - GitHub IS the login method since this is a GitHub-centric tool.

## Development

```bash
# Build frontend (required after Svelte changes)
cd ui && npm run build

# Generate sqlc (after .sql changes)
sqlc generate

# Run tests
go test ./...

# Live reload
make dev
```

## Security

- **Sandbox Escape Prevention**: All user file/shell operations go through Bubblewrap
- **Path Traversal**: Bind mounts prevent access outside user's workspace
- **Resource Limits**: 10min timeout, 1MB output per execution
- **Token Security**: GitHub tokens stored encrypted in PostgreSQL
- **Cross-User Isolation**: Row-level isolation via `user_id`, separate workspace directories

## License

FSL-1.1-MIT (Functional Source License)

- Internal use, non-commercial education/research permitted
- Converts to MIT on January 5, 2028

See [LICENSE](LICENSE) for full terms.
