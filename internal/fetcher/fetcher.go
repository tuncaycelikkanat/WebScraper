// Package fetcher defines the interface and result types for web scraping engines.
package fetcher

import (
	"context"
	"time"

	"colly-chromedp-scraper/internal/config"
)

// Result represents the output of a fetch operation.
type Result struct {
	HTML       []byte            `json:"html,omitempty"`
	Screenshot []byte            `json:"-"`
	StatusCode int               `json:"status_code"`
	Headers    map[string]string `json:"headers,omitempty"`
	URL        string            `json:"url"`
	Engine     string            `json:"engine"`
	FetchedAt  time.Time         `json:"fetched_at"`
	Duration   time.Duration     `json:"duration"`
}

// Fetcher is the common interface implemented by all scraping engines.
// This allows swapping engines (Colly, Chromedp, Playwright, etc.)
// without changing the calling code.
type Fetcher interface {
	// Name returns the human-readable name of the engine.
	Name() string
	// Fetch retrieves the content at the given URL using the engine.
	Fetch(ctx context.Context, url string, cfg *config.Config) (*Result, error)
}
