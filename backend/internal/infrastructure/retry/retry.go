package retry

import (
	"context"
	"fmt"
	"time"
)

// Do retries fn up to attempts times with exponential backoff. Returns the last error if all attempts fail.
func Do(ctx context.Context, attempts int, fn func() error) error {
	var err error
	for i := 1; i <= attempts; i++ {
		err = fn()
		if err == nil {
			return nil
		}
		t := time.NewTimer(time.Duration(1<<i) * time.Second)
		select {
		case <-t.C:
		case <-ctx.Done():
			t.Stop()
			return fmt.Errorf("retry: context cancelled: %w", ctx.Err())
		}
	}
	return fmt.Errorf("retry: all %d attempts failed: %w", attempts, err)
}
