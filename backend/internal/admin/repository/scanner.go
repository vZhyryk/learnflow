package adminrepository

import (
	admindomain "learnflow_backend/internal/admin/domain"
	"learnflow_backend/internal/infrastructure/convert"
	"learnflow_backend/internal/shared/repository"

	"github.com/jackc/pgx/v5/pgtype"
)

const dobLayout = "2006-01-02"

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

func scanUserData(row repository.RowScanner) (*admindomain.UserData, error) {
	userData := &admindomain.UserData{}
	var dob pgtype.Date
	err := row.Scan(
		&userData.UserID,
		&userData.FirstName,
		&userData.LastName,
		&userData.PhoneNumber,
		&userData.Country,
		&userData.City,
		&dob,
		&userData.Gender,
		&userData.AvatarURL,
		&userData.Bio,
		&userData.CreatedAt,
		&userData.DeletedAt,
		&userData.LastLoginAt,
		&userData.Status,
		&userData.Role,
	)
	if err != nil {
		return nil, err
	}
	userData.DateOfBirth = convert.FormatNullableDate(dob, dobLayout)
	return userData, nil
}
