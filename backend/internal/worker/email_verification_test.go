package worker

import (
	"learnflow_backend/internal/events"
	"learnflow_backend/internal/shared/testutil"
	"testing"
)

func TestEmailVerificationWorker(t *testing.T) {
	runEmailWorkerCase(t, "email_verification", emailWorkerCase[events.TokenPayload]{
		worker: NewEmailVerificationWorker(nil, nil, testutil.NewTestLogger(), nil, testBaseURL),
		valid: events.TokenPayload{
			UserID: "user-123", Email: "user@example.com", RawToken: "token", UserName: "John Doe", ExpiresAt: testExpiresAt,
		},
		invalid:     events.TokenPayload{UserID: "user-123", Email: "user@example.com"},
		wantMissing: []string{"RawToken", "UserName", "ExpiresAt"},
		wantKey:     "processed:email_verification:user-123:" + tokenKeyPart,
		wantData: map[string]string{
			"name":            "John Doe",
			"verificationUrl": testBaseURL + "/api/v1/auth/email/verify?token=token",
			"expirationTime":  "2 Jan 2026, 15:04 UTC",
		},
		wantTo: "user@example.com",
	})
}
