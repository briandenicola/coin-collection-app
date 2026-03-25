package handlers

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/briandenicola/ancient-coins-api/models"
	"github.com/briandenicola/ancient-coins-api/repository"
	"github.com/briandenicola/ancient-coins-api/services"
	"github.com/gin-gonic/gin"
)

type AgentHandler struct {
	client   *http.Client
	repo     *repository.AgentRepository
	userRepo *repository.UserRepository
}

func NewAgentHandler(repo *repository.AgentRepository, userRepo *repository.UserRepository) *AgentHandler {
	return &AgentHandler{
		client:   &http.Client{Timeout: 300 * time.Second},
		repo:     repo,
		userRepo: userRepo,
	}
}

// Chat request/response types

type AgentChatRequest struct {
	Message string             `json:"message" binding:"required"`
	History []AgentChatMessage `json:"history"`
}

type AgentChatMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type CoinSuggestion struct {
	Name         string `json:"name"`
	Description  string `json:"description"`
	Category     string `json:"category"`
	Era          string `json:"era"`
	Ruler        string `json:"ruler"`
	Material     string `json:"material"`
	Denomination string `json:"denomination"`
	EstPrice     string `json:"estPrice"`
	ImageURL     string `json:"imageUrl"`
	SourceURL    string `json:"sourceUrl"`
	SourceName   string `json:"sourceName"`
}

type AgentChatResponse struct {
	Message     string           `json:"message"`
	Suggestions []CoinSuggestion `json:"suggestions"`
}

// Anthropic API types

type anthropicRequest struct {
	Model     string             `json:"model"`
	MaxTokens int                `json:"max_tokens"`
	Stream    bool               `json:"stream"`
	System    string             `json:"system"`
	Tools     []anthropicTool    `json:"tools"`
	Messages  []anthropicMessage `json:"messages"`
}

type anthropicTool struct {
	Type    string `json:"type"`
	Name    string `json:"name"`
	MaxUses int    `json:"max_uses,omitempty"`
}

type anthropicMessage struct {
	Role    string      `json:"role"`
	Content interface{} `json:"content"`
}

type anthropicResponse struct {
	Content []anthropicContent `json:"content"`
	Error   *anthropicError    `json:"error,omitempty"`
}

type anthropicContent struct {
	Type string `json:"type"`
	Text string `json:"text,omitempty"`
}

type anthropicError struct {
	Type    string `json:"type"`
	Message string `json:"message"`
}

// SSE stream event types
type streamEvent struct {
	Type         string          `json:"type"`
	Index        int             `json:"index,omitempty"`
	ContentBlock *anthropicContent `json:"content_block,omitempty"`
	Delta        *streamDelta    `json:"delta,omitempty"`
}

type streamDelta struct {
	Type string `json:"type"`
	Text string `json:"text,omitempty"`
}

