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

// setRequiredEnv sets the four vars BuildDSNFromEnv requires, plus resets the
// two optional ones to unset. Everything goes through t.Setenv/unsetEnv so
// each var is restored to its pre-test value once TestBuildDSNFromEnv
// returns — a raw os.Setenv/Unsetenv leaks DB_* overrides for the rest of the
// process, breaking any test running later in the same binary that needs the
// real DB_* env (e.g. testutil.NewTestPool, used by other _test.go files in
// this package for real-Postgres integration tests).
func setRequiredEnv(t *testing.T) {
	t.Helper()
	t.Setenv(envDBName, "testdb")
	t.Setenv(envDBUser, "testuser")
	t.Setenv(envDBHost, "localhost")
	t.Setenv(envDBPassword, "testpass")
	unsetEnv(t, envDBPort)
	unsetEnv(t, envDBSSLMode)
}

// unsetEnv temporarily unsets key, restoring its prior value (set or unset)
// once the current test returns.
func unsetEnv(t *testing.T, key string) {
	t.Helper()

	orig, wasSet := os.LookupEnv(key)
	if err := os.Unsetenv(key); err != nil {
		t.Fatalf("unsetEnv(%s): %v", key, err)
	}
	t.Cleanup(func() {
		if wasSet {
			os.Setenv(key, orig) //nolint:errcheck,gosec // best-effort env restore in test cleanup
		}
	})
}

func TestBuildDSNFromEnv(t *testing.T) {
	Convey("BuildDSNFromEnv", t, func() {
		Convey("When all required vars are set", func() {
			setRequiredEnv(t)
			dsn, err := db.BuildDSNFromEnv()
			So(err, ShouldBeNil)
			So(dsn, ShouldStartWith, "postgres://")
			So(dsn, ShouldContainSubstring, "testuser")
			So(dsn, ShouldContainSubstring, "testdb")
			So(dsn, ShouldContainSubstring, "localhost")
			So(dsn, ShouldContainSubstring, "sslmode=require")
		})

		Convey("When DB_PORT is overridden", func() {
			setRequiredEnv(t)
			t.Setenv(envDBPort, "5433")
			dsn, err := db.BuildDSNFromEnv()
			So(err, ShouldBeNil)
			So(dsn, ShouldContainSubstring, ":5433/")
		})

		Convey("When DB_SSLMODE is overridden", func() {
			setRequiredEnv(t)
			t.Setenv(envDBSSLMode, "disable")
			dsn, err := db.BuildDSNFromEnv()
			So(err, ShouldBeNil)
			So(dsn, ShouldContainSubstring, "sslmode=disable")
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
				setRequiredEnv(t)
				unsetEnv(t, tc.unset)
				_, err := db.BuildDSNFromEnv()
				So(err, ShouldNotBeNil)
				So(strings.Contains(err.Error(), tc.msg), ShouldBeTrue)
			})
		}
	})
}
