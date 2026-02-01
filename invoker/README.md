# Invoker - Counterspell Control Plane

The control plane service for Counterspell Multi-Tenant SaaS, handling authentication, billing, and VM provisioning.

## Architecture

Invoker is the central management point that:
- Authenticates users via Supabase
- Manages user subscriptions (Stripe)
- Provisions Fly.io VMs dynamically
- Maintains routing table for subdomains
- Tracks machine health and status

## Tech Stack

- **Language**: Go
- **HTTP Router**: Chi (go-chi/chi/v5)
- **UI**: Svelte 5 with SvelteKit
- **Database**: Supabase PostgreSQL (via pgx/v5)
- **Auth**: Supabase JWT validation
- **VM Provider**: Fly.io API

## Project Structure

```
/invoker
├── cmd/
│   └── invoker/
│       └── main.go           # Application entry point
├── ui/                     # Svelte 5 UI
│   ├── src/
│   │   ├── routes/           # SvelteKit routes
│   │   └── lib/
│   ├── build/               # Production build output
│   ├── embed.go             # Go embed for UI files
│   └── package.json
├── internal/
│   ├── auth/                 # Supabase JWT validation and handlers
│   │   ├── supabase.go       # JWT validation logic
│   │   └── handler.go        # Auth HTTP handlers (register, login)
│   ├── config/               # Configuration management
│   │   └── config.go
│   ├── db/                   # Database operations
│   │   ├── pool.go           # Connection pool
│   │   ├── db.go             # Generated SQLc code
│   │   ├── service.go        # DB service interface
│   │   └── converter.go      # Model converters
│   ├── fly/                  # Fly.io API client (TODO)
│   └── billing/              # Stripe integration (TODO)
├── pkg/
│   └── models/               # Shared data models
│       └── models.go
├── schema.sql                # Database schema
├── .env.example              # Environment variables template
├── go.mod
└── go.sum
```

## Setup

### Prerequisites
- Go 1.25+
- Node.js 18+ (for UI development)
- Supabase account
- Fly.io account

### Environment Variables

Copy `.env.example` to `.env` and configure:

```bash
# Supabase Configuration
SUPABASE_URL=https://your-project.supabase.co
SUPABASE_ANON_KEY=your-anon-key
SUPABASE_SERVICE_ROLE_KEY=your-service-role-key

# Server Configuration
PORT=8080
APP_VERSION=0.1.0
ENVIRONMENT=development

# Fly.io Configuration
FLY_API_TOKEN=your-fly-api-token
FLY_ORG=your-fly-org
FLY_REGION=iad

# Database Configuration
DATABASE_URL=your-database-url

# Security
JWT_SECRET=your-jwt-secret-key
```

### Database Setup

Initialize the database schema:

```bash
psql $DATABASE_URL < schema.sql
```

### Build and Run

#### Quick Start
```bash
# Build UI and Go application
make build

# Run
./bin/invoker
```

#### Development

**Backend**:
```bash
go run ./cmd/invoker
```

**UI** (separate development):
```bash
cd ui
npm install
npm run dev
```

**Full Build**:
```bash
# Build UI for production
make ui-build

# Build Go application
make build

# Run
./bin/invoker
```

### Available Make Targets

```bash
make ui-dev         # Run Svelte UI in development mode
make ui-build       # Build Svelte UI for production
make ui-clean       # Clean UI build artifacts
make ui-deps        # Install UI dependencies
make build          # Build UI and Go application
make run            # Build and run
make dev            # Run with hot reload (requires 'air')
```

## API Endpoints

### Health Checks
- `GET /health` - Health check (no auth required)
- `GET /ready` - Readiness check (checks DB connection)

### Authentication
- `POST /api/auth/register` - Register new user
  - Body: `{ "email": "...", "first_name": "...", "last_name": "...", "password": "..." }`
  - Response: `{ "token": "...", "user": { ... } }`

- `POST /api/auth/login` - Login user
  - Body: `{ "email": "...", "password": "..." }`
  - Response: `{ "token": "...", "user": { ... } }`

### VM Management (TODO - Task 1.3)
- `POST /api/vm/start` - Start/resume user's VM
- `GET /api/vm/status` - Get VM status
- `DELETE /api/vm/stop` - Stop user's VM

### Machine Registry (TODO - Task 1.4)
- `GET /api/machines` - List user's VMs
- `GET /api/machines/:id` - Get VM details

### Routing (TODO - Task 1.5)
- `GET /api/routing/:subdomain` - Get VM URL for subdomain

## Development

### Adding Dependencies

```bash
go get <package>
go mod tidy
```

### Running Tests

```bash
go test ./...
```

### Building for Production

```bash
go build -ldflags="-s -w" -o bin/invoker ./cmd/invoker
```

## Current Status

**Completed (Tasks 1.1-1.2)**:
- ✅ Project structure and Go module
- ✅ Database schema (users, subscriptions, machine_registry, routing_table, etc.)
- ✅ Basic HTTP server with Chi router
- ✅ Supabase JWT validation (public key only, no DB calls)
- ✅ Auth endpoints (register, login)
- ✅ User database operations
- ✅ Health and readiness endpoints

**In Progress**:
- ⏳ Database schema conversion (SQLite → PostgreSQL syntax)

**TODO**:
- ⏳ Task 1.3: Fly.io API integration
- ⏳ Task 1.4: Machine registry & health monitoring
- ⏳ Task 1.5: Dynamic subdomain routing table
- ⏳ Task 1.6: Cloudflare Worker deployment
- ⏳ Supabase Auth REST API integration for login

## Known Issues

1. **Database Schema**: Uses SQLite syntax, needs conversion to PostgreSQL for Supabase
2. **Login Handler**: Returns placeholder token, needs Supabase Auth API integration
3. **JWT Generation**: Returns placeholder token, needs proper JWT generation

## Contributing

See WORK.md for detailed notes on implementation and next steps.
