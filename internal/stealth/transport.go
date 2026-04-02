package stealth

import (
	"crypto/tls"
	"fmt"
	"net/http"
	"net/url"
	"time"
)

// NewTransport creates an *http.Transport with a TLS configuration
// that mimics a real Chrome browser's TLS fingerprint as closely as
// possible using Go's standard crypto/tls library.
//
// What this changes vs Go's default:
//   - Forces TLS 1.2+ (Chrome doesn't do TLS 1.0/1.1)
//   - Sets cipher suite order to match Chrome's preferences
//   - Sets curve preferences to match Chrome (X25519 first)
//   - Enables HTTP/2 (Chrome always tries h2)
//   - Configures realistic timeouts
//
// Limitations:
//   - Go's TLS stack has a distinctive JA3 fingerprint that differs from
//     Chrome's BoringSSL. For full JA3 spoofing, use uTLS:
//     github.com/refraction-networking/utls
//
// proxyURL can be empty (direct), "http://host:port", or "socks5://host:port".
func NewTransport(proxyURL string) (*http.Transport, error) {
	tlsConfig := &tls.Config{
		MinVersion: tls.VersionTLS12,
		MaxVersion: tls.VersionTLS13,
		// Chrome's preferred cipher suite order
		CipherSuites: []uint16{
			tls.TLS_AES_128_GCM_SHA256,
			tls.TLS_AES_256_GCM_SHA384,
			tls.TLS_CHACHA20_POLY1305_SHA256,
			tls.TLS_ECDHE_ECDSA_WITH_AES_128_GCM_SHA256,
			tls.TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256,
			tls.TLS_ECDHE_ECDSA_WITH_AES_256_GCM_SHA384,
			tls.TLS_ECDHE_RSA_WITH_AES_256_GCM_SHA384,
			tls.TLS_ECDHE_ECDSA_WITH_CHACHA20_POLY1305_SHA256,
			tls.TLS_ECDHE_RSA_WITH_CHACHA20_POLY1305_SHA256,
		},
		CurvePreferences: []tls.CurveID{
			tls.X25519,
			tls.CurveP256,
			tls.CurveP384,
		},
		InsecureSkipVerify: false,
	}

	transport := &http.Transport{
		TLSClientConfig:       tlsConfig,
		ForceAttemptHTTP2:     true,
		MaxIdleConns:          100,
		MaxIdleConnsPerHost:   10,
		IdleConnTimeout:       90 * time.Second,
		TLSHandshakeTimeout:   10 * time.Second,
		ResponseHeaderTimeout: 30 * time.Second,
		DisableCompression:    false, // accept gzip/br like Chrome
	}

	// Apply proxy if specified
	if proxyURL != "" {
		parsed, err := url.Parse(proxyURL)
		if err != nil {
			return nil, fmt.Errorf("invalid proxy URL %q: %w", proxyURL, err)
		}
		transport.Proxy = http.ProxyURL(parsed)
	}

	return transport, nil
}
