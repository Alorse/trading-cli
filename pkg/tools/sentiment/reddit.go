package sentiment

import (
	"context"
	"fmt"
	"math"
	"sort"
	"strings"
	"time"

	"github.com/alorse/trading-cli/internal/config"
	"github.com/alorse/trading-cli/pkg/client"
	"github.com/alorse/trading-cli/pkg/utils"
)

// RedditPostScore holds a post with its calculated sentiment score
type RedditPostScore struct {
	Post  client.RedditPost
	Score float64
}

// RedditTopPost represents a top post for JSON output
type RedditTopPost struct {
	Title        string  `json:"title"`
	Subreddit    string  `json:"subreddit"`
	Score        int     `json:"score"`
	NumComments  int     `json:"numComments"`
	URL          string  `json:"url"`
	SentimentScore float64 `json:"sentimentScore"`
}

// SentimentOutput represents the JSON output for market sentiment analysis
type SentimentOutput struct {
	Symbol          string          `json:"symbol"`
	SentimentScore  float64         `json:"sentimentScore"`
	SentimentLabel  string          `json:"sentimentLabel"`
	PostsAnalyzed   int             `json:"postsAnalyzed"`
	BullishCount    int             `json:"bullishCount"`
	BearishCount    int             `json:"bearishCount"`
	NeutralCount    int             `json:"neutralCount"`
	TopPosts        []RedditTopPost `json:"topPosts"`
	Sources         []string        `json:"sources"`
	Timestamp       time.Time       `json:"timestamp"`
}

// RunMarketSentiment analyzes market sentiment from Reddit
func RunMarketSentiment(cfg *config.Config, symbol string, category string, limit int) error {
	httpClient := client.NewHTTPClient(cfg)
	redditClient := client.NewRedditClient(httpClient)

	ctx, cancel := context.WithTimeout(context.Background(), cfg.HTTPTimeout)
	defer cancel()

	// Select subreddits based on category
	subreddits := selectSubreddits(category)
	if len(subreddits) == 0 {
		return fmt.Errorf("invalid category: %s", category)
	}

	// Search each subreddit for the symbol
	var allPosts []client.RedditPost
	postsPerSubreddit := limit / len(subreddits)
	if postsPerSubreddit <= 0 {
		postsPerSubreddit = 1
	}

	// Track which subreddits had posts for sources
	sourcesMap := make(map[string]bool)

	for _, subreddit := range subreddits {
		posts, err := redditClient.Search(ctx, subreddit, symbol, postsPerSubreddit)
		if err != nil {
			continue // Skip failed subreddits
		}
		if len(posts) > 0 {
			sourcesMap[subreddit] = true
			allPosts = append(allPosts, posts...)
		}
	}

	if len(allPosts) == 0 {
		// Return empty response with zeros
		output := SentimentOutput{
			Symbol:         symbol,
			SentimentScore: 0.0,
			SentimentLabel: "Neutral",
			PostsAnalyzed:  0,
			BullishCount:   0,
			BearishCount:   0,
			NeutralCount:   0,
			TopPosts:       []RedditTopPost{},
			Sources:        []string{},
			Timestamp:      time.Now().UTC(),
		}
		return utils.PrintJSON(output)
	}

	// Score each post based on sentiment keywords
	scoredPosts := make([]RedditPostScore, len(allPosts))
	bullishCount := 0
	bearishCount := 0
	neutralCount := 0
	totalScore := 0.0

	for i, post := range allPosts {
		text := post.Title + " " + post.Selftext
		bullCount, bearCount := countKeywords(text)
		postScore := calculateSentimentScore(bullCount, bearCount)

		scoredPosts[i] = RedditPostScore{
			Post:  post,
			Score: postScore,
		}

		// Categorize
		if postScore > 0.2 || postScore < -0.2 {
			if postScore > 0 {
				bullishCount++
			} else {
				bearishCount++
			}
		} else {
			neutralCount++
		}

		totalScore += postScore
	}

	// Calculate aggregate sentiment score
	aggregateScore := totalScore / float64(len(allPosts))

	// Determine label
	label := determineSentimentLabel(aggregateScore)

	// Get top 5 posts
	sort.Slice(scoredPosts, func(i, j int) bool {
		return math.Abs(scoredPosts[i].Score) > math.Abs(scoredPosts[j].Score)
	})

	topPosts := make([]RedditTopPost, 0, 5)
	for i := 0; i < len(scoredPosts) && i < 5; i++ {
		sp := scoredPosts[i]
		topPosts = append(topPosts, RedditTopPost{
			Title:           sp.Post.Title,
			Subreddit:       sp.Post.Subreddit,
			Score:           sp.Post.Score,
			NumComments:     sp.Post.NumComments,
			URL:             sp.Post.URL,
			SentimentScore:  sp.Score,
		})
	}

	// Collect sources
	var sources []string
	for sr := range sourcesMap {
		sources = append(sources, "r/"+sr)
	}
	sort.Strings(sources)

	output := SentimentOutput{
		Symbol:         symbol,
		SentimentScore: aggregateScore,
		SentimentLabel: label,
		PostsAnalyzed:  len(allPosts),
		BullishCount:   bullishCount,
		BearishCount:   bearishCount,
		NeutralCount:   neutralCount,
		TopPosts:       topPosts,
		Sources:        sources,
		Timestamp:      time.Now().UTC(),
	}

	return utils.PrintJSON(output)
}

// selectSubreddits returns the list of subreddits to search based on category
func selectSubreddits(category string) []string {
	switch strings.ToLower(category) {
	case "crypto":
		return []string{"CryptoCurrency", "Bitcoin", "ethereum", "CryptoMarkets", "altcoin"}
	case "stocks":
		return []string{"stocks", "investing", "wallstreetbets", "StockMarket", "ValueInvesting"}
	case "all":
		return []string{"wallstreetbets", "stocks", "investing", "CryptoCurrency", "StockMarket"}
	default:
		return []string{}
	}
}

// countKeywords counts bull and bear keywords in text
func countKeywords(text string) (int, int) {
	textLower := strings.ToLower(text)

	bullKeywords := []string{"buy", "bull", "moon", "pump", "long", "call", "up", "gain", "strong", "breakout", "bullish", "rally", "surge", "upside", "accumulate", "undervalued", "support", "bottom", "recovery"}
	bearKeywords := []string{"sell", "bear", "dump", "short", "put", "down", "loss", "weak", "crash", "drop", "bearish", "tank", "decline", "downside", "overvalued", "resistance", "top", "overbought", "bubble"}

	bullCount := 0
	for _, kw := range bullKeywords {
		bullCount += strings.Count(textLower, kw)
	}

	bearCount := 0
	for _, kw := range bearKeywords {
		bearCount += strings.Count(textLower, kw)
	}

	return bullCount, bearCount
}

// calculateSentimentScore calculates sentiment score from keyword counts
func calculateSentimentScore(bullCount, bearCount int) float64 {
	total := bullCount + bearCount
	if total == 0 {
		return 0.0
	}
	return float64(bullCount-bearCount) / float64(total)
}

// determineSentimentLabel returns a human-readable sentiment label
func determineSentimentLabel(score float64) string {
	if score > 0.2 {
		return "Strongly Bullish"
	}
	if score > 0.05 {
		return "Bullish"
	}
	if score < -0.2 {
		return "Strongly Bearish"
	}
	if score < -0.05 {
		return "Bearish"
	}
	return "Neutral"
}
