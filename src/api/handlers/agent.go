package handlers

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/briandenicola/ancient-coins-api/repository"
	"github.com/briandenicola/ancient-coins-api/services"
	"github.com/gin-gonic/gin"
)

type AgentHandler struct {
	repo     *repository.AgentRepository
	userRepo *repository.UserRepository
	proxy    *services.AgentProxy
}

func NewAgentHandler(repo *repository.AgentRepository, userRepo *repository.UserRepository, proxy *services.AgentProxy) *AgentHandler {
	return &AgentHandler{
		repo:     repo,
		userRepo: userRepo,
		proxy:    proxy,
	}
}

// resolveLLMConfig reads the explicit AIProvider setting and builds LLMConfig.
// Returns an error message if the provider is not configured.
func resolveLLMConfig() (services.LLMConfig, string) {
	provider := services.GetSetting(services.SettingAIProvider)
	if provider == "" {
		return services.LLMConfig{}, "AI provider not configured. Please select Anthropic or Ollama in Admin Settings."
	}

	cfg := services.LLMConfig{
		Provider:   provider,
		OllamaURL:  services.GetSetting(services.SettingOllamaURL),
		SearXNGURL: services.GetSetting(services.SettingSearXNGURL),
	}

	switch provider {
	case "anthropic":
		cfg.APIKey = services.GetSetting(services.SettingAnthropicAPIKey)
		cfg.Model = services.GetSetting(services.SettingAnthropicModel)
		if cfg.APIKey == "" {
			return services.LLMConfig{}, "Anthropic API key is required. Configure it in Admin Settings."
		}
	case "ollama":
		cfg.Model = services.GetSetting(services.SettingOllamaModel)
		if cfg.OllamaURL == "" {
			return services.LLMConfig{}, "Ollama URL is required. Configure it in Admin Settings."
		}
	default:
		return services.LLMConfig{}, "Invalid AI provider. Please select Anthropic or Ollama in Admin Settings."
	}

	if cfg.Model == "" {
		cfg.Model = "claude-sonnet-4-20250514"
	}

	return cfg, ""
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



const DefaultCoinSearchPrompt = `You are a numismatic search specialist focused on Greek and Roman coinage up through the Byzantine Era. You specialize in finding that rare gem of a coin for just the right price.

CRITICAL RULES:
- Search for coins that are CURRENTLY FOR SALE — never return sold items or past auction results
- ONLY search reputable dealer sites: vcoins.com, ma-shops.com, forumancientcoins.com, biddr.com, catawiki.com, hjbltd.com
- Add "for sale" or "buy now" to your search queries
- For EACH result, you MUST provide the exact URL to the listing page
- NEVER invent, guess, or recall URLs from memory — only use URLs from search results
- Return ONLY results you actually found in your search
- If a listing says "SOLD", "Auction ended", or "Realized price" — SKIP IT
- ACSSearch.info is a PAST auction archive — do NOT use it
- Quality over quantity — 2 verified, available results beat 5 questionable ones
- Flag any concerns about authenticity or condition
- Mention dealer/auction house reputation if known`

const DefaultCoinShowsPrompt = `You are a coin show search specialist focused on numismatic conventions and collecting events.

Search for upcoming coin shows, expos, and conventions, especially those featuring ancient, Greek, Roman, or Byzantine coinage. Also include major general numismatic shows.

Key organizations and websites to search:
- coinshows.com (comprehensive coin show directory)
- money.org (ANA — American Numismatic Association events)
- pngdealers.org (PNG show schedule)
- nyinc.info (New York International Numismatic Convention)
- biddr.com (live auctions tied to shows)

CRITICAL RULES:
- ONLY return shows with FUTURE dates — never list past events
- Focus on shows within the next 30 days unless the user specifies a different timeframe
- When the user says "near me" or asks for nearby shows, only include shows within approximately 50 miles of their location
- For each show, find: name, dates, location, venue, website URL
- If the user mentions a location, prioritize shows near that area
- Note any special exhibits, notable dealers, or auction events at the show`

func (h *AgentHandler) getCoinSearchPrompt() string {
	prompt := services.GetSetting(services.SettingCoinSearchPrompt)
	if prompt == "" {
		prompt = DefaultCoinSearchPrompt
	}
	return prompt
}

func (h *AgentHandler) getCoinShowsPrompt(userID uint) string {
	prompt := services.GetSetting(services.SettingCoinShowsPrompt)
	if prompt == "" {
		prompt = DefaultCoinShowsPrompt
	}

	// Inject current date so the agent knows what "upcoming" means
	now := time.Now()
	deadline := now.AddDate(0, 0, 30)
	datePreamble := fmt.Sprintf("Today's date is %s. Unless the user specifies a different timeframe, only return shows between now and %s (next 30 days).\n\n",
		now.Format("January 2, 2006"), deadline.Format("January 2, 2006"))
	prompt = datePreamble + prompt

	if user, err := h.userRepo.FindByID(userID); err == nil && user.ZipCode != "" {
		prompt = fmt.Sprintf("The user's location ZIP code is %s. Use this to prioritize nearby coin shows, dealers, and events when relevant.\n\n%s", user.ZipCode, prompt)
	} else {
		prompt = fmt.Sprintf("The user has not set a ZIP code. If they ask about nearby coin shows or local events, ask them where they are located before searching.\n\n%s", prompt)
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

	var req AgentChatRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Message is required"})
		return
	}

	// Resolve LLM provider from explicit setting
	llmCfg, errMsg := resolveLLMConfig()
	if errMsg != "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": errMsg})
		return
	}

	// Build history for the proxy
	history := make([]services.ChatMessageProxy, 0, len(req.History))
	for _, msg := range req.History {
		history = append(history, services.ChatMessageProxy{
			Role:    msg.Role,
			Content: msg.Content,
		})
	}

	// Look up user zip code for location context
	var zipCode string
	if user, err := h.userRepo.FindByID(userID); err == nil {
		zipCode = user.ZipCode
	}

	proxyReq := services.AgentChatProxyRequest{
		LLM: llmCfg,
		User: services.UserContextProxy{
			UserID:  userID,
			ZipCode: zipCode,
		},
		Message:          req.Message,
		History:          history,
		CoinSearchPrompt: h.getCoinSearchPrompt(),
		CoinShowsPrompt:  h.getCoinShowsPrompt(userID),
	}

	if err := h.proxy.StreamChat(c.Request.Context(), c.Writer, proxyReq); err != nil {
		logger.Error("agent", "Chat stream proxy failed: %v", err)
		// Only send JSON error if headers haven't been sent yet
		if !c.Writer.Written() {
			c.JSON(http.StatusServiceUnavailable, gin.H{"error": "Agent service unavailable"})
		}
	}
}

