package client

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"
)

const redditBase = "https://www.reddit.com"

type RedditPost struct {
	Title     string  `json:"title"`
	Subreddit string  `json:"subreddit"`
	Score     int     `json:"score"`
	NumComments int   `json:"num_comments"`
	URL       string  `json:"url"`
	Permalink string  `json:"permalink"`
	CreatedUTC float64 `json:"created_utc"`
	Selftext  string  `json:"selftext"`
}

type RedditClient struct {
	http *HTTPClient
}

func NewRedditClient(http *HTTPClient) *RedditClient {
	return &RedditClient{http: http}
}

func (c *RedditClient) Search(ctx context.Context, subreddit, query string, limit int) ([]RedditPost, error) {
	if limit <= 0 {
		limit = 25
	}
	if limit > 100 {
		limit = 100
	}

	params := url.Values{}
	params.Set("q", query)
	params.Set("sort", "new")
	params.Set("t", "week")
	params.Set("limit", fmt.Sprintf("%d", limit))

	reqURL := fmt.Sprintf("%s/r/%s/search.json?%s", redditBase, subreddit, params.Encode())

	headers := map[string]string{
		"User-Agent": "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36",
	}

	data, err := c.http.GetWithHeaders(ctx, reqURL, headers)
	if err != nil {
		return nil, fmt.Errorf("reddit search: %w", err)
	}

	return parseRedditResponse(data)
}

type redditListing struct {
	Data struct {
		Children []struct {
			Data RedditPost `json:"data"`
		} `json:"children"`
	} `json:"data"`
}

func parseRedditResponse(data []byte) ([]RedditPost, error) {
	var listing redditListing
	if err := json.Unmarshal(data, &listing); err != nil {
		return nil, fmt.Errorf("parse reddit response: %w", err)
	}

	posts := make([]RedditPost, 0, len(listing.Data.Children))
	for _, child := range listing.Data.Children {
		posts = append(posts, child.Data)
	}

	return posts, nil
}
