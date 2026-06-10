package authrepository

import (
	"context"
	"errors"
	"fmt"
	authdomain "learnflow_backend/internal/auth/domain"

	"github.com/jackc/pgx/v5"
)

// DeleteExpiredTokens removes expired tokens of all types and returns the total count deleted.
func (rep *Repository) DeleteExpiredTokens(ctx context.Context) (int, error) {
	queries := []string{
		deleteExpiredEmailVerificationTokensSQL,
		deleteExpiredPasswordResetTokensSQL,
		deleteExpiredEmailChangeTokensSQL,
		deleteExpiredAccountRecoveryTokensSQL,
	}
	total := 0
	for _, q := range queries {
		tag, err := rep.db.Exec(ctx, q)
		if err != nil {
			return total, fmt.Errorf("repository.DeleteExpiredTokens: %w", err)
		}
		total += int(tag.RowsAffected())
	}
	return total, nil
}

// CreateEmailVerificationToken persists a new email verification token and returns it with DB-generated fields.
func (rep *Repository) CreateEmailVerificationToken(ctx context.Context, token *authdomain.EmailVerificationToken) (*authdomain.EmailVerificationToken, error) {
	base, err := scanToken(rep.db.QueryRow(ctx, createEmailVerificationTokenSQL, token.UserID, token.TokenHash, token.ExpiresAt))
	if err != nil {
		return nil, fmt.Errorf("repository.CreateEmailVerificationToken: %w", err)
	}

	return &authdomain.EmailVerificationToken{TokenBase: *base}, nil
}

// GetEmailVerificationToken retrieves an email verification token by its hash.
func (rep *Repository) GetEmailVerificationToken(ctx context.Context, tokenHash string) (*authdomain.EmailVerificationToken, error) {
	token, err := scanToken(rep.db.QueryRow(ctx, getEmailVerificationTokenByHashSQL, tokenHash))
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, authdomain.ErrInvalidToken
	}
	if err != nil {
		return nil, fmt.Errorf("repository.GetEmailVerificationToken: %w", err)
	}

	return &authdomain.EmailVerificationToken{TokenBase: *token}, nil
}

// MarkEmailVerificationTokenUsed marks the token as used so it cannot be reused.
func (rep *Repository) MarkEmailVerificationTokenUsed(ctx context.Context, tokenHash string) error {
	tag, err := rep.db.Exec(ctx, markEmailVerificationTokenUsedSQL, tokenHash)
	if err != nil {
		return fmt.Errorf("repository.MarkEmailVerificationTokenUsed: %w", err)
	}

	if tag.RowsAffected() == 0 {
		return authdomain.ErrTokenUsed
	}

	return nil
}

// CreatePasswordResetToken persists a new password reset token and returns it with DB-generated fields.
func (rep *Repository) CreatePasswordResetToken(ctx context.Context, token *authdomain.PasswordResetToken) (*authdomain.PasswordResetToken, error) {
	base, err := scanToken(rep.db.QueryRow(ctx, createPasswordResetTokenSQL, token.UserID, token.TokenHash, token.ExpiresAt))
	if err != nil {
		return token, fmt.Errorf("repository.CreatePasswordResetToken: %w", err)
	}

	return &authdomain.PasswordResetToken{TokenBase: *base}, nil
}

// GetPasswordResetToken retrieves a password reset token by its hash.
func (rep *Repository) GetPasswordResetToken(ctx context.Context, tokenHash string) (*authdomain.PasswordResetToken, error) {
	token, err := scanToken(rep.db.QueryRow(ctx, getPasswordResetTokenByHashSQL, tokenHash))
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, authdomain.ErrInvalidToken
	}
	if err != nil {
		return nil, fmt.Errorf("repository.GetPasswordResetToken: %w", err)
	}

	return &authdomain.PasswordResetToken{TokenBase: *token}, nil
}

