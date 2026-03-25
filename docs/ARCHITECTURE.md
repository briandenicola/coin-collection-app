# Architecture

This document describes the layered architecture of the Ancient Coins API (`src/api/`). All contributors (human and AI) should follow these conventions.

## Layer Diagram

```
Request
  |
  v
[Handler]  ----  Parses HTTP request, calls service/repo, returns HTTP response
  |
  v
[Service]  ----  Business logic, orchestration, domain rules (HTTP-agnostic)
  |
  v
[Repository]  --  Database queries, GORM operations (one repo per domain)
  |
  v
[Database]  ----  *gorm.DB (initialized once in main.go, injected everywhere)
```

## Rules

### 1. Only `main.go` references the `database` package

Every other package receives `*gorm.DB` (or a repository/service) through its constructor. This is enforced by `architecture_test.go`.

```go
// GOOD (main.go)
coinRepo := repository.NewCoinRepository(database.DB)

// BAD (anywhere else)
database.DB.Find(&coins)  // violates DI rule
```

### 2. Handlers are thin

Handlers should only:
- Parse the request (bind JSON, extract path/query params)
- Call a service or repository method
- Map the result to an HTTP response

Handlers must NOT contain business logic, multi-step orchestration, or raw SQL.

### 3. Services contain business logic

Services orchestrate multiple repository calls, enforce domain rules, and perform data transformations. They are HTTP-agnostic (no `gin.Context`, no `net/http`).

```go
// Service method
func (s *CoinService) UpdateCoin(existing *models.Coin, updates *models.Coin, ...) error {
    s.repo.Update(existing, updates)
    if valueChanged { s.repo.RecordValueHistory(...) }
    s.repo.RecordValueSnapshot(userID)
}
```

### 4. Repositories own all database access

Every GORM query lives in a repository file. Repositories take `*gorm.DB` via constructor and expose typed methods.

```go
func (r *CoinRepository) FindByID(id, userID uint) (*models.Coin, error) {
    var coin models.Coin
    err := r.db.Scopes(OwnedByID(id, userID)).Preload("Images").First(&coin).Error
    return &coin, err
}
```

### 5. Multi-step writes use transactions

When a method performs 2+ database writes that must succeed or fail together, wrap them in `r.db.Transaction()`. Key examples:
- `CoinRepository.Delete` (coin + images + journals + value history)
- `AdminRepository.DeleteUserCascade` (all user data)
- `AuthRepository.RotateRefreshToken` (revoke old + create new)
- `ImageRepository.SetPrimaryAndCreate` (clear primary + insert)

### 6. Errors stay internal

Never expose raw `err.Error()` from internal packages (GORM, Ollama, WebAuthn) to API clients. Log the real error server-side and return a generic message. Validation errors from `ShouldBindJSON` are acceptable to return.

## Package Map

| Package | Responsibility | Imports allowed |
|---|---|---|
| `main.go` | Composition root, wiring | Everything |
| `handlers/` | HTTP layer | `services/`, `repository/`, `models/` |
| `services/` | Business logic | `repository/`, `models/` |
| `repository/` | Database access | `models/`, `gorm.io/gorm` |
| `models/` | Data structures | Standard library only |
| `middleware/` | Auth, rate limiting | `models/`, `gorm.io/gorm` |
| `config/` | Environment config | Standard library only |
| `database/` | DB initialization | `gorm.io/gorm`, driver |

## Shared Scopes

Reusable GORM scopes live in `repository/scopes.go`:
- `OwnedBy(userID)` — filters by `user_id`
- `ByID(id)` — filters by primary key
- `OwnedByID(id, userID)` — filters by both
- `ActiveCollection(userID)` — non-wishlist, non-sold coins
- `PublicCoins(userID)` — public, non-wishlist, non-sold
- `ByCoinID(coinID)` — filters by `coin_id`

Use scopes instead of repeating `Where("user_id = ? AND ...")` queries.

## Running Architecture Tests

```bash
cd src/api
go test -v -run "TestNoDirectDatabase|TestHandlersDoNotUseRawSQL" .
```
