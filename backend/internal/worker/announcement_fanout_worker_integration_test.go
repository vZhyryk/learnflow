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

func insertCourseAccess(t *testing.T, ctx context.Context, tx pgx.Tx, userID, courseID, status string, expiresAt *time.Time) {
	t.Helper()

	_, err := tx.Exec(ctx,
		`INSERT INTO user_course_access (user_id, course_id, access_type, status, granted_at, expires_at)
		 VALUES ($1, $2, 'admin_granted', $3, now(), $4)`,
		userID, courseID, status, expiresAt,
	)
	if err != nil {
		t.Fatalf("insertCourseAccess: %v", err)
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
			insertCourseAccess(t, ctx, tx, activeUser, courseID, "active", nil)

			revokedUser := testutil.InsertTestUser(t, tx, testutil.RandomTestEmail(t, "fanout-course-revoked"))
			insertNotificationPreferences(t, ctx, tx, revokedUser, true)
			insertCourseAccess(t, ctx, tx, revokedUser, courseID, "revoked", nil)

			expired := time.Now().Add(-1 * time.Hour)
			expiredUser := testutil.InsertTestUser(t, tx, testutil.RandomTestEmail(t, "fanout-course-expired"))
			insertNotificationPreferences(t, ctx, tx, expiredUser, true)
			insertCourseAccess(t, ctx, tx, expiredUser, courseID, "active", &expired)

			optedOutUser := testutil.InsertTestUser(t, tx, testutil.RandomTestEmail(t, "fanout-course-opted-out"))
			insertNotificationPreferences(t, ctx, tx, optedOutUser, false)
			insertCourseAccess(t, ctx, tx, optedOutUser, courseID, "active", nil)

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

			_, err := tx.Exec(ctx,
				`INSERT INTO course_content_items (course_id, content_item_id, position, is_required) VALUES ($1, $2, 1, true)`,
				courseID, contentItemID,
			)
			So(err, ShouldBeNil)

			viaCourse := testutil.InsertTestUser(t, tx, testutil.RandomTestEmail(t, "fanout-content-via-course"))
			insertNotificationPreferences(t, ctx, tx, viaCourse, true)
			insertCourseAccess(t, ctx, tx, viaCourse, courseID, "active", nil)

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
			announcementID := insertAnnouncementTx(t, ctx, tx, optedIn)

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

			Convey("When fanOut runs twice for the same announcement, the second run violates no unique constraint and duplicates the row", func() {
				// Documents current behavior: announcement_email_deliveries has no
				// UNIQUE(user_id, announcement_id) constraint, so re-running fan-out for
				// the same announcement (e.g. after a re-approve, see NOTIFICATION_FLOW.md
				// decision #7) inserts a second delivery row instead of failing or
				// upserting — the recipient would get the email twice.
				So(w.fanOut(ctx, &admindomain.Announcement{ID: announcementID}), ShouldBeNil)
				So(w.fanOut(ctx, &admindomain.Announcement{ID: announcementID}), ShouldBeNil)

				var count int
				err := tx.QueryRow(ctx,
					`SELECT COUNT(*) FROM announcement_email_deliveries WHERE user_id = $1 AND announcement_id = $2`,
					optedIn, announcementID,
				).Scan(&count)
				So(err, ShouldBeNil)
				So(count, ShouldEqual, 2)
			})
		})
	})
}
