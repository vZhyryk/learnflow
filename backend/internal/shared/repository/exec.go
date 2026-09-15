package repository

import (
	"context"
	"fmt"
)

// ExecUpdateByID runs a single-row status-change UPDATE, wrapping errors and mapping 0 rows
// affected to notFoundErr.
func ExecUpdateByID(ctx context.Context, rep *BaseRepository, sql, methodName, itemID, userID string, notFoundErr error) error {
	tag, err := rep.QueryRunner(ctx).Exec(ctx, sql, itemID, userID)
	if err != nil {
		return fmt.Errorf("repository.%s: %w", methodName, err)
	}

	if tag.RowsAffected() == 0 {
		return notFoundErr
	}

	return nil
}
