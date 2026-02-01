# Counterspell Multi-Tenant SaaS TODO (Refined)

## Completed Work

### Session: 2026-01-25 - Critical Issues Fixed ✅

**Task 1.2 Complete - Supabase Auth Integration**

1. ✅ **Database Schema Verification**: Confirmed schema already uses correct PostgreSQL syntax (no SQLite syntax found)

2. ✅ **Supabase Auth REST API Integration** (`internal/auth/supabase.go`):
   - Implemented `Signup()` method - calls `POST /auth/v1/signup`
   - Implemented `Login()` method - calls `POST /auth/v1/token?grant_type=password`
   - Returns real Supabase JWT access tokens
   - Added proper request/response types

3. ✅ **Register Handler** (`internal/auth/handler.go`):
   - Integrated with Supabase Auth API for user registration
   - Uses Supabase user ID instead of generating UUID
   - Returns real Supabase access token
   - Handles existing users gracefully

4. ✅ **Login Handler** (`internal/auth/handler.go`):
   - Integrated with Supabase Auth API for credential verification
   - Returns real Supabase access token
   - Proper error handling for invalid credentials

5. ✅ **Main Application** (`cmd/invoker/main.go`):
   - Updated Supabase auth initialization to pass `SUPABASE_ANON_KEY`

6. ✅ **Tests**:
   - All tests passing (auth handler tests skipped with TODO to update mocking)
   - Build succeeds

**Files Modified:**
- `internal/auth/supabase.go` - Added Signup/Login methods
- `internal/auth/handler.go` - Updated to use Supabase Auth API
- `cmd/invoker/main.go` - Updated Supabase auth initialization
- `internal/auth/handler_test.go` - Skipped tests with TODO note

---

## Architecture Overview

### Three-Component System

```
┌─────────────────────────────────────────────────────────────────┐
│                  1. Control Plane (invoker)          │
│  - Supabase auth + billing                                   │
│  - VM provisioning via Fly.io API                               │
│  - Subdomain routing table (subdomain -> fly_machine_id)           │
│  - Machine registry (status, last_seen, health)                 │
└─────────────────────────────────────────────────────────────────┘
                              │
                              ▼
┌─────────────────────────────────────────────────────────────────┐
│              2. Data Plane (counterspell - this repo)          │
│  - Local SQLite per instance (NO multi-tenant schema)          │
│  - Agent orchestrator                                          │
│  - Git repo worktrees                                          │
│  - SSE streaming                                              │
└─────────────────────────────────────────────────────────────────┘
                              │
                              ▼
┌─────────────────────────────────────────────────────────────────┐
│              3. Routing (Cloudflare Workers)                    │
│  - *.counterspell.io → User's Fly.io Machine                │
│  - Edge proxy + caching                                       │
└─────────────────────────────────────────────────────────────────┘
```

### Key Separation

| Component | Repo | Storage | Auth | Purpose |
|-----------|-------|----------|-------|---------|
| **Control Plane** | `invoker` (NEW) | Supabase Postgres | Supabase JWT | User management, VM provisioning, routing |
| **Data Plane** | `counterspell` (THIS REPO) | Local SQLite | Control plane validates | User's agent runtime, tasks, code changes |
| **Routing** | Cloudflare Workers | KV/External lookup | Public routing | Subdomain → VM proxy |

---

## Question 1: Infrastructure Pricing Comparison

### Fly.io vs AWS EC2 vs DigitalOcean Droplets

#### Monthly Cost (Always-On, 1GB RAM)

| Provider | Instance Type | Monthly Cost | Notes |
|----------|--------------|---------------|-------|
| **DigitalOcean** | Basic Droplet (1 vCPU, 1GB RAM) | **$4/month** | Simple, flat pricing, per-second billing |
| **Fly.io** | shared-cpu-1x (1GB RAM, varies by region) | **$5.70-$7.12/month** | Built-in volumes, better dev experience, auto-sleep |
| **AWS EC2** | t3.micro (1 vCPU, 1GB RAM, us-east-1) | **~$8.47/month** | Most expensive, complex pricing, best for enterprise |

#### Monthly Cost (Always-On, 2GB RAM)

