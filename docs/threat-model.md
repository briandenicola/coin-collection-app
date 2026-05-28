# Threat Model

> The live inventory of known security threats and security-review findings for Ancient Coins. This document keeps the original **backend / frontend / supply-chain** split because those are the operational owners and remediation surfaces in this repo.

## How to add a new threat finding

1. Add the finding to the correct surface area below (`B-*`, `F-*`, or `SC-*`).
2. Record **severity**, **status**, **location**, **description**, and **recommended remediation** in the same edit.
3. Use `Open`, `Mitigated`, or `Accepted` as the status. If prioritization matters, append it in parentheses (for example `Open (P1)`).
4. Link any architecture or policy change needed for remediation to an ADR or to [security-principles.md](security-principles.md).
5. If the finding becomes an active incident, manage the operational response in [incident-response.md](incident-response.md) and keep this document focused on the durable risk record.

## Status summary

| Domain | Findings | Mitigated | Open | Accepted |
|---|---:|---:|---:|---:|
| Backend API | 9 | 4 | 5 | 0 |
| Frontend | 8 | 0 | 7 | 1 |
| Supply chain & infrastructure | 7 | 0 | 7 | 0 |
| **Total** | **24 enumerated** | **4** | **19** | **1** |

> The legacy executive summary says 25 findings, but the source document enumerates 24 finding IDs (`B-1`…`B-9`, `F-1`…`F-8`, `SC-1`…`SC-7`). This split preserves the enumerated record and flags the count mismatch for the next audit refresh.

## Backend API findings

| ID | Severity | Status | Location | Description | Recommended remediation |
|---|---|---|---|---|---|
| B-1 | Critical | Mitigated | `src/api/main.go`, `src/api/config/config.go` | CORS previously reflected arbitrary origins while allowing credentials. | Keep `CORS_ORIGINS` allowlisted and review any new origin-handling logic against Principle XI. |
| B-2 | Critical | Open (P1) | `src/api/handlers/analysis.go` | A user-controlled column name is concatenated into a query path, creating a classical SQL-injection shape even if exploitability is constrained. | Replace dynamic column usage with an explicit whitelist or fixed query map. |
| B-3 | High | Mitigated | `src/api/config/config.go` | The application previously allowed a weak default JWT secret. | Preserve the startup fail-fast and minimum-length checks; never weaken them for convenience. |
| B-4 | High | Mitigated | `src/api/handlers/images.go` | Uploads now enforce an extension allowlist, but the original finding still has a residual MIME-validation gap. | Add magic-byte validation so uploads are checked by content, not extension alone. |
| B-5 | High | Mitigated | `src/api/main.go` | Auth endpoints previously allowed unlimited brute-force attempts. | Keep auth rate limiting in place and extend it to new auth-adjacent endpoints when added. |
| B-6 | Medium | Open (P2) | `src/api/main.go` | Request bodies and multipart uploads lack explicit size caps, enabling avoidable memory pressure or DoS risk. | Set `MaxMultipartMemory` and add JSON body-size middleware. |
| B-7 | Medium | Open (P2) | `src/api/handlers/webauthn.go` | In-memory WebAuthn ceremony sessions never expire, so abandoned sessions can accumulate indefinitely. | Add a short TTL and periodic cleanup for challenge state. |
| B-8 | Medium | Open (P2) | `src/api/handlers/webauthn.go` | WebAuthn origin validation dynamically trusts the request origin, weakening relying-party protections. | Restrict allowed origins to configured values only. |
| B-9 | Low | Open (P3) | `src/api/handlers/numista.go` | Some error responses expose more internal detail than clients need. | Return generic client-facing errors and keep specifics in logs only. |

## Frontend findings

| ID | Severity | Status | Location | Description | Recommended remediation |
|---|---|---|---|---|---|
| F-1 | Critical | Open (P0) | `src/web/src/pages/CoinDetailPage.vue` | AI analysis Markdown is rendered through `v-html` without a sanitizer, creating an XSS path. | Sanitize rendered HTML before injection. |
| F-2 | Critical | Open (P0) | `src/web/src/components/CoinSearchChat.vue` | Chat messages are transformed into HTML and injected without sanitization. | Sanitize message output before `v-html`. |
| F-3 | Critical | Open (P3) | `src/web/src/stores/auth.ts` | JWT and refresh tokens live in `localStorage`, which turns any XSS into token theft. | Short term: eliminate XSS sinks. Long term: evaluate `HttpOnly` cookie transport with an ADR. |
| F-4 | High | Open (P1) | `src/web/package.json` | Rich-text rendering existed without a dedicated sanitization dependency. | Keep a maintained sanitizer dependency in the web surface and use it consistently. |
| F-5 | High | Open (P3) | `src/web/src/api/client.ts` | Refresh tokens travel in a JSON body rather than an `HttpOnly` cookie-based flow. | Revisit transport if auth is redesigned; document the trade-off in ADRs. |
| F-6 | Medium | Open (P2) | frontend auth responses | Sensitive responses are not explicitly marked `Cache-Control: no-store`. | Set no-store headers on auth and token-bearing responses. |
| F-7 | Medium | Open (P2) | `src/web/src/api/client.ts` | WebAuthn login finish sends the username in the query string, leaking it into histories and logs. | Move the username into the request body. |
| F-8 | Low | Accepted (platform limitation) | `src/web/src/pages/LoginPage.vue` | Password state cannot be meaningfully scrubbed from browser memory after use in a reliable, portable way. | Accept the platform limit; keep lifetime short in reactive state and avoid unnecessary copies. |

## Supply chain and infrastructure findings

| ID | Severity | Status | Location | Description | Recommended remediation |
|---|---|---|---|---|---|
| SC-1 | Critical | Open (P1) | `.github/workflows/docker-publish*.yml` | GitHub Actions use mutable tags instead of immutable SHAs, exposing the CI chain to supply-chain drift or maintainer compromise. | Pin every action by commit SHA. |
| SC-2 | Critical | Open (P0) | `Taskfile.yml` | A development JWT secret is hardcoded in a tracked file, which means it is already in git history. | Move the secret to ignored local env files and assess whether history cleanup is warranted. |
| SC-3 | Critical | Open (P2) | `src/web/package.json` | `@imgly/background-removal` downloads large runtime models from external CDNs without integrity verification. | Evaluate bundling, server-side processing, or a lower-risk alternative. |
| SC-4 | High | Open (P1) | `src/api/go.mod` | `golang.org/x/*` dependencies were called out for lagging current versions in the original review. | Re-check with `govulncheck` / dependency review and upgrade deliberately. |
| SC-5 | High | Open (P2) | GitHub repository settings | Branch protection is not documented as enforced for `main`/`beta`. | Require PR reviews, required checks, and restricted direct pushes. |
| SC-6 | Medium | Open (P2) | `Dockerfile` | Base images use mutable tags instead of fully reproducible digests or patch pins. | Pin trusted base images more tightly for production builds. |
| SC-7 | Medium | Open (P2) | `Dockerfile` | Final application containers run as root by default. | Create and switch to a non-root runtime user. |

## Related documents

- [security-principles.md](security-principles.md)
- [incident-response.md](incident-response.md)
- [authentication.md](authentication.md)
- [SECURITY.md](../SECURITY.md)
