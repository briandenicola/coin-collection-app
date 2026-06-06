# Research: Coin Sets with Trend Tracking

## Decision: Evolve tags into `CoinSet` rather than add a separate set concept

**Rationale**: Existing tags already provide user-owned labels, colors, many-to-many coin membership, collection filtering, bulk tagging, and settings management. Evolving this model preserves current behavior and minimizes user migration risk.

**Alternatives considered**:
- Add separate `sets` tables and keep tags unchanged: rejected because users would have two overlapping grouping concepts.
- Rename all tag endpoints immediately: rejected because existing frontend and possible clients depend on `/tags`.

## Decision: Keep backward-compatible `/tags` endpoints while adding `/sets`

**Rationale**: `/sets` can expose richer analytics and set-specific capabilities while `/tags` remains a compatibility facade over open sets. This enables incremental UI migration and reduces rollout risk.

**Alternatives considered**:
- Replace `/tags` in one release: rejected because it is higher risk and requires all UI surfaces to move at once.
- Keep only `/tags` with expanded payloads: rejected because set analytics, templates, trends, and compare endpoints are clearer under `/sets`.

## Decision: Implement in staged slices

**Rationale**: Issue #240 spans several feature families. A staged plan keeps each slice independently testable: open sets and summaries, defined templates/completion, snapshots/trends, smart sets, goals/milestones, then advanced analytics.

**Alternatives considered**:
- Implement all set types in one release: rejected due to high UX, migration, and performance risk.
- Build only smart sets first: rejected because it does not preserve the current tag workflow.

## Decision: Store smart criteria as typed JSON with backend validation

**Rationale**: Criteria need nested AND/OR groups but should be restricted to known coin fields and operators. A typed JSON contract offers flexibility while allowing safe validation and parameterized repository queries.

**Alternatives considered**:
- Store raw SQL: rejected for security and portability.
- Store a fixed column per criterion: rejected because nested combinations and future fields would require repeated migrations.

## Decision: Manual sets store memberships; smart sets derive membership at read time

**Rationale**: Open, defined, and goal sets need user-curated membership and notes. Smart sets should reflect current coin data automatically and avoid stale join rows.

**Alternatives considered**:
- Materialize smart set membership on every coin update: rejected initially because it creates synchronization complexity.
- Store no memberships for any set: rejected because manual curation is a core existing tag behavior.

## Decision: Add set snapshots as aggregate time-series rows

**Rationale**: Trend charts and comparisons need stable historical aggregates even if coins are later deleted or edited. Aggregate snapshots are compact and avoid recalculating every historical point from coin histories.

**Alternatives considered**:
- Calculate all trends from per-coin value history on demand: rejected for performance and because not all historical membership states are available.
- Store full coin-level snapshot details for every set every day: rejected for storage growth in the first version.

## Decision: Reuse scheduler patterns from valuation jobs

**Rationale**: The Go API already has scheduler patterns for valuation and availability jobs. Set snapshot scheduling belongs in the Go API because it aggregates persisted collection data and does not require AI inference.

**Alternatives considered**:
- Use the Python agent service: rejected by constitution because Python agent is stateless and must not access the database.
- Require only manual snapshots: rejected because daily trends are an acceptance criterion.

## Decision: Keep public sharing private-by-default and limited in v1

**Rationale**: Collection values are sensitive. Public set URLs need explicit opt-in, random share tokens, and value/privacy controls before external sharing is exposed.

**Alternatives considered**:
- Make all sets shareable immediately: rejected due to privacy risk.
- Omit sharing fields entirely: rejected because issue #240 includes `is_public` and `share_token`, and future social sharing is anticipated.

## Decision: Use existing validation and architecture gates

**Rationale**: The feature touches Go API, Vue UI, and SQLite persistence. It must follow Handler -> Service -> Repository, Swagger annotations, frontend API client usage, and design tokens.

**Alternatives considered**:
- Put aggregation logic in handlers for speed: rejected by constitution.
- Add direct frontend calls to non-API services: rejected by service boundary rules.