// AgentStatus returns the current AI provider configuration status.
func (h *AgentHandler) AgentStatus(c *gin.Context) {
	provider := services.GetSetting(services.SettingAIProvider)
	configured := provider == "anthropic" || provider == "ollama"

	c.JSON(http.StatusOK, gin.H{
		"provider":   provider,
		"configured": configured,
	})
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

// GetCoinSearchPrompt returns the current coin search prompt.
//
//	@Summary		Get coin search prompt
//	@Tags			Agent
//	@Produce		json
//	@Success		200	{object}	object
//	@Security		BearerAuth
//	@Router			/agent/coin-search-prompt [get]
func (h *AgentHandler) GetCoinSearchPrompt(c *gin.Context) {
	prompt := services.GetSetting(services.SettingCoinSearchPrompt)
	if prompt == "" {
		prompt = DefaultCoinSearchPrompt
	}
	c.JSON(http.StatusOK, gin.H{
		"prompt":  prompt,
		"default": DefaultCoinSearchPrompt,
	})
}

// GetCoinShowsPrompt returns the current coin shows prompt.
//
//	@Summary		Get coin shows prompt
//	@Tags			Agent
//	@Produce		json
//	@Success		200	{object}	object
//	@Security		BearerAuth
//	@Router			/agent/coin-shows-prompt [get]
func (h *AgentHandler) GetCoinShowsPrompt(c *gin.Context) {
	prompt := services.GetSetting(services.SettingCoinShowsPrompt)
	if prompt == "" {
		prompt = DefaultCoinShowsPrompt
	}
	c.JSON(http.StatusOK, gin.H{
		"prompt":  prompt,
		"default": DefaultCoinShowsPrompt,
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

// EstimateValue estimates the current market value of a coin using the agent service.
func (h *AgentHandler) EstimateValue(c *gin.Context) {
	logger := services.AppLogger

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

	// Resolve LLM provider from explicit setting
	llmCfg, errMsg := resolveLLMConfig()
	if errMsg != "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": errMsg})
		return
	}

	var zipCode string
	if user, err := h.userRepo.FindByID(userID); err == nil {
		zipCode = user.ZipCode
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

	proxyReq := services.PortfolioReviewProxyRequest{
		LLM: llmCfg,
		User: services.UserContextProxy{
			UserID:  userID,
			ZipCode: zipCode,
		},
		Message:         userMessage,
		ValuationPrompt: h.getValuationPrompt(),
	}

	if err := h.proxy.StreamPortfolioReview(c.Request.Context(), c.Writer, proxyReq); err != nil {
		logger.Error("agent", "EstimateValue proxy failed: %v", err)
		if !c.Writer.Written() {
			c.JSON(http.StatusServiceUnavailable, gin.H{"error": "Agent service unavailable"})
		}
		return
	}

	// Auto-record value history and journal entry are now handled
	// by the Python agent service or by the frontend after parsing the SSE response.
}
