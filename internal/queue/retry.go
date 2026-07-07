package queue

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/avast/retry-go/v4"
)

// RetryConfig holds exponential backoff + jitter settings for SSE-aware retries.
type RetryConfig struct {
	Attempts     uint
	Delay        time.Duration
	MaxDelay     time.Duration
	JitterFactor float64 // 0.0 = no jitter, 0.5 = ±50%
}

// DefaultRetryConfig is suitable for LLM calls with SSE streaming.
var DefaultRetryConfig = RetryConfig{
	Attempts:     3,
	Delay:        2 * time.Second,
	MaxDelay:     30 * time.Second,
	JitterFactor: 0.2, // ±20% jitter
}

// Do runs fn with retry, sending SSE progress updates on each attempt.
// If hub is nil, SSE updates are skipped (silent retry).
func (r *RetryConfig) Do(ctx context.Context, hub *SSEHub, clientID string, operation string, fn func() error) error {
	attempt := uint(0)

	return retry.Do(
		func() error {
			attempt++
			err := fn()

			if hub != nil && clientID != "" {
				status := "attempt"
				if err == nil {
					status = "success"
				}
				msg := fmt.Sprintf(`{"type":"retry","operation":"%s","attempt":%d,"status":"%s"}`, operation, attempt, status)
				hub.Send(clientID, []byte(msg))
			}

			if err != nil {
				slog.Warn("retry: attempt failed", "operation", operation, "attempt", attempt, "max_attempts", r.Attempts, "error", err)
			}
			return err
		},
		retry.Context(ctx),
		retry.Attempts(r.Attempts),
		retry.Delay(r.Delay),
		retry.MaxDelay(r.MaxDelay),
		retry.MaxJitter(time.Duration(float64(r.Delay) * r.JitterFactor)),
		retry.DelayType(func(n uint, err error, config *retry.Config) time.Duration {
			// Exponential backoff: delay * 2^(n-1) + jitter
			_ = err // delay independent of error
			d := time.Duration(float64(r.Delay) * float64(int(1)<<(n-1)))
			if d > r.MaxDelay {
				d = r.MaxDelay
			}
			return d
		}),
		retry.LastErrorOnly(true),
	)
}

// DoSilent runs fn with retry but NO SSE feedback (for internal/non-user-facing jobs).
func (r *RetryConfig) DoSilent(ctx context.Context, fn func() error) error {
	return r.Do(ctx, nil, "", "", fn)
}
