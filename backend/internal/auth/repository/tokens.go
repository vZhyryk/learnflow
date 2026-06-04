package authrepository

import (
	"context"
	authdomain "learnflow_backend/internal/auth/domain"
)

func (rep *Repository) DeleteExpiredTokens(ctx context.Context) (int, error) {
	return 0, nil
}
func (rep *Repository) CreateEmailVerificationToken(ctx context.Context, token *authdomain.EmailVerificationToken) error {
	return nil
}
func (rep *Repository) GetEmailVerificationToken(ctx context.Context, token string) (*authdomain.EmailVerificationToken, error) {
	return nil, nil
}
func (rep *Repository) MarkEmailVerificationTokenUsed(ctx context.Context, tokenHash string) error {
	return nil
}
func (rep *Repository) CreatePasswordResetToken(ctx context.Context, token *authdomain.PasswordResetToken) error {
	return nil
}
func (rep *Repository) GetPasswordResetToken(ctx context.Context, token string) (*authdomain.PasswordResetToken, error) {
	return nil, nil
}
func (rep *Repository) MarkPasswordResetTokenUsed(ctx context.Context, tokenHash string) error {
	return nil
}
func (rep *Repository) CreateEmailChangeToken(ctx context.Context, token *authdomain.EmailChangeToken) error {
	return nil
}
func (rep *Repository) GetEmailChangeToken(ctx context.Context, token string) (*authdomain.EmailChangeToken, error) {
	return nil, nil
}
func (rep *Repository) MarkEmailChangeTokenUsed(ctx context.Context, tokenHash string) error {
	return nil
}
func (rep *Repository) CreateAccountRecoveryToken(ctx context.Context, token *authdomain.AccountRecoveryToken) error {
	return nil
}
func (rep *Repository) GetAccountRecoveryToken(ctx context.Context, token string) (*authdomain.AccountRecoveryToken, error) {
	return nil, nil
}
func (rep *Repository) MarkAccountRecoveryTokenUsed(ctx context.Context, tokenHash string) error {
	return nil
}
