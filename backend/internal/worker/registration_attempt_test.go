package worker

import (
	"learnflow_backend/internal/events"
	"learnflow_backend/internal/shared/testutil"
	"testing"
)

func TestRegistrationAttemptWorker(t *testing.T) {
	runEmailWorkerCase(t, "registration_attempt", emailWorkerCase[events.RegistrationAttemptPayload]{
		worker:      NewRegistrationAttemptsWorker(nil, nil, testutil.NewTestLogger(), nil, testBaseURL),
		valid:       events.RegistrationAttemptPayload{UserID: "user-123", Email: "user@example.com", UserName: "John Doe"},
		invalid:     events.RegistrationAttemptPayload{UserID: "user-123"},
		wantMissing: []string{"Email", "UserName"},
		wantKey:     "processed:registration_attempt:user-123",
		wantData:    map[string]string{"name": "John Doe", "email": "user@example.com"},
		wantTo:      "user@example.com",
	})
}
