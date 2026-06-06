# Quickstart: Coin Sets with Trend Tracking

## Prerequisites

- Go API dependencies installed in `src\api`
- Vue dependencies installed in `src\web`
- Task runner available from the repository root

## Backend development flow

1. Add or migrate set models in `src\api\models\`.
2. Register new models in `src\api\database\database.go`.
3. Add `SetRepository` methods in `src\api\repository\` for set CRUD, membership, summaries, templates, snapshots, trends, comparison, criteria validation, and smart query evaluation.
4. Add `SetService` methods in `src\api\services\` for completion calculations, snapshot aggregation, comparison metrics, milestone alerts, and smart criteria validation.
5. Add thin handlers in `src\api\handlers\sets.go` with Swagger annotations.
6. Wire repository, service, scheduler, and protected routes in `src\api\main.go`.
7. Keep existing `/tags` endpoints as compatibility wrappers over open sets until all UI surfaces are migrated.

## Frontend development flow

1. Extend set-related types in `src\web\src\types\index.ts`.
2. Add set API functions in `src\web\src\api\client.ts`.
3. Add set dashboard, set detail, creation wizard, completion checklist, trend chart, and compare components under `src\web\src\components\sets\` and pages under `src\web\src\pages\`.
4. Update settings tag management to use set language where appropriate while preserving existing tag flows during migration.
5. Use design tokens and global `.btn`, `.chip`, `.chip-sm`, and `.badge` classes.
6. Avoid emojis in UI copy; use `lucide-vue-next` icons.

## Validation commands

From `src\api`:

```powershell
go test -v ./...
go vet ./...
go build ./...
```

From `src\web`:

```powershell
npm run build
npm test
npm run lint
```

From repository root:

```powershell
task build
task test
```

## Manual smoke test

1. Start the app with `task up`.
2. Sign in as a user with existing tagged coins.
3. Open the sets dashboard and confirm existing tags appear as open sets.
4. Create a new open set and add two coins.
5. Confirm set detail shows coin count, total current value, and average value.
6. Create a manual snapshot for the set and confirm trend data includes today's point.
7. Create a defined set from a template and confirm missing targets appear.
8. Compare two sets and confirm value and percentage changes are displayed.

## Validation record

2026-06-06 final implementation cleanup:

- `task openapi` completed successfully on Windows after making the copy/check steps platform-specific.
- From `src\api`: `go test ./...`, `go vet ./...`, and `go build ./...` completed successfully.
- From `src\web`: `npm.cmd test`, `npm.cmd run lint`, and `npm.cmd run build` completed successfully.
- Focused review confirmed set repository queries are user-scoped and smart criteria use whitelisted fields plus parameter binding.
- Focused review confirmed new set UI uses design-token patterns and no emoji UI copy.
