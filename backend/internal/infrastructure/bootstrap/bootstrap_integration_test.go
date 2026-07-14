//go:build integration

package bootstrap_test

import (
	"context"
	"learnflow_backend/internal/infrastructure/bootstrap"
	"learnflow_backend/internal/shared/testutil"
	"os"
	"testing"
	"time"

	. "github.com/smartystreets/goconvey/convey"
)

const (
	redisPoolSize        = "REDIS_POOL_SIZE"
	redisMinIdleConns    = "REDIS_MIN_IDLE_CONNS"
	redisMaxRetries      = "REDIS_MAX_RETRIES"
	redisConnMaxLifetime = "REDIS_CONN_MAX_LIFETIME"
	redisAddr            = "REDIS_ADDR"
	redisPassword        = "REDIS_PASSWORD"
)

func unsetRedisConfigEnv() {
	for _, key := range []string{
		redisPoolSize, redisMinIdleConns, redisMaxRetries, redisConnMaxLifetime, redisAddr, redisPassword} {
		So(os.Unsetenv(key), ShouldBeNil)
	}
}

func unsetAllConfigEnv() {
	unsetDatabaseConfigEnv()
	unsetRedisConfigEnv()
}

func setRedisConfigEnv() {
	So(os.Setenv(redisAddr, "127.0.0.1:6379"), ShouldBeNil)
	So(os.Setenv(redisPoolSize, "20"), ShouldBeNil)
	So(os.Setenv(redisMinIdleConns, "4"), ShouldBeNil)
	So(os.Setenv(redisMaxRetries, "5"), ShouldBeNil)
	So(os.Setenv(redisConnMaxLifetime, "10m"), ShouldBeNil)
	So(os.Setenv(redisPassword, "20"), ShouldBeNil)
}

func TestGetRedis_Integration(t *testing.T) {
	Convey("GetRedis", t, func() {
		Reset(unsetRedisConfigEnv)
		Convey("When no env vars are set", func() {
			unsetRedisConfigEnv()
			redisClient, err := bootstrap.GetRedis()
			So(err, ShouldNotBeNil)
			So(redisClient, ShouldBeNil)
		})
		Convey("When env vars override the defaults", func() {
			setRedisConfigEnv()
			redisClient, err := bootstrap.GetRedis()
			So(err, ShouldBeNil)
			So(redisClient.Options().Addr, ShouldEqual, "127.0.0.1:6379")
			So(redisClient.Options().Password, ShouldEqual, "20")
			So(redisClient.Options().MaxRetries, ShouldEqual, 5)
			So(redisClient.Options().MinIdleConns, ShouldEqual, 4)
			So(redisClient.Options().PoolSize, ShouldEqual, 20)
			So(redisClient.Options().ConnMaxLifetime, ShouldEqual, 10*time.Minute)
		})
	})
}

func TestMustInitInfra_Integration(t *testing.T) {
	Convey("MustInitInfra_Integration", t, func() {
		Reset(unsetAllConfigEnv)
		Convey("When required env vars point at a real, reachable Postgres and Redis", func() {
			So(os.Setenv(redisAddr, "localhost:6379"), ShouldBeNil)
			cfg, err := bootstrap.LoadDatabaseConfig()
			So(err, ShouldBeNil)
			jsonLogger := testutil.NewTestLogger()

			db, redisClient, cleanup := bootstrap.MustInitInfra(cfg, jsonLogger)
			defer cleanup()

			So(db.Ping(context.Background()), ShouldBeNil)
			_, err = redisClient.Ping(context.Background()).Result()
			So(err, ShouldBeNil)
		})
	})
}
