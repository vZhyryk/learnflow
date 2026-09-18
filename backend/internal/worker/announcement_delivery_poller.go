package worker

import (
	"encoding/json"
	"fmt"
	"learnflow_backend/internal/events"
	"learnflow_backend/internal/infrastructure/db"
	"learnflow_backend/internal/infrastructure/logger"
	"time"
)

// AnnouncementDeliveryPoller is a phantom type parameterizing Poller[T] for
// announcement_email_deliveries — step 4 of the announcement notification flow. It reads
// pending delivery rows (joined with announcements/users/user_profiles) and publishes a
// self-contained AnnouncementDeliver payload to Redis for Worker Б (AnnouncementDeliveryWorker).
type AnnouncementDeliveryPoller struct{}

const (
	querySelectAnnouncements = `
    	SELECT
			aed.id,
			aed.user_id,
			aed.announcement_id,
			up.first_name,
			an.title,
			an.body,
			u.email
    	FROM announcement_email_deliveries aed
		LEFT JOIN user_profiles up ON up.user_id = aed.user_id
		LEFT JOIN announcements an ON an.id = aed.announcement_id
		LEFT JOIN users u ON u.id = aed.user_id
		WHERE aed.status = 'pending'
    	ORDER BY aed.created_at
		LIMIT 100
    	FOR UPDATE SKIP LOCKED
	`

	queryMarkAnnouncementSend = `
    	UPDATE announcement_email_deliveries
    	SET status = 'sent',
		updated_at = now()
    	WHERE id = $1
	`

	queryMarkAnnouncementFailed = `
    	UPDATE announcement_email_deliveries
    	SET status = 'failed',
		updated_at = now(),
		last_error = $2
    	WHERE id = $1
	`
)

// AnnouncementDeliver is the self-contained, denormalized payload the delivery poller
// publishes to Redis — it also doubles as the send-time payload for
// AnnouncementDeliveryWorker (an EmailWorker[AnnouncementDeliver] instantiation).
type AnnouncementDeliver struct {
	ID             string `json:"id"`
	UserID         string `json:"user_id"`
	AnnouncementID string `json:"announcement_id"`
	FirstName      string `json:"first_name"`
	Title          string `json:"title"`
	Body           string `json:"body"`
	Email          string `json:"email"`
}

func scanAnnouncementDeliveryPoller(row entryScanner) (PollerEntry[AnnouncementDeliveryPoller], error) {
	var entry AnnouncementDeliver
	var pollerEntry PollerEntry[AnnouncementDeliveryPoller]
	err := row.Scan(&entry.ID, &entry.UserID, &entry.AnnouncementID, &entry.FirstName, &entry.Title, &entry.Body, &entry.Email)
	if err != nil {
		return pollerEntry, err
	}

	pollerEntry.ID = entry.ID
	pollerEntry.EventType = events.EventAnnouncementDeliver

	val, err := json.Marshal(entry)
	if err != nil {
		return pollerEntry, fmt.Errorf("scanAnnouncementDeliveryPoller: marshal: %w", err)
	}

	pollerEntry.PayloadJSON = string(val)

	return pollerEntry, nil
}

// NewAnnouncementDeliveryPoller returns a Poller that reads pending
// announcement_email_deliveries rows and publishes them to Redis, polling on a ticker
// until ctx is cancelled.
func NewAnnouncementDeliveryPoller(queryRunner db.QueryRunner, publisher events.Publisher, jsonLogger *logger.Logger, transactor Transactor) *Poller[AnnouncementDeliveryPoller] {
	return NewPoller[AnnouncementDeliveryPoller](queryRunner, publisher, jsonLogger, transactor, "announcementDeliveryPoller", 5*time.Second, SQLList[AnnouncementDeliveryPoller]{
		selectSQL:      querySelectAnnouncements,
		markSuccessSQL: queryMarkAnnouncementSend,
		markFailedSQL:  queryMarkAnnouncementFailed,
		scanEntry:      scanAnnouncementDeliveryPoller,
	})
}
