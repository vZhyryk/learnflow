//go:build integration

package worker

import (
	"context"
	"testing"
	"time"

	"learnflow_backend/internal/infrastructure/db"
	"learnflow_backend/internal/shared/testutil"

	. "github.com/smartystreets/goconvey/convey"
)

const insertTestAnnouncementSQL = `
	INSERT INTO announcements (title, body, created_by_user_id)
	VALUES ('Integration Test Announcement', 'body', $1)
	RETURNING id`

func insertTestAnnouncement(t *testing.T, pool db.QueryRunner, createdByUserID string) string {
	t.Helper()

	var id string
	if err := pool.QueryRow(context.Background(), insertTestAnnouncementSQL, createdByUserID).Scan(&id); err != nil {
		t.Fatalf("insertTestAnnouncement: %v", err)
	}
	return id
}

// dummyUserPasswordHash mirrors testutil.dummyUserPasswordHash (unexported there) — a
// valid bcrypt hash for seeding test users, not a real credential.
const dummyUserPasswordHash = "$2a$10$N9qo8uLOickgx2ZMRZoMyeIjZAgcfl7p92ldGxad68LJZdL17lhWy"

// insertTestUserViaPool inserts a minimal active user directly against the pool — the
// shared testutil.InsertTestUser requires a pgx.Tx, which pgxpool.Pool doesn't satisfy.
func insertTestUserViaPool(t *testing.T, pool db.QueryRunner, email string) string {
	t.Helper()

	var id string
	err := pool.QueryRow(context.Background(),
		`INSERT INTO users (email, password_hash, role, status) VALUES ($1, $2, 'user', 'active') RETURNING id`,
		email, dummyUserPasswordHash,
	).Scan(&id)
	if err != nil {
		t.Fatalf("insertTestUserViaPool: %v", err)
	}
	return id
}

func insertAnnouncementDeliveryRowWithStatus(t *testing.T, pool db.QueryRunner, announcementID, userID, status string, updatedAt time.Time) string {
	t.Helper()

	var id string
	err := pool.QueryRow(context.Background(),
		`INSERT INTO announcement_email_deliveries (user_id, announcement_id, status, updated_at)
		 VALUES ($1, $2, $3, $4)
		 RETURNING id`,
		userID, announcementID, status, updatedAt,
	).Scan(&id)
	if err != nil {
		t.Fatalf("insertAnnouncementDeliveryRowWithStatus: %v", err)
	}
	return id
}

func announcementDeliveryRowExists(t *testing.T, pool db.QueryRunner, id string) bool {
	t.Helper()

	var exists bool
	if err := pool.QueryRow(context.Background(),
		`SELECT EXISTS(SELECT 1 FROM announcement_email_deliveries WHERE id = $1)`, id).Scan(&exists); err != nil {
		t.Fatalf("announcementDeliveryRowExists: %v", err)
	}
	return exists
}

func TestAnnouncementCleanUpWorkerPoll_Integration(t *testing.T) {
	pool := testutil.NewTestPool(t)
	w := NewAnnouncementCleanUpWorker(pool, testutil.NewTestLogger(), 24*time.Hour)

	Convey("Given announcement_email_deliveries rows backed by real Postgres", t, func() {
		userID := insertTestUserViaPool(t, pool, testutil.RandomTestEmail(t, "announcement-cleanup"))
		announcementID := insertTestAnnouncement(t, pool, userID)

		Convey("When a sent row is older than the 7-day retention window", func() {
			old := time.Now().Add(-8 * 24 * time.Hour)
			id := insertAnnouncementDeliveryRowWithStatus(t, pool, announcementID, userID, "sent", old)
			t.Cleanup(func() {
				pool.Exec(context.Background(), `DELETE FROM announcement_email_deliveries WHERE id = $1`, id) //nolint:errcheck // best-effort cleanup
			})

			w.poll(context.Background())

			So(announcementDeliveryRowExists(t, pool, id), ShouldBeFalse)
		})

		Convey("When a failed row is older than the 7-day retention window", func() {
			old := time.Now().Add(-8 * 24 * time.Hour)
			id := insertAnnouncementDeliveryRowWithStatus(t, pool, announcementID, userID, "failed", old)
			t.Cleanup(func() {
				pool.Exec(context.Background(), `DELETE FROM announcement_email_deliveries WHERE id = $1`, id) //nolint:errcheck // best-effort cleanup
			})

			w.poll(context.Background())

			So(announcementDeliveryRowExists(t, pool, id), ShouldBeFalse)
		})

		Convey("When a sent row is within the 7-day retention window", func() {
			recent := time.Now().Add(-1 * 24 * time.Hour)
			id := insertAnnouncementDeliveryRowWithStatus(t, pool, announcementID, userID, "sent", recent)
			t.Cleanup(func() {
				pool.Exec(context.Background(), `DELETE FROM announcement_email_deliveries WHERE id = $1`, id) //nolint:errcheck // best-effort cleanup
			})

			w.poll(context.Background())

			So(announcementDeliveryRowExists(t, pool, id), ShouldBeTrue)
		})

		Convey("When a pending row is old, poll leaves it untouched (only sent/failed are retention-eligible)", func() {
			old := time.Now().Add(-8 * 24 * time.Hour)
			id := insertAnnouncementDeliveryRowWithStatus(t, pool, announcementID, userID, "pending", old)
			t.Cleanup(func() {
				pool.Exec(context.Background(), `DELETE FROM announcement_email_deliveries WHERE id = $1`, id) //nolint:errcheck // best-effort cleanup
			})

			w.poll(context.Background())

			So(announcementDeliveryRowExists(t, pool, id), ShouldBeTrue)
		})
	})
}
