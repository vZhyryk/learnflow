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

// GetAndParseList runs a paginated list query without extra args, scanning each row via scan and
// prefixing errors with the repository method name.
func GetAndParseList[T any](ctx context.Context,
	rep *BaseRepository,
	query, methodName string,
	params *pagination.Params,
	scan func(RowScanner) (T, error)) ([]T, error) {
	return GetAndParseListWithArgs(ctx, rep, query, methodName, params, scan, nil)
}

// GetAndParseListWithArgs runs a list query with leading args; when params is set, limit and offset are appended
// after them. args is never modified.
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

// GetCountAndParseListWithArgs is GetAndParseListWithArgs plus the total row count: countQuery runs first with args
// only (no limit/offset), then the list query runs with limit and offset appended.
func GetCountAndParseListWithArgs[T any](ctx context.Context,
	rep *BaseRepository,
	query, countQuery, methodName string,
	params pagination.Params,
	scan func(RowScanner) (T, error),
	args []any,
) (items []T, total int, err error) {
	if err = rep.QueryRunner(ctx).QueryRow(ctx, countQuery, args...).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("repository.%s count: %w", methodName, err)
	}

	if items, err = GetAndParseListWithArgs(ctx, rep, query, methodName, &params, scan, args); err != nil {
		return nil, 0, err
	}

	return items, total, nil
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
