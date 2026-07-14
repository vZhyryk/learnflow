package worker

import (
	"context"
	"fmt"
	"learnflow_backend/internal/events"
	"learnflow_backend/internal/shared/testutil"
	"testing"

	"github.com/redis/go-redis/v9"

	. "github.com/smartystreets/goconvey/convey"
)

func newTestEmailWorker() *EmailWorker[map[string]string] {
	return &EmailWorker[map[string]string]{
		logger: testutil.NewTestLogger(),
		cfg: Config[map[string]string]{
			EventType: "test.event",
			Validate:  func(_ map[string]string) error { return nil },
		},
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

func TestHandleMessage(t *testing.T) {
	Convey("handleMessage", t, func() {
		Convey("Invalid JSON", func() {
			w := newTestEmailWorker()

			result, idempotencyKey, err := w.handleMessage(context.Background(), "")
			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldContainSubstring, "unmarshal")
			So(result, ShouldBeNil)
			So(idempotencyKey, ShouldBeEmpty)
		})

		Convey("Break validate", func() {
			w := newTestEmailWorker()
			w.cfg.Validate = func(_ map[string]string) error {
				return fmt.Errorf("some error")
			}

			result, idempotencyKey, err := w.handleMessage(context.Background(), `{"value": "value"}`)
			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldContainSubstring, "some error")
			So(result, ShouldBeNil)
			So(idempotencyKey, ShouldBeEmpty)
		})

		Convey("JSON null payload (e.g. a nil interface published upstream)", func() {
			w := &EmailWorker[events.RegistrationAttemptPayload]{
				logger: testutil.NewTestLogger(),
				cfg: Config[events.RegistrationAttemptPayload]{
					EventType: "registration_attempt",
					Validate:  ValidateRegistrationAttemptsPayload,
				},
			}

			result, idempotencyKey, err := w.handleMessage(context.Background(), "null")

			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldContainSubstring, "missing fields")
			So(result, ShouldBeNil)
			So(idempotencyKey, ShouldBeEmpty)
		})

		Convey("No redis (broken SetNX)", func() {
			w := newTestEmailWorker()
			w.cfg.Validate = func(_ map[string]string) error {
				return nil
			}

			w.cfg.IdempotencyKey = func(_ map[string]string) string {
				return "test_key:test_key"
			}

			So(func() {
				_, _, _ = w.handleMessage(context.Background(), `{"value": "value"}`) //nolint:errcheck // panics before returning; asserted via ShouldPanic
			}, ShouldPanic)
		})
	})
}
