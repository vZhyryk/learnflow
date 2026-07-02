package worker

import (
	"learnflow_backend/internal/events"
	"learnflow_backend/internal/shared/mailer"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestValidateRegistrationAttemptsPayload(t *testing.T) {
	Convey("Given an RegistrationAttemptPayload payload", t, func() {
		Convey("When all required fields are present", func() {
			payload := events.RegistrationAttemptPayload{
				UserID: "user-123",
				Email:  "user@example.com",
			}

			err := ValidateRegistrationAttemptsPayload(payload)

			So(err, ShouldBeNil)
		})

		Convey("When required fields are missing", func() {
			payload := events.RegistrationAttemptPayload{
				UserID: "user-123",
			}

			err := ValidateRegistrationAttemptsPayload(payload)

			So(err, ShouldNotBeNil)
		})
	})
}

func TestGenerateRegistrationAttemptsIdempotencyKey(t *testing.T) {
	Convey("Given an RegistrationAttemptPayload payload", t, func() {
		payload := events.RegistrationAttemptPayload{
			UserID: "user-123",
			Email:  "user@example.com",
		}

		Convey("When generating the idempotency key", func() {
			key := GenerateRegistrationAttemptsIdempotencyKey(payload)
			So(key, ShouldEqual, "processed:registration_attempt:user-123")
		})
	})
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
				send: func(templateFile string, data any, ccUser mailer.CCuser, _ []string) error {
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
