package db_test

import (
	"os"
	"strings"
	"testing"

	"learnflow_backend/internal/infrastructure/db"

	. "github.com/smartystreets/goconvey/convey"
)

const (
	envDBName     = "DB_NAME"
	envDBUser     = "DB_USER"
	envDBHost     = "DB_HOST"
	envDBPassword = "DB_PASSWORD"
	envDBPort     = "DB_PORT"
	envDBSSLMode  = "DB_SSLMODE"
)

func setRequiredEnv() {
	So(os.Setenv(envDBName, "testdb"), ShouldBeNil)
	So(os.Setenv(envDBUser, "testuser"), ShouldBeNil)
	So(os.Setenv(envDBHost, "localhost"), ShouldBeNil)
	So(os.Setenv(envDBPassword, "testpass"), ShouldBeNil)
}

func unsetAllEnv() {
	So(os.Unsetenv(envDBName), ShouldBeNil)
	So(os.Unsetenv(envDBUser), ShouldBeNil)
	So(os.Unsetenv(envDBHost), ShouldBeNil)
	So(os.Unsetenv(envDBPassword), ShouldBeNil)
	So(os.Unsetenv(envDBPort), ShouldBeNil)
	So(os.Unsetenv(envDBSSLMode), ShouldBeNil)
}

func TestBuildDSNFromEnv(t *testing.T) {
	Convey("BuildDSNFromEnv", t, func() {
		Reset(unsetAllEnv)

		Convey("When all required vars are set", func() {
			setRequiredEnv()
			dsn, err := db.BuildDSNFromEnv()
			So(err, ShouldBeNil)
			So(dsn, ShouldStartWith, "postgres://")
			So(dsn, ShouldContainSubstring, "testuser")
			So(dsn, ShouldContainSubstring, "testdb")
			So(dsn, ShouldContainSubstring, "localhost")
			So(dsn, ShouldContainSubstring, "sslmode=disable")
		})

		Convey("When DB_PORT is overridden", func() {
			setRequiredEnv()
			So(os.Setenv(envDBPort, "5433"), ShouldBeNil)
			dsn, err := db.BuildDSNFromEnv()
			So(err, ShouldBeNil)
			So(dsn, ShouldContainSubstring, ":5433/")
		})

		Convey("When DB_SSLMODE is overridden", func() {
			setRequiredEnv()
			So(os.Setenv(envDBSSLMode, "require"), ShouldBeNil)
			dsn, err := db.BuildDSNFromEnv()
			So(err, ShouldBeNil)
			So(dsn, ShouldContainSubstring, "sslmode=require")
		})

		missing := []struct {
			name  string
			unset string
			msg   string
		}{
			{"DB_NAME is missing", envDBName, "DB_NAME"},
			{"DB_USER is missing", envDBUser, "DB_USER"},
			{"DB_HOST is missing", envDBHost, "DB_HOST"},
			{"DB_PASSWORD is missing", envDBPassword, "DB_PASSWORD"},
		}
		for _, tc := range missing {
			Convey("When "+tc.name, func() {
				setRequiredEnv()
				So(os.Unsetenv(tc.unset), ShouldBeNil)
				_, err := db.BuildDSNFromEnv()
				So(err, ShouldNotBeNil)
				So(strings.Contains(err.Error(), tc.msg), ShouldBeTrue)
			})
		}
	})
}
