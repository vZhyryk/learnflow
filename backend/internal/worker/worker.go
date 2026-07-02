package worker

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"learnflow_backend/internal/infrastructure/db"
	"learnflow_backend/internal/infrastructure/logger"
	"learnflow_backend/internal/shared/mailer"
)

// Transactor wraps a unit of work in a database transaction.
type Transactor interface {
	InTransaction(ctx context.Context, fn func(ctx context.Context) error) error
}

// Worker is the interface implemented by all background workers.
type Worker interface {
	Run(ctx context.Context)
}

// Mailer is the subset of *mailer.Mailer used by email workers to send templated emails.
type Mailer interface {
	Send(templateFile string, data any, ccUser mailer.CCuser, attachmentList []string) error
}

const (
	queryInsertDLQ = `
    	INSERT INTO failed_jobs (event_type, queue_name, payload_json, error_message, attempt_count)
    	VALUES ($1, $2, $3, $4, $5)
	`
)

var errAlreadyProcessed = errors.New("already processed")

// NewDLQ returns a DLQWriter backed by the given query runner.
func NewDLQ(queryRunner db.QueryRunner, jsonLogger *logger.Logger) *DLQWriter {
	return &DLQWriter{jsonLogger: jsonLogger, db: queryRunner}
}

// DLQWriter persists failed events to the dead-letter queue (failed_jobs table).
type DLQWriter struct {
	db         db.QueryRunner
	jsonLogger *logger.Logger
}

func (d *DLQWriter) Write(ctx context.Context, eventName, queueName string, payload any, processErr error, attempts int) {
	if processErr == nil {
		d.jsonLogger.Error(fmt.Errorf("DLQ Write: Process error is empty"), nil)
		return
	}

	data, err := json.Marshal(payload)
	if err != nil {
		d.jsonLogger.Error(fmt.Errorf("DLQ Write: Marshal: %w", err), nil)
		return
	}

	if _, err := d.db.Exec(ctx, queryInsertDLQ, eventName, queueName, string(data), processErr.Error(), attempts); err != nil {
		d.jsonLogger.Error(fmt.Errorf("DLQ Write: Exec: %w", err), nil)
	}
}
