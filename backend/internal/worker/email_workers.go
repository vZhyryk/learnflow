package worker

import (
	"learnflow_backend/internal/infrastructure/db"
	"learnflow_backend/internal/infrastructure/logger"

	"github.com/redis/go-redis/v9"
)

// NewEmailWorkers returns every email-sending worker, so cmd/worker registers them in one place.
func NewEmailWorkers(
	queryRunner db.QueryRunner,
	redisClient *redis.Client,
	jsonLogger *logger.Logger,
	m Mailer,
	baseURL string,
) []Worker {
	return []Worker{
		NewEmailVerificationWorker(queryRunner, redisClient, jsonLogger, m, baseURL),
		NewEmailChangeWorker(queryRunner, redisClient, jsonLogger, m, baseURL),
		NewPasswordResetWorker(queryRunner, redisClient, jsonLogger, m, baseURL),
		NewRegistrationAttemptsWorker(queryRunner, redisClient, jsonLogger, m, baseURL),
		NewAccountRecoveryWorker(queryRunner, redisClient, jsonLogger, m, baseURL),
		NewAnnouncementDeliveryWorker(queryRunner, redisClient, jsonLogger, m, baseURL),
		NewEmailUserBlockWorker(queryRunner, redisClient, jsonLogger, m, baseURL),
		NewEmailUserUnBlockWorker(queryRunner, redisClient, jsonLogger, m, baseURL),
		NewEmailUserDeleteWorker(queryRunner, redisClient, jsonLogger, m, baseURL),
		NewEmailUserRestoreWorker(queryRunner, redisClient, jsonLogger, m, baseURL),
		NewEmailGrantAccessWorker(queryRunner, redisClient, jsonLogger, m, baseURL),
	}
}
