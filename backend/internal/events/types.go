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

// AggregationType identifies the domain aggregate that owns the event.
type AggregationType string

// Aggregate type constants.
const (
	AggregationTypeUser         AggregationType = "user"
	AggregationTypeBrief        AggregationType = "brief"
	AggregationTypeBooking      AggregationType = "booking"
	AggregationTypePayment      AggregationType = "payment"
	AggregationTypeNotification AggregationType = "notification"
)
