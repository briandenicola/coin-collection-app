package services

import (
	"fmt"
	"io"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"regexp"
	"strconv"
	"strings"

	"golang.org/x/net/html"
)

const (
	numisbidsBase    = "https://www.numisbids.com"
	numisbidsLoginURL = numisbidsBase + "/registration/login.php"
	numisbidsWatchlistURL = numisbidsBase + "/watchlist"
	numisbidsUserAgent = "Mozilla/5.0 (Windows NT 10.0; Win64; x64) " +
		"AppleWebKit/537.36 (KHTML, like Gecko) " +
		"Chrome/131.0.0.0 Safari/537.36"
)

var (
	lotLinkRe    = regexp.MustCompile(`href="(/sale/(\d+)/lot/(\d+))"`)
	imgSrcRe     = regexp.MustCompile(`<img[^>]*src="([^"]*)"`)
	estimateRe   = regexp.MustCompile(`Estimate:\s*([\d,]+(?:\.\d+)?\s*\w+)`)
	currencyValRe = regexp.MustCompile(`([\d,]+(?:\.\d+)?)\s*(USD|EUR|GBP|CHF)`)
)

// WatchlistLot represents a single lot parsed from a NumisBids watchlist page.
type WatchlistLot struct {
	URL       string   `json:"url"`
	SaleID    string   `json:"saleId"`
	LotNumber int      `json:"lotNumber"`
	Title     string   `json:"title"`
	ImageURL  string   `json:"imageUrl"`
	Estimate  *float64 `json:"estimate"`
	Currency  string   `json:"currency"`
}

// NumisBidsService handles HTTP interactions with numisbids.com.
type NumisBidsService struct{}

// NewNumisBidsService creates a new NumisBidsService.
func NewNumisBidsService() *NumisBidsService {
	return &NumisBidsService{}
}

