package authhttp_test

import (
	"context"
	"encoding/json"
	"errors"
	authdomain "learnflow_backend/internal/auth/domain"
	appcontext "learnflow_backend/internal/shared/context"
	"net/http"
	"net/http/httptest"
	"testing"
)

// errWriter is an http.ResponseWriter whose Write always fails — used to
// exercise the "response write failed" branches after a handler has already
// decided what to respond with.
type errWriter struct {
	httptest.ResponseRecorder
}

func (e *errWriter) Write(_ []byte) (int, error) {
	return 0, errors.New("write failed")
}

type mockService struct {
	login                 func(ctx context.Context, req authdomain.LoginRequest) (*authdomain.AuthTokens, error)
	logout                func(ctx context.Context, req authdomain.LogoutRequest) (string, error)
	register              func(ctx context.Context, req authdomain.RegisterRequest) (string, error)
	refresh               func(ctx context.Context, req authdomain.RefreshRequest) (*authdomain.AuthTokens, error)
	verifyEmail           func(ctx context.Context, req authdomain.VerifyEmailRequest) (string, error)
	changePassword        func(ctx context.Context, req authdomain.ChangePasswordRequest) error
	initiatePasswordReset func(ctx context.Context, req authdomain.RequestPasswordResetRequest) error
	resetPassword         func(ctx context.Context, req authdomain.ResetPasswordRequest) error
	initiateEmailChange   func(ctx context.Context, req authdomain.RequestEmailChangeRequest) error
	changeEmail           func(ctx context.Context, req authdomain.EmailChangeRequest) error
	recoverAccount        func(ctx context.Context, req authdomain.RecoverAccountRequest) error
	initRecoverAccount    func(ctx context.Context, req authdomain.RequestRecoverAccountRequest) error
}

func (m *mockService) Login(ctx context.Context, req authdomain.LoginRequest) (*authdomain.AuthTokens, error) {
	return m.login(ctx, req)
}
func (m *mockService) Logout(ctx context.Context, req authdomain.LogoutRequest) (string, error) {
	return m.logout(ctx, req)
}
func (m *mockService) Register(ctx context.Context, req authdomain.RegisterRequest) (string, error) {
	return m.register(ctx, req)
}
func (m *mockService) Refresh(ctx context.Context, req authdomain.RefreshRequest) (*authdomain.AuthTokens, error) {
	return m.refresh(ctx, req)
}
func (m *mockService) VerifyEmail(ctx context.Context, req authdomain.VerifyEmailRequest) (string, error) {
	return m.verifyEmail(ctx, req)
}
func (m *mockService) ChangePassword(ctx context.Context, req authdomain.ChangePasswordRequest) error {
	return m.changePassword(ctx, req)
}
func (m *mockService) InitiatePasswordReset(ctx context.Context, req authdomain.RequestPasswordResetRequest) error {
	return m.initiatePasswordReset(ctx, req)
}
func (m *mockService) ResetPassword(ctx context.Context, req authdomain.ResetPasswordRequest) error {
	return m.resetPassword(ctx, req)
}
func (m *mockService) InitiateEmailChange(ctx context.Context, req authdomain.RequestEmailChangeRequest) error {
	return m.initiateEmailChange(ctx, req)
}
func (m *mockService) ChangeEmail(ctx context.Context, req authdomain.EmailChangeRequest) error {
	return m.changeEmail(ctx, req)
}
func (m *mockService) RecoverAccount(ctx context.Context, req authdomain.RecoverAccountRequest) error {
	return m.recoverAccount(ctx, req)
}
func (m *mockService) InitRecoverAccount(ctx context.Context, req authdomain.RequestRecoverAccountRequest) error {
	return m.initRecoverAccount(ctx, req)
}

func decodeBody(t *testing.T, body []byte) map[string]any {
	t.Helper()
	var m map[string]any
	if err := json.Unmarshal(body, &m); err != nil {
		t.Fatalf("failed to decode response body: %v", err)
	}
	return m
}

func withUser(r *http.Request) *http.Request {
	user := &authdomain.User{ID: "user-123"}
	return r.WithContext(appcontext.WithUser(r.Context(), user))
}
