package worker

import (
	"learnflow_backend/internal/events"
	"learnflow_backend/internal/shared/mailer"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestValidateInitEmailChangePayload(t *testing.T) {
	Convey("Given an InitEmailChangeToken payload", t, func() {
		Convey("When all required fields are present", func() {
			payload := events.InitEmailChangeToken{
				UserID:   "user-123",
				Email:    "user@example.com",
				RawToken: "token",
			}

			err := ValidateInitEmailChangePayload(payload)

			So(err, ShouldBeNil)
		})

		Convey("When required fields are missing", func() {
			payload := events.InitEmailChangeToken{
				UserID: "user-123",
				Email:  "user@example.com",
			}

			err := ValidateInitEmailChangePayload(payload)

			So(err, ShouldNotBeNil)
		})
	})
}

func TestGenerateInitEmailChangeIdempotencyKey(t *testing.T) {
	Convey("Given an InitEmailChangeToken payload", t, func() {
		payload := events.InitEmailChangeToken{
			UserID:   "user-123",
			RawToken: "token",
		}

		Convey("When generating the idempotency key", func() {
			key := GenerateInitEmailChangeIdempotencyKey(payload)
			So(key, ShouldEqual, "processed:email_change:user-123:token")
		})
	})
}

func TestHandleInitEmailChangeProcess(t *testing.T) {
	Convey("Given an InitEmailChangeToken payload", t, func() {
		payload := events.InitEmailChangeToken{
			UserID:   "user-123",
			UserName: "John Doe",
			Email:    "user@example.com",
		}
		baseURL := "https://example.com"

		Convey("When sending the email change email", func() {
			m := &mockMailer{
				send: func(templateFile string, data any, ccUser mailer.CCUser, _ []string) error {
					So(templateFile, ShouldEqual, "email_change.html")
					dataMap, ok := data.(map[string]string)
					So(ok, ShouldBeTrue)
					So(dataMap["name"], ShouldEqual, "John Doe")
					So(dataMap["confirmationUrl"], ShouldStartWith, baseURL)
					So(ccUser.Mail, ShouldEqual, "user@example.com")
					return nil
				},
			}
			err := HandleInitEmailChangeProcess(payload, baseURL, m)
			So(err, ShouldBeNil)
		})
	})
}
