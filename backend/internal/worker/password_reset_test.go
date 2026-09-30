package worker

import (
	"learnflow_backend/internal/events"
	"learnflow_backend/internal/shared/testutil"
	"testing"
)

func TestPasswordResetWorker(t *testing.T) {
	runEmailWorkerCase(t, "password_reset", emailWorkerCase[events.TokenPayload]{
		worker: NewPasswordResetWorker(nil, nil, testutil.NewTestLogger(), nil, testBaseURL),
		valid: events.TokenPayload{
			UserID: "user-123", Email: "user@example.com", RawToken: "token", UserName: "John Doe", ExpiresAt: testExpiresAt,
		},
		invalid:     events.TokenPayload{UserID: "user-123", Email: "user@example.com"},
		wantMissing: []string{"RawToken", "UserName", "ExpiresAt"},
		wantKey:     "processed:password_reset:user-123:" + tokenKeyPart,
		wantData: map[string]string{
			"name":           "John Doe",
			"resetUrl":       testBaseURL + "/api/v1/auth/password/reset?token=token",
			"expirationTime": "2 Jan 2026, 15:04 UTC",
		},
		wantTo: "user@example.com",
	})
}
