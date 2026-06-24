package worker

import (
	"fmt"
	"learnflow_backend/internal/events"
	"learnflow_backend/internal/infrastructure/db"
	"learnflow_backend/internal/infrastructure/logger"
	"learnflow_backend/internal/shared/mailer"

	"github.com/redis/go-redis/v9"
)

func NewAccountRecoveryWorker(
	queryRunner db.QueryRunner,
	redisClient *redis.Client,
	jsonLogger *logger.Logger,
	m *mailer.Mailer,
	baseUrl string,
) *EmailWorker[events.InitAccountRecoveryToken] {
	return NewEmailWorker(queryRunner, redisClient, jsonLogger, m, baseUrl, WorkerConfig[events.InitAccountRecoveryToken]{
		EventType:       string(events.EventAccountRecovery),
		AggregationType: string(events.AggregationTypeAccount),
		IdempotencyKey:  GenerateInitAccountRecoveryIdempotencyKey,
		Validate:        ValidateAccountRecoveryPayload,
		Process:         HandleInitAccountRecoveryProcess,
	})
}

func ValidateAccountRecoveryPayload(p events.InitAccountRecoveryToken) error {
	if p.UserID == "" || p.Email == "" || p.RawToken == "" {
		return fmt.Errorf("accountRecovery: invalid payload: missing fields")
	}
	return nil
}

func HandleInitAccountRecoveryProcess(p events.InitAccountRecoveryToken, baseUrl string, m *mailer.Mailer) error {
	data := map[string]string{
		"name":           p.UserName,
		"recoveryUrl":    fmt.Sprintf("%s/api/v1/users/auth/account/recover?token=%s", baseUrl, p.RawToken),
		"expirationTime": p.ExpiresAt.UTC().Format("2 Jan 2006, 15:04 UTC"),
	}

	return m.Send("account_recovery.html", data, mailer.CCuser{Mail: p.Email}, nil)
}

func GenerateInitAccountRecoveryIdempotencyKey(p events.InitAccountRecoveryToken) string {
	return fmt.Sprintf("processed:account_recovery:%s:%s", p.UserID, p.RawToken)
}
