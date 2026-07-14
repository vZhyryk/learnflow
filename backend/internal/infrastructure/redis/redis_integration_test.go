//go:build integration

package redis_test

import (
	"context"
	"testing"

	"learnflow_backend/internal/infrastructure/redis"

	goredis "github.com/redis/go-redis/v9"
	. "github.com/smartystreets/goconvey/convey"
)

// withRequirePass enables requirepass on the shared docker-compose.tests.yml
// Redis instance for the duration of fn, then unsets it again via t.Cleanup.
// Each of the three integration tests below has exactly one Convey leaf, so
// GoConvey's top-down re-execution (which would otherwise replay this setup
// mid-test and leak requirepass into unrelated leaves) never applies here.
func withRequirePass(t *testing.T, password string) {
	admin := goredis.NewClient(&goredis.Options{Addr: "localhost:6379"})
	if err := admin.ConfigSet(context.Background(), "requirepass", password).Err(); err != nil {
		admin.Close() //nolint:errcheck // best-effort cleanup on setup failure
		t.Fatalf("withRequirePass: enable: %v", err)
	}
	admin.Close() //nolint:errcheck // best-effort cleanup

	t.Cleanup(func() {
		resetAdmin := goredis.NewClient(&goredis.Options{Addr: "localhost:6379", Password: password})
		resetAdmin.ConfigSet(context.Background(), "requirepass", "") //nolint:errcheck // best-effort cleanup
		resetAdmin.Close()                                            //nolint:errcheck // best-effort cleanup
	})
}

func TestInitRedisNoPasswordConfigured_Integration(t *testing.T) {
	Convey("InitRedis, when the server is reachable and no password is configured or sent", t, func() {
		client, err := redis.InitRedis("localhost:6379", "", redis.PoolConfig{PoolSize: 1, MinIdleConns: 0, MaxRetries: 0})

		So(err, ShouldBeNil)
		So(client, ShouldNotBeNil)
		client.Close() //nolint:errcheck // test-local client, nothing to react to
	})
}

func TestInitRedisWrongPassword_Integration(t *testing.T) {
	withRequirePass(t, "integration-test-pass")

	Convey("InitRedis, when a password is configured and the wrong one is sent", t, func() {
		client, err := redis.InitRedis("localhost:6379", "wrong-password", redis.PoolConfig{PoolSize: 1, MinIdleConns: 0, MaxRetries: 0})
		So(err, ShouldNotBeNil)
		So(err.Error(), ShouldContainSubstring, "failed to ping")
		So(err.Error(), ShouldContainSubstring, "WRONGPASS")
		So(client, ShouldBeNil)
	})
}

func TestInitRedisCorrectPassword_Integration(t *testing.T) {
	withRequirePass(t, "integration-test-pass")

	Convey("InitRedis, when a password is configured and the correct one is sent", t, func() {
		client, err := redis.InitRedis("localhost:6379", "integration-test-pass", redis.PoolConfig{PoolSize: 1, MinIdleConns: 0, MaxRetries: 0})

		So(err, ShouldBeNil)
		So(client, ShouldNotBeNil)
		client.Close() //nolint:errcheck // test-local client, nothing to react to
	})
}
