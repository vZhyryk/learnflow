package router

import (
	"context"
	"testing"
	"time"

	. "github.com/smartystreets/goconvey/convey"
)

// TestRedisRateLimitFailClosed is the regression test for the fail-open bug fixed in
// commit 44868d9: when Redis is unreachable, redisRateLimit must return (false, err) —
// i.e. deny the request — rather than allowing it through.
func TestRefillRate(t *testing.T) {
	Convey("refillRate converts requests per window into tokens per second", t, func() {
		So(refillRate(5, time.Minute), ShouldAlmostEqual, 5.0/60, 1e-9)
		So(refillRate(3, time.Hour), ShouldAlmostEqual, 3.0/3600, 1e-9)
		So(refillRate(10, time.Second), ShouldEqual, 10)
		So(refillRate(7, 0), ShouldEqual, 7)
	})
}

func TestRedisRateLimitFailClosed(t *testing.T) {
	Convey("redisRateLimit", t, func() {
		Convey("When Redis is unreachable, it fails closed (allowed=false) and returns an error", func() {
			route := newTestRouteHandler()
			defer route.App.Redis.Close() //nolint:errcheck // best-effort cleanup

			allowed, err := route.redisRateLimit(context.Background(), "test-key", 1, 1, time.Second)

			So(allowed, ShouldBeFalse)
			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldContainSubstring, "router.redisRateLimit")
		})

		Convey("When the context is already canceled, it fails closed (allowed=false) and returns an error", func() {
			route := newTestRouteHandler()
			defer route.App.Redis.Close() //nolint:errcheck // best-effort cleanup

			ctx, cancel := context.WithCancel(context.Background())
			cancel()

			allowed, err := route.redisRateLimit(ctx, "test-key", 1, 1, time.Second)

			So(allowed, ShouldBeFalse)
			So(err, ShouldNotBeNil)
		})
	})
}
