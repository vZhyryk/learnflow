package worker

import (
	"learnflow_backend/internal/infrastructure/db"
	"learnflow_backend/internal/infrastructure/logger"
	"time"
)

const announcementCleanUpQuery = `
	DELETE FROM announcement_email_deliveries
	WHERE id IN (
        SELECT id FROM announcement_email_deliveries
        WHERE status IN ('sent', 'failed') AND updated_at < now() - interval '7 days'
        LIMIT 1000
	)
`

// AnnouncementCleanUpWorker is a phantom type parameterizing CleanupWorker[T] for
// announcement_email_deliveries, mirroring OutboxCleanupWorker for event_outbox.
type AnnouncementCleanUpWorker string

// NewAnnouncementCleanUpWorker returns a CleanupWorker that deletes old
// sent/failed announcement_email_deliveries rows on pollInterval.
func NewAnnouncementCleanUpWorker(queryRunner db.QueryRunner, jsonLogger *logger.Logger, pollInterval time.Duration) *CleanupWorker[AnnouncementCleanUpWorker] {
	return NewCleanupWorker[AnnouncementCleanUpWorker](queryRunner, jsonLogger, pollInterval, announcementCleanUpQuery, "AnnouncementCleanupWorker")
}
