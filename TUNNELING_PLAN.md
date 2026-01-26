# Counterspell Tunneling Architecture & Plan

## 1. The Vision: "Local First, Globally Accessible"

We are building a **Split-Brain Architecture** where:
- **The Brain (Logic/Agents)** runs locally on the user's machine (Data Plane).
- **The Face (UI/Auth)** runs on the Edge (Control Plane).
- **The Connection** is a secure tunnel linking them.

**Goal:** Allow users to access their local coding agent from any device (phone, tablet, laptop) via a secure `username.counterspell.app` URL, without complex networking setup.

---

## 2. Architecture Overview

```mermaid
graph TD
    User[User's Browser] -->|https://alice.counterspell.app| Edge[Control Plane (Cloudflare/Vercel)]
    
    subgraph "Control Plane (counterspell.io)"
        Edge -->|Serving Static UI| UI_Bucket[UI Assets]
        Edge -->|Auth & Routing| Supabase[Auth / DB]
        Edge -->|Tunnel Request| TunnelServer[Tunnel Ingress]
    end
    
    subgraph "Data Plane (User's Machine)"
        TunnelClient[Binary Tunnel Client] -->|Secure WebSocket| TunnelServer
        TunnelClient -->|Local API Request| GoServer[Go Backend :8710]
        GoServer -->|SQL| SQLite[Local DB]
        GoServer -->|Exec| Agents[Claude/Native Agents]
    end
```

### Key Components

| Component | Responsibility | Location | Status |
|-----------|----------------|----------|--------|
| **Control Plane** | Auth, Billing, Tunnel Routing, Subdomain Management | `counterspell.io` | Spec Defined |
| **Data Plane** | Agent Logic, File Access, Local DB, Tunnel Client | User's Machine | **In Progress** |
| **The Tunnel** | Secure pipe between Control & Data Plane | Cloudflare / Ngrok | Next Step |

---

## 3. Implementation Status: Authentication (Completed)

We have implemented a secure **OAuth 2.0 Authorization Code Flow with PKCE**.

### Why this approach?
- **Security:** The JWT *never* appears in the URL or browser history.
- **UX:** No copy-pasting codes. The browser creates a magical hand-off to the CLI.
- **Standards:** Compliant with OAuth 2.0 best practices for public clients.

### The Flow
1. **CLI Start:** Generates PKCE `code_verifier` & `code_challenge`.
2. **Request:** CLI asks Control Plane for a login URL (sending `code_challenge`).
3. **Login:** User logs in at `counterspell.io` (Supabase handled internally).
4. **Redirect:** Control Plane redirects browser to `http://localhost:8711/callback?code=abc...`.
5. **Exchange:** CLI (listening on :8711) takes `code`, validates `state`, and POSTs to Control Plane to get JWT.
6. **Token:** JWT is stored locally in `counterspell.db` (sqlite).

**Files Implemented:**
- `internal/services/auth.go`: PKCE generation, Token storage.
- `internal/services/controlplane.go`: API client for Auth & Machine Registration.
- `internal/cli/callback.go`: Local HTTP server for zero-friction handoff.
- `cmd/app/main.go`: Integration into startup flow.

---

## 4. Next Step: The Tunneling Implementation

### 4.1. Registration
Once authenticated, the binary calls `POST /api/v1/machines/register`.
- **Input:** Machine info (OS, Arch, CPU).
- **Output:** Authorized Subdomain (e.g., `alice`) and Tunnel Configuration.

### 4.2. Tunnel Establishment
We will use **Cloudflare Tunnels (cloudflared)** via the Go SDK or wrapper.

**Command:**
```bash
# Concept
cloudflared tunnel run --token <token_from_control_plane>
```

**Implementation Plan:**
1. **Tunnel Service:** Create `internal/services/tunnel.go`.
2. **Auto-Download:** If `cloudflared` isn't found, download the binary for the OS.
3. **Configuration:** Configure ingress rules to forward `https://alice.counterspell.app` -> `http://localhost:8710`.
4. **Lifecycle:** Start tunnel on app launch, stop on shutdown.

---

## 5. Control Plane API Specification

The Data Plane (Binary) expects these endpoints to exist on the Control Plane:

### Auth
- `POST /api/v1/auth/url`
  - Body: `{ "machine_name": "...", "redirect_url": "...", "code_challenge": "...", "state": "..." }`
  - Returns: `{ "auth_url": "..." }`

- `POST /api/v1/auth/exchange`
  - Body: `{ "code": "...", "code_verifier": "...", "state": "..." }`
  - Returns: `{ "jwt": "...", "user_id": "...", "email": "..." }`

### Machines & Tunnel
- `POST /api/v1/machines/register` (Protected)
  - Headers: `Authorization: Bearer <jwt>`
  - Body: `{ "machine_id": "...", "capabilities": {...} }`
  - Returns: `{ "subdomain": "alice", "tunnel_token": "..." }`

---

## 6. Subdomain Routing (The "Split Brain" Logic)

The Frontend (Svelte) is a **Single Page App** served from the Edge, but it needs to talk to the *correct* backend.

**Logic:**
1. User visits `alice.counterspell.app`.
2. Frontend loads assets from CDN.
3. Frontend API Client (`api.ts`) detects hostname `alice.counterspell.app`.
4. Requests sent to `/api/*` are routed by the Control Plane's Edge Worker.
5. Edge Worker looks up `alice` -> Finds Tunnel ID -> Proxies request down the tunnel.
6. Local Binary receives request -> Executes Agent -> Returns JSON.

This architecture allows for **infinite scaling** of users without running infinite servers in the cloud. We only pay for the "pipe" (tunneling), while users pay for the "compute" (running the agent locally).
