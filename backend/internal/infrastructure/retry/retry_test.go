package retry_test

import (
	"context"
	"errors"
	"testing"

	"learnflow_backend/internal/infrastructure/retry"

	. "github.com/smartystreets/goconvey/convey"
)

func TestDo(t *testing.T) {
	Convey("retry.Do", t, func() {
		Convey("When attempts is zero", func() {
			called := 0
			err := retry.Do(context.Background(), 0, func() error {
				called++
				return errors.New("fail")
			})
			So(called, ShouldEqual, 0)
			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldContainSubstring, "all 0 attempts failed")
		})

		Convey("When fn succeeds on first attempt", func() {
			called := 0
			err := retry.Do(context.Background(), 3, func() error {
				called++
				return nil
			})
			So(err, ShouldBeNil)
			So(called, ShouldEqual, 1)
		})

		Convey("When context is pre-cancelled", func() {
			ctx, cancel := context.WithCancel(context.Background())
			cancel()
			err := retry.Do(ctx, 3, func() error {
				return errors.New("fail")
			})
			So(errors.Is(err, context.Canceled), ShouldBeTrue)
		})

		// NOTE: sleeps 2s (backoff for attempt i=1 = 1<<1 seconds).
		Convey("When all attempts exhausted (attempts=1)", func() {
			called := 0
			err := retry.Do(context.Background(), 1, func() error {
				called++
				return errors.New("permanent failure")
			})
			So(called, ShouldEqual, 1)
			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldContainSubstring, "all 1 attempts failed")
			So(err.Error(), ShouldContainSubstring, "permanent failure")
		})
	})
}
