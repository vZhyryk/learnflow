package main

import (
	"learnflow_backend/internal/shared/testutil"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestNewApp(t *testing.T) {
	Convey("NewApp", t, func() {
		cfg := Config{
			Env: "test",
			SMTP: SMTP{
				Host:     "smtp.example.com",
				Port:     587,
				Username: "user",
				Password: "pass",
				Sender:   "noreply@example.com",
			},
		}
		log := testutil.NewTestLogger()
		redisClient := testutil.UnreachableRedis()
		defer redisClient.Close() //nolint:errcheck // best-effort cleanup

		app := NewApp(cfg, log, nil, redisClient)

		Convey("It wires all dependencies into the App container", func() {
			So(app.Config, ShouldResemble, cfg)
			So(app.Logger, ShouldEqual, log)
			So(app.RedisClient, ShouldEqual, redisClient)
			So(app.Outbox, ShouldNotBeNil)
			So(app.Publisher, ShouldNotBeNil)
			So(app.Mailer, ShouldNotBeNil)
		})
	})
}
