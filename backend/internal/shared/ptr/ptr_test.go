package ptr_test

import (
	"testing"

	"learnflow_backend/internal/shared/ptr"

	. "github.com/smartystreets/goconvey/convey"
)

func TestStringOrNil(t *testing.T) {
	Convey("StringOrNil", t, func() {
		Convey("empty string returns nil", func() {
			So(ptr.StringOrNil(""), ShouldBeNil)
		})
		Convey("non-empty string returns a pointer to it", func() {
			got := ptr.StringOrNil("Alice")
			So(got, ShouldNotBeNil)
			So(*got, ShouldEqual, "Alice")
		})
	})
}

func TestStringOrEmpty(t *testing.T) {
	Convey("StringOrEmpty", t, func() {
		Convey("nil returns empty string", func() {
			So(ptr.StringOrEmpty(nil), ShouldEqual, "")
		})
		Convey("non-nil returns the pointed-to value", func() {
			s := "Alice"
			So(ptr.StringOrEmpty(&s), ShouldEqual, "Alice")
		})
	})
}
