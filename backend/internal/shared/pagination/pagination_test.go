package pagination_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"learnflow_backend/internal/shared/pagination"

	. "github.com/smartystreets/goconvey/convey"
)

func TestNewParams(t *testing.T) {
	Convey("NewParams", t, func() {
		Convey("valid page and pageSize are kept as-is", func() {
			p := pagination.NewParams(3, 50)
			So(p.Page, ShouldEqual, 3)
			So(p.PageSize, ShouldEqual, 50)
		})

		Convey("page 0 normalizes to 1", func() {
			p := pagination.NewParams(0, 20)
			So(p.Page, ShouldEqual, 1)
		})

		Convey("negative page normalizes to 1", func() {
			p := pagination.NewParams(-5, 20)
			So(p.Page, ShouldEqual, 1)
		})

		Convey("page 1 is kept as-is (boundary)", func() {
			p := pagination.NewParams(1, 20)
			So(p.Page, ShouldEqual, 1)
		})

		Convey("pageSize 0 normalizes to defaultPageSize (20)", func() {
			p := pagination.NewParams(1, 0)
			So(p.PageSize, ShouldEqual, 20)
		})

		Convey("negative pageSize normalizes to defaultPageSize (20)", func() {
			p := pagination.NewParams(1, -10)
			So(p.PageSize, ShouldEqual, 20)
		})

		Convey("pageSize at maxPageSize (100) is kept as-is (boundary)", func() {
			p := pagination.NewParams(1, 100)
			So(p.PageSize, ShouldEqual, 100)
		})

		Convey("pageSize over maxPageSize (101) is capped to 100 (boundary)", func() {
			p := pagination.NewParams(1, 101)
			So(p.PageSize, ShouldEqual, 100)
		})

		Convey("pageSize far over maxPageSize is capped to 100", func() {
			p := pagination.NewParams(1, 100000)
			So(p.PageSize, ShouldEqual, 100)
		})
	})
}

func TestParamsLimit(t *testing.T) {
	Convey("Params.Limit", t, func() {
		Convey("returns PageSize", func() {
			p := pagination.NewParams(1, 30)
			So(p.Limit(), ShouldEqual, 30)
		})
	})
}

func TestParamsOffset(t *testing.T) {
	Convey("Params.Offset", t, func() {
		Convey("page 1 has offset 0", func() {
			p := pagination.NewParams(1, 20)
			So(p.Offset(), ShouldEqual, 0)
		})

		Convey("page 2 with pageSize 20 has offset 20", func() {
			p := pagination.NewParams(2, 20)
			So(p.Offset(), ShouldEqual, 20)
		})

		Convey("page 5 with pageSize 10 has offset 40", func() {
			p := pagination.NewParams(5, 10)
			So(p.Offset(), ShouldEqual, 40)
		})
	})
}

func TestParsePaginationParams(t *testing.T) {
	Convey("ParsePaginationParams", t, func() {
		Convey("valid page and page_size are used as-is", func() {
			r := httptest.NewRequestWithContext(context.Background(), http.MethodGet, "/?page=3&page_size=50", http.NoBody)
			p := pagination.ParsePaginationParams(r)
			So(p.Page, ShouldEqual, 3)
			So(p.PageSize, ShouldEqual, 50)
		})

		Convey("missing query values fall back to NewParams defaults", func() {
			r := httptest.NewRequestWithContext(context.Background(), http.MethodGet, "/", http.NoBody)
			p := pagination.ParsePaginationParams(r)
			So(p.Page, ShouldEqual, 1)
			So(p.PageSize, ShouldEqual, 20)
		})

		Convey("non-numeric page falls back to default (page=1)", func() {
			r := httptest.NewRequestWithContext(context.Background(), http.MethodGet, "/?page=abc&page_size=20", http.NoBody)
			p := pagination.ParsePaginationParams(r)
			So(p.Page, ShouldEqual, 1)
		})

		Convey("non-numeric page_size falls back to defaultPageSize (20)", func() {
			r := httptest.NewRequestWithContext(context.Background(), http.MethodGet, "/?page=1&page_size=abc", http.NoBody)
			p := pagination.ParsePaginationParams(r)
			So(p.PageSize, ShouldEqual, 20)
		})

		Convey("negative page and page_size normalize through NewParams", func() {
			r := httptest.NewRequestWithContext(context.Background(), http.MethodGet, "/?page=-1&page_size=-10", http.NoBody)
			p := pagination.ParsePaginationParams(r)
			So(p.Page, ShouldEqual, 1)
			So(p.PageSize, ShouldEqual, 20)
		})

		Convey("page_size over maxPageSize is capped to 100", func() {
			r := httptest.NewRequestWithContext(context.Background(), http.MethodGet, "/?page=1&page_size=500", http.NoBody)
			p := pagination.ParsePaginationParams(r)
			So(p.PageSize, ShouldEqual, 100)
		})
	})
}
