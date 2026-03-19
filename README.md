# Ancient Coins

> **Note:** This application is 100% vibe coded. It's exclusively for me to learn and experiment with GitHub Copilot CLI.

Ancient Coins is a full-stack web application for cataloging and managing your personal coin collection. Track details like denomination, ruler, era, mint, material, grade, inscriptions, RIC rarity ratings, photos, and more — with AI-powered coin analysis via Ollama vision models and an Anthropic-powered coin search agent. Every coin is scoped to your authenticated account using JWT-based authentication.

It includes a **wish list** with an AI search agent for discovering coins, a **stats dashboard** with grade distribution charts and portfolio value tracking over time, per-coin **activity journals**, **Numista catalog lookups**, and collection **export/import**.

On first launch, the first user to register is automatically assigned as the admin and can configure application settings including AI integrations.

## Architecture

| Layer    | Tech                                  | Path       |
| -------- | ------------------------------------- | ---------- |
| Backend  | Go (Gin), GORM, Pure-Go SQLite       | `src/api/` |
| Frontend | Vue 3, TypeScript, Vite, Pinia (PWA)  | `src/web/` |

The Vue SPA communicates with the Go API exclusively via REST (`/api/*`). In production the API serves the SPA as static files from a single container, so no separate web server is needed.

The frontend is a Progressive Web App (PWA) and can be installed on iOS (Safari → Share → Add to Home Screen), Android, and desktop browsers for a native app-like experience with offline caching.

### Development (two processes, two ports)

```
Browser → localhost:5173 (Vite dev server, serves Vue SPA)
           └─ /api/* requests → proxied to localhost:8080 (via vite.config.ts)

Browser → localhost:8080 (Go/Gin API, serves REST endpoints)
```

In development, the Vite dev server runs on `:5173` with hot-reload and proxies any `/api/*` or `/uploads/*` request to the Go API on `:8080`. The browser only talks to the Vite server.

### Production (single container, single port)

The Dockerfile uses a multi-stage build to combine both into one image:

```
Stage 1: node:24-alpine     → npm run build  → produces dist/ (static HTML/JS/CSS)
Stage 2: golang:1.26-alpine → go build       → produces ancient-coins-api binary
Stage 3: alpine:3.21        → copies both:
           /app/ancient-coins-api    (Go binary)
           /app/wwwroot/             (Vue dist/ output)
```

The Go binary serves the Vue SPA as static files **and** handles API routes — one process does both jobs:

```
Browser → localhost:8080 → Go binary inside container
              ├─ /api/*      → Gin REST handlers
              ├─ /uploads/*  → serves uploaded images from volume
              └─ /*          → serves Vue SPA from /app/wwwroot/
```

No nginx or reverse proxy needed. Docker volumes persist the SQLite database and uploaded images across container restarts.

## Prerequisites

