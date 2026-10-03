//go:build integration

package redis_test

import (
	"context"
	"fmt"
	"testing"
	"time"

	"learnflow_backend/internal/infrastructure/redis"

	. "github.com/smartystreets/goconvey/convey"
)

func newBlocklistInstance(t *testing.T) *redis.Instance {
	t.Helper()

	ri, err := redis.InitRedis("localhost:6379", "", redis.PoolConfig{PoolSize: 1})
	if err != nil {
		t.Fatalf("InitRedis: %v", err)
	}
	t.Cleanup(func() { ri.Close() }) //nolint:errcheck // test-local client, nothing to react to

	return ri
}

func uniqueID(prefix string) string {
	return fmt.Sprintf("%s-%d", prefix, time.Now().UnixNano())
}

func TestBlockUser_Integration(t *testing.T) {
	ri := newBlocklistInstance(t)
	ctx := context.Background()

	Convey("BlockUser / UnBlockUser / IsBlocked against real Redis", t, func() {
		userID := uniqueID("user")

		blocked, err := ri.IsBlocked(ctx, userID, "unused-jti")
		So(err, ShouldBeNil)
		So(blocked, ShouldBeFalse)

		So(ri.BlockUser(ctx, userID, time.Minute), ShouldBeNil)

		blocked, err = ri.IsBlocked(ctx, userID, "unused-jti")
		So(err, ShouldBeNil)
		So(blocked, ShouldBeTrue)

		ttl, err := ri.Raw().TTL(ctx, "user_blocked:"+userID).Result()
		So(err, ShouldBeNil)
		So(ttl, ShouldBeBetween, time.Duration(0), time.Minute+time.Second)

		So(ri.UnBlockUser(ctx, userID), ShouldBeNil)

		blocked, err = ri.IsBlocked(ctx, userID, "unused-jti")
		So(err, ShouldBeNil)
		So(blocked, ShouldBeFalse)
	})

	Convey("UnBlockUser on a user that was never blocked is not an error", t, func() {
		So(ri.UnBlockUser(ctx, uniqueID("never-blocked")), ShouldBeNil)
	})
}

func TestBlockToken_Integration(t *testing.T) {
	ri := newBlocklistInstance(t)
	ctx := context.Background()

	Convey("BlockToken / IsBlocked against real Redis", t, func() {
		jti := uniqueID("jti")

		blocked, err := ri.IsBlocked(ctx, "unused-user", jti)
		So(err, ShouldBeNil)
		So(blocked, ShouldBeFalse)

		So(ri.BlockToken(ctx, jti, time.Minute), ShouldBeNil)

		blocked, err = ri.IsBlocked(ctx, "unused-user", jti)
		So(err, ShouldBeNil)
		So(blocked, ShouldBeTrue)
	})

	Convey("A blocked token and a blocked user with the same id do not affect each other", t, func() {
		id := uniqueID("same")

		So(ri.BlockToken(ctx, id, time.Minute), ShouldBeNil)

		userBlocked, err := ri.IsBlocked(ctx, id, "unused-jti")
		So(err, ShouldBeNil)
		So(userBlocked, ShouldBeFalse)
	})
}

func TestIsBlocked_Integration(t *testing.T) {
	ri := newBlocklistInstance(t)
	ctx := context.Background()

	Convey("IsBlocked checks the user and the token in one call", t, func() {
		userID, jti := uniqueID("user"), uniqueID("jti")

		blocked, err := ri.IsBlocked(ctx, userID, jti)
		So(err, ShouldBeNil)
		So(blocked, ShouldBeFalse)

		So(ri.BlockUser(ctx, userID, time.Minute), ShouldBeNil)
		blocked, err = ri.IsBlocked(ctx, userID, jti)
		So(err, ShouldBeNil)
		So(blocked, ShouldBeTrue)

		So(ri.UnBlockUser(ctx, userID), ShouldBeNil)
		So(ri.BlockToken(ctx, jti, time.Minute), ShouldBeNil)
		blocked, err = ri.IsBlocked(ctx, userID, jti)
		So(err, ShouldBeNil)
		So(blocked, ShouldBeTrue)
	})
}
