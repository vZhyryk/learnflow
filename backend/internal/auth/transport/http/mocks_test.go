package authhttp_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"

	authdomain "learnflow_backend/internal/auth/domain"
	authhttp "learnflow_backend/internal/auth/transport/http"
	"learnflow_backend/internal/shared/testutil"
)

type errWriter = testutil.ErrWriter

var decodeBody = testutil.DecodeBody
var withUser = testutil.WithUser

func newAuthMux(svc *mockService) *http.ServeMux {
	h := authhttp.NewHTTPHandler(svc, testutil.NewTestLogger())
	mux := http.NewServeMux()
	h.RegisterRoutes(mux, authhttp.AuthRouteChains{})
	return mux
}

// httpFixture wires a mockService-backed mux and a request builder for a single
// route, shared by every per-handler fixture in this package (loginFixture,
// registerFixture, ...). Embed it and add the handler-specific svcResult/svcErr
// fields on top.
type httpFixture struct {
	mux    *http.ServeMux
	newReq func(body string) *http.Request
}

func newHTTPFixture(svc *mockService, method, path string) *httpFixture {
	return &httpFixture{
		mux: newAuthMux(svc),
		newReq: func(body string) *http.Request {
			return httptest.NewRequestWithContext(context.Background(), method, path, strings.NewReader(body))
		},
	}
}

// doRequest fires body through f.mux and returns the recorded response.
func (f *httpFixture) doRequest(body string) *httptest.ResponseRecorder {
	return testutil.ServeHTTP(f.mux, f.newReq(body))
}

// doRequestWithWriter fires body through f.mux against an arbitrary
// http.ResponseWriter (e.g. errWriter, to exercise response-write-failure branches).
func (f *httpFixture) doRequestWithWriter(w http.ResponseWriter, body string) {
	f.mux.ServeHTTP(w, f.newReq(body))
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
