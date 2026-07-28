package worker

import (
	"learnflow_backend/internal/events"
	"learnflow_backend/internal/shared/mailer"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestValidateInitEmailChangePayload(t *testing.T) {
	runValidatePayloadTest(t, "InitEmailChangeToken",
		events.InitEmailChangeToken{UserID: "user-123", Email: "user@example.com", RawToken: "token"},
		events.InitEmailChangeToken{UserID: "user-123", Email: "user@example.com"},
		ValidateInitEmailChangePayload,
	)
}

func TestGenerateInitEmailChangeIdempotencyKey(t *testing.T) {
	runIdempotencyKeyTest(t, "InitEmailChangeToken",
		events.InitEmailChangeToken{UserID: "user-123", RawToken: "token"},
		GenerateInitEmailChangeIdempotencyKey,
		"processed:email_change:user-123:token",
	)
}

func TestHandleInitEmailChangeProcess(t *testing.T) {
	Convey("Given an InitEmailChangeToken payload", t, func() {
		payload := events.InitEmailChangeToken{UserID: "user-123", UserName: "John Doe", Email: "user@example.com"}
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
