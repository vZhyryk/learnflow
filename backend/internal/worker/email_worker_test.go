package worker

import (
	"context"
	"fmt"
	"learnflow_backend/internal/shared/testutil"
	"testing"

	"github.com/redis/go-redis/v9"

	. "github.com/smartystreets/goconvey/convey"
)

// newTestEmailWorker returns an EmailWorker[string] with a discard logger — the only
// dependency handleRunBLPopErrors touches. Redis and Process/Validate callbacks are
// intentionally left unset: EmailWorker.redisClient is a concrete *redis.Client, so
// handleMessage/Run (which call SetNX/BLPop against it) aren't unit-testable here.
func newTestEmailWorker() *EmailWorker[string] {
	return &EmailWorker[string]{
		logger: testutil.NewTestLogger(),
		cfg:    Config[string]{EventType: "test.event"},
	}
}

func TestHandleRunBLPopErrorsRedisNil(t *testing.T) {
	Convey("Given an EmailWorker", t, func() {
		Convey("When BLPop returns redis.Nil (no message available)", func() {
			w := newTestEmailWorker()

			shouldContinue, shouldReturn := w.handleRunBLPopErrors(redis.Nil)

			So(shouldContinue, ShouldBeTrue)
			So(shouldReturn, ShouldBeFalse)
		})
	})
}

func TestHandleRunBLPopErrorsContextCanceled(t *testing.T) {
	Convey("Given an EmailWorker", t, func() {
		Convey("When BLPop fails because the context was canceled", func() {
			w := newTestEmailWorker()

			shouldContinue, shouldReturn := w.handleRunBLPopErrors(context.Canceled)

			So(shouldContinue, ShouldBeFalse)
			So(shouldReturn, ShouldBeTrue)
		})

		Convey("When BLPop fails because the context deadline was exceeded", func() {
			w := newTestEmailWorker()

			shouldContinue, shouldReturn := w.handleRunBLPopErrors(context.DeadlineExceeded)

			So(shouldContinue, ShouldBeFalse)
			So(shouldReturn, ShouldBeTrue)
		})

		Convey("When BLPop fails with a context error wrapped by another error", func() {
			w := newTestEmailWorker()

			shouldContinue, shouldReturn := w.handleRunBLPopErrors(fmt.Errorf("blpop: %w", context.Canceled))

			So(shouldContinue, ShouldBeFalse)
			So(shouldReturn, ShouldBeTrue)
		})
	})
}

func TestHandleRunBLPopErrorsUnexpectedError(t *testing.T) {
	Convey("Given an EmailWorker", t, func() {
		Convey("When BLPop fails with an unexpected, non-context error", func() {
			w := newTestEmailWorker()

			shouldContinue, shouldReturn := w.handleRunBLPopErrors(testutil.ErrRedisUnavailable)

			So(shouldContinue, ShouldBeTrue)
			So(shouldReturn, ShouldBeFalse)
		})
	})
}

func TestHandleRunBLPopErrorsNoError(t *testing.T) {
	Convey("Given an EmailWorker", t, func() {
		Convey("When BLPop succeeds", func() {
			w := newTestEmailWorker()

			shouldContinue, shouldReturn := w.handleRunBLPopErrors(nil)

			So(shouldContinue, ShouldBeFalse)
			So(shouldReturn, ShouldBeFalse)
		})
	})
}
