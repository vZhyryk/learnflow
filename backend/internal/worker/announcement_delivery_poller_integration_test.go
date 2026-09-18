//go:build integration

package worker

import (
	"context"
	"testing"

	"learnflow_backend/internal/events"
	"learnflow_backend/internal/shared/testutil"

	"github.com/jackc/pgx/v5"

	. "github.com/smartystreets/goconvey/convey"
)

func insertUserProfile(t *testing.T, ctx context.Context, tx pgx.Tx, userID, firstName string) {
	t.Helper()

	_, err := tx.Exec(ctx, `INSERT INTO user_profiles (user_id, first_name) VALUES ($1, $2)`, userID, firstName)
	if err != nil {
		t.Fatalf("insertUserProfile: %v", err)
	}
}

func insertAnnouncementTx(t *testing.T, ctx context.Context, tx pgx.Tx, createdByUserID string) string {
	t.Helper()

	var id string
	if err := tx.QueryRow(ctx, insertTestAnnouncementSQL, createdByUserID).Scan(&id); err != nil {
		t.Fatalf("insertAnnouncementTx: %v", err)
	}
	return id
}

func insertPendingDeliveryTx(t *testing.T, ctx context.Context, tx pgx.Tx, userID, announcementID string) string {
	t.Helper()

	var id string
	err := tx.QueryRow(ctx,
		`INSERT INTO announcement_email_deliveries (user_id, announcement_id) VALUES ($1, $2) RETURNING id`,
		userID, announcementID,
	).Scan(&id)
	if err != nil {
		t.Fatalf("insertPendingDeliveryTx: %v", err)
	}
	return id
}

func TestAnnouncementDeliveryPollerGetList_Integration(t *testing.T) {
	pool := testutil.NewTestPool(t)

	Convey("Given a pending announcement_email_deliveries row backed by real Postgres", t, func() {
		Convey("When the recipient has a user_profiles row, getList scans successfully", func() {
			testutil.WithTestTx(t, pool, func(ctx context.Context, tx pgx.Tx) {
				userID := testutil.InsertTestUser(t, tx, testutil.RandomTestEmail(t, "delivery-poller"))
				insertUserProfile(t, ctx, tx, userID, "Jane")
				announcementID := insertAnnouncementTx(t, ctx, tx, userID)
				insertPendingDeliveryTx(t, ctx, tx, userID, announcementID)

				p := NewAnnouncementDeliveryPoller(tx, nil, testutil.NewTestLogger(), nil)
				entries, err := p.getList(ctx)

				So(err, ShouldBeNil)
				So(entries, ShouldHaveLength, 1)
				So(entries[0].EventType, ShouldEqual, "announcement.deliver")
			})
		})

		// Documents a real gap: nothing in the app creates a user_profiles row on
		// registration (confirmed via grep — no INSERT INTO user_profiles anywhere
		// outside this test file), so every real newly-registered user hits this path.
		// querySelectAnnouncements LEFT JOINs user_profiles and scans first_name into a
		// plain (non-pointer) string — Postgres returns NULL for a missing profile row,
		// and pgx refuses to scan NULL into a non-nullable Go string.
		Convey("When the recipient has no user_profiles row, getList fails to scan", func() {
			testutil.WithTestTx(t, pool, func(ctx context.Context, tx pgx.Tx) {
				userID := testutil.InsertTestUser(t, tx, testutil.RandomTestEmail(t, "delivery-poller-noprofile"))
				announcementID := insertAnnouncementTx(t, ctx, tx, userID)
				insertPendingDeliveryTx(t, ctx, tx, userID, announcementID)

				p := NewAnnouncementDeliveryPoller(tx, nil, testutil.NewTestLogger(), nil)
				_, err := p.getList(ctx)

				So(err, ShouldNotBeNil)
			})
		})
	})
}

func TestAnnouncementDeliveryPollerPoll_Integration(t *testing.T) {
	pool := testutil.NewTestPool(t)

	Convey("Given a pending delivery row backed by real Postgres", t, func() {
		Convey("When poll runs successfully, the row is marked sent", func() {
			testutil.WithTestTx(t, pool, func(ctx context.Context, tx pgx.Tx) {
				userID := testutil.InsertTestUser(t, tx, testutil.RandomTestEmail(t, "delivery-poller-poll"))
				insertUserProfile(t, ctx, tx, userID, "Jane")
				announcementID := insertAnnouncementTx(t, ctx, tx, userID)
				deliveryID := insertPendingDeliveryTx(t, ctx, tx, userID, announcementID)

				publisher := &mockPublisher{publish: func(_ context.Context, _ events.EventType, _ any) error { return nil }}
				p := NewAnnouncementDeliveryPoller(tx, publisher, testutil.NewTestLogger(), testutil.NoopTransactor{})

				p.poll(ctx)

				var status string
				err := tx.QueryRow(ctx, `SELECT status FROM announcement_email_deliveries WHERE id = $1`, deliveryID).Scan(&status)
				So(err, ShouldBeNil)
				So(status, ShouldEqual, "sent")
			})
		})
	})
}
