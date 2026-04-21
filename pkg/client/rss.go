package client

import (
	"context"
	"encoding/xml"
	"fmt"
	"strings"
	"time"
)

type RSSItem struct {
	Title       string    `json:"title"`
	Link        string    `json:"link"`
	Description string    `json:"description"`
	PubDate     time.Time `json:"pub_date"`
	Source      string    `json:"source"`
}

var NewsFeeds = map[string]string{
	"coindesk":    "https://www.coindesk.com/arc/outboundfeeds/rss/",
	"cointelegraph": "https://cointelegraph.com/rss",
	"reuters-business": "https://feeds.reuters.com/reuters/businessNews",
	"reuters-companies": "https://feeds.reuters.com/reuters/companyNews",
}

type RSSClient struct {
	http *HTTPClient
}

func NewRSSClient(http *HTTPClient) *RSSClient {
	return &RSSClient{http: http}
}

func (c *RSSClient) FetchFeed(ctx context.Context, feedName string) ([]RSSItem, error) {
	feedURL, ok := NewsFeeds[feedName]
	if !ok {
		return nil, fmt.Errorf("unknown feed: %s", feedName)
	}
	return c.FetchURL(ctx, feedURL, feedName)
}

func (c *RSSClient) FetchURL(ctx context.Context, feedURL, source string) ([]RSSItem, error) {
	data, err := c.http.Get(ctx, feedURL)
	if err != nil {
		return nil, fmt.Errorf("fetch rss feed %s: %w", source, err)
	}
	return parseRSS(data, source)
}

// FetchMultiple aggregates items from multiple feeds. Individual feed failures
// are skipped so a single unavailable source does not block the rest.
func (c *RSSClient) FetchMultiple(ctx context.Context, feedNames []string) ([]RSSItem, error) {
	var all []RSSItem
	for _, name := range feedNames {
		items, err := c.FetchFeed(ctx, name)
		if err != nil {
			continue
		}
		all = append(all, items...)
	}
	return all, nil
}

type rssChannel struct {
	Items []rssXMLItem `xml:"item"`
}

type rssXMLItem struct {
	Title       string `xml:"title"`
	Link        string `xml:"link"`
	Description string `xml:"description"`
	PubDate     string `xml:"pubDate"`
}

type rssFeed struct {
	XMLName xml.Name   `xml:"rss"`
	Channel rssChannel `xml:"channel"`
}

func parseRSS(data []byte, source string) ([]RSSItem, error) {
	var feed rssFeed
	if err := xml.Unmarshal(data, &feed); err != nil {
		return nil, fmt.Errorf("parse rss xml: %w", err)
	}

	items := make([]RSSItem, 0, len(feed.Channel.Items))
	for _, xi := range feed.Channel.Items {
		item := RSSItem{
			Title:       strings.TrimSpace(xi.Title),
			Link:        strings.TrimSpace(xi.Link),
			Description: strings.TrimSpace(xi.Description),
			Source:      source,
		}

		if xi.PubDate != "" {
			formats := []string{
				time.RFC1123Z,
				time.RFC1123,
				"Mon, 02 Jan 2006 15:04:05 -0700",
				"Mon, 02 Jan 2006 15:04:05 MST",
			}
			for _, format := range formats {
				if t, err := time.Parse(format, strings.TrimSpace(xi.PubDate)); err == nil {
					item.PubDate = t
					break
				}
			}
		}

		items = append(items, item)
	}

	return items, nil
}
