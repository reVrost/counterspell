# MVP (2-Day) — Mobile UX for Agent Runs

**Date:** January 31, 2026
**Goal:** Prove that a phone can be the best control surface for coding agents by enabling: create task → watch live run → approve changes → review diff.

---

## Success Criteria (2-Day Demo)

- A user can start a task from mobile in under 60 seconds.
- Live run view shows streaming updates (text/tool/status) with no stalls.
- After run, user sees a readable diff and can accept/reject.
- Works with at least one real backend (Claude Code) and one fallback (Native).

---

## Scope (Must Ship)

### Backend
- **Adapter for Claude Code** (stream-json parsing already exists).
- **Native backend** as fallback.
- **Normalized event stream** over SSE to UI:
  - `run.started`, `assistant.delta`, `tool.call`, `tool.result`, `run.error`, `run.completed`.
- **Run persistence:** store events + final message + diff in SQLite.
- **Pairing URL:** CLI prints local URL and QR for mobile.

### Mobile UI (PWA)
- **Inbox view** (list of tasks with status).
- **Run view** (timeline of events, streaming).
- **Diff view** (file list + unified diff).
- **Action bar:** Start, Stop, Approve, Reject.

---

## Out of Scope (Explicitly Not MVP)

- Cloud mode (Fly.io), billing, or Supabase auth.
- Push notifications.
- Multi-device sync.
- Agent marketplace, plugins, or MCP settings UI.
- Advanced permission rules.

---

## 2-Day Build Plan

### Day 1 — Backend + Streaming
- Add a normalized event struct and SSE endpoint.
- Map Claude Code events → normalized events.
- Store per-run events in SQLite.
- Build minimal mobile shell (task list + run view).

### Day 2 — Review + Diff
- Generate git diff after run.
- Render diff in mobile view (file list + unified diff).
- Add accept/reject actions (persist decision).
- Polish: basic error states + empty states.

---

## Demo Script
1. Open mobile PWA at local pairing URL.
2. Create task: “Fix the failing test in `auth.go`”.
3. Watch live event stream (tool calls, logs, assistant text).
4. View diff summary and accept.

---

## Post-MVP Next Steps (If Demo Lands)

- Add permissions UI (ask/approve tool calls).
- Add push notifications for approvals and completion.
- Add Codex + OpenCode adapters.
- Add “resume run” and session history timeline.
