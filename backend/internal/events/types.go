package events

// EventType identifies the kind of domain event being emitted.
type EventType string

// Domain event type constants.
const (
	EventUserRegistered                     EventType = "user.registered"
	EventUserBlocked                        EventType = "user.blocked"
	EventUserUnBlocked                      EventType = "user.unblocked"
	EventUserDeleted                        EventType = "user.deleted"
	EventUserRestored                       EventType = "user.restored"
	EventEmailChange                        EventType = "email.change"
	EventAccountRecovery                    EventType = "account.recovery"
	EventPasswordReset                      EventType = "password.reset"
	EventBriefSubmitted                     EventType = "brief.submitted"
	EventBookingCreated                     EventType = "booking.created"
	EventPaymentCompleted                   EventType = "payment.completed"
	EventNotificationSend                   EventType = "notification.send"
	EventAnnouncementApprove                EventType = "announcement.approved"
	EventRegistrationAttemptOnExistingEmail EventType = "user.existed.register"
	EventAnnouncementDeliver                EventType = "announcement.deliver"
	EventGrantAccess                        EventType = "user.grant.access"
)

// IsKnownEventType reports whether t is a registered event type.
func IsKnownEventType(t EventType) bool {
	switch t {
	case
		EventUserRegistered,
		EventUserBlocked,
		EventUserUnBlocked,
		EventUserRestored,
		EventUserDeleted,
		EventEmailChange,
		EventAccountRecovery,
		EventPasswordReset,
		EventBriefSubmitted,
		EventBookingCreated,
		EventPaymentCompleted,
		EventNotificationSend,
		EventAnnouncementApprove,
		EventRegistrationAttemptOnExistingEmail,
		EventAnnouncementDeliver,
		EventGrantAccess:
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
	AggregationTypeAccount      AggregationType = "account"
	AggregationTypePassword     AggregationType = "password"
	AggregationTypeBrief        AggregationType = "brief"
	AggregationTypeBooking      AggregationType = "booking"
	AggregationTypePayment      AggregationType = "payment"
	AggregationTypeAnnouncement AggregationType = "announcement"
)

// IsKnownAggregationType reports whether t is a registered aggregate type.
func IsKnownAggregationType(t AggregationType) bool {
	switch t {
	case
		AggregationTypeUser,
		AggregationTypeEmail,
		AggregationTypeAccount,
		AggregationTypePassword,
		AggregationTypeBrief,
		AggregationTypeBooking,
		AggregationTypePayment,
		AggregationTypeAnnouncement:
		return true
	}
	return false
}
