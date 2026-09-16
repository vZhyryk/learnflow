package adminservice

import (
	"context"
	admindomain "learnflow_backend/internal/admin/domain"
	"learnflow_backend/internal/shared/pagination"
	"learnflow_backend/internal/shared/testutil"
)

// mockAnnouncementRepo implements admindomain.AnnouncementRepository via function fields.
type mockAnnouncementRepo struct {
	createAnnouncement         func(ctx context.Context, announcement *admindomain.Announcement) (*admindomain.Announcement, error)
	updateAnnouncement         func(ctx context.Context, announcement *admindomain.Announcement) error
	approveAnnouncement        func(ctx context.Context, announcementID, userID string) error
	getAnnouncements           func(ctx context.Context, params pagination.Params) ([]*admindomain.Announcement, error)
	getUnApprovedAnnouncements func(ctx context.Context, params pagination.Params) ([]*admindomain.Announcement, error)
	getApprovedAnnouncements   func(ctx context.Context, params pagination.Params) ([]*admindomain.Announcement, error)
	getExpiredAnnouncements    func(ctx context.Context, params pagination.Params) ([]*admindomain.Announcement, error)
	getAnnouncementByID        func(ctx context.Context, id string) (*admindomain.Announcement, error)
}

func (m *mockAnnouncementRepo) CreateAnnouncement(ctx context.Context, announcement *admindomain.Announcement) (*admindomain.Announcement, error) {
	if m.createAnnouncement == nil {
		panic("mockAnnouncementRepo.CreateAnnouncement not set")
	}
	return m.createAnnouncement(ctx, announcement)
}

func (m *mockAnnouncementRepo) UpdateAnnouncement(ctx context.Context, announcement *admindomain.Announcement) error {
	if m.updateAnnouncement == nil {
		panic("mockAnnouncementRepo.UpdateAnnouncement not set")
	}
	return m.updateAnnouncement(ctx, announcement)
}

func (m *mockAnnouncementRepo) ApproveAnnouncement(ctx context.Context, announcementID, userID string) error {
	if m.approveAnnouncement == nil {
		panic("mockAnnouncementRepo.ApproveAnnouncement not set")
	}
	return m.approveAnnouncement(ctx, announcementID, userID)
}

func (m *mockAnnouncementRepo) GetAnnouncements(ctx context.Context, params pagination.Params) ([]*admindomain.Announcement, error) {
	if m.getAnnouncements == nil {
		panic("mockAnnouncementRepo.GetAnnouncements not set")
	}
	return m.getAnnouncements(ctx, params)
}

func (m *mockAnnouncementRepo) GetUnApprovedAnnouncements(ctx context.Context, params pagination.Params) ([]*admindomain.Announcement, error) {
	if m.getUnApprovedAnnouncements == nil {
		panic("mockAnnouncementRepo.GetUnApprovedAnnouncements not set")
	}
	return m.getUnApprovedAnnouncements(ctx, params)
}

func (m *mockAnnouncementRepo) GetApprovedAnnouncements(ctx context.Context, params pagination.Params) ([]*admindomain.Announcement, error) {
	if m.getApprovedAnnouncements == nil {
		panic("mockAnnouncementRepo.GetApprovedAnnouncements not set")
	}
	return m.getApprovedAnnouncements(ctx, params)
}

func (m *mockAnnouncementRepo) GetExpiredAnnouncements(ctx context.Context, params pagination.Params) ([]*admindomain.Announcement, error) {
	if m.getExpiredAnnouncements == nil {
		panic("mockAnnouncementRepo.GetExpiredAnnouncements not set")
	}
	return m.getExpiredAnnouncements(ctx, params)
}

func (m *mockAnnouncementRepo) GetAnnouncementByID(ctx context.Context, id string) (*admindomain.Announcement, error) {
	if m.getAnnouncementByID == nil {
		panic("mockAnnouncementRepo.GetAnnouncementByID not set")
	}
	return m.getAnnouncementByID(ctx, id)
}

func newTestService(repo *mockAnnouncementRepo) *Service {
	return New(repo, &testutil.NoopTransactor{}, testutil.NewNoopOutbox())
}
