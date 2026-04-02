package fetcher

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"time"

	"colly-chromedp-scraper/internal/config"
	"colly-chromedp-scraper/internal/stealth"

	"github.com/gocolly/colly/v2"
)

// CollyFetcher implements Fetcher using the Colly HTTP scraping library.
type CollyFetcher struct{}

// NewCollyFetcher creates a new CollyFetcher instance.
func NewCollyFetcher() *CollyFetcher {
	return &CollyFetcher{}
}

// Name returns the engine identifier.
func (f *CollyFetcher) Name() string {
	return "colly"
}

// Fetch retrieves the HTML content of the target URL using Colly.
// When stealth mode is enabled, it applies:
//   - Random User-Agent rotation
//   - Realistic browser headers (Sec-CH-UA, Sec-Fetch-*, etc.)
//   - Chrome-like TLS fingerprint
//   - Proxy support
//   - Random referrer
func (f *CollyFetcher) Fetch(ctx context.Context, targetURL string, cfg *config.Config) (*Result, error) {
	start := time.Now()

	// Determine User-Agent
	ua := cfg.UserAgent
	if ua == "" && cfg.Stealth {
		ua = stealth.RandomUserAgent()
	} else if ua == "" {
		ua = "Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/131.0.0.0 Safari/537.36"
	}

	c := colly.NewCollector(
		colly.UserAgent(ua),
		colly.AllowURLRevisit(),
	)

	c.SetRequestTimeout(cfg.Timeout)

	// Apply stealth transport (TLS fingerprint + proxy)
	if cfg.Stealth || cfg.ProxyURL != "" {
		transport, err := stealth.NewTransport(cfg.ProxyURL)
		if err != nil {
			return nil, fmt.Errorf("create stealth transport: %w", err)
		}
		c.WithTransport(transport)
		slog.Info("Colly stealth transport applied",
			"proxy", cfg.ProxyURL != "",
			"ua", ua[:40]+"...",
		)
	}

	if err := c.Limit(&colly.LimitRule{
		DomainGlob:  "*",
		Delay:       2 * time.Second,
		RandomDelay: 1 * time.Second,
	}); err != nil {
		return nil, fmt.Errorf("set limit rule: %w", err)
	}

	result := &Result{
		URL:     targetURL,
		Engine:  "colly",
		Headers: make(map[string]string),
	}

	var fetchErr error

	c.OnRequest(func(r *colly.Request) {
		if cfg.Stealth {
			// Apply full set of realistic browser headers
			stealth.ApplyCollyHeaders(r.Headers.Set, ua)
		} else {
			r.Headers.Set("Accept-Language", "tr-TR,tr;q=0.9,en-US;q=0.8")
			r.Headers.Set("Referer", "https://www.google.com/")
		}
		slog.Info("Colly request", "url", r.URL, "stealth", cfg.Stealth)
	})

	c.OnResponse(func(r *colly.Response) {
		result.HTML = r.Body
		result.StatusCode = r.StatusCode

		// Capture response headers
		for _, key := range []string{"Server", "Content-Type", "Content-Length", "X-Powered-By", "Set-Cookie"} {
			if v := r.Headers.Get(key); v != "" {
				result.Headers[key] = v
			}
		}

		slog.Info("Colly response",
			"status", r.StatusCode,
			"size", len(r.Body),
			"server", r.Headers.Get("Server"),
		)
	})

	c.OnError(func(r *colly.Response, err error) {
		if r != nil {
			result.StatusCode = r.StatusCode
			slog.Error("Colly error",
				"status", r.StatusCode,
				"error", err,
			)
		} else {
			slog.Error("Colly error", "error", err)
		}
		fetchErr = err
	})

	if err := c.Visit(targetURL); err != nil {
		return nil, fmt.Errorf("colly visit failed: %w", err)
	}

	if fetchErr != nil {
		return nil, fmt.Errorf("colly fetch error: %w", fetchErr)
	}

	result.Duration = time.Since(start)
	result.FetchedAt = time.Now()

	return result, nil
}

// collyTransportAdapter wraps *http.Transport to implement http.RoundTripper
// interface for Colly's WithTransport method.
type collyTransportAdapter struct {
	transport *http.Transport
}

func (a *collyTransportAdapter) RoundTrip(req *http.Request) (*http.Response, error) {
	return a.transport.RoundTrip(req)
}
