package worker

import (
	"learnflow_backend/internal/events"
	"learnflow_backend/internal/infrastructure/db"
	"learnflow_backend/internal/infrastructure/logger"

	"github.com/redis/go-redis/v9"
)

// NewEmailUserBlockWorker returns an EmailWorker that notifies a user their account was blocked.
func NewEmailUserBlockWorker(
	queryRunner db.QueryRunner,
	redisClient *redis.Client,
	jsonLogger *logger.Logger,
	m Mailer,
	baseURL string,
) *EmailWorker[events.UserNotificationPayload] {
	return NewEmailWorker(queryRunner, redisClient, jsonLogger, m, baseURL, Config[events.UserNotificationPayload]{
		EventType:       string(events.EventUserBlocked),
		AggregationType: string(events.AggregationTypeUser),
		Name:            "email_user_block",
		Schema:          "email_user_block.html",
		SchemaFields:    userNotificationFields(),
	})
}
