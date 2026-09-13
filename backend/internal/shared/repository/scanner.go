package repository

import (
	"context"
	"fmt"
	"learnflow_backend/internal/shared/pagination"
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
	params pagination.Params,
	scan func(RowScanner) (T, error)) ([]T, error) {
	rows, err := rep.QueryRunner(ctx).Query(ctx, query, params.Limit(), params.Offset())
	if err != nil {
		return nil, fmt.Errorf("repository.%s: %w", methodName, err)
	}

	defer rows.Close()

	var items []T
	for rows.Next() {
		course, err := scan(rows)
		if err != nil {
			return nil, fmt.Errorf("repository.%s scan: %w", methodName, err)
		}
		items = append(items, course)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("repository.%s rows: %w", methodName, err)
	}

	return items, nil
}

// GetAndParseListWithArgs runs a paginated list query with extra leading args (appended before
// limit/offset), scanning each row via scan and wrapping errors with methodName.
func GetAndParseListWithArgs[T any](ctx context.Context,
	rep *BaseRepository,
	query, methodName string,
	params pagination.Params,
	scan func(RowScanner) (T, error),
	args []any,
) ([]T, error) {
	args = append(args, params.Limit(), params.Offset())
	rows, err := rep.QueryRunner(ctx).Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("repository.%s: %w", methodName, err)
	}

	defer rows.Close()

	var items []T
	for rows.Next() {
		course, err := scan(rows)
		if err != nil {
			return nil, fmt.Errorf("repository.%s scan: %w", methodName, err)
		}
		items = append(items, course)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("repository.%s rows: %w", methodName, err)
	}

	return items, nil
}
