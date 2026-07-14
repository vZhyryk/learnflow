//go:build integration

package db

import (
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestInitDatabase_Integration(t *testing.T) {
	Convey("InitDatabase", t, func() {
		Convey("When the address is reachable", func() {
			validDSN, _ := BuildDSNFromEnv()
			pool, err := InitDatabase(validDSN, "5m", "1h", 10, 2)

			So(err, ShouldBeNil)
			So(pool, ShouldNotBeNil)
		})
	})
}
