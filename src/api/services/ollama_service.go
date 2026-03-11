package services

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/briandenicola/ancient-coins-api/models"
)

type OllamaService struct {
	BaseURL string
	Client  *http.Client
}

type ollamaRequest struct {
	Model  string    `json:"model"`
	Prompt string    `json:"prompt"`
	Images []string  `json:"images"`
	Stream bool      `json:"stream"`
}

type ollamaResponse struct {
	Response string `json:"response"`
	Done     bool   `json:"done"`
}

func NewOllamaService(baseURL string) *OllamaService {
	return &OllamaService{
		BaseURL: strings.TrimRight(baseURL, "/"),
		Client: &http.Client{
			Timeout: 120 * time.Second,
		},
	}
}

func (s *OllamaService) AnalyzeCoinImages(imagePaths []string, coin models.Coin) (string, error) {
	var base64Images []string

	for _, path := range imagePaths {
		data, err := os.ReadFile(path)
		if err != nil {
			continue
		}
		base64Images = append(base64Images, base64.StdEncoding.EncodeToString(data))
	}

	if len(base64Images) == 0 {
		return "", fmt.Errorf("no valid images found")
	}

	prompt := buildPrompt(coin)

	reqBody := ollamaRequest{
		Model:  "llava",
		Prompt: prompt,
		Images: base64Images,
		Stream: false,
	}

	jsonBody, err := json.Marshal(reqBody)
	if err != nil {
		return "", fmt.Errorf("failed to marshal request: %w", err)
	}

	resp, err := s.Client.Post(
		s.BaseURL+"/api/generate",
		"application/json",
		bytes.NewReader(jsonBody),
	)
	if err != nil {
		return "", fmt.Errorf("failed to call Ollama: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("Ollama returned status %d: %s", resp.StatusCode, string(body))
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read response: %w", err)
	}

	var result ollamaResponse
	if err := json.Unmarshal(body, &result); err != nil {
		return "", fmt.Errorf("failed to parse response: %w", err)
	}

	return result.Response, nil
}

func buildPrompt(coin models.Coin) string {
	var sb strings.Builder
	sb.WriteString("You are an expert numismatist specializing in ancient and modern coins. ")
	sb.WriteString("Analyze the coin image(s) provided and give a detailed assessment.\n\n")

	if coin.Name != "" {
		sb.WriteString(fmt.Sprintf("The coin is cataloged as: %s\n", coin.Name))
	}
	if coin.Category != "" {
		sb.WriteString(fmt.Sprintf("Category: %s\n", coin.Category))
	}
	if coin.Denomination != "" {
		sb.WriteString(fmt.Sprintf("Denomination: %s\n", coin.Denomination))
	}
	if coin.Ruler != "" {
		sb.WriteString(fmt.Sprintf("Ruler: %s\n", coin.Ruler))
	}

	sb.WriteString("\nPlease provide:\n")
	sb.WriteString("1. **Identification**: Confirm or correct the identification of the coin\n")
	sb.WriteString("2. **Obverse Description**: Describe what you see on the obverse (front)\n")
	sb.WriteString("3. **Reverse Description**: Describe what you see on the reverse (back)\n")
	sb.WriteString("4. **Condition Assessment**: Assess the coin's condition/grade\n")
	sb.WriteString("5. **Inscriptions**: Read any visible inscriptions\n")
	sb.WriteString("6. **Historical Context**: Brief historical context about this coin\n")
	sb.WriteString("7. **Notable Features**: Any die varieties, errors, or notable features\n")
	sb.WriteString("8. **Authenticity Notes**: Any observations about authenticity\n")

	return sb.String()
}
