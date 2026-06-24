package authrepository

import (
	"context"
	"errors"
	"fmt"
	authdomain "learnflow_backend/internal/auth/domain"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

// CreateUser inserts a new user and returns the generated ID.
func (rep *Repository) CreateUser(ctx context.Context, user *authdomain.User) (string, error) {
	err := rep.queryRunner(ctx).QueryRow(ctx, createUserSQL, user.Email, user.PasswordHash, user.Role).Scan(&user.ID)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return "", authdomain.ErrUserAlreadyExists
		}
		return "", fmt.Errorf("repository.CreateUser: %w", err)
	}

	return user.ID, nil
}

// CreateUserProfile inserts a profile row linked to the given user ID.
func (rep *Repository) CreateUserProfile(ctx context.Context, profile *authdomain.UserProfile) error {
	_, err := rep.queryRunner(ctx).Exec(ctx, createUserProfileSQL,
		profile.UserID,
		profile.FirstName,
		profile.LastName,
		profile.PhoneNumber,
		profile.Country,
		profile.City,
		profile.DateOfBirth,
		profile.Gender,
		profile.UiLanguage,
		profile.AvatarUrl,
		profile.Timezone,
		profile.Bio,
	)
	if err != nil {
		return fmt.Errorf("repository.CreateUserProfile: %w", err)
	}
	return nil
}

// GetDeletedUserByID returns a soft-deleted user record by its ID.
func (rep *Repository) GetDeletedUserByID(ctx context.Context, userID string) (*authdomain.User, error) {
	user, err := scanUser(rep.queryRunner(ctx).QueryRow(ctx, getDeletedUserByIDSQL, userID))
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, authdomain.ErrUserNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("repository.GetDeletedUserByID: %w", err)
	}
	return user, nil
}

// GetUserProfileByUserID returns the profile row for the given user ID.
func (rep *Repository) GetUserProfileByUserID(ctx context.Context, userID string) (*authdomain.UserProfile, error) {
	profile, err := scanUserProfile(rep.queryRunner(ctx).QueryRow(ctx, getUserProfileByUserIdSQL, userID))
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, authdomain.ErrUserNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("repository.GetUserProfileByUserID: %w", err)
	}
	return profile, nil
}

// RestoreUser clears the deleted_at timestamp for the given user ID.
func (rep *Repository) RestoreUser(ctx context.Context, userID string) error {
	tag, err := rep.queryRunner(ctx).Exec(ctx, restoreUserSQL, userID)
	if err != nil {
		return fmt.Errorf("repository.RestoreUser: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return authdomain.ErrUserNotFound
	}
	return nil
}

// GetUserByID returns a user by primary key.
func (rep *Repository) GetUserByID(ctx context.Context, userID string) (*authdomain.User, error) {
	user, err := scanUser(rep.queryRunner(ctx).QueryRow(ctx, getUserByIDSQL, userID))
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, authdomain.ErrUserNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("repository.GetUserByID: %w", err)
	}

	return user, nil
}

// GetUserByEmail returns a user by email address.
func (rep *Repository) GetUserByEmail(ctx context.Context, email string) (*authdomain.User, error) {
	user, err := scanUser(rep.queryRunner(ctx).QueryRow(ctx, getUserByEmailSQL, email))
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, authdomain.ErrUserNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("repository.GetUserByEmail: %w", err)
	}

	return user, nil
}

// UpdateStatus sets the account status for the given user.
func (rep *Repository) UpdateStatus(ctx context.Context, userID string, status authdomain.UserStatus) error {
	tag, err := rep.queryRunner(ctx).Exec(ctx, updateUserStatusSQL, status, userID)
	if err != nil {
		return fmt.Errorf("repository.UpdateStatus: %w", err)
	}

	if tag.RowsAffected() == 0 {
		return authdomain.ErrUserNotFound
	}

	return nil
}

