package admindomain

import (
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestUserDataDisplayName(t *testing.T) {
	first, empty := "John", ""

	Convey("UserData.DisplayName", t, func() {
		Convey("returns the first name when set", func() {
			So((&UserData{FirstName: &first}).DisplayName(), ShouldEqual, "John")
		})

		Convey("falls back to a generic name when the first name is nil", func() {
			So((&UserData{}).DisplayName(), ShouldEqual, "User")
		})

		Convey("falls back to a generic name when the first name is empty", func() {
			So((&UserData{FirstName: &empty}).DisplayName(), ShouldEqual, "User")
		})
	})
}
