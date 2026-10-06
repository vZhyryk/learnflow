package auditdomain

import (
	"errors"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

const validTargetID = "11111111-1111-1111-1111-111111111111"

func TestGetInstanceAdminActionsRequestValidate(t *testing.T) {
	Convey("Given a GetInstanceAdminActionsRequest", t, func() {
		Convey("When id and target type are valid", func() {
			req := GetInstanceAdminActionsRequest{ItemID: validTargetID, TargetType: TargetUser}
			So(req.Validate(), ShouldBeNil)
		})

		Convey("When the id is empty", func() {
			req := GetInstanceAdminActionsRequest{TargetType: TargetUser}
			So(errors.Is(req.Validate(), ErrInvalidItemID), ShouldBeTrue)
		})

		Convey("When the id is not a UUID", func() {
			req := GetInstanceAdminActionsRequest{ItemID: "nope", TargetType: TargetUser}
			So(errors.Is(req.Validate(), ErrInvalidItemID), ShouldBeTrue)
		})

		Convey("When the target type is unknown", func() {
			req := GetInstanceAdminActionsRequest{ItemID: validTargetID, TargetType: "video"}
			So(errors.Is(req.Validate(), ErrInvalidItemType), ShouldBeTrue)
		})

		Convey("When the target type is empty", func() {
			req := GetInstanceAdminActionsRequest{ItemID: validTargetID}
			So(errors.Is(req.Validate(), ErrInvalidItemType), ShouldBeTrue)
		})

		Convey("Every declared target type is accepted", func() {
			for _, target := range []AdminTargetType{
				TargetUser, TargetBooking, TargetCourse, TargetFailedJob, TargetPayment, TargetSupportChat,
				TargetReview, TargetAnnouncement, TargetArticle, TargetGiftCoupon, TargetContentItem, TargetExpense,
			} {
				req := GetInstanceAdminActionsRequest{ItemID: validTargetID, TargetType: target}
				So(req.Validate(), ShouldBeNil)
			}
		})
	})
}
