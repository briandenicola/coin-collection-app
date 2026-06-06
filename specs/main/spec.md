# Feature Specification: Coin Sets with Trend Tracking

**Feature Branch**: `main`  
**Created**: 2026-06-06  
**Status**: Draft  
**Input**: GitHub issue #240, "Expand Tagging System to Support Coin Sets with Trend Tracking"

## User Scenarios & Testing

### User Story 1 - Manage sets evolved from tags (Priority: P1)

Collectors can create and maintain coin sets using the existing tag workflow, with added set metadata, descriptions, and collection value totals.

**Why this priority**: This preserves current tag behavior while creating the foundation for all set analytics.

**Independent Test**: A user can create a set, add and remove coins, view set details, and still filter the collection by that set.

**Acceptance Scenarios**:

1. **Given** a user has existing tags, **When** the feature is enabled, **Then** those tags appear as open coin sets with their current coin memberships preserved.
2. **Given** a user creates an open set, **When** they add coins to it, **Then** the set detail view shows the selected coins, total current value, coin count, and average value.
3. **Given** a coin belongs to multiple sets, **When** the user views each set, **Then** the coin is included independently without duplicating the coin record.

---

### User Story 2 - Track completion for defined sets (Priority: P2)

Series collectors can use defined templates or custom target lists to track completion and identify missing coins.

**Why this priority**: Completion tracking is the core value beyond existing tags and supports popular collecting workflows.

**Independent Test**: A user can create a defined set from a template, match owned coins to targets, and see completion percentage and missing items.

**Acceptance Scenarios**:

1. **Given** a defined set has 50 targets and the user owns 20 matching coins, **When** the set details load, **Then** completion is shown as 40% with 30 missing targets.
2. **Given** the user adds a coin matching a missing target, **When** the set recalculates, **Then** the target is marked complete and completion percentage increases.
3. **Given** a user imports custom targets, **When** the import succeeds, **Then** the set uses those targets for completion checks.

---

### User Story 3 - Monitor value trends for sets (Priority: P3)

Collectors can view historical valuation snapshots for each set and compare performance over time.

**Why this priority**: Trend tracking answers the investment and analytics questions in the issue while building on set aggregates.

**Independent Test**: A user can capture a snapshot for a set and view trend data over selected time ranges.

**Acceptance Scenarios**:

1. **Given** a set has daily snapshots, **When** the user selects a one-year range, **Then** the chart displays total value, coin count, and completion history across that range.
2. **Given** the user compares two sets, **When** both have snapshots, **Then** the comparison shows value change and percentage change for each set.
3. **Given** a set crosses a configured value milestone, **When** snapshots are processed, **Then** a notification is generated once for that milestone crossing.

---

### User Story 4 - Maintain smart sets (Priority: P4)

Collectors can define rule-based sets that automatically include matching coins.

**Why this priority**: Smart sets add power-user organization but are not required for the MVP set foundation.

**Independent Test**: A user can define criteria, preview matching coins, save the set, and see membership update when coins change.

**Acceptance Scenarios**:

1. **Given** a smart set rule for silver coins, **When** the user adds a silver coin, **Then** the coin appears in the smart set without manual tagging.
2. **Given** the user changes a coin so it no longer matches, **When** smart membership is recalculated, **Then** the coin is removed from the smart set view.

## Edge Cases

- Existing tag names remain unique per user after migration to open sets.
- Deleting a coin removes manual memberships but preserves historical set snapshots by retaining aggregate values and the deleted coin's last recorded contribution only in historical data.
- Deleting a set removes memberships, target definitions, criteria, alerts, and future snapshots, but does not delete coins.
- Smart set criteria with invalid fields or operators are rejected with validation errors.
- Defined sets with duplicate target rows are rejected during import.
- Snapshot generation skips empty sets but records a zero-value state for sets that previously had coins and now contain none.
- Nested sets cannot create cycles.

## Requirements

### Functional Requirements

