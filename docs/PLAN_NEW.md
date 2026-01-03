***

# ARCHITECTURE.md: Agentic Workflow Engine

## 1. Executive Summary

We are building a **self-hosting, mobile-first agent orchestration system**. The system acts as a "Command & Control" center where humans define intent, and AI agents execute, review, and report back at "inference speed."

**Core Philosophy:** "Live Hypermedia."
The UI is a real-time reflection of server state, avoiding the complexity of React SPAs in favor of Go-driven HTML streaming (SSE), type-safe templates, and optimistic interactions.

---

## 2. Visual Architecture

### 2.1 System Context Diagram

```ascii
      +---------------------+                  +----------------------+
      | User Device (Web)   |                  |  External LLM API    |
      | (Mobile / Desktop)  |                  | (Claude / OpenAI)    |
      +----------+----------+                  +-----------+----------+
                 |                                         ^
      HTTP/HTML  | SSE Stream                              | JSON/API
      (HTMX)     | (Logs/Updates)                          |
                 v                                         v
      +-----------------------------------------------------------+
      |                   GO SERVER (App Engine)                  |
      |                                                           |
      |  +-------------+      +-------------+     +-------------+ |
      |  | HTTP Router |----->| Task Domain |<--->| Agent Worker| |
      |  | (Chi/Mux)   |      | Logic       |     | (Executor)  | |
      |  +-------------+      +------+------+     +-------------+ |
      |                              |                            |
      |                       +------+------+                     |
      |                       | Persistence |                     |
      |                       | (SQLite)    |                     |
      |                       +-------------+                     |
      +-----------------------------------------------------------+
                                  ^
                                  |
                        +---------+---------+
                        |  Local Filesystem |
                        |  (Artifact Storage)|
                        +-------------------+
```

### 2.2 The "Sleek" Interaction Loop (Optimistic UI)

This details how we achieve the perception of zero latency.

```ascii
[Client Browser]                [Go Server]                  [Database]
       |                             |                            |
   (1) User Drags Card               |                            |
       |                             |                            |
   (2) UI Updates Instantly          |                            |
       (SortableJS)                  |                            |
       |                             |                            |
   (3) HTMX POST /move  ------------>|                            |
       |                             | (4) Validate State         |
       |                             |--------------------------->| (5) UPDATE status
       |                             |<---------------------------|
       |                             |                            |
       |                             | (6) Trigger Agent (Async)  |
   (7) 200 OK (Empty) <--------------|                            |
       |                             |                            |
       |                       (8) SSE Stream Begins              |
   (9) Agent Logs Appear <-----------+                            |
```

---

## 3. Technical Stack & Decisions

| Component | Choice | Rationale |
| :--- | :--- | :--- |
| **Language** | Go 1.22+ | High concurrency, single binary, strict typing. |
| **Routing** | `go-chi/chi` | Lightweight, idiomatic router. Better middleware support than `net/http` alone. |
| **Templating** | `a-h/templ` | **Type-safe HTML.** Prevents runtime UI bugs. Best-in-class for Go. |
| **Frontend** | HTMX 2.0 | Handles generic AJAX, DOM swapping, and SSE. |
| **Interactions** | Alpine.js | Handles local state (modals, dropdowns) and bridges events. |
| **Sortable** | SortableJS | Gold standard for JS drag-and-drop. |
| **Database** | SQLite + `mattn/go-sqlite3` | WAL mode enabled. Zero-config, massive read concurrency. |
| **Real-time** | SSE (Server-Sent Events) | Unidirectional stream. simpler than WebSockets for this use case. |
| **Styles** | TailwindCSS | Utility-first. Essential for rapid mobile layout adjustments. |

---

## 4. Data Model (Schema)

We use a rigid state machine to ensure agents and humans stay synchronized.

```sql
-- Enforce WAL mode for concurrency
PRAGMA journal_mode = WAL;
PRAGMA foreign_keys = ON;

-- 1. Tasks: The core unit of work
CREATE TABLE tasks (
    id TEXT PRIMARY KEY,           -- ULID (Time-sortable UUID)
    title TEXT NOT NULL,
    intent TEXT NOT NULL,          -- User prompt
    -- Status Constraints: Rigid State Machine
    status TEXT NOT NULL CHECK(status IN ('todo', 'in_progress', 'review', 'human_review', 'done')),
    position INTEGER DEFAULT 0,    -- Kanban column ordering (Lexorank or Integer gap)
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
);

-- 2. Agent Logs: "The Matrix" Code Stream
-- Ephemeral-feeling logs that show the agent thinking
CREATE TABLE agent_logs (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    task_id TEXT NOT NULL REFERENCES tasks(id) ON DELETE CASCADE,
    level TEXT CHECK(level IN ('info', 'plan', 'code', 'error', 'success')),
    message TEXT NOT NULL,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP
);

-- Indices for performance
CREATE INDEX idx_tasks_status ON tasks(status);
CREATE INDEX idx_agent_logs_task ON agent_logs(task_id);
```

---

## 5. detailed Workflow & Logic

### 5.1 The Task Lifecycle (State Machine)

1.  **Todo (Human):** Creation state. Editable.
2.  **In Progress (Agent):** Triggered when moved here.
    *   *Logic:* Spawns `AgentWorker`. Locks "Edit" capability on UI.
    *   *Visuals:* Card shows spinner + streaming log console.
