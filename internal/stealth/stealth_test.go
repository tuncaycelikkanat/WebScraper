package stealth

import (
	"net/http"
	"testing"
)

func TestRandomUserAgent(t *testing.T) {
	seen := make(map[string]bool)
	for range 100 {
		ua := RandomUserAgent()
		if ua == "" {
			t.Fatal("RandomUserAgent returned empty string")
		}
		seen[ua] = true
	}
	// With 10 UAs and 100 draws, we should see at least 3 different ones
	if len(seen) < 3 {
		t.Errorf("expected variety in UAs, got only %d unique", len(seen))
	}
}

func TestRandomViewport(t *testing.T) {
	vp := RandomViewport()
	if vp.Width == 0 || vp.Height == 0 {
		t.Errorf("invalid viewport: %dx%d", vp.Width, vp.Height)
	}
	if vp.Width < vp.Height {
		t.Errorf("desktop viewport should be landscape: %dx%d", vp.Width, vp.Height)
	}
}

func TestRandomReferrer(t *testing.T) {
	// Just ensure it doesn't panic
	for range 50 {
		_ = RandomReferrer()
	}
}

func TestApplyHeaders(t *testing.T) {
	h := make(http.Header)
	ua := "TestAgent/1.0"
	ApplyHeaders(h, ua)

	// Check critical headers are set
	requiredHeaders := []string{
		"Accept",
		"Accept-Language",
		"Sec-Ch-Ua",
		"Sec-Ch-Ua-Mobile",
		"Sec-Ch-Ua-Platform",
		"Sec-Fetch-Dest",
		"Sec-Fetch-Mode",
		"Sec-Fetch-Site",
		"Sec-Fetch-User",
		"User-Agent",
		"Upgrade-Insecure-Requests",
	}

	for _, key := range requiredHeaders {
		if h.Get(key) == "" {
			t.Errorf("missing header: %s", key)
		}
	}

	if h.Get("User-Agent") != ua {
		t.Errorf("User-Agent mismatch: got %q, want %q", h.Get("User-Agent"), ua)
	}
}

func TestChromeStealthScripts(t *testing.T) {
	scripts := ChromeStealthScripts()
	if len(scripts) == 0 {
		t.Fatal("expected stealth scripts, got none")
	}
	// Currently 12 scripts
	if len(scripts) < 10 {
		t.Errorf("expected at least 10 stealth scripts, got %d", len(scripts))
	}
	for i, s := range scripts {
		if s == "" {
			t.Errorf("script %d is empty", i)
		}
	}
}

func TestHumanBehaviorScript(t *testing.T) {
	script := HumanBehaviorScript()
	if script == "" {
		t.Fatal("HumanBehaviorScript returned empty string")
	}
}

func TestNewTransport(t *testing.T) {
	// Without proxy
	transport, err := NewTransport("")
	if err != nil {
		t.Fatalf("NewTransport failed: %v", err)
	}
	if transport == nil {
		t.Fatal("transport is nil")
	}
	if transport.TLSClientConfig == nil {
		t.Fatal("TLS config is nil")
	}
	if !transport.ForceAttemptHTTP2 {
		t.Error("HTTP/2 should be enabled")
	}

	// With HTTP proxy
	transport, err = NewTransport("http://proxy.example.com:8080")
	if err != nil {
		t.Fatalf("NewTransport with proxy failed: %v", err)
	}
	if transport.Proxy == nil {
		t.Error("proxy should be set")
	}

	// With SOCKS5 proxy
	transport, err = NewTransport("socks5://proxy.example.com:1080")
	if err != nil {
		t.Fatalf("NewTransport with SOCKS5 failed: %v", err)
	}
	if transport.Proxy == nil {
		t.Error("SOCKS5 proxy should be set")
	}

	// Invalid proxy URL
	_, err = NewTransport("://invalid")
	if err == nil {
		t.Error("expected error for invalid proxy URL")
	}
}
