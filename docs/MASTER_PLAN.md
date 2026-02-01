

# RFC: Counterspell (The "Linear" for Autonomous Agents)

**Date:** January 19, 2026
**Version:** 2.0 (Local-First Pivot)
**License:** FSL 1.1 (Source Available). Converts to MIT after 2 years.
**Stack:** Go (Backend/CLI), Svelte (Frontend), SQLite (Local State), Supabase (Auth/Billing).

---

### 1. Executive Summary

Counterspell is a task-based autonomous coding environment. Unlike "Chat" bots (Claude/ChatGPT), Counterspell uses a **Ticket-Based Workflow** (Inbox -> Plan -> Execute -> Review).

It operates on a **Local-First** model:

1. **The Brain runs locally:** The Agent executes on the user's machine (or private cloud), accessing files directly.
2. **The State is local:** All prompts, logs, and keys are stored in a local SQLite file.
3. **The UI is global:** A secure tunnel exposes the local interface to a PWA (`alice.counterspell.io`), allowing users to manage tasks from anywhere.

---

### 2. High-Level Architecture

The system is strictly divided into the **Control Plane** (SaaS) and the **Data Plane** (User).

#### A. The Control Plane (Supabase)

* **Role:** The Gatekeeper.
* **Data:** Users, Billing (Stripe), Tunnel Routing maps.
* **Location:** Hosted Cloud (Supabase).

#### B. The Data Plane (The "CLI")

* **Role:** The Worker.
* **Data:** Source Code, SQLite DB (Tasks, Chat History), OpenAI API Keys.
* **Location:** User's Laptop (Local Mode) or Fly.io Machine (Cloud Mode).
* **Connectivity:** Opens a secure WebSocket tunnel out to the internet.

---

### 3. Project Structure (Monorepo)

We use a standard Go layout that embeds the Svelte frontend into a single binary.

```text
/counterspell
├── cmd/
│   └── counterspell/       # The Main Binary
│       └── main.go         # Entrypoint: specifices 'server' or 'tunnel' mode
│
├── pkg/
│   ├── agent/              # LLM Logic (Step, Plan, ToolUse)
│   ├── db/                 # SQLite Schema & Queries (GORM or sqlc)
│   ├── sandbox/            # Native OS execution wrappers
│   ├── tunnel/             # Cloudflare/Ngrok SDK wrapper
│   └── server/             # HTTP API (Chi/Fiber) serving the UI
│
├── web/                    # The Svelte Kit Frontend
│   ├── src/
│   │   ├── routes/         # /inbox, /ticket/[id], /settings
│   │   ├── lib/            # UI Components (DiffView, TerminalOutput)
│   │   └── api.ts          # Client for local API
│   └── dist/               # Compiled static assets (embedded in Go)
│
├── internal/
│   └── supabase/           # JWT Validation Logic (Public Keys)
│
├── go.mod
└── Dockerfile              # For Cloud Mode (Fly.io)

```

Separate repo between data-plane and control-plane, build data plane first get adoption, feedback, virality only build the sass when its proven lots of users

---

### 4. Data Topology: Who Stores What?

We strictly separate "SaaS Data" from "User Data."

| Feature               | Storage Location | Technology | Schema Example                                       |
| --------------------- | ---------------- | ---------- | ---------------------------------------------------- |
| **User Login**        | **Supabase**     | Postgres   | `users(id, email, plan_tier)`                        |
| **Billing**           | **Supabase**     | Postgres   | `subscriptions(stripe_id, status)`                   |
| **Task History**      | **Local**        | SQLite     | `tasks(id, title, status, machine_id, created_at)`   |
| **Chat Logs**         | **Local**        | SQLite     | `activities(task_id, role, content)`                 |
| **LLM Keys**          | **Local**        | SQLite     | `secrets(key_name, value)` (Encrypted - see Roadmap) |
| **Code Diffs**        | **Local**        | SQLite     | `diffs(id, file_path, old_blob, new_blob)`           |
| **Machine Instances** | **Local**        | SQLite     | `machines(id, name, mode, last_seen_at)`             |

#### Why Machine ID Matters

