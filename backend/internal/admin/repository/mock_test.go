package adminrepository

import (
	"fmt"
	admindomain "learnflow_backend/internal/admin/domain"
	"learnflow_backend/internal/shared/repository"
	"learnflow_backend/internal/shared/testutil"
	"time"

	"github.com/jackc/pgx/v5/pgtype"
)

func newTestRepo(runner *testutil.MockQueryRunner) *Repository {
	return &Repository{repository.BaseRepository{DB: runner}}
}

func fakeAnnouncement(now time.Time) *admindomain.Announcement {
	updatedByUserID := "user-456"
	approvedByUserID := "user-456"
	entityID := "course-123"
	entityType := admindomain.CourseEntityType

	return &admindomain.Announcement{
		ID:               "announcement-123",
		Title:            "Some Title",
		Body:             "Some body",
		CreatedAt:        now,
		CreatedByUserID:  "user-123",
		UpdatedAt:        &now,
		UpdatedByUserID:  &updatedByUserID,
		ApprovedAt:       &now,
		ApprovedByUserID: &approvedByUserID,
		ExpiresAt:        now.Add(24 * time.Hour),
		EntityID:         &entityID,
		EntityType:       &entityType,
		Channels:         []admindomain.Channel{admindomain.EmailChannel, admindomain.BannerChannel},
	}
}

// fakeAnnouncementScan simulates rows.Scan populating an Announcement from column order,
// matching scanAnnouncement in scanner.go.
func fakeAnnouncementScan(a *admindomain.Announcement) func(dest ...any) error {
	return func(dest ...any) error {
		*testutil.CastStr(dest[0], 0) = a.ID
		*testutil.CastStr(dest[1], 1) = a.Title
		*testutil.CastStr(dest[2], 2) = a.Body
		*testutil.CastTime(dest[3], 3) = a.CreatedAt
		*testutil.CastStr(dest[4], 4) = a.CreatedByUserID
		*testutil.CastPtrTime(dest[5], 5) = a.UpdatedAt
		*testutil.CastPtrStr(dest[6], 6) = a.UpdatedByUserID
		*testutil.CastPtrTime(dest[7], 7) = a.ApprovedAt
		*testutil.CastPtrStr(dest[8], 8) = a.ApprovedByUserID
		*testutil.CastTime(dest[9], 9) = a.ExpiresAt
		*testutil.CastPtrStr(dest[10], 10) = a.EntityID
		*testutil.CastEnum[*admindomain.EntityType](dest[11], 11) = a.EntityType
		*testutil.CastEnum[[]admindomain.Channel](dest[12], 12) = a.Channels
		return nil
	}
}

func fakeAnnouncementPublic(n int) *admindomain.AnnouncementPublic {
	entityID := "course-1"
	entityType := admindomain.CourseEntityType
	return &admindomain.AnnouncementPublic{
		ID:         fmt.Sprintf("announcement-%d", n),
		Title:      "title",
		Body:       "body",
		ApprovedAt: time.Now().UTC().Truncate(time.Second),
		ExpiresAt:  time.Now().Add(24 * time.Hour).UTC().Truncate(time.Second),
		EntityID:   &entityID,
		EntityType: &entityType,
	}
}

func fakeAnnouncementPublicScan(a *admindomain.AnnouncementPublic) func(dest ...any) error {
	return func(dest ...any) error {
		*testutil.CastStr(dest[0], 0) = a.ID
		*testutil.CastStr(dest[1], 1) = a.Title
		*testutil.CastStr(dest[2], 2) = a.Body
		*testutil.CastTime(dest[3], 3) = a.ApprovedAt
		*testutil.CastTime(dest[4], 4) = a.ExpiresAt
		*testutil.CastPtrStr(dest[5], 5) = a.EntityID
		*testutil.CastEnum[*admindomain.EntityType](dest[6], 6) = a.EntityType
		return nil
	}
}

func fakeUserData(n int) *admindomain.UserData {
	firstName := fmt.Sprintf("First%d", n)
	dob := "1990-05-17"
	now := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)

	return &admindomain.UserData{
		UserID:      fmt.Sprintf("user-%d", n),
		FirstName:   &firstName,
		DateOfBirth: &dob,
		CreatedAt:   now,
		Status:      admindomain.StatusActive,
		Role:        admindomain.RoleUser,
	}
}

// fakeUserDataScan simulates rows.Scan populating a UserData from column order,
// matching scanUserData in scanner.go.
func fakeUserDataScan(u *admindomain.UserData) func(dest ...any) error {
	return func(dest ...any) error {
		*testutil.CastStr(dest[0], 0) = u.UserID
		*testutil.CastPtrStr(dest[1], 1) = u.FirstName
		*testutil.CastPtrStr(dest[2], 2) = u.LastName
		*testutil.CastPtrStr(dest[3], 3) = u.PhoneNumber
		*testutil.CastPtrStr(dest[4], 4) = u.Country
		*testutil.CastPtrStr(dest[5], 5) = u.City
		if u.DateOfBirth != nil {
			parsed, err := time.Parse(dobLayout, *u.DateOfBirth)
			if err != nil {
				return fmt.Errorf("parse date of birth: %w", err)
			}
			*testutil.CastPgtypeDate(dest[6], 6) = pgtype.Date{Time: parsed, Valid: true}
		}
		*testutil.CastPtrStr(dest[7], 7) = u.Gender
		*testutil.CastPtrStr(dest[8], 8) = u.AvatarURL
		*testutil.CastPtrStr(dest[9], 9) = u.Bio
		*testutil.CastTime(dest[10], 10) = u.CreatedAt
		*testutil.CastPtrTime(dest[11], 11) = u.DeletedAt
		*testutil.CastPtrTime(dest[12], 12) = u.LastLoginAt
		*testutil.CastEnum[admindomain.UserStatus](dest[13], 13) = u.Status
		*testutil.CastEnum[admindomain.UserRole](dest[14], 14) = u.Role
		return nil
	}
}
