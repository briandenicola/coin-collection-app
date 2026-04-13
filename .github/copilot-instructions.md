# Copilot Instructions

> Repository-level instructions for GitHub Copilot (IDE, CLI, and code review).
> This file is automatically read by Copilot on every interaction.

## Project Overview

Ancient Coins is a full-stack PWA for managing a personal ancient coin collection. Go/Gin API backend with Vue 3/TypeScript frontend, SQLite database, and a Python LangGraph multi-agent service for AI features.

| Layer | Tech | Path |
|---|---|---|
| Backend | Go 1.26, Gin, GORM, SQLite | `src/api/` |
| Frontend | Vue 3, TypeScript, Pinia, Vite, PWA | `src/web/` |
| Agent | Python 3.12, FastAPI, LangGraph, LangChain | `src/agent/` |
| Build | Multi-stage Docker (2 containers) | `Dockerfile`, `src/agent/Dockerfile` |

## Build, Test, and Lint

A [Taskfile](../Taskfile.yml) wraps common commands. Run `task --list` to see all targets.

```bash
# Go API (from src/api/)
go build ./...                          # compile
go vet ./...                            # lint
go test -v ./...                        # all tests (architecture + unit)
go test -v -run TestNoDirectDatabase .  # single test by name

# Vue frontend (from src/web/)
npm run build                           # production build (type-check + vite)
npm run type-check                      # vue-tsc only
npx vue-tsc --noEmit                    # alternative type check

# Python agent (from src/agent/)
pip install -e ".[dev]"                 # install with dev deps
ruff check app/ tests/                  # lint
pytest tests/ -v                        # all tests
pytest tests/test_foo.py::test_bar -v   # single test

# Task runner shortcuts (from repo root)
task build                              # build API + web
task test                               # Go tests
task up                                 # API + web dev servers
task up-all                             # API + web + agent dev servers
task test-agent                         # Python tests
task lint-agent                         # Python lint
```

## Architecture

See `docs/ARCHITECTURE.md` for full details.

### Go API — Layered Architecture

```
Handler → Service → Repository → Database
```

**Rules (enforced by `architecture_test.go`):**

1. **Only `main.go` imports the `database` package.** All other packages receive `*gorm.DB` or a repository/service via constructor injection.
2. **Handlers are thin.** Parse request, call service/repo, return response. No business logic, no raw SQL.
3. **Services contain business logic.** Orchestrate repos, enforce domain rules. HTTP-agnostic (no `gin.Context`).
4. **Repositories own all DB access.** Every GORM query lives in `src/api/repository/`.
5. **Multi-step writes use transactions** (`r.db.Transaction()`).
6. **Never leak internal errors to clients.** Log server-side, return generic messages.
7. **Go API contains zero LLM/agent logic.** All AI inference is proxied to the Python agent service.

**Package import rules:**

| Package | May import |
|---|---|
| `handlers/` | `services/`, `repository/`, `models/` |
| `services/` | `repository/`, `models/` |
| `repository/` | `models/`, `gorm.io/gorm` |
| `models/` | Standard library only |
| `middleware/` | `models/`, `gorm.io/gorm` |

**DI wiring in `main.go`:** `config.Load()` → `database.Connect()` → construct repos → construct services → construct handlers → register routes. Three route groups: `api` (public auth), `protected` (JWT required), `admin` (JWT + admin role).

### Multi-Agent Architecture (Python)

```
Vue SPA → Go API (8080) → Python Agent Service (8081)
```

The Python agent is a **stateless** FastAPI service — no database access. All configuration (API keys, models, prompts, user context) is passed per-request from the Go API. SSE streams flow Python → Go → Vue (Go proxies the byte stream via `services/agent_proxy.go`).

**Team pipelines:**

| Team | Pipeline |
|---|---|
| Coin Search | Search → Fetch dealer pages → Format |
| Coin Shows | Search → Verify dates are future → Format |
| Coin Analysis | Vision model analysis → Format |
| Portfolio Review | Read holdings → Valuate → Analyze |
| Availability Check | Check URLs → Analyze results → Verdict |

**Key design rules:**
- Search agents pass only tool-returned data downstream — never invented details
- Verification agents confirm every URL is live and every date is in the future
- All worker agent outputs conform to a defined Pydantic schema — no free-form text
- Top-level supervisor (`app/supervisor.py`) enforces max iteration count to prevent loops

### AI Provider Configuration

Users choose one provider in Admin Settings (`AIProvider` key):

- **Anthropic** — Claude models. Web search uses Claude's built-in `web_search_20250305` tool.
- **Ollama** — Self-hosted models. Web search uses a `create_react_agent` with SearXNG tool.

**Important:** Anthropic's `web_search` is NOT available by default on `ChatAnthropic`. Use `get_search_model()` from `app/llm/provider.py` (which calls `bind_tools`) for any agent node that needs web search. Use `get_chat_model()` for nodes that don't search.

## Code Conventions

### Go
- Constructor injection for all dependencies (`NewXxxHandler(repo, service)` pattern)
- Sentinel errors in services (e.g., `ErrNotFound`, `ErrInvalidCredentials`)
- Use GORM scopes from `repository/scopes.go` (`OwnedBy`, `OwnedByID`, `ActiveCollection`, `PublicCoins`, `ByCoinID`) instead of repeating `.Where()` clauses
- Swagger annotations on all public handler methods
- Settings use key-value `AppSetting` model; constants and defaults live in `services/settings_service.go`

### Python (Agent)
- Pydantic models for all request/response schemas (in `app/models/`)
- LangGraph `StateGraph` for team pipelines
- `create_react_agent()` for tool-using agents
- Structured logging via `app/logging_config.py` (ring buffer + stdout)

### TypeScript / Vue
- `<script setup lang="ts">` with Composition API
- **Always** use optional chaining (`?.`) and nullish coalescing (`??`) on array index access — Docker builds use stricter TS checking than local `vue-tsc`
- All API calls go through `src/web/src/api/client.ts` (Axios with JWT interceptor and 401 refresh queue)
- Agent chat streaming uses `fetch` + manual SSE parsing, not Axios
- `sanitizeCoin()` in the API client normalizes `''`/`undefined` → `null` before sending
- CSS variables: `--accent-gold`, `--bg-card`, `--border-subtle`, `--text-primary`
- Icons: `lucide-vue-next`

### UI / UX
- No emojis in UI text, prompts, or AI responses
- Dark theme is default
- PWA-compatible — test on mobile viewports

### Adding a New API Feature

1. Model in `src/api/models/` → add to `AutoMigrate` in `database/database.go`
2. Repository in `src/api/repository/*_repository.go`
3. Service (if business logic needed) in `src/api/services/*_service.go`
4. Thin handler in `src/api/handlers/` with `NewXxxHandler()` constructor
5. Wire in `src/api/main.go` (create repo → service → handler, register routes under correct group)
6. Run `go test ./...` to verify architecture rules pass

## Commit Convention

Use conventional prefixes: `feat:`, `fix:`, `docs:`, `refactor:`, `chore:`

Always include the co-author trailer:
```
Co-authored-by: Copilot <223556219+Copilot@users.noreply.github.com>
```
