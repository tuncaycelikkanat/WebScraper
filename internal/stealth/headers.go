package stealth

import "net/http"

// ApplyHeaders sets realistic browser headers on an http.Header.
// These headers match what a real Chrome browser sends, including
// Sec-CH-UA (Client Hints), Sec-Fetch-* headers, and proper ordering.
//
// Anti-bot systems like Cloudflare and DataDome check:
// 1. Header presence — missing Sec-CH-UA is a red flag
// 2. Header order — Go's net/http canonicalizes, but values matter
// 3. Header consistency — UA version must match Sec-CH-UA version
func ApplyHeaders(h http.Header, ua string) {
	h.Set("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,image/avif,image/webp,image/apng,*/*;q=0.8,application/signed-exchange;v=b3;q=0.7")
	h.Set("Accept-Language", "tr-TR,tr;q=0.9,en-US;q=0.8,en;q=0.7")
	h.Set("Accept-Encoding", "gzip, deflate, br, zstd")
	h.Set("Cache-Control", "max-age=0")
	h.Set("Connection", "keep-alive")
	h.Set("DNT", "1")
	h.Set("Upgrade-Insecure-Requests", "1")

	// Client Hints — critical for bypassing modern anti-bot
	h.Set("Sec-Ch-Ua", `"Chromium";v="131", "Google Chrome";v="131", "Not_A Brand";v="24"`)
	h.Set("Sec-Ch-Ua-Mobile", "?0")
	h.Set("Sec-Ch-Ua-Platform", `"Windows"`)

	// Sec-Fetch headers — indicate the request context
	h.Set("Sec-Fetch-Dest", "document")
	h.Set("Sec-Fetch-Mode", "navigate")
	h.Set("Sec-Fetch-Site", "none")
	h.Set("Sec-Fetch-User", "?1")

	h.Set("User-Agent", ua)
}

// ApplyCollyHeaders sets realistic headers on a Colly request.
// This mirrors what ApplyHeaders does but works with Colly's header API.
func ApplyCollyHeaders(setFunc func(key, value string), ua string) {
	setFunc("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,image/avif,image/webp,image/apng,*/*;q=0.8,application/signed-exchange;v=b3;q=0.7")
	setFunc("Accept-Language", "tr-TR,tr;q=0.9,en-US;q=0.8,en;q=0.7")
	setFunc("Accept-Encoding", "gzip, deflate, br, zstd")
	setFunc("Cache-Control", "max-age=0")
	setFunc("DNT", "1")
	setFunc("Upgrade-Insecure-Requests", "1")

	// Client Hints
	setFunc("Sec-Ch-Ua", `"Chromium";v="131", "Google Chrome";v="131", "Not_A Brand";v="24"`)
	setFunc("Sec-Ch-Ua-Mobile", "?0")
	setFunc("Sec-Ch-Ua-Platform", `"Windows"`)

	// Sec-Fetch
	setFunc("Sec-Fetch-Dest", "document")
	setFunc("Sec-Fetch-Mode", "navigate")
	setFunc("Sec-Fetch-Site", "none")
	setFunc("Sec-Fetch-User", "?1")

	// Referrer
	ref := RandomReferrer()
	if ref != "" {
		setFunc("Referer", ref)
	}
}
