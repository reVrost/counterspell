# Counterspell

A multi-tenant AI agent orchestration platform for parallel coding tasks. Fire-and-forget task execution with real-time feedback, sandboxed code execution, and per-user isolation.

> **Philosophy:** "Persistent Brain (Go), Ephemeral Hands (Bubblewrap), Segmented Memory (SQLite)"

## Features

- **Multi-Tenant SaaS**: Per-user SQLite databases, isolated workspaces, Supabase authentication
- **Single-Player Mode**: Self-host with `MULTI_TENANT=false` for personal use
- **Bubblewrap Sandboxing**: Secure code execution with Linux kernel namespaces (<5ms startup)
- **GitHub Integration**: OAuth-based repo access, shared bare repos, user-isolated worktrees
- **Mobile-First PWA**: Native-feeling interface with offline support
- **Real-Time Logs**: SSE streaming of agent execution (the "Matrix Rain" console)
- **Task State Machine**: Backlog → In Progress → Review → Done

## Architecture

```
┌─────────────────┐     ┌──────────────┐     ┌─────────────────┐
│  Browser/PWA    │────▶│   Supabase   │────▶│  GitHub OAuth   │
└─────────────────┘     └──────────────┘     └─────────────────┘
         │                                           │
         └──────────────┬────────────────────────────┘
                        ▼
┌─────────────────────────────────────────────────────────────┐
│                     Go Server                                │
│  ┌─────────────┐   ┌─────────────┐   ┌──────────────────┐   │
│  │ Auth Layer  │──▶│ User Manager│──▶│ Worker Pool (20) │   │
│  │ (JWT/OAuth) │   │ (per-user)  │   │ (FIFO scheduling)│   │
│  └─────────────┘   └─────────────┘   └──────────────────┘   │
│         │                │                     │             │
│         ▼                ▼                     ▼             │
│  ┌─────────────────────────────────────────────────────┐    │
│  │              Bubblewrap Sandbox                      │    │
│  │  • Isolated filesystem (user workspace only)         │    │
│  │  • 10min timeout, 1MB output limit                   │    │
│  │  • Full network (npm, go get, etc.)                  │    │
│  └─────────────────────────────────────────────────────┘    │
└─────────────────────────────────────────────────────────────┘
```

### Directory Structure

```
data/
├── db/
│   ├── {user_uuid}.db      # Per-user SQLite (history, settings, logs)
│   └── default.db          # Single-player mode
├── repos/                  # Shared bare git repos (deduplicated)
│   └── {owner}/{repo}.git
└── workspaces/
    └── {user_uuid}/
        └── worktrees/
            └── {repo}_{task_id}/   # Isolated agent workspace
```

## Tech Stack

| Component | Technology |
|-----------|------------|
| Language | Go 1.25+ |
| Frontend | HTMX + Alpine.js + TailwindCSS |
| Templates | templ (type-safe) |
| Database | SQLite (WAL mode, per-user) |
| Auth | Supabase (GitHub OAuth) |
| Isolation | Bubblewrap (bwrap) |
| Agent Engine | Orion (LLM orchestration) |

## Getting Started

### Prerequisites

- Go 1.25+
- Git
- Bubblewrap (`bwrap`) - for sandboxed execution

### Installation

```bash
git clone https://github.com/revrost/counterspell.git
cd counterspell

go build -o counterspell ./cmd/app
./counterspell
```

### Environment Variables

```bash
# Mode (default: single-player)
MULTI_TENANT=false          # Set to true for SaaS mode

# Supabase (required when MULTI_TENANT=true)
SUPABASE_URL=https://xxx.supabase.co
SUPABASE_ANON_KEY=eyJ...
SUPABASE_JWT_SECRET=your-jwt-secret

# Worker pool
WORKER_POOL_SIZE=20         # Concurrent workers
MAX_TASKS_PER_USER=5        # Per-user task limit
USER_MANAGER_TTL=2h         # Inactive user cleanup

# Sandbox
SANDBOX_TIMEOUT=600         # 10 minutes
SANDBOX_OUTPUT_LIMIT=1048576  # 1MB
NATIVE_ALLOWLIST=git,ls,cat,head,tail,grep,find,wc,sort,uniq
```

### Docker

```bash
docker build -t counterspell .
docker run -p 8710:8710 -v $(pwd)/data:/app/data counterspell
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
# Live reload
make dev

# Run tests
go test ./...

# Generate templates
templ generate

# Generate sqlc
sqlc generate
```

## Documentation

| Document | Description |
|----------|-------------|
| [docs/MULTI_TENANT_DESIGN.md](docs/MULTI_TENANT_DESIGN.md) | Multi-tenant architecture design |
| [docs/STACK_GUIDE.md](docs/STACK_GUIDE.md) | templ + HTMX + Alpine.js guide |
| [gobox.md](gobox.md) | Original architecture vision |

## Security

- **Sandbox Escape Prevention**: All user file/shell operations go through Bubblewrap
- **Path Traversal**: Bind mounts prevent access outside user's workspace
- **Resource Limits**: 10min timeout, 1MB output per execution
- **Token Security**: GitHub tokens in Supabase vault + user SQLite
- **Cross-User Isolation**: Separate SQLite files, separate workspace directories

## License

FSL-1.1-MIT (Functional Source License)

- Internal use, non-commercial education/research permitted
- Converts to MIT on January 5, 2028

See [LICENSE](LICENSE) for full terms.
