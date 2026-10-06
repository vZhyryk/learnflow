package worker

import (
	"learnflow_backend/internal/events"
	"learnflow_backend/internal/infrastructure/db"
	"learnflow_backend/internal/infrastructure/logger"

	"github.com/redis/go-redis/v9"
)

// NewEmailGrantAccessWorker returns an EmailWorker configured to handle access-granted events.
func NewEmailGrantAccessWorker(
	queryRunner db.QueryRunner,
	redisClient *redis.Client,
	jsonLogger *logger.Logger,
	m Mailer,
	baseURL string,
) *EmailWorker[events.GrantAccessPayload] {
	return NewEmailWorker(queryRunner, redisClient, jsonLogger, m, baseURL, Config[events.GrantAccessPayload]{
		EventType:       string(events.EventGrantAccess),
		AggregationType: string(events.AggregationTypeUser),
		Name:            "email_grant_access",
		Schema:          "email_grant_access.html",
		SchemaFields: []SchemaField[events.GrantAccessPayload]{
			{Name: "name", Value: func(p events.GrantAccessPayload, _ string) string { return p.UserName }},
			{Name: "item_name", Value: func(p events.GrantAccessPayload, _ string) string { return p.ItemName }},
			{Name: "item_type", Value: func(p events.GrantAccessPayload, _ string) string { return p.ItemType }},
		},
	})
}
