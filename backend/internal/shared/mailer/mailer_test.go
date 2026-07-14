package mailer_test

import (
	"testing"

	"learnflow_backend/internal/shared/mailer"

	. "github.com/smartystreets/goconvey/convey"
)

func newMailer() *mailer.Mailer {
	return mailer.New(587, "localhost", "user", "pass", "noreply@example.com")
}

func TestSendValidatesRecipient(t *testing.T) {
	Convey("Mailer.Send recipient validation", t, func() {
		m := newMailer()

		Convey("empty email returns error", func() {
			err := m.Send("email_verification.html", nil, mailer.CCUser{Mail: "", Username: "Alice"}, nil)
			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldContainSubstring, "invalid recipient address")
		})

		Convey("malformed email returns error", func() {
			err := m.Send("email_verification.html", nil, mailer.CCUser{Mail: "notanemail", Username: "Alice"}, nil)
			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldContainSubstring, "invalid recipient address")
		})
	})
}
