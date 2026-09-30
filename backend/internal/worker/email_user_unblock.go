package worker

import (
	"learnflow_backend/internal/events"
	"learnflow_backend/internal/infrastructure/db"
	"learnflow_backend/internal/infrastructure/logger"

	"github.com/redis/go-redis/v9"
)

// NewEmailUserUnBlockWorker returns an EmailWorker that notifies a user their account was unblocked.
func NewEmailUserUnBlockWorker(
	queryRunner db.QueryRunner,
	redisClient *redis.Client,
	jsonLogger *logger.Logger,
	m Mailer,
	baseURL string,
) *EmailWorker[events.UserNotificationPayload] {
	return NewEmailWorker(queryRunner, redisClient, jsonLogger, m, baseURL, Config[events.UserNotificationPayload]{
		EventType:       string(events.EventUserUnBlocked),
		AggregationType: string(events.AggregationTypeUser),
		Name:            "email_user_unblock",
		Schema:          "email_user_unblock.html",
		SchemaFields:    userNotificationFields(),
	})
}
