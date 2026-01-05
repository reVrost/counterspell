# AGENTS.md

> Shared operating rules for all agents.
> Optimize for **future readers and future change**, not speed.

---

## 1. Philosophy (Why)

- Contract first
- Clarity > cleverness
- Minimal change
- Simplicity ⇒ reliability
- Fix root causes only

---

## 2. Non‑Negotiables (Never)

Agents must **never**:
- Make unrelated or aesthetic refactors
- Change outputs without approval
- Rebuild infrastructure
- Re‑implement existing utilities or mocks
- Patch symptoms instead of causes
- Start / restart / manage dev servers
- Add dependencies without explicit approval

---

## 3. Execution Rules (How)

- Atomic changes only
- No drive‑bys
- Unrecognized changes
  ⇒ assume other agent
  → don’t revert
  → narrow scope
  → conflicts = stop + ask

**Scope gate**
- Small → proceed
- Medium → confirm
- Large → design first

**Done =**
- Works as specified
- Behavior tested
- No collateral changes

---

## 4. Uncertainty & Escalation (When to Stop)

**No guessing.**

Stop and ask if:
- Logic is underspecified
- Public contracts change
- New dependencies required
- Scope crosses subsystems
- Multiple valid interpretations exist

When blocked:
1. Read more code
2. Ask with **2–3 short options**

Escalate with **impact**, not speculation.

---

## 5. Implementation Checklist (What to Do)

1. Clarify intent
2. Define contract
3. Implement minimal change
4. Test behavior
5. Verify end‑to‑end
6. Leave breadcrumbs (decisions, assumptions)

---

## 6. Engineering Standards

**Architecture**
- Contracts = source of truth
- Separate interfaces, logic, infrastructure, entrypoints
- Shared utils: minimal, generic, stable

**Code Quality**
- Explicit > implicit
- Structured data > free text
- Patterns > novelty
- Propagate context / cancellation
- Wrap errors with intent
- Surprising change → re‑check assumptions

**Shape**
- Small functions
- Bounded files
- Explicit complexity
- Bias simple + safe

---

## 7. Testing & Observability

**Test behavior, not wiring**

Test:
- Business logic
- Math
- Conditionals
- Loops
- Transformations

Skip:
- Pass‑throughs
- Constructors
- Simple error bubbling
- Getters / setters

**Bugs:** fix root cause + regression test

**Observability**
- Log at boundaries
- Signal > noise
- No sensitive data
- Note behavior changes

---

## 8. Communication & Style

- Telegraphic
- Drop filler
- Noun phrases OK
- Minimize tokens
- Escalate with impact

---

## 9. Frontend Aesthetics

<details>
<summary><strong>Design Guardrails</strong></summary>

Real typography.
Committed palette.
Bold accents.
Intentional motion.
Depth, not flatness.
Avoid clichés and generic grids.

</details>
## Tech Stack & Constraints

### Backend (Go 1.25+)
- **Testing:** `testify/assert` @latest + `gomock` v0.6.0.
