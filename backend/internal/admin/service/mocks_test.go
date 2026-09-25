package adminservice

import (
	"context"
	admindomain "learnflow_backend/internal/admin/domain"
	auditdomain "learnflow_backend/internal/audit/domain"
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
	getPublicAnnouncements     func(ctx context.Context, params pagination.Params, userID string) ([]*admindomain.AnnouncementPublic, error)
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

func (m *mockAnnouncementRepo) GetPublicAnnouncements(ctx context.Context, params pagination.Params, userID string) ([]*admindomain.AnnouncementPublic, error) {
	if m.getPublicAnnouncements == nil {
		panic("mockAnnouncementRepo.GetPublicAnnouncements not set")
	}
	return m.getPublicAnnouncements(ctx, params, userID)
}

func newTestService(repo *mockAnnouncementRepo) *Service {
	return newTestServiceWithActions(repo, noopAdminActions())
}

func newTestServiceWithActions(repo *mockAnnouncementRepo, actions *mockAdminActionRepo) *Service {
	return New(repo, &mockUserRepo{}, actions, &testutil.NoopTransactor{}, testutil.NewNoopOutbox())
}

// mockUserRepo implements admindomain.UserRepository via function fields.
type mockUserRepo struct {
	revokeUserRole  func(ctx context.Context, userID string) error
	assignUserRole  func(ctx context.Context, userID string) error
	deleteUser      func(ctx context.Context, userID string) error
	restoreUser     func(ctx context.Context, userID string) error
	blockUser       func(ctx context.Context, userID string) error
	unBlockUser     func(ctx context.Context, userID string) error
	getUsersData    func(ctx context.Context, params pagination.Params) ([]*admindomain.UserData, int, error)
	getUserDataByID func(ctx context.Context, userID string) (*admindomain.UserData, error)
}

func (m *mockUserRepo) RevokeUserRole(ctx context.Context, userID string) error {
	if m.revokeUserRole == nil {
		panic("mockUserRepo.RevokeUserRole not set")
	}
	return m.revokeUserRole(ctx, userID)
}

func (m *mockUserRepo) AssignUserRole(ctx context.Context, userID string) error {
	if m.assignUserRole == nil {
		panic("mockUserRepo.AssignUserRole not set")
	}
	return m.assignUserRole(ctx, userID)
}

func (m *mockUserRepo) DeleteUser(ctx context.Context, userID string) error {
	if m.deleteUser == nil {
		panic("mockUserRepo.DeleteUser not set")
	}
	return m.deleteUser(ctx, userID)
}

func (m *mockUserRepo) RestoreUser(ctx context.Context, userID string) error {
	if m.restoreUser == nil {
		panic("mockUserRepo.RestoreUser not set")
	}
	return m.restoreUser(ctx, userID)
}

func (m *mockUserRepo) BlockUser(ctx context.Context, userID string) error {
	if m.blockUser == nil {
		panic("mockUserRepo.BlockUser not set")
	}
	return m.blockUser(ctx, userID)
}

func (m *mockUserRepo) UnBlockUser(ctx context.Context, userID string) error {
	if m.unBlockUser == nil {
		panic("mockUserRepo.UnBlockUser not set")
	}
	return m.unBlockUser(ctx, userID)
}

func (m *mockUserRepo) GetUsersData(ctx context.Context, params pagination.Params) ([]*admindomain.UserData, int, error) {
	if m.getUsersData == nil {
		panic("mockUserRepo.GetUsersData not set")
	}
	return m.getUsersData(ctx, params)
}

func (m *mockUserRepo) GetUserDataByID(ctx context.Context, userID string) (*admindomain.UserData, error) {
	if m.getUserDataByID == nil {
		panic("mockUserRepo.GetUserDataByID not set")
	}
	return m.getUserDataByID(ctx, userID)
}

// mockAdminActionRepo implements admindomain.AdminActionRepository via function fields.
type mockAdminActionRepo struct {
	createAdminAction func(ctx context.Context, action *auditdomain.AdminAction) error
}

func (m *mockAdminActionRepo) CreateAdminAction(ctx context.Context, action *auditdomain.AdminAction) error {
	if m.createAdminAction == nil {
		panic("mockAdminActionRepo.CreateAdminAction not set")
	}
	return m.createAdminAction(ctx, action)
}

func noopAdminActions() *mockAdminActionRepo {
	return &mockAdminActionRepo{createAdminAction: func(_ context.Context, _ *auditdomain.AdminAction) error { return nil }}
}

func capturingAdminActions(got *[]*auditdomain.AdminAction, err error) *mockAdminActionRepo {
	return &mockAdminActionRepo{createAdminAction: func(_ context.Context, action *auditdomain.AdminAction) error {
		*got = append(*got, action)
		return err
	}}
}

func newTestUserService(users *mockUserRepo, actions *mockAdminActionRepo) *Service {
	return New(&mockAnnouncementRepo{}, users, actions, &testutil.NoopTransactor{}, testutil.NewNoopOutbox())
}
