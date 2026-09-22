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

func TestAnnouncementDeliveryPollerGetList_Integration(t *testing.T) {
	pool := testutil.NewTestPool(t)

	Convey("Given a pending announcement_email_deliveries row backed by real Postgres", t, func() {
		Convey("When the recipient has a user_profiles row, getList scans successfully", func() {
			testutil.WithTestTx(t, pool, func(ctx context.Context, tx pgx.Tx) {
				userID := testutil.InsertTestUser(t, tx, testutil.RandomTestEmail(t, "delivery-poller"))
				insertUserProfile(t, ctx, tx, userID, "Jane")
				announcementID := insertTestAnnouncement(t, tx, userID)
				insertAnnouncementDelivery(t, tx, announcementID, userID, "pending", nil)

				p := NewAnnouncementDeliveryPoller(tx, nil, testutil.NewTestLogger(), nil)
				entries, err := p.getList(ctx)

				So(err, ShouldBeNil)
				So(entries, ShouldHaveLength, 1)
				So(entries[0].EventType, ShouldEqual, events.EventAnnouncementDeliver)
			})
		})

		// querySelectAnnouncements LEFT JOINs user_profiles, which may have no row for a
		// recipient (registration always inserts one, but its first_name is optional and
		// can be NULL) — COALESCE(up.first_name, '') keeps that NULL/missing-row case
		// from breaking the scan into AnnouncementDeliver.FirstName (a plain string).
		Convey("When the recipient has no user_profiles row, getList still scans (empty first name)", func() {
			testutil.WithTestTx(t, pool, func(ctx context.Context, tx pgx.Tx) {
				userID := testutil.InsertTestUser(t, tx, testutil.RandomTestEmail(t, "delivery-poller-noprofile"))
				announcementID := insertTestAnnouncement(t, tx, userID)
				insertAnnouncementDelivery(t, tx, announcementID, userID, "pending", nil)

				p := NewAnnouncementDeliveryPoller(tx, nil, testutil.NewTestLogger(), nil)
				entries, err := p.getList(ctx)

				So(err, ShouldBeNil)
				So(entries, ShouldHaveLength, 1)
			})
		})

		Convey("When the recipient has a user_profiles row with a NULL first name, getList still scans", func() {
			testutil.WithTestTx(t, pool, func(ctx context.Context, tx pgx.Tx) {
				userID := testutil.InsertTestUser(t, tx, testutil.RandomTestEmail(t, "delivery-poller-nullname"))
				_, err := tx.Exec(ctx, `INSERT INTO user_profiles (user_id, first_name) VALUES ($1, NULL)`, userID)
				So(err, ShouldBeNil)
				announcementID := insertTestAnnouncement(t, tx, userID)
				insertAnnouncementDelivery(t, tx, announcementID, userID, "pending", nil)

				p := NewAnnouncementDeliveryPoller(tx, nil, testutil.NewTestLogger(), nil)
				entries, err := p.getList(ctx)

				So(err, ShouldBeNil)
				So(entries, ShouldHaveLength, 1)
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
				announcementID := insertTestAnnouncement(t, tx, userID)
				deliveryID := insertAnnouncementDelivery(t, tx, announcementID, userID, "pending", nil)

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
