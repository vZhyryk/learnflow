package worker

import (
	"errors"
	"learnflow_backend/internal/shared/testutil"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

var errSendFailed = errors.New("send failed")

func validAnnouncementDeliver() AnnouncementDeliver {
	return AnnouncementDeliver{
		ID: "delivery-1", UserID: "user-1", AnnouncementID: "ann-1",
		FirstName: "John", Title: "Big news", Body: "Something happened", Email: "user@example.com",
	}
}

func TestAnnouncementDeliveryWorker(t *testing.T) {
	runEmailWorkerCase(t, "announcement_delivery", emailWorkerCase[AnnouncementDeliver]{
		worker:      NewAnnouncementDeliveryWorker(nil, nil, testutil.NewTestLogger(), nil, testBaseURL),
		valid:       validAnnouncementDeliver(),
		invalid:     AnnouncementDeliver{},
		wantMissing: []string{"UserID", "Email", "AnnouncementID", "FirstName", "Title", "Body"},
		wantKey:     "processed:announcement_delivery:user-1:ann-1",
		wantData:    map[string]string{"name": "John", "title": "Big news", "body": "Something happened"},
		wantTo:      "user@example.com",
	})
}

func TestAnnouncementDeliverRequiredFields(t *testing.T) {
	Convey("AnnouncementDeliver validation", t, func() {
		cases := map[string]func(p *AnnouncementDeliver){
			"UserID":         func(p *AnnouncementDeliver) { p.UserID = "" },
			"Email":          func(p *AnnouncementDeliver) { p.Email = "" },
			"AnnouncementID": func(p *AnnouncementDeliver) { p.AnnouncementID = "" },
			"FirstName":      func(p *AnnouncementDeliver) { p.FirstName = "" },
			"Title":          func(p *AnnouncementDeliver) { p.Title = "" },
			"Body":           func(p *AnnouncementDeliver) { p.Body = "" },
		}

		for field, clear := range cases {
			Convey("When "+field+" is missing, only that field is reported", func() {
				p := validAnnouncementDeliver()
				clear(&p)

				err := ValidatePayload("announcement_delivery", p)

				So(err, ShouldNotBeNil)
				So(err.Error(), ShouldContainSubstring, "["+field+"]")
			})
		}
	})
}