// Login authenticates with NumisBids and returns a cookie-jar-enabled client.
func (s *NumisBidsService) Login(username, password string) (*http.Client, error) {
	jar, err := cookiejar.New(nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create cookie jar: %w", err)
	}

	client := &http.Client{Jar: jar}

	form := url.Values{
		"email":    {username},
		"password": {password},
		"login":    {"Login"},
	}

	req, err := http.NewRequest("POST", numisbidsLoginURL, strings.NewReader(form.Encode()))
	if err != nil {
		return nil, fmt.Errorf("failed to create login request: %w", err)
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("User-Agent", numisbidsUserAgent)
	req.Header.Set("Referer", numisbidsBase+"/")
	req.Header.Set("X-Requested-With", "XMLHttpRequest")

	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("login request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusFound {
		return nil, fmt.Errorf("login returned HTTP %d", resp.StatusCode)
	}

	// Read body to check for login errors (AJAX form returns HTML fragment)
	body, _ := io.ReadAll(resp.Body)
	bodyStr := string(body)
	if strings.Contains(bodyStr, "Incorrect") || strings.Contains(bodyStr, "invalid") {
		return nil, fmt.Errorf("invalid credentials")
	}

	// Verify session cookie was set by checking for a PHPSESSID or similar
	parsedURL, _ := url.Parse(numisbidsBase)
	if len(jar.Cookies(parsedURL)) == 0 {
		return nil, fmt.Errorf("no session cookie received — login may have failed")
	}

	return client, nil
}

// FetchWatchlist retrieves the authenticated user's watchlist HTML.
func (s *NumisBidsService) FetchWatchlist(client *http.Client) (string, error) {
	req, err := http.NewRequest("GET", numisbidsWatchlistURL, nil)
	if err != nil {
		return "", fmt.Errorf("failed to create watchlist request: %w", err)
	}
	req.Header.Set("User-Agent", numisbidsUserAgent)

	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("watchlist request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("watchlist returned HTTP %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read watchlist body: %w", err)
	}

	return string(body), nil
}

// ParseWatchlist extracts lot data from NumisBids watchlist HTML.
// Mirrors the Python scrape_numisbids_watchlist logic.
func (s *NumisBidsService) ParseWatchlist(rawHTML string) []WatchlistLot {
	// Find all lot link positions, then extract the block between each pair.
	// Go's regexp engine (RE2) doesn't support lookaheads, so we split manually.
	indices := lotLinkRe.FindAllStringIndex(rawHTML, -1)

	var lots []WatchlistLot
	for i, idx := range indices {
		start := idx[0]
		end := len(rawHTML)
		if i+1 < len(indices) {
			end = indices[i+1][0]
		}
		block := rawHTML[start:end]

		linkMatch := lotLinkRe.FindStringSubmatch(block)
		if linkMatch == nil {
			continue
		}

		lotNum, _ := strconv.Atoi(linkMatch[3])
		lot := WatchlistLot{
			URL:       numisbidsBase + linkMatch[1],
			SaleID:    linkMatch[2],
			LotNumber: lotNum,
			Currency:  "USD",
		}

		// Image URL
		if imgMatch := imgSrcRe.FindStringSubmatch(block); imgMatch != nil {
			imgURL := imgMatch[1]
			if strings.HasPrefix(imgURL, "//") {
				imgURL = "https:" + imgURL
			}
			lot.ImageURL = imgURL
		}

		// Title: extract only the text inside the lot anchor tag
		lot.Title = extractLotTitle(block, linkMatch[1])
		if len(lot.Title) > 200 {
			lot.Title = lot.Title[:200]
		}

		// Estimate
		if estMatch := estimateRe.FindStringSubmatch(block); estMatch != nil {
			val, cur := parseCurrencyValue(estMatch[1])
			lot.Estimate = val
			if cur != "" {
				lot.Currency = cur
			}
		}

		lots = append(lots, lot)
	}

	return lots
}

// parseCurrencyValue extracts a numeric value and currency code from a string
// like "150 USD" or "1,200.50 EUR".
func parseCurrencyValue(text string) (*float64, string) {
	match := currencyValRe.FindStringSubmatch(text)
	if match == nil {
		return nil, "USD"
	}
	numStr := strings.ReplaceAll(match[1], ",", "")
	val, err := strconv.ParseFloat(numStr, 64)
	if err != nil {
		return nil, match[2]
	}
	return &val, match[2]
}

// extractLotTitle extracts the text content of the anchor tag that links to the lot.
// It walks the HTML tokens looking for <a href="...lotPath...">, then collects
// text until the closing </a> tag.
func extractLotTitle(block, lotPath string) string {
	tokenizer := html.NewTokenizer(strings.NewReader(block))
	inLotLink := false
	var result strings.Builder

	for {
		tt := tokenizer.Next()
		if tt == html.ErrorToken {
			break
		}

		switch tt {
		case html.StartTagToken:
			t := tokenizer.Token()
			if t.Data == "a" && !inLotLink {
				for _, attr := range t.Attr {
					if attr.Key == "href" && attr.Val == lotPath {
						inLotLink = true
						break
					}
				}
			}
		case html.EndTagToken:
			if inLotLink && tokenizer.Token().Data == "a" {
				goto done
			}
		case html.TextToken:
			if inLotLink {
				result.WriteString(tokenizer.Token().Data)
			}
		}
	}

done:
	text := strings.Join(strings.Fields(result.String()), " ")
	return strings.TrimSpace(text)
}

// cleanHTML strips HTML tags and normalizes whitespace.
func cleanHTML(s string) string {
	tokenizer := html.NewTokenizer(strings.NewReader(s))
	var result strings.Builder
	for {
		tt := tokenizer.Next()
		if tt == html.ErrorToken {
			break
		}
		if tt == html.TextToken {
			result.WriteString(tokenizer.Token().Data)
		}
	}
	// Normalize whitespace
	text := result.String()
	text = strings.Join(strings.Fields(text), " ")
	return strings.TrimSpace(text)
}
