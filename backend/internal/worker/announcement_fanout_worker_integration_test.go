//go:build integration

package worker

import (
	"context"
	admindomain "learnflow_backend/internal/admin/domain"
	"learnflow_backend/internal/shared/testutil"
	"testing"
	"time"

	"github.com/jackc/pgx/v5"

	. "github.com/smartystreets/goconvey/convey"
)

func insertNotificationPreferences(t *testing.T, ctx context.Context, tx pgx.Tx, userID string, emailOnAnnouncement bool) {
	t.Helper()

	_, err := tx.Exec(ctx,
		`INSERT INTO notification_preferences (user_id, email_on_announcement) VALUES ($1, $2)`,
		userID, emailOnAnnouncement,
	)
	if err != nil {
		t.Fatalf("insertNotificationPreferences: %v", err)
	}
}

func TestAnnouncementFanOutWorkerQueryRecipients_PlatformWide_Integration(t *testing.T) {
	pool := testutil.NewTestPool(t)

	Convey("Given users backed by real Postgres", t, func() {
		testutil.WithTestTx(t, pool, func(ctx context.Context, tx pgx.Tx) {
			w := newTestAnnouncementFanOutWorker(tx, nil, &mockAnnouncementRepo{}, defaultTestEntityConfigs())

			optedIn := testutil.InsertTestUser(t, tx, testutil.RandomTestEmail(t, "fanout-opted-in"))
			insertNotificationPreferences(t, ctx, tx, optedIn, true)

			optedOut := testutil.InsertTestUser(t, tx, testutil.RandomTestEmail(t, "fanout-opted-out"))
			insertNotificationPreferences(t, ctx, tx, optedOut, false)

			// Documents a real gap: nothing in the app inserts a notification_preferences
			// row on registration (confirmed via grep — no writer of this table exists
			// outside this test file). recipientBaseSelectSQL LEFT JOINs it and filters
			// "np.email_on_announcement = true" — NULL = true is NULL in SQL, not true,
			// so a user with no preferences row (i.e. every real user today) is silently
			// excluded, even though the column's own DEFAULT is true.
			noPrefsRow := testutil.InsertTestUser(t, tx, testutil.RandomTestEmail(t, "fanout-no-prefs-row"))

			Convey("When querying platform-wide recipients", func() {
				recipients, err := w.queryRecipients(ctx, getAnnouncementUserListSQL, nil)

				So(err, ShouldBeNil)
				So(recipients, ShouldContain, optedIn)
				So(recipients, ShouldNotContain, optedOut)
				So(recipients, ShouldNotContain, noPrefsRow)
			})
		})
	})
}

func TestAnnouncementFanOutWorkerQueryRecipients_Course_Integration(t *testing.T) {
	pool := testutil.NewTestPool(t)

	Convey("Given a course with mixed-access users backed by real Postgres", t, func() {
		testutil.WithTestTx(t, pool, func(ctx context.Context, tx pgx.Tx) {
			w := newTestAnnouncementFanOutWorker(tx, nil, &mockAnnouncementRepo{}, defaultTestEntityConfigs())
			courseID := testutil.InsertTestCourse(t, tx)

			activeUser := testutil.InsertTestUser(t, tx, testutil.RandomTestEmail(t, "fanout-course-active"))
			insertNotificationPreferences(t, ctx, tx, activeUser, true)
			testutil.GrantCourseAccess(t, tx, activeUser, courseID, testutil.AccessGrant{})

			revokedUser := testutil.InsertTestUser(t, tx, testutil.RandomTestEmail(t, "fanout-course-revoked"))
			insertNotificationPreferences(t, ctx, tx, revokedUser, true)
			testutil.GrantCourseAccess(t, tx, revokedUser, courseID, testutil.AccessGrant{Status: "revoked"})

			created := time.Now().Add(-3 * time.Hour)
			granted := time.Now().Add(-2 * time.Hour)
			expired := time.Now().Add(-1 * time.Hour)
			expiredUser := testutil.InsertTestUser(t, tx, testutil.RandomTestEmail(t, "fanout-course-expired"))
			insertNotificationPreferences(t, ctx, tx, expiredUser, true)
			testutil.GrantCourseAccess(t, tx, expiredUser, courseID, testutil.AccessGrant{
				CreatedAt: &created, GrantedAt: &granted, ExpiresAt: &expired,
			})

			optedOutUser := testutil.InsertTestUser(t, tx, testutil.RandomTestEmail(t, "fanout-course-opted-out"))
			insertNotificationPreferences(t, ctx, tx, optedOutUser, false)
			testutil.GrantCourseAccess(t, tx, optedOutUser, courseID, testutil.AccessGrant{})

			Convey("When querying course recipients", func() {
				recipients, err := w.queryRecipients(ctx, getCourseAnnouncementUserListSQL, []any{courseID})

				So(err, ShouldBeNil)
				So(recipients, ShouldContain, activeUser)
				So(recipients, ShouldNotContain, revokedUser)
				So(recipients, ShouldNotContain, expiredUser)
				So(recipients, ShouldNotContain, optedOutUser)
			})
		})
	})
}

