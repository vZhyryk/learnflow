package testutil

import (
	"context"
	"testing"

	"learnflow_backend/internal/shared/pagination"
	"learnflow_backend/internal/shared/repository"
)

type fakeListItem struct{ ID int }

func scanFakeListItem(row repository.RowScanner) (fakeListItem, error) {
	var item fakeListItem
	err := row.Scan(&item.ID)
	return item, err
}

func fakeCountList(runner *MockQueryRunner) func(context.Context, pagination.Params) ([]fakeListItem, int, error) {
	repo := &repository.BaseRepository{DB: runner}
	return func(ctx context.Context, params pagination.Params) ([]fakeListItem, int, error) {
		return repository.GetCountAndParseListWithArgs(ctx, repo, "list", "count", "GetItems", params, scanFakeListItem, nil)
	}
}

func TestTestCountListMethod(t *testing.T) {
	TestCountListMethod(t, "GetItems", fakeCountList,
		func(n int) fakeListItem { return fakeListItem{ID: n} },
		func(item fakeListItem) func(dest ...any) error {
			return func(dest ...any) error {
				*CastInt(dest[0], 0) = item.ID
				return nil
			}
		},
	)
}
