package worker

import (
	"fmt"
	"learnflow_backend/internal/events"
	"learnflow_backend/internal/infrastructure/db"
	"learnflow_backend/internal/infrastructure/logger"
	"learnflow_backend/internal/shared/mailer"

	"github.com/redis/go-redis/v9"
)

func NewEmailVerificationWorker(
	queryRunner db.QueryRunner,
	redisClient *redis.Client,
	jsonLogger *logger.Logger,
	m *mailer.Mailer,
	baseUrl string,
) *EmailWorker[events.UserRegisteredPayload] {
	return NewEmailWorker(queryRunner, redisClient, jsonLogger, m, baseUrl, WorkerConfig[events.UserRegisteredPayload]{
		EventType:       string(events.EventUserRegistered),
		AggregationType: string(events.AggregationTypeEmail),
		IdempotencyKey:  GenerateEmailVerificationIdempotencyKey,
		Validate:        ValidateEmailVerificationPayload,
		Process:         HandleEmailVerificationProcess,
	})
}

func ValidateEmailVerificationPayload(p events.UserRegisteredPayload) error {
	if p.UserID == "" || p.Email == "" || p.RawToken == "" {
		return fmt.Errorf("email_verification: invalid payload: missing fields")
	}
	return nil
}

func HandleEmailVerificationProcess(p events.UserRegisteredPayload, baseUrl string, m *mailer.Mailer) error {
	data := map[string]string{
		"name":            p.UserName,
		"verificationUrl": fmt.Sprintf("%s/api/v1/users/auth/email/verify?token=%s", baseUrl, p.RawToken),
		"expirationTime":  p.ExpiresAt.UTC().Format("2 Jan 2006, 15:04 UTC"),
	}
	return m.Send("email_verification.html", data, mailer.CCuser{Mail: p.Email}, nil)
}

func GenerateEmailVerificationIdempotencyKey(p events.UserRegisteredPayload) string {
	return fmt.Sprintf("processed:email_verification:%s:%s", p.UserID, p.RawToken)
}
