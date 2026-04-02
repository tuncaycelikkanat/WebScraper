package fetcher

import (
	"context"
	"fmt"
	"log/slog"
	"math/rand/v2"
	"time"

	"colly-chromedp-scraper/internal/config"
	"colly-chromedp-scraper/internal/stealth"

	"github.com/chromedp/cdproto/page"
	"github.com/chromedp/chromedp"
)

// ChromeFetcher implements Fetcher using Chromedp (headless Chrome).
type ChromeFetcher struct{}

// NewChromeFetcher creates a new ChromeFetcher instance.
func NewChromeFetcher() *ChromeFetcher {
	return &ChromeFetcher{}
}

// Name returns the engine identifier.
func (f *ChromeFetcher) Name() string {
	return "chromedp"
}

// Fetch renders a page in a real browser (headless or visible), extracts the
// full DOM HTML, and takes a full-page screenshot.
//
// When stealth mode is enabled, it:
//   - Injects stealth scripts BEFORE page load (navigator.webdriver, plugins, etc.)
//   - Uses random User-Agent from the pool
//   - Randomizes viewport size
//   - Simulates human-like scrolling and mouse movements
//   - Applies random delays between actions
//   - Routes through proxy if configured
func (f *ChromeFetcher) Fetch(ctx context.Context, targetURL string, cfg *config.Config) (*Result, error) {
	start := time.Now()

	// Determine User-Agent
	ua := cfg.UserAgent
	if ua == "" && cfg.Stealth {
		ua = stealth.RandomUserAgent()
	} else if ua == "" {
		ua = "Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/131.0.0.0 Safari/537.36"
	}

	// Random viewport for stealth
	vp := stealth.Viewport{Width: 1920, Height: 1080}
	if cfg.Stealth {
		vp = stealth.RandomViewport()
	}

	opts := append(chromedp.DefaultExecAllocatorOptions[:],
		chromedp.Flag("headless", cfg.Headless),
		chromedp.Flag("disable-blink-features", "AutomationControlled"),
		chromedp.Flag("window-size", fmt.Sprintf("%d,%d", vp.Width, vp.Height)),
		chromedp.UserAgent(ua),

		// Additional anti-detection flags
		chromedp.Flag("disable-features", "AutomationControlled,TranslateUI"),
		chromedp.Flag("disable-infobars", true),
		chromedp.Flag("excludeSwitches", "enable-automation"),
		chromedp.Flag("useAutomationExtension", false),
		chromedp.Flag("disable-extensions", true),

		// Make the browser appear more realistic
		chromedp.Flag("lang", "tr-TR,tr"),
		chromedp.Flag("accept-lang", "tr-TR,tr;q=0.9,en-US;q=0.8,en;q=0.7"),
	)

	// Proxy support for Chromedp
	if cfg.ProxyURL != "" {
		opts = append(opts, chromedp.ProxyServer(cfg.ProxyURL))
		slog.Info("Chromedp using proxy", "proxy", cfg.ProxyURL)
	}

	// Move window off-screen when not headless (avoids UI flash)
	if !cfg.Headless {
		opts = append(opts, chromedp.Flag("window-position", "-32000,-32000"))
	}

	allocCtx, cancel := chromedp.NewExecAllocator(ctx, opts...)
	defer cancel()

	taskCtx, cancel := chromedp.NewContext(allocCtx)
	defer cancel()

	taskCtx, cancel = context.WithTimeout(taskCtx, cfg.Timeout)
	defer cancel()

	slog.Info("Chromedp opening browser",
		"url", targetURL,
		"headless", cfg.Headless,
		"stealth", cfg.Stealth,
		"viewport", fmt.Sprintf("%dx%d", vp.Width, vp.Height),
	)

	// Phase 1: Inject stealth scripts BEFORE navigation
	if cfg.Stealth {
		scripts := stealth.ChromeStealthScripts()
		for _, script := range scripts {
			if err := chromedp.Run(taskCtx,
				chromedp.ActionFunc(func(ctx context.Context) error {
					_, err := page.AddScriptToEvaluateOnNewDocument(script).Do(ctx)
					return err
				}),
			); err != nil {
				slog.Warn("Failed to inject stealth script", "error", err)
			}
		}
		slog.Info("Stealth scripts injected", "count", len(scripts))
	}

	var html string
	var screenshot []byte

	// Phase 2: Navigate + human-like interactions
	actions := []chromedp.Action{
		chromedp.Navigate(targetURL),
		chromedp.WaitReady("body"),
	}

	// Random initial wait (3-6 seconds)
	actions = append(actions, chromedp.Sleep(time.Duration(3+rand.IntN(4))*time.Second))

	if cfg.Stealth {
		// Inject human behavior simulation (mouse moves, random scroll)
		actions = append(actions, chromedp.Evaluate(stealth.HumanBehaviorScript(), nil))
		// Wait for smooth scroll to complete
		actions = append(actions, chromedp.Sleep(time.Duration(1+rand.IntN(2))*time.Second))
		// Second scroll interaction
		actions = append(actions, chromedp.Evaluate(`window.scrollTo({top: 0, behavior: 'smooth'});`, nil))
		actions = append(actions, chromedp.Sleep(time.Duration(1+rand.IntN(2))*time.Second))
		// Scroll back down to capture full page
		actions = append(actions, chromedp.Evaluate(`window.scrollTo({top: document.body.scrollHeight, behavior: 'smooth'});`, nil))
		actions = append(actions, chromedp.Sleep(time.Duration(2+rand.IntN(3))*time.Second))
	} else {
		actions = append(actions,
			chromedp.Evaluate(`window.scrollTo(0, document.body.scrollHeight);`, nil),
			chromedp.Sleep(time.Duration(2+rand.IntN(3))*time.Second),
		)
	}

	// Extract HTML and take screenshot
	actions = append(actions,
		chromedp.OuterHTML("html", &html),
		chromedp.FullScreenshot(&screenshot, 90),
	)

	if err := chromedp.Run(taskCtx, actions...); err != nil {
		return nil, fmt.Errorf("chromedp run failed: %w", err)
	}

	slog.Info("Chromedp fetch complete",
		"html_size", len(html),
		"screenshot_size", len(screenshot),
		"duration", time.Since(start),
	)

	return &Result{
		HTML:       []byte(html),
		Screenshot: screenshot,
		StatusCode: 200,
		Headers:    map[string]string{},
		URL:        targetURL,
		Engine:     "chromedp",
		FetchedAt:  time.Now(),
		Duration:   time.Since(start),
	}, nil
}
