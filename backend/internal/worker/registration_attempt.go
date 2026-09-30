package worker

import (
	"learnflow_backend/internal/events"
	"learnflow_backend/internal/infrastructure/db"
	"learnflow_backend/internal/infrastructure/logger"

	"github.com/redis/go-redis/v9"
)

// NewRegistrationAttemptsWorker returns an EmailWorker configured to handle registration attempt events.
func NewRegistrationAttemptsWorker(
	queryRunner db.QueryRunner,
	redisClient *redis.Client,
	jsonLogger *logger.Logger,
	m Mailer,
	baseURL string,
) *EmailWorker[events.RegistrationAttemptPayload] {
	return NewEmailWorker(queryRunner, redisClient, jsonLogger, m, baseURL, Config[events.RegistrationAttemptPayload]{
		EventType:       string(events.EventRegistrationAttemptOnExistingEmail),
		AggregationType: string(events.AggregationTypeUser),
		Name:            "registration_attempt",
		Schema:          "registration_attempt.html",
		SchemaFields: []SchemaField[events.RegistrationAttemptPayload]{
			{Name: "email", Value: func(p events.RegistrationAttemptPayload, _ string) string { return p.Email }},
			{Name: "name", Value: func(p events.RegistrationAttemptPayload, _ string) string { return p.UserName }},
		},
	})
}
