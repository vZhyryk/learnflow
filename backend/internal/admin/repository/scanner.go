package adminrepository

import (
	admindomain "learnflow_backend/internal/admin/domain"
	"learnflow_backend/internal/shared/repository"
)

func scanAnnouncement(row repository.RowScanner) (*admindomain.Announcement, error) {
	announcement := &admindomain.Announcement{}
	err := row.Scan(
		&announcement.ID,
		&announcement.Title,
		&announcement.Body,
		&announcement.CreatedAt,
		&announcement.CreatedByUserID,
		&announcement.UpdatedAt,
		&announcement.UpdatedByUserID,
		&announcement.ApprovedAt,
		&announcement.ApprovedByUserID,
		&announcement.ExpiresAt,
		&announcement.EntityID,
		&announcement.EntityType,
		&announcement.Channels,
	)
	if err != nil {
		return nil, err
	}
	return announcement, nil
}

func scanPublicAnnouncement(row repository.RowScanner) (*admindomain.AnnouncementPublic, error) {
	announcement := &admindomain.AnnouncementPublic{}
	err := row.Scan(
		&announcement.ID,
		&announcement.Title,
		&announcement.Body,
		&announcement.ApprovedAt,
		&announcement.ExpiresAt,
		&announcement.EntityID,
		&announcement.EntityType,
	)
	if err != nil {
		return nil, err
	}
	return announcement, nil
}
