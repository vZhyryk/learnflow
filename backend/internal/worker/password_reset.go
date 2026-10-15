package worker

import (
	"learnflow_backend/internal/events"
	"learnflow_backend/internal/infrastructure/db"
	"learnflow_backend/internal/infrastructure/logger"

	"github.com/redis/go-redis/v9"
)

// NewPasswordResetWorker returns an EmailWorker configured to handle password reset events.
func NewPasswordResetWorker(
	queryRunner db.QueryRunner,
	redisClient *redis.Client,
	jsonLogger *logger.Logger,
	m Mailer,
	baseURL string,
) *EmailWorker[events.TokenPayload] {
	return NewEmailWorker(queryRunner, redisClient, jsonLogger, m, baseURL, Config[events.TokenPayload]{
		EventType:       string(events.EventPasswordReset),
		AggregationType: string(events.AggregationTypePassword),
		Name:            "password_reset",
		Schema:          "password_reset.html",
		SchemaFields:    tokenFields("resetUrl", "/api/v1/auth/password/reset"),
	})
}
