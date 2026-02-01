# Backend v2 (Agent Adapter Contract)

**Date:** January 31, 2026
**Status:** Draft (design only, no code changes yet)

## Goal
Define a durable backend interface so the UI only depends on:
- A normalized event stream
- A capability matrix
- Minimal session/run lifecycle controls

This keeps Counterspell stable as agent CLIs evolve (Claude Code, Codex, OpenCode, Amp, etc.).

---

## Design Principles

1. **UI stability > backend detail**
   The UI listens only to normalized events. Agent-specific schemas never leak into UI code.

2. **Capabilities first**
   Backends declare what they can do. The UI adapts (e.g., no reasoning → no reasoning panel).

3. **Event sourcing**
   All events can be stored and replayed to rebuild UI state.

4. **Opaque state**
   Backends can persist/restore opaque state blobs without exposing internal format.

---

## Core Types (Proposed)

### Backend + Session

```go
// Backend is a factory + capability descriptor.
type Backend interface {
    Info() BackendInfo
    Capabilities() Capabilities
    NewSession(ctx context.Context, opts SessionOptions) (Session, error)
    Close() error
}

// Session owns a conversation context.
type Session interface {
    ID() string
    Run(ctx context.Context, input TurnInput) (RunHandle, error)
    Continue(ctx context.Context, input TurnInput) (RunHandle, error)
    Checkpoint() ([]byte, error)
    Restore(state []byte) error
    Cancel(runID string) error
}
```

### RunHandle + Events

```go
// RunHandle streams normalized events for a single run.
type RunHandle interface {
    RunID() string
    Events() <-chan Event
    Wait() error
}

// Event is the normalized payload for UI and persistence.
type Event struct {
    Type      EventType
    SessionID string
    RunID     string
    Time      time.Time
    Seq       uint64
    Payload   any
    Raw       json.RawMessage
}
```

### Capabilities

```go
type Capabilities struct {
    StreamingDeltas bool
    ToolCalls       bool
    ToolResults     bool
    Reasoning       bool
    StatusUpdates   bool
    Permissions     bool
    Questions       bool
    FileChanges     bool
    CommandExec     bool
    MCPTools        bool
    SharedProcess   bool
}
```

---

## Normalized Event Types

```go
const (
    EventSessionStarted   EventType = "session.started"
    EventSessionEnded     EventType = "session.ended"
    EventRunStarted       EventType = "run.started"
    EventRunCompleted     EventType = "run.completed"
    EventRunError         EventType = "run.error"

    EventAssistantDelta   EventType = "assistant.delta"   // streaming text
    EventAssistantMessage EventType = "assistant.message" // full message

    EventToolCall         EventType = "tool.call"
    EventToolResult       EventType = "tool.result"

    EventFileChange       EventType = "file.change"
    EventCommandOutput    EventType = "command.output"

    EventPermissionReq    EventType = "permission.requested"
    EventPermissionRes    EventType = "permission.resolved"

    EventQuestionReq      EventType = "question.requested"
    EventQuestionRes      EventType = "question.resolved"

    EventStatus           EventType = "status"
    EventRaw              EventType = "raw" // fallback passthrough
)
```

---

## Mapping Guidance (Current Backends)

### Claude Code (CLI)
- JSONL stream → map to:
  - `assistant.delta` for streaming text chunks
  - `tool.call` and `tool.result`
  - `run.completed` or `run.error`
- Capture `session_id` and store as session/run metadata.

### Native Backend
- Emit `assistant.delta` while generating
- Emit `tool.call` and `tool.result` around tool execution
- Emit `run.completed` when agent loop finishes

---

## Persistence Strategy

- **Store every event** in SQLite with `(session_id, run_id, seq)`.
- **Rebuild UI state** by replaying events (no reliance on backend internals).
- **Checkpoint blobs** are optional and opaque.

---

## Migration Notes (From v1)

Current interface (`Run/Send/FinalMessage/Messages/Todos`) couples UI to backend data structures.
Migration path:

1. Add normalized event stream alongside current callbacks.
2. Switch UI to event playback.
3. Deprecate direct message/todo getters.

---

## MVP Event Subset (Must Ship)

For a first usable mobile UI, only these are required:

- `run.started`, `run.completed`, `run.error`
- `assistant.delta`
- `tool.call`, `tool.result`
- `status`

Everything else can be feature‑flagged by capabilities.

---

## Open Questions / TODO

- Should sessions be long‑lived across app restarts, or run‑scoped only?
- Should `assistant.message` always be emitted, or only when deltas are unsupported?
- Define a stable payload schema for each event type (JSON schema or Go structs).
- Decide how to map file diffs: real‑time `file.change` vs post‑run diff only.

