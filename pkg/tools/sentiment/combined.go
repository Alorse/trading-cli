package sentiment

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/alorse/trading-cli/internal/config"
	"github.com/alorse/trading-cli/pkg/client"
	"github.com/alorse/trading-cli/pkg/tools/analysis"
	"github.com/alorse/trading-cli/pkg/tools/screener"
	"github.com/alorse/trading-cli/pkg/utils"
)

// ConfluenceResult holds the confluence analysis between different signals
type ConfluenceResult struct {
	TechBullish    bool   `json:"techBullish"`
	SentBullish    bool   `json:"sentBullish"`
	SignalsAgree   bool   `json:"signalsAgree"`
	Confidence     string `json:"confidence"`
	Recommendation string `json:"recommendation"`
}

// CombinedAnalysisOutput represents the merged analysis output
type CombinedAnalysisOutput struct {
	Symbol     string           `json:"symbol"`
	Exchange   string           `json:"exchange"`
	Timeframe  string           `json:"timeframe"`
	Technical  json.RawMessage  `json:"technical"`
	Sentiment  json.RawMessage  `json:"sentiment"`
	News       json.RawMessage  `json:"news"`
	Confluence ConfluenceResult `json:"confluence"`
	Timestamp  time.Time        `json:"timestamp"`
}

// RunCombinedAnalysis performs integrated analysis combining technical, sentiment, and news
func RunCombinedAnalysis(cfg *config.Config, symbol, exchange, timeframe, category string) error {
	// Validate inputs
	if symbol == "" {
		return fmt.Errorf("symbol cannot be empty")
	}
	if exchange == "" {
		return fmt.Errorf("exchange cannot be empty")
	}
	if timeframe == "" {
		return fmt.Errorf("timeframe cannot be empty")
	}
	if category == "" {
		category = "all"
	}

	ctx, cancel := context.WithTimeout(context.Background(), cfg.HTTPTimeout)
	defer cancel()

	// 1. Get technical analysis
	technicalData, err := getTechnicalAnalysis(ctx, cfg, symbol, exchange, timeframe)
	if err != nil {
		return fmt.Errorf("technical analysis: %w", err)
	}

	// 2. Get sentiment analysis
	sentimentData, sentimentScore, err := getSentimentAnalysis(ctx, cfg, symbol, category)
	if err != nil {
		return fmt.Errorf("sentiment analysis: %w", err)
	}

	// 3. Get news
	newsData, err := getNewsAnalysis(ctx, cfg, symbol, category)
	if err != nil {
		return fmt.Errorf("news analysis: %w", err)
	}

	// 4. Compute confluence
	confluence := computeConfluence(technicalData, sentimentScore)

	// 5. Build output
	output := CombinedAnalysisOutput{
		Symbol:     symbol,
		Exchange:   exchange,
		Timeframe:  timeframe,
		Technical:  technicalData,
		Sentiment:  sentimentData,
		News:       newsData,
		Confluence: confluence,
		Timestamp:  time.Now().UTC(),
	}

	return utils.PrintJSON(output)
}

// getTechnicalAnalysis fetches technical analysis data
func getTechnicalAnalysis(ctx context.Context, cfg *config.Config, symbol, exchange, timeframe string) (json.RawMessage, error) {
	ticker := screener.FormatTicker(exchange, symbol)
	screenName, err := client.ScreenerForExchange(exchange)
	if err != nil {
		return nil, fmt.Errorf("invalid exchange: %w", err)
	}

	httpClient := client.NewHTTPClient(cfg)
	tvClient := client.NewTradingViewClient(httpClient)

	results, err := tvClient.GetMultipleAnalysis(ctx, screenName, []string{ticker}, client.DefaultColumns)
	if err != nil {
		return nil, err
	}

	if len(results) == 0 {
		return nil, fmt.Errorf("no data for symbol")
	}

	output := analysis.BuildCoinAnalysisOutput(ticker, exchange, timeframe, results[0].Values)
	return json.Marshal(output)
}

