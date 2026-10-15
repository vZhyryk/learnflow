package worker

import (
	"context"
	"fmt"
	"learnflow_backend/internal/events"
	"learnflow_backend/internal/shared/mailer"
	"testing"

	"github.com/jackc/pgx/v5/pgconn"
	. "github.com/smartystreets/goconvey/convey"
)

// runIdempotencyKeyTest is shared across the fan-out worker idempotency-key generator tests.
func runIdempotencyKeyTest[T any](t *testing.T, payloadType string, payload T, generate func(T) string, want string) {
	t.Helper()
	Convey(fmt.Sprintf("Given an %s payload", payloadType), t, func() {
		Convey("When generating the idempotency key", func() {
			So(generate(payload), ShouldEqual, want)
		})
	})
}

// mockMailer implements the Mailer interface via a function field.
type mockMailer struct {
	send func(templateFile string, data any, ccUser mailer.CCUser, attachmentList []string) error
}

func (m *mockMailer) Send(templateFile string, data any, ccUser mailer.CCUser, attachmentList []string) error {
	if m.send == nil {
		panic("mockMailer.send not set")
	}
	return m.send(templateFile, data, ccUser, attachmentList)
}

// mockPublisher implements events.Publisher via a function field.
type mockPublisher struct {
	publish func(ctx context.Context, eventType events.EventType, payload any) error
}

func (m *mockPublisher) Publish(ctx context.Context, eventType events.EventType, payload any) error {
	if m.publish == nil {
		panic("mockPublisher.publish not set")
	}
	return m.publish(ctx, eventType, payload)
}

// castEventType safely type-asserts a scan destination to *events.EventType.
func castEventType(v any, idx int) *events.EventType {
	s, ok := v.(*events.EventType)
	if !ok {
		panic(fmt.Sprintf("dest[%d]: expected *events.EventType, got %T", idx, v))
	}
	return s
}

func capturingExecFn(query *string, args *[]any) func(context.Context, string, ...any) (pgconn.CommandTag, error) {
	return func(_ context.Context, sql string, a ...any) (pgconn.CommandTag, error) {
		*query, *args = sql, a
		return pgconn.NewCommandTag("UPDATE 1"), nil
	}
}

// testPayload is a minimal requiredFielder for generic EmailWorker tests.
type testPayload struct {
	Value string `json:"value"`
}

func (p testPayload) RequiredFields() []events.Field {
	return []events.Field{{Name: "Value", IsValid: func() bool { return p.Value != "" }}}
}

func (p testPayload) GetIdempotencyKey() []string { return []string{p.Value} }

func (p testPayload) GetEmail() string { return "user@example.com" }
