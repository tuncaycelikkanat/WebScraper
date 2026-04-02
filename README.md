# 🕷️ Web Scraper — Go

A dual-engine web scraper built in Go that fetches HTML content and screenshots using [Colly](https://github.com/gocolly/colly) (HTTP-based) and [Chromedp](https://github.com/chromedp/chromedp) (headless Chrome).

## Features

- **Dual Engine** — Colly for speed, Chromedp for JavaScript-rendered pages
- **Retry with Backoff** — Automatic retries with linear backoff on failure
- **Structured Logging** — `log/slog` based logging throughout
- **JSON Export** — Machine-readable metadata output alongside HTML
- **Screenshots** — Full-page PNG screenshots via Chromedp
- **CLI Flags** — Flexible configuration via command-line arguments
- **Docker Support** — Run anywhere with built-in Chrome

## Quick Start

```bash
# Build
make build

# Scrape a URL (both engines)
make scrape URL=https://example.com

# Or run directly
go run ./cmd/scraper https://example.com
```

## Usage

```
scraper [url] [flags]

Flags:
  -e, --engine string    Scraping engine: colly, chrome, both (default "both")
  -f, --format string    Output format: html, json, all (default "html")
      --headless         Run browser in headless mode (default true)
  -o, --output string    Output directory (default "outputs")
  -r, --retries int      Number of retry attempts (default 3)
  -t, --timeout duration Request timeout per engine (default 2m0s)
  -v, --version          Version info
  -h, --help             Help
```

### Examples

```bash
# Colly only (fast, no JS)
scraper https://example.com --engine colly

# Chrome only with visible browser (debugging)
scraper https://example.com --engine chrome --headless=false

# Both engines, JSON output, 5 retries
scraper https://example.com --format all --retries 5

# Custom timeout
scraper https://example.com --timeout 30s --engine colly
```

## Output Structure

```
outputs/
└── 2026-04-02_09-58-00_example_com/
    ├── 2026-04-02_09-58-00_example_com_colly.html
    ├── 2026-04-02_09-58-00_example_com_chromedp.html
    ├── 2026-04-02_09-58-00_example_com.png
    └── 2026-04-02_09-58-00_example_com_results.json  (if --format json|all)
```

## Project Structure

```
scraper_go/
├── cmd/
│   └── scraper/
│       └── main.go              # CLI entry point (Cobra)
├── internal/
│   ├── config/
│   │   ├── config.go            # Configuration & URL normalization
│   │   └── config_test.go
│   ├── fetcher/
│   │   ├── fetcher.go           # Fetcher interface & Result type
│   │   ├── colly.go             # Colly engine implementation
│   │   └── chrome.go            # Chromedp engine implementation
│   ├── output/
│   │   ├── writer.go            # File/directory management
│   │   └── writer_test.go
│   └── retry/
│       ├── retry.go             # Retry with backoff
│       └── retry_test.go
├── Makefile
├── Dockerfile
├── .golangci.yml
├── go.mod
└── go.sum
```

## Development

```bash
# Run tests
make test

# Format code
make fmt

# Lint (requires golangci-lint)
make lint

# Tidy dependencies
make tidy
```

## Docker

```bash
# Build image
docker build -t scraper .

# Run
docker run --rm -v $(pwd)/outputs:/data/outputs scraper https://example.com -o /data/outputs
```

## License

MIT
