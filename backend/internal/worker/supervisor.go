package worker

import (
	"context"
	"fmt"
	"runtime/debug"
	"strings"
	"time"

	"learnflow_backend/internal/events"
	"learnflow_backend/internal/infrastructure/logger"
	"learnflow_backend/internal/shared/mailer"
)

// restartBackoff is the pause before restarting w.Run after it returns or panics —
// intentionally no max-retry cap, since a worker that exits forever on a transient error is worse.
const restartBackoff = time.Second

// RunWithRecovery runs w.Run in a loop, recovering/logging panics and restarting after
// restartBackoff, until ctx is done. Callers own their own WaitGroup bookkeeping.
func RunWithRecovery(ctx context.Context, log *logger.Logger, w Worker) {
	for {
		func() {
			defer handleRecover(log)
			w.Run(ctx)
		}()
		if ctx.Err() != nil {
			return
		}
		time.Sleep(restartBackoff)
	}
}

func handleRecover(log *logger.Logger) {
	if r := recover(); r != nil {
		log.Error(fmt.Errorf("worker panic: %v", r), map[string]any{
			"stack": string(debug.Stack()),
		})
	}
}

// requiredFielder is a payload that lists its mandatory fields and the parts of its idempotency key.
type requiredFielder interface {
	RequiredFields() []events.Field
	GetIdempotencyKey() []string
	GetEmail() string
}

// SchemaField maps one template variable to a value computed from the payload.
type SchemaField[T any] struct {
	Name  string
	Value func(p T, baseURL string) string
}

// ValidatePayload returns an error naming every empty required field of p, prefixed with event.
func ValidatePayload[T requiredFielder](event string, p T) error {
	var missing []string
	for _, f := range p.RequiredFields() {
		if !f.IsValid() {
			missing = append(missing, f.Name)
		}
	}
	if len(missing) > 0 {
		return fmt.Errorf("%s: invalid payload: missing fields: %v", event, missing)
	}
	return nil
}

// GenerateIdempotencyKey builds the Redis dedupe key: processed:{event}:{key parts joined by ':'}.
func GenerateIdempotencyKey[T requiredFielder](event string, p T) string {
	val := strings.Builder{}
	fmt.Fprintf(&val, "processed:%s", event)
	for _, key := range p.GetIdempotencyKey() {
		fmt.Fprintf(&val, ":%s", key)
	}

	return val.String()
}

// HandleProcess renders template with the values of fields and sends it to the payload's recipient.
func HandleProcess[T requiredFielder](p T, m Mailer, fields []SchemaField[T], baseURL, template string) error {
	data := make(map[string]string, len(fields))
	for _, f := range fields {
		data[f.Name] = f.Value(p, baseURL)
	}

	return m.Send(template, data, mailer.CCUser{Mail: p.GetEmail()}, nil)
}
