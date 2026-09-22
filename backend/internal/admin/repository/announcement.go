package adminrepository

import (
	"context"
	"errors"
	"fmt"
	admindomain "learnflow_backend/internal/admin/domain"
	"learnflow_backend/internal/infrastructure/db"
	"learnflow_backend/internal/shared/pagination"
	"learnflow_backend/internal/shared/repository"

	"github.com/jackc/pgx/v5"
)

const announcementEntityPairingCheckConstraint = "announcements_entity_pairing_check"

// CreateAnnouncement persists a new announcement.
func (rep *Repository) CreateAnnouncement(ctx context.Context, announcement *admindomain.Announcement) (*admindomain.Announcement, error) {
	result, err := scanAnnouncement(rep.QueryRunner(ctx).QueryRow(ctx, createAnnouncementSQL, announcement.Title, announcement.Body, announcement.CreatedByUserID, announcement.ExpiresAt, announcement.EntityID, announcement.EntityType, announcement.Channels))
	if db.IsCheckViolation(err, announcementEntityPairingCheckConstraint) {
		return nil, admindomain.ErrEntityDataMisMatch
	}
	if err != nil {
		return nil, fmt.Errorf("repository.CreateAnnouncement: %w", err)
	}

	return result, nil
}

// UpdateAnnouncement updates an existing announcement.
func (rep *Repository) UpdateAnnouncement(ctx context.Context, announcement *admindomain.Announcement) error {
	tag, err := rep.QueryRunner(ctx).Exec(ctx, updateAnnouncementSQL, announcement.ID, announcement.Title, announcement.Body, announcement.EntityID, announcement.EntityType, announcement.Channels, announcement.UpdatedByUserID, announcement.ExpiresAt)
	if db.IsCheckViolation(err, announcementEntityPairingCheckConstraint) {
		return admindomain.ErrEntityDataMisMatch
	}
	if err != nil {
		return fmt.Errorf("repository.UpdateAnnouncement: %w", err)
	}

	if tag.RowsAffected() == 0 {
		return admindomain.ErrAnnouncementNotFound
	}

	return nil
}

// ApproveAnnouncement marks an announcement as approved by the given user.
func (rep *Repository) ApproveAnnouncement(ctx context.Context, announcementID, userID string) error {
	tag, err := rep.QueryRunner(ctx).Exec(ctx, approveAnnouncementSQL, announcementID, userID)
	if err != nil {
		return fmt.Errorf("repository.ApproveAnnouncement: %w", err)
	}

	if tag.RowsAffected() == 0 {
		return admindomain.ErrAnnouncementNotFound
	}

	return nil
}

// GetAnnouncements returns a paginated list of all announcements.
func (rep *Repository) GetAnnouncements(ctx context.Context, params pagination.Params) ([]*admindomain.Announcement, error) {
	return repository.GetAndParseList(ctx, &rep.BaseRepository, getAnnouncementsSQL, "GetAnnouncements", params, scanAnnouncement)
}

// GetUnApprovedAnnouncements returns a paginated list of announcements pending approval.
func (rep *Repository) GetUnApprovedAnnouncements(ctx context.Context, params pagination.Params) ([]*admindomain.Announcement, error) {
	return repository.GetAndParseList(ctx, &rep.BaseRepository, getUnApprovedAnnouncementsSQL, "GetUnApprovedAnnouncements", params, scanAnnouncement)
}

// GetApprovedAnnouncements returns a paginated list of approved announcements.
func (rep *Repository) GetApprovedAnnouncements(ctx context.Context, params pagination.Params) ([]*admindomain.Announcement, error) {
	return repository.GetAndParseList(ctx, &rep.BaseRepository, getApprovedAnnouncementsSQL, "GetApprovedAnnouncements", params, scanAnnouncement)
}

// GetExpiredAnnouncements returns a paginated list of expired announcements.
func (rep *Repository) GetExpiredAnnouncements(ctx context.Context, params pagination.Params) ([]*admindomain.Announcement, error) {
	return repository.GetAndParseList(ctx, &rep.BaseRepository, getExpiredAnnouncementsSQL, "GetExpiredAnnouncements", params, scanAnnouncement)
}

// GetPublicAnnouncements returns a paginated list of approved banner announcements userID may see.
func (rep *Repository) GetPublicAnnouncements(ctx context.Context, params pagination.Params, userID string) ([]*admindomain.AnnouncementPublic, error) {
	return repository.GetAndParseListWithArgs(ctx, &rep.BaseRepository, getPublicAnnouncementsSQL, "GetPublicAnnouncements", params, scanPublicAnnouncement, []any{userID})
}

// GetAnnouncementByID retrieves an announcement by ID.
func (rep *Repository) GetAnnouncementByID(ctx context.Context, id string) (*admindomain.Announcement, error) {
	result, err := scanAnnouncement(rep.QueryRunner(ctx).QueryRow(ctx, getAnnouncementByIDSQL, id))
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, admindomain.ErrAnnouncementNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("repository.GetAnnouncementByID: %w", err)
	}

	return result, nil
}
