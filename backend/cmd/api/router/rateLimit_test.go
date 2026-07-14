package router

import (
	"context"
	"learnflow_backend/internal/shared/testutil"
	"testing"
	"time"

	. "github.com/smartystreets/goconvey/convey"
)

// TestRedisRateLimitFailClosed is the regression test for the fail-open bug fixed in
// commit 44868d9: when Redis is unreachable, redisRateLimit must return (false, err) —
// i.e. deny the request — rather than allowing it through.
func TestRedisRateLimitFailClosed(t *testing.T) {
	Convey("redisRateLimit", t, func() {
		Convey("When Redis is unreachable, it fails closed (allowed=false) and returns an error", func() {
			rdb := testutil.UnreachableRedis()
			defer rdb.Close() //nolint:errcheck // best-effort cleanup

			allowed, err := redisRateLimit(context.Background(), rdb, "test-key", 1, 1, time.Second)

			So(allowed, ShouldBeFalse)
			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldContainSubstring, "redisRateLimit")
		})

		Convey("When the context is already canceled, it fails closed (allowed=false) and returns an error", func() {
			rdb := testutil.UnreachableRedis()
			defer rdb.Close() //nolint:errcheck // best-effort cleanup

			ctx, cancel := context.WithCancel(context.Background())
			cancel()

			allowed, err := redisRateLimit(ctx, rdb, "test-key", 1, 1, time.Second)

			So(allowed, ShouldBeFalse)
			So(err, ShouldNotBeNil)
		})
	})
}
