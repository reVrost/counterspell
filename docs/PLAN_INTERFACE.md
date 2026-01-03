Here is your final, execution-ready **PLAN.md**. Save this to your repository root. It unifies the architecture, the "Mobile First" strategy, and the Git/Worktree engine into a concrete Step-by-Step guide.

***

# PLAN.md: The "Pocket CTO" Engine

**Objective:** Build a self-hosted, mobile-first AI agent orchestration system that allows parallel coding tasks ("fire and forget").
**Tech Stack:** Go (Backend) + HTMX/Alpine (Frontend) + SQLite (DB) + SSE (Realtime).
**Core Moat:** A "Sleek" Mobile PWA interface that unlocks asynchronous management of AI coding agents, leveraging Git Worktrees for safety.

---

## 📅 Phase 1: The "Pocket Board" (Mobile Foundation)
**Goal:** A "native-feeling" PWA on your phone. No AI yet. Just a flawless manual Kanban board accessible via Tailscale.

### Week 1: Scaffold & Mobile Layout
- [ ] **Init Project:**
    - [ ] `go mod init` & setup standard folder structure (`cmd/`, `internal/`, `tools/`).
    - [ ] Install `templ`, `chi`, `sqlite3`.
    - [ ] Configure `TailwindCSS`.
- [ ] **Responsive Shell (HTML/Templ):**
    - [ ] Create `layout.templ`.
    - [ ] **Crucial:** Implement "Tabs vs Grid" logic.
        - *Mobile:* Show Bottom Tab Bar (Todo | Doing | Done). Only render 1 column container.
        - *Desktop:* Show standard 3-column Grid.
- [ ] **Database Setup:**
    - [ ] Create `tasks` table in SQLite.
    - [ ] Implement `TaskService` (Create, List, Move).

### Week 2: "Sleek" Interactions
- [ ] **Drag & Drop (SortableJS):**
    - [ ] Init SortableJS with `delay: 150` (prevents scrolling conflict on touch).
    - [ ] Implement **Optimistic UI:** Move DOM element -> Send AJAX -> Revert on Error.
    - [ ] Add **Haptics:** `navigator.vibrate(10)` on drag start using Alpine.js.
- [ ] **Manifest & PWA:**
    - [ ] create `manifest.json`. Set `display: standalone`.
    - [ ] Add app icons.
    - [ ] **Test:** "Add to Home Screen" on iOS/Android. Ensure no browser URL bar shows.

---

## 📅 Phase 2: The "Asynchronous Factory" (Agent Engine)
**Goal:** Parallel agents working without blocking. The "Iron Man" suit.

### Week 3: Git Worktree Engine
- [ ] **Worktree Manager (`internal/git`):**
    - [ ] Implement `CreateWorktree(taskID string, baseBranch string) (path string, err error)`.
    - [ ] Implement `CleanupWorktree(path string)`.
- [ ] **Agent Sandbox:**
    - [ ] Update `AgentRunner` to accept a `WorkDir`.
    - [ ] Ensure all CLI commands (`go test`, `ls`) run inside that isolated path.

### Week 4: The "Brain" & Real-Time Stream
- [ ] **SSE Pipeline:**
    - [ ] create `internal/events/bus.go`.
    - [ ] implement `GET /events` handler.
    - [ ] **Visual:** Implement the "Matrix Rain" console log on the card UI using SSE.
- [ ] **The "Brain Lift" (LLM Integration):**
    - [ ] Integrate `anthropic-go` SDK.
    - [ ] Implement the **Planner** prompt (Read File -> Think -> Propose Edit).
    - [ ] Stream "Thinking..." tokens to the SSE bus.

---

## 📅 Phase 3: The 10x Mobile Workflow (Validation)
**Goal:** Review and Merge code from the toilet/bed/commute.

### Week 5: The Pocket Reviewer
- [ ] **Diff Viewer:**
    - [ ] Create a "Review" View (Mobile Optimized).
    - [ ] Use `chroma` (Go syntax highlighter) to render diffs server-side.
    - [ ] **UI:** Accordion-style "Changed Files". Click header to expand diff.
- [ ] **Gestures:**
    - [ ] Add "Swipe Left/Right" on the Review Card using Alpine.js (`x-on:touchstart` logic).
    - [ ] **Right:** Merge & Deploy.
    - [ ] **Left:** Reject (Open comment modal).

### Week 6: Notifications & Polish
- [ ] **Push Notifications:**
    - [ ] Implement Web Push API (VAPID keys).
    - [ ] Trigger Push when Agent moves task to `Review` or `Done`.
- [ ] **Offline Resilience:**
    - [ ] Add basic Service Worker to cache HTML shell/CSS/JS.
    - [ ] Show "Offline Mode" badge if websocket/SSE creates connection error.

---

## 🛡️ Edge Cases & Testing Strategy

### Critical E2E Tests (Playwright)
1.  **The "Subway Tunnel" Test:**
    -   Disconnect Network -> Drag Card -> Reconnect.
    -   *Expectation:* UI Optimistically moves, shows "Syncing...", then Retries or Reverts.
2.  **The "Git Conflict" Test:**
    -   Two agents edit same file in parallel worktrees.
    -   *Expectation:* Second merge fails gracefully. Task moves to "Human Review" with "Merge Conflict" error tag.

### Security Checklist
- [ ] **Host Binding:** Bind server to `0.0.0.0` (for Docker/Tailscale).
- [ ] **Directory Traversal:** Ensure Agent `ReadFile` tool cannot reach `../` outside the worktree.
- [ ] **Rate Limits:** Simple semaphore to limit MaxConcurrentAgents (e.g., 2) to save API costs.

---

## 🚀 Execution Order (Day 0)

1.  **Initialize:** `mkdir pocket-cto && cd pocket-cto && go mod init github.com/you/pocket-cto`.
2.  **Architecture:** Copy `ARCHITECTURE.md` to root.
3.  **Plan:** Copy this `PLAN.md` to root.
4.  **First Code:** Create `cmd/server/main.go` and `internal/ui/layout.templ`. Output "Hello World" to your phone via Tailscale/ngrok.

*"The moat is not the AI. The moat is being able to wield the AI while walking your dog."*
