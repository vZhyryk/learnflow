package events

import (
	"crypto/sha256"
	"encoding/hex"
	"time"
)

// Field is one required payload field: its name and whether its value is valid.
type Field struct {
	Name    string
	Value   string
	IsValid func() bool
}

// TokenPayload carries a one-time token link email (verification, reset, recovery, email change).
type TokenPayload struct {
	UserID    string    `json:"user_id"`
	Email     string    `json:"email"`
	RawToken  string    `json:"raw_token"`
	UserName  string    `json:"user_name"`
	ExpiresAt time.Time `json:"expires_at"`
}

// RequiredFields lists the fields that must be valid before the email is sent.
func (p TokenPayload) RequiredFields() []Field {
	return []Field{
		{Name: "UserID", IsValid: func() bool { return p.UserID != "" }},
		{Name: "Email", IsValid: func() bool { return p.Email != "" }},
		{Name: "RawToken", IsValid: func() bool { return p.RawToken != "" }},
		{Name: "UserName", IsValid: func() bool { return p.UserName != "" }},
		{Name: "ExpiresAt", IsValid: func() bool { return !p.ExpiresAt.IsZero() }},
	}
}

// GetIdempotencyKey returns the parts of the dedupe key: user and the token's SHA-256, never the raw token.
func (p TokenPayload) GetIdempotencyKey() []string {
	sum := sha256.Sum256([]byte(p.RawToken))
	return []string{p.UserID, hex.EncodeToString(sum[:])}
}

// GetEmail returns the recipient address.
func (p TokenPayload) GetEmail() string {
	return p.Email
}

// RegistrationAttemptPayload carries data for a registration attempt event.
type RegistrationAttemptPayload struct {
	UserID   string `json:"user_id"`
	Email    string `json:"email"`
	UserName string `json:"user_name"`
}

// GetEmail returns the recipient address.
func (p RegistrationAttemptPayload) GetEmail() string {
	return p.Email
}

// RequiredFields lists the fields that must be valid before the email is sent.
func (p RegistrationAttemptPayload) RequiredFields() []Field {
	return []Field{
		{Name: "UserID", IsValid: func() bool { return p.UserID != "" }},
		{Name: "Email", IsValid: func() bool { return p.Email != "" }},
		{Name: "UserName", IsValid: func() bool { return p.UserName != "" }},
	}
}

// GetIdempotencyKey returns the parts of the dedupe key: the user only, so at most one warning is sent
// per user per 24h and repeated attempts cannot flood the victim's inbox.
func (p RegistrationAttemptPayload) GetIdempotencyKey() []string {
	return []string{p.UserID}
}

// BriefSubmittedPayload is the event payload emitted when a consultation brief is submitted.
type BriefSubmittedPayload struct {
	BriefID string `json:"brief_id"`
	UserID  string `json:"user_id"`
}

// AnnouncementPayload is the event payload emitted when an announcement is approved.
type AnnouncementPayload struct {
	AnnouncementID string `json:"announcement_id"`
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

// UserNotificationPayload carries an account status email (block, unblock, delete, restore); EventID makes each action unique.
type UserNotificationPayload struct {
	EventID  string `json:"event_id"`
	UserID   string `json:"user_id"`
	Email    string `json:"email"`
	UserName string `json:"user_name"`
}

// RequiredFields lists the fields that must be valid before the email is sent.
func (p UserNotificationPayload) RequiredFields() []Field {
	return []Field{
		{Name: "UserID", IsValid: func() bool { return p.UserID != "" }},
		{Name: "Email", IsValid: func() bool { return p.Email != "" }},
		{Name: "UserName", IsValid: func() bool { return p.UserName != "" }},
		{Name: "EventID", IsValid: func() bool { return p.EventID != "" }},
	}
}

// GetIdempotencyKey returns the parts of the dedupe key: user and event.
func (p UserNotificationPayload) GetIdempotencyKey() []string {
	return []string{p.UserID, p.EventID}
}

// GetEmail returns the recipient address.
func (p UserNotificationPayload) GetEmail() string {
	return p.Email
}

type GrantAccessPayload struct {
	UserID   string `json:"user_id"`
	ItemName string `json:"item_name"`
	ItemID   string `json:"item_id"`
	ItemType string `json:"item_type"`
	UserName string `json:"user_name"`
	Email    string `json:"email"`
}

func (p GrantAccessPayload) RequiredFields() []Field {
	return []Field{
		{Name: "UserID", IsValid: func() bool { return p.UserID != "" }},
		{Name: "ItemID", IsValid: func() bool { return p.ItemID != "" }},
		{Name: "ItemType", IsValid: func() bool { return p.ItemType != "" }},
		{Name: "ItemName", IsValid: func() bool { return p.ItemName != "" }},
		{Name: "UserName", IsValid: func() bool { return p.UserName != "" }},
		{Name: "Email", IsValid: func() bool { return p.Email != "" }},
	}
}

func (p GrantAccessPayload) GetIdempotencyKey() []string {
	return []string{p.UserID, p.ItemID, p.ItemType}
}

func (p GrantAccessPayload) GetEmail() string {
	return p.Email
}
