package auditdomain

import (
	"errors"
	"testing"
	"time"

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

func TestAdminActionFilterValidate(t *testing.T) {
	from := time.Date(2026, 10, 1, 0, 0, 0, 0, time.UTC)
	to := from.Add(24 * time.Hour)

	cases := []struct {
		name   string
		filter AdminActionFilter
		want   error
	}{
		{"an empty filter", AdminActionFilter{}, nil},
		{"every field set", AdminActionFilter{AdminUserID: validTargetID, ActionType: ActionBlockUser, From: &from, To: &to}, nil},
		{"only from", AdminActionFilter{From: &from}, nil},
		{"only to", AdminActionFilter{To: &to}, nil},
		{"an admin id that is not a UUID", AdminActionFilter{AdminUserID: "nope"}, ErrInvalidAdminUserID},
		{"an unknown action type", AdminActionFilter{ActionType: "launch_rocket"}, ErrInvalidActionType},
		{"from after to", AdminActionFilter{From: &to, To: &from}, ErrInvalidDateRange},
		{"from equal to (an empty range)", AdminActionFilter{From: &from, To: &from}, ErrInvalidDateRange},
	}

	for _, tc := range cases {
		Convey("AdminActionFilter.Validate with "+tc.name, t, func() {
			err := tc.filter.Validate()

			if tc.want == nil {
				So(err, ShouldBeNil)
			} else {
				So(errors.Is(err, tc.want), ShouldBeTrue)
			}
		})
	}

	Convey("Every action type constant is accepted by the filter", t, func() {
		for _, action := range []AdminActionType{
			ActionAssignSubadmin, ActionRevokeSubadmin, ActionDeleteUser, ActionRestoreUser, ActionBlockUser, ActionUnblockUser,
			ActionConfirmBooking, ActionCancelBooking, ActionGrantItemAccess, ActionIssueRefund, ActionRecordExpense,
			ActionRescheduleBooking, ActionCloseSupportChat, ActionCreateGiftCoupon, ActionRevokeGiftCoupon,
			ActionPublishItem, ActionDeleteItem, ActionArchiveItem, ActionCreateItem, ActionUpdateItem, ActionApproveItem,
		} {
			filter := AdminActionFilter{ActionType: action}
			So(filter.Validate(), ShouldBeNil)
		}
	})
}
