package admindomain

import (
	"context"
	"learnflow_backend/internal/shared/pagination"
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
}
