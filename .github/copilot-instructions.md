# Copilot Instructions

> Repository-level instructions for GitHub Copilot (IDE, CLI, and code review).
> This file is automatically read by Copilot on every interaction.

## Project Overview

Ancient Coins is a full-stack PWA for managing a personal coin collection. Go/Gin API backend with Vue 3/TypeScript frontend, SQLite database, AI integrations (Ollama + Anthropic).

| Layer | Tech | Path |
|---|---|---|
| Backend | Go 1.26, Gin, GORM, SQLite | `src/api/` |
| Frontend | Vue 3, TypeScript, Pinia, Vite, PWA | `src/web/` |
| Build | Multi-stage Docker | `Dockerfile` |

## Architecture (CRITICAL)

The Go API follows a strict layered architecture. See `docs/ARCHITECTURE.md` for full details.

```
Handler → Service → Repository → Database
```

### Rules (enforced by architecture_test.go)

1. **Only `main.go` imports the `database` package.** All other packages receive `*gorm.DB` or a repository/service via constructor injection.
2. **Handlers are thin.** Parse request, call service/repo, return response. No business logic, no raw SQL.
3. **Services contain business logic.** Orchestrate repos, enforce domain rules. HTTP-agnostic (no gin.Context).
4. **Repositories own all DB access.** Every GORM query lives in `src/api/repository/`.
5. **Multi-step writes use transactions** (`r.db.Transaction()`).
6. **Never leak internal errors to clients.** Log server-side, return generic messages.

### Adding a New API Feature

1. Model in `src/api/models/` → add to AutoMigrate in `database/database.go`
2. Repository methods in `src/api/repository/*_repository.go`
3. Service logic (if needed) in `src/api/services/*_service.go`
4. Thin handler in `src/api/handlers/`
5. Wire in `src/api/main.go` (create repo → service → handler, register routes)
6. Run `go test ./...` to verify architecture rules pass

### Package Import Rules

| Package | May import |
|---|---|
| `handlers/` | `services/`, `repository/`, `models/` |
| `services/` | `repository/`, `models/` |
| `repository/` | `models/`, `gorm.io/gorm` |
| `models/` | Standard library only |
| `middleware/` | `models/`, `gorm.io/gorm` |

## Code Style

### Go
- Standard Go conventions (gofmt, go vet)
- Constructor injection for all dependencies
- Sentinel errors in services (e.g., `ErrNotFound`, `ErrInvalidCredentials`)
- Use GORM scopes from `repository/scopes.go` instead of repeating Where clauses
- Swagger annotations on all public handler methods

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

# Both
task build            # build API + web
task test             # run Go tests
task up               # run API + web dev servers
```

## Environment

| Variable | Default | Description |
|---|---|---|
| `JWT_SECRET` | (generated) | JWT signing key (min 32 chars) |
| `DB_PATH` | `./ancientcoins.db` | SQLite database path |
| `PORT` | `8080` | HTTP server port |
| `UPLOAD_DIR` | `./uploads` | Image upload directory |
| `WEBAUTHN_RP_ID` | `localhost` | WebAuthn Relying Party ID |
| `WEBAUTHN_ORIGIN` | `http://localhost:8080` | WebAuthn origin |

Admin-managed settings (Ollama URL/model, Anthropic API key/model, Numista key, prompts) are stored in the database and configured via Admin UI.

## Commit Convention

Use conventional prefixes: `feat:`, `fix:`, `docs:`, `refactor:`, `chore:`

Always include the co-author trailer:
```
Co-authored-by: Copilot <223556219+Copilot@users.noreply.github.com>
```
