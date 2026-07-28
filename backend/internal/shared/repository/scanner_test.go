package repository

import (
	"context"
	"testing"

	"learnflow_backend/internal/shared/pagination"
	"learnflow_backend/internal/shared/testutil"

	"github.com/jackc/pgx/v5"
	. "github.com/smartystreets/goconvey/convey"
)

// fakeListItem and scanFakeListItem stand in for a real domain type + scan function
// (e.g. *coursedomain.Course + scanCourse) to exercise GetAndParseList's generic shape.
type fakeListItem struct {
	ID string
}

func scanFakeListItem(row RowScanner) (fakeListItem, error) {
	item := fakeListItem{}
	err := row.Scan(&item.ID)
	return item, err
}

func TestGetAndParseList(t *testing.T) {
	Convey("Given a BaseRepository", t, func() {
		var rows *testutil.MockRows
		var queryErr error
		rep := &BaseRepository{DB: &testutil.MockQueryRunner{
			QueryFn: func(_ context.Context, _ string, _ ...any) (pgx.Rows, error) {
				return rows, queryErr
			},
		}}

		Convey("When the query fails", func() {
			queryErr = testutil.ErrDBUnexpected
			_, err := GetAndParseList(context.Background(), rep, "SELECT id FROM items", "MyMethod", pagination.NewParams(1, 20), scanFakeListItem)
			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldContainSubstring, "repository.MyMethod")
		})

		Convey("When a row fails to scan", func() {
			rows = &testutil.MockRows{Rows: []*testutil.MockRow{
				{ScanFn: func(_ ...any) error { return testutil.ErrDBUnexpected }},
			}}
			_, err := GetAndParseList(context.Background(), rep, "SELECT id FROM items", "MyMethod", pagination.NewParams(1, 20), scanFakeListItem)
			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldContainSubstring, "scan")
		})

		Convey("When rows.Err() reports a failure after iteration", func() {
			rows = &testutil.MockRows{RowsErr: testutil.ErrDBUnexpected}
			_, err := GetAndParseList(context.Background(), rep, "SELECT id FROM items", "MyMethod", pagination.NewParams(1, 20), scanFakeListItem)
			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldContainSubstring, "rows")
		})

		Convey("When rows return 2 items", func() {
			rows = &testutil.MockRows{Rows: []*testutil.MockRow{
				{ScanFn: func(dest ...any) error {
					*testutil.CastStr(dest[0], 0) = "item-1"
					return nil
				}},
				{ScanFn: func(dest ...any) error {
					*testutil.CastStr(dest[0], 0) = "item-2"
					return nil
				}},
			}}
			got, err := GetAndParseList(context.Background(), rep, "SELECT id FROM items", "MyMethod", pagination.NewParams(1, 20), scanFakeListItem)
			So(err, ShouldBeNil)
			So(got, ShouldHaveLength, 2)
			So(got[0].ID, ShouldEqual, "item-1")
			So(got[1].ID, ShouldEqual, "item-2")
		})
	})
}
