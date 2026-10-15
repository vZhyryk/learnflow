package validator

import (
	"errors"
	"strings"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

var errTestInvalid = errors.New("test invalid")

func TestRequireOptionalContentDescription(t *testing.T) {
	Convey("RequireOptionalContentDescription", t, func() {
		Convey("nil is valid", func() {
			So(RequireOptionalContentDescription(nil, errTestInvalid), ShouldBeNil)
		})

		Convey("empty is invalid", func() {
			empty := ""
			So(RequireOptionalContentDescription(&empty, errTestInvalid), ShouldEqual, errTestInvalid)
		})

		Convey("10000 runes is valid (boundary)", func() {
			val := strings.Repeat("A", 10000)
			So(RequireOptionalContentDescription(&val, errTestInvalid), ShouldBeNil)
		})

		Convey("10001 runes is invalid (boundary)", func() {
			val := strings.Repeat("A", 10001)
			So(RequireOptionalContentDescription(&val, errTestInvalid), ShouldEqual, errTestInvalid)
		})
	})
}

func TestRequireOptionalContentBody(t *testing.T) {
	Convey("RequireOptionalContentBody", t, func() {
		Convey("nil is valid", func() {
			So(RequireOptionalContentBody(nil, errTestInvalid), ShouldBeNil)
		})

		Convey("empty is invalid", func() {
			empty := ""
			So(RequireOptionalContentBody(&empty, errTestInvalid), ShouldEqual, errTestInvalid)
		})

		Convey("10000 runes is valid (boundary)", func() {
			val := strings.Repeat("A", 10000)
			So(RequireOptionalContentBody(&val, errTestInvalid), ShouldBeNil)
		})

		Convey("10001 runes is invalid (boundary)", func() {
			val := strings.Repeat("A", 10001)
			So(RequireOptionalContentBody(&val, errTestInvalid), ShouldEqual, errTestInvalid)
		})
	})
}

func TestRequireOptionalSeoTitle(t *testing.T) {
	Convey("RequireOptionalSeoTitle", t, func() {
		Convey("nil is valid", func() {
			So(RequireOptionalSeoTitle(nil, errTestInvalid), ShouldBeNil)
		})

		Convey("empty is invalid", func() {
			empty := ""
			So(RequireOptionalSeoTitle(&empty, errTestInvalid), ShouldEqual, errTestInvalid)
		})

		Convey("70 runes is valid (boundary)", func() {
			val := strings.Repeat("A", 70)
			So(RequireOptionalSeoTitle(&val, errTestInvalid), ShouldBeNil)
		})

		Convey("71 runes is invalid (boundary)", func() {
			val := strings.Repeat("A", 71)
			So(RequireOptionalSeoTitle(&val, errTestInvalid), ShouldEqual, errTestInvalid)
		})
	})
}

func TestRequireOptionalSeoDescription(t *testing.T) {
	Convey("RequireOptionalSeoDescription", t, func() {
		Convey("nil is valid", func() {
			So(RequireOptionalSeoDescription(nil, errTestInvalid), ShouldBeNil)
		})

		Convey("empty is invalid", func() {
			empty := ""
			So(RequireOptionalSeoDescription(&empty, errTestInvalid), ShouldEqual, errTestInvalid)
		})

		Convey("160 runes is valid (boundary)", func() {
			val := strings.Repeat("A", 160)
			So(RequireOptionalSeoDescription(&val, errTestInvalid), ShouldBeNil)
		})

		Convey("161 runes is invalid (boundary)", func() {
			val := strings.Repeat("A", 161)
			So(RequireOptionalSeoDescription(&val, errTestInvalid), ShouldEqual, errTestInvalid)
		})
	})
}

func TestRequireOptionalHTTPSURL(t *testing.T) {
	Convey("RequireOptionalHTTPSURL", t, func() {
		Convey("nil is valid", func() {
			So(RequireOptionalHTTPSURL(nil, errTestInvalid), ShouldBeNil)
		})

		Convey("empty is invalid", func() {
			empty := ""
			So(RequireOptionalHTTPSURL(&empty, errTestInvalid), ShouldEqual, errTestInvalid)
		})

		Convey("non-https is invalid", func() {
			val := "http://example.com/image.png"
			So(RequireOptionalHTTPSURL(&val, errTestInvalid), ShouldEqual, errTestInvalid)
		})

		Convey("valid https URL", func() {
			val := "https://example.com/image.png"
			So(RequireOptionalHTTPSURL(&val, errTestInvalid), ShouldBeNil)
		})
	})
}

func TestRequireOptionalPositiveInt(t *testing.T) {
	Convey("RequireOptionalPositiveInt", t, func() {
		Convey("nil is valid", func() {
			So(RequireOptionalPositiveInt(nil, errTestInvalid), ShouldBeNil)
		})

		Convey("positive value is valid", func() {
			val := 30
			So(RequireOptionalPositiveInt(&val, errTestInvalid), ShouldBeNil)
		})

		Convey("zero is invalid", func() {
			val := 0
			So(RequireOptionalPositiveInt(&val, errTestInvalid), ShouldEqual, errTestInvalid)
		})

		Convey("negative value is invalid", func() {
			val := -1
			So(RequireOptionalPositiveInt(&val, errTestInvalid), ShouldEqual, errTestInvalid)
		})
	})
}