// getSentimentAnalysis fetches market sentiment data
func getSentimentAnalysis(ctx context.Context, cfg *config.Config, symbol, category string) (json.RawMessage, float64, error) {
	httpClient := client.NewHTTPClient(cfg)
	redditClient := client.NewRedditClient(httpClient)

	subreddits := selectSubreddits(category)
	if len(subreddits) == 0 {
		return nil, 0, fmt.Errorf("invalid category: %s", category)
	}

	var allPosts []client.RedditPost
	postsPerSubreddit := 50 / len(subreddits)
	if postsPerSubreddit <= 0 {
		postsPerSubreddit = 1
	}

	for _, subreddit := range subreddits {
		posts, err := redditClient.Search(ctx, subreddit, symbol, postsPerSubreddit)
		if err != nil {
			continue
		}
		allPosts = append(allPosts, posts...)
	}

	if len(allPosts) == 0 {
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
		data, _ := json.Marshal(output)
		return json.RawMessage(data), 0.0, nil
	}

	bullishCount := 0
	bearishCount := 0
	neutralCount := 0
	totalScore := 0.0

	for _, post := range allPosts {
		text := post.Title + " " + post.Selftext
		bullCount, bearCount := countKeywords(text)
		postScore := calculateSentimentScore(bullCount, bearCount)

		if postScore > 0.2 {
			bullishCount++
		} else if postScore < -0.2 {
			bearishCount++
		} else {
			neutralCount++
		}

		totalScore += postScore
	}

	aggregateScore := totalScore / float64(len(allPosts))
	label := determineSentimentLabel(aggregateScore)

	output := SentimentOutput{
		Symbol:         symbol,
		SentimentScore: aggregateScore,
		SentimentLabel: label,
		PostsAnalyzed:  len(allPosts),
		BullishCount:   bullishCount,
		BearishCount:   bearishCount,
		NeutralCount:   neutralCount,
		TopPosts:       []RedditTopPost{},
		Sources:        []string{},
		Timestamp:      time.Now().UTC(),
	}

	data, _ := json.Marshal(output)
	return json.RawMessage(data), aggregateScore, nil
}

// getNewsAnalysis fetches news data
func getNewsAnalysis(ctx context.Context, cfg *config.Config, symbol, category string) (json.RawMessage, error) {
	httpClient := client.NewHTTPClient(cfg)
	rssClient := client.NewRSSClient(httpClient)

	feeds := selectFeeds(category)
	if len(feeds) == 0 {
		return nil, fmt.Errorf("invalid category: %s", category)
	}

	items, err := rssClient.FetchMultiple(ctx, feeds)
	if err != nil {
		items = []client.RSSItem{}
	}

	if symbol != "" {
		items = filterBySymbol(items, symbol)
	}

	newsItems := make([]NewsItem, 0, len(items))
	for i, item := range items {
		if i >= 10 {
			break
		}
		newsItems = append(newsItems, NewsItem{
			Title:     item.Title,
			URL:       item.Link,
			Published: item.PubDate,
			Summary:   stripHTML(item.Description),
			Source:    item.Source,
		})
	}

	output := NewsOutput{
		Symbol:    symbol,
		Category:  category,
		Count:     len(newsItems),
		Items:     newsItems,
		Timestamp: time.Now().UTC(),
	}

	data, _ := json.Marshal(output)
	return json.RawMessage(data), nil
}

// computeConfluence computes confluence between technical and sentiment signals
func computeConfluence(technicalData json.RawMessage, sentimentScore float64) ConfluenceResult {
	// Parse technical data to check trend
	var techAnalysis analysis.CoinAnalysisOutput
	_ = json.Unmarshal(technicalData, &techAnalysis)

	techBullish := techAnalysis.MarketStructure.Trend == "bullish"
	sentBullish := sentimentScore > 0.1

	signalsAgree := (techBullish && sentBullish) || (!techBullish && !sentBullish)

	confidence := "MIXED"
	if signalsAgree {
		confidence = "HIGH"
	}

	recommendation := buildRecommendation(techBullish, sentBullish, signalsAgree)

	return ConfluenceResult{
		TechBullish:    techBullish,
		SentBullish:    sentBullish,
		SignalsAgree:   signalsAgree,
		Confidence:     confidence,
		Recommendation: recommendation,
	}
}

// buildRecommendation constructs natural language recommendation
func buildRecommendation(techBullish, sentBullish, agree bool) string {
	if agree {
		if techBullish {
			return "Strong bullish confluence: both technical and sentiment indicators align positively"
		}
		return "Strong bearish confluence: both technical and sentiment indicators align negatively"
	}

	if techBullish && !sentBullish {
		return "Caution: Technical indicators are bullish but sentiment is bearish - monitor for divergence"
	}
	if !techBullish && sentBullish {
		return "Mixed signals: Sentiment is bullish but technical indicators are bearish - wait for confirmation"
	}

	return "No clear direction - awaiting further confluence"
}
