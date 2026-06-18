package events

// UserRegisteredPayload is the event payload emitted when a new user completes registration.
type UserRegisteredPayload struct {
	UserID string `json:"user_id"`
	Email  string `json:"email"`
	URL    string `json:"verification_url"`
}

// BriefSubmittedPayload is the event payload emitted when a consultation brief is submitted.
type BriefSubmittedPayload struct {
	BriefID string `json:"brief_id"`
	UserID  string `json:"user_id"`
}

// NotificationSendPayload is the event payload emitted when a notification should be delivered.
type NotificationSendPayload struct {
	UserID   string            `json:"user_id"`
	Template string            `json:"template"`
	Data     map[string]string `json:"data"`
}

// BookingCreatedPayload is the event payload emitted when a booking is created.
type BookingCreatedPayload struct {
	BookingID string `json:"booking_id"`
	UserID    string `json:"user_id"`
}

// PaymentCompletedPayload is the event payload emitted when a payment is completed.
type PaymentCompletedPayload struct {
	PaymentID   string `json:"payment_id"`
	UserID      string `json:"user_id"`
	AmountCents int64  `json:"amount_cents"`
}
