# Copilot Instructions

> Repository-level instructions for GitHub Copilot (IDE, CLI, and code review).
> This file is automatically read by Copilot on every interaction.

## Project Overview

Ancient Coins is a full-stack PWA for managing a personal coin collection. Go/Gin API backend with Vue 3/TypeScript frontend, SQLite database, and a Python LangGraph multi-agent service for AI features.

| Layer | Tech | Path |
|---|---|---|
| Backend | Go 1.26, Gin, GORM, SQLite | `src/api/` |
| Frontend | Vue 3, TypeScript, Pinia, Vite, PWA | `src/web/` |
| Agent | Python 3.12, FastAPI, LangGraph, LangChain | `src/agent/` |
| Build | Multi-stage Docker (2 containers) | `Dockerfile`, `src/agent/Dockerfile` |

## Architecture (CRITICAL)

See `docs/ARCHITECTURE.md` for full details.

### Go API — Layered Architecture

```
Handler → Service → Repository → Database
```

**Rules (enforced by architecture_test.go):**

1. **Only `main.go` imports the `database` package.** All other packages receive `*gorm.DB` or a repository/service via constructor injection.
2. **Handlers are thin.** Parse request, call service/repo, return response. No business logic, no raw SQL.
3. **Services contain business logic.** Orchestrate repos, enforce domain rules. HTTP-agnostic (no gin.Context).
4. **Repositories own all DB access.** Every GORM query lives in `src/api/repository/`.
5. **Multi-step writes use transactions** (`r.db.Transaction()`).
6. **Never leak internal errors to clients.** Log server-side, return generic messages.
7. **Go API contains zero LLM/agent logic.** All AI inference is proxied to the Python agent service.

### Multi-Agent Architecture (Python)

```
Vue SPA → Go API (8080) → Python Agent Service (8081)
```

The Python agent is a **stateless** FastAPI service — it has no database access. All configuration (API keys, models, prompts, user context) is passed per-request from the Go API.

**Team Structure:**

| Team | Purpose | Pipeline |
|---|---|---|
| Team 1: Coin Search | Find coins for sale | Search → Verify URLs live/unsold → Format |
| Team 2: Coin Shows | Find upcoming events | Search → Verify dates future → Format |
| Team 3: Coin Analysis | Analyze coin images | Vision model analysis → Format |
| Team 4: Portfolio Review | Assess collection | Read holdings → Valuate (via Team 1) → Analyze |

**Key Design Rules:**
- Search agents pass only tool-returned data downstream — never invented details
- Verification agents confirm every URL is live and every date is in the future
- All worker agent outputs conform to a defined schema — no free-form text
- Top-level supervisor enforces max iteration count to prevent loops

### AI Provider Configuration

Users explicitly choose one AI provider in Admin Settings:

- **Anthropic (Recommended)** — Claude models with built-in `web_search` tool
- **Ollama** — Self-hosted models, requires external SearXNG for web search

The `AIProvider` setting must be set before agent features work. The agent chat shows a configuration banner when it's empty. The `resolveLLMConfig()` helper in `handlers/agent.go` reads the explicit setting — there is no implicit fallback.

### Package Import Rules

| Package | May import |
|---|---|
| `handlers/` | `services/`, `repository/`, `models/` |
| `services/` | `repository/`, `models/` |
| `repository/` | `models/`, `gorm.io/gorm` |
| `models/` | Standard library only |
| `middleware/` | `models/`, `gorm.io/gorm` |

### Adding a New API Feature

1. Model in `src/api/models/` → add to AutoMigrate in `database/database.go`
2. Repository methods in `src/api/repository/*_repository.go`
3. Service logic (if needed) in `src/api/services/*_service.go`
4. Thin handler in `src/api/handlers/`
5. Wire in `src/api/main.go` (create repo → service → handler, register routes)
6. Run `go test ./...` to verify architecture rules pass

## Code Style