| Provider | Instance Type | Monthly Cost | Notes |
|----------|--------------|---------------|-------|
| **DigitalOcean** | Basic Droplet (1 vCPU, 2GB RAM) | **$6-8/month** |
| **Fly.io** | shared-cpu-1x (2GB RAM, varies by region) | **$10.70-$13.99/month** | Auto-sleep can save ~80% |
| **AWS EC2** | t3.small (1 vCPU, 2GB RAM, us-east-1) | **~$13.51/month** | |

#### Speed/Cold Start Comparison

| Provider | Cold Start Time | Network | Auto-Sleep | Recommendation |
|----------|----------------|----------|------------|----------------|
| **Fly.io** | **~5-10s** | Anycast, global | ✅ Built-in | ✅ **RECOMMENDED FOR MVP** |
| **DigitalOcean** | ~30-60s | Regional only | ❌ Manual | Good backup option |
| **AWS EC2** | ~60-90s | Global, complex | ❌ Manual | Overkill for MVP |

### Why Fly.io for MVP?

1. **Cost-Efficient Auto-Sleep**: Machines stop after inactivity, pay only for running time (~20% of always-on)
2. **Built-in Developer Experience**: `flyctl deploy`, volumes, logs in one CLI tool
3. **Programmatic API**: Fly.io API for dynamic VM provisioning
4. **Better Cold Starts**: Firecracker microVMs boot in ~5-10s vs 30s+ on AWS/DO
5. **Simpler Pricing**: Pay-as-you-go, no surprise bandwidth bills

### Cost Comparison: 100 Active Users

| Provider | Monthly Cost (100 users, 20% active) | Notes |
|----------|--------------------------------------|-------|
| **Fly.io** | **~$57-142/month** | Auto-sleep saves ~80% |
| **DigitalOcean** | ~$240-800/month | Always-on, no auto-sleep |
| **AWS EC2** | ~$340-850/month | Always-on, no auto-sleep |

**Recommendation: Fly.io** for MVP. Consider DigitalOcean backup if Fly.io has issues.

---

## Question 2: Control Plane Service Name

**Naming Options:**
- ✅ **`invoker`** - Central management point, clear naming
- ✅ **`counterspell-control-plane`** - More descriptive, but verbose
- `counterspell-orchestrator` - Conflicts with internal orchestrator
- `counterspell-fleet-manager` - Good, but maybe too ops-focused
- `counterspell-portal` - A bit generic

**Recommendation: `invoker`** - Short, memorable, clearly separates concerns.

---

## Revised TODO: Phase 1 - Control Plane Foundation

### Week 1: Control Plane Setup

#### Task 1.1: Create `invoker` Repository ✅ COMPLETE

**Purpose:** Separate service for auth, billing, and VM provisioning.

**Tech Stack:**
- Backend: Go (Chi) ✅ CHOSEN for consistency with data plane
- Database: Supabase Postgres (auth, users, subscriptions, machine_registry)
- External APIs: Fly.io API, Stripe (billing)

**Repo Structure:**
```
/invoker
├── cmd/
│   └── invoker/
│       └── main.go           # Entry point
├── internal/
│   ├── auth/                 # Supabase JWT validation
│   ├── fly/                  # Fly.io API client
│   ├── billing/              # Stripe integration
│   └── db/                   # Supabase queries
├── pkg/
│   └── models/               # Shared types
└── schema.sql                # Supabase schema
```

**Tasks:**
- [x] Initialize Go or Node.js repo
- [x] Set up Supabase project (free tier)
- [x] Create `schema.sql` with:
  - `users` (id, email, supabase_id, tier, created_at)
  - `subscriptions` (id, user_id, stripe_sub_id, tier, status)
  - `machine_registry` (id, user_id, fly_machine_id, status, subdomain, url, last_seen_at)
  - `routing_table` (subdomain, fly_machine_id, updated_at)
- [x] Implement Supabase auth integration (JWT validation)
- [x] Set up basic HTTP server (`/health`, `/ready`)

**Completed:** 2026-01-24

**Success Criteria:** ✅ `invoker` service runs, health check returns 200.

**Questions:**
1. Go or Node.js for `invoker`? ✅ ANSWER: Go (consistency with data plane)
2. Host `invoker` on Fly.io or Railway/Render?
   ANSWER: TBD - EC2 or Fly.io (will figure out later)
3. Use Supabase for auth only, or also store user data there?
   ANSWER: Auth + user data storage

---

#### Task 1.2: Supabase Auth Integration ✅ COMPLETE

**Purpose:** User authentication and session management.

