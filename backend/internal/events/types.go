package events

// EventType identifies the kind of domain event being emitted.
type EventType string

// Domain event type constants.
const (
	EventUserRegistered   EventType = "user.registered"
	EventBriefSubmitted   EventType = "brief.submitted"
	EventBookingCreated   EventType = "booking.created"
	EventPaymentCompleted EventType = "payment.completed"
	EventNotificationSend EventType = "notification.send"
)

// IsKnownEventType reports whether t is a registered event type.
func IsKnownEventType(t EventType) bool {
	switch t {
	case EventUserRegistered, EventBriefSubmitted, EventBookingCreated,
		EventPaymentCompleted, EventNotificationSend:
		return true
	}
	return false
}

// AggregationType identifies the domain aggregate that owns the event.
type AggregationType string

// Aggregate type constants.
const (
	AggregationTypeUser         AggregationType = "user"
	AggregationTypeEmail        AggregationType = "email"
	AggregationTypeBrief        AggregationType = "brief"
	AggregationTypeBooking      AggregationType = "booking"
	AggregationTypePayment      AggregationType = "payment"
	AggregationTypeNotification AggregationType = "notification"
)

// IsKnownAggregationType reports whether t is a registered aggregate type.
func IsKnownAggregationType(t AggregationType) bool {
	switch t {
	case AggregationTypeUser, AggregationTypeBrief, AggregationTypeBooking,
		AggregationTypePayment, AggregationTypeNotification, AggregationTypeEmail:
		return true
	}
	return false
}
