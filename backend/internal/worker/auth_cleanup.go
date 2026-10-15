package worker

import (
	"fmt"
	"learnflow_backend/internal/infrastructure/db"
	"learnflow_backend/internal/infrastructure/logger"
	"time"
)

// sessionRetention keeps expired/revoked sessions long enough for refresh-token reuse detection
// (previous_refresh_hash lookups) and support questions before they are deleted.
const sessionRetention = "30 days"

// tokenTables are the single-use token tables whose expired rows are removed.
var tokenTables = []string{
	"email_verification_tokens",
	"password_reset_tokens",
	"email_change_tokens",
	"account_recovery_tokens",
}

func deleteExpiredTokensQuery(table string) string {
	return fmt.Sprintf(`
	DELETE FROM %[1]s
	WHERE id IN (
		SELECT id FROM %[1]s
		WHERE expires_at < now()
		LIMIT %[2]d
	)
`, table, cleanupBatchSize)
}

var deleteExpiredSessionsQuery = fmt.Sprintf(`
	DELETE FROM user_sessions
	WHERE id IN (
		SELECT id FROM user_sessions
		WHERE expires_at < now() - interval '%[1]s'
			OR revoked_at < now() - interval '%[1]s'
		LIMIT %[2]d
	)
`, sessionRetention, cleanupBatchSize)

// TokenCleanupWorker is a phantom type parameterizing CleanupWorker[T] for the one-time token tables.
type TokenCleanupWorker string

// NewTokenCleanupWorkers returns one CleanupWorker per token table, deleting expired rows on pollInterval.
func NewTokenCleanupWorkers(queryRunner db.QueryRunner, jsonLogger *logger.Logger, pollInterval time.Duration) []Worker {
	workers := make([]Worker, 0, len(tokenTables))
	for _, table := range tokenTables {
		workers = append(workers, NewCleanupWorker[TokenCleanupWorker](
			queryRunner, jsonLogger, pollInterval, deleteExpiredTokensQuery(table), "TokenCleanupWorker:"+table))
	}

	return workers
}

// SessionCleanupWorker is a phantom type parameterizing CleanupWorker[T] for user_sessions.
type SessionCleanupWorker string

// NewSessionCleanupWorker returns a CleanupWorker that deletes sessions expired or revoked longer than
// sessionRetention ago.
func NewSessionCleanupWorker(queryRunner db.QueryRunner, jsonLogger *logger.Logger, pollInterval time.Duration) *CleanupWorker[SessionCleanupWorker] {
	return NewCleanupWorker[SessionCleanupWorker](queryRunner, jsonLogger, pollInterval, deleteExpiredSessionsQuery, "SessionCleanupWorker")
}
