# Implementation Plan: Coin Sets with Trend Tracking

**Branch**: `main` | **Date**: 2026-06-06 | **Spec**: `specs\main\spec.md`
**Input**: Feature specification derived from GitHub issue #240

**Note**: Issue #240 was used as the source because no existing `specs\main\spec.md` was present when planning began.

## Summary

Extend the current tagging system into user-owned coin sets that preserve existing tag behavior while adding set metadata, completion tracking, aggregate valuations, historical trend snapshots, smart criteria, and comparison analytics. The implementation should introduce `/api/sets` as the rich set API, keep `/api/tags` backward-compatible as an open-set facade during migration, and deliver the work in staged slices so open sets and value summaries land before defined templates, snapshots, smart sets, and advanced analytics.

## Technical Context

**Language/Version**: Go 1.26.1 API, Vue 3 + TypeScript frontend  
**Primary Dependencies**: Gin, GORM, SQLite, Pinia/Vite, existing notification and valuation scheduler patterns  
**Storage**: SQLite via GORM AutoMigrate; new set, target, snapshot, and alert tables plus migration from existing `Tag`/`CoinTag` data  
**Testing**: `go test -v ./...`, `go vet ./...`, `go build ./...` from `src\api`; `npm run build`, `npm test`, `npm run lint` from `src\web`  
**Target Platform**: Go API + Vue PWA served by existing application container  
**Project Type**: Full-stack web application with REST API and PWA frontend  
**Performance Goals**: List 50 sets over 5,000 coins within 2 seconds; trend query for one year of daily snapshots under 1 second for 100 sets; smart set preview deterministic and parameterized  
**Constraints**: Preserve current tag behavior and `/tags` compatibility; user-scope every set query; no raw SQL from user criteria; no AI/Python agent dependency; frontend must use design tokens and API client  
**Scale/Scope**: Existing personal collection scale plus planned users with dozens to hundreds of sets, thousands of coins, and daily snapshot history

## Constitution Check

*GATE: Must pass before Phase 0 research. Re-check after Phase 1 design.*

**Gate I - Layered Architecture**: PASS. Set handlers stay thin, set services own completion/snapshot/criteria business logic, and repositories own all GORM queries.

**Gate II - Dependency Injection**: PASS. New set repository, service, scheduler, and handler will be constructed in `main.go`.

**Gate III - Service Boundary Separation**: PASS. This feature is Go API + Vue only; Python agent remains uninvolved.

**Gate IV - Strict Typing & Build Parity**: PASS. New Go models/types and Vue types must pass documented build commands.

**Gate V - Design Token System**: PASS. Set dashboard/detail/wizard components must use variables and global button/chip classes.

**Gate VI - AI/Agent Isolation**: PASS. No LLM behavior is planned.

**Gate VII - Schema-Driven Contracts**: PASS. Draft OpenAPI contract is in `contracts\sets-openapi.yaml`; implementation must add Swagger annotations and frontend client types.

**Gate VIII - Conventional Commits & Workflow**: PASS. Future implementation commits must use conventional prefixes and co-author trailer.

**Gate IX - UI/UX Consistency**: PASS. UI copy avoids emojis and uses lucide icons.

**Gate X - Architecture Enforcement**: PASS. API implementation must satisfy existing architecture tests.

**Gate XI - Security Hardening**: PASS. Criteria are validated and parameterized; errors remain generic to clients.

**Gate XII - Authentication & Token Policy**: PASS. All set management endpoints are protected user-scoped routes.

**Gate XIII - PWA / Mobile Interaction Rules**: PASS. Set views must remain mobile/PWA compatible.

**Gate XIV - Social & Privacy Model**: PASS. Public sharing remains opt-in and limited; private values are not exposed publicly in v1.

**Gate XV - Supply Chain & CI Integrity**: PASS. No new dependency is mandated by this plan; any chart library must be evaluated and pinned through existing package management.

**Gate XVI - Account Lifecycle**: PASS. No account lifecycle changes.

## Project Structure

### Documentation (this feature)

```text
specs\main\
├── spec.md
├── plan.md
├── research.md
├── data-model.md
├── quickstart.md
├── contracts\
│   └── sets-openapi.yaml
└── tasks.md
```

### Source Code (repository root)
```text
src\api\
├── models\
│   └── set.go
├── repository\
│   └── set_repository.go
├── services\
│   ├── set_service.go
│   └── set_snapshot_scheduler.go
├── handlers\
│   └── sets.go
├── database\
│   └── database.go
└── main.go

src\web\src\
├── api\
│   └── client.ts
├── types\
│   └── index.ts
├── components\
│   └── sets\
│       ├── SetDashboardCard.vue
│       ├── SetCompletionChecklist.vue
│       ├── SetCreationWizard.vue
│       ├── SetTrendChart.vue
│       └── SetComparePanel.vue
├── pages\
│   ├── SetsPage.vue
│   └── SetDetailPage.vue
└── router\
    └── index.ts
```

**Structure Decision**: Use the existing full-stack web app layout. The Go API owns persistence, aggregation, scheduling, and REST contracts; the Vue frontend owns dashboard/detail/wizard/chart experiences through the existing API client.

## Phase Plan

### Phase 1 - Open sets and summaries

- Migrate existing tags into open sets while preserving `/tags` compatibility.
- Add `/api/sets` CRUD, membership, detail, and summary endpoints.
- Add dashboard and detail pages with count/value summary metrics.

### Phase 2 - Defined sets and completion

- Add templates, copied set targets, custom CSV import, completion calculation, and missing-target checklist.
- Add UI for template selection and completion visualization.

### Phase 3 - Snapshots and trends

- Add manual and scheduled set snapshots.
- Add trend endpoints and chart UI for 1m, 3m, 1y, and all ranges.

### Phase 4 - Smart sets

- Add validated criteria schema, preview endpoint, smart membership query evaluation, and rule-builder UI.

### Phase 5 - Goal sets, milestones, and projections

- Add target date, acquisition velocity, projected completion, milestone alerts, and notifications.

### Phase 6 - Advanced analytics and comparison

- Add best/worst performers, ROI comparison, CSV trend export, and multi-set compare UI.

## Post-Design Constitution Check

**Status**: PASS. The design keeps DB access in repositories, business logic in services, all routes authenticated and user-scoped, and frontend access through `client.ts`. No constitution violations require complexity tracking.

## Complexity Tracking

> **Fill ONLY if Constitution Check has violations that must be justified**

| Violation | Why Needed | Simpler Alternative Rejected Because |
|-----------|------------|-------------------------------------|
| None | N/A | N/A |
