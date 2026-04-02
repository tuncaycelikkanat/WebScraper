package config

import "testing"

func TestNormalizeURL(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{"bare domain", "example.com", "https://example.com"},
		{"with https", "https://example.com", "https://example.com"},
		{"with http", "http://example.com", "http://example.com"},
		{"with www", "www.example.com", "https://www.example.com"},
		{"with path", "example.com/path", "https://example.com/path"},
		{"with port", "localhost:8080", "https://localhost:8080"},
		{"empty string", "", "https://"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := NormalizeURL(tt.input)
			if got != tt.expected {
				t.Errorf("NormalizeURL(%q) = %q, want %q", tt.input, got, tt.expected)
			}
		})
	}
}

func TestDefault(t *testing.T) {
	cfg := Default()

	if cfg.Engine != "both" {
		t.Errorf("expected engine 'both', got %q", cfg.Engine)
	}
	if !cfg.Headless {
		t.Error("expected headless to be true by default")
	}
	if cfg.Retries != 3 {
		t.Errorf("expected retries 3, got %d", cfg.Retries)
	}
	if cfg.OutputDir != "outputs" {
		t.Errorf("expected outputDir 'outputs', got %q", cfg.OutputDir)
	}
	if cfg.Format != "html" {
		t.Errorf("expected format 'html', got %q", cfg.Format)
	}
	if cfg.UserAgent != "" {
		t.Error("expected empty UserAgent (random rotation)")
	}
	if !cfg.Stealth {
		t.Error("expected stealth to be true by default")
	}
	if cfg.ProxyURL != "" {
		t.Error("expected empty ProxyURL by default")
	}
}