Users will have multiple devices running agents:
- **Laptop** (primary workspace)
- **Desktop** (home setup)
- **Cloud Mode** (Fly.io instance)

Without a first-class `machine_id` concept:
- You can't distinguish which device executed a task
- Cloud mode feels disconnected from local workflow
- Debugging becomes ambiguous ("which machine ran this task?")
- Multi-device auth becomes a mess

**Machine Schema:**
```sql
CREATE TABLE machines (
    id TEXT PRIMARY KEY,
    name TEXT NOT NULL,  -- e.g., "Alice's MacBook Pro", "Fly Agent #1"
    mode TEXT CHECK(mode IN ('local', 'cloud')),
    capabilities TEXT,  -- JSON: {"os": "darwin", "cpus": 8}
    last_seen_at TIMESTAMPTZ DEFAULT NOW()
);
```

**Usage:**
- Every task includes `machine_id` to track origin
- Tunnel registration includes machine_id
- UI shows active machines: "Alice's MacBook Pro (● Online)"
- Cloud mode machines appear alongside local devices

---

### 5. The "Magic Tunnel" & Routing

This is how we give a Localhost app a Public URL without complex user setup.

#### Step 1: The Tunnel (Go)

Inside `pkg/tunnel`, we implement a wrapper around **Cloudflare Tunnel (cloudflared)** or a simple WireGuard implementation.

* When `counterspell start` runs, it:
1. Spins up the HTTP Server on `localhost:3000`.
2. Establishes an outbound tunnel to `tunnel.counterspell.io`.
3. Receives a stable subdomain: `https://alice-dev.counterspell.io`.



#### Step 2: Serving the UI

The Go binary uses Go 1.16+ embedding to serve the Svelte app.

```go
// pkg/server/server.go
//go:embed ../../web/dist/*
var svelteAssets embed.FS

func Start() {
    // 1. API Routes (Data)
    router.Get("/api/tasks", handleGetTasks) 
    
    // 2. Static Assets (UI)
    // If route not found in API, serve index.html (SPA Fallback)
    router.Get("/*", http.FileServer(http.FS(svelteAssets)))
}

```

#### Step 3: Auth Handshake

1. User visits `alice-dev.counterspell.io`.
2. **Svelte App** checks LocalStorage for Supabase JWT.
* *Missing?* Redirects to `auth.counterspell.io` (Supabase Login).


3. **Supabase** redirects back with `#access_token=xyz`.
4. **Svelte App** sends `Authorization: Bearer xyz` to the Go Backend.
5. **Go Backend** verifies the JWT signature using Supabase's Public Key. **It does NOT need to talk to the Supabase DB.**

---

### 6. The User Experience (The "Linear" Flow)

**A. The Inbox (Home)**

* A Kanban board or List View.
* Statuses: `Draft`, `Planning`, `Running`, `Review`, `Done`.
* *Action:* User types "Fix the bug in auth.go" -> Creates a new Ticket.

**B. The Agent Loop (Ticket View)**

1. **Planning Phase:** Agent scans files, proposes a plan. User clicks "Approve Plan."
2. **Execution Phase:** Agent runs tools (`grep`, `read_file`).
* *UI:* Shows a streaming terminal log (`> installing deps...`).


3. **Coding Phase:** Agent writes changes.

**C. The Review (Diff View)**

* The Agent stops and marks the ticket `Review`.
* The UI renders a **Split-View Diff** (Red/Green).
* User can highlight lines and comment: *"This variable name is confusing."*
* Agent reads comment -> Fixes code -> Updates Diff.
* User clicks **"Merge"** -> Agent runs `git commit`.

---

### 7. Cloud Strategy (MVP)

For users who want "Always On" agents (closing their laptop), we offer **Cloud Mode**.

**Infrastructure: Fly.io**

* **Why?** Fly.io converts a Docker container into a Firecracker MicroVM instantly. It is cheaper and simpler than managing raw EC2 or Droplets.
* **Orchestration:**
1. User clicks "Deploy to Cloud" in the UI.
2. Our SaaS triggers the Fly.io API.
3. Spins up a standard `counterspell` Docker image.
4. Attaches a persistent Volume (for the SQLite DB & Repo).



