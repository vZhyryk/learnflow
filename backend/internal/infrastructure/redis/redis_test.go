package redis_test

import (
	"testing"

	"learnflow_backend/internal/infrastructure/redis"

	. "github.com/smartystreets/goconvey/convey"
)

func TestInitRedis(t *testing.T) {
	Convey("InitRedis", t, func() {
		Convey("When the address is unreachable", func() {
			client, err := redis.InitRedis("127.0.0.1:9", "", redis.PoolConfig{PoolSize: 1, MinIdleConns: 0, MaxRetries: 0})

			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldContainSubstring, "failed to ping")
			So(client, ShouldBeNil)
		})
	})
}
