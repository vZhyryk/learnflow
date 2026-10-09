package redis_test

import (
	"context"
	"testing"
	"time"

	"learnflow_backend/internal/shared/testutil"

	. "github.com/smartystreets/goconvey/convey"
)

func TestBlocklistUnreachableRedis(t *testing.T) {
	Convey("Given a Redis instance that cannot be reached", t, func() {
		ri := testutil.UnreachableRedisInstance()
		defer ri.Close() //nolint:errcheck // best-effort cleanup

		ctx := context.Background()

		Convey("BlockUser wraps the error", func() {
			err := ri.BlockUser(ctx, "user-1", time.Minute)
			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldContainSubstring, "redis.BlockUser")
		})

		Convey("UnBlockUser wraps the error", func() {
			err := ri.UnBlockUser(ctx, "user-1")
			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldContainSubstring, "redis.UnBlockUser")
		})

		Convey("BlockToken wraps the error", func() {
			err := ri.BlockToken(ctx, "jti-1", time.Minute)
			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldContainSubstring, "redis.BlockToken")
		})

		Convey("A non-positive ttl is rejected before touching Redis", func() {
			for _, ttl := range []time.Duration{0, -time.Second} {
				err := ri.BlockUser(ctx, "user-1", ttl)
				So(err, ShouldNotBeNil)
				So(err.Error(), ShouldContainSubstring, "redis.BlockUser: ttl must be positive")

				err = ri.BlockToken(ctx, "jti-1", ttl)
				So(err, ShouldNotBeNil)
				So(err.Error(), ShouldContainSubstring, "redis.BlockToken: ttl must be positive")

				err = ri.RevokeUserRole(ctx, "user-1", ttl)
				So(err, ShouldNotBeNil)
				So(err.Error(), ShouldContainSubstring, "redis.RevokeUserRole: ttl must be positive")
			}
		})

		Convey("IsBlocked reports false with the error, so callers can fail closed", func() {
			blocked, err := ri.IsBlocked(ctx, "user-1", "jti-1", time.Now())
			So(blocked, ShouldBeFalse)
			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldContainSubstring, "redis.IsBlocked")
		})
	})
}
