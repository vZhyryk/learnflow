package worker

import (
	"fmt"
	"learnflow_backend/internal/events"
	"learnflow_backend/internal/infrastructure/db"
	"learnflow_backend/internal/infrastructure/logger"
	"learnflow_backend/internal/shared/mailer"

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
		IdempotencyKey:  GenerateRegistrationAttemptsIdempotencyKey,
		Validate:        ValidateRegistrationAttemptsPayload,
		Process:         HandleRegistrationAttemptsProcess,
	})
}

// ValidateRegistrationAttemptsPayload checks that all required fields are present in the payload.
func ValidateRegistrationAttemptsPayload(p events.RegistrationAttemptPayload) error {
	var missing []string
	if p.UserID == "" {
		missing = append(missing, "UserID")
	}
	if p.Email == "" {
		missing = append(missing, "Email")
	}
	if len(missing) > 0 {
		return fmt.Errorf("registration_attempt: invalid payload: missing fields: %v", missing)
	}
	return nil
}

// HandleRegistrationAttemptsProcess sends the registration attempt notification email for the given payload.
func HandleRegistrationAttemptsProcess(p events.RegistrationAttemptPayload, _ string, m Mailer) error {
	data := map[string]string{
		"name": p.UserName,
	}

	return m.Send("registration_attempt.html", data, mailer.CCUser{Mail: p.Email}, nil)
}

// GenerateRegistrationAttemptsIdempotencyKey returns a Redis key used to deduplicate registration attempt processing.
func GenerateRegistrationAttemptsIdempotencyKey(p events.RegistrationAttemptPayload) string {
	return fmt.Sprintf("processed:registration_attempt:%s", p.UserID)
}
