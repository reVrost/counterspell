# counterspell

A mobile-first, self-hosted AI agent orchestration system that allows parallel coding tasks ("fire and forget"). The system acts as a "Command & Control" center where humans define intent, and AI agents execute, review, and report back.

## Features

- **GitHub Integration**: Connect your personal account or organization (similar to Vercel)
- **Mobile-First PWA**: Native-feeling mobile interface with tab navigation for phones, grid for desktop
- **Optimistic UI**: Drag-and-drop with instant feedback using HTMX and Alpine.js
- **Real-Time Logs**: SSE streaming of agent execution logs (the "Matrix Rain" console)
- **Git Worktrees**: Isolated agent workspaces for safe code execution
- **Task State Machine**: Rigid status flow (backlog → in_progress → review → done)
- **Offline Support**: Service worker caching for offline access

## Tech Stack

- **Backend**: Go 1.25+
- **Routing**: `go-chi/chi`
- **Database**: SQLite (WAL mode)
- **Frontend**: HTMX + Alpine.js + TailwindCSS
- **Drag & Drop**: SortableJS with haptic feedback
- **Real-time**: SSE (Server-Sent Events)
- **Agent Engine**: Orion (LLM orchestration)

## Project Structure

```
counterspell/
├── cmd/app/           # Application entry point
├── internal/
│   ├── db/           # SQLite database layer
│   ├── git/          # Git worktree management
│   ├── handlers/      # HTTP handlers
│   ├── models/        # Data models (tasks, projects, github connections)
│   ├── services/      # Business logic (tasks, agents, events)
│   └── ui/           # HTML templates (templ)
├── web/static/       # Static assets, PWA files
└── pkg/orion/       # Existing agent brain library
```

## Getting Started

### Prerequisites

- Go 1.25+
- Git (for worktree functionality)

### Installation

```bash
# Clone repository
git clone https://github.com/revrost/code/counterspell.git
cd counterspell

# Build
go build -o counterspell ./cmd/app

# Run
./counterspell -addr :8710 -db ./data/counterspell.db
```

### Docker

```bash
docker build -t counterspell .
docker run -p 8710:8710 -v $(pwd)/data:/app/data counterspell
```

## Usage

### 1. Connect GitHub

Access the app at **http://localhost:8710** and connect your GitHub account or organization. This grants counterspell access to your repositories.

### 2. Select a Project

After connecting, you'll see a list of your available repositories. Select one to start creating tasks.

### 3. Create Tasks

- Click "+ New Task" button
- Enter a title (e.g., "Fix login bug")
- Describe the intent for the agent

### 4. Manage Tasks

**Desktop**: Drag tasks between columns:
- **Backlog** → **In Progress** → Triggers agent execution
- **Review** → **Done** → Approves and merges changes

**Mobile**: 
- Tap tabs to switch between status columns
- Drag tasks within columns to reorganize
- Moving to "In Progress" triggers agent

## Agent Workflow

1. **Planning**: Agent analyzes the task and creates a plan
2. **Worktree Creation**: Isolated git worktree is created
3. **Execution**: Agent writes code in the sandbox
4. **Review**: Task moves to review status with diffs
5. **Approval**: Human approves or rejects the changes

## PWA Installation

1. Access the app via HTTPS or localhost
2. Tap "Share" → "Add to Home Screen" (iOS)
3. Tap "Install App" (Android)
4. The app runs in standalone mode (no browser chrome)

## Development

```bash
# Run with live reload
make dev

# Run tests
go test ./...

# Generate templates
templ generate
```

## Mobile Testing

Test on iOS Safari or Chrome DevTools Device Mode:

```bash
# Use ngrok for tunneling
ngrok http 8710

# Then access https://<ngrok-id>.ngrok.io from your phone
```

## GitHub Integration

counterspell supports two connection types:

1. **Personal Account**: Connect your personal GitHub repositories
2. **Organization**: Connect your entire organization (requires admin approval)

Both provide read/write access to:
- Repository code
- Issues
- Pull requests

The system creates isolated worktrees for each task, keeping your main branch safe.

## License

FSL-1.1-MIT (Functional Source License with MIT Future Grant)

This software is licensed under the FSL-1.1-MIT license, which permits:
- Internal use and access
- Non-commercial education
- Non-commercial research
- Professional services with licensed clients

On the second anniversary of release (2028-01-05), the license automatically converts to standard MIT.

See the [LICENSE](LICENSE) file for full terms.
