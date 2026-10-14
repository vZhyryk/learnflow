//go:build integration

package access

import (
	"context"
	"testing"
	"time"

	"learnflow_backend/internal/shared/testutil"

	"github.com/jackc/pgx/v5"
	. "github.com/smartystreets/goconvey/convey"
)

// pastGrant builds an expired AccessGrant satisfying created_at <= granted_at < expires_at.
func pastGrant() testutil.AccessGrant {
	created := time.Now().Add(-3 * 24 * time.Hour)
	granted := time.Now().Add(-2 * 24 * time.Hour)
	expires := time.Now().Add(-24 * time.Hour)
	return testutil.AccessGrant{AccessType: "purchased", CreatedAt: &created, GrantedAt: &granted, ExpiresAt: &expires}
}

func TestHasAccessCourseIntegration(t *testing.T) {
	pool := testutil.NewTestPool(t)

	Convey("Given an access Checker backed by real Postgres", t, func() {
		Convey("An active grant gives access", func() {
			testutil.WithTestTx(t, pool, func(ctx context.Context, tx pgx.Tx) {
				userID, courseID := testutil.InsertRandomTestUser(t, tx), testutil.InsertTestCourse(t, tx)
				testutil.GrantCourseAccess(t, tx, userID, courseID, testutil.AccessGrant{})

				ok, err := New(tx).HasAccessCourse(ctx, userID, courseID)

				So(err, ShouldBeNil)
				So(ok, ShouldBeTrue)
			})
		})

		Convey("An expired grant gives no access", func() {
			testutil.WithTestTx(t, pool, func(ctx context.Context, tx pgx.Tx) {
				userID, courseID := testutil.InsertRandomTestUser(t, tx), testutil.InsertTestCourse(t, tx)
				testutil.GrantCourseAccess(t, tx, userID, courseID, pastGrant())

				ok, err := New(tx).HasAccessCourse(ctx, userID, courseID)

				So(err, ShouldBeNil)
				So(ok, ShouldBeFalse)
			})
		})

		Convey("A revoked grant gives no access", func() {
			testutil.WithTestTx(t, pool, func(ctx context.Context, tx pgx.Tx) {
				userID, courseID := testutil.InsertRandomTestUser(t, tx), testutil.InsertTestCourse(t, tx)
				testutil.GrantCourseAccess(t, tx, userID, courseID, testutil.AccessGrant{Status: "revoked"})

				ok, err := New(tx).HasAccessCourse(ctx, userID, courseID)

				So(err, ShouldBeNil)
				So(ok, ShouldBeFalse)
			})
		})

		Convey("A grant of another user gives no access", func() {
			testutil.WithTestTx(t, pool, func(ctx context.Context, tx pgx.Tx) {
				owner, other := testutil.InsertRandomTestUser(t, tx), testutil.InsertRandomTestUser(t, tx)
				courseID := testutil.InsertTestCourse(t, tx)
				testutil.GrantCourseAccess(t, tx, owner, courseID, testutil.AccessGrant{})

				ok, err := New(tx).HasAccessCourse(ctx, other, courseID)

				So(err, ShouldBeNil)
				So(ok, ShouldBeFalse)
			})
		})
	})
}

func TestHasAccessContentIntegration(t *testing.T) {
	pool := testutil.NewTestPool(t)

	Convey("Given an access Checker backed by real Postgres", t, func() {
		Convey("An active direct content grant gives access", func() {
			testutil.WithTestTx(t, pool, func(ctx context.Context, tx pgx.Tx) {
				userID, contentID := testutil.InsertRandomTestUser(t, tx), testutil.InsertTestContentItem(t, tx)
				testutil.GrantContentAccess(t, tx, userID, contentID, testutil.AccessGrant{})

				ok, err := New(tx).HasAccessContent(ctx, userID, contentID)

				So(err, ShouldBeNil)
				So(ok, ShouldBeTrue)
			})
		})

		Convey("An expired or revoked direct grant gives no access", func() {
			testutil.WithTestTx(t, pool, func(ctx context.Context, tx pgx.Tx) {
				userID := testutil.InsertRandomTestUser(t, tx)
				expiredID, revokedID := testutil.InsertTestContentItem(t, tx), testutil.InsertTestContentItem(t, tx)
				testutil.GrantContentAccess(t, tx, userID, expiredID, pastGrant())
				testutil.GrantContentAccess(t, tx, userID, revokedID, testutil.AccessGrant{Status: "revoked"})
				checker := New(tx)

				expiredOK, err := checker.HasAccessContent(ctx, userID, expiredID)
				So(err, ShouldBeNil)
				revokedOK, err := checker.HasAccessContent(ctx, userID, revokedID)
				So(err, ShouldBeNil)

				So(expiredOK, ShouldBeFalse)
				So(revokedOK, ShouldBeFalse)
			})
		})

		Convey("An active course grant gives access to the course's content items", func() {
			testutil.WithTestTx(t, pool, func(ctx context.Context, tx pgx.Tx) {
				userID, courseID := testutil.InsertRandomTestUser(t, tx), testutil.InsertTestCourse(t, tx)
				contentID := testutil.InsertTestContentItem(t, tx)
				testutil.LinkContentItemToCourse(t, tx, courseID, contentID, 1, true)
				testutil.GrantCourseAccess(t, tx, userID, courseID, testutil.AccessGrant{})

				ok, err := New(tx).HasAccessContent(ctx, userID, contentID)

				So(err, ShouldBeNil)
				So(ok, ShouldBeTrue)
			})
		})

		Convey("An expired course grant gives no access to the course's content items", func() {
			testutil.WithTestTx(t, pool, func(ctx context.Context, tx pgx.Tx) {
				userID, courseID := testutil.InsertRandomTestUser(t, tx), testutil.InsertTestCourse(t, tx)
				contentID := testutil.InsertTestContentItem(t, tx)
				testutil.LinkContentItemToCourse(t, tx, courseID, contentID, 1, true)
				testutil.GrantCourseAccess(t, tx, userID, courseID, pastGrant())

				ok, err := New(tx).HasAccessContent(ctx, userID, contentID)

				So(err, ShouldBeNil)
				So(ok, ShouldBeFalse)
			})
		})

		Convey("A content item outside the granted course gives no access", func() {
			testutil.WithTestTx(t, pool, func(ctx context.Context, tx pgx.Tx) {
				userID, courseID := testutil.InsertRandomTestUser(t, tx), testutil.InsertTestCourse(t, tx)
				unlinked := testutil.InsertTestContentItem(t, tx)
				testutil.GrantCourseAccess(t, tx, userID, courseID, testutil.AccessGrant{})

				ok, err := New(tx).HasAccessContent(ctx, userID, unlinked)

				So(err, ShouldBeNil)
				So(ok, ShouldBeFalse)
			})
		})
	})
}