func TestAnnouncementFanOutWorkerQueryRecipients_ContentViaCourse_Integration(t *testing.T) {
	pool := testutil.NewTestPool(t)

	Convey("Given a content item reachable only via course access, backed by real Postgres", t, func() {
		testutil.WithTestTx(t, pool, func(ctx context.Context, tx pgx.Tx) {
			w := newTestAnnouncementFanOutWorker(tx, nil, &mockAnnouncementRepo{}, defaultTestEntityConfigs())
			courseID := testutil.InsertTestCourse(t, tx)
			contentItemID := testutil.InsertTestContentItem(t, tx)

			testutil.LinkContentItemToCourse(t, tx, courseID, contentItemID, 1, true)

			viaCourse := testutil.InsertTestUser(t, tx, testutil.RandomTestEmail(t, "fanout-content-via-course"))
			insertNotificationPreferences(t, ctx, tx, viaCourse, true)
			testutil.GrantCourseAccess(t, tx, viaCourse, courseID, testutil.AccessGrant{})

			noAccess := testutil.InsertTestUser(t, tx, testutil.RandomTestEmail(t, "fanout-content-no-access"))
			insertNotificationPreferences(t, ctx, tx, noAccess, true)

			Convey("When querying content recipients", func() {
				recipients, err := w.queryRecipients(ctx, getContentAnnouncementUserListSQL, []any{contentItemID})

				So(err, ShouldBeNil)
				So(recipients, ShouldContain, viaCourse)
				So(recipients, ShouldNotContain, noAccess)
			})
		})
	})
}

func TestAnnouncementFanOutWorkerFanOut_Integration(t *testing.T) {
	pool := testutil.NewTestPool(t)

	Convey("Given a platform-wide announcement and an opted-in user, backed by real Postgres", t, func() {
		testutil.WithTestTx(t, pool, func(ctx context.Context, tx pgx.Tx) {
			w := newTestAnnouncementFanOutWorker(tx, nil, &mockAnnouncementRepo{}, defaultTestEntityConfigs())

			optedIn := testutil.InsertTestUser(t, tx, testutil.RandomTestEmail(t, "fanout-e2e"))
			insertNotificationPreferences(t, ctx, tx, optedIn, true)
			announcementID := insertTestAnnouncement(t, tx, optedIn)

			Convey("When fanOut runs, it bulk-inserts announcement_email_deliveries rows", func() {
				err := w.fanOut(ctx, &admindomain.Announcement{ID: announcementID})
				So(err, ShouldBeNil)

				var status string
				err = tx.QueryRow(ctx,
					`SELECT status FROM announcement_email_deliveries WHERE user_id = $1 AND announcement_id = $2`,
					optedIn, announcementID,
				).Scan(&status)
				So(err, ShouldBeNil)
				So(status, ShouldEqual, "pending")
			})

			Convey("When fanOut runs twice for the same announcement, the unique index keeps a single delivery row", func() {
				So(w.fanOut(ctx, &admindomain.Announcement{ID: announcementID}), ShouldBeNil)
				So(w.fanOut(ctx, &admindomain.Announcement{ID: announcementID}), ShouldBeNil)

				var count int
				err := tx.QueryRow(ctx,
					`SELECT COUNT(*) FROM announcement_email_deliveries WHERE user_id = $1 AND announcement_id = $2`,
					optedIn, announcementID,
				).Scan(&count)
				So(err, ShouldBeNil)
				So(count, ShouldEqual, 1)
			})
		})
	})
}
