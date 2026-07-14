package worker

import (
	"learnflow_backend/internal/events"
	"learnflow_backend/internal/shared/mailer"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestValidatePasswordResetPayload(t *testing.T) {
	Convey("Given an InitPasswordResetToken payload", t, func() {
		Convey("When all required fields are present", func() {
			payload := events.InitPasswordResetToken{
				UserID:   "user-123",
				Email:    "user@example.com",
				RawToken: "token",
			}

			err := ValidatePasswordResetPayload(payload)

			So(err, ShouldBeNil)
		})

		Convey("When required fields are missing", func() {
			payload := events.InitPasswordResetToken{
				UserID: "user-123",
				Email:  "user@example.com",
			}

			err := ValidatePasswordResetPayload(payload)

			So(err, ShouldNotBeNil)
		})
	})
}

func TestGeneratePasswordResetIdempotencyKey(t *testing.T) {
	Convey("Given an InitPasswordResetToken payload", t, func() {
		payload := events.InitPasswordResetToken{
			UserID:   "user-123",
			RawToken: "token",
		}

		Convey("When generating the idempotency key", func() {
			key := GeneratePasswordResetIdempotencyKey(payload)
			So(key, ShouldEqual, "processed:password_reset:user-123:token")
		})
	})
}

func TestHandlePasswordResetProcess(t *testing.T) {
	Convey("Given an InitPasswordResetToken payload", t, func() {
		payload := events.InitPasswordResetToken{
			UserID:   "user-123",
			UserName: "John Doe",
			Email:    "user@example.com",
		}
		baseURL := "https://example.com"

		Convey("When sending the password reset email", func() {
			m := &mockMailer{
				send: func(templateFile string, data any, ccUser mailer.CCUser, _ []string) error {
					So(templateFile, ShouldEqual, "password_reset.html")
					dataMap, ok := data.(map[string]string)
					So(ok, ShouldBeTrue)
					So(dataMap["name"], ShouldEqual, "John Doe")
					So(dataMap["resetUrl"], ShouldStartWith, baseURL)
					So(ccUser.Mail, ShouldEqual, "user@example.com")
					return nil
				},
			}
			err := HandlePasswordResetProcess(payload, baseURL, m)
			So(err, ShouldBeNil)
		})
	})
}
