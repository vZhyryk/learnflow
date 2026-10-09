package repository

import (
	"context"
	"fmt"
	"learnflow_backend/internal/shared/pagination"
	"slices"

	"github.com/jackc/pgx/v5"
)

// RowScanner abstracts pgx.Row and pgx.Rows for use in repository scan functions.
type RowScanner interface {
	Scan(dest ...any) error
}

// GetAndParseList runs a paginated list query, scanning each row via scan and wrapping
// errors with methodName.
func GetAndParseList[T any](ctx context.Context,
	rep *BaseRepository,
	query, methodName string,
	params *pagination.Params,
	scan func(RowScanner) (T, error)) ([]T, error) {
	var args []any
	if params != nil {
		args = append(args, params.Limit(), params.Offset())
	}

	rows, err := rep.QueryRunner(ctx).Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("repository.%s: %w", methodName, err)
	}

	defer rows.Close()

	return CollectRows(rows, fmt.Sprintf("repository.%s", methodName), scan)
}

// GetAndParseListWithArgs runs a paginated list query with extra leading args (appended before
// limit/offset), scanning each row via scan and wrapping errors with methodName.
func GetAndParseListWithArgs[T any](ctx context.Context,
	rep *BaseRepository,
	query, methodName string,
	params *pagination.Params,
	scan func(RowScanner) (T, error),
	args []any,
) ([]T, error) {
	if params != nil {
		args = append(slices.Clone(args), params.Limit(), params.Offset())
	}

	rows, err := rep.QueryRunner(ctx).Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("repository.%s: %w", methodName, err)
	}

	defer rows.Close()

	return CollectRows(rows, fmt.Sprintf("repository.%s", methodName), scan)
}

// CollectRows scans every row via scan and returns a non-nil slice; errors are prefixed with method.
func CollectRows[T any](rows pgx.Rows, method string, scan func(RowScanner) (T, error)) ([]T, error) {
	items := make([]T, 0)
	for rows.Next() {
		item, err := scan(rows)
		if err != nil {
			return nil, fmt.Errorf("%s scan: %w", method, err)
		}
		items = append(items, item)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("%s rows: %w", method, err)
	}

	return items, nil
}
