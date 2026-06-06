# Tasks: Coin Sets with Trend Tracking

**Input**: Design documents from `specs\main\`
**Prerequisites**: `plan.md`, `spec.md`, `research.md`, `data-model.md`, `contracts\sets-openapi.yaml`, `quickstart.md`

**Tests**: Test implementation tasks are not listed separately because TDD was not requested; validation tasks are included in the final phase.

**Organization**: Tasks are grouped by user story so each story can be implemented and tested independently after the foundational phase.

## Format: `[ID] [P?] [Story] Description`

- **[P]**: Can run in parallel because it touches different files and has no dependency on incomplete tasks in the same phase
- **[Story]**: Maps to user stories from `specs\main\spec.md`
- All task descriptions include exact repository paths

---

## Phase 1: Setup (Shared Infrastructure)

**Purpose**: Prepare the codebase for set implementation without changing runtime behavior.

- [X] T001 Review the draft API contract and align endpoint names in `specs\main\contracts\sets-openapi.yaml`
- [X] T002 [P] Create backend set implementation files `src\api\models\set.go`, `src\api\repository\set_repository.go`, `src\api\services\set_service.go`, `src\api\handlers\sets.go`
- [X] T003 [P] Create frontend set component directory and placeholder files in `src\web\src\components\sets\`
- [X] T004 [P] Create frontend set page files `src\web\src\pages\SetsPage.vue` and `src\web\src\pages\SetDetailPage.vue`
- [X] T005 [P] Add a route planning placeholder for set pages in `src\web\src\router\index.ts`

---

## Phase 2: Foundational (Blocking Prerequisites)

**Purpose**: Define shared data structures, repository/service boundaries, and compatibility seams required before any user story implementation.

**CRITICAL**: No user story work can begin until this phase is complete.

- [X] T006 Define `CoinSetType`, `CoinSet`, `CoinSetMembership`, `CoinSetTarget`, `CoinSetValuationSnapshot`, and `CoinSetMilestoneAlert` models in `src\api\models\set.go`
- [X] T007 Register set models in `AutoMigrate` and preserve existing tag migration order in `src\api\database\database.go`
- [X] T008 Implement `CoinSet` migration/backfill helpers from existing `Tag` and `CoinTag` data in `src\api\repository\set_repository.go`
- [X] T009 Implement user-scoped `CoinSet` CRUD repository methods with `OwnedBy` and `OwnedByID` scopes in `src\api\repository\set_repository.go`
- [X] T010 Implement manual membership repository methods with same-user coin/set validation in `src\api\repository\set_repository.go`
- [X] T011 Implement shared set summary aggregate queries for coin count, total value, total invested, average value, and highest-value coin in `src\api\repository\set_repository.go`
- [X] T012 Implement set validation helpers for name uniqueness, set type rules, color/icon limits, and parent cycle prevention in `src\api\services\set_service.go`
- [X] T013 Implement compatibility mapping between legacy `Tag` payloads and open `CoinSet` records in `src\api\services\set_service.go`
- [X] T014 Update existing tag repository or compatibility facade to read/write open sets without breaking `/tags` responses in `src\api\repository\tag_repository.go`
- [X] T015 Update existing tag handler compatibility behavior for `/tags` and `/coins/:id/tags` in `src\api\handlers\tag.go`
- [X] T016 Define shared TypeScript set types and extend `Tag` compatibility types in `src\web\src\types\index.ts`
- [X] T017 Add base set API client functions for `/sets` endpoints in `src\web\src\api\client.ts`
- [X] T018 Wire `SetRepository`, `SetService`, and protected `/sets` routes in `src\api\main.go`
- [X] T019 Add Swagger type definitions for set request and response payloads in `src\api\handlers\swagger_types.go`

**Checkpoint**: Foundation ready; user stories can now proceed in priority order or in parallel where staffing allows.

---

## Phase 3: User Story 1 - Manage sets evolved from tags (Priority: P1) MVP

**Goal**: Users can see existing tags as open sets, create and edit sets, add/remove coins, view summary totals, and continue filtering collection by set/tag.

**Independent Test**: Sign in with tagged coins, open the sets dashboard, confirm existing tags appear as open sets, create a new open set, add two coins, view set details with count/value totals, and filter the collection by that set.

### Implementation for User Story 1

- [X] T020 [US1] Implement set list, create, get, update, and delete service methods for open sets in `src\api\services\set_service.go`
- [X] T021 [US1] Implement set list, create, get, update, and delete HTTP handlers with Swagger annotations in `src\api\handlers\sets.go`
- [X] T022 [US1] Implement list/create/get/update/delete frontend API calls in `src\web\src\api\client.ts`
- [X] T023 [P] [US1] Implement `SetDashboardCard` summary UI using design tokens in `src\web\src\components\sets\SetDashboardCard.vue`
- [X] T024 [US1] Implement sets dashboard loading, empty state, create action, and card grid in `src\web\src\pages\SetsPage.vue`
- [X] T025 [US1] Implement set detail loading, summary header, and coin grid/list area in `src\web\src\pages\SetDetailPage.vue`
- [X] T026 [US1] Implement set creation and edit wizard for open sets in `src\web\src\components\sets\SetCreationWizard.vue`
- [X] T027 [US1] Implement add/remove coin membership service methods for manual sets in `src\api\services\set_service.go`
- [X] T028 [US1] Implement `GET /sets/{id}/coins`, `POST /sets/{id}/coins`, and `DELETE /sets/{id}/coins/{coinId}` handlers in `src\api\handlers\sets.go`
- [X] T029 [US1] Implement add/remove coin membership frontend API calls in `src\web\src\api\client.ts`
- [X] T030 [US1] Wire set pages into authenticated navigation and routes in `src\web\src\router\index.ts`
- [x] T031 [US1] Update collection filtering to treat selected tags as open sets while preserving existing filter behavior in `src\web\src\composables\useCollectionFilters.ts`
- [x] T032 [US1] Update desktop and PWA collection headers to use compatible set/tag labels without breaking current tag selection in `src\web\src\components\collection\DesktopCollectionHeader.vue` and `src\web\src\components\collection\PwaCollectionHeader.vue`
- [x] T033 [US1] Update coin detail tag controls to use open-set membership APIs while preserving current tag UI behavior in `src\web\src\components\coin\CoinTagsSection.vue`
- [x] T034 [US1] Update settings data tag management to present open sets and maintain legacy tag CRUD compatibility in `src\web\src\components\settings\SettingsDataSection.vue`
- [x] T035 [US1] Update bulk tag application to use open-set membership compatibility in `src\web\src\pages\CollectionPage.vue`

**Checkpoint**: User Story 1 is a usable MVP that preserves current tags and adds open set dashboard/detail summaries.

---

## Phase 4: User Story 2 - Track completion for defined sets (Priority: P2)

**Goal**: Users can create defined or goal sets from templates or imported targets, then view completion percentage and missing targets.

**Independent Test**: Create a defined set from a template, confirm target count and missing list, add a matching coin, and confirm completion increases deterministically.

### Implementation for User Story 2

- [X] T036 [P] [US2] Add built-in set template definitions for popular US series in `src\api\services\set_templates.go`
- [X] T037 [US2] Implement template list and copy-to-target repository methods in `src\api\repository\set_repository.go`
- [X] T038 [US2] Implement defined and goal set creation from template in `src\api\services\set_service.go`
- [X] T039 [US2] Implement custom CSV target import parsing and duplicate target validation in `src\api\services\set_target_import.go`
- [X] T040 [US2] Implement deterministic target-to-coin matching and completion calculation in `src\api\services\set_completion.go`
- [X] T041 [US2] Implement target list, missing target, and completion repository queries in `src\api\repository\set_repository.go`
- [X] T042 [US2] Implement `GET /sets/templates` and `GET /sets/{id}/completion` handlers with Swagger annotations in `src\api\handlers\sets.go`
- [X] T043 [US2] Add template, target import, and completion frontend API calls in `src\web\src\api\client.ts`
- [X] T044 [US2] Extend TypeScript set target, template, and completion types in `src\web\src\types\index.ts`
- [x] T045 [US2] Extend set creation wizard with set type selection, template selection, target date, and import options in `src\web\src\components\sets\SetCreationWizard.vue`
- [x] T046 [P] [US2] Implement completion checklist and missing target UI in `src\web\src\components\sets\SetCompletionChecklist.vue`
- [x] T047 [US2] Add completion panel to set detail page for defined and goal sets in `src\web\src\pages\SetDetailPage.vue`
- [x] T048 [US2] Add completion percentage and target count display to set dashboard cards in `src\web\src\components\sets\SetDashboardCard.vue`

**Checkpoint**: User Story 2 independently supports defined/goal completion tracking on top of the set foundation.

---

## Phase 5: User Story 3 - Monitor value trends for sets (Priority: P3)

**Goal**: Users can capture and view set valuation snapshots, compare set performance, and receive milestone notifications.

**Independent Test**: Capture a manual snapshot for a set, view trend data for a selected range, compare two sets with snapshots, and verify milestone notification behavior.

### Implementation for User Story 3

- [x] T049 [US3] Implement snapshot creation, same-day recapture, range filtering, and comparison repository methods in `src\api\repository\set_repository.go`
- [x] T050 [US3] Implement set snapshot aggregation, trend range selection, ROI metrics, and comparison calculations in `src\api\services\set_service.go`
- [x] T051 [US3] Implement milestone alert evaluation and notification creation after snapshots in `src\api\services\set_service.go`
- [x] T052 [US3] Implement set snapshot scheduler using existing scheduler patterns in `src\api\services\set_snapshot_scheduler.go`
- [x] T053 [US3] Wire set snapshot scheduler startup and shutdown in `src\api\main.go`
- [x] T054 [US3] Implement `POST /sets/{id}/snapshot`, `GET /sets/{id}/trends`, `GET /sets/{id}/analytics`, and `POST /sets/compare` handlers with Swagger annotations in `src\api\handlers\sets.go`
- [x] T055 [US3] Add snapshot, trends, analytics, and compare frontend API calls in `src\web\src\api\client.ts`
- [x] T056 [US3] Extend TypeScript snapshot, analytics, and comparison types in `src\web\src\types\index.ts`
- [x] T057 [P] [US3] Implement value trend chart component with existing dependencies or a vetted lightweight chart dependency in `src\web\src\components\sets\SetTrendChart.vue`
- [x] T058 [P] [US3] Implement set comparison panel UI in `src\web\src\components\sets\SetComparePanel.vue`
- [x] T059 [US3] Add trend chart, range selector, manual snapshot action, analytics panel, and compare entry point to `src\web\src\pages\SetDetailPage.vue`
- [x] T060 [US3] Add trend summary and value change percentage to dashboard cards in `src\web\src\components\sets\SetDashboardCard.vue`
- [x] T061 [US3] Add set milestone notification type display support in `src\web\src\pages\NotificationsPage.vue`

**Checkpoint**: User Story 3 independently supports set snapshots, trends, analytics, comparison, and milestone notifications.

---

## Phase 6: User Story 4 - Maintain smart sets (Priority: P4)

**Goal**: Users can define validated criteria, preview matching coins, save smart sets, and see membership update from coin data.

**Independent Test**: Create a smart set for silver coins, preview matches, save it, add or edit a coin to match the rule, and confirm the set membership reflects the new data without manual tagging.

### Implementation for User Story 4

- [x] T062 [US4] Define smart criteria Go request models and validation rules in `src\api\services\set_criteria.go`
- [x] T063 [US4] Implement safe criteria-to-GORM query translation with parameter binding in `src\api\repository\set_repository.go`
- [x] T064 [US4] Implement smart set preview and smart membership evaluation service methods in `src\api\services\set_service.go`
- [x] T065 [US4] Prevent manual add/remove membership for smart sets in `src\api\services\set_service.go`
- [x] T066 [US4] Implement smart set preview handler and smart-set create/update validation with Swagger annotations in `src\api\handlers\sets.go`
- [x] T067 [US4] Add smart criteria and preview frontend API calls in `src\web\src\api\client.ts`
- [x] T068 [US4] Extend TypeScript smart criteria types in `src\web\src\types\index.ts`
- [x] T069 [P] [US4] Implement smart rule builder UI in `src\web\src\components\sets\SetSmartRuleBuilder.vue`
- [x] T070 [US4] Extend set creation wizard with smart criteria builder and preview results in `src\web\src\components\sets\SetCreationWizard.vue`
- [x] T071 [US4] Update set detail coin loading to use derived smart membership for smart sets in `src\web\src\pages\SetDetailPage.vue`
- [x] T072 [US4] Add smart set criteria summary display to dashboard cards in `src\web\src\components\sets\SetDashboardCard.vue`

**Checkpoint**: User Story 4 independently supports rule-based smart sets with deterministic preview and derived membership.

---

## Phase 7: Polish & Cross-Cutting Concerns

**Purpose**: Final consistency, documentation, generated contracts, and validation across all selected stories.

- [x] T073 [P] Update API documentation for set endpoints in `docs\api-reference.md`
- [x] T074 [P] Update feature documentation and user-facing workflow notes in `docs\ARCHITECTURE.md`
- [x] T075 Run Swagger regeneration for new set annotations and update generated artifacts via `task openapi`
- [x] T076 Validate Go API architecture, tests, vet, and build from `src\api`
- [x] T077 Validate Vue frontend build, tests, and lint from `src\web`
- [x] T078 Run quickstart smoke test steps and record any follow-up fixes in `specs\main\quickstart.md`
- [x] T079 Review set queries for user scoping, parameterized criteria, and private value exposure in `src\api\repository\set_repository.go`
- [x] T080 Review set UI for design token usage, no emoji copy, and PWA/mobile compatibility in `src\web\src\components\sets\`

---

## Dependencies & Execution Order

### Phase Dependencies

- **Phase 1 Setup**: No dependencies; starts immediately.
- **Phase 2 Foundational**: Depends on Phase 1; blocks all user stories.
- **Phase 3 US1**: Depends on Phase 2; MVP scope.
- **Phase 4 US2**: Depends on Phase 2 and benefits from US1 UI/API foundation.
- **Phase 5 US3**: Depends on Phase 2 and uses summaries from US1 plus completion data from US2 when available.
- **Phase 6 US4**: Depends on Phase 2 and uses set CRUD from US1.
- **Phase 7 Polish**: Depends on all selected user stories for the release.

### User Story Dependencies

- **US1 (P1)**: Required MVP; no dependency on other user stories.
- **US2 (P2)**: Can begin after foundation, but dashboard/detail integration is easiest after US1.
- **US3 (P3)**: Can snapshot open sets after foundation and US1; completion trend fields become richer after US2.
- **US4 (P4)**: Can begin after foundation and set CRUD; independent from US2 and US3 except shared summary rendering.

### Within Each User Story

- Backend models and repositories before services.
- Services before handlers.
- Handlers before frontend API client integration.
- Frontend types and API client before pages/components that consume them.
- Story checkpoint must pass before treating the story as complete.

---

## Parallel Execution Examples

### User Story 1

```text
Task: "Implement SetDashboardCard summary UI using design tokens in src\web\src\components\sets\SetDashboardCard.vue"
Task: "Implement set list, create, get, update, and delete HTTP handlers with Swagger annotations in src\api\handlers\sets.go"
Task: "Implement list/create/get/update/delete frontend API calls in src\web\src\api\client.ts"
```

### User Story 2

```text
Task: "Add built-in set template definitions for popular US series in src\api\services\set_templates.go"
Task: "Implement completion checklist and missing target UI in src\web\src\components\sets\SetCompletionChecklist.vue"
Task: "Extend TypeScript set target, template, and completion types in src\web\src\types\index.ts"
```

### User Story 3

```text
Task: "Implement value trend chart component with existing dependencies or a vetted lightweight chart dependency in src\web\src\components\sets\SetTrendChart.vue"
Task: "Implement set comparison panel UI in src\web\src\components\sets\SetComparePanel.vue"
Task: "Implement set snapshot scheduler using existing scheduler patterns in src\api\services\set_snapshot_scheduler.go"
```

### User Story 4

```text
Task: "Implement smart rule builder UI in src\web\src\components\sets\SetSmartRuleBuilder.vue"
Task: "Define smart criteria Go request models and validation rules in src\api\services\set_criteria.go"
Task: "Add smart criteria and preview frontend API calls in src\web\src\api\client.ts"
```

---

## Implementation Strategy

### MVP First

1. Complete Phase 1 and Phase 2.
2. Complete Phase 3 (US1) only.
3. Validate that existing tags appear as open sets, CRUD works, membership works, and collection filtering still works.
4. Demo or ship the MVP before adding defined, trend, or smart set complexity.

### Incremental Delivery

1. Ship US1 to preserve tags and add open sets with value summaries.
2. Add US2 for templates, target lists, and completion tracking.
3. Add US3 for snapshots, trend charts, analytics, comparison, and milestones.
4. Add US4 for smart criteria and derived membership.
5. Finish with cross-cutting documentation, generated OpenAPI, and validation.

### Parallel Team Strategy

After Phase 2, backend and frontend work can split by file ownership. US2, US3, and US4 can proceed in parallel if developers coordinate shared changes in `src\api\services\set_service.go`, `src\api\repository\set_repository.go`, `src\api\handlers\sets.go`, `src\web\src\api\client.ts`, and `src\web\src\types\index.ts`.

---

## Summary

- **Total tasks**: 80
- **Setup tasks**: 5
- **Foundational tasks**: 14
- **US1 tasks**: 16
- **US2 tasks**: 13
- **US3 tasks**: 13
- **US4 tasks**: 11
- **Polish tasks**: 8
- **Suggested MVP scope**: Phase 1, Phase 2, and Phase 3 only
- **Format validation**: All implementation tasks use markdown checkboxes, sequential IDs, optional `[P]`, required `[US#]` labels for user-story tasks, and exact file paths
