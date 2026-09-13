package worker

import (
	"fmt"
	"learnflow_backend/internal/events"
	"learnflow_backend/internal/infrastructure/db"
	"learnflow_backend/internal/infrastructure/logger"
	"learnflow_backend/internal/shared/mailer"

	"github.com/redis/go-redis/v9"
)

// NewEmailVerificationWorker returns an EmailWorker configured to handle email verification events.
func NewEmailVerificationWorker(
	queryRunner db.QueryRunner,
	redisClient *redis.Client,
	jsonLogger *logger.Logger,
	m Mailer,
	baseURL string,
) *EmailWorker[events.UserRegisteredPayload] {
	return NewEmailWorker(queryRunner, redisClient, jsonLogger, m, baseURL, Config[events.UserRegisteredPayload]{
		EventType:       string(events.EventUserRegistered),
		AggregationType: string(events.AggregationTypeEmail),
		IdempotencyKey:  GenerateEmailVerificationIdempotencyKey,
		Validate:        ValidateEmailVerificationPayload,
		Process:         HandleEmailVerificationProcess,
	})
}

// ValidateEmailVerificationPayload checks that all required fields are present in the payload.
func ValidateEmailVerificationPayload(p events.UserRegisteredPayload) error {
	var missing []string
	if p.UserID == "" {
		missing = append(missing, "UserID")
	}
	if p.Email == "" {
		missing = append(missing, "Email")
	}
	if p.RawToken == "" {
		missing = append(missing, "RawToken")
	}
	if len(missing) > 0 {
		return fmt.Errorf("email_verification: invalid payload: missing fields: %v", missing)
	}
	return nil
}

// HandleEmailVerificationProcess sends the email verification email for the given payload.
func HandleEmailVerificationProcess(p events.UserRegisteredPayload, baseURL string, m Mailer) error {
	data := map[string]string{
		"name":            p.UserName,
		"verificationUrl": fmt.Sprintf("%s/api/v1/auth/email/verify?token=%s", baseURL, p.RawToken),
		"expirationTime":  p.ExpiresAt.UTC().Format("2 Jan 2006, 15:04 UTC"),
	}
	return m.Send("email_verification.html", data, mailer.CCUser{Mail: p.Email}, nil)
}

// GenerateEmailVerificationIdempotencyKey returns a Redis key used to deduplicate email verification processing.
func GenerateEmailVerificationIdempotencyKey(p events.UserRegisteredPayload) string {
	return fmt.Sprintf("processed:email_verification:%s:%s", p.UserID, p.RawToken)
}
