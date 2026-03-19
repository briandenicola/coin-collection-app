# Ancient Coins – Agent Instructions

> This document helps AI coding agents understand and work with this codebase.
> Read this before making changes.

## What This App Does

A full-stack PWA for managing a personal ancient coin collection. Users can catalog coins with photos, track values, maintain wishlists, get AI-powered coin analysis, search for coins via a conversational agent, and view collection statistics. Multi-user with JWT + WebAuthn authentication.

## Tech Stack

| Layer    | Technology                                       | Path         |
|----------|--------------------------------------------------|--------------|
| Backend  | Go 1.26, Gin, GORM, Pure-Go SQLite (WAL mode)   | `src/api/`   |
| Frontend | Vue 3, TypeScript, Pinia, Vite, VitePWA          | `src/web/`   |
| Build    | Multi-stage Docker (Node → Go → Alpine ~40 MB)   | `Dockerfile` |
| CI/CD    | GitHub Actions → GHCR                            | `.github/`   |

## Project Layout

```
src/
├── api/                              # Go backend
│   ├── main.go                       # Entry point, route registration
│   ├── config/config.go              # Env var loading (JWT_SECRET, DB_PATH, PORT, etc.)
│   ├── database/database.go          # SQLite connection, GORM AutoMigrate
│   ├── middleware/auth.go            # JWT + API key auth middleware
│   ├── handlers/                     # HTTP handlers (one file per domain)
│   │   ├── auth.go                   # Register, login, token refresh
│   │   ├── coins.go                  # Coin CRUD, list/filter/sort, stats, value history
│   │   ├── images.go                 # Image upload/delete (multipart + base64)
│   │   ├── analysis.go              # Ollama AI coin analysis, text extraction
│   │   ├── agent.go                  # Anthropic chat agent with web_search tool
│   │   ├── journal.go               # Per-coin activity log CRUD
│   │   ├── numista.go               # Numista catalog search proxy
│   │   ├── snapshots.go             # Portfolio value snapshot recording
│   │   ├── admin.go                 # User management, settings, logs
│   │   ├── user.go                  # Password change, profile
│   │   ├── export.go                # Collection export/import
│   │   ├── api_keys.go              # API key generation/revocation
│   │   └── webauthn.go              # FIDO2/WebAuthn passwordless auth
│   ├── models/                       # GORM models
│   │   ├── coin.go                   # Coin (35+ fields, isWishlist flag)
│   │   ├── user.go                   # User (username, password hash, role)
│   │   ├── coin_journal.go           # Activity log entries
│   │   ├── value_snapshot.go         # Portfolio value snapshots
│   │   ├── api_key.go                # API keys (SHA256 hashed)
│   │   ├── appsetting.go            # Key-value app settings
│   │   ├── refresh_token.go         # JWT refresh tokens
│   │   └── webauthn_credential.go   # FIDO2 credentials
│   └── services/
│       ├── settings_service.go       # App settings with defaults (Ollama, Anthropic, Numista keys)
│       ├── ollama_service.go         # Ollama vision model integration
│       └── logger.go                 # Structured logger with in-memory buffer
│
└── web/                              # Vue 3 SPA
    └── src/
        ├── api/client.ts             # Axios HTTP client, auth interceptors, all API methods
        ├── types/index.ts            # TypeScript interfaces for all models
        ├── router/index.ts           # Vue Router (11 routes, auth guard)
        ├── stores/
        │   ├── auth.ts               # Auth state, login/logout/refresh
        │   └── coins.ts              # Coin list, current coin, stats, value history
        ├── pages/                    # Route-level components
        │   ├── CollectionPage.vue    # Main gallery (search, filter, sort, swipe/grid toggle)
        │   ├── WishlistPage.vue      # Wishlist gallery with AI search agent
        │   ├── CoinDetailPage.vue    # Full coin view, journal, Numista lookup, AI analysis
        │   ├── AddCoinPage.vue       # Create coin form
        │   ├── EditCoinPage.vue      # Edit coin form
        │   ├── StatsPage.vue         # Charts: grade distribution, value over time
        │   ├── SettingsPage.vue      # User prefs, password, WebAuthn, API keys, export/import
        │   ├── AdminPage.vue         # Users, AI config, system settings, logs
        │   ├── ImageProcessorPage.vue # OCR text extraction from store cards
        │   ├── LoginPage.vue         # Login (password + WebAuthn)
        │   └── RegisterPage.vue      # Registration
        └── components/               # Reusable components
            ├── CoinCard.vue          # Gallery card (supports wishlist variant)
            ├── CoinForm.vue          # Shared create/edit form with autocomplete
            ├── CoinSearchChat.vue    # AI agent chat drawer (Anthropic + web search)
            ├── SearchBar.vue         # Search input
            ├── CategoryFilter.vue    # Category pill filters
            ├── SortSelect.vue        # Sort dropdown (date, value)
            ├── ImageGallery.vue      # Image grid with lightbox
            ├── SwipeGallery.vue      # Mobile swipe carousel
            ├── ImageProcessor.vue    # Store card OCR upload
            └── AutocompleteInput.vue # Autocomplete text input
```

## Architecture Patterns

### Backend Conventions

- **Handler pattern**: Each domain gets its own file in `handlers/`. Handlers are structs with methods, instantiated via `NewXxxHandler()`, registered in `main.go`.
- **Route registration**: All routes are wired in `main.go` under three groups:
  - `api` (public) — auth endpoints
  - `protected` (JWT required) — all user-facing endpoints
  - `admin` (JWT + admin role) — admin-only endpoints