const DefaultAgentPrompt = `You are a knowledgeable numismatist with a focus on Greek and Roman coinage up through the Byzantine Era. You specialize in finding that rare gem of a coin for just the right price. You know how to spot a fake but also a great deal. You are enthusiastic but informative, helpful and friendly.

Core Capabilities:
- Discover coins that are CURRENTLY FOR SALE from reputable dealers and active auction listings
- Identify coins that match the user's requirements and pricing guidelines
- Verify that the seller has a great reputation
- Verify every link is real and currently accessible — never hallucinate or fabricate URLs
- Filter out unwanted or potentially fake coins

CRITICAL — Availability Rules:
- ONLY return coins that are CURRENTLY AVAILABLE FOR PURCHASE or in an UPCOMING/ACTIVE auction
- NEVER return sold items, past auction results, or price archives
- If a listing says "SOLD", "Auction ended", "Realized price", or similar — SKIP IT
- ACSSearch.info is an archive of PAST auction results — do NOT use it for current listings
- When searching, add keywords like "buy now", "for sale", "in stock", or "available" to your queries
- Verify each result page shows an active "Buy" or "Add to Cart" button, or an auction with a future end date

Website Hints (search these but also search beyond them):
- https://www.vcoins.com/ (dealer marketplace — items listed are for sale)
- https://www.forumancientcoins.com/ (dealer with direct sales)
- https://www.hjbltd.com/ (auction house)
- https://www.biddr.com/ (live auction aggregator — check auction dates)
- https://www.catawiki.com/ (online auctions — check if auction is active)
- https://www.ma-shops.com/ (dealer marketplace)

Important Rules:
1. ALWAYS use the web_search tool to find real, currently available coins. Never invent listings.
2. Every sourceUrl you return MUST be a real URL you found during your web search. If you cannot find a direct link, omit the suggestion entirely.
3. Verify each result came from your search results. Do not guess or construct URLs.
4. Include the actual listed price, not an estimate or a past realized price.
5. Mention the dealer/auction house reputation if known.
6. Flag any concerns about authenticity or condition.
7. For imageUrl: if you can see a direct image URL (ending in .jpg, .png, etc.) in your search results, include it. If you cannot find a direct image URL — for example because the page uses dynamic/lazy image loading — set imageUrl to an empty string "". The system will automatically extract the image from the listing page. Do NOT skip a coin suggestion just because you cannot find its image URL.
8. The sourceUrl is the most important link — always provide the direct URL to the listing page. The system uses it to extract images automatically when imageUrl is unavailable.

After searching, provide an enthusiastic but informative response about what you found. Include a JSON block with structured coin suggestions. The JSON block MUST be wrapped in ` + "```json" + ` and ` + "```" + ` markers.

The JSON should be an array of objects with these fields:
- name: Full coin name/title as listed by the seller
- description: Brief description including notable features, condition notes, and any authenticity observations
- category: One of "Roman", "Greek", "Byzantine", "Modern", or "Other"
- era: Time period (e.g., "27 BC - 14 AD")
- ruler: Ruler or authority (if applicable)
- material: One of "Gold", "Silver", "Bronze", "Copper", "Electrum", or "Other"
- denomination: Coin denomination (e.g., "Denarius", "Tetradrachm")
- estPrice: Actual listed price from the listing (e.g., "$150", "$200-300") — NOT a past realized price
- imageUrl: Direct URL to the coin image file if available, or empty string "" if not found (the system will auto-extract from the listing page)
- sourceUrl: Direct URL to the actual listing page (required — must be a real link from your search)
- sourceName: Name of the dealer, auction house, or website

Example format:
` + "```json" + `
[
  {
    "name": "Augustus AR Denarius - Lugdunum mint",
    "description": "Silver denarius of Augustus, laureate head right. Rev: Gaius and Lucius Caesars. Good VF, nice cabinet tone. Reputable dealer with 20+ years experience.",
    "category": "Roman",
    "era": "27 BC - 14 AD",
    "ruler": "Augustus",
    "material": "Silver",
    "denomination": "Denarius",
    "estPrice": "$275",
    "imageUrl": "https://www.vcoins.com/images/coin12345.jpg",
    "sourceUrl": "https://www.vcoins.com/en/stores/example/1234",
    "sourceName": "VCoins - Example Numismatics"
  }
]
` + "```" + `

Only include coins you actually found in your search results. Quality over quantity — 2 verified, currently available results are better than 5 sold or fabricated ones.

Portfolio Analysis:
When the user asks you to analyze their portfolio or collection, they will provide a portfolio summary with their collection composition (categories, materials, eras, rulers, top coins, total value). Use this data to:
1. Assess the collection's strengths and identify well-represented areas
2. Identify gaps — missing eras, under-represented categories, or rulers that would complement existing holdings
3. Suggest specific acquisitions that would diversify or strengthen the collection
4. Provide market context using web_search — are certain areas appreciating? Are there opportunities?
5. Consider budget based on average coin value in their collection

When doing portfolio analysis, DO NOT include a JSON suggestion block. Instead, provide a detailed written analysis with clear sections and actionable recommendations.

Coin Shows & Events:
When the user asks about upcoming coin shows, conventions, or numismatic events:
1. Use web_search to find upcoming coin shows, expos, and numismatic conventions
2. Focus on shows that feature ancient, Greek, Roman, or Byzantine coinage — but include major general numismatic shows as well
3. Search for events from organizations like ANA (American Numismatic Association), PNG, NYINC (New York International Numismatic Convention), and regional coin clubs
4. Website hints for coin show listings:
   - https://www.coinshows.com/ (comprehensive coin show directory)
   - https://www.money.org/ (ANA events and conventions)
   - https://www.pngdealers.org/ (PNG show schedule)
   - https://www.nyinc.info/ (NYINC annual convention)
   - https://www.biddr.com/ (live auctions tied to shows)
5. For each show found, provide: name, dates, location (city/venue), website link, and a brief note on relevance to ancient coin collectors
6. Only include shows with future dates — do not list past events
7. If the user mentions a location or region, prioritize shows near that area
8. Mention any notable dealers, auction events, or special exhibits associated with the show`