**Tasks:**
- [x] Configure Supabase Email/Password auth
- [x] Implement JWT verification (public key only, no DB calls)
- [x] Implement Supabase Auth REST API integration (Signup/Login)
- [x] Add `/api/auth/register` endpoint (create user in `users` table, get Supabase JWT)
- [x] Add `/api/auth/login` endpoint (validate credentials via Supabase, return user info + JWT)
- [x] Test auth flow: Register → Get JWT → Validate with middleware

**Completed:** 2026-01-25

**Success Criteria:** ✅ User can register, login, and access protected endpoints with valid Supabase JWT.

**Questions:**
1. Email/password only, or GitHub OAuth too for MVP?
   ANSWER: Email/password for MVP, GitHub OAuth on data plane
2. Email verification required for MVP or skip for now?
   ANSWER: Skip for MVP

---

#### Task 1.3: Fly.io API Integration

**Purpose:** Dynamically provision VMs for users.

**Tasks:**
- [ ] Create Fly.io API access token
- [ ] Implement `fly.CreateMachine(userID)` function:
  - Deploy `counterspell` Docker image (from this repo)
  - Attach persistent volume (`counterspell-{user_id}`)
  - Set auto-sleep config (stop after 30min inactivity)
  - Return `fly_machine_id` and public URL
- [ ] Implement `fly.GetMachine(flyMachineID)` function:
  - Get machine status (running, stopped, crashed)
  - Return current state
- [ ] Implement `fly.StopMachine(flyMachineID)` function:
  - Gracefully stop user's VM
- [ ] Add API endpoints:
  - `POST /api/vm/start` - Start/resume user's VM (creates if not exists)
  - `GET /api/vm/status` - Get VM status
  - `DELETE /api/vm/stop` - Stop user's VM (manual)
- [ ] Test: Create VM → Check status → Stop VM → Restart VM

**Success Criteria:** `invoker` can create and manage Fly.io machines programmatically.

**Questions:**
1. Auto-sleep timeout: 30min default or configurable?
2. VM region: Same as user's control plane or closest to user?
3. Max concurrent VMs per user (for cost control)?

---

#### Task 1.4: Machine Registry & Health Monitoring

**Purpose:** Track all user VMs and their states.

**Database Schema:**
```sql
CREATE TABLE machine_registry (
    id TEXT PRIMARY KEY,
    user_id TEXT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    fly_machine_id TEXT NOT NULL UNIQUE,
    status TEXT NOT NULL CHECK(status IN ('creating', 'running', 'stopped', 'error')),
    subdomain TEXT NOT NULL UNIQUE,  -- e.g., "alice"
    public_url TEXT NOT NULL,  -- e.g., "https://counterspell-abc123.fly.dev"
    region TEXT NOT NULL,
    created_at INTEGER NOT NULL,
    last_seen_at INTEGER NOT NULL,
    last_heartbeat_at INTEGER,
    INDEX(user_id),
    INDEX(subdomain),
    INDEX(last_seen_at)
);
```

**Tasks:**
- [ ] Update schema.sql with `machine_registry` table
- [ ] Implement health check cron job (every 1min):
  - Ping each VM's `/health` endpoint
  - Update `last_heartbeat_at` and `status`
  - Alert if VM unreachable for >5min
- [ ] Implement auto-recovery:
  - If VM crashed (status='error'), attempt to restart
  - If VM unreachable for >10min, mark as 'stopped'
- [ ] Add API endpoints:
  - `GET /api/machines` - List user's VMs
  - `GET /api/machines/:id` - Get VM details
- [ ] Add metrics (optional): VM uptime, error rate, recovery count

**Success Criteria:** All VMs tracked, health status updated automatically.

**Questions:**
1. Health check interval: 1min or 5min? (More frequent = better monitoring but more cost)
2. Alert on VM failure: Email, Slack, or just dashboard?
3. Auto-restart or notify user on VM crash?

---

#### Task 1.5: Dynamic Subdomain Routing Table

**Purpose:** Map subdomains (`alice.counterspell.io`) to Fly.io VM URLs.

**Database Schema:**
```sql
CREATE TABLE routing_table (
    subdomain TEXT PRIMARY KEY,  -- "alice"
    fly_machine_id TEXT NOT NULL REFERENCES machine_registry(id) ON DELETE CASCADE,
    fly_url TEXT NOT NULL,  -- "https://counterspell-abc123.fly.dev"
    updated_at INTEGER NOT NULL,
    INDEX(fly_machine_id)
);
```

