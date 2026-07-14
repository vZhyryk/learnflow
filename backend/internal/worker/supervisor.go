package worker

import (
	"context"
	"fmt"
	"runtime/debug"
	"time"

	"learnflow_backend/internal/infrastructure/logger"
)

// restartBackoff is the pause between w.Run returning (or panicking) and being restarted.
// Workers are expected to run forever via their own internal blocking loop (e.g. Redis
// BLPop) — Run returning is itself unexpected, so this loop restarts indefinitely rather
// than giving up after N attempts: a long-lived background worker that permanently exits
// on a transient error (e.g. a momentary DB blip) is worse than one that keeps retrying at
// a fixed, cheap interval. There is intentionally no max-retry cap.
const restartBackoff = time.Second

// RunWithRecovery runs w.Run in a loop, recovering from and logging any panic, then
// restarting w.Run after restartBackoff. It returns once ctx is done. Callers are
// responsible for their own WaitGroup bookkeeping around this call.
func RunWithRecovery(ctx context.Context, log *logger.Logger, w Worker) {
	for {
		func() {
			defer func() {
				if r := recover(); r != nil {
					log.Error(fmt.Errorf("worker panic: %v\n%s", r, debug.Stack()), nil)
				}
			}()
			w.Run(ctx)
		}()
		if ctx.Err() != nil {
			return
		}
		time.Sleep(restartBackoff)
	}
}