- **FR-001**: The system MUST evolve existing user tags into coin sets without losing existing names, colors, or coin memberships.
- **FR-002**: Users MUST be able to create, read, update, and delete sets with name, description, color, icon name, set type, parent set, target completion date, and visibility settings.
- **FR-003**: Users MUST be able to manually add and remove coins from open, defined, and goal sets.
- **FR-004**: A coin MUST be allowed to belong to multiple sets.
- **FR-005**: The system MUST calculate set summary metrics including coin count, total current value, total invested, average value, highest-value coin, and ROI where cost data exists.
- **FR-006**: Defined and goal sets MUST support target coin definitions and completion percentage.
- **FR-007**: The system MUST provide built-in templates for popular US coin series listed in issue #240 and allow custom CSV target import.
- **FR-008**: Smart sets MUST support rule criteria for material, date range, country/category, mint, grade range, value range, acquisition date, wishlist status, sold status, and AND/OR grouping.
- **FR-009**: Smart set membership MUST be evaluated from coin data and must not store manual membership rows as the source of truth.
- **FR-010**: The system MUST store set valuation snapshots with snapshot date, total value, total invested, coin count, completion percentage, average value, and highest-value coin.
- **FR-011**: Users MUST be able to manually create a set snapshot.
- **FR-012**: The system MUST support scheduled set snapshots using existing scheduler patterns.
- **FR-013**: Users MUST be able to retrieve trend data for a set over standard time ranges.
- **FR-014**: Users MUST be able to compare multiple sets by value, value change, percentage change, coin count, completion, and ROI.
- **FR-015**: The frontend MUST provide a set dashboard, set detail view, creation/edit wizard, completion checklist, trend chart, and comparison experience.
- **FR-016**: All set APIs MUST be scoped to the authenticated user and must not expose private collection values through public sharing.
- **FR-017**: Existing `/tags` and coin tag endpoints MUST remain backward compatible until the UI and clients are migrated.
- **FR-018**: The implementation MUST include Swagger annotations for new public API handlers and frontend API client/types for all new endpoints.

### Key Entities

- **CoinSet**: A user-owned grouping evolved from Tag, with display metadata, type, optional parent, sharing fields, and target completion date.
- **CoinSetMembership**: Manual relationship between a coin and a set, with added date and optional notes.
- **CoinSetTarget**: Expected item in a defined or goal set used to calculate completion and missing coins.
- **CoinSetCriteria**: JSON rule tree for smart sets.
- **CoinSetValuationSnapshot**: Time-series aggregate of set value, invested cost, coin count, completion, and highest-value coin.
- **CoinSetTemplate**: Built-in reusable target definition for common series.
- **CoinSetMilestoneAlert**: User-owned alert for set value or completion thresholds.

## Success Criteria

### Measurable Outcomes

- **SC-001**: Existing tags are visible as open sets with unchanged memberships after migration.
- **SC-002**: A user can create an open set and add coins in under one minute from the dashboard.
- **SC-003**: Set summary metrics load for 50 sets and 5,000 coins within two seconds on typical local hardware.
- **SC-004**: Defined set completion is recalculated correctly after coin add, update, delete, and target import events.
- **SC-005**: Trend endpoints return one year of daily snapshots for a set in under one second for a user with 100 sets.
- **SC-006**: Smart set preview and evaluation return deterministic results for the same criteria and coin data.
- **SC-007**: The frontend build, Go tests, and architecture tests pass after implementation.

## Assumptions

- Initial implementation should deliver a staged MVP: open sets plus summaries first, then defined templates/completion, then snapshots/trends, then smart sets and advanced analytics.
- Current `Tag` and `CoinTag` tables are the migration base rather than creating a disconnected set system.
- Public set sharing is planned as data model/API groundwork only; full social showcase sharing is out of scope for the first implementation slice.
- Existing valuation data from `currentValue`, `purchasePrice`, and `CoinValueHistory` is sufficient for set-level aggregation until Feature #1 improves valuations.
- Chart rendering can use an existing dependency if already present; otherwise a lightweight Vue-compatible chart library should be evaluated during implementation.
