package events

import (
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestIsKnownEventType(t *testing.T) {
	Convey("IsKnownEventType", t, func() {
		Convey("returns true for known event types", func() {
			So(IsKnownEventType(EventUserRegistered), ShouldBeTrue)
			So(IsKnownEventType(EventEmailChange), ShouldBeTrue)
			So(IsKnownEventType(EventAccountRecovery), ShouldBeTrue)
			So(IsKnownEventType(EventPasswordReset), ShouldBeTrue)
			So(IsKnownEventType(EventBriefSubmitted), ShouldBeTrue)
			So(IsKnownEventType(EventBookingCreated), ShouldBeTrue)
			So(IsKnownEventType(EventPaymentCompleted), ShouldBeTrue)
			So(IsKnownEventType(EventNotificationSend), ShouldBeTrue)
			So(IsKnownEventType(EventRegistrationAttemptOnExistingEmail), ShouldBeTrue)
		})

		Convey("returns false for unknown event types", func() {
			So(IsKnownEventType("unknown.event"), ShouldBeFalse)
		})
	})
}

func TestIsKnownAggregationType(t *testing.T) {
	Convey("IsKnownAggregationType", t, func() {
		Convey("returns true for known event types", func() {
			So(IsKnownAggregationType(AggregationTypeUser), ShouldBeTrue)
			So(IsKnownAggregationType(AggregationTypeEmail), ShouldBeTrue)
			So(IsKnownAggregationType(AggregationTypeAccount), ShouldBeTrue)
			So(IsKnownAggregationType(AggregationTypePassword), ShouldBeTrue)
			So(IsKnownAggregationType(AggregationTypeBrief), ShouldBeTrue)
			So(IsKnownAggregationType(AggregationTypeBooking), ShouldBeTrue)
			So(IsKnownAggregationType(AggregationTypePayment), ShouldBeTrue)
			So(IsKnownAggregationType(AggregationTypeNotification), ShouldBeTrue)
		})

		Convey("returns false for unknown event types", func() {
			So(IsKnownAggregationType("unknown.event"), ShouldBeFalse)
		})
	})
}
