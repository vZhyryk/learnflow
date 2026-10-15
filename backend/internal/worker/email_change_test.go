package worker

import (
	"learnflow_backend/internal/events"
	"learnflow_backend/internal/shared/testutil"
	"testing"
)

func TestEmailChangeWorker(t *testing.T) {
	runEmailWorkerCase(t, "email_change", emailWorkerCase[events.TokenPayload]{
		worker: NewEmailChangeWorker(nil, nil, testutil.NewTestLogger(), nil, testBaseURL),
		valid: events.TokenPayload{
			UserID: "user-123", Email: "new@example.com", RawToken: "token", UserName: "John Doe", ExpiresAt: testExpiresAt,
		},
		invalid:     events.TokenPayload{UserID: "user-123", Email: "new@example.com"},
		wantMissing: []string{"RawToken", "UserName", "ExpiresAt"},
		wantKey:     "processed:email_change:user-123:" + tokenKeyPart,
		wantData: map[string]string{
			"name":            "John Doe",
			"newEmail":        "new@example.com",
			"confirmationUrl": testBaseURL + "/api/v1/auth/email/change?token=token",
			"expirationTime":  "2 Jan 2026, 15:04 UTC",
		},
		wantTo: "new@example.com",
	})
}
