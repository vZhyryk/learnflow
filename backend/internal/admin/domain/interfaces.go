package admindomain

import (
	"context"
	auditdomain "learnflow_backend/internal/audit/domain"
	"learnflow_backend/internal/shared/pagination"
	"time"
)

// Transactor executes a function within a database transaction.
type Transactor interface {
	InTransaction(ctx context.Context, fn func(ctx context.Context) error) error
}

// AnnouncementRepository defines persistence operations for Announcement.
type AnnouncementRepository interface {
	CreateAnnouncement(ctx context.Context, announcement *Announcement) (*Announcement, error)
	UpdateAnnouncement(ctx context.Context, announcement *Announcement) error
	ApproveAnnouncement(ctx context.Context, announcementID, userID string) error
	GetAnnouncements(ctx context.Context, params pagination.Params) ([]*Announcement, error)
	GetUnApprovedAnnouncements(ctx context.Context, params pagination.Params) ([]*Announcement, error)
	GetApprovedAnnouncements(ctx context.Context, params pagination.Params) ([]*Announcement, error)
	GetExpiredAnnouncements(ctx context.Context, params pagination.Params) ([]*Announcement, error)
	GetAnnouncementByID(ctx context.Context, id string) (*Announcement, error)
	GetPublicAnnouncements(ctx context.Context, params pagination.Params, userID string) ([]*AnnouncementPublic, error)
}

// AdminActionRepository persists the admin audit trail.
type AdminActionRepository interface {
	CreateAdminAction(ctx context.Context, action *auditdomain.AdminAction) error
}

// SessionRepository revokes user sessions on behalf of an admin.
type SessionRepository interface {
	RevokeAllUserSessionsAdmin(ctx context.Context, userID string, revokedByUserID string) error
}

// UserBlocklist marks users as blocked (or clears the mark) so their already-issued access tokens are rejected.
type UserBlocklist interface {
	BlockUser(ctx context.Context, userID string, ttl time.Duration) error
	UnBlockUser(ctx context.Context, userID string) error
}

// UserRepository defines admin persistence operations for user accounts.
type UserRepository interface {
	RevokeUserRole(ctx context.Context, userID string) error
	AssignUserRole(ctx context.Context, userID string) error
	DeleteUser(ctx context.Context, userID string) error
	RestoreUser(ctx context.Context, userID string) error
	BlockUser(ctx context.Context, userID string) error
	UnBlockUser(ctx context.Context, userID string) error
	GetUsersData(ctx context.Context, params pagination.Params) ([]*UserData, int, error)
	GetUserDataByID(ctx context.Context, userID string) (*UserData, error)
}

// Service defines the admin module's announcement business logic.
type Service interface {
	CreateAnnouncement(ctx context.Context, req CreateAnnouncementRequest) (string, error)
	UpdateAnnouncement(ctx context.Context, req UpdateAnnouncementRequest) error
	ApproveAnnouncement(ctx context.Context, announcementID, userID string) error
	GetAnnouncements(ctx context.Context, params pagination.Params) ([]*Announcement, error)
	GetUnApprovedAnnouncements(ctx context.Context, params pagination.Params) ([]*Announcement, error)
	GetApprovedAnnouncements(ctx context.Context, params pagination.Params) ([]*Announcement, error)
	GetExpiredAnnouncements(ctx context.Context, params pagination.Params) ([]*Announcement, error)
	GetPublicAnnouncements(ctx context.Context, params pagination.Params, userID string) ([]*AnnouncementPublic, error)

	GetUsersData(ctx context.Context, params pagination.Params) ([]*UserData, int, error)
	GetUserDataByID(ctx context.Context, userID string) (*UserData, error)
	ChangeUserField(ctx context.Context, operationName, userID, adminID string) error
}
