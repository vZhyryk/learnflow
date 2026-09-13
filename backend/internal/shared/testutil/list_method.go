package testutil

import (
	"context"
	"testing"

	"learnflow_backend/internal/shared/pagination"

	"github.com/jackc/pgx/v5"
	. "github.com/smartystreets/goconvey/convey"
)

// TestListMethod covers the shared shape of every GetAll*/List repository method
// (query -> scan loop -> rows.Err), per go-testing.md's coverage-per-branch rule.
//
// bind receives the MockQueryRunner for the test case and returns the bound repository
// method to exercise. makeFake builds the nth fake item (n distinguishes the two rows in
// the "returns 2 items" case — e.g. by setting a distinct ID). scanFake simulates
// rows.Scan populating an item, matching the repository's own scan function.
func TestListMethod[T any](
	t *testing.T,
	methodName string,
	bind func(*MockQueryRunner) func(context.Context, pagination.Params) ([]T, error),
	makeFake func(n int) T,
	scanFake func(T) func(dest ...any) error,
) {
	Convey("Given a repository", t, func() {
		var rows *MockRows
		var queryErr error
		call := bind(&MockQueryRunner{
			QueryFn: func(_ context.Context, _ string, _ ...any) (pgx.Rows, error) {
				return rows, queryErr
			},
		})

		Convey("When the query fails", func() {
			queryErr = ErrDBUnexpected
			_, err := call(context.Background(), pagination.NewParams(1, 20))
			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldContainSubstring, "repository."+methodName)
		})

		Convey("When a row fails to scan", func() {
			rows = &MockRows{Rows: []*MockRow{
				{ScanFn: func(_ ...any) error { return ErrDBUnexpected }},
			}}
			_, err := call(context.Background(), pagination.NewParams(1, 20))
			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldContainSubstring, "scan")
		})

		Convey("When rows.Err() reports a failure after iteration", func() {
			rows = &MockRows{RowsErr: ErrDBUnexpected}
			_, err := call(context.Background(), pagination.NewParams(1, 20))
			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldContainSubstring, "rows")
		})

		Convey("When rows return 2 items", func() {
			item1, item2 := makeFake(1), makeFake(2)
			rows = &MockRows{Rows: []*MockRow{
				{ScanFn: scanFake(item1)},
				{ScanFn: scanFake(item2)},
			}}
			got, err := call(context.Background(), pagination.NewParams(1, 20))
			So(err, ShouldBeNil)
			So(got, ShouldHaveLength, 2)
			So(got[0], ShouldResemble, item1)
			So(got[1], ShouldResemble, item2)
		})
	})
}

func TestListMethodWithStringArg[T any](
	t *testing.T,
	methodName string,
	bind func(*MockQueryRunner) func(context.Context, pagination.Params, string) ([]T, error),
	makeFake func(n int) T,
	scanFake func(T) func(dest ...any) error,
) {
	Convey("Given a repository", t, func() {
		var rows *MockRows
		var queryErr error
		id := "test_id"
		call := bind(&MockQueryRunner{
			QueryFn: func(_ context.Context, _ string, _ ...any) (pgx.Rows, error) {
				return rows, queryErr
			},
		})

		Convey("When the query fails", func() {
			queryErr = ErrDBUnexpected
			_, err := call(context.Background(), pagination.NewParams(1, 20), id)
			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldContainSubstring, "repository."+methodName)
		})

		Convey("When a row fails to scan", func() {
			rows = &MockRows{Rows: []*MockRow{
				{ScanFn: func(_ ...any) error { return ErrDBUnexpected }},
			}}
			_, err := call(context.Background(), pagination.NewParams(1, 20), id)
			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldContainSubstring, "scan")
		})

		Convey("When rows.Err() reports a failure after iteration", func() {
			rows = &MockRows{RowsErr: ErrDBUnexpected}
			_, err := call(context.Background(), pagination.NewParams(1, 20), id)
			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldContainSubstring, "rows")
		})

		Convey("When rows return 2 items", func() {
			item1, item2 := makeFake(1), makeFake(2)
			rows = &MockRows{Rows: []*MockRow{
				{ScanFn: scanFake(item1)},
				{ScanFn: scanFake(item2)},
			}}
			got, err := call(context.Background(), pagination.NewParams(1, 20), id)
			So(err, ShouldBeNil)
			So(got, ShouldHaveLength, 2)
			So(got[0], ShouldResemble, item1)
			So(got[1], ShouldResemble, item2)
		})
	})
}
