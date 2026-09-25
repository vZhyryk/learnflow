package admindomain

import "time"

// UserRole represents the permission level of a user account.
type UserRole string

// UserStatus represents the current state of a user account.
type UserStatus string

// Role constants.
const (
	RoleAdmin    UserRole = "admin"
	RoleSubAdmin UserRole = "subadmin"
	RoleUser     UserRole = "user"
)

// Status constants.
const (
	StatusActive              UserStatus = "active"
	StatusBlocked             UserStatus = "blocked"
	StatusDeleted             UserStatus = "deleted"
	StatusPendingVerification UserStatus = "pending_verification"
)

// UserData is the admin view of a user account joined with its profile.
type UserData struct {
	UserID      string     `json:"user_id"`
	FirstName   *string    `json:"first_name"`
	LastName    *string    `json:"last_name"`
	PhoneNumber *string    `json:"phone_number"`
	Country     *string    `json:"country"`
	City        *string    `json:"city"`
	DateOfBirth *string    `json:"date_of_birth"`
	Gender      *string    `json:"gender"`
	AvatarURL   *string    `json:"avatar_url"`
	Bio         *string    `json:"bio"`
	DeletedAt   *time.Time `json:"deleted_at"`
	CreatedAt   time.Time  `json:"created_at"`
	LastLoginAt *time.Time `json:"last_login_at"`
	// Purchases   []any      `json:"purchases"`
	Status UserStatus `json:"status"`
	Role   UserRole   `json:"role"`
}
