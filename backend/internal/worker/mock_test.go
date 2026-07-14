package worker

import (
	"context"
	"fmt"
	"learnflow_backend/internal/events"
	"learnflow_backend/internal/shared/mailer"

	"github.com/jackc/pgx/v5/pgconn"
)

type mockMailer struct {
	send func(templateFile string, data any, ccUser mailer.CCUser, attachmentList []string) error
}

func (m *mockMailer) Send(templateFile string, data any, ccUser mailer.CCUser, attachmentList []string) error {
	if m.send != nil {
		return m.send(templateFile, data, ccUser, attachmentList)
	}
	return nil
}

// mockPublisher implements events.Publisher via a function field.
type mockPublisher struct {
	publish func(ctx context.Context, eventType events.EventType, payload any) error
}

func (m *mockPublisher) Publish(ctx context.Context, eventType events.EventType, payload any) error {
	if m.publish != nil {
		return m.publish(ctx, eventType, payload)
	}
	return nil
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
