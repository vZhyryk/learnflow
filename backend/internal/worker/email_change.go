package worker

import (
	"fmt"
	"learnflow_backend/internal/events"
	"learnflow_backend/internal/infrastructure/db"
	"learnflow_backend/internal/infrastructure/logger"
	"learnflow_backend/internal/shared/mailer"

	"github.com/redis/go-redis/v9"
)

func NewEmailChangeWorker(
	queryRunner db.QueryRunner,
	redisClient *redis.Client,
	jsonLogger *logger.Logger,
	m *mailer.Mailer,
	baseUrl string,
) *EmailWorker[events.InitEmailChangeToken] {
	return NewEmailWorker(queryRunner, redisClient, jsonLogger, m, baseUrl, WorkerConfig[events.InitEmailChangeToken]{
		EventType:       string(events.EventEmailChange),
		AggregationType: string(events.AggregationTypeEmail),
		IdempotencyKey:  GenerateInitEmailChangeIdempotencyKey,
		Validate:        ValidateInitEmailChangePayload,
		Process:         HandleInitEmailChangeProcess,
	})
}

func ValidateInitEmailChangePayload(p events.InitEmailChangeToken) error {
	if p.UserID == "" || p.Email == "" || p.RawToken == "" {
		return fmt.Errorf("emailChangeWorker: invalid payload: missing fields")
	}
	return nil
}

func HandleInitEmailChangeProcess(p events.InitEmailChangeToken, baseUrl string, m *mailer.Mailer) error {
	data := map[string]string{
		"name":            p.UserName,
		"newEmail":        p.Email,
		"confirmationUrl": fmt.Sprintf("%s/api/v1/users/auth/email/change?token=%s", baseUrl, p.RawToken),
		"expirationTime":  p.ExpiresAt.UTC().Format("2 Jan 2006, 15:04 UTC"),
	}

	return m.Send("email_change.html", data, mailer.CCuser{Mail: p.Email}, nil)
}

func GenerateInitEmailChangeIdempotencyKey(p events.InitEmailChangeToken) string {
	return fmt.Sprintf("processed:email_change:%s:%s", p.UserID, p.RawToken)
}
