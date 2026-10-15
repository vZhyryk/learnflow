package testutil_test

import (
	"os"
	"testing"

	"learnflow_backend/internal/infrastructure/db"
	"learnflow_backend/internal/shared/testutil"

	. "github.com/smartystreets/goconvey/convey"
)

func TestSetRequiredDBEnv(t *testing.T) {
	Convey("Given DB_* vars in a known pre-test state", t, func() {
		t.Setenv("DB_NAME", "original-name")
		t.Setenv("DB_HOST", "")
		So(os.Unsetenv("DB_HOST"), ShouldBeNil)

		var duringName, duringUser, duringHost, duringPassword, dsn string
		var dsnErr error

		Convey("When SetRequiredDBEnv runs inside a subtest", func() {
			// Assertions can't run inside t.Run: Convey's context is bound to the outer goroutine.
			t.Run("inner", func(inner *testing.T) {
				testutil.SetRequiredDBEnv(inner)
				duringName = os.Getenv("DB_NAME")
				duringUser = os.Getenv("DB_USER")
				duringHost = os.Getenv("DB_HOST")
				duringPassword = os.Getenv("DB_PASSWORD")
				dsn, dsnErr = db.BuildDSNFromEnv()
			})

			Convey("Then all four required vars are set while the subtest runs", func() {
				So(duringName, ShouldEqual, "testdb")
				So(duringUser, ShouldEqual, "testuser")
				So(duringHost, ShouldEqual, "localhost")
				So(duringPassword, ShouldEqual, "testpass")
			})

			Convey("Then db.BuildDSNFromEnv accepts them", func() {
				So(dsnErr, ShouldBeNil)
				So(dsn, ShouldContainSubstring, "testdb")
			})

			Convey("Then the prior state is restored after the subtest", func() {
				So(os.Getenv("DB_NAME"), ShouldEqual, "original-name")
				_, hostSet := os.LookupEnv("DB_HOST")
				So(hostSet, ShouldBeFalse)
			})
		})
	})
}
