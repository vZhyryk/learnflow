package adminservice

import (
	"context"
	"fmt"
	admindomain "learnflow_backend/internal/admin/domain"
	auditdomain "learnflow_backend/internal/audit/domain"
	"learnflow_backend/internal/events"
	"learnflow_backend/internal/shared/pagination"
)

// CreateAnnouncement creates a new announcement and returns its ID.
func (srv *Service) CreateAnnouncement(ctx context.Context, req admindomain.CreateAnnouncementRequest) (announcementID string, err error) {
	err = srv.transactor.InTransaction(ctx, func(ctx context.Context) error {
		announcement, createErr := srv.announRepo.CreateAnnouncement(ctx, &admindomain.Announcement{
			Title:           req.Title,
			Body:            req.Body,
			EntityID:        req.EntityID,
			EntityType:      req.EntityType,
			Channels:        req.Channels,
			ExpiresAt:       req.ExpiresAt,
			CreatedByUserID: req.CreatedByUserID,
		})

		if createErr != nil {
			return fmt.Errorf("service.CreateAnnouncement: %w", createErr)
		}

		announcementID = announcement.ID

		return srv.actionRepo.CreateAdminAction(ctx, &auditdomain.AdminAction{
			AdminUserID: req.CreatedByUserID,
			ActionType:  auditdomain.ActionCreateItem,
			TargetType:  auditdomain.TargetAnnouncement,
			TargetID:    announcementID,
		})
	})

	return announcementID, err
}

// UpdateAnnouncement applies req to an existing announcement.
func (srv *Service) UpdateAnnouncement(ctx context.Context, req admindomain.UpdateAnnouncementRequest) error {
	return srv.transactor.InTransaction(ctx, func(ctx context.Context) error {
		announcement, err := srv.announRepo.GetAnnouncementByID(ctx, req.ID)
		if err != nil {
			return fmt.Errorf("service.UpdateAnnouncement: %w", err)
		}

		if announcement.ApprovedAt != nil {
			return admindomain.ErrAnnouncementApproved
		}

		req.Apply(announcement)

		if (announcement.EntityType == nil) != (announcement.EntityID == nil) {
			return admindomain.ErrEntityDataMisMatch
		}

		if err := srv.announRepo.UpdateAnnouncement(ctx, announcement); err != nil {
			return fmt.Errorf("service.UpdateAnnouncement: %w", err)
		}

		return srv.actionRepo.CreateAdminAction(ctx, &auditdomain.AdminAction{
			AdminUserID: req.UpdatedByUserID,
			ActionType:  auditdomain.ActionUpdateItem,
			TargetType:  auditdomain.TargetAnnouncement,
			TargetID:    req.ID,
		})
	})
}

// ApproveAnnouncement marks an announcement as approved by the given user.
func (srv *Service) ApproveAnnouncement(ctx context.Context, announcementID, userID string) error {
	return srv.transactor.InTransaction(ctx, func(ctx context.Context) error {
		err := srv.announRepo.ApproveAnnouncement(ctx, announcementID, userID)
		if err != nil {
			return fmt.Errorf("service.ApproveAnnouncement: %w", err)
		}

		payload := events.AnnouncementPayload{
			AnnouncementID: announcementID,
		}

		err = srv.actionRepo.CreateAdminAction(ctx, &auditdomain.AdminAction{
			AdminUserID: userID,
			ActionType:  auditdomain.ActionApproveItem,
			TargetType:  auditdomain.TargetAnnouncement,
			TargetID:    announcementID,
		})

		if err != nil {
			return fmt.Errorf("service.ApproveAnnouncement: %w", err)
		}

		return srv.outbox.Emit(ctx, events.AggregationTypeAnnouncement, announcementID, events.EventAnnouncementApprove, payload)
	})
}

// GetAnnouncements returns a paginated list of all announcements.
func (srv *Service) GetAnnouncements(ctx context.Context, params pagination.Params) ([]*admindomain.Announcement, error) {
	list, err := srv.announRepo.GetAnnouncements(ctx, params)
	if err != nil {
		return nil, fmt.Errorf("service.GetAnnouncements: %w", err)
	}

	return list, nil
}

// GetUnApprovedAnnouncements returns a paginated list of announcements pending approval.
func (srv *Service) GetUnApprovedAnnouncements(ctx context.Context, params pagination.Params) ([]*admindomain.Announcement, error) {
	list, err := srv.announRepo.GetUnApprovedAnnouncements(ctx, params)
	if err != nil {
		return nil, fmt.Errorf("service.GetUnApprovedAnnouncements: %w", err)
	}

	return list, nil
}

// GetApprovedAnnouncements returns a paginated list of approved announcements.
func (srv *Service) GetApprovedAnnouncements(ctx context.Context, params pagination.Params) ([]*admindomain.Announcement, error) {
	list, err := srv.announRepo.GetApprovedAnnouncements(ctx, params)
	if err != nil {
		return nil, fmt.Errorf("service.GetApprovedAnnouncements: %w", err)
	}

	return list, nil
}

// GetPublicAnnouncements returns a paginated list of approved banner announcements visible to userID.
func (srv *Service) GetPublicAnnouncements(ctx context.Context, params pagination.Params, userID string) ([]*admindomain.AnnouncementPublic, error) {
	list, err := srv.announRepo.GetPublicAnnouncements(ctx, params, userID)
	if err != nil {
		return nil, fmt.Errorf("service.GetPublicAnnouncements: %w", err)
	}

	return list, nil
}

// GetExpiredAnnouncements returns a paginated list of expired announcements.
func (srv *Service) GetExpiredAnnouncements(ctx context.Context, params pagination.Params) ([]*admindomain.Announcement, error) {
	list, err := srv.announRepo.GetExpiredAnnouncements(ctx, params)
	if err != nil {
		return nil, fmt.Errorf("service.GetExpiredAnnouncements: %w", err)
	}

	return list, nil
}
