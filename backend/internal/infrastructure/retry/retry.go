package retry

import (
	"context"
	"fmt"
	"math/rand/v2"
	"time"
)

// maxBackoff caps the exponential backoff delay so a long attempts count can't leave a
// caller waiting for minutes (or overflow the 1<<i shift) between retries.
const maxBackoff = 30 * time.Second

// Do retries fn up to attempts times with backoff (capped, jittered against thundering
// herd). Returns the last error if all attempts fail.
func Do(ctx context.Context, attempts int, fn func() error) error {
	var err error
	for i := 1; i <= attempts; i++ {
		err = fn()
		if err == nil {
			return nil
		}

		// Shift is capped at 5 (1<<5s = 32s, already above maxBackoff) so large `attempts`
		// values can't overflow the shift before the maxBackoff clamp below applies.
		shift := min(i, 5)
		backoff := min(time.Duration(1<<shift)*time.Second, maxBackoff)
		jitter := time.Duration(rand.Int64N(int64(backoff) / 5)) //nolint:gosec // jitter timing, not security-sensitive

		t := time.NewTimer(backoff + jitter)
		select {
		case <-t.C:
		case <-ctx.Done():
			t.Stop()
			return fmt.Errorf("retry: context cancelled: %w", ctx.Err())
		}
	}
	return fmt.Errorf("retry: all %d attempts failed: %w", attempts, err)
}
