package worker

import (
	"learnflow_backend/internal/events"
	"learnflow_backend/internal/shared/mailer"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestValidateEmailVerificationPayload(t *testing.T) {
	runValidatePayloadTest(t, "UserRegisteredPayload",
		events.UserRegisteredPayload{UserID: "user-123", Email: "user@example.com", RawToken: "token"},
		events.UserRegisteredPayload{UserID: "user-123", Email: "user@example.com"},
		ValidateEmailVerificationPayload,
	)
}

func TestGenerateEmailVerificationIdempotencyKey(t *testing.T) {
	runIdempotencyKeyTest(t, "UserRegisteredPayload",
		events.UserRegisteredPayload{UserID: "user-123", RawToken: "token"},
		GenerateEmailVerificationIdempotencyKey,
		"processed:email_verification:user-123:token",
	)
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
				send: func(templateFile string, data any, ccUser mailer.CCUser, _ []string) error {
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
