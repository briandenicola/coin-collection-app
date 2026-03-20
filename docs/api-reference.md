# API Reference

Complete REST API reference for Ancient Coins. All endpoints are served under the `/api` prefix (e.g., `http://localhost:8080/api/coins`). Interactive documentation is also available via Swagger UI at `/swagger/index.html`.

## Authentication

Most endpoints require authentication via one of two methods:

| Method | Header | Example |
| ------ | ------ | ------- |
| JWT Bearer Token | `Authorization: Bearer <token>` | Obtained from `/api/auth/login` |
| API Key | `X-API-Key: <key>` | Generated via `/api/auth/api-keys` |

To obtain a JWT token, call the [login](#post-apiauthlogin) endpoint. Tokens can be refreshed via the [refresh](#post-apiauthrefresh) endpoint. API keys are managed through the [API Keys](#api-keys) endpoints.

---

## Public Endpoints

These endpoints do not require authentication.

### GET /api/auth/setup

Check whether any users exist. Used by the frontend to determine if the app needs first-time setup.

**Response:**

```json
{ "needsSetup": true }
```

### POST /api/auth/register

Create a new user account. The **first user** to register is automatically assigned the **admin** role.

**Request Body:**

```json
{
  "username": "collector",
  "password": "s3cur3P@ss"
}
```

**Response:**

```json
{
  "token": "eyJhbGciOi...",
  "refreshToken": "dGhpcyBpcyBh...",
  "user": {
    "id": 1,
    "username": "collector",
    "role": "admin"
  }
}
```

### POST /api/auth/login

Authenticate and receive a JWT token.

**Request Body:**

```json
{
  "username": "collector",
  "password": "s3cur3P@ss"
}
```

**Response:**

```json
{
  "token": "eyJhbGciOi...",
  "refreshToken": "dGhpcyBpcyBh...",
  "user": {
    "id": 1,
    "username": "collector",
    "role": "admin"
  }
}
```

**Example:**

```sh
curl -X POST http://localhost:8080/api/auth/login \
  -H "Content-Type: application/json" \
  -d '{"username": "collector", "password": "s3cur3P@ss"}'
```

### POST /api/auth/refresh

Refresh an expired JWT using a refresh token.

**Request Body:**

```json
{ "refreshToken": "dGhpcyBpcyBh..." }
```

**Response:** Same shape as the login response (`token`, `refreshToken`, `user`).

### POST /api/auth/webauthn/login/begin

Begin a WebAuthn (biometric/passkey) login ceremony.

**Request Body:**

```json
{ "username": "collector" }
```

### POST /api/auth/webauthn/login/finish

Complete a WebAuthn login ceremony. The request body contains the authenticator assertion response from the browser WebAuthn API.

**Query Parameters:**

| Param | Description |
| ----- | ----------- |
| `username` | The username that initiated the ceremony |

**Response:** Same shape as the login response (`token`, `refreshToken`, `user`).

### GET /api/auth/webauthn/check

Check whether a user has registered biometric/passkey credentials.

**Query Parameters:**

| Param | Description |
| ----- | ----------- |
| `username` | The username to check |

**Response:**

```json
{ "hasCredentials": true }
```

---

## Protected Endpoints

All endpoints below require a valid `Authorization: Bearer <token>` or `X-API-Key` header.

---

### Coins

#### GET /api/coins

List coins in the collection with filtering, sorting, and pagination.

**Query Parameters:**

| Param | Type | Default | Description |
| ----- | ---- | ------- | ----------- |
| `category` | string | — | Filter by category (`Roman`, `Greek`, `Byzantine`, `Modern`, `Other`) |
| `search` | string | — | Full-text search across name, ruler, denomination, and other fields |
| `wishlist` | string | — | `true` for wishlist only, `false` for collection only |
| `sort` | string | `createdAt` | Field name to sort by (e.g., `name`, `purchasePrice`, `currentValue`, `createdAt`) |
| `order` | string | `desc` | Sort direction: `asc` or `desc` |
| `page` | int | `1` | Page number (1-indexed) |
| `limit` | int | `50` | Results per page |

**Response:**

```json
{
  "coins": [ { "id": 1, "name": "Augustus Denarius", "..." : "..." } ],
  "total": 142,
  "page": 1,
  "limit": 50
}
```

**Example — search Roman coins sorted by value:**

```sh
curl "http://localhost:8080/api/coins?category=Roman&sort=currentValue&order=desc&limit=10" \
  -H "Authorization: Bearer $TOKEN"
```

#### GET /api/coins/:id

Get a single coin with all fields and associated images.

**Example:**

```sh
curl http://localhost:8080/api/coins/42 \
  -H "Authorization: Bearer $TOKEN"
```

#### POST /api/coins

Create a new coin. Only `name` is required; all other fields are optional. See the [field reference](getting-started.md#field-reference-for-import) for the full list.

**Request Body:**

```json
{
  "name": "Nero Sestertius - Port of Ostia",
  "category": "Roman",
  "material": "Bronze",
  "denomination": "Sestertius",
  "ruler": "Nero",
  "purchasePrice": 1200.00,
  "isWishlist": false
}
```

**Example:**

```sh
curl -X POST http://localhost:8080/api/coins \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"name": "Nero Sestertius", "category": "Roman", "material": "Bronze"}'
```

#### PUT /api/coins/:id

Update an existing coin. Send any fields you want to change.

**Request Body:** Same schema as create — only include the fields you want to update.

#### POST /api/coins/:id/purchase

Move a wishlist coin into the collection by setting `isWishlist` to `false`. Use this when you acquire a coin you were tracking on your wishlist.

#### POST /api/coins/:id/sell

Mark a coin as sold.

**Request Body:**

```json
{
  "salePrice": 750.00,
  "buyerName": "John Smith"
}
```

#### DELETE /api/coins/:id

Delete a coin and all of its associated images (both database records and uploaded files).

---

### Images

#### POST /api/coins/:id/images

Upload an image file via multipart form data.

**Form Fields:**

| Field | Type | Required | Description |
| ----- | ---- | -------- | ----------- |
| `file` | file | Yes | The image file |
| `imageType` | string | No | One of: `obverse`, `reverse`, `detail`, `other` |
| `isPrimary` | string | No | `true` to set as the primary/cover image |

**Example:**

```sh
curl -X POST http://localhost:8080/api/coins/42/images \
  -H "Authorization: Bearer $TOKEN" \
  -F "file=@coin-obverse.jpg" \
  -F "imageType=obverse" \
  -F "isPrimary=true"
```

#### POST /api/coins/:id/images/base64

Upload an image as a base64-encoded string. Useful for programmatic uploads or when working with image data directly.

**Supported extensions:** `jpg`, `jpeg`, `png`, `gif`, `webp`, `bmp`, `tiff`

**Request Body:**

```json
{
  "data": "/9j/4AAQSkZJRg...",
  "fileExtension": "jpg",
  "imageType": "obverse",
  "isPrimary": true
}
```

#### DELETE /api/coins/:id/images/:imageId

Delete an image by ID. Removes both the database record and the uploaded file.

#### GET /api/proxy-image

Proxy an external image URL through the server. Useful to avoid CORS issues when displaying images from third-party sites.

**Query Parameters:**

| Param | Description |
| ----- | ----------- |
| `url` | The external image URL to proxy |

**Example:**

```sh
curl "http://localhost:8080/api/proxy-image?url=https://example.com/coin.jpg" \
  -H "Authorization: Bearer $TOKEN" \
  --output coin.jpg
```

#### GET /api/scrape-image

Scrape the `og:image` meta tag from a URL. Returns the image URL found in the page's OpenGraph metadata.

**Query Parameters:**

| Param | Description |
| ----- | ----------- |
| `url` | The page URL to scrape |

---

### Journal

Each coin can have journal entries for tracking research notes, provenance details, or condition observations over time.

#### GET /api/coins/:id/journal

List all journal entries for a coin.

#### POST /api/coins/:id/journal

Add a journal entry to a coin.

**Request Body:**

```json
{ "entry": "Identified mint mark as Rome based on comparison with RIC plates." }
```

**Example:**

```sh
curl -X POST http://localhost:8080/api/coins/42/journal \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"entry": "Cleaned with distilled water. No change to surfaces."}'
```

#### DELETE /api/coins/:id/journal/:entryId

Delete a journal entry by ID.

---

### AI Analysis

AI-powered coin analysis using Ollama vision models. Requires Ollama to be running and configured in admin settings.

#### POST /api/coins/:id/analyze

Run AI analysis on a coin's images. Analyzes either the obverse or reverse side.

**Query Parameters:**

| Param | Values | Description |
| ----- | ------ | ----------- |
| `side` | `obverse`, `reverse` | Which side to analyze |

**Example:**

```sh
curl -X POST "http://localhost:8080/api/coins/42/analyze?side=obverse" \
  -H "Authorization: Bearer $TOKEN"
```

#### DELETE /api/coins/:id/analyze

Delete a stored AI analysis for a coin.

**Query Parameters:**

| Param | Values | Description |
| ----- | ------ | ----------- |
| `side` | `obverse`, `reverse` | Which analysis to delete |

#### POST /api/extract-text

Extract text from an image using OCR. Accepts a multipart form upload.

**Form Fields:**

| Field | Type | Description |
| ----- | ---- | ----------- |
| `file` | file | The image file to extract text from |

**Example:**

```sh
curl -X POST http://localhost:8080/api/extract-text \
  -H "Authorization: Bearer $TOKEN" \
  -F "file=@inscription.jpg"
```

#### GET /api/ollama-status

Check Ollama connectivity and whether the configured vision model is available.

**Response:**

```json
{
  "connected": true,
  "model": "llava",
  "modelAvailable": true
}
```

---

### Statistics

#### GET /api/stats

Get collection statistics including totals, category breakdowns, material breakdowns, and grade distributions.

**Example:**

```sh
curl http://localhost:8080/api/stats \
  -H "Authorization: Bearer $TOKEN"
```

#### GET /api/value-history

Get portfolio value snapshots over time, useful for charting collection value trends.

#### GET /api/suggestions

Get autocomplete suggestions for coin fields (e.g., rulers, mints, denominations). Used by the frontend to power form autocomplete.

---

### Numista

#### GET /api/numista/search

Search the [Numista](https://en.numista.com/) coin catalog.

**Query Parameters:**

| Param | Description |
| ----- | ----------- |
| `q` | Search terms (e.g., `Augustus denarius`) |

**Example:**

```sh
curl "http://localhost:8080/api/numista/search?q=Augustus+denarius" \
  -H "Authorization: Bearer $TOKEN"
```

---

### AI Agent

Chat with an AI-powered coin search agent backed by Anthropic models. The agent can search your collection, answer numismatic questions, and provide research assistance.

#### POST /api/agent/chat

Send a message to the AI agent. The response is streamed via **Server-Sent Events (SSE)**.

**Request Body:**

```json
{
  "messages": [
    { "role": "user", "content": "What Roman denarii do I have from the Julio-Claudian dynasty?" }
  ],
  "conversationId": "optional-conversation-id"
}
```

**SSE Stream Format:**

The response uses `Content-Type: text/event-stream`. Each event is a JSON object:

```
data: {"type": "text", "content": "Based on your collection..."}

data: {"type": "text", "content": " I found 3 denarii..."}

data: {"type": "done"}
```

**Example:**

```sh
curl -N -X POST http://localhost:8080/api/agent/chat \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"messages": [{"role": "user", "content": "Show me my most valuable coins"}]}'
```

> **Note:** Use `curl -N` (no-buffer) to see streamed events in real time.

#### GET /api/agent/models

List available Anthropic models that can be used with the agent.

#### GET /api/agent/prompt

Get the current system prompt used by the AI agent.

---

### Conversations

Save and manage AI agent conversation history.

#### GET /api/agent/conversations

List all saved conversations.

#### GET /api/agent/conversations/:id

Get a saved conversation by ID, including the full message history.

#### POST /api/agent/conversations

Save a conversation.

**Request Body:**

```json
{
  "title": "Julio-Claudian Research",
  "messages": [
    { "role": "user", "content": "What coins do I have from Augustus?" },
    { "role": "assistant", "content": "You have 5 coins from Augustus..." }
  ]
}
```

#### DELETE /api/agent/conversations/:id

Delete a saved conversation.

---

### User

#### GET /api/auth/me

Get the current authenticated user's information.

**Response:**

```json
{
  "id": 1,
  "username": "collector",
  "role": "admin"
}
```

#### POST /api/auth/change-password

Change the current user's password.

**Request Body:**

```json
{
  "currentPassword": "oldP@ss",
  "newPassword": "n3wS3cur3P@ss"
}
```

#### GET /api/user/export

Export the entire collection as a JSON file download. The response has `Content-Disposition: attachment` headers.

**Example:**

```sh
curl http://localhost:8080/api/user/export \
  -H "Authorization: Bearer $TOKEN" \
  --output my-collection.json
```

#### POST /api/user/import

Import coins from a JSON array. See the [Getting Started guide](getting-started.md#import--export) for the full field reference and import behavior.

**Request Body:** A JSON array of coin objects.

```sh
curl -X POST http://localhost:8080/api/user/import \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d @my-coins.json
```

**Response:**

```json
{ "message": "Import complete", "imported": 42 }
```

---

### API Keys

Manage API keys for programmatic access. API keys are an alternative to JWT tokens and are useful for scripts, integrations, and automation.

#### POST /api/auth/api-keys

Generate a new API key. The full key is returned **only once** in the response — store it securely.

**Request Body:**

```json
{ "name": "My Integration" }
```

The `name` field is optional and helps you identify the key later.

**Response:**

```json
{
  "id": 1,
  "name": "My Integration",
  "key": "ak_abc123def456...",
  "prefix": "ak_abc1",
  "createdAt": "2024-03-15T10:30:00Z"
}
```

**Example:**

```sh
# Generate a key
curl -X POST http://localhost:8080/api/auth/api-keys \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"name": "Backup Script"}'

# Use the key
curl http://localhost:8080/api/coins \
  -H "X-API-Key: ak_abc123def456..."
```

#### GET /api/auth/api-keys

List all API keys for the current user. Only the key prefix is shown (not the full key).

#### DELETE /api/auth/api-keys/:id

Revoke an API key. The key will immediately stop working.

---

### WebAuthn (Credential Registration)

Manage WebAuthn/passkey credentials for biometric login. Login ceremonies use the [public endpoints](#public-endpoints) above.

#### POST /api/auth/webauthn/register/begin

Begin a WebAuthn credential registration ceremony. Returns challenge data for the browser WebAuthn API.

#### POST /api/auth/webauthn/register/finish

Complete credential registration. The request body contains the authenticator attestation response from the browser WebAuthn API.

#### GET /api/auth/webauthn/credentials

List all registered WebAuthn credentials for the current user.

#### DELETE /api/auth/webauthn/credentials/:id

Delete a registered WebAuthn credential.

---

## Admin Endpoints

These endpoints require the **admin** role. Regular users will receive a `403 Forbidden` response.

### GET /api/admin/users

List all registered users.

### DELETE /api/admin/users/:id

Delete a user account.

### POST /api/admin/users/:id/reset-password

Reset a user's password (admin override).

**Request Body:**

```json
{ "newPassword": "t3mpP@ss" }
```

### GET /api/admin/settings

Get all application settings.

### GET /api/admin/settings/defaults

Get the default values for all settings.

### PUT /api/admin/settings

Update one or more settings.

**Request Body:**

```json
[
  { "key": "ollama_url", "value": "http://localhost:11434" },
  { "key": "ollama_model", "value": "llava" },
  { "key": "log_level", "value": "info" }
]
```

**Example:**

```sh
curl -X PUT http://localhost:8080/api/admin/settings \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '[{"key": "ollama_url", "value": "http://ollama:11434"}]'
```

### GET /api/admin/logs

Get application logs with optional filtering.

**Query Parameters:**

| Param | Type | Default | Description |
| ----- | ---- | ------- | ----------- |
| `level` | string | — | Filter by log level (`trace`, `debug`, `info`, `warn`, `error`) |
| `limit` | int | `100` | Number of log entries to return |

**Example:**

```sh
curl "http://localhost:8080/api/admin/logs?level=error&limit=50" \
  -H "Authorization: Bearer $TOKEN"
```

---

## Static Resources

| Path | Description |
| ---- | ----------- |
| `/swagger/index.html` | Interactive Swagger UI API documentation |
| `/uploads/*` | Uploaded coin images (served as static files) |

---

## Common Patterns

### Pagination

List endpoints support pagination via `page` and `limit` query parameters. The default page size is **50**.

```sh
# Get page 3 with 20 results per page
curl "http://localhost:8080/api/coins?page=3&limit=20" \
  -H "Authorization: Bearer $TOKEN"
```

The response includes `total`, `page`, and `limit` fields so clients can calculate the total number of pages.

### Sorting

List endpoints support sorting via `sort` and `order` query parameters.

```sh
# Sort by purchase price, highest first
curl "http://localhost:8080/api/coins?sort=purchasePrice&order=desc" \
  -H "Authorization: Bearer $TOKEN"
```

### Error Responses

Errors return an appropriate HTTP status code with a JSON body:

```json
{ "error": "coin not found" }
```

| Status | Meaning |
| ------ | ------- |
| `400` | Bad request (invalid input) |
| `401` | Unauthorized (missing or invalid token/key) |
| `403` | Forbidden (insufficient permissions) |
| `404` | Resource not found |
| `500` | Internal server error |
