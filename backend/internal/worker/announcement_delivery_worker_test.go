package worker

import (
	"errors"
	"learnflow_backend/internal/shared/mailer"
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

func TestValidateAnnouncementPayload(t *testing.T) {
	Convey("ValidateAnnouncementPayload", t, func() {
		Convey("When all required fields are present", func() {
			So(ValidateAnnouncementPayload(validAnnouncementDeliver()), ShouldBeNil)
		})

		Convey("When UserID is missing", func() {
			p := validAnnouncementDeliver()
			p.UserID = ""
			err := ValidateAnnouncementPayload(p)
			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldContainSubstring, "UserID")
		})

		Convey("When Email is missing", func() {
			p := validAnnouncementDeliver()
			p.Email = ""
			err := ValidateAnnouncementPayload(p)
			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldContainSubstring, "Email")
		})

		Convey("When AnnouncementID is missing", func() {
			p := validAnnouncementDeliver()
			p.AnnouncementID = ""
			err := ValidateAnnouncementPayload(p)
			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldContainSubstring, "AnnouncementID")
		})

		Convey("When FirstName is missing", func() {
			p := validAnnouncementDeliver()
			p.FirstName = ""
			err := ValidateAnnouncementPayload(p)
			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldContainSubstring, "FirstName")
		})

		Convey("When Title is missing", func() {
			p := validAnnouncementDeliver()
			p.Title = ""
			err := ValidateAnnouncementPayload(p)
			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldContainSubstring, "Title")
		})

		Convey("When Body is missing", func() {
			p := validAnnouncementDeliver()
			p.Body = ""
			err := ValidateAnnouncementPayload(p)
			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldContainSubstring, "Body")
		})

		Convey("When multiple fields are missing", func() {
			err := ValidateAnnouncementPayload(AnnouncementDeliver{})
			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldContainSubstring, "UserID")
			So(err.Error(), ShouldContainSubstring, "Email")
			So(err.Error(), ShouldContainSubstring, "AnnouncementID")
		})
	})
}

func TestHandleAnnouncementProcess(t *testing.T) {
	Convey("Given an AnnouncementDeliver payload", t, func() {
		payload := validAnnouncementDeliver()

		Convey("When sending the announcement email", func() {
			m := &mockMailer{
				send: func(templateFile string, data any, ccUser mailer.CCUser, _ []string) error {
					So(templateFile, ShouldEqual, "announcement_delivery.html")
					dataMap, ok := data.(map[string]string)
					So(ok, ShouldBeTrue)
					So(dataMap["name"], ShouldEqual, "John")
					So(dataMap["title"], ShouldEqual, "Big news")
					So(dataMap["body"], ShouldEqual, "Something happened")
					So(ccUser.Mail, ShouldEqual, "user@example.com")
					return nil
				},
			}
			err := HandleAnnouncementProcess(payload, "https://example.com", m)
			So(err, ShouldBeNil)
		})

		Convey("When the mailer fails", func() {
			m := &mockMailer{
				send: func(_ string, _ any, _ mailer.CCUser, _ []string) error {
					return errSendFailed
				},
			}
			err := HandleAnnouncementProcess(payload, "https://example.com", m)
			So(err, ShouldEqual, errSendFailed)
		})
	})
}

func TestGenerateAnnouncementIdempotencyKey(t *testing.T) {
	runIdempotencyKeyTest(t, "AnnouncementDeliver",
		AnnouncementDeliver{UserID: "user-1", AnnouncementID: "ann-1"},
		GenerateAnnouncementIdempotencyKey,
		"processed:announcement_delivery:user-1:ann-1",
	)
}
