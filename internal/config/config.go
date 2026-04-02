// Package config provides configuration management for the scraper.
package config

import (
	"strings"
	"time"
)

// Config holds all scraper configuration options.
type Config struct {
	URL       string        // Target URL to scrape
	Engine    string        // Scraping engine: "colly", "chrome", "both"
	Timeout   time.Duration // Request timeout per engine
	OutputDir string        // Base output directory
	Headless  bool          // Run browser in headless mode
	Retries   int           // Number of retry attempts on failure
	Format    string        // Output format: "html", "json", "all"
	UserAgent string        // HTTP User-Agent header (empty = random rotation)
	ProxyURL  string        // Proxy URL: "http://host:port" or "socks5://host:port"
	Stealth   bool          // Enable stealth mode (anti-bot bypass)
}

// Default returns a Config with sensible default values.
func Default() *Config {
	return &Config{
		Engine:    "both",
		Timeout:   120 * time.Second,
		OutputDir: "outputs",
		Headless:  true,
		Retries:   3,
		Format:    "html",
		UserAgent: "", // empty = use random UA rotation in stealth mode
		Stealth:   true,
	}
}

// NormalizeURL ensures the URL has a scheme prefix (defaults to https).
func NormalizeURL(rawURL string) string {
	if !strings.HasPrefix(rawURL, "http://") &&
		!strings.HasPrefix(rawURL, "https://") {
		return "https://" + rawURL
	}
	return rawURL
}