- **Settings**: Key-value `AppSetting` model. Constants defined in `services/settings_service.go`. `GetSetting(key)` returns DB value or hardcoded default. Never expose API keys (Anthropic, Numista) to the frontend.
- **Database**: GORM with SQLite. Schema changes happen via `AutoMigrate` in `database.go`. Add new models there.
- **Auth**: JWT (Bearer token) + API key (`X-API-Key` header). Middleware populates `userId` and `userRole` in Gin context.
- **Error responses**: `c.JSON(status, gin.H{"error": "message"})`.
- **Logging**: `services.AppLogger` — structured logger with `.Info()`, `.Warn()`, `.Error()`, `.Debug()` methods. Format: `logger.Info("category", "message %s", arg)`.

### Frontend Conventions

- **API client**: All backend calls go through `src/web/src/api/client.ts`. Export one function per endpoint. Import types from `@/types`.
- **State**: Pinia stores in `stores/`. `coins.ts` manages coin list + current coin. `auth.ts` manages authentication state.
- **Styling**: CSS variables defined in theme files. Use `var(--bg-card)`, `var(--accent-gold)`, `var(--border-subtle)`, etc. Dark theme is default.
- **Icons**: Use `lucide-vue-next` for all icons. Import specific icons by name.
- **Components**: Use `<script setup lang="ts">` with Composition API. Props via `defineProps`, emits via `defineEmits`.
- **Coin data**: The `Coin` interface has 35+ fields. Nullable number fields use `number | null`. The `sanitizeCoin()` function in `client.ts` handles empty→null conversion before sending to Go.

### Adding a New Feature (Checklist)

1. **Model** — Add GORM model in `src/api/models/`, add to `AutoMigrate` in `database/database.go`
2. **Handler** — Create handler file in `src/api/handlers/`, follow `NewXxxHandler()` pattern
3. **Routes** — Register in `main.go` under the appropriate group
4. **Settings** (if needed) — Add constant + default in `services/settings_service.go`
5. **Types** — Add TypeScript interface in `src/web/src/types/index.ts`
6. **API client** — Add method in `src/web/src/api/client.ts`
7. **Store** (if needed) — Add state/actions in relevant Pinia store
8. **UI** — Create page in `pages/` or component in `components/`, register route in `router/index.ts`

## Environment Variables

| Variable          | Default                                               | Description                          |
|-------------------|-------------------------------------------------------|--------------------------------------|
| `JWT_SECRET`      | `dev-secret-key-change-in-production-min32chars`      | JWT signing key (min 32 chars)       |
| `DB_PATH`         | `./ancientcoins.db`                                   | SQLite database file path            |
| `PORT`            | `8080`                                                | HTTP server port                     |
| `UPLOAD_DIR`      | `./uploads`                                           | Directory for uploaded coin images   |
| `WEBAUTHN_RP_ID`  | `localhost`                                           | WebAuthn Relying Party ID            |
| `WEBAUTHN_ORIGIN` | `http://localhost:8080`                               | WebAuthn origin URL                  |

## Admin-Managed Settings (stored in DB)

These are configured in the Admin UI, not environment variables:

| Key                | Purpose                                    |
|--------------------|--------------------------------------------|
| `OllamaURL`       | Ollama server URL for AI analysis          |
| `OllamaModel`     | Vision model name (e.g., `llava`)          |
| `OllamaTimeout`   | Request timeout in seconds                 |
| `AnthropicAPIKey`  | API key for Claude agent chat              |
| `AnthropicModel`   | Claude model (e.g., `claude-sonnet-4-20250514`) |
| `NumistaAPIKey`    | Numista catalog API key                    |
| `ObversePrompt`    | Custom prompt for obverse analysis         |
| `ReversePrompt`    | Custom prompt for reverse analysis         |
| `TextExtractionPrompt` | Custom prompt for OCR text extraction  |
| `LogLevel`         | Application log level                      |

## Running Locally

```sh
task run          # Starts both API (:8080) and Vite dev server (:5173)
task run-api      # API only
task run-web      # Frontend only (proxies /api/* to :8080)
```

## Building

```sh
task build        # Build both API binary and Vue dist/
task docker-build # Build Docker image
```

## Key API Endpoints

```
POST   /api/auth/register          # First user becomes admin
POST   /api/auth/login              # Returns JWT + refresh token
POST   /api/auth/refresh            # Refresh JWT

GET    /api/coins                   # List (params: category, search, wishlist, sort, order, page, limit)
POST   /api/coins                   # Create
GET    /api/coins/:id               # Get one
PUT    /api/coins/:id               # Update
DELETE /api/coins/:id               # Delete
POST   /api/coins/:id/purchase      # Move from wishlist to collection

POST   /api/coins/:id/images        # Upload image (multipart)
POST   /api/coins/:id/analyze       # AI analysis via Ollama

GET    /api/coins/:id/journal       # List journal entries
POST   /api/coins/:id/journal       # Add journal entry
DELETE /api/coins/:id/journal/:eid  # Delete journal entry

POST   /api/agent/chat              # AI coin search agent (Anthropic + web search)
GET    /api/numista/search?q=       # Numista catalog search
GET    /api/stats                   # Collection statistics
GET    /api/value-history           # Portfolio value snapshots

GET    /api/admin/settings          # All settings (admin only)
PUT    /api/admin/settings          # Update settings (admin only)
```

## Testing Notes

- Go backend: `go vet` and `go build` in `src/api/`
- Frontend: `npx vue-tsc --noEmit` for type-checking, `npm run build` for production build
- No automated test suite exists — verify changes manually
- The Go module specifies Go 1.26.1 — ensure your Go toolchain is compatible

## Commit Convention

Use conventional commit prefixes: `feat:`, `fix:`, `docs:`, `refactor:`, `style:`, `chore:`

Always include the co-author trailer:
```
Co-authored-by: Copilot <223556219+Copilot@users.noreply.github.com>
```