func (h *AgentHandler) getSystemPrompt(userID uint) string {
	prompt := services.GetSetting(services.SettingAgentPrompt)
	if prompt == "" {
		prompt = DefaultAgentPrompt
	}

	if user, err := h.userRepo.FindByID(userID); err == nil && user.ZipCode != "" {
		prompt = fmt.Sprintf("The user's location ZIP code is %s. Use this to prioritize nearby coin shows, dealers, and events when relevant.\n\n%s", user.ZipCode, prompt)
	}

	return prompt
}

// ChatStream handles a streaming conversation with the AI agent via SSE.
//
//	@Summary		Chat with coin search agent (streaming)
//	@Description	Send a message to the AI agent. Response is streamed as Server-Sent Events.
//	@Tags			Agent
//	@Accept			json
//	@Produce		text/event-stream
//	@Param			body	body		AgentChatRequest	true	"Chat message"
//	@Success		200		{object}	AgentChatResponse
//	@Failure		400		{object}	ErrorResponse
//	@Failure		503		{object}	ErrorResponse
//	@Security		BearerAuth
//	@Router			/agent/chat [post]
func (h *AgentHandler) ChatStream(c *gin.Context) {
	logger := services.AppLogger
	userID := c.GetUint("userId")

	apiKey := services.GetSetting(services.SettingAnthropicAPIKey)
	if apiKey == "" {
		c.JSON(http.StatusServiceUnavailable, gin.H{
			"error": "Anthropic API key not configured. Set it in Admin → Settings → AI Config.",
		})
		return
	}

	var req AgentChatRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Message is required"})
		return
	}

	model := services.GetSetting(services.SettingAnthropicModel)
	if model == "" {
		model = "claude-sonnet-4-20250514"
	}

	// Build messages array from history + new message
	messages := make([]anthropicMessage, 0, len(req.History)+1)
	for _, msg := range req.History {
		messages = append(messages, anthropicMessage{
			Role:    msg.Role,
			Content: msg.Content,
		})
	}
	messages = append(messages, anthropicMessage{
		Role:    "user",
		Content: req.Message,
	})

	anthropicReq := anthropicRequest{
		Model:     model,
		MaxTokens: 4096,
		Stream:    true,
		System:    h.getSystemPrompt(userID),
		Tools: []anthropicTool{
			{
				Type:    "web_search_20250305",
				Name:    "web_search",
				MaxUses: 20,
			},
		},
		Messages: messages,
	}

	body, err := json.Marshal(anthropicReq)
	if err != nil {
		logger.Error("agent", "Failed to marshal request: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to build request"})
		return
	}

	httpReq, err := http.NewRequest("POST", "https://api.anthropic.com/v1/messages", bytes.NewReader(body))
	if err != nil {
		logger.Error("agent", "Failed to create HTTP request: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create request"})
		return
	}
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("x-api-key", apiKey)
	httpReq.Header.Set("anthropic-version", "2023-06-01")

	resp, err := h.client.Do(httpReq)
	if err != nil {
		logger.Error("agent", "Anthropic API call failed: %v", err)
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "Failed to reach Anthropic API"})
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		respBody, _ := io.ReadAll(resp.Body)
		logger.Error("agent", "Anthropic API returned %d: %s", resp.StatusCode, string(respBody))
		c.JSON(http.StatusServiceUnavailable, gin.H{
			"error": fmt.Sprintf("Anthropic API error (HTTP %d)", resp.StatusCode),
		})
		return
	}

	// Set SSE headers
	c.Header("Content-Type", "text/event-stream")
	c.Header("Cache-Control", "no-cache")
	c.Header("Connection", "keep-alive")
	c.Header("X-Accel-Buffering", "no")

	flusher, ok := c.Writer.(http.Flusher)
	if !ok {
		logger.Error("agent", "Response writer does not support flushing")
		return
	}

	var fullText strings.Builder

	scanner := bufio.NewScanner(resp.Body)
	// Increase buffer for large SSE events
	buf := make([]byte, 0, 64*1024)
	scanner.Buffer(buf, 1024*1024)

	for scanner.Scan() {
		line := scanner.Text()

		if !strings.HasPrefix(line, "data: ") {
			continue
		}

		data := strings.TrimPrefix(line, "data: ")
		if data == "[DONE]" {
			break
		}

		var event streamEvent
		if err := json.Unmarshal([]byte(data), &event); err != nil {
			continue
		}

		switch event.Type {
		case "content_block_delta":
			if event.Delta != nil && event.Delta.Type == "text_delta" && event.Delta.Text != "" {
				fullText.WriteString(event.Delta.Text)
				// Send text chunk to client
				chunk, _ := json.Marshal(map[string]string{
					"type": "text",
					"text": event.Delta.Text,
				})
				fmt.Fprintf(c.Writer, "data: %s\n\n", chunk)
				flusher.Flush()
			}

		case "message_stop":
			// Parse suggestions from full accumulated text
			text := fullText.String()
			suggestions := extractSuggestions(text)
			cleanMessage := removeJSONBlock(text)

			done, _ := json.Marshal(map[string]interface{}{
				"type":        "done",
				"message":     cleanMessage,
				"suggestions": suggestions,
			})
			fmt.Fprintf(c.Writer, "data: %s\n\n", done)
			flusher.Flush()

			logger.Info("agent", "Stream complete: %d suggestions found", len(suggestions))
		}
	}
}