// UpdateRole sets the role for the given user.
func (rep *Repository) UpdateRole(ctx context.Context, userID string, role authdomain.UserRole) error {
	tag, err := rep.queryRunner(ctx).Exec(ctx, updateUserRoleSQL, role, userID)
	if err != nil {
		return fmt.Errorf("repository.UpdateRole: %w", err)
	}

	if tag.RowsAffected() == 0 {
		return authdomain.ErrUserNotFound
	}

	return nil
}

// UpdateLastLoginAt records the current time as last_login_at for the given user.
func (rep *Repository) UpdateLastLoginAt(ctx context.Context, userID string) error {
	tag, err := rep.queryRunner(ctx).Exec(ctx, updateLastLoginSQL, userID)
	if err != nil {
		return fmt.Errorf("repository.UpdateLastLoginAt: %w", err)
	}

	if tag.RowsAffected() == 0 {
		return authdomain.ErrUserNotFound
	}
	return nil
}

// UpdatePasswordHash replaces the stored password hash for the given user.
func (rep *Repository) UpdatePasswordHash(ctx context.Context, userID, passwordHash string) error {
	tag, err := rep.queryRunner(ctx).Exec(ctx, updatePasswordSQL, passwordHash, userID)
	if err != nil {
		return fmt.Errorf("repository.UpdatePasswordHash: %w", err)
	}

	if tag.RowsAffected() == 0 {
		return authdomain.ErrUserNotFound
	}
	return nil
}

// UpdateEmail changes the email address for the given user.
func (rep *Repository) UpdateEmail(ctx context.Context, userID, newEmail string) error {
	tag, err := rep.queryRunner(ctx).Exec(ctx, updateEmailSQL, newEmail, userID)
	if err != nil {
		return fmt.Errorf("repository.UpdateEmail: %w", err)
	}

	if tag.RowsAffected() == 0 {
		return authdomain.ErrUserNotFound
	}
	return nil
}

// UpdateEmailVerifiedAt marks the user's email as verified.
func (rep *Repository) UpdateEmailVerifiedAt(ctx context.Context, userID string) error {
	tag, err := rep.queryRunner(ctx).Exec(ctx, updateEmailVerifiedAtSQL, userID)
	if err != nil {
		return fmt.Errorf("repository.UpdateEmailVerifiedAt: %w", err)
	}

	if tag.RowsAffected() == 0 {
		return authdomain.ErrUserNotFound
	}

	return nil
}

// DeleteUser soft-deletes the user with the given ID.
func (rep *Repository) DeleteUser(ctx context.Context, userID string) error {
	tag, err := rep.queryRunner(ctx).Exec(ctx, deleteUserSQL, userID)
	if err != nil {
		return fmt.Errorf("repository.DeleteUser: %w", err)
	}

	if tag.RowsAffected() == 0 {
		return authdomain.ErrUserNotFound
	}
	return nil
}

// IncrementFailedLogin increments the failed login counter and locks the user after reaching the limit.
func (rep *Repository) IncrementFailedLogin(ctx context.Context, userID, lockInterval string, loginCountLimit int) error {
	tag, err := rep.queryRunner(ctx).Exec(ctx, incrementFailedLoginSQL, userID, loginCountLimit, lockInterval)
	if err != nil {
		return fmt.Errorf("repository.IncrementFailedLogin: %w", err)
	}

	if tag.RowsAffected() == 0 {
		return authdomain.ErrUserNotFound
	}

	return nil
}

// ResetFailedLogin clears the failed login counter and lock for the given user.
func (rep *Repository) ResetFailedLogin(ctx context.Context, userID string) error {
	tag, err := rep.queryRunner(ctx).Exec(ctx, resetFailedLoginSQL, userID)
	if err != nil {
		return fmt.Errorf("repository.ResetFailedLogin: %w", err)
	}

	if tag.RowsAffected() == 0 {
		return authdomain.ErrUserNotFound
	}
	return nil
}
