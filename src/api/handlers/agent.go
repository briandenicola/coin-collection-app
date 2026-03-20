package handlers

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/briandenicola/ancient-coins-api/services"
	"github.com/gin-gonic/gin"
)

type AgentHandler struct {
	client *http.Client
}

func NewAgentHandler() *AgentHandler {
	return &AgentHandler{
		client: &http.Client{Timeout: 120 * time.Second},
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
	Model     string               `json:"model"`
	MaxTokens int                  `json:"max_tokens"`
	System    string               `json:"system"`
	Tools     []anthropicTool      `json:"tools"`
	Messages  []anthropicMessage   `json:"messages"`
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

const agentSystemPrompt = `You are a knowledgeable numismatist with a focus on Greek and Roman coinage up through the Byzantine Era. You specialize in finding that rare gem of a coin for just the right price. You know how to spot a fake but also a great deal. You are enthusiastic but informative, helpful and friendly.

Core Capabilities:
- Discover current and upcoming auctions and dealer listings
- Identify coins that match the user's requirements and pricing guidelines
- Verify that the seller has a great reputation
- Verify every link is real and currently accessible — never hallucinate or fabricate URLs
- Filter out unwanted or potentially fake coins

Website Hints (search these but also search beyond them):
- https://www.forumancientcoins.com/
- https://www.vcoins.com/
- https://www.hjbltd.com/
- https://www.acsearch.info/
- https://www.biddr.com/
- https://www.catawiki.com/

Important Rules:
1. ALWAYS use the web_search tool to find real, currently available coins. Never invent listings.
2. Every sourceUrl you return MUST be a real URL you found during your web search. If you cannot find a direct link, omit the suggestion entirely.
3. Verify each result came from your search results. Do not guess or construct URLs.
4. Include the actual price from the listing when available, not an estimate.
5. Mention the dealer/auction house reputation if known.
6. Flag any concerns about authenticity or condition.
7. ALWAYS include an imageUrl for every coin suggestion. Look for the coin photo URL in the listing or search results. The image is used as the obverse photo when added to the wishlist. Only omit imageUrl if absolutely no image exists.

After searching, provide an enthusiastic but informative response about what you found. Include a JSON block with structured coin suggestions. The JSON block MUST be wrapped in ` + "```json" + ` and ` + "```" + ` markers.

The JSON should be an array of objects with these fields:
- name: Full coin name/title as listed by the seller
- description: Brief description including notable features, condition notes, and any authenticity observations
- category: One of "Roman", "Greek", "Byzantine", "Modern", or "Other"
- era: Time period (e.g., "27 BC - 14 AD")
- ruler: Ruler or authority (if applicable)
- material: One of "Gold", "Silver", "Bronze", "Copper", "Electrum", or "Other"
- denomination: Coin denomination (e.g., "Denarius", "Tetradrachm")
- estPrice: Actual listed price or price range from the listing (e.g., "$150", "$200-300")
- imageUrl: URL to the coin image from the listing (required — the image is saved as the coin's obverse photo when added to the wishlist)
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
    "imageUrl": "",
    "sourceUrl": "https://www.vcoins.com/en/stores/example/1234",
    "sourceName": "VCoins - Example Numismatics"
  }
]
` + "```" + `

Only include coins you actually found in your search results. Quality over quantity — 2 verified results are better than 5 fabricated ones.`

// Chat handles a conversation with the AI agent.
//
//	@Summary		Chat with coin search agent
//	@Description	Send a message to the AI agent that searches the web for coins matching your description.
//	@Tags			Agent
//	@Accept			json
//	@Produce		json
//	@Param			body	body		AgentChatRequest	true	"Chat message"
//	@Success		200		{object}	AgentChatResponse
//	@Failure		400		{object}	ErrorResponse
//	@Failure		503		{object}	ErrorResponse
//	@Security		BearerAuth
//	@Router			/agent/chat [post]
func (h *AgentHandler) Chat(c *gin.Context) {
	logger := services.AppLogger

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
		System:    agentSystemPrompt,
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

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		logger.Error("agent", "Failed to read response: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to read response"})
		return
	}

	if resp.StatusCode != http.StatusOK {
		logger.Error("agent", "Anthropic API returned %d: %s", resp.StatusCode, string(respBody))
		c.JSON(http.StatusServiceUnavailable, gin.H{
			"error": fmt.Sprintf("Anthropic API error (HTTP %d)", resp.StatusCode),
		})
		return
	}

	var anthropicResp anthropicResponse
	if err := json.Unmarshal(respBody, &anthropicResp); err != nil {
		logger.Error("agent", "Failed to parse response: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to parse response"})
		return
	}

	if anthropicResp.Error != nil {
		logger.Error("agent", "Anthropic error: %s", anthropicResp.Error.Message)
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": anthropicResp.Error.Message})
		return
	}

	// Extract text content from response
	var fullText string
	for _, block := range anthropicResp.Content {
		if block.Type == "text" {
			fullText += block.Text
		}
	}

	// Parse coin suggestions from JSON block in the response
	suggestions := extractSuggestions(fullText)

	// Clean the message text by removing the JSON block
	cleanMessage := removeJSONBlock(fullText)

	logger.Info("agent", "Chat response: %d suggestions found", len(suggestions))

	c.JSON(http.StatusOK, AgentChatResponse{
		Message:     cleanMessage,
		Suggestions: suggestions,
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
