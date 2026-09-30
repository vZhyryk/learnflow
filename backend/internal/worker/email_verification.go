package worker

import (
	"learnflow_backend/internal/events"
	"learnflow_backend/internal/infrastructure/db"
	"learnflow_backend/internal/infrastructure/logger"

	"github.com/redis/go-redis/v9"
)

// NewEmailVerificationWorker returns an EmailWorker configured to handle email verification events.
func NewEmailVerificationWorker(
	queryRunner db.QueryRunner,
	redisClient *redis.Client,
	jsonLogger *logger.Logger,
	m Mailer,
	baseURL string,
) *EmailWorker[events.TokenPayload] {
	return NewEmailWorker(queryRunner, redisClient, jsonLogger, m, baseURL, Config[events.TokenPayload]{
		EventType:       string(events.EventUserRegistered),
		AggregationType: string(events.AggregationTypeEmail),
		Name:            "email_verification",
		Schema:          "email_verification.html",
		SchemaFields:    tokenFields("verificationUrl", "/api/v1/auth/email/verify"),
	})
}
