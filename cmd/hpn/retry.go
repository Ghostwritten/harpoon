package main

import (
	"context"
	"fmt"
	"time"

	containerruntime "github.com/harpoon/hpn/internal/runtime"
)

// retryWithBackoff calls fn up to cfg.MaxAttempts times.
// Between attempts it waits for an exponentially increasing delay (capped at cfg.MaxDelay).
// Returns nil on the first successful call, or a wrapped error after exhausting all attempts.
// Returns ctx.Err() immediately if the context is cancelled before retrying.
func retryWithBackoff(ctx context.Context, fn func() error, cfg containerruntime.RetryConfig) error {
	var lastErr error
	delay := cfg.Delay

	for attempt := 1; attempt <= cfg.MaxAttempts; attempt++ {
		if ctx.Err() != nil {
			return ctx.Err()
		}

		if err := fn(); err == nil {
			return nil
		} else {
			lastErr = err
		}

		if attempt < cfg.MaxAttempts {
			select {
			case <-ctx.Done():
				return ctx.Err()
			case <-time.After(delay):
			}
			delay = time.Duration(float64(delay) * 1.5)
			if delay > cfg.MaxDelay {
				delay = cfg.MaxDelay
			}
		}
	}

	return fmt.Errorf("failed after %d attempts: %w", cfg.MaxAttempts, lastErr)
}
