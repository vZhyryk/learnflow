package worker

import (
	"learnflow_backend/internal/events"
	"learnflow_backend/internal/shared/mailer"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestValidateEmailVerificationPayload(t *testing.T) {
	Convey("Given an UserRegisteredPayload payload", t, func() {
		Convey("When all required fields are present", func() {
			payload := events.UserRegisteredPayload{
				UserID:   "user-123",
				Email:    "user@example.com",
				RawToken: "token",
			}

			err := ValidateEmailVerificationPayload(payload)

			So(err, ShouldBeNil)
		})

		Convey("When required fields are missing", func() {
			payload := events.UserRegisteredPayload{
				UserID: "user-123",
				Email:  "user@example.com",
			}

			err := ValidateEmailVerificationPayload(payload)

			So(err, ShouldNotBeNil)
		})
	})
}

func TestGenerateEmailVerificationIdempotencyKey(t *testing.T) {
	Convey("Given an UserRegisteredPayload payload", t, func() {
		payload := events.UserRegisteredPayload{
			UserID:   "user-123",
			RawToken: "token",
		}

		Convey("When generating the idempotency key", func() {
			key := GenerateEmailVerificationIdempotencyKey(payload)
			So(key, ShouldEqual, "processed:email_verification:user-123:token")
		})
	})
}

func TestHandleEmailVerificationProcess(t *testing.T) {
	Convey("Given an UserRegisteredPayload payload", t, func() {
		payload := events.UserRegisteredPayload{
			UserID:   "user-123",
			UserName: "John Doe",
			Email:    "user@example.com",
		}
		baseURL := "https://example.com"

		Convey("When sending the email verification email", func() {
			m := &mockMailer{
				send: func(templateFile string, data any, ccUser mailer.CCuser, _ []string) error {
					So(templateFile, ShouldEqual, "email_verification.html")
					dataMap, ok := data.(map[string]string)
					So(ok, ShouldBeTrue)
					So(dataMap["name"], ShouldEqual, "John Doe")
					So(dataMap["verificationUrl"], ShouldStartWith, baseURL)
					So(ccUser.Mail, ShouldEqual, "user@example.com")
					return nil
				},
			}
			err := HandleEmailVerificationProcess(payload, baseURL, m)
			So(err, ShouldBeNil)
		})
	})
}
