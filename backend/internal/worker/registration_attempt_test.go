package worker

import (
	"learnflow_backend/internal/events"
	"learnflow_backend/internal/shared/mailer"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestValidateRegistrationAttemptsPayload(t *testing.T) {
	runValidatePayloadTest(t, "RegistrationAttemptPayload",
		events.RegistrationAttemptPayload{UserID: "user-123", Email: "user@example.com"},
		events.RegistrationAttemptPayload{UserID: "user-123"},
		ValidateRegistrationAttemptsPayload,
	)
}

func TestGenerateRegistrationAttemptsIdempotencyKey(t *testing.T) {
	runIdempotencyKeyTest(t, "RegistrationAttemptPayload",
		events.RegistrationAttemptPayload{UserID: "user-123", Email: "user@example.com"},
		GenerateRegistrationAttemptsIdempotencyKey,
		"processed:registration_attempt:user-123",
	)
}

func TestHandleRegistrationAttemptsProcess(t *testing.T) {
	Convey("Given an RegistrationAttemptPayload payload", t, func() {
		payload := events.RegistrationAttemptPayload{
			UserID:   "user-123",
			UserName: "John Doe",
			Email:    "user@example.com",
		}
		baseURL := "https://example.com"

		Convey("When sending the registration attempt email", func() {
			m := &mockMailer{
				send: func(templateFile string, data any, ccUser mailer.CCUser, _ []string) error {
					So(templateFile, ShouldEqual, "registration_attempt.html")
					dataMap, ok := data.(map[string]string)
					So(ok, ShouldBeTrue)
					So(dataMap["name"], ShouldEqual, "John Doe")
					So(ccUser.Mail, ShouldEqual, "user@example.com")
					return nil
				},
			}
			err := HandleRegistrationAttemptsProcess(payload, baseURL, m)
			So(err, ShouldBeNil)
		})
	})
}
