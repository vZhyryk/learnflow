//go:build integration

package worker

import (
	"context"
	"testing"
	"time"

	"learnflow_backend/internal/infrastructure/db"
)

const insertTestAnnouncementSQL = `
	INSERT INTO announcements (title, body, created_by_user_id)
	VALUES ('Integration Test Announcement', 'body', $1)
	RETURNING id`

const insertAnnouncementDeliverySQL = `
	INSERT INTO announcement_email_deliveries (user_id, announcement_id, status, updated_at)
	VALUES ($1, $2, $3, COALESCE($4::timestamptz, now()))
	RETURNING id`

func insertTestAnnouncement(t *testing.T, q db.QueryRunner, createdByUserID string) string {
	t.Helper()

	var id string
	if err := q.QueryRow(context.Background(), insertTestAnnouncementSQL, createdByUserID).Scan(&id); err != nil {
		t.Fatalf("insertTestAnnouncement: %v", err)
	}
	return id
}

// insertAnnouncementDelivery inserts a delivery row; a nil updatedAt keeps the column default.
func insertAnnouncementDelivery(t *testing.T, q db.QueryRunner, announcementID, userID, status string, updatedAt *time.Time) string {
	t.Helper()

	var id string
	if err := q.QueryRow(context.Background(), insertAnnouncementDeliverySQL, userID, announcementID, status, updatedAt).Scan(&id); err != nil {
		t.Fatalf("insertAnnouncementDelivery: %v", err)
	}
	return id
}
