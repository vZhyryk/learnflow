package convert_test

import (
	"testing"

	"learnflow_backend/internal/infrastructure/convert"

	. "github.com/smartystreets/goconvey/convey"
)

func TestToMapStringAny(t *testing.T) {
	Convey("ToMapStringAny", t, func() {
		Convey("When v is nil", func() {
			m, ok := convert.ToMapStringAny(nil)
			So(ok, ShouldBeFalse)
			So(m, ShouldBeNil)
		})

		Convey("When v is map[string]any", func() {
			in := map[string]any{"key": 42}
			m, ok := convert.ToMapStringAny(in)
			So(ok, ShouldBeTrue)
			So(m["key"], ShouldEqual, 42)
		})

		Convey("When v is map[string]string", func() {
			in := map[string]string{"a": "b"}
			m, ok := convert.ToMapStringAny(in)
			So(ok, ShouldBeTrue)
			So(m["a"], ShouldEqual, "b")
		})

		Convey("When v is a JSON-serialisable struct", func() {
			in := struct {
				Name string `json:"name"`
				Age  int    `json:"age"`
			}{Name: "Alice", Age: 30}
			m, ok := convert.ToMapStringAny(in)
			So(ok, ShouldBeTrue)
			So(m["name"], ShouldEqual, "Alice")
			So(m["age"], ShouldEqual, float64(30))
		})

		Convey("When v is a non-object JSON type (slice)", func() {
			m, ok := convert.ToMapStringAny([]string{"a", "b"})
			So(ok, ShouldBeFalse)
			So(m, ShouldBeNil)
		})

		Convey("When v is a scalar int", func() {
			m, ok := convert.ToMapStringAny(42)
			So(ok, ShouldBeFalse)
			So(m, ShouldBeNil)
		})
	})
}
