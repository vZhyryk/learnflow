package db

import (
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

const validDSN = "postgres://user:pass@localhost:5432/testdb"
const unreachableDSN = "postgres://user:pass@127.0.0.1:1/testdb"

func TestParseConfigs(t *testing.T) {
	Convey("parseConfigs", t, func() {
		Convey("When the DSN is invalid", func() {
			_, err := parseConfigs("not-a-dsn", "5m", "1h", 10, 2)
			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldContainSubstring, "failed to parse config")
		})

		Convey("When maxOpenConns is zero", func() {
			_, err := parseConfigs(validDSN, "5m", "1h", 0, 2)
			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldContainSubstring, "maxOpenConns must be between 1 and 100")
		})

		Convey("When maxOpenConns is negative", func() {
			_, err := parseConfigs(validDSN, "5m", "1h", -1, 2)
			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldContainSubstring, "maxOpenConns must be between 1 and 100")
		})

		Convey("When maxOpenConns exceeds 100", func() {
			_, err := parseConfigs(validDSN, "5m", "1h", 101, 2)
			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldContainSubstring, "maxOpenConns must be between 1 and 100")
		})

		Convey("When maxIdleTime cannot be parsed", func() {
			_, err := parseConfigs(validDSN, "not-a-duration", "1h", 10, 2)
			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldContainSubstring, "failed to parse max idle time")
		})

		Convey("When maxIdleTime is zero", func() {
			_, err := parseConfigs(validDSN, "0s", "1h", 10, 2)
			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldContainSubstring, "max idle time must be positive")
		})

		Convey("When maxIdleTime is negative", func() {
			_, err := parseConfigs(validDSN, "-5m", "1h", 10, 2)
			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldContainSubstring, "max idle time must be positive")
		})

		Convey("When maxLifetime cannot be parsed", func() {
			_, err := parseConfigs(validDSN, "5m", "not-a-duration", 10, 2)
			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldContainSubstring, "failed to parse max lifetime")
		})

		Convey("When maxLifetime is zero", func() {
			_, err := parseConfigs(validDSN, "5m", "0s", 10, 2)
			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldContainSubstring, "max lifetime must be positive")
		})

		Convey("When minOpenConns equals maxOpenConns", func() {
			_, err := parseConfigs(validDSN, "5m", "1h", 10, 10)
			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldContainSubstring, "minOpenConns")
		})

		Convey("When minOpenConns exceeds maxOpenConns", func() {
			_, err := parseConfigs(validDSN, "5m", "1h", 10, 20)
			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldContainSubstring, "minOpenConns")
		})

		Convey("When all inputs are valid", func() {
			cfg, err := parseConfigs(validDSN, "5m", "1h", 10, 2)
			So(err, ShouldBeNil)
			So(cfg.MaxConns, ShouldEqual, int32(10))
			So(cfg.MinConns, ShouldEqual, int32(2))
			So(cfg.MaxConnIdleTime.String(), ShouldEqual, "5m0s")
			So(cfg.MaxConnLifetime.String(), ShouldEqual, "1h0m0s")
		})
	})
}

func TestInitDatabase(t *testing.T) {
	Convey("InitDatabase", t, func() {
		Convey("When the address is unreachable", func() {
			pool, err := InitDatabase(unreachableDSN, "5m", "1h", 10, 2)

			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldContainSubstring, "failed to ping")
			So(pool, ShouldBeNil)
		})
	})
}
