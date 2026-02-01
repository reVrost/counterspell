# AI-Native Codebase Configuration

## What "AI-Native" Means

Making a codebase "AI-native" means structuring it so an AI coding assistant (like me) can:

1. **Navigate quickly** - Find code without exhaustive searching
2. **Understand context** - Know how components relate to each other
3. **Follow patterns** - See consistent conventions across the codebase
4. **Debug efficiently** - Know where errors typically originate
5. **Make safe changes** - Understand dependencies and impact of changes

## What I've Set Up

### 1. AGENTS.md (Trimmed)
**Purpose:** My primary instruction file
- Development principles and constraints
- Code style conventions (Go + Svelte 5)
- Quick commands reference
- Common pitfalls to avoid
- Environment variables

**Before:** 375 lines with verbose examples
**After:** ~120 lines focused on essential guidance

### 2. CODEBASE.md
**Purpose:** Navigation map
- Where to find specific functionality
- Architecture flow diagram
- Service dependency graph
- Error pattern locations
- File naming conventions
- How to add new features

### 3. TASKS.md
**Purpose:** Work tracking
- Active tasks with checkboxes
- Recent changes log
- Technical debt list
- Questions for human input

## What Would Help Me Work Better

### Navigation & Context

**✅ Already have:**
- CODEBASE.md for quick lookups
- AGENTS.md for conventions
- Standard directory structure

**Would be helpful:**
- Inline comments explaining WHY (not WHAT) for complex logic
- ADOPTED_DECISIONS.md capturing architecture choices
- API_CONTRACTS.md documenting request/response formats

### Code Quality

**✅ Already have:**
- Structured logging with slog
- Consistent error wrapping
- Type hints in TypeScript

**Would be helpful:**
- JSDoc comments on exported functions (Go and TS)
- Examples of common patterns in CODEBASE.md
- More test coverage for logic (not wiring)

### Tooling

**✅ Already have:**
- Makefile with all commands
- sqlc for type-safe DB access
- LSP diagnostics integrated

**Would be helpful:**
- Pre-commit hooks for formatting (gofmt, prettier)
- A `make format` command that runs all formatters
- A `make lint` command that checks for common issues
- Integration tests (not just unit tests)

### Debugging

**✅ Already have:**
- Centralized logging to `server.log`
- Structured logging with key-value pairs
- SSE events for real-time updates

**Would be helpful:**
- Log correlation IDs for request tracing
- A `make debug` command that tail logs with filtering
- Example error scenarios and how to debug them

### Testing

**✅ Already have:**
- Go test framework
- Testify assertions
- Clear policy on what to test (logic vs wiring)

**Would be helpful:**
- Test utilities/mocks for common services
- Golden file tests for API responses
- E2E test setup with test database

## Suggested Next Steps

### Quick Wins ✅ (Completed)

1. ✅ **Added `make format` and `make check-all`**
   - `make format` runs prettier on UI + go fmt on Go
   - `make check-all` runs linters and type checks

2. ✅ **Updated ui/.prettierrc**
   - Added consistent formatting rules (semi: true, singleQuote: true, etc.)

3. ✅ **Set up Playwright E2E test framework**
   - Lightweight, text-based testing (blind agent friendly)
   - Runs against localhost:5173
   - Auto-starts dev server via webServer config
   - Console error capture, uncaught exception detection
   - 6 initial tests for dashboard
   - See `ui/tests/e2e/README.md` for usage

### Quick Wins (Remaining - 1-2 hours each)

1. **Add JSDoc to key exports**
   - Focus on `internal/services/` interfaces
   - Focus on `ui/src/lib/api.ts` functions

2. **Create API_CONTRACTS.md**
   - List all endpoints with request/response examples
   - Authentication requirements
   - Error response formats

### Medium Effort (half-day each)

1. **Add test utilities**
   - Mock implementations for external services (GitHub, etc.)
   - Test DB setup helpers

2. **Add integration tests**
   - Test real request flows end-to-end
   - Use test database

3. **Add request tracing**
   - Generate unique request ID
   - Log it at each layer

### Larger Effort (multi-day)

1. **Pre-commit hooks**
   - Run formatters before commit
   - Run lint in background

2. **Comprehensive E2E tests**
   - Test user flows from browser to database
   - Use Playwright or similar

## What I Can Do With Current Setup

✅ Navigate to any file quickly using CODEBASE.md
✅ Understand patterns from AGENTS.md conventions
✅ Make safe changes knowing testing policy
✅ Debug using server.log patterns
✅ Add new features following existing architecture
✅ Track work in TASKS.md
✅ Format code with `make format`
✅ Check code with `make check-all`
✅ Run E2E tests with `make test-e2e` (blind agent friendly)
✅ Verify UI changes with text-based assertions

## How E2E Tests Help Blind Agents

```bash
# After UI changes, run:
make test-e2e

# Output example:
Running 6 tests
✓ dashboard page loads without errors (2.3s)
✓ dashboard feed section is visible (1.1s)
✓ dashboard loading state appears initially (0.8s)
✓ dashboard console logs contain no uncaught errors (1.2s)
✓ dashboard page has accessible heading (0.5s)
✓ dashboard feed loaded state appears (1.0s)

6 passed (7.2s)
```

**Text-based checks:**
- Console errors captured as text
- Uncaught exceptions logged with message + name
- Element visibility checked via data-testid
- No screenshots needed for blind agents
- Results diffable via CLI/JSON output

## What I'd Struggle With

❌ Understanding WHY a particular architecture decision was made (no ADOPTED_DECISIONS.md)
❌ Testing complex integrations (no test DB setup)
❌ Understanding complete API contracts without reading code
❌ Request tracing across layers (no correlation IDs)

## Recommendation

**Quick Wins completed** - format and check commands added. Next, tackle JSDoc documentation and API_CONTRACTS.md for even better context. Then proceed to **Medium Effort** items as they come up. Only consider **Larger Effort** if we hit specific blockers.

The current setup (AGENTS.md + CODEBASE.md + TASKS.md + make format/check-all) is excellent for an AI assistant to work effectively with this codebase.
