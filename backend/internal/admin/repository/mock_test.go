package adminrepository

import (
	"fmt"
	admindomain "learnflow_backend/internal/admin/domain"
	"learnflow_backend/internal/shared/repository"
	"learnflow_backend/internal/shared/testutil"
	"time"
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
