package bootstrap_test

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"os"
	"os/exec"
	"testing"

	"learnflow_backend/internal/infrastructure/bootstrap"

	. "github.com/smartystreets/goconvey/convey"
)

func TestNewLogger(t *testing.T) {
	Convey("NewLogger", t, func() {
		Convey("When environment is dev", func() {
			l := bootstrap.NewLogger("dev")
			So(l, ShouldNotBeNil)
		})

		Convey("When environment is not dev", func() {
			l := bootstrap.NewLogger("production")
			So(l, ShouldNotBeNil)
		})
	})
}

const (
	envDBOpenConnLimit = "DB_OPEN_CONNECTION_LIMIT"
	envDBMinConnLimit  = "DB_MIN_CONNECTION_LIMIT"
	envDBMaxIdleTime   = "DB_MAX_IDLE_TIME"
	envDBMaxLifetime   = "DB_MAX_LIFETIME"
	envDBName          = "DB_NAME"
	envDBUser          = "DB_USER"
	envDBHost          = "DB_HOST"
	envDBPassword      = "DB_PASSWORD"
)

func unsetDatabaseConfigEnv() {
	for _, key := range []string{envDBOpenConnLimit, envDBMinConnLimit, envDBMaxIdleTime, envDBMaxLifetime, envDBName, envDBUser, envDBHost, envDBPassword} {
		So(os.Unsetenv(key), ShouldBeNil)
	}
}

func setDSNData() {
	So(os.Setenv(envDBName, "testdb"), ShouldBeNil)
	So(os.Setenv(envDBUser, "testuser"), ShouldBeNil)
	So(os.Setenv(envDBHost, "localhost"), ShouldBeNil)
	So(os.Setenv(envDBPassword, "testpass"), ShouldBeNil)
}

func TestGetDatabaseConfig(t *testing.T) {
	Convey("GetDatabaseConfig", t, func() {
		Reset(unsetDatabaseConfigEnv)

		Convey("When no env vars are set", func() {
			unsetDatabaseConfigEnv()
			maxOpenConns, minOpenConns, maxIdleTime, maxLifetime := bootstrap.GetDatabaseConfig()

			So(maxOpenConns, ShouldBeGreaterThanOrEqualTo, 25)
			So(minOpenConns, ShouldEqual, 2)
			So(maxIdleTime, ShouldEqual, "30m")
			So(maxLifetime, ShouldEqual, "1h")
		})

		Convey("When env vars override the defaults", func() {
			unsetDatabaseConfigEnv()

			So(os.Setenv(envDBOpenConnLimit, "100"), ShouldBeNil)
			So(os.Setenv(envDBMinConnLimit, "5"), ShouldBeNil)
			So(os.Setenv(envDBMaxIdleTime, "10m"), ShouldBeNil)
			So(os.Setenv(envDBMaxLifetime, "2h"), ShouldBeNil)

			maxOpenConns, minOpenConns, maxIdleTime, maxLifetime := bootstrap.GetDatabaseConfig()

			So(maxOpenConns, ShouldEqual, 100)
			So(minOpenConns, ShouldEqual, 5)
			So(maxIdleTime, ShouldEqual, "10m")
			So(maxLifetime, ShouldEqual, "2h")
		})
	})
}

func TestLoadDatabaseConfig(t *testing.T) {
	Convey("LoadDatabaseConfig", t, func() {
		Reset(unsetDatabaseConfigEnv)

		Convey("When required DSN env vars are set", func() {
			unsetDatabaseConfigEnv()
			setDSNData()

			cfg, err := bootstrap.LoadDatabaseConfig()
			So(err, ShouldBeNil)
			So(cfg.DSN, ShouldStartWith, "postgres://")
			So(cfg.DSN, ShouldContainSubstring, "testdb")
			So(cfg.DSN, ShouldContainSubstring, "testuser")
			So(cfg.DSN, ShouldContainSubstring, "localhost")
			So(cfg.DSN, ShouldContainSubstring, "testpass")
			So(cfg.MaxIdleTime, ShouldEqual, "30m")
			So(cfg.MaxLifetime, ShouldEqual, "1h")
			So(cfg.MinOpenConns, ShouldEqual, 2)
		})

		Convey("When a required DSN env var is missing", func() {
			unsetDatabaseConfigEnv()

			_, err := bootstrap.LoadDatabaseConfig()

			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldContainSubstring, "failed to resolve database DSN")
		})
	})
}

// TestMustInitInfraCrasher is not a real test — it's the subprocess entry point
// spawned by TestMustInitInfraExitsOnDatabaseFailure. MustInitInfra calls
// Logger.Fatal on failure, which calls os.Exit(1); calling it in-process would
// kill the whole test binary before any assertion runs (see the parent test's
// comment for the confirmed failure mode). Re-exec'ing the test binary lets us
// observe that os.Exit(1) without losing the rest of the suite.
func TestMustInitInfraCrasher(t *testing.T) {
	if os.Getenv("KB_MUST_INIT_INFRA_CRASH") != "1" {
		return
	}
	jsonLogger := bootstrap.NewLogger("production")
	cfg := bootstrap.DatabaseConfig{
		DSN:          "not-a-valid-dsn",
		MaxIdleTime:  "30m",
		MaxLifetime:  "1h",
		MaxOpenConns: 5,
		MinOpenConns: 2,
	}
	bootstrap.MustInitInfra(cfg, jsonLogger)
	t.Fatal("MustInitInfra returned instead of exiting — Fatal path did not trigger")
}

func TestMustInitInfraExitsOnDatabaseFailure(t *testing.T) {
	Convey("MustInitInfra", t, func() {
		Convey("When the database fails to initialize, it logs and exits(1) instead of returning", func() {
			cmd := exec.CommandContext(context.Background(), os.Args[0], "-test.run=^TestMustInitInfraCrasher$") //nolint:gosec // fixed args, test-only re-exec of this same binary
			cmd.Env = append(os.Environ(), "KB_MUST_INIT_INFRA_CRASH=1")
			var out bytes.Buffer
			cmd.Stdout = &out
			cmd.Stderr = &out

			runErr := cmd.Run()

			var exitErr *exec.ExitError
			So(errors.As(runErr, &exitErr), ShouldBeTrue)
			So(exitErr.ExitCode(), ShouldEqual, 1)

			var entry map[string]any
			So(json.Unmarshal(out.Bytes(), &entry), ShouldBeNil)
			So(entry["message"], ShouldContainSubstring, "db: failed to parse config")
		})
	})
}
