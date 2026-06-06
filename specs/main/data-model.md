# Data Model: Coin Sets with Trend Tracking

## CoinSet

Represents a user-owned set evolved from the current tag concept.

| Field | Type | Rules |
|---|---|---|
| id | uint | Primary key |
| user_id | uint | Required; indexed; owner scope |
| name | string | Required; 1-80 characters; unique per user, case-insensitive |
| description | text | Optional; max 2,000 characters |
| color | string | Optional hex color; defaults to current tag gray |
| icon | string | Optional lucide icon name; max 50 characters |
| set_type | enum | `open`, `defined`, `smart`, `goal` |
| parent_set_id | uint nullable | Must belong to same user; no cycles |
| target_completion_date | date nullable | Only meaningful for goal/defined sets |
| is_public | bool | Default false |
| share_token | string nullable | Random, unique, only present when public sharing is enabled |
| smart_criteria | JSON nullable | Required for smart sets; validated criteria tree |
| created_at | timestamp | Set on create |
| updated_at | timestamp | Set on update |

### Relationships

- Has many `CoinSetMembership`
- Has many `CoinSetTarget`
- Has many `CoinSetValuationSnapshot`
- Has many child `CoinSet` rows through `parent_set_id`

### State/Validation

- Existing `Tag` rows migrate to `CoinSet` with `set_type=open`.
- For `smart` sets, manual membership endpoints are disabled or ignored with a validation error.
- `parent_set_id` cannot reference the same set or any descendant.
- Deleting a set cascades set-owned membership, targets, snapshots, and alerts but never deletes coins.

## CoinSetMembership

Manual membership for open, defined, and goal sets.

| Field | Type | Rules |
|---|---|---|
| set_id | uint | Composite primary key; required |
| coin_id | uint | Composite primary key; required |
| added_at | timestamp | Required; defaults to now |
| notes | text | Optional; max 1,000 characters |

### Relationships

- Belongs to `CoinSet`
- Belongs to `Coin`

### State/Validation

- The set and coin must belong to the same user.
- Inserts are idempotent.
- Sold and wishlist coins may be included manually unless a set-specific filter excludes them.

## CoinSetTarget

Defines expected coins for completion tracking.

| Field | Type | Rules |
|---|---|---|
| id | uint | Primary key |
| set_id | uint | Required; indexed |
| label | string | Required; display name for checklist |
| year | int nullable | Optional |
| mint_mark | string nullable | Optional; max 20 characters |
| denomination | string nullable | Optional; max 200 characters |
| country | string nullable | Optional; max 100 characters |
| material | string nullable | Optional; must match allowed material when present |
| match_rules | JSON nullable | Optional structured matching hints |
| sort_order | int | Required; stable checklist ordering |
| created_at | timestamp | Set on create |

### Relationships

- Belongs to `CoinSet`

### State/Validation

- Only valid for `defined` and `goal` sets.
- Duplicate target identities within one set are rejected.
- Completion is calculated by matching owned coins to targets using deterministic rules.

## CoinSetCriteria

Stored as JSON on `CoinSet.smart_criteria`.

```json
{
  "operator": "and",
  "rules": [
    { "field": "material", "op": "eq", "value": "Silver" },
    { "field": "purchaseDate", "op": "between", "value": ["2026-01-01", "2026-12-31"] }
  ]
}
```

### Allowed fields

`material`, `category`, `denomination`, `ruler`, `era`, `mint`, `grade`, `currentValue`, `purchasePrice`, `purchaseDate`, `createdAt`, `isWishlist`, `isSold`, `isPrivate`.

### Allowed operators

`eq`, `neq`, `contains`, `startsWith`, `in`, `between`, `gte`, `lte`, `isNull`, `isNotNull`.

### State/Validation

- Field/operator combinations must be compatible with the field type.
- All SQL generated from criteria must use parameter binding.
- Criteria preview returns matched coin IDs and summary metrics before save.

## CoinSetValuationSnapshot

Aggregate time-series row for trend tracking.

| Field | Type | Rules |
|---|---|---|
| id | uint | Primary key |
| set_id | uint | Required; indexed |
| user_id | uint | Required; indexed |
| snapshot_date | date | Required; unique with set_id |
| total_value | decimal | Required; defaults to 0 |
| total_invested | decimal | Required; defaults to 0 |
| coin_count | int | Required |
| completion_percentage | decimal nullable | Present for defined/goal sets |
| avg_value_per_coin | decimal nullable | Null when coin_count is 0 |
| highest_value_coin_id | uint nullable | Coin must belong to user at snapshot time |
| created_at | timestamp | Set on create |

### Relationships

- Belongs to `CoinSet`
- Optionally references `Coin` for current highest-value coin

### State/Validation

- One snapshot per set per date.
- Historical rows are immutable except for same-day manual recapture.
- Snapshot calculations include only user-owned coins visible to that set's rules.

## CoinSetTemplate

Built-in target definitions for common collecting series.

| Field | Type | Rules |
|---|---|---|
| id | string | Stable template key |
| name | string | Required |
| category | string | Required grouping |
| description | text | Optional |
| targets | JSON | Required array of target definitions |
| version | int | Incremented when template target list changes |

### State/Validation

- Templates are application-owned and read-only to normal users.
- Creating a set from a template copies target definitions into `CoinSetTarget`.

## CoinSetMilestoneAlert

Tracks one-time or repeatable alerts for set milestones.

| Field | Type | Rules |
|---|---|---|
| id | uint | Primary key |
| set_id | uint | Required |
| user_id | uint | Required |
| metric | enum | `total_value`, `completion_percentage`, `coin_count` |
| threshold | decimal | Required |
| direction | enum | `crosses_above`, `crosses_below` |
| last_triggered_at | timestamp nullable | Used for idempotency |
| enabled | bool | Default true |

### State/Validation

- Alert evaluation runs after snapshot creation.
- Notifications reuse the existing notification infrastructure.
