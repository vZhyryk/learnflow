package worker

import (
	"fmt"
	"learnflow_backend/internal/events"
	"learnflow_backend/internal/infrastructure/db"
	"learnflow_backend/internal/infrastructure/logger"
	"learnflow_backend/internal/shared/mailer"

	"github.com/redis/go-redis/v9"
)

// NewAccountRecoveryWorker returns an EmailWorker configured to handle account recovery events.
func NewAccountRecoveryWorker(
	queryRunner db.QueryRunner,
	redisClient *redis.Client,
	jsonLogger *logger.Logger,
	m Mailer,
	baseURL string,
) *EmailWorker[events.InitAccountRecoveryToken] {
	return NewEmailWorker(queryRunner, redisClient, jsonLogger, m, baseURL, Config[events.InitAccountRecoveryToken]{
		EventType:       string(events.EventAccountRecovery),
		AggregationType: string(events.AggregationTypeAccount),
		IdempotencyKey:  GenerateInitAccountRecoveryIdempotencyKey,
		Validate:        ValidateAccountRecoveryPayload,
		Process:         HandleInitAccountRecoveryProcess,
	})
}

// ValidateAccountRecoveryPayload checks that all required fields are present in the payload.
func ValidateAccountRecoveryPayload(p events.InitAccountRecoveryToken) error {
	if p.UserID == "" || p.Email == "" || p.RawToken == "" {
		return fmt.Errorf("accountRecovery: invalid payload: missing fields")
	}
	return nil
}

// HandleInitAccountRecoveryProcess sends the account recovery email for the given payload.
func HandleInitAccountRecoveryProcess(p events.InitAccountRecoveryToken, baseURL string, m Mailer) error {
	data := map[string]string{
		"name":           p.UserName,
		"recoveryUrl":    fmt.Sprintf("%s/api/v1/users/auth/account/recover?token=%s", baseURL, p.RawToken),
		"expirationTime": p.ExpiresAt.UTC().Format("2 Jan 2006, 15:04 UTC"),
	}

	return m.Send("account_recovery.html", data, mailer.CCuser{Mail: p.Email}, nil)
}

// GenerateInitAccountRecoveryIdempotencyKey returns a Redis key used to deduplicate account recovery processing.
func GenerateInitAccountRecoveryIdempotencyKey(p events.InitAccountRecoveryToken) string {
	return fmt.Sprintf("processed:account_recovery:%s:%s", p.UserID, p.RawToken)
}
