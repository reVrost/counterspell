# Auth + Tunnel Workflow (v1)

**Date:** January 31, 2026
**Status:** Draft (design-only)

## Goal
Enable this flow as soon as possible:
1. User runs `counterspell` (local agent server starts).
2. CLI authenticates against Invoker (hosted on AWS App Runner).
3. Invoker provisions a subdomain + tunnel for the machine.
4. User opens `https://<subdomain>.counterspell.app` on phone and uses the Svelte UI to drive Claude Code.

This document focuses on **workflow and contracts**, not implementation.

---

## System Roles

- **Counterspell CLI** (local machine): runs agent backend + serves UI/API on localhost.
- **Invoker API** (remote, App Runner on `counterspell.io`): auth, machine registry, tunnel provisioning.
- **Cloudflare Tunnel**: reverse proxy from public subdomain to local server.
- **Mobile PWA**: uses the public URL to talk to the local server.

---

## Happy Path (Browser OAuth)

1) **CLI starts**
- Starts local HTTP server (e.g. `http://localhost:8710`).
- Generates a `machine_id` (stable per device).

2) **CLI requests auth URL**
- `POST https://counterspell.io/api/v1/auth/url`
- Includes `redirect_uri` and optional `state`/`code_challenge`.

3) **CLI opens browser**
- User completes Supabase OAuth login.
- Supabase redirects to `https://counterspell.io/api/v1/auth/callback` (must be allow‑listed).
- Invoker stores the `code` against the pending login.

4) **CLI polls for auth code**
- `POST /api/v1/auth/poll` with `state` until status is `ready`.

5) **CLI exchanges code for machine JWT**
- `POST /api/v1/auth/exchange` with `code`, `state`, optional `code_verifier`.
- Response returns `machine_jwt`.

6) **CLI registers machine + provisions tunnel**
- `POST /api/v1/machines/register` with `machine_id` + system info.
- Response returns `subdomain` + `tunnel_token` (and later `tunnel_id`).

7) **CLI starts Cloudflare Tunnel**
- Uses `tunnel_token` to connect local server to Cloudflare.
- Public URL becomes `https://<subdomain>.counterspell.app` (data‑plane domain).

8) **User opens phone**
- Visits `https://<subdomain>.counterspell.app`.
- Svelte UI loads and connects to local agent server through the tunnel.

---

## Headless Flow (Device Code) — Recommended

This avoids needing a browser on the machine running Counterspell.

1) CLI requests a device code
- `POST /api/v1/auth/device/start`
- Response: `{ device_code, user_code, verification_url, interval }`

2) CLI prints login URL + code
- Example: `Go to https://counterspell.io/device and enter ABCD-EFGH`.

3) User completes login on phone
- User visits `https://counterspell.io/device`.
- Invoker links device_code to user and issues machine JWT.
- Web UI calls `/api/v1/auth/device/approve` with `user_code` (Supabase JWT).

4) CLI polls for token
- `POST /api/v1/auth/device/poll` with `device_code`.
- Response returns `machine_jwt` when approved.

5) Continue with machine registration and tunnel provisioning (same as above).

**Note:** Device code endpoints are implemented in Invoker; UI approval lives at `https://counterspell.io/device`.

---

## Invoker Endpoints (Current)

- `POST /api/v1/auth/url`
  - Creates Supabase OAuth URL.
- `GET /api/v1/auth/callback`
  - Browser redirect endpoint (stores auth code for polling).
- `POST /api/v1/auth/poll`
  - CLI polls for auth code (state → code).
- `POST /api/v1/auth/exchange`
  - Exchanges OAuth code → **machine JWT** (currently mock).
- `POST /api/v1/machines/register`
  - Returns `{ subdomain, tunnel_token, tunnel_provider }`.

---

## Invoker Enhancements (Status)

1) **Real JWT signing**
- Implemented (HS256 via `JWT_SECRET`, 30‑day expiry).

2) **Device code flow**
- Implemented: `/api/v1/auth/device/start`, `/api/v1/auth/device/poll`, `/api/v1/auth/device/approve`.

3) **OAuth polling**
- Implemented: `/api/v1/auth/poll` (CLI polls for auth code).

4) **Tunnel provisioning**
- Invoker creates Cloudflare tunnel + DNS record and returns:
  - `subdomain`
  - `tunnel_id`
  - `tunnel_token`
  - (optional) `account_id` or `tunnel_name`

---

## Local Storage (SQLite)

Persist auth + machine data in the local SQLite DB (via sqlc).

**Machine metadata** lives in `machine_identity`:

```sql
CREATE TABLE IF NOT EXISTS machine_identity (
    machine_id TEXT PRIMARY KEY,
    user_id TEXT NOT NULL,
    subdomain TEXT NOT NULL UNIQUE,
    tunnel_provider TEXT NOT NULL CHECK(tunnel_provider IN ('cloudflare')),
    tunnel_token TEXT NOT NULL,
    created_at INTEGER NOT NULL,
    last_seen_at INTEGER
);
```

**Auth token** lives in `settings` (single‑tenant local DB). Add fields such as:
`machine_jwt TEXT` and `machine_id TEXT` to store the current machine token and
stable machine identity. Optionally also store `machine_user_id`,
`machine_subdomain`, and `token_expires_at` for convenience.

---

## Open Questions

- Should Invoker return a **full URL** (`https://kenley.counterspell.app`) to avoid client assumptions?
- Do we want a **refresh token** / re‑auth strategy for long‑running machines?
- Should `tunnel_id` be stored alongside `tunnel_token` in local DB?

---

## Minimum Viable Implementation Notes

- Use browser OAuth for v1 (already implemented endpoints).
- Use Cloudflare tunnel token returned by Invoker to expose local server.
- Persist machine JWT + tunnel token locally and reuse on next run.
- Store `machine_jwt` in the `settings` table (single‑tenant local DB).
- Serve the Svelte UI from the local Counterspell server (simplest); tunnel just exposes it.
