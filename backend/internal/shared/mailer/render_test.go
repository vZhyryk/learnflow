package mailer

import (
	"context"
	"net"
	"testing"
	"time"

	. "github.com/smartystreets/goconvey/convey"
)

func TestRenderEmail(t *testing.T) {
	Convey("renderEmail", t, func() {
		Convey("valid template and data renders subject, plain and html bodies", func() {
			data := map[string]any{
				"name":            "Alice",
				"verificationUrl": "https://learnflow.dev/verify?token=abc",
				"expirationTime":  "24 hours",
			}
			subject, plainBody, htmlBody, err := renderEmail("email_verification.html", data)

			So(err, ShouldBeNil)
			So(subject, ShouldContainSubstring, "Verify your email")
			So(plainBody, ShouldContainSubstring, "Alice")
			So(htmlBody, ShouldContainSubstring, "Alice")
		})

		Convey("missing template file returns a parse error", func() {
			_, _, _, err := renderEmail("does_not_exist.html", nil)

			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldContainSubstring, "renderEmail parse template")
		})

		Convey("data incompatible with template fields returns an execute error", func() {
			data := struct{ Name string }{Name: "Alice"}
			_, _, _, err := renderEmail("email_verification.html", data)

			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldContainSubstring, "renderEmail execute")
		})
	})
}

func TestMailerSendDialFailure(t *testing.T) {
	Convey("Mailer.Send", t, func() {
		var lc net.ListenConfig
		lis, lisErr := lc.Listen(context.Background(), "tcp", "127.0.0.1:0")
		So(lisErr, ShouldBeNil)
		tcpAddr, ok := lis.Addr().(*net.TCPAddr)
		So(ok, ShouldBeTrue)
		port := tcpAddr.Port
		So(lis.Close(), ShouldBeNil)

		m := New(port, "127.0.0.1", "user", "pass", "noreply@example.com")
		m.dialer.Timeout = 2 * time.Second

		Convey("SMTP dial failure is wrapped", func() {
			err := m.Send("email_verification.html", map[string]any{
				"name":            "Alice",
				"verificationUrl": "https://learnflow.dev/verify?token=abc",
				"expirationTime":  "24 hours",
			}, CCuser{Mail: "user@example.com", Username: "Alice"}, nil)

			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldContainSubstring, "mailer.Send")
		})

		Convey("template render failure is returned before dialing", func() {
			err := m.Send("does_not_exist.html", nil, CCuser{Mail: "user@example.com", Username: "Alice"}, nil)

			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldContainSubstring, "renderEmail parse template")
		})
	})
}
