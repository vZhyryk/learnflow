//go:build integration

package worker

import (
	"context"
	"testing"
	"time"

	"learnflow_backend/internal/infrastructure/db"
	"learnflow_backend/internal/shared/testutil"

	"github.com/jackc/pgx/v5"
	. "github.com/smartystreets/goconvey/convey"
)

func announcementDeliveryRowExists(t *testing.T, q db.QueryRunner, id string) bool {
	t.Helper()

	var exists bool
	if err := q.QueryRow(context.Background(),
		`SELECT EXISTS(SELECT 1 FROM announcement_email_deliveries WHERE id = $1)`, id).Scan(&exists); err != nil {
		t.Fatalf("announcementDeliveryRowExists: %v", err)
	}
	return exists
}

func seedAnnouncementDelivery(t *testing.T, tx pgx.Tx, status string, age time.Duration) string {
	t.Helper()

	userID := testutil.InsertRandomTestUser(t, tx)
	announcementID := insertTestAnnouncement(t, tx, userID)
	updatedAt := time.Now().Add(-age)
	return insertAnnouncementDelivery(t, tx, announcementID, userID, status, &updatedAt)
}

func TestAnnouncementCleanUpWorkerPoll_Integration(t *testing.T) {
	pool := testutil.NewTestPool(t)

	Convey("Given announcement_email_deliveries rows backed by real Postgres", t, func() {
		Convey("When a sent row is older than the 7-day retention window", func() {
			testutil.WithTestTx(t, pool, func(ctx context.Context, tx pgx.Tx) {
				w := NewAnnouncementCleanUpWorker(tx, testutil.NewTestLogger(), 24*time.Hour)
				id := seedAnnouncementDelivery(t, tx, "sent", 8*24*time.Hour)

				w.poll(ctx)

				So(announcementDeliveryRowExists(t, tx, id), ShouldBeFalse)
			})
		})

		Convey("When a failed row is older than the 7-day retention window", func() {
			testutil.WithTestTx(t, pool, func(ctx context.Context, tx pgx.Tx) {
				w := NewAnnouncementCleanUpWorker(tx, testutil.NewTestLogger(), 24*time.Hour)
				id := seedAnnouncementDelivery(t, tx, "failed", 8*24*time.Hour)

				w.poll(ctx)

				So(announcementDeliveryRowExists(t, tx, id), ShouldBeFalse)
			})
		})

		Convey("When a sent row is within the 7-day retention window", func() {
			testutil.WithTestTx(t, pool, func(ctx context.Context, tx pgx.Tx) {
				w := NewAnnouncementCleanUpWorker(tx, testutil.NewTestLogger(), 24*time.Hour)
				id := seedAnnouncementDelivery(t, tx, "sent", 24*time.Hour)

				w.poll(ctx)

				So(announcementDeliveryRowExists(t, tx, id), ShouldBeTrue)
			})
		})

		Convey("When a pending row is old, poll leaves it untouched (only sent/failed are retention-eligible)", func() {
			testutil.WithTestTx(t, pool, func(ctx context.Context, tx pgx.Tx) {
				w := NewAnnouncementCleanUpWorker(tx, testutil.NewTestLogger(), 24*time.Hour)
				id := seedAnnouncementDelivery(t, tx, "pending", 8*24*time.Hour)

				w.poll(ctx)

				So(announcementDeliveryRowExists(t, tx, id), ShouldBeTrue)
			})
		})
	})
}