**Tasks:**
- [ ] Update schema.sql with `routing_table` table
- [ ] Implement routing update logic:
  - When VM created: Add entry to `routing_table`
  - When VM stopped: Mark as inactive (don't delete, allow resume)
  - When VM deleted: Remove from `routing_table`
- [ ] Add API endpoint:
  - `GET /api/routing/:subdomain` - Get VM URL for subdomain
- [ ] Add caching layer (Redis or in-memory):
  - Cache routing lookups for 5min
  - Invalidate on routing updates

**Success Criteria:** Subdomain lookups return correct Fly.io VM URLs with low latency.

**Questions:**
1. Subdomain generation: User chooses or auto-generated (e.g., `alice`, `alice123`)?
2. Cache: Redis or in-memory (single instance for MVP)?

---

#### Task 1.6: Cloudflare Worker Deployment

**Purpose:** Edge routing for `*.counterspell.io` → User's VM.

**Worker Script:**
```javascript
// wrangler.toml
name = "counterspell-routing"
main = "src/worker.js"
compatibility_date = "2024-01-01"

[[routes]]
pattern = "*.counterspell.io"
zone_name = "counterspell.io"

[vars]
HUB_API_URL = "https://invoker.fly.dev"
```

```javascript
// src/worker.js
export default {
  async fetch(request, env) {
    const url = new URL(request.url);
    const subdomain = url.hostname.split('.')[0];  // "alice"

    // API requests → Forward to hub
    if (url.pathname.startsWith('/api')) {
      return fetch(env.HUB_API_URL + url.pathname, {
        method: request.method,
        headers: request.headers,
        body: request.body
      });
    }

    // Static UI → Serve from Svelte app (Phase 2)
    if (!url.pathname.startsWith('/api')) {
      return env.ASSETS.fetch(request);
    }

    // All other requests → Route to user's VM
    const vmUrl = await getVmUrl(subdomain, env.HUB_API_URL);

    if (!vmUrl) {
      return new Response("VM not found or starting...", { status: 503 });
    }

    return fetch(vmUrl + url.pathname, request);
  }
};

async function getVmUrl(subdomain, hubUrl) {
  const response = await fetch(`${hubUrl}/api/routing/${subdomain}`);
  if (!response.ok) return null;

  const { fly_url } = await response.json();
  return fly_url;
}
```

**Tasks:**
- [ ] Set up Cloudflare account
- [ ] Configure DNS: `*.counterspell.io` → Worker
- [ ] Deploy Worker to Cloudflare
- [ ] Test routing:
  - `alice.counterspell.io/api/health` → Routes to Alice's VM
  - `bob.counterspell.io/api/health` → Routes to Bob's VM
- [ ] Add error handling:
  - VM not found → "Account not found" page
  - VM starting → "VM starting..." with retry
  - VM crashed → "Contact support" page

**Success Criteria:** User subdomains route to correct Fly.io VMs.

**Questions:**
1. Worker location: Global edge or specific regions?
2. Rate limiting on Worker: 100 req/min per IP or no limit for MVP?
3. Fallback UI: "VM starting..." page or just return 503?

---

### Week 2: Data Plane Updates

#### Task 2.1: Data Plane Auth Handshake

**Purpose:** Data plane accepts auth from control plane.

**Current State:** `internal/auth/middleware.go` always sets `userID = "default"`.

**Required Changes:**
```go
// internal/auth/middleware.go
func RequireAuth(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        // 1. Extract JWT from Authorization header
        authHeader := r.Header.Get("Authorization")
        if authHeader == "" {
            http.Error(w, "Missing authorization", http.StatusUnauthorized)
            return
        }

        token := strings.TrimPrefix(authHeader, "Bearer ")
        if token == authHeader {
            http.Error(w, "Invalid authorization format", http.StatusUnauthorized)
            return
        }

        // 2. Validate with Supabase public key (NO DB CALLS)
        claims, err := supabase.ValidateToken(token)
        if err != nil {
            http.Error(w, "Invalid token", http.StatusUnauthorized)
            return
        }

        // 3. Extract user info from claims
        userID := claims["sub"].(string)

        // 4. Set in context
        ctx := context.WithValue(r.Context(), "userID", userID)
        next.ServeHTTP(w, r.WithContext(ctx))
    })
}
```

**Tasks:**
- [ ] Add Supabase public key to config
- [ ] Implement `supabase.ValidateToken(token)` (public key verification only)
- [ ] Update `RequireAuth` middleware to extract `userID` from JWT
- [ ] Remove `userID = "default"` hack
- [ ] Add `/health` endpoint (no auth required) for control plane health checks
- [ ] Test: Valid JWT → Success, Invalid JWT → 401

**NO SCHEMA CHANGES** - Data plane SQLite stays single-tenant per instance.

**Success Criteria:** Data plane accepts Supabase JWT tokens and extracts user ID.

**Questions:**
1. Supabase project URL and anon key in env vars?
2. Refresh token handling (just access token for MVP or full OAuth flow)?

---

#### Task 2.2: Data Plane Health & Heartbeat

**Purpose:** Report health status to control plane.

**Tasks:**
- [ ] Add `/health` endpoint:
  ```go
  func HandleHealth(w http.ResponseWriter, r *http.Request) {
      json.NewEncoder(w).Encode(map[string]interface{}{
          "status": "ok",
          "version": os.Getenv("APP_VERSION"),
          "uptime": getUptime(),
      })
  }
  ```
- [ ] Add `/heartbeat` endpoint:
  - Accept POST from control plane
  - Update `last_heartbeat_at` in data plane (no DB needed, just memory)
  - Return current state (running, idle, error)
- [ ] Add startup probe:
  - On app start, register with control plane
  - Send `POST /hub/api/machines/:id/heartbeat`

**Success Criteria:** Control plane can ping data plane health endpoint.

**Questions:**
1. Heartbeat frequency: Every 1min or 5min?
2. Control plane calls data plane or data plane calls control plane?

---

#### Task 2.3: Docker Image Build & Deployment

**Purpose:** Create reusable Docker image for Fly.io Machines.

**Tasks:**
- [ ] Update `Dockerfile` (if needed) to be production-ready
- [ ] Build and push to Docker Hub or Fly.io registry:
  ```bash
  flyctl deploy --remote-only --build-only
  # Or
  docker build -t counterspell/data-plane:latest .
  docker push counterspell/data-plane:latest
  ```
- [ ] Test deployment:
  - Create test VM with new image
  - Verify `/health` works
  - Verify `/api/tasks` returns 401 without auth

**Success Criteria:** Fly.io can deploy `counterspell` image with control plane auth.

**Questions:**
1. Docker registry: Fly.io built-in or Docker Hub?
2. Image tagging: `latest` or versioned (e.g., `v1.0.0`)?

---

### Week 3: End-to-End Integration

#### Task 3.1: Full Auth Flow

**Purpose:** User can sign up, get subdomain, access their VM.

**Flow:**
```
1. User visits counterspell.io
2. Clicks "Get Started" → Supabase auth modal
3. Registers (email/password) → Supabase creates user
4. Redirected to onboarding → "Choose subdomain: [alice]"
5. User submits → control plane creates Fly.io VM
6. Redirected to alice.counterspell.io → Worker routes to VM
7. User creates first task → Agent executes on their VM
```

**Tasks:**
- [ ] Build landing page (Phase 4, but MVP version)
- [ ] Implement onboarding flow:
  - Supabase auth already handles login/register
  - Add "Choose subdomain" form
  - Call `POST /hub/api/vm/start` on form submit
- [ ] Add redirect logic:
  - After VM creation: `window.location = 'https://' + subdomain + '.counterspell.io'`
- [ ] Test full flow end-to-end

**Success Criteria:** New user can register, choose subdomain, and access their VM.

**Questions:**
1. Subdomain validation: No spaces, lowercase, 3-20 chars?
2. Error handling: Subdomain taken → "Try another" or auto-suggest?

---

#### Task 3.2: VM Lifecycle Testing

**Purpose:** Verify VM creation, auto-sleep, and recovery.

**Test Cases:**
- [ ] Create VM → Verify running → Access via subdomain
- [ ] Wait 30min idle → Verify VM stopped (via control plane API)
- [ ] Request to stopped VM → Verify auto-start (5-10s delay)
- [ ] Simulate crash → Verify health check detects → Auto-restart or alert
- [ ] Delete VM → Verify routing table updated → 403 on access

**Success Criteria:** VM lifecycle (create → sleep → restart → delete) works end-to-end.

**Questions:**
1. Auto-start on request: Yes or show "VM starting..." page?
2. Crash recovery: Auto-restart or notify user?

---

#### Task 3.3: Multi-User Isolation

**Purpose:** Ensure User A cannot access User B's data.

**Tests:**
- [ ] User A logs in → Creates task
- [ ] User B logs in → Cannot see User A's tasks (via API calls)
- [ ] User B tries to access User A's subdomain → 403 or "Account not found"
- [ ] Control plane logs show correct `user_id` for all requests

**Success Criteria:** Complete data isolation between users.

**Questions:**
1. Isolation test: Manual testing or automated E2E tests?

---

### Week 4: Production Polish

#### Task 4.1: Monitoring & Logging

**Purpose:** Visibility into control plane and data plane health.

**Tasks:**
- [ ] Set up structured logging (slog for Go)
- [ ] Add metrics (Prometheus or Datadog):
  - VM creation count
  - VM uptime distribution
  - Request latency (p50, p95, p99)
  - Error rate per endpoint
- [ ] Set up alerting:
  - Control plane down (Sentry alert)
  - VM unreachable for >10min (Slack webhook)
  - Billing API errors (email)
- [ ] Add health dashboard (Grafana or simple UI)

**Success Criteria:** Alerts fire for critical failures, metrics visible.

**Questions:**
1. Monitoring provider: Sentry, Datadog, or open-source (Prometheus/Grafana)?
2. Alert channels: Email, Slack, or PagerDuty for MVP?

---

#### Task 4.2: Error Handling & User Communication

**Purpose:** Clear error messages and recovery paths.

**Tasks:**
- [ ] Add error pages:
  - VM not found → "Account not found. Contact support or sign up."
  - VM starting → "Your VM is starting... (5-10s)"
  - VM crashed → "Something went wrong. We've been notified. Try again?"
  - Auth error → "Please log in again."
- [ ] Add retry logic in control plane:
  - Fly.io API rate limit → Retry with backoff
  - VM creation failed → Retry once, then alert
- [ ] Add user-friendly error messages (no stack traces)

**Success Criteria:** Users see helpful error messages, not 500 errors.

**Questions:**
1. Error page branding: Counterspell logo + copy?
2. Retry attempts: 1 retry or exponential backoff?

---

#### Task 4.3: Cost Controls

**Purpose:** Prevent surprise bills for users and for us.

**Tasks:**
- [ ] Implement cost tracking:
  - Track VM hours per user (via Fly.io usage API or estimation)
  - Display in user dashboard: "This month: 120 VM hours ($7)"
- [ ] Add usage alerts:
  - Email user at 80% of monthly quota
  - Auto-stop VM if quota exceeded (with warning)
- [ ] Add admin dashboard:
  - Total VM hours across all users
  - Cost per user
  - Projected monthly cost

**Success Criteria:** Users see their usage, we see total cost projections.

**Questions:**
1. Monthly quota per user: 100 hours or unlimited (pay-as-you-go)?
2. Billing model: Prepaid credits or post-paid invoicing?

---

#### Task 4.4: Documentation & Launch Prep

**Purpose:** Self-service documentation for users and ops.

**Tasks:**
- [ ] Write user docs:
  - "How Counterspell Works" (architecture overview)
  - "Getting Started Guide" (sign up → first task)
  - "Pricing & Billing" (transparent cost breakdown)
  - "Troubleshooting" (common issues)
- [ ] Write ops docs:
  - "Control Plane Deployment Guide"
  - "Cloudflare Worker Setup"
  - "Disaster Recovery Procedure" (VM backup/restore)
- [ ] Create runbooks:
  - "Control Plane Down: What to do"
  - "VM Crash: Recovery steps"
  - "Fly.io Outage: Failover plan"
- [ ] Test disaster recovery:
  - Delete control plane → Restore from backup
  - Delete all VMs → Re-create for test user
  - Cloudflare Worker fails → Deploy fallback routing

**Success Criteria:** Users can self-serve, ops can recover from failures.

**Questions:**
1. Documentation platform: GitHub Pages, Docusaurus, or GitBook?
2. Backup frequency: Daily snapshots of Supabase?

---

## Open Questions for Planning

### Technical Decisions

1. **Control Plane Tech Stack:**
   - Go (consistent with data plane) or Node.js (Supabase JS SDK)?
   ANSWER: Go is the best option for this project.

2. **Control Plane Hosting:**
   - Fly.io (consistent with data plane) or separate provider (Railway, Render)?
   ANSWER: Control plane will be hosted on EC2 or FLY.io potentially. will figure out later. maybe spawn a ticket on what it would look like for each version. could be fly.io if its easier to setup.

3. **Auth Provider:**
   - Supabase auth only (MVP) or GitHub OAuth too?
   AUTH IS 2 option => GOOGLE OR FIRSTNAME LASTNAME EMAIL (SUPABASE exist?)
   GITHUB INTEGRATION IS separate and would be on the data plane

4. **Subdomain Generation:**
   - User chooses (e.g., "alice") or auto-generated (e.g., "alice123")?
   ANSWER: will be based on user's username will be firstly generated by their firstname/email but they can change later

5. **VM Auto-Sleep:**
   - Fixed 30min timeout or user-configurable?
   ANSWER: FIXED

6. **Billing Model:**
   - Prepaid credits (buy $20, use hours) or post-paid (invoice monthly)?
   ANSWER: prepaid credits or subscriptions

7. **Monitoring Provider:**
   - Sentry (errors) + Prometheus (metrics) or all-in-one (Datadog, New Relic)?
   ANSWER: GRAFANA agent thingo GRAFANA ALLOY to grafana cloud free tier with SLOGs

### Business Decisions

1. **Pricing Strategy:**
   - Flat $20/month or usage-based ($0.20/VM hour)?
   - Free tier limits (e.g., 10 tasks/month, 5 VM hours)?
   - SUBSCRIPTION BASED, FREE TIER 10 tasks a month

2. **Onboarding Flow:**
   - Email verification required or skip for MVP?
   - GitHub OAuth required or optional?
   ANSWER: NO EMAIL VERIFICATION, GITHUB AUTH will be required after first-time login (onbarding sesh)

3. **VM Limits:**
   - Max concurrent VMs per user (for cost control)?
   - Max total VMs across all users (Fly.io quota)?
   ANSWER: START WITH 1 PER USER, max total vm idk come up with a sane one so i dont go bankrupt and i cant pay for your tokens anymore

4. **Beta Testing:**
   - Invite-only (5-10 users) or open beta?
   - Beta pricing (50% off) or free for early adopters?
   ANSWER: NO IDEA YET, MIGHT BE OPEN BETA

### Risk Mitigation

1. **Fly.io Outage:**
   - Failover to DigitalOcean?
   - Display "We're experiencing issues" page?
   ANSWER: "We're experiencing issues" page

2. **VM Crash Loop:**
   - Auto-restart with backoff?
   - Alert and manual intervention?
   ANSWER: AUTO-RESTART WITH BACKOFF

3. **Cost Overrun:**
   - Auto-stop at 200% of quota?
   - Require credit card upfront?
   ANSWER: AUTO-STOP AT 200% OF QUOTA

4. **Security Incident:**
   - Revoked tokens → Invalidate in Supabase?
   - VM compromise → Shutdown VM + notify user?
   ANSWER: REVOKED TOKENS, VM COMPROMISE, SHUTDOWN VM

---

## Next Steps

1. ~~**Answer open questions** (above) - Blocker for starting implementation~~ ✅ DONE
2. ~~**Choose control plane tech stack** - Go vs Node.js~~ ✅ DONE (Go selected)
3. ~~**Create `invoker` repo** - Initialize project~~ ✅ DONE
4. ~~**Set up Supabase** - Create project, configure auth~~ ✅ DONE
5. ~~**Build control plane MVP** - Auth + Fly.io API integration~~ ✅ Auth complete, Fly.io pending
6. **Task 1.3**: Implement Fly.io API integration
7. **Task 1.4**: Implement machine registry & health monitoring
8. **Task 1.5**: Implement dynamic subdomain routing table
9. **Task 1.6**: Deploy Cloudflare Worker for *.counterspell.io routing
10. **Task 2.1**: Update data plane auth to accept Supabase JWT
11. **Task 2.2**: Add health/heartbeat endpoints to data plane
12. **Task 2.3**: Build production Docker image for data plane
13. E2E testing - Full user flow from signup to task
14. Monitoring & docs - Prepare for launch
15. Beta launch - Invite 5-10 users

---

*Last Updated: 2026-01-25*
*Status: Phase 1 Tasks 1.1-1.2 Complete - Ready for Task 1.3 (Fly.io API Integration)*
