package worker

import (
	"encoding/json"
	"learnflow_backend/internal/events"
	"learnflow_backend/internal/shared/testutil"
	"testing"
	"time"

	. "github.com/smartystreets/goconvey/convey"
)

func fakeAnnouncementDeliverScan(entry AnnouncementDeliver) func(dest ...any) error {
	return func(dest ...any) error {
		*testutil.CastStr(dest[0], 0) = entry.ID
		*testutil.CastStr(dest[1], 1) = entry.UserID
		*testutil.CastStr(dest[2], 2) = entry.AnnouncementID
		*testutil.CastStr(dest[3], 3) = entry.FirstName
		*testutil.CastStr(dest[4], 4) = entry.Title
		*testutil.CastStr(dest[5], 5) = entry.Body
		*testutil.CastStr(dest[6], 6) = entry.Email
		return nil
	}
}

func TestScanAnnouncementDeliveryPoller(t *testing.T) {
	Convey("scanAnnouncementDeliveryPoller", t, func() {
		Convey("When the row fails to scan", func() {
			_, err := scanAnnouncementDeliveryPoller(&testutil.MockRow{
				ScanFn: func(_ ...any) error { return testutil.ErrDBUnexpected },
			})

			So(err, ShouldNotBeNil)
		})

		Convey("When the row scans successfully", func() {
			entry := AnnouncementDeliver{
				ID: "delivery-1", UserID: "user-1", AnnouncementID: "ann-1",
				FirstName: "John", Title: "Big news", Body: "Something happened", Email: "user@example.com",
			}

			pollerEntry, err := scanAnnouncementDeliveryPoller(&testutil.MockRow{ScanFn: fakeAnnouncementDeliverScan(entry)})

			So(err, ShouldBeNil)
			So(pollerEntry.ID, ShouldEqual, "delivery-1")
			So(pollerEntry.EventType, ShouldEqual, events.EventAnnouncementDeliver)

			var gotPayload AnnouncementDeliver
			So(json.Unmarshal([]byte(pollerEntry.PayloadJSON), &gotPayload), ShouldBeNil)
			So(gotPayload, ShouldResemble, entry)
		})
	})
}

func TestNewAnnouncementDeliveryPoller(t *testing.T) {
	Convey("Given a NewAnnouncementDeliveryPoller", t, func() {
		Convey("When wiring the poller's SQLList", func() {
			p := NewAnnouncementDeliveryPoller(&testutil.MockQueryRunner{}, nil, testutil.NewTestLogger(), testutil.NoopTransactor{})

			So(p.SQLList.selectSQL, ShouldEqual, querySelectAnnouncements)
			So(p.SQLList.markSuccessSQL, ShouldEqual, queryMarkAnnouncementSend)
			So(p.SQLList.markFailedSQL, ShouldEqual, queryMarkAnnouncementFailed)
			So(p.pollInterval, ShouldEqual, 5*time.Second)
		})
	})
}