// ListModels returns the list of available Anthropic models.
//
//	@Summary		List available Anthropic models
//	@Description	Returns a curated list of Anthropic models suitable for the coin search agent.
//	@Tags			Agent
//	@Produce		json
//	@Success		200	{array}	object
//	@Security		BearerAuth
//	@Router			/agent/models [get]
func (h *AgentHandler) ListModels(c *gin.Context) {
	apiKey := services.GetSetting(services.SettingAnthropicAPIKey)
	if apiKey == "" {
		// Return defaults if no API key
		c.JSON(http.StatusOK, getDefaultModels())
		return
	}

	// Try to fetch from Anthropic API
	req, err := http.NewRequest("GET", "https://api.anthropic.com/v1/models", nil)
	if err != nil {
		c.JSON(http.StatusOK, getDefaultModels())
		return
	}
	req.Header.Set("x-api-key", apiKey)
	req.Header.Set("anthropic-version", "2023-06-01")

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil || resp.StatusCode != http.StatusOK {
		c.JSON(http.StatusOK, getDefaultModels())
		return
	}
	defer resp.Body.Close()

	var result struct {
		Data []struct {
			ID          string `json:"id"`
			DisplayName string `json:"display_name"`
		} `json:"data"`
	}
	body, _ := io.ReadAll(resp.Body)
	if err := json.Unmarshal(body, &result); err != nil || len(result.Data) == 0 {
		c.JSON(http.StatusOK, getDefaultModels())
		return
	}

	models := make([]map[string]string, 0, len(result.Data))
	for _, m := range result.Data {
		models = append(models, map[string]string{
			"id":   m.ID,
			"name": m.DisplayName,
		})
	}
	c.JSON(http.StatusOK, models)
}

