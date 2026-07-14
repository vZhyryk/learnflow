package convert_test

import (
	"testing"
	"time"

	"learnflow_backend/internal/infrastructure/convert"

	"github.com/jackc/pgx/v5/pgtype"
	. "github.com/smartystreets/goconvey/convey"
)

func TestFormatNullableDate(t *testing.T) {
	Convey("FormatNullableDate", t, func() {
		Convey("When the date is NULL", func() {
			d := pgtype.Date{Valid: false}
			got := convert.FormatNullableDate(d, "2006-01-02")
			So(got, ShouldBeNil)
		})

		Convey("When the date is valid", func() {
			d := pgtype.Date{Time: time.Date(1990, 5, 17, 0, 0, 0, 0, time.UTC), Valid: true}
			got := convert.FormatNullableDate(d, "2006-01-02")
			So(got, ShouldNotBeNil)
			So(*got, ShouldEqual, "1990-05-17")
		})

		Convey("When a different layout is given", func() {
			d := pgtype.Date{Time: time.Date(1990, 5, 17, 0, 0, 0, 0, time.UTC), Valid: true}
			got := convert.FormatNullableDate(d, "02/01/2006")
			So(got, ShouldNotBeNil)
			So(*got, ShouldEqual, "17/05/1990")
		})
	})
}
