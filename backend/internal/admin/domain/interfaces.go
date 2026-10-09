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
	SetExpiredNowAnnouncement(ctx context.Context, announcementID string) error
}

// AdminActionRepository persists the admin audit trail.
type AdminActionRepository interface {
	CreateAdminAction(ctx context.Context, action *auditdomain.AdminAction) error
	GetInstanceAdminActions(ctx context.Context, targetType auditdomain.AdminTargetType, targetID string, params pagination.Params) ([]*auditdomain.AdminAction, int, error)
	GetAdminActions(ctx context.Context, filter auditdomain.AdminActionFilter, params pagination.Params) ([]*auditdomain.AdminAction, int, error)
	GetFailedJobs(ctx context.Context, params pagination.Params) ([]*auditdomain.FailedJob, int, error)
}

// SessionRepository revokes user sessions on behalf of an admin.
type SessionRepository interface {
	RevokeAllUserSessionsAdmin(ctx context.Context, userID, revokedByUserID string) error
}

// CourseRepository looks up courses for admin access grants.
type CourseRepository interface {
	GetCourseTitleByID(ctx context.Context, courseID string) (string, error)
}

// ContentItemRepository looks up content items for admin access grants.
type ContentItemRepository interface {
	GetContentItemTitleByID(ctx context.Context, contentItemID string) (string, error)
}

// UserBlockList marks users as blocked (or clears the mark) so their already-issued access tokens are rejected.
type UserBlockList interface {
	BlockUser(ctx context.Context, userID string, ttl time.Duration) error
	UnBlockUser(ctx context.Context, userID string) error
	RevokeUserRole(ctx context.Context, userID string, ttl time.Duration) error
	ClearUserRoleRevoked(ctx context.Context, userID string) error
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
	GrantUserCourseAccess(ctx context.Context, userID, courseID string) error
	GrantUserContentAccess(ctx context.Context, userID, contentItemID string) error
}

// Service defines the admin module's business logic: announcements and user management.
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
	ChangeUserField(ctx context.Context, userID, adminID string, operationName UserAdminOperation) error
	GrantUserCourseAccess(ctx context.Context, userID, courseID, adminID string, operationName auditdomain.AdminActionType) error
	GrantUserContentAccess(ctx context.Context, userID, contentItemID, adminID string, operationName auditdomain.AdminActionType) error
	GetInstanceAdminActions(ctx context.Context, actorID string, targetType auditdomain.AdminTargetType, targetID string, params pagination.Params) (actions []*auditdomain.AdminAction, count int, err error)
	GetAdminActions(ctx context.Context, actorID string, filter auditdomain.AdminActionFilter, params pagination.Params) (actions []*auditdomain.AdminAction, count int, err error)
	GetFailedJobs(ctx context.Context, params pagination.Params) (failedJobs []*auditdomain.FailedJob, count int, err error)
	SetExpiredNowAnnouncement(ctx context.Context, announcementID, userID string) error
}
