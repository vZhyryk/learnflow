package worker

import (
	"context"
	"fmt"
	"runtime/debug"
	"time"

	"learnflow_backend/internal/infrastructure/logger"
)

// restartBackoff is the pause before restarting w.Run after it returns or panics —
// intentionally no max-retry cap, since a worker that exits forever on a transient error is worse.
const restartBackoff = time.Second

// RunWithRecovery runs w.Run in a loop, recovering/logging panics and restarting after
// restartBackoff, until ctx is done. Callers own their own WaitGroup bookkeeping.
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
