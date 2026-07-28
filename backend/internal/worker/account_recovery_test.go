package worker

import (
	"learnflow_backend/internal/events"
	"learnflow_backend/internal/shared/mailer"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestValidateAccountRecoveryPayload(t *testing.T) {
	runValidatePayloadTest(t, "InitAccountRecoveryToken",
		events.InitAccountRecoveryToken{UserID: "user-123", Email: "user@example.com", RawToken: "token"},
		events.InitAccountRecoveryToken{UserID: "user-123", Email: "user@example.com"},
		ValidateAccountRecoveryPayload,
	)
}

func TestGenerateInitAccountRecoveryIdempotencyKey(t *testing.T) {
	runIdempotencyKeyTest(t, "InitAccountRecoveryToken",
		events.InitAccountRecoveryToken{UserID: "user-123", RawToken: "token"},
		GenerateInitAccountRecoveryIdempotencyKey,
		"processed:account_recovery:user-123:token",
	)
}

func TestHandleInitAccountRecoveryProcess(t *testing.T) {
	Convey("Given an InitAccountRecoveryToken payload", t, func() {
		payload := events.InitAccountRecoveryToken{
			UserID:   "user-123",
			UserName: "John Doe",
			Email:    "user@example.com",
		}
		baseURL := "https://example.com"

		Convey("When sending the account recovery email", func() {
			m := &mockMailer{
				send: func(templateFile string, data any, ccUser mailer.CCUser, _ []string) error {
					So(templateFile, ShouldEqual, "account_recovery.html")
					dataMap, ok := data.(map[string]string)
					So(ok, ShouldBeTrue)
					So(dataMap["name"], ShouldEqual, "John Doe")
					So(dataMap["recoveryUrl"], ShouldStartWith, baseURL)
					So(ccUser.Mail, ShouldEqual, "user@example.com")
					return nil
				},
			}
			err := HandleInitAccountRecoveryProcess(payload, baseURL, m)
			So(err, ShouldBeNil)
		})
	})
}
