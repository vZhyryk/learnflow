//go:build integration

package worker

import (
	"context"
	"errors"
	"testing"

	"learnflow_backend/internal/shared/testutil"

	"github.com/jackc/pgx/v5"

	. "github.com/smartystreets/goconvey/convey"
)

// insertFailedJob inserts a failed_jobs row with the given starting attempt_count
// and returns its id, so markAsRetriedFailedSQL's attempt_count + 1 boundary can
// be exercised against a real Postgres.
func insertFailedJob(t *testing.T, ctx context.Context, tx pgx.Tx, attemptCount int) (id string) {
	t.Helper()

	err := tx.QueryRow(ctx,
		`INSERT INTO failed_jobs (event_type, queue_name, payload_json, attempt_count, failed_at)
		 VALUES ($1, $2, $3, $4, now())
		 RETURNING id`,
		"user.registered", "email", `{"id":"u-1"}`, attemptCount,
	).Scan(&id)
	if err != nil {
		t.Fatalf("insertFailedJob: %v", err)
	}

	return id
}

func getFailedJobState(t *testing.T, ctx context.Context, tx pgx.Tx, id string) (attemptCount int, resolutionNote *string, resolved bool) {
	t.Helper()

	err := tx.QueryRow(ctx,
		`SELECT attempt_count, resolution_note, resolved_at IS NOT NULL FROM failed_jobs WHERE id = $1`,
		id,
	).Scan(&attemptCount, &resolutionNote, &resolved)
	if err != nil {
		t.Fatalf("getFailedJobState: %v", err)
	}

	return attemptCount, resolutionNote, resolved
}

func TestMarkAsRetriedFailedSQL_MaxRetriesBoundary_Integration(t *testing.T) {
	pool := testutil.NewTestPool(t)

	Convey("Given a DLQRetryWorker backed by real Postgres", t, func() {
		Convey("When attempt_count + 1 is still below the ceiling, the row stays unresolved", func() {
			testutil.WithTestTx(t, pool, func(ctx context.Context, tx pgx.Tx) {
				w := NewDLQRetryWorker(tx, nil, testutil.NewTestLogger(), nil)
				id := insertFailedJob(t, ctx, tx, 4)

				w.markEntryFailed(ctx, errors.New("boom"), id)

				attemptCount, resolutionNote, resolved := getFailedJobState(t, ctx, tx, id)
				So(attemptCount, ShouldEqual, 5)
				So(resolutionNote, ShouldBeNil)
				So(resolved, ShouldBeFalse)
			})
		})

		Convey("When attempt_count + 1 reaches the ceiling, the row is resolved as max_retries_exceeded", func() {
			testutil.WithTestTx(t, pool, func(ctx context.Context, tx pgx.Tx) {
				w := NewDLQRetryWorker(tx, nil, testutil.NewTestLogger(), nil)
				id := insertFailedJob(t, ctx, tx, 5)

				w.markEntryFailed(ctx, errors.New("boom"), id)

				attemptCount, resolutionNote, resolved := getFailedJobState(t, ctx, tx, id)
				So(attemptCount, ShouldEqual, 6)
				So(resolutionNote, ShouldNotBeNil)
				So(*resolutionNote, ShouldEqual, "max_retries_exceeded")
				So(resolved, ShouldBeTrue)
			})
		})
	})
}
