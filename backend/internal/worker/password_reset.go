package worker

import (
	"fmt"
	"learnflow_backend/internal/events"
	"learnflow_backend/internal/infrastructure/db"
	"learnflow_backend/internal/infrastructure/logger"
	"learnflow_backend/internal/shared/mailer"

	"github.com/redis/go-redis/v9"
)

// NewPasswordResetWorker returns an EmailWorker configured to handle password reset events.
func NewPasswordResetWorker(
	queryRunner db.QueryRunner,
	redisClient *redis.Client,
	jsonLogger *logger.Logger,
	m *mailer.Mailer,
	baseURL string,
) *EmailWorker[events.InitPasswordResetToken] {
	return NewEmailWorker(queryRunner, redisClient, jsonLogger, m, baseURL, Config[events.InitPasswordResetToken]{
		EventType:       string(events.EventPasswordReset),
		AggregationType: string(events.AggregationTypePassword),
		IdempotencyKey:  GeneratePasswordResetIdempotencyKey,
		Validate:        ValidatePasswordResetPayload,
		Process:         HandlePasswordResetProcess,
	})
}

// ValidatePasswordResetPayload checks that all required fields are present in the payload.
func ValidatePasswordResetPayload(p events.InitPasswordResetToken) error {
	if p.UserID == "" || p.Email == "" || p.RawToken == "" {
		return fmt.Errorf("password_reset: invalid payload: missing fields")
	}
	return nil
}

// HandlePasswordResetProcess sends the password reset email for the given payload.
func HandlePasswordResetProcess(p events.InitPasswordResetToken, baseURL string, m *mailer.Mailer) error {
	data := map[string]string{
		"name":           p.UserName,
		"resetUrl":       fmt.Sprintf("%s/api/v1/users/auth/password/reset?token=%s", baseURL, p.RawToken),
		"expirationTime": p.ExpiresAt.UTC().Format("2 Jan 2006, 15:04 UTC"),
	}
	return m.Send("password_reset.html", data, mailer.CCuser{Mail: p.Email}, nil)
}

// GeneratePasswordResetIdempotencyKey returns a Redis key used to deduplicate password reset processing.
func GeneratePasswordResetIdempotencyKey(p events.InitPasswordResetToken) string {
	return fmt.Sprintf("processed:password_reset:%s:%s", p.UserID, p.RawToken)
}
