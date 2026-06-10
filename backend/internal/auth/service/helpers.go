package authservice

import (
	"context"
	"errors"
	"fmt"
	authdomain "learnflow_backend/internal/auth/domain"
	"time"
)

func retryOnConflict(ctx context.Context, delays []time.Duration, fn func() error) error {
	if len(delays) == 0 {
		delays = []time.Duration{50 * time.Millisecond, 100 * time.Millisecond, 200 * time.Millisecond}
	}

	for i, d := range delays {
		err := fn()
		if err == nil {
			return nil
		}
		if !errors.Is(err, authdomain.ErrConflict) {
			return err
		}
		if i < len(delays)-1 {
			select {
			case <-ctx.Done():
				return fmt.Errorf("context deadline: %w", ctx.Err())
			case <-time.After(d):
			}
		}
	}
	return fn()
}
