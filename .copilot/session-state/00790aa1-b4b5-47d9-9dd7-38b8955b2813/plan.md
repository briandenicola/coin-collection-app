# AI Coin Value Estimator

## Problem
Users have no automated way to estimate the current market value of coins in their collection. The `currentValue` field must be manually entered. The existing Numista lookup provides catalog data but not pricing. The AI Chat agent can find live listings but requires manual interaction per coin.

## Approach
Add an "Estimate Value" button to the coin detail page that sends the coin's metadata to Claude (Anthropic API) with web search enabled, asks it to research current market prices for similar coins, and returns a value estimate. The estimate auto-populates the `currentValue` field (with user confirmation).

Reuses the existing Anthropic client infrastructure (API key, model selection, settings).

## Todos

### Backend
- **api-estimate-endpoint**: Create `POST /api/coins/:id/estimate-value` endpoint in `agent.go`
  - Loads coin from DB (name, category, denomination, ruler, era, material, grade, weight, diameter)
  - Builds a focused valuation prompt asking Claude to search for comparable sales/listings
  - Uses web_search tool so Claude can check live dealer sites
  - Returns single JSON response: `{ estimatedValue, confidence, reasoning, comparables }`
  - Requires auth (same middleware as agent/chat)

### Frontend
- **fe-types**: Add `ValueEstimate` interface to `types/index.ts`
- **fe-api-client**: Add `estimateCoinValue(coinId)` to `api/client.ts`
- **fe-estimate-button**: Add "Estimate Value" UI to CoinDetailPage.vue
  - Button near value/price section
  - Loading spinner with "Researching market value..."
  - Estimate card: value, confidence badge, reasoning, comparable listings
  - "Apply Estimate" updates currentValue, "Dismiss" closes

### Polish
- **error-handling**: API key not configured (503), coin not found (404), disabled state

## Notes
- Web search tool enabled so Claude checks real dealer pricing
- Confidence: high (3+ comparables), medium (1-2), low (estimate only)
- Does NOT auto-save — user confirms before applying
- Reuses existing AgentHandler HTTP client (300s timeout)