func getDefaultModels() []map[string]string {
	return []map[string]string{
		{"id": "claude-sonnet-4-20250514", "name": "Claude Sonnet 4"},
		{"id": "claude-haiku-4-20250414", "name": "Claude Haiku 4"},
		{"id": "claude-opus-4-20250514", "name": "Claude Opus 4"},
	}
}

// GetPrompt returns the current agent prompt.
//
//	@Summary		Get agent prompt
//	@Tags			Agent
//	@Produce		json
//	@Success		200	{object}	object
//	@Security		BearerAuth
//	@Router			/agent/prompt [get]
func (h *AgentHandler) GetPrompt(c *gin.Context) {
	prompt := services.GetSetting(services.SettingAgentPrompt)
	if prompt == "" {
		prompt = DefaultAgentPrompt
	}
	c.JSON(http.StatusOK, gin.H{
		"prompt":  prompt,
		"default": DefaultAgentPrompt,
	})
}

// GetValuationPrompt returns the current valuation prompt.
//
//	@Summary		Get valuation prompt
//	@Tags			Agent
//	@Produce		json
//	@Success		200	{object}	object
//	@Security		BearerAuth
//	@Router			/agent/valuation-prompt [get]
func (h *AgentHandler) GetValuationPrompt(c *gin.Context) {
	prompt := services.GetSetting(services.SettingValuationPrompt)
	if prompt == "" {
		prompt = DefaultValuationPrompt
	}
	c.JSON(http.StatusOK, gin.H{
		"prompt":  prompt,
		"default": DefaultValuationPrompt,
	})
}

// PortfolioSummary returns aggregated collection data for AI portfolio analysis.
//
//	@Summary		Get portfolio summary for AI analysis
//	@Tags			Agent
//	@Produce		json
//	@Success		200	{object}	object
//	@Security		BearerAuth
//	@Router			/agent/portfolio-summary [get]
func (h *AgentHandler) PortfolioSummary(c *gin.Context) {
	userID := c.GetUint("userId")

	summary, _ := h.repo.GetPortfolioSummary(userID)

	c.JSON(http.StatusOK, gin.H{
		"totalCoins":    summary.TotalCoins,
		"totalValue":    summary.TotalValue,
		"totalInvested": summary.TotalInvested,
		"categories":    summary.Categories,
		"materials":     summary.Materials,
		"eras":          summary.Eras,
		"rulers":        summary.Rulers,
		"topCoins":      summary.TopCoins,
	})
}

// extractSuggestions finds a JSON array inside ```json ... ``` markers
func extractSuggestions(text string) []CoinSuggestion {
	start := -1
	end := -1

	// Find ```json marker
	jsonStart := "```json"
	jsonEnd := "```"

	startIdx := indexOf(text, jsonStart)
	if startIdx == -1 {
		return nil
	}
	start = startIdx + len(jsonStart)

	// Find closing ``` after the opening
	endIdx := indexOf(text[start:], jsonEnd)
	if endIdx == -1 {
		return nil
	}
	end = start + endIdx

	jsonStr := text[start:end]

	var suggestions []CoinSuggestion
	if err := json.Unmarshal([]byte(jsonStr), &suggestions); err != nil {
		return nil
	}
	return suggestions
}

// removeJSONBlock strips the ```json ... ``` block from the message
func removeJSONBlock(text string) string {
	jsonStart := "```json"
	jsonEnd := "```"

	startIdx := indexOf(text, jsonStart)
	if startIdx == -1 {
		return text
	}

	remaining := text[startIdx+len(jsonStart):]
	endIdx := indexOf(remaining, jsonEnd)
	if endIdx == -1 {
		return text
	}

	return text[:startIdx] + text[startIdx+len(jsonStart)+endIdx+len(jsonEnd):]
}

func indexOf(s, substr string) int {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return i
		}
	}
	return -1
}

// Value estimation types

type ValueComparable struct {
	Source string `json:"source"`
	Price  string `json:"price"`
	URL    string `json:"url"`
}

type ValueEstimateResponse struct {
	EstimatedValue float64           `json:"estimatedValue"`
	Confidence     string            `json:"confidence"`
	Reasoning      string            `json:"reasoning"`
	Comparables    []ValueComparable `json:"comparables"`
}

