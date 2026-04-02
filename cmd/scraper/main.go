// Package main is the entry point for the scraper CLI application.
package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"time"

	"colly-chromedp-scraper/internal/config"
	"colly-chromedp-scraper/internal/fetcher"
	"colly-chromedp-scraper/internal/output"
	"colly-chromedp-scraper/internal/retry"

	"github.com/spf13/cobra"
)

var cfg = config.Default()

func main() {
	rootCmd := &cobra.Command{
		Use:   "scraper [url]",
		Short: "Web scraper powered by Colly and Chromedp",
		Long: `A dual-engine web scraper that fetches HTML content and screenshots.

Engines:
  colly    - Fast HTTP-based scraping (no JavaScript)
  chrome   - Full browser rendering via headless Chrome
  both     - Run both engines and save results side-by-side

Examples:
  scraper https://example.com
  scraper example.com --engine colly --timeout 30s
  scraper https://example.com --engine chrome --headless=false
  scraper https://example.com --format json --retries 5
  scraper https://example.com --stealth --proxy socks5://127.0.0.1:1080`,
		Args:    cobra.ExactArgs(1),
		RunE:    runScraper,
		Version: "3.0.0",
	}

	// CLI flags
	rootCmd.Flags().StringVarP(&cfg.Engine, "engine", "e", cfg.Engine, "Scraping engine: colly, chrome, both")
	rootCmd.Flags().DurationVarP(&cfg.Timeout, "timeout", "t", cfg.Timeout, "Request timeout per engine")
	rootCmd.Flags().StringVarP(&cfg.OutputDir, "output", "o", cfg.OutputDir, "Output directory")
	rootCmd.Flags().BoolVar(&cfg.Headless, "headless", cfg.Headless, "Run browser in headless mode")
	rootCmd.Flags().IntVarP(&cfg.Retries, "retries", "r", cfg.Retries, "Number of retry attempts")
	rootCmd.Flags().StringVarP(&cfg.Format, "format", "f", cfg.Format, "Output format: html, json, all")
	rootCmd.Flags().StringVarP(&cfg.ProxyURL, "proxy", "p", cfg.ProxyURL, "Proxy URL (http://host:port or socks5://host:port)")
	rootCmd.Flags().BoolVar(&cfg.Stealth, "stealth", cfg.Stealth, "Enable stealth mode (anti-bot bypass)")

	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}

func runScraper(cmd *cobra.Command, args []string) error {
	cfg.URL = config.NormalizeURL(args[0])

	// Setup structured logging
	handler := slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	})
	slog.SetDefault(slog.New(handler))

	slog.Info("Starting scraper",
		"url", cfg.URL,
		"engine", cfg.Engine,
		"headless", cfg.Headless,
		"stealth", cfg.Stealth,
		"proxy", cfg.ProxyURL,
		"timeout", cfg.Timeout,
		"retries", cfg.Retries,
		"format", cfg.Format,
	)

	// Master context with generous timeout (beyond individual engine timeouts)
	ctx, cancel := context.WithTimeout(context.Background(), cfg.Timeout*2)
	defer cancel()

	// Select engines
	fetchers, err := buildFetchers(cfg.Engine)
	if err != nil {
		return err
	}

	// Prepare output directory
	writer := output.NewWriter(cfg.OutputDir)
	paths, err := writer.Prepare(cfg.URL)
	if err != nil {
		return fmt.Errorf("prepare output: %w", err)
	}

	results := make([]*fetcher.Result, 0, len(fetchers))
	anySuccess := false

	for _, f := range fetchers {
		slog.Info("Fetching with engine", "engine", f.Name())

		var result *fetcher.Result
		retryErr := retry.Do(ctx, cfg.Retries, 2*time.Second, func() error {
			var fetchErr error
			result, fetchErr = f.Fetch(ctx, cfg.URL, cfg)
			return fetchErr
		})

		if retryErr != nil {
			slog.Error("Fetch failed after retries",
				"engine", f.Name(),
				"error", retryErr,
			)
			continue
		}

		results = append(results, result)
		anySuccess = true

		// Save outputs per engine
		if err := saveEngineResults(writer, paths, f.Name(), result); err != nil {
			slog.Error("Failed to save results", "engine", f.Name(), "error", err)
		}
	}

	// Save JSON results if requested
	if cfg.Format == "json" || cfg.Format == "all" {
		if err := saveJSONResults(writer, paths, results); err != nil {
			slog.Error("Save JSON failed", "error", err)
		}
	}

	// Cleanup on total failure
	if !anySuccess {
		slog.Error("All engines failed, cleaning up output directory")
		_ = writer.Cleanup(paths.Dir)
		return fmt.Errorf("all scraping engines failed for %s", cfg.URL)
	}

	slog.Info("Scraping complete",
		"url", cfg.URL,
		"engines_succeeded", len(results),
		"output_dir", paths.Dir,
	)

	return nil
}

// buildFetchers returns the fetcher(s) based on the selected engine.
func buildFetchers(engine string) ([]fetcher.Fetcher, error) {
	switch engine {
	case "colly":
		return []fetcher.Fetcher{fetcher.NewCollyFetcher()}, nil
	case "chrome":
		return []fetcher.Fetcher{fetcher.NewChromeFetcher()}, nil
	case "both":
		return []fetcher.Fetcher{
			fetcher.NewCollyFetcher(),
			fetcher.NewChromeFetcher(),
		}, nil
	default:
		return nil, fmt.Errorf("unknown engine: %s (choose: colly, chrome, both)", engine)
	}
}

// saveEngineResults persists HTML and screenshot for a given engine.
func saveEngineResults(writer *output.Writer, paths *output.OutputPaths, engine string, result *fetcher.Result) error {
	switch engine {
	case "colly":
		if err := writer.SaveFile(paths.CollyHTMLPath, result.HTML); err != nil {
			return err
		}
		slog.Info("HTML saved", "engine", engine, "path", paths.CollyHTMLPath)

	case "chromedp":
		if err := writer.SaveFile(paths.ChromeHTMLPath, result.HTML); err != nil {
			return err
		}
		slog.Info("HTML saved", "engine", engine, "path", paths.ChromeHTMLPath)

		if err := writer.SaveFile(paths.ScreenshotPath, result.Screenshot); err != nil {
			return err
		}
		slog.Info("Screenshot saved", "path", paths.ScreenshotPath)
	}
	return nil
}

// saveJSONResults writes a summary JSON with metadata (no HTML body).
func saveJSONResults(writer *output.Writer, paths *output.OutputPaths, results []*fetcher.Result) error {
	type jsonEntry struct {
		URL           string            `json:"url"`
		Engine        string            `json:"engine"`
		StatusCode    int               `json:"status_code"`
		Headers       map[string]string `json:"headers"`
		HTMLSizeBytes int               `json:"html_size_bytes"`
		HasScreenshot bool              `json:"has_screenshot"`
		FetchedAt     time.Time         `json:"fetched_at"`
		Duration      string            `json:"duration"`
	}

	entries := make([]jsonEntry, 0, len(results))
	for _, r := range results {
		entries = append(entries, jsonEntry{
			URL:           r.URL,
			Engine:        r.Engine,
			StatusCode:    r.StatusCode,
			Headers:       r.Headers,
			HTMLSizeBytes: len(r.HTML),
			HasScreenshot: len(r.Screenshot) > 0,
			FetchedAt:     r.FetchedAt,
			Duration:      r.Duration.String(),
		})
	}

	if err := writer.SaveJSON(paths.JSONPath, entries); err != nil {
		return err
	}
	slog.Info("JSON results saved", "path", paths.JSONPath)
	return nil
}
