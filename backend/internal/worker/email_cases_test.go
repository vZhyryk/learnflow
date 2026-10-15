package worker

import (
	"learnflow_backend/internal/shared/mailer"
	"os"
	"path/filepath"
	"regexp"
	"testing"
	"time"

	. "github.com/smartystreets/goconvey/convey"
)

const (
	testBaseURL = "https://example.com"
	// tokenKeyPart is hex(sha256("token")): the idempotency key carries the token hash, never the raw token.
	tokenKeyPart = "3c469e9d6c5875d37a43f353d4f88e61fcf812c66eee3457465a40b0da4153e0"
)

var (
	templateVarRe = regexp.MustCompile(`\{\{\s*\.(\w+)\s*\}\}`)
	testExpiresAt = time.Date(2026, 1, 2, 15, 4, 0, 0, time.UTC)
)

// emailWorkerCase describes what one EmailWorker must validate, dedupe and send.
type emailWorkerCase[T requiredFielder] struct {
	worker      *EmailWorker[T]
	valid       T
	invalid     T
	wantMissing []string
	wantKey     string
	wantData    map[string]string
	wantTo      string
}

// runEmailWorkerCase covers validation, idempotency key, template data and template-variable coverage.
func runEmailWorkerCase[T requiredFielder](t *testing.T, name string, c emailWorkerCase[T]) {
	t.Helper()
	Convey("Given the "+name+" worker", t, func() {
		cfg := c.worker.cfg

		Convey("When the payload is complete, validation passes", func() {
			So(ValidatePayload(cfg.Name, c.valid), ShouldBeNil)
		})

		Convey("When the payload is incomplete, validation names the missing fields", func() {
			err := ValidatePayload(cfg.Name, c.invalid)

			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldStartWith, cfg.Name+": invalid payload")
			for _, field := range c.wantMissing {
				So(err.Error(), ShouldContainSubstring, field)
			}
		})

		Convey("When generating the idempotency key", func() {
			So(GenerateIdempotencyKey(cfg.Name, c.valid), ShouldEqual, c.wantKey)
		})

		Convey("When sending, the template gets the mapped data and the recipient", func() {
			var gotTemplate, gotTo string
			var gotData map[string]string
			m := &mockMailer{send: func(tmpl string, data any, cc mailer.CCUser, _ []string) error {
				dataMap, ok := data.(map[string]string)
				So(ok, ShouldBeTrue)
				gotTemplate, gotData, gotTo = tmpl, dataMap, cc.Mail
				return nil
			}}

			err := HandleProcess(c.valid, m, cfg.SchemaFields, testBaseURL, cfg.Schema)

			So(err, ShouldBeNil)
			So(gotTemplate, ShouldEqual, cfg.Schema)
			So(gotData, ShouldResemble, c.wantData)
			So(gotTo, ShouldEqual, c.wantTo)

			raw, readErr := os.ReadFile(filepath.Join("..", "shared", "mailer", "templates", cfg.Schema))
			So(readErr, ShouldBeNil)
			for _, match := range templateVarRe.FindAllStringSubmatch(string(raw), -1) {
				So(gotData, ShouldContainKey, match[1])
			}
		})

		Convey("When the mailer fails, the error is returned", func() {
			m := &mockMailer{send: func(_ string, _ any, _ mailer.CCUser, _ []string) error {
				return errSendFailed
			}}

			err := HandleProcess(c.valid, m, cfg.SchemaFields, testBaseURL, cfg.Schema)

			So(err, ShouldEqual, errSendFailed)
		})
	})
}