const DefaultValuationPrompt = `You are an expert numismatist and coin appraiser. Your task is to estimate the current fair market value of a specific coin based on its attributes.

Instructions:
1. Use web_search to find CURRENT listings and RECENT sales of comparable coins from reputable dealers and auction houses.
2. Focus on coins with similar: denomination, ruler, era, material, and grade/condition.
3. Check multiple sources: VCoins, MA-Shops, CNG, Heritage Auctions, Biddr, ForumAncientCoins.
4. Consider the grade/condition when comparing — a VF coin is worth less than an EF example.
5. If the coin has a purchase price, note whether it appears to have appreciated or depreciated.

Return your response as a JSON object (wrapped in ` + "```json" + ` and ` + "```" + ` markers) with these fields:
- estimatedValue: number (your best estimate in USD, as a single number — not a range)
- confidence: "high" (3+ comparable listings found), "medium" (1-2 comparables), or "low" (estimate based on general knowledge)
- reasoning: string (2-3 sentences explaining your valuation methodology and what you found)
- comparables: array of objects with { "source": "dealer/site name", "price": "$X" or "$X-Y", "url": "listing URL" }

Example:
` + "```json" + `
{
  "estimatedValue": 275,
  "confidence": "high",
  "reasoning": "Based on 4 current listings of Augustus denarii in similar VF condition, the market range is $250-300. Your coin's grade and strike quality place it at the mid-range.",
  "comparables": [
    { "source": "VCoins - Example Dealer", "price": "$285", "url": "https://www.vcoins.com/..." },
    { "source": "MA-Shops", "price": "$250", "url": "https://www.ma-shops.com/..." }
  ]
}
` + "```" + `

Only include real listings from your search. Do not fabricate URLs or prices.`

func (h *AgentHandler) getValuationPrompt() string {
	prompt := services.GetSetting(services.SettingValuationPrompt)
	if prompt == "" {
		return DefaultValuationPrompt
	}
	return prompt
}

