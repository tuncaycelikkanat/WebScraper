// Package retry provides a generic retry mechanism with exponential backoff
// and context cancellation support.
package retry

import (
	"context"
	"fmt"
	"log/slog"
	"time"
)

// Do executes fn up to `attempts` times. If fn returns nil, Do returns nil
// immediately. Between retries, it waits with linear backoff (baseDelay * attempt).
// If the context is cancelled during a wait, Do returns the context error.
func Do(ctx context.Context, attempts int, baseDelay time.Duration, fn func() error) error {
	var lastErr error
	for i := range attempts {
		if i > 0 {
			delay := baseDelay * time.Duration(i)
			slog.Info("Retrying",
				"attempt", i+1,
				"max", attempts,
				"delay", delay,
			)
			select {
			case <-ctx.Done():
				return fmt.Errorf("retry cancelled: %w", ctx.Err())
			case <-time.After(delay):
			}
		}

		if err := fn(); err != nil {
			lastErr = err
			slog.Warn("Attempt failed",
				"attempt", i+1,
				"error", err,
			)
			continue
		}
		return nil
	}
	return fmt.Errorf("all %d attempts failed, last error: %w", attempts, lastErr)
}
