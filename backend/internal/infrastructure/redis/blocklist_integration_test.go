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

		blocked, err := ri.IsBlocked(ctx, userID, "unused-jti", time.Now())
		So(err, ShouldBeNil)
		So(blocked, ShouldBeFalse)

		So(ri.BlockUser(ctx, userID, time.Minute), ShouldBeNil)

		blocked, err = ri.IsBlocked(ctx, userID, "unused-jti", time.Now())
		So(err, ShouldBeNil)
		So(blocked, ShouldBeTrue)

		ttl, err := ri.Raw().TTL(ctx, "user_blocked:"+userID).Result()
		So(err, ShouldBeNil)
		So(ttl, ShouldBeBetween, time.Duration(0), time.Minute+time.Second)

		So(ri.UnBlockUser(ctx, userID), ShouldBeNil)

		blocked, err = ri.IsBlocked(ctx, userID, "unused-jti", time.Now())
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

		blocked, err := ri.IsBlocked(ctx, "unused-user", jti, time.Now())
		So(err, ShouldBeNil)
		So(blocked, ShouldBeFalse)

		So(ri.BlockToken(ctx, jti, time.Minute), ShouldBeNil)

		blocked, err = ri.IsBlocked(ctx, "unused-user", jti, time.Now())
		So(err, ShouldBeNil)
		So(blocked, ShouldBeTrue)
	})

	Convey("A blocked token and a blocked user with the same id do not affect each other", t, func() {
		id := uniqueID("same")

		So(ri.BlockToken(ctx, id, time.Minute), ShouldBeNil)

		userBlocked, err := ri.IsBlocked(ctx, id, "unused-jti", time.Now())
		So(err, ShouldBeNil)
		So(userBlocked, ShouldBeFalse)
	})
}

func TestIsBlocked_Integration(t *testing.T) {
	ri := newBlocklistInstance(t)
	ctx := context.Background()

	Convey("IsBlocked checks the user and the token in one call", t, func() {
		userID, jti := uniqueID("user"), uniqueID("jti")

		blocked, err := ri.IsBlocked(ctx, userID, jti, time.Now())
		So(err, ShouldBeNil)
		So(blocked, ShouldBeFalse)

		So(ri.BlockUser(ctx, userID, time.Minute), ShouldBeNil)
		blocked, err = ri.IsBlocked(ctx, userID, jti, time.Now())
		So(err, ShouldBeNil)
		So(blocked, ShouldBeTrue)

		So(ri.UnBlockUser(ctx, userID), ShouldBeNil)
		So(ri.BlockToken(ctx, jti, time.Minute), ShouldBeNil)
		blocked, err = ri.IsBlocked(ctx, userID, jti, time.Now())
		So(err, ShouldBeNil)
		So(blocked, ShouldBeTrue)
	})
}

func TestRevokeUserTokens_Integration(t *testing.T) {
	ri := newBlocklistInstance(t)
	ctx := context.Background()

	Convey("RevokeUserTokens revokes the tokens issued up to now and not the ones issued later", t, func() {
		userID := uniqueID("user")
		now := time.Now()

		revoked, err := ri.IsBlocked(ctx, userID, "unused-jti", now)
		So(err, ShouldBeNil)
		So(revoked, ShouldBeFalse)

		So(ri.RevokeUserTokens(ctx, userID, time.Minute), ShouldBeNil)

		for name, issuedAt := range map[string]time.Time{
			"issued a minute ago":       now.Add(-time.Minute),
			"issued in the same second": now,
		} {
			revoked, err = ri.IsBlocked(ctx, userID, "unused-jti", issuedAt)
			So(err, ShouldBeNil)
			So(name+": "+boolText(revoked), ShouldEqual, name+": revoked")
		}

		revoked, err = ri.IsBlocked(ctx, userID, "unused-jti", now.Add(2*time.Second))
		So(err, ShouldBeNil)
		So(revoked, ShouldBeFalse)
	})

	Convey("The revocation marker expires with its ttl", t, func() {
		userID := uniqueID("user")
		So(ri.RevokeUserTokens(ctx, userID, time.Minute), ShouldBeNil)

		ttl, err := ri.Raw().TTL(ctx, "tokens_revoked_before:"+userID).Result()

		So(err, ShouldBeNil)
		So(ttl, ShouldBeBetween, time.Duration(0), time.Minute+time.Second)
	})

	Convey("Revoking one user's tokens does not affect another user", t, func() {
		revokedUser, otherUser := uniqueID("user-a"), uniqueID("user-b")
		So(ri.RevokeUserTokens(ctx, revokedUser, time.Minute), ShouldBeNil)

		revoked, err := ri.IsBlocked(ctx, otherUser, "unused-jti", time.Now().Add(-time.Hour))

		So(err, ShouldBeNil)
		So(revoked, ShouldBeFalse)
	})

	Convey("A non-positive ttl is rejected", t, func() {
		So(ri.RevokeUserTokens(ctx, uniqueID("user"), 0), ShouldNotBeNil)
	})
}

func boolText(revoked bool) string {
	if revoked {
		return "revoked"
	}

	return "not revoked"
}