**Comparison:**

* **Local:** Free. User pays for electricity.
* **Cloud:** Paid ($XX/mo). We pay Fly.io ~$5/mo per active VM.

---

### 8. Licensing Strategy (FSL 1.1)

We use the **Functional Source License (FSL)**.

* **For Users:** It feels like Open Source. They can read the code, fork it, and modify it for personal/internal use.
* **For Competitors:** Prevents AWS or a startup competitor from taking our code and selling "Counterspell-as-a-Service."
* **Safety Valve:** The license automatically converts to standard **MIT/Apache** after 2 years. This assures users that the code will eventually be free even if we disappear.

---

### 9. Immediate Next Steps (Implementation)

1. **Init Repo:** Set up the folder structure.
2. **The "Hello World" Tunnel:**
* Build a simple Go server that returns JSON.
* Use `cloudflared` to expose it to a public URL.


3. **The Svelte Shell:**
* Create the Svelte project in `/web`.
* Configure `npm build` to output to `/web/dist`.
* Wire up Go `embed` to serve `index.html`.


4. **SQLite & Schema:**
* Initialize `gorm` or `sqlc`.
* Create the `tasks` table.

5. **Agent Adapter Contract:**
* Define a normalized event schema + capabilities map.
* Implement adapters for Claude Code and Native.
* Stream events via SSE to the UI.



This RFC provides a complete blueprint for the **"Linear for Agents"** pivot. It minimizes server cost, maximizes privacy, and offers a clear upgrade path to paid Cloud hosting.


Yes, this is a standard "Multi-Tenant" SaaS pattern (used by Vercel, Slack, Notion).1 You **do not** need to deploy a new Svelte server for every user. You deploy **one** Svelte app, and it behaves differently based on the URL the user is visiting.2

Here is exactly how to achieve the **"One Deployment, Infinite Subdomains"** architecture to minimize costs.

---

### The Architecture: "The Wildcard Shell"

You will use **Cloudflare Pages** (or Vercel) to host a single instance of your frontend.

1. **DNS:** You configure `*.counterspell.io` to point to that single deployment.
    
2. **SvelteKit:** The app reads the `host` header (e.g., `alice.counterspell.io`).3
    
3. **Routing:**
    
    - **Static Assets (HTML/CSS):** Served immediately by Cloudflare (Global CDN).
        
    - **API Calls (`/api/*`):** The browser (client-side) routes these requests through the **Tunnel** to the user's local laptop.
        

---

### Step 1: Infrastructure Setup (DNS)

You need a **Wildcard CNAME**.

1. **In Cloudflare DNS:**
    
    - Record: `CNAME`
        
    - Name: `*` (Wildcard)
        
    - Target: `counterspell-web.pages.dev` (Your single Svelte deployment)
        
2. **In Cloudflare Pages:**
    
    - Add Custom Domain: `counterspell.io` (This automatically covers wildcards on Enterprise, but on Free/Pro, you might need to add specific subdomains or use a **Cloudflare Worker** to handle the wildcard routing, which is cheaper).
        

**Cost:** $0 (Static hosting is free).

---

### Step 2: The Svelte Implementation (Client-Side Routing)

The trick is that the **UI is the same for everyone**, but the **Data Source** changes.

In your `web/src/lib/api.ts` (your API client wrapper), you dynamically determine the API target.

TypeScript

