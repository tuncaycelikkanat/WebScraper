// Package stealth provides anti-bot bypass techniques for web scraping.
// It includes User-Agent rotation, realistic browser headers,
// Chromedp stealth scripts, and TLS fingerprint manipulation.
package stealth

import "math/rand/v2"

// Desktop Chrome User-Agents from recent stable versions.
// Each UA matches a specific OS + Chrome version combination.
var desktopUAs = []string{
	// Chrome 131 — Windows
	"Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/131.0.0.0 Safari/537.36",
	// Chrome 131 — macOS
	"Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/131.0.0.0 Safari/537.36",
	// Chrome 131 — Linux
	"Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/131.0.0.0 Safari/537.36",
	// Chrome 130 — Windows
	"Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/130.0.0.0 Safari/537.36",
	// Chrome 130 — macOS
	"Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/130.0.0.0 Safari/537.36",
	// Chrome 129 — Windows
	"Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/129.0.0.0 Safari/537.36",
	// Chrome 129 — macOS
	"Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/129.0.0.0 Safari/537.36",
	// Chrome 128 — Windows
	"Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/128.0.0.0 Safari/537.36",
	// Edge 131 — Windows (Chromium-based, same engine)
	"Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/131.0.0.0 Safari/537.36 Edg/131.0.0.0",
	// Chrome 131 — Windows 11
	"Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/131.0.6778.86 Safari/537.36",
}

// RandomUserAgent returns a random desktop Chrome User-Agent string.
func RandomUserAgent() string {
	return desktopUAs[rand.IntN(len(desktopUAs))]
}

// Viewport represents a common screen resolution.
type Viewport struct {
	Width  int
	Height int
}

// Common desktop screen resolutions (from StatCounter global stats).
var viewports = []Viewport{
	{1920, 1080},
	{1366, 768},
	{1536, 864},
	{1440, 900},
	{1280, 720},
	{2560, 1440},
	{1600, 900},
}

// RandomViewport returns a randomly selected common screen resolution.
func RandomViewport() Viewport {
	return viewports[rand.IntN(len(viewports))]
}

// Referrers commonly seen in real user traffic.
var referrers = []string{
	"https://www.google.com/",
	"https://www.google.com.tr/",
	"https://www.google.com/search?q=",
	"https://yandex.com.tr/",
	"https://search.yahoo.com/",
	"", // direct traffic (no referrer)
}

// RandomReferrer returns a realistic referrer URL.
func RandomReferrer() string {
	return referrers[rand.IntN(len(referrers))]
}