// EstimateValue estimates the current market value of a coin using AI with web search.
func (h *AgentHandler) EstimateValue(c *gin.Context) {
	logger := services.AppLogger

	apiKey := services.GetSetting(services.SettingAnthropicAPIKey)
	if apiKey == "" {
		c.JSON(http.StatusServiceUnavailable, gin.H{
			"error": "Anthropic API key not configured. Set it in Admin → Settings → AI.",
		})
		return
	}

	coinID, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid coin ID"})
		return
	}

	userID := c.GetUint("userId")
	coin, err := h.repo.FindCoinForUser(uint(coinID), userID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Coin not found"})
		return
	}

	// Build description of the coin for the AI
	parts := []string{}
	if coin.Name != "" {
		parts = append(parts, fmt.Sprintf("Name: %s", coin.Name))
	}
	if coin.Category != "" {
		parts = append(parts, fmt.Sprintf("Category: %s", string(coin.Category)))
	}
	if coin.Denomination != "" {
		parts = append(parts, fmt.Sprintf("Denomination: %s", coin.Denomination))
	}
	if coin.Ruler != "" {
		parts = append(parts, fmt.Sprintf("Ruler: %s", coin.Ruler))
	}
	if coin.Era != "" {
		parts = append(parts, fmt.Sprintf("Era: %s", coin.Era))
	}
	if coin.Material != "" {
		parts = append(parts, fmt.Sprintf("Material: %s", string(coin.Material)))
	}
	if coin.Grade != "" {
		parts = append(parts, fmt.Sprintf("Grade/Condition: %s", coin.Grade))
	}
	if coin.WeightGrams != nil {
		parts = append(parts, fmt.Sprintf("Weight: %.2fg", *coin.WeightGrams))
	}
	if coin.DiameterMm != nil {
		parts = append(parts, fmt.Sprintf("Diameter: %.1fmm", *coin.DiameterMm))
	}
	if coin.RarityRating != "" {
		parts = append(parts, fmt.Sprintf("Rarity/RIC: %s", coin.RarityRating))
	}
	if coin.Mint != "" {
		parts = append(parts, fmt.Sprintf("Mint: %s", coin.Mint))
	}
	if coin.PurchasePrice != nil {
		parts = append(parts, fmt.Sprintf("Purchase Price: $%.2f", *coin.PurchasePrice))
	}

	userMessage := fmt.Sprintf("Please estimate the current market value of this coin:\n\n%s", strings.Join(parts, "\n"))

	model := services.GetSetting(services.SettingAnthropicModel)
	if model == "" {
		model = "claude-sonnet-4-20250514"
	}

	anthropicReq := anthropicRequest{
		Model:     model,
		MaxTokens: 4096,
		Stream:    false,
		System:    h.getValuationPrompt(),
		Tools: []anthropicTool{
			{
				Type:    "web_search_20250305",
				Name:    "web_search",
				MaxUses: 10,
			},
		},
		Messages: []anthropicMessage{
			{Role: "user", Content: userMessage},
		},
	}

	body, err := json.Marshal(anthropicReq)
	if err != nil {
		logger.Error("agent", "Failed to marshal estimate request: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to build request"})
		return
	}

	httpReq, err := http.NewRequest("POST", "https://api.anthropic.com/v1/messages", bytes.NewReader(body))
	if err != nil {
		logger.Error("agent", "Failed to create estimate HTTP request: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create request"})
		return
	}
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("x-api-key", apiKey)
	httpReq.Header.Set("anthropic-version", "2023-06-01")

	resp, err := h.client.Do(httpReq)
	if err != nil {
		logger.Error("agent", "Anthropic estimate API call failed: %v", err)
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "Failed to reach Anthropic API"})
		return
	}
	defer resp.Body.Close()

	respBody, _ := io.ReadAll(resp.Body)

	if resp.StatusCode != http.StatusOK {
		logger.Error("agent", "Anthropic estimate API returned %d: %s", resp.StatusCode, string(respBody))
		c.JSON(http.StatusServiceUnavailable, gin.H{
			"error": fmt.Sprintf("Anthropic API error (HTTP %d)", resp.StatusCode),
		})
		return
	}

	var apiResp anthropicResponse
	if err := json.Unmarshal(respBody, &apiResp); err != nil {
		logger.Error("agent", "Failed to parse estimate response: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to parse AI response"})
		return
	}

	// Extract text content from response
	var fullText string
	for _, block := range apiResp.Content {
		if block.Type == "text" {
			fullText += block.Text
		}
	}

	// Parse the JSON estimate from the response
	var estimate ValueEstimateResponse
	parsed := false
	jsonStart := indexOf(fullText, "```json")
	if jsonStart != -1 {
		start := jsonStart + len("```json")
		remaining := fullText[start:]
		jsonEnd := indexOf(remaining, "```")
		if jsonEnd != -1 {
			jsonStr := remaining[:jsonEnd]
			if err := json.Unmarshal([]byte(jsonStr), &estimate); err != nil {
				logger.Error("agent", "Failed to parse estimate JSON: %v — raw: %s", err, jsonStr)
			} else {
				parsed = true
			}
		}
	}

	// Fallback if JSON parsing failed — return raw text as reasoning
	if !parsed {
		estimate.Reasoning = fullText
		estimate.Confidence = "low"
	}

	// Auto-record value history entry
	if estimate.EstimatedValue > 0 {
		h.repo.RecordValueHistory(&models.CoinValueHistory{
			CoinID:     uint(coinID),
			UserID:     userID,
			Value:      estimate.EstimatedValue,
			Confidence: estimate.Confidence,
			RecordedAt: time.Now(),
		})
	}

	// Auto-add journal entry
	journalText := fmt.Sprintf("AI Value Estimate: $%.2f (%s confidence)", estimate.EstimatedValue, estimate.Confidence)
	h.repo.CreateJournalEntry(&models.CoinJournal{
		CoinID: uint(coinID),
		UserID: userID,
		Entry:  journalText,
	})

	c.JSON(http.StatusOK, estimate)
}