```
// web/src/lib/api.ts
import { browser } from '$app/environment';

export function getApiUrl() {
  if (!browser) return 'http://localhost:8080'; // Server-side fallback

  const hostname = window.location.hostname; // e.g., "alice.counterspell.io"
  
  // 1. Dev Mode
  if (hostname.includes('localhost')) {
    return 'http://localhost:8080';
  }

  // 2. Production: Extract subdomain
  const subdomain = hostname.split('.')[0]; // "alice"
  
  // 3. Routing Logic
  // Option A: If using a Tunnel service like Ngrok/Cloudflare Tunnels
  // The tunnel is likely exposed at a predictable URL or we proxy via the current domain.
  
  // If "Cloud Mode" (User is visiting alice.counterspell.io), 
  // we want API calls to go to the SAME domain, but /api/ path.
  // The Tunnel (Step 3) will intercept these.
  return `https://${hostname}/api`; 
}
```

---

### Step 3: The "Split-Brain" Routing (The Magic)

This is the hardest part: **How do we serve the UI from Cloudflare, but send API requests to Alice's Laptop?**

You use a **Cloudflare Worker** (Edge Proxy) sitting in front of your domain.4

**⚠️ Critical Infrastructure Warning**

The Cloudflare Worker becomes the **choke point** for all users. If it goes down, everyone's UI breaks.

**Must-haves for the Worker:**
- **Minimal logic:** No auth decisions, no heavy computation
- **Fail-closed:** If routing table lookup fails, return 404 (don't leak)
- **Rate limiting:** Prevent abuse/tunnel flooding
- **Health monitoring:** Alert if worker becomes unresponsive
- **Graceful degradation:** Serve static assets even if tunnel routing fails

**The Worker Script (`wrangler.toml`):**

JavaScript

```
export default {
  async fetch(request, env) {
    const url = new URL(request.url);
    const subdomain = url.hostname.split('.')[0]; // "alice"

    // ROUTE 1: Is this an API call? -> Send to User's Tunnel
    if (url.pathname.startsWith('/api')) {
      // You need a mapping of "alice" -> "tunnel-id"
      // Or, if using Cloudflare Tunnels, you route to the specific tunnel tag.
      
      // Example: Forward to the specific tunnel URL you assigned this user
      const userTunnelUrl = `https://${subdomain}-tunnel.counterspell-infra.com`;
      
      // Rewrite the request to target the tunnel
      const newUrl = new URL(request.url);
      newUrl.hostname = new URL(userTunnelUrl).hostname;
      
      return fetch(newUrl.toString(), request);
    }

    // ROUTE 2: Is this the Frontend? -> Serve the Svelte App
    // Fallback to fetching the static assets from Cloudflare Pages
    return env.ASSETS.fetch(request);
  }
};
```

**Why this saves money:**

1. **Frontend:** Cached globally. You pay $0 for bandwidth/hosting.
    
2. **Backend:** You aren't hosting it. You are just **proxying** bytes from the user's browser to the user's laptop.
    
3. **Compute:** The Worker runtime costs are tiny ($5/mo for millions of requests).
    

---

### Step 4: Configuring the Local Tunnel (Go)

When the user runs `./counterspell start`, your Go binary needs to create that tunnel endpoint.

You can use the **Cloudflare Tunnel SDK** (embedded in Go) or just use **ngrok-go**.

**Example (Using Cloudflare Tunnels):**

1. User runs `counterspell start`.
    
2. Go Binary starts:
    
    - HTTP Server on `:8080`.
        
    - Tunnel Client: Connects to Cloudflare.
        
    - **Register:** Tells your Supabase DB: _"I am Alice, and my active tunnel ID is `cf-tunnel-xyz-123`."_
        
3. **Result:**
    
    - The Cloudflare Worker (Step 3) sees the request to `alice.counterspell.io/api`.
        
    - It looks up the tunnel ID.
        
    - It streams the request down to the Go Binary.
        

### Summary of the "Single Deployment" Flow

1. **User Browser** -> `alice.counterspell.io`
    
2. **Cloudflare Worker** intercepts:
    
    - Request for `index.html`? -> Serve from **Svelte Bucket** (Fast).
        
    - Request for `/api/tasks`? -> Forward through **Tunnel** -> **Alice's Laptop**.
        
3. **Alice's Laptop (Go)**:
    
    - Receives request.
        
    - Queries SQLite.
        
    - Returns JSON.
        

Next Step for you:

When you build the Svelte app, build it as a Static Adapter (adapter-static). This ensures it's just a pile of HTML/JS files that are cheap to host, making the "Split-Brain" routing much easier.

---

### 10. Critical Success Factors

This pivot succeeds or fails based on execution. Here's what actually matters.

#### 10.1 Success Conditions (Non-Negotiable)

**Onboarding Must Be Seamless**
- First task must succeed in <10 minutes
- Zero tunnel confusion (auto-generated URL works instantly)
- Great defaults (sensible agent behavior out of the box)

**Trust Through Transparency**
- Clear diff view before any changes are applied
- Explicit approvals for every write operation
- No surprise file modifications
- Logs show exactly what the agent is doing

**Perceived Intelligence**
- Planning phase must feel "thoughtful" (agent scans files first)
- Even mediocre LLMs feel good if UX is right
- Streaming logs show progress, not just a spinner

**Reliability**
- Agent never "hangs silently"
- Always explain what it's doing
- Graceful failure with recovery paths

#### 10.2 Failure Modes (What Kills Adoption)

These will cause users to abandon the product:

- ❌ **"Too much setup for not much value"** - Complexity of local + tunnel should be invisible
- ❌ **Tunnel issues without good UX** - Clear error messages, auto-reconnect, status indicator
- ❌ **Agent breaks repo or runs wild** - Sandbox restrictions, confirm before dangerous operations
- ❌ **UI feels slow or janky** - Svelte should be instant, SSE for real-time updates
- ❌ **Users don't understand why it's better than chat** - Emphasize control, review workflow, git integration

#### 10.3 Strategic Insight: You're Not Selling AI

**You are selling control.**

Best messaging:
- "Your code never leaves your machine"
- "Review every change before it happens"
- "Agents you can trust"

If you lean into this:
- Higher willingness to pay
- Lower churn (users feel ownership)
- Slower but stickier growth

---

### 11. Mobile UX Wedge (The Differentiator)

**The phone is the control plane.** The laptop/VM runs the agent; the phone is where users approve, monitor, and ship.

**Must-have mobile flows:**
- **One-thumb approvals:** Accept/reject edits and permissions without opening a laptop.
- **Glanceable run feed:** Streaming events with clear state (planning, running, waiting, done).
- **Diff-first review:** Show a readable diff on mobile with file-level summary.
- **Interrupt & resume:** Stop a run, ask a question, or resume later.

**Why this wins:**
- Every other agent UX assumes a desktop screen.
- Mobile turns “background agent” into a continuous workflow.
- It forces good abstractions: small, clear events and explicit approvals.

---

### 12. Agent Adapter Contract (Durable Interface)

All agent backends should conform to a **normalized event stream + capability map**. This keeps the UI stable while backends evolve.

**Backend capabilities (matrix):**
- `streaming_deltas`, `tool_calls`, `tool_results`
- `reasoning`, `status_updates`, `permissions`, `questions`
- `file_changes`, `command_execution`, `mcp_tools`

**Normalized events (UI contract):**
- `session.started`, `session.ended`
- `run.started`, `run.completed`, `run.error`
- `assistant.delta`, `assistant.message`
- `tool.call`, `tool.result`
- `file.change`, `command.output`
- `permission.requested`, `permission.resolved`
- `question.requested`, `question.resolved`
- `raw` (fallback passthrough for unknown agent output)

**Key principle:** UI depends on these events, not on any specific agent’s native schema.

---

### 13. Roadmap: Post-MVP Features

#### 11.1 API Key Encryption (Priority: Medium)

Currently stored plaintext in SQLite. Future enhancement:

```go
// Use OS keychain or encrypted SQLite
type SecureStore interface {
    Set(key, value string) error
    Get(key string) (string, error)
}

// Implementations:
// - macOS: keychain package
// - Linux: libsecret / GNOME Keyring
// - Windows: Windows Data Protection API
// - Fallback: AES-256-GCM with user password
```

**Why not now?** Local-first means only user has access. Encryption adds complexity without immediate benefit. Add when cloud sync or multi-device sync is needed.

#### 11.2 Multi-Device Sync (Priority: Low)

Users want to start task on laptop, continue on phone. Options:
- **Simple:** Export/import JSON backups
- **Better:** Optional sync to encrypted cloud storage
- **Full:** Real-time sync via Control Plane (user opt-in)

**Recommendation:** Start with manual export/import. Let users request sync features.

#### 11.3 Advanced Agent Capabilities

- Parallel execution (multiple tasks across machines)
- Agent handoff (local planning → cloud execution)
- Scheduled tasks (cron-like for agents)
- Agent marketplace (community-contributed workflows)
