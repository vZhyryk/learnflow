package usershttp_test

import (
	"context"
	"encoding/json"
	"io"
	authdomain "learnflow_backend/internal/auth/domain"
	"learnflow_backend/internal/infrastructure/logger"
	"learnflow_backend/internal/infrastructure/sanitizer"
	appcontext "learnflow_backend/internal/shared/context"
	usersdomain "learnflow_backend/internal/users/domain"
	"net/http"
	"testing"

	"github.com/justinas/alice"
)

func noopChain() alice.Chain { return alice.New() }

// mockService implements usersdomain.Service via function fields.
type mockService struct {
	getUserProfile    func(ctx context.Context, userID string) (*usersdomain.UserProfile, error)
	changeUserProfile func(ctx context.Context, req usersdomain.ChangeUserProfileRequest) error
}

func (m *mockService) GetUserProfile(ctx context.Context, userID string) (*usersdomain.UserProfile, error) {
	return m.getUserProfile(ctx, userID)
}

func (m *mockService) ChangeUserProfile(ctx context.Context, req usersdomain.ChangeUserProfileRequest) error {
	return m.changeUserProfile(ctx, req)
}

func newTestLogger() *logger.Logger {
	return logger.New(io.Discard, sanitizer.NewSanitizer("***", 100, nil), logger.LevelFatal)
}

func withUser(r *http.Request) *http.Request {
	user := &authdomain.User{ID: "user-123"}
	return r.WithContext(appcontext.WithUser(r.Context(), user))
}

func decodeBody(t *testing.T, body []byte) map[string]any {
	t.Helper()
	var m map[string]any
	if err := json.Unmarshal(body, &m); err != nil {
		t.Fatalf("failed to decode response body: %v", err)
	}
	return m
}
