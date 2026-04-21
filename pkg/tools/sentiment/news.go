package sentiment

import (
	"context"
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/alorse/trading-cli/internal/config"
	"github.com/alorse/trading-cli/pkg/client"
	"github.com/alorse/trading-cli/pkg/utils"
)

// NewsItem represents a news article for JSON output
type NewsItem struct {
	Title     string    `json:"title"`
	URL       string    `json:"url"`
	Published time.Time `json:"published"`
	Summary   string    `json:"summary"`
	Source    string    `json:"source"`
}

// NewsOutput represents the JSON output for financial news
type NewsOutput struct {
	Symbol    string      `json:"symbol"`
	Category  string      `json:"category"`
	Count     int         `json:"count"`
	Items     []NewsItem  `json:"items"`
	Timestamp time.Time   `json:"timestamp"`
}

// RunFinancialNews fetches and displays financial news based on symbol and category
func RunFinancialNews(cfg *config.Config, symbol string, category string, limit int) error {
	httpClient := client.NewHTTPClient(cfg)
	rssClient := client.NewRSSClient(httpClient)

	ctx, cancel := context.WithTimeout(context.Background(), cfg.HTTPTimeout)
	defer cancel()

	// Select feeds based on category
	feeds := selectFeeds(category)
	if len(feeds) == 0 {
		return fmt.Errorf("invalid category: %s", category)
	}

	// Fetch all feeds
	items, err := rssClient.FetchMultiple(ctx, feeds)
	if err != nil {
		return fmt.Errorf("failed to fetch RSS feeds: %w", err)
	}

	// Filter by symbol if provided
	if symbol != "" {
		items = filterBySymbol(items, symbol)
	}

	// Strip HTML and truncate descriptions
	for i := range items {
		items[i].Description = stripHTML(items[i].Description)
		if len(items[i].Description) > 300 {
			items[i].Description = items[i].Description[:300]
		}
	}

	// Take first `limit` items
	if limit > 0 && len(items) > limit {
		items = items[:limit]
	}

	// Convert to output format
	newsItems := make([]NewsItem, len(items))
	for i, item := range items {
		newsItems[i] = NewsItem{
			Title:     item.Title,
			URL:       item.Link,
			Published: item.PubDate,
			Summary:   item.Description,
			Source:    item.Source,
		}
	}

	output := NewsOutput{
		Symbol:    symbol,
		Category:  category,
		Count:     len(newsItems),
		Items:     newsItems,
		Timestamp: time.Now().UTC(),
	}

	return utils.PrintJSON(output)
}

// selectFeeds returns the list of feeds to use based on category
func selectFeeds(category string) []string {
	switch strings.ToLower(category) {
	case "crypto":
		return []string{"coindesk", "cointelegraph"}
	case "stocks":
		return []string{"reuters-business", "reuters-companies"}
	case "all":
		return []string{"coindesk", "cointelegraph", "reuters-business", "reuters-companies"}
	default:
		return []string{}
	}
}

// filterBySymbol filters items by symbol (case-insensitive substring match)
func filterBySymbol(items []client.RSSItem, symbol string) []client.RSSItem {
	if symbol == "" {
		return items
	}

	var filtered []client.RSSItem
	symbolUpper := strings.ToUpper(symbol)

	for _, item := range items {
		if strings.Contains(strings.ToUpper(item.Title), symbolUpper) ||
			strings.Contains(strings.ToUpper(item.Description), symbolUpper) {
			filtered = append(filtered, item)
		}
	}

	return filtered
}

// stripHTML removes HTML tags from a string
func stripHTML(html string) string {
	re := regexp.MustCompile(`<[^>]+>`)
	return re.ReplaceAllString(html, "")
}