3.  **Review (Agent Judge):** Agent self-moves here when complete.
    *   *Logic:* Spawns `ValidationWorker` (LLM Judge). Reviews output code/diffs.
    *   *Result:* If Pass -> `Human Review`. If Fail -> Back to `In Progress` (auto-correction).
4.  **Human Review (Human):** The gatekeeper.
    *   *Actions:* Approve (-> `Done`) or Reject (add comment -> `Todo`).
5.  **Done:** Archived state.

### 5.2 Agent Communication Protocol
Agents do not speak directly to the UI. They speak to the **Event Bus**.

*   **Go Channel:** `type EventBus chan Event`
*   **Event Struct:**
    ```go
    type Event struct {
        TaskID      string
        Type        string // "log", "mutation", "error"
        HTMLPayload string // Pre-rendered HTML fragment
    }
    ```
*   **SSE Handler:** Subscribes to the Bus and writes to `http.ResponseWriter`.

---

## 6. Edge Cases & Resilience Strategy

These are the specific failure modes we must handle for a robust MVP.

### 6.1 Network / UI Edge Cases
*   **The "Ghost" Drag:** User is offline but drags a card.
    *   *Behavior:* UI assumes success. HTMX POST fails (Network Error).
    *   *Recovery:* Use `htmx:responseError` to trigger a JS function `resetBoard()`. The page reloads or the card snaps back with a red toast notification.
*   **The "Double" Drag:** User drags card A, then immediately drags card B before A finishes syncing.
    *   *Behavior:* HTMX queues requests. Server processes sequentially.
    *   *Risk:* Position integer collision.
    *   *Mitigation:* Server recalculates position based on neighbors, trusting the `target_id` (neighbor) sent by SortableJS, not absolute index.
*   **Mobile Scroll vs. Drag:** User tries to scroll down the list but accidentally drags a card.
    *   *Mitigation:* Configure SortableJS with `delay: 150` (ms) on touch devices. User must "long press" to grab.

### 6.2 Agent / Backend Edge Cases
*   **The Hallucinating Agent:** Agent returns invalid JSON or code that doesn't compile.
    *   *Mitigation:* The `AgentWorker` has a retry loop (Max 3 retries). If all fail, Task state moves to `human_review` with a flag `needs_attention`.
*   **Server Restart mid-Agent execution:**
    *   *Risk:* Goroutine dies. Task stuck in `in_progress` forever.
    *   *Recovery:* On Server Startup (`main.go`), run a cleanup query: `UPDATE tasks SET status='todo' WHERE status='in_progress'`.

---

## 7. MVP Test Plan (End-to-End)

We will not rely solely on Unit Tests. The value is in the flow. We uses **Playwright** (Go or Node version) for E2E.

### 7.1 Test Setup
*   **Environment:** Docker container running the App + SQLite (Test DB).
*   **Mock LLM:** Do **not** hit real OpenAI/Claude APIs in tests.
    *   Create a `MockLLM` struct implementing the `Agent` interface.
    *   It sleeps 50ms and returns deterministic text depending on the prompt.

### 7.2 Critical Test Cases (E2E)

#### Scenario A: Happy Path (One-Shot)
1.  **Given** I am on the dashboard.
2.  **When** I click "New Task" and type "Print Hello World".
3.  **Then** a card appears in `Todo`.
4.  **When** I drag the card to `In Progress`.
5.  **Then** the card visually snaps to `In Progress`.
6.  **And** within 2 seconds, I see "Agent logs: Writing code..." appear on the card (via SSE).
7.  **And** the card automatically jumps to `Review` (simulating Agent finish).

#### Scenario B: The "Optimistic Failure"
1.  **Given** the server is configured to reject moves (simulated failure).
2.  **When** I drag a card to `Done`.
3.  **Then** the card momentarily stays in `Done`.
4.  **But** after <500ms, the card snaps back to `Todo`.
5.  **And** a red toast message appears: "Sync failed".

#### Scenario C: Mobile Layout
1.  **Given** I load the page with Viewport `375x812` (iPhone X).
2.  **Then** I should **not** see 4 horizontal columns (Desktop Kanban).
3.  **I should see** a Tab Bar (Todo | In Progress | Done) or a vertical carousel.

---

## 8. Implementation Roadmap (Execution Order)

### Phase 0: Foundation (Day 1-2)
- [ ] Initialize Go Module.
- [ ] Setup `sqlite` with migration files.
- [ ] Setup `chi` router and `templ` integration.
- [ ] Create `board.templ` with basic Tailwind grid.

### Phase 1: Interactive Board (Day 3-4)
- [ ] Implement `GET /tasks` (Render keys).
- [ ] Implement `POST /tasks` (Create).
- [ ] **Core:** Add SortableJS + Alpine.
- [ ] Implement `POST /tasks/{id}/move` (Update position/status).
- [ ] **Test:** Manual test of drag & drop functionality on mobile.

### Phase 2: Live Events (Day 5)
- [ ] Create `SSE` handler in Go.
- [ ] Connect client-side logs: `<div class="logs" sse-swap="...">`.
- [ ] Create a "Dummy Agent" button that just pushes log lines to SSE.

### Phase 3: The Brain (Day 6-7)
- [ ] Integrate `anthropic-sdk-go` or `openai-go`.
- [ ] Build key Agent Prompts ("Planner", "Coder", "Reviewer").
- [ ] Wire `In Progress` status change to trigger Agent Goroutine.

### Phase 4: Hardening (Day 8+)
- [ ] Implement Server Restart Recovery (Reset stuck tasks).
- [ ] Write Playwright E2E happy path.
