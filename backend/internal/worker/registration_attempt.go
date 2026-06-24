package worker

import (
	"fmt"
	"learnflow_backend/internal/events"
	"learnflow_backend/internal/infrastructure/db"
	"learnflow_backend/internal/infrastructure/logger"
	"learnflow_backend/internal/shared/mailer"

	"github.com/redis/go-redis/v9"
)

func NewRegistrationAttemptsWorker(
	queryRunner db.QueryRunner,
	redisClient *redis.Client,
	jsonLogger *logger.Logger,
	m *mailer.Mailer,
	baseUrl string,
) *EmailWorker[events.RegistrationAttemptPayload] {
	return NewEmailWorker(queryRunner, redisClient, jsonLogger, m, baseUrl, WorkerConfig[events.RegistrationAttemptPayload]{
		EventType:       string(events.EventEmailChange),
		AggregationType: string(events.AggregationTypeEmail),
		IdempotencyKey:  GenerateRegistrationAttemptsIdempotencyKey,
		Validate:        ValidateRegistrationAttemptsPayload,
		Process:         HandleRegistrationAttemptsProcess,
	})
}

func ValidateRegistrationAttemptsPayload(p events.RegistrationAttemptPayload) error {
	if p.UserID == "" || p.Email == "" {
		return fmt.Errorf("registrationAttempt: invalid payload: missing fields")
	}
	return nil
}

func HandleRegistrationAttemptsProcess(p events.RegistrationAttemptPayload, baseUrl string, m *mailer.Mailer) error {
	data := map[string]string{
		"name": p.UserName,
	}

	return m.Send("registration_attempt.html", data, mailer.CCuser{Mail: p.Email}, nil)
}

func GenerateRegistrationAttemptsIdempotencyKey(p events.RegistrationAttemptPayload) string {
	return fmt.Sprintf("processed:registration_attempt:%s", p.UserID)
}
