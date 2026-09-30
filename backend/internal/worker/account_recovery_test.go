package worker

import (
	"learnflow_backend/internal/events"
	"learnflow_backend/internal/shared/testutil"
	"testing"
)

func TestAccountRecoveryWorker(t *testing.T) {
	runEmailWorkerCase(t, "account_recovery", emailWorkerCase[events.TokenPayload]{
		worker: NewAccountRecoveryWorker(nil, nil, testutil.NewTestLogger(), nil, testBaseURL),
		valid: events.TokenPayload{
			UserID: "user-123", Email: "user@example.com", RawToken: "token", UserName: "John Doe", ExpiresAt: testExpiresAt,
		},
		invalid:     events.TokenPayload{UserID: "user-123", Email: "user@example.com"},
		wantMissing: []string{"RawToken", "UserName", "ExpiresAt"},
		wantKey:     "processed:account_recovery:user-123:" + tokenKeyPart,
		wantData: map[string]string{
			"name":           "John Doe",
			"recoveryUrl":    testBaseURL + "/api/v1/auth/account/recover?token=token",
			"expirationTime": "2 Jan 2026, 15:04 UTC",
		},
		wantTo: "user@example.com",
	})
}
