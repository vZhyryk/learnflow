package events

import "time"

// UserRegisteredPayload is the event payload emitted when a new user completes registration.
type UserRegisteredPayload struct {
	UserID    string    `json:"user_id"`
	Email     string    `json:"email"`
	RawToken  string    `json:"raw_token"`
	UserName  string    `json:"user_name"`
	ExpiresAt time.Time `json:"expires_at"`
}

// InitEmailChangeToken carries the token data for an email-change initiation event.
type InitEmailChangeToken struct {
	UserID    string    `json:"user_id"`
	Email     string    `json:"email"`
	RawToken  string    `json:"raw_token"`
	UserName  string    `json:"user_name"`
	ExpiresAt time.Time `json:"expires_at"`
}

// RegistrationAttemptPayload carries data for a registration attempt event.
type RegistrationAttemptPayload struct {
	UserID   string `json:"user_id"`
	Email    string `json:"email"`
	UserName string `json:"user_name"`
}

// InitPasswordResetToken carries the token data for a password-reset initiation event.
type InitPasswordResetToken struct {
	UserID    string    `json:"user_id"`
	Email     string    `json:"email"`
	RawToken  string    `json:"raw_token"`
	UserName  string    `json:"user_name"`
	ExpiresAt time.Time `json:"expires_at"`
}

// InitAccountRecoveryToken carries the token data for an account-recovery initiation event.
type InitAccountRecoveryToken struct {
	UserID    string    `json:"user_id"`
	Email     string    `json:"email"`
	RawToken  string    `json:"raw_token"`
	UserName  string    `json:"user_name"`
	ExpiresAt time.Time `json:"expires_at"`
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
