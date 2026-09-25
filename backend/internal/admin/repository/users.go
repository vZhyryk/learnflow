package adminrepository

import (
	"context"
	"errors"
	"fmt"
	admindomain "learnflow_backend/internal/admin/domain"
	"learnflow_backend/internal/shared/pagination"

	"github.com/jackc/pgx/v5"
)

// GetUsersData returns a page of users including soft-deleted ones (deliberate: admin view) and the total user count.
func (rep *Repository) GetUsersData(ctx context.Context, params pagination.Params) ([]*admindomain.UserData, int, error) {
	var total int
	if err := rep.QueryRunner(ctx).QueryRow(ctx, countUsersSQL).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("repository.GetUsersData count: %w", err)
	}

	rows, err := rep.QueryRunner(ctx).Query(ctx, getUserDataSQL, params.Limit(), params.Offset())
	if err != nil {
		return nil, 0, fmt.Errorf("repository.GetUsersData query: %w", err)
	}

	defer rows.Close()

	var users []*admindomain.UserData
	for rows.Next() {
		user, err := scanUserData(rows)
		if err != nil {
			return nil, 0, fmt.Errorf("repository.GetUsersData scan: %w", err)
		}
		users = append(users, user)
	}

	if err := rows.Err(); err != nil {
		return nil, 0, fmt.Errorf("repository.GetUsersData rows: %w", err)
	}

	return users, total, nil
}

// GetUserDataByID returns one user's admin view, including soft-deleted ones (deliberate: admin view).
func (rep *Repository) GetUserDataByID(ctx context.Context, userID string) (*admindomain.UserData, error) {
	user, err := scanUserData(rep.QueryRunner(ctx).QueryRow(ctx, getUserDetailsByIDSQL, userID))
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, admindomain.ErrUserNotFound
	}

	if err != nil {
		return nil, fmt.Errorf("repository.GetUserDataByID: %w", err)
	}

	return user, nil
}

// RevokeUserRole demotes a subadmin to a regular user.
func (rep *Repository) RevokeUserRole(ctx context.Context, userID string) error {
	return rep.ChangeUserField(ctx, revokeUserRoleSQL, "RevokeUserRole", userID)
}

// AssignUserRole promotes a regular user to subadmin.
func (rep *Repository) AssignUserRole(ctx context.Context, userID string) error {
	return rep.ChangeUserField(ctx, assignUserRoleSQL, "AssignUserRole", userID)
}

// DeleteUser soft-deletes a user account.
func (rep *Repository) DeleteUser(ctx context.Context, userID string) error {
	return rep.ChangeUserField(ctx, deleteUserSQL, "DeleteUser", userID)
}

// RestoreUser reactivates a soft-deleted user account.
func (rep *Repository) RestoreUser(ctx context.Context, userID string) error {
	return rep.ChangeUserField(ctx, restoreUserSQL, "RestoreUser", userID)
}

// BlockUser blocks an active user account.
func (rep *Repository) BlockUser(ctx context.Context, userID string) error {
	return rep.ChangeUserField(ctx, blockUserSQL, "BlockUser", userID)
}

// UnBlockUser reactivates a blocked user account.
func (rep *Repository) UnBlockUser(ctx context.Context, userID string) error {
	return rep.ChangeUserField(ctx, unblockUserSQL, "UnBlockUser", userID)
}

// ChangeUserField executes a single-row account mutation. 0 rows affected maps to ErrUserNotFound
// when the user does not exist, or ErrInvalidUserState when the state/role guard in the query rejected it.
func (rep *Repository) ChangeUserField(ctx context.Context, sql, methodName, userID string) error {
	tag, err := rep.QueryRunner(ctx).Exec(ctx, sql, userID)
	if err != nil {
		return fmt.Errorf("repository.%s: %w", methodName, err)
	}

	if tag.RowsAffected() > 0 {
		return nil
	}

	var exists bool
	if err = rep.QueryRunner(ctx).QueryRow(ctx, existsUserSQL, userID).Scan(&exists); err != nil {
		return fmt.Errorf("repository.%s exists: %w", methodName, err)
	}

	if !exists {
		return admindomain.ErrUserNotFound
	}

	return admindomain.ErrInvalidUserState
}