- [Go](https://go.dev/dl/) (1.22+)
- [Node.js](https://nodejs.org/) (v20+)
- [Task](https://taskfile.dev/) — optional, for task runner commands
- [Docker & Docker Compose](https://docs.docker.com/get-docker/) — optional, for containerized deployment
- [Ollama](https://ollama.ai/) — optional, for AI coin analysis

## Getting Started

Clone the repository and start the development servers:

```sh
git clone <repo-url> && cd AncientCoins
task run        # starts both API and frontend in parallel
```

The API runs on `http://localhost:8080` and the Vite dev server on `http://localhost:5173`. You can also start them individually:

```sh
task run-api    # API only
task run-web    # frontend only
```

When the app launches for the first time, register your first account — it is automatically assigned as the admin. You can then configure Ollama and other settings from the Admin page.

For a detailed walkthrough of first-time setup, adding coins, import/export, and AI analysis, see the [Getting Started Guide](docs/getting-started.md).

## Task Commands

| Command            | Description                              |
| ------------------ | ---------------------------------------- |
| `task init`        | Generate a `.env` file with a random JWT secret |
| `task run`         | Run API and frontend in parallel         |
| `task run-api`     | Run the Go API server                    |
| `task run-web`     | Run the Vite dev server                  |
| `task build`       | Build both API and frontend              |
| `task build-api`   | Build the Go API binary                  |
| `task build-web`   | Build the Vue frontend                   |
| `task docker-build`| Build the Docker container image         |
| `task docker-run`  | Run the Docker container locally         |

## CI/CD

A GitHub Actions workflow (`.github/workflows/docker-publish.yml`) builds and pushes the Docker image to Docker Hub:

- **Triggers** — Push to `main` branch and manual dispatch
- **Tags** — Full commit SHA, short commit SHA, and `latest` (on main)
- **Caching** — Uses GitHub Actions cache for Docker layer caching

### Required Secrets

| Secret | Description |
| ------ | ----------- |
| `DOCKERHUB_USERNAME` | Your Docker Hub username |
| `DOCKERHUB_TOKEN` | A Docker Hub access token |

## Deployment

The application ships as a single Docker container that serves both the API and the Vue SPA. The pre-built image is hosted on Docker Hub.

### Configuration

Generate a `.env` file with a random JWT signing key:

```sh
task init
```

This creates a `.env` file with a cryptographically random `JWT_SECRET`:

```
JWT_SECRET=<auto-generated>
```

### Running with Docker Compose

```sh
docker compose up
```

### Running with Task

```sh
task docker-build           # build the image from source
task docker-run             # run the container (reads .env automatically)
```

### Running with Docker directly

```sh
docker run -p 8080:8080 -d \
  -e JWT_SECRET=YourSecretKeyHere \
  -v ancient-coins-data:/app/data \
  -v ancient-coins-uploads:/app/uploads \
  ghcr.io/briandenicola/ancient-coins:latest
```

The application will be available on `http://localhost:8080`. On first launch, the first user to register becomes the admin.

## Features

### Collection Management

The main collection page supports browsing your coins with filtering, search, and sorting:

- **Card Gallery** — Responsive card grid showing each coin's primary image with name, ruler, denomination, and category. Supports filtering by category (Roman, Greek, Byzantine, Modern) and full-text search across all fields.
- **Sorting** — Sort by date added, date last updated, or current value (ascending or descending).
- **Swipe / Grid Toggle** — Switch between a swipeable card carousel (ideal for mobile/PWA) and a traditional grid layout. Default view preference is configurable in Settings.
- **Pagination** — Coins are loaded with pagination for large collections.
- **Category Colors** — Each category has a distinct color accent: purple for Roman, olive for Greek, red for Byzantine, and steel blue for Modern.

### Wish List

Track coins you'd like to acquire with an AI-powered search agent:

- **Add to Wish List** — When creating a coin, toggle the "Wishlist" flag. The coin appears in the dedicated Wish List view instead of the main collection.
- **AI Coin Search Agent** — Click "Find Coins" to open a chat drawer powered by Anthropic Claude with web search. Describe the coins you're looking for (e.g., "Roman silver denarii of Julius Caesar under $500") and the agent searches the web for real listings and references. Results appear as cards with metadata, estimated prices, and source links — each with an "Add to Wishlist" button for one-click import.
- **Purchase** — Move a wishlist coin to your main collection when you acquire it.
- **Wish List Gallery** — A separate page showing only wishlist items with sorting support.

### Coin Details

Each coin can store:

- **Core fields** — Name, denomination, ruler/authority, year/era, category, material, weight, diameter
- **Numismatic details** — Mint mark, obverse/reverse inscriptions, grade, RIC number, rarity rating
- **Financial data** — Purchase price, current value, acquisition date, dealer/source
- **Images** — Multiple image uploads per coin (obverse, reverse, edge, detail, full) with a gallery viewer
- **AI Analysis** — Markdown-formatted analysis from Ollama, stored with the coin (obverse and reverse analyzed separately)
- **Activity Journal** — Timestamped log entries per coin (e.g., "cleaned", "sent to NGC for grading", "displayed at coin show"). Add and delete entries directly from the detail page.
- **Numista Catalog Lookup** — Search the Numista coin catalog directly from a coin's detail page. Results show thumbnails, title, issuer, and year range with links to the full Numista catalog entry.
- **Notes** — Free-text notes field

### AI Coin Analysis

Upload photos of a coin and click **Analyze with AI** to get an AI-powered numismatic analysis via Ollama. The analysis covers identification, obverse/reverse descriptions, inscriptions, condition assessment, historical context, and estimated market value. Obverse and reverse sides are analyzed separately with dedicated prompts. If accepted, the analysis is saved to the coin's record.

To enable AI analysis:

1. Install [Ollama](https://ollama.ai/) and pull a vision model: `ollama pull llava`
2. Start Ollama: `ollama serve`
3. Configure the Ollama URL and model in **Admin → AI Configuration**

### AI Coin Search Agent

Chat with an AI agent that searches the web for coins matching your description. Powered by Anthropic Claude with the web search tool, the agent returns structured coin suggestions with names, categories, materials, price estimates, and source links. Each suggestion can be added to your wishlist with one click.

To enable the search agent:

1. Get an API key from [console.anthropic.com](https://console.anthropic.com/)
2. Configure it in **Admin → AI Configuration → Anthropic API Key**

### Numista Catalog Integration

Look up coins in the [Numista](https://en.numista.com/) catalog directly from any coin's detail page. The search uses the coin's name, denomination, and ruler to find matching entries, displaying thumbnails and linking to full catalog pages.

To enable Numista lookup:

1. Get a free API key at [numista.com/api](https://en.numista.com/api/) (2,000 requests/month free)
2. Configure it in **Admin → AI Configuration → Numista API Key**

### Collection Statistics

The **Stats** page shows:

- **Summary cards** — Total coins, total value, average value, unique rulers
- **Category breakdown** — Distribution across Roman, Greek, Byzantine, Modern, and Other
- **Material distribution** — Breakdown by gold, silver, bronze, copper, and electrum
- **Grade distribution** — Bar chart showing coin counts by grade (VF, EF, AU, etc.) with a blue gradient
- **Value over time** — SVG line chart tracking total portfolio value and total invested over time, built from automatic snapshots recorded after every coin create, update, or delete
- **Top coins by value** — Ranked list of the most valuable coins in your collection

### Image Text Extraction

Upload photos of store cards, certificates, or coin holder labels and extract text via OCR powered by Ollama. Useful for quickly capturing dealer information, grades, and reference numbers when adding coins.

### User Settings

All authenticated users can access **Settings** to configure:

- **Change Password** — Update your account password (requires current password).
- **Theme** — Switch between dark (museum) and light mode. Persists across sessions.
- **Default View** — Choose between swipe carousel or grid layout for the collection gallery on mobile/PWA.
- **Time Zone** — Select your preferred time zone for date/time display.
- **WebAuthn / Passkeys** — Register FIDO2 credentials for passwordless login.
- **API Keys** — Generate and manage API keys for programmatic access.
- **Export / Import** — Export your entire collection as JSON, or import coins from a JSON file. See the [Getting Started Guide](docs/getting-started.md#import--export) for the full import file format and field reference.

### Admin Settings

The first registered user is the admin. Admins can access **Admin** to manage:

- **Users** — View all registered users, delete accounts, and reset passwords.
- **AI Configuration** — Configure Ollama (URL, vision model, timeout, analysis prompts), Anthropic (API key, model for search agent), and Numista (API key for catalog lookups).
- **System** — Set the application log level (trace, debug, info, warn, error).
- **Logs** — View real-time application logs with level filtering and auto-refresh.

### Environment Variables

| Variable          | Default                         | Description |
| ----------------- | ------------------------------- | ----------- |
| `JWT_SECRET`      | *(auto-generated for dev)*      | JWT signing key (`openssl rand -base64 48`) |
| `DB_PATH`         | `./ancientcoins.db`             | SQLite database file path |
| `PORT`            | `8080`                          | HTTP server port |
| `UPLOAD_DIR`      | `./uploads`                     | Directory for uploaded coin images |
| `WEBAUTHN_RP_ID`  | `localhost`                     | Relying Party ID for FIDO2/WebAuthn |
| `WEBAUTHN_ORIGIN` | `http://localhost:8080`         | Origin URL for WebAuthn |

#### Admin-Managed Settings

These are configured in the Admin UI (not environment variables) and stored in the database:

| Setting              | Description |
| -------------------- | ----------- |
| `OllamaURL`          | Ollama server URL for AI coin analysis (default: `http://localhost:11434`) |
| `OllamaModel`        | Vision model name (default: `llava`) |
| `OllamaTimeout`      | AI request timeout in seconds (default: `300`) |
| `AnthropicAPIKey`     | API key for the Claude-powered coin search agent |
| `AnthropicModel`      | Claude model to use (default: `claude-sonnet-4-20250514`) |
| `NumistaAPIKey`       | Numista catalog API key for coin lookups |
| `ObversePrompt`       | Custom prompt for obverse image analysis |
| `ReversePrompt`       | Custom prompt for reverse image analysis |
| `TextExtractionPrompt`| Custom prompt for OCR text extraction |
| `LogLevel`            | Application log level (trace/debug/info/warn/error) |

## Project Structure

```
AncientCoins/
├── .devcontainer/                    # Dev container configuration
│   ├── Dockerfile
│   ├── devcontainer.json
│   └── post-create.sh
├── src/
│   ├── api/                          # Go backend
│   │   ├── main.go                   # App entry point & route wiring
│   │   ├── config/                   # Environment-based configuration
│   │   ├── database/                 # SQLite connection (pure-Go driver)
│   │   ├── handlers/                 # HTTP handlers
│   │   │   ├── auth.go               # Registration, login, token refresh
│   │   │   ├── coins.go              # Coin CRUD, list/filter/sort, stats, value history
│   │   │   ├── images.go             # Image upload/delete
│   │   │   ├── analysis.go           # Ollama AI coin analysis & OCR
│   │   │   ├── agent.go              # Anthropic chat agent with web search
│   │   │   ├── journal.go            # Per-coin activity log
│   │   │   ├── numista.go            # Numista catalog search proxy
│   │   │   ├── snapshots.go          # Portfolio value snapshots
│   │   │   ├── admin.go              # User/settings management
│   │   │   ├── user.go               # Password change, profile
│   │   │   ├── export.go             # Collection export/import
│   │   │   ├── api_keys.go           # API key management
│   │   │   └── webauthn.go           # FIDO2/WebAuthn auth
│   │   ├── middleware/               # JWT & API key auth middleware
│   │   ├── models/                   # GORM entities (Coin, User, CoinJournal, ValueSnapshot, etc.)
│   │   └── services/                 # Business logic (Ollama, settings, logger)
│   └── web/                          # Vue 3 SPA
│       ├── src/
│       │   ├── api/                  # Axios API client
│       │   ├── assets/styles/        # CSS variables & global styles
│       │   ├── components/           # Reusable components
│       │   │   ├── CoinCard.vue      # Gallery card (collection + wishlist variants)
│       │   │   ├── CoinForm.vue      # Shared create/edit form with autocomplete
│       │   │   ├── CoinSearchChat.vue # AI agent chat drawer
│       │   │   ├── SearchBar.vue     # Search input
│       │   │   ├── CategoryFilter.vue # Category pill filters
│       │   │   ├── SortSelect.vue    # Sort dropdown
│       │   │   ├── ImageGallery.vue  # Image grid with lightbox
│       │   │   ├── SwipeGallery.vue  # Mobile swipe carousel
│       │   │   ├── ImageProcessor.vue # Store card OCR upload
│       │   │   └── AutocompleteInput.vue
│       │   ├── pages/                # Route pages
│       │   ├── stores/               # Pinia stores (auth, coins)
│       │   ├── router/               # Vue Router configuration
│       │   └── types/                # TypeScript type definitions
│       ├── public/                   # PWA icons & coin logo
│       └── vite.config.ts
├── docs/
│   └── getting-started.md            # User walkthrough guide
├── instructions.md                   # Agent instructions for AI coding assistants
├── Dockerfile                        # Multi-stage build (Vue + Go → Alpine)
├── Taskfile.yml                      # Task runner configuration
├── docker-compose.yaml               # Container orchestration
└── README.md
```

## Backlog

Feature ideas and completed enhancements:

- [x] **CI/CD Pipeline** — GitHub Actions workflow to build and push Docker image
- [x] **Sorting** — Sort coins by date added, date updated, or value
- [x] **Swipe / Grid Toggle** — Mobile-friendly view preference with PWA support
- [x] **PWA Viewport Stability** — Fixed scrolling/interaction wobble in installed PWA
- [x] **Grade Distribution Chart** — Bar chart of coins by grade
- [x] **Value Over Time Chart** — SVG line chart tracking portfolio value and investment
- [x] **Activity Journal** — Per-coin timestamped activity log
- [x] **Numista Catalog Lookup** — Search the Numista coin database from detail pages
- [x] **AI Coin Search Agent** — Anthropic-powered chat agent with web search for discovering coins
- [ ] **Wear Heatmap** — GitHub-style heatmap for tracking when coins were viewed or handled
- [ ] **Collection Timeline** — Visual timeline of when each coin was acquired
- [ ] **Coin Comparison** — Side-by-side spec comparison of any two coins
- [ ] **Advanced Search** — Filter by date range, price range, grade, material
- [ ] **Batch Import** — Import coins from CSV or numismatic database exports
- [ ] **PWA Icons** — Generate proper PWA icons from the EID MAR coin logo
- [ ] **Price Alerts** — Notifications when watched coins appear below a target price
- [ ] **Share Collection** — Public shareable link for a subset of your collection
- [ ] **Camera Capture** — Take coin photos directly in the app with cropping

## License

This project is licensed under the [MIT License](LICENSE).
