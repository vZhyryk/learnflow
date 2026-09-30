package worker

import (
	"learnflow_backend/internal/events"
	"learnflow_backend/internal/shared/testutil"
	"testing"
)

func TestEmailUserNotificationWorkers(t *testing.T) {
	logger := testutil.NewTestLogger()
	workers := map[string]*EmailWorker[events.UserNotificationPayload]{
		"email_user_block":   NewEmailUserBlockWorker(nil, nil, logger, nil, testBaseURL),
		"email_user_unblock": NewEmailUserUnBlockWorker(nil, nil, logger, nil, testBaseURL),
		"email_user_delete":  NewEmailUserDeleteWorker(nil, nil, logger, nil, testBaseURL),
		"email_user_restore": NewEmailUserRestoreWorker(nil, nil, logger, nil, testBaseURL),
	}

	for name, w := range workers {
		t.Run(name, func(t *testing.T) {
			runEmailWorkerCase(t, name, emailWorkerCase[events.UserNotificationPayload]{
				worker:      w,
				valid:       events.UserNotificationPayload{UserID: "user-123", Email: "user@example.com", UserName: "John", EventID: "evt-1"},
				invalid:     events.UserNotificationPayload{UserID: "user-123"},
				wantMissing: []string{"Email", "UserName", "EventID"},
				wantKey:     "processed:" + name + ":user-123:evt-1",
				wantData:    map[string]string{"name": "John"},
				wantTo:      "user@example.com",
			})
		})
	}
}
