package worker

import (
	"learnflow_backend/internal/events"
	"learnflow_backend/internal/infrastructure/db"
	"learnflow_backend/internal/infrastructure/logger"

	"github.com/redis/go-redis/v9"
)

// NewEmailChangeWorker returns an EmailWorker configured to handle email change events.
func NewEmailChangeWorker(
	queryRunner db.QueryRunner,
	redisClient *redis.Client,
	jsonLogger *logger.Logger,
	m Mailer,
	baseURL string,
) *EmailWorker[events.TokenPayload] {
	fields := append(tokenFields("confirmationUrl", "/api/v1/auth/email/change"), SchemaField[events.TokenPayload]{
		Name:  "newEmail",
		Value: func(p events.TokenPayload, _ string) string { return p.Email },
	})

	return NewEmailWorker(queryRunner, redisClient, jsonLogger, m, baseURL, Config[events.TokenPayload]{
		EventType:       string(events.EventEmailChange),
		AggregationType: string(events.AggregationTypeEmail),
		Name:            "email_change",
		Schema:          "email_change.html",
		SchemaFields:    fields,
	})
}