### Go
- Standard Go conventions (gofmt, go vet)
- Constructor injection for all dependencies
- Sentinel errors in services (e.g., `ErrNotFound`, `ErrInvalidCredentials`)
- Use GORM scopes from `repository/scopes.go` instead of repeating Where clauses
- Swagger annotations on all public handler methods

### Python (Agent)
- Pydantic models for all request/response schemas
- LangGraph `StateGraph` for team pipelines
- `create_react_agent()` for tool-using agents
- Structured logging via `app/logging_config.py` (ring buffer + stdout)
- Ruff for linting, pytest for tests

### TypeScript / Vue
- `<script setup lang="ts">` with Composition API
- Always use optional chaining (`?.`) and nullish coalescing (`??`) on array index access
- Docker build uses stricter TS checking than local vue-tsc
- All API calls go through `src/web/src/api/client.ts`
- CSS variables: `--accent-gold`, `--bg-card`, `--border-subtle`, `--text-primary`
- Icons: `lucide-vue-next`

### UI / UX
- No emojis in UI text, prompts, or AI responses
- Dark theme is default
- PWA-compatible — test on mobile viewports

## Build and Test

```bash
# Go API
cd src/api
go build ./...        # compile
go vet ./...          # lint
go test -v ./...      # architecture tests

# Vue frontend
cd src/web
npm run build         # production build
npx vue-tsc --noEmit  # type check

# Python agent
cd src/agent
ruff check app/ tests/  # lint
pytest tests/ -q         # tests

# All
task build            # build API + web
task test             # run Go tests
task up               # run API + web dev servers
```

## Deployment

Two Docker containers orchestrated via `docker-compose.yaml`:

| Container | Image | Port | Purpose |
|---|---|---|---|
| `app` | `<user>/ancient-coins:latest` | 8080 | Go API + Vue SPA |
| `agent` | `<user>/ancient-coins-agent:latest` | 8081 | Python LangGraph agent |

External services (not deployed by us):
- **Ollama** — self-hosted LLM (only if Ollama provider selected)
- **SearXNG** — web search engine (only needed for Ollama mode)

CI builds and pushes both images via GitHub Actions. `docker-compose` only references images — no local builds.

## Environment

| Variable | Default | Description |
|---|---|---|
| `JWT_SECRET` | (generated) | JWT signing key (min 32 chars) |
| `DB_PATH` | `./ancientcoins.db` | SQLite database path |
| `PORT` | `8080` | HTTP server port |
| `UPLOAD_DIR` | `./uploads` | Image upload directory |
| `WEBAUTHN_RP_ID` | `localhost` | WebAuthn Relying Party ID |
| `WEBAUTHN_ORIGIN` | `http://localhost:8080` | WebAuthn origin |
| `AGENT_SERVICE_URL` | `http://agent:8081` | Python agent service URL |
| `AGENT_LOG_LEVEL` | `INFO` | Python agent log level |

### Admin-Managed Settings (stored in DB)

| Key | Purpose |
|---|---|
| `AIProvider` | Explicit provider choice: `anthropic` or `ollama` (empty = unconfigured) |
| `AnthropicAPIKey` | API key for Claude models |
| `AnthropicModel` | Claude model (e.g., `claude-sonnet-4-20250514`) |
| `OllamaURL` | Ollama server URL |
| `OllamaModel` | Vision model name (e.g., `llava`) |
| `OllamaTimeout` | Request timeout in seconds |
| `SearXNGURL` | SearXNG search engine URL (required for Ollama web search) |
| `NumistaAPIKey` | Numista catalog API key |
| `AgentPrompt` | System prompt for coin search agent |
| `ValuationPrompt` | System prompt for value estimator |
| `ObversePrompt` | Prompt for obverse image analysis |
| `ReversePrompt` | Prompt for reverse image analysis |
| `TextExtractionPrompt` | Prompt for OCR text extraction |
| `LogLevel` | Application log level |

## Commit Convention

Use conventional prefixes: `feat:`, `fix:`, `docs:`, `refactor:`, `chore:`

Always include the co-author trailer:
```
Co-authored-by: Copilot <223556219+Copilot@users.noreply.github.com>
```
