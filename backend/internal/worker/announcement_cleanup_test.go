package worker

import (
	"learnflow_backend/internal/shared/testutil"
	"testing"
	"time"

	. "github.com/smartystreets/goconvey/convey"
)

// TestNewAnnouncementCleanUpWorker only checks wiring — poll's own branches (partial
// batch, full-batch loop, ctx cancellation, Exec failure) are exercised generically by
// TestOutboxCleanupWorkerPoll* in outbox_cleanup_test.go against the same CleanupWorker[T]
// engine, so they aren't duplicated here.
func TestNewAnnouncementCleanUpWorker(t *testing.T) {
	Convey("Given a NewAnnouncementCleanUpWorker", t, func() {
		Convey("When wiring the cleanup worker", func() {
			w := NewAnnouncementCleanUpWorker(&testutil.MockQueryRunner{}, testutil.NewTestLogger(), 24*time.Hour)

			So(w.cleanUpQuery, ShouldEqual, announcementCleanUpQuery)
			So(w.workerName, ShouldEqual, "AnnouncementCleanupWorker")
			So(w.pollInterval, ShouldEqual, 24*time.Hour)
		})
	})
}