// MarkPasswordResetTokenUsed marks the token as used so it cannot be reused.
func (rep *Repository) MarkPasswordResetTokenUsed(ctx context.Context, tokenHash string) error {
	tag, err := rep.db.Exec(ctx, markPasswordResetTokenUsedSQL, tokenHash)
	if err != nil {
		return fmt.Errorf("repository.MarkPasswordResetTokenUsed: %w", err)
	}

	if tag.RowsAffected() == 0 {
		return authdomain.ErrTokenUsed
	}

	return nil
}

// CreateEmailChangeToken persists a new email change token and returns it with DB-generated fields.
func (rep *Repository) CreateEmailChangeToken(ctx context.Context, token *authdomain.EmailChangeToken) (*authdomain.EmailChangeToken, error) {
	token, err := scanEmailChangeToken(rep.db.QueryRow(ctx, createEmailChangeTokenSQL, token.UserID, token.TokenHash, token.ExpiresAt, token.NewEmail))
	if err != nil {
		return nil, fmt.Errorf("repository.CreateEmailChangeToken: %w", err)
	}

	return token, nil
}

// GetEmailChangeToken retrieves an email change token by its hash.
func (rep *Repository) GetEmailChangeToken(ctx context.Context, tokenHash string) (*authdomain.EmailChangeToken, error) {
	token, err := scanEmailChangeToken(rep.db.QueryRow(ctx, getEmailChangeTokenByHashSQL, tokenHash))
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, authdomain.ErrInvalidToken
	}
	if err != nil {
		return nil, fmt.Errorf("repository.GetEmailChangeToken: %w", err)
	}

	return token, nil
}

// MarkEmailChangeTokenUsed marks the token as used so it cannot be reused.
func (rep *Repository) MarkEmailChangeTokenUsed(ctx context.Context, tokenHash string) error {
	tag, err := rep.db.Exec(ctx, markEmailChangeTokenUsedSQL, tokenHash)
	if err != nil {
		return fmt.Errorf("repository.MarkEmailChangeTokenUsed: %w", err)
	}

	if tag.RowsAffected() == 0 {
		return authdomain.ErrTokenUsed
	}
	return nil
}

// CreateAccountRecoveryToken persists a new account recovery token and returns it with DB-generated fields.
func (rep *Repository) CreateAccountRecoveryToken(ctx context.Context, token *authdomain.AccountRecoveryToken) (*authdomain.AccountRecoveryToken, error) {
	base, err := scanToken(rep.db.QueryRow(ctx, createAccountRecoveryTokenSQL, token.UserID, token.TokenHash, token.ExpiresAt))
	if err != nil {
		return nil, fmt.Errorf("repository.CreateAccountRecoveryToken: %w", err)
	}

	return &authdomain.AccountRecoveryToken{TokenBase: *base}, nil
}

// GetAccountRecoveryToken retrieves an account recovery token by its hash.
func (rep *Repository) GetAccountRecoveryToken(ctx context.Context, tokenHash string) (*authdomain.AccountRecoveryToken, error) {
	token, err := scanToken(rep.db.QueryRow(ctx, getAccountRecoveryTokenByHashSQL, tokenHash))
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, authdomain.ErrInvalidToken
	}
	if err != nil {
		return nil, fmt.Errorf("repository.GetAccountRecoveryToken: %w", err)
	}

	return &authdomain.AccountRecoveryToken{TokenBase: *token}, nil
}

// MarkAccountRecoveryTokenUsed marks the token as used so it cannot be reused.
func (rep *Repository) MarkAccountRecoveryTokenUsed(ctx context.Context, tokenHash string) error {
	tag, err := rep.db.Exec(ctx, markAccountRecoveryTokenUsedSQL, tokenHash)
	if err != nil {
		return fmt.Errorf("repository.MarkAccountRecoveryTokenUsed: %w", err)
	}

	if tag.RowsAffected() == 0 {
		return authdomain.ErrTokenUsed
	}
	return nil
}
