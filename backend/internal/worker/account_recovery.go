package worker

import (
	"learnflow_backend/internal/events"
	"learnflow_backend/internal/infrastructure/db"
	"learnflow_backend/internal/infrastructure/logger"

	"github.com/redis/go-redis/v9"
)

// NewAccountRecoveryWorker returns an EmailWorker configured to handle account recovery events.
func NewAccountRecoveryWorker(
	queryRunner db.QueryRunner,
	redisClient *redis.Client,
	jsonLogger *logger.Logger,
	m Mailer,
	baseURL string,
) *EmailWorker[events.TokenPayload] {
	return NewEmailWorker(queryRunner, redisClient, jsonLogger, m, baseURL, Config[events.TokenPayload]{
		EventType:       string(events.EventAccountRecovery),
		AggregationType: string(events.AggregationTypeAccount),
		Name:            "account_recovery",
		Schema:          "account_recovery.html",
		SchemaFields:    tokenFields("recoveryUrl", "/api/v1/auth/account/recover"),
	})
}
