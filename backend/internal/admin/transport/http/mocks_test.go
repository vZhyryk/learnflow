package adminhttp_test

import (
	"context"
	"net/http"

	admindomain "learnflow_backend/internal/admin/domain"
	adminhttp "learnflow_backend/internal/admin/transport/http"
	auditdomain "learnflow_backend/internal/audit/domain"
	"learnflow_backend/internal/shared/pagination"
	"learnflow_backend/internal/shared/testutil"

	"github.com/justinas/alice"
)

type errWriter = testutil.ErrWriter

var decodeBody = testutil.DecodeBody
var withUser = testutil.WithUser

func newAuthMux(svc *mockService) *http.ServeMux {
	h := adminhttp.NewHTTPHandler(svc, testutil.NewTestLogger())
	mux := http.NewServeMux()
	h.RegisterRoutes(mux, alice.Chain{}, alice.Chain{})
	return mux
}

// httpFixture wires a mockService-backed mux and a request builder for a single route.
type httpFixture struct {
	mux    *http.ServeMux
	newReq func(body string, urlParams map[string]string) *http.Request
}

func newHTTPFixture(svc *mockService, method, path string) *httpFixture {
	f := testutil.NewHTTPFixture(newAuthMux(svc), method, path)
	return &httpFixture{mux: f.Mux, newReq: f.NewReq}
}

type mockService struct {
	createAnnouncement         func(ctx context.Context, req admindomain.CreateAnnouncementRequest) (string, error)
	updateAnnouncement         func(ctx context.Context, req admindomain.UpdateAnnouncementRequest) error
	approveAnnouncement        func(ctx context.Context, announcementID, userID string) error
	setExpiredNowAnnouncement  func(ctx context.Context, announcementID, userID string) error
	getAnnouncements           func(ctx context.Context, params pagination.Params) ([]*admindomain.Announcement, error)
	getUnApprovedAnnouncements func(ctx context.Context, params pagination.Params) ([]*admindomain.Announcement, error)
	getApprovedAnnouncements   func(ctx context.Context, params pagination.Params) ([]*admindomain.Announcement, error)
	getPublicAnnouncements     func(ctx context.Context, params pagination.Params, userID string) ([]*admindomain.AnnouncementPublic, error)
	getExpiredAnnouncements    func(ctx context.Context, params pagination.Params) ([]*admindomain.Announcement, error)
	getUsersData               func(ctx context.Context, params pagination.Params) ([]*admindomain.UserData, int, error)
	getUserDataByID            func(ctx context.Context, userID string) (*admindomain.UserData, error)
	changeUserField            func(ctx context.Context, userID, adminID string, operationName admindomain.UserAdminOperation) error
	grantUserCourseAccess      func(ctx context.Context, userID, courseID, adminID string, action auditdomain.AdminActionType) error
	grantUserContentAccess     func(ctx context.Context, userID, contentItemID, adminID string, action auditdomain.AdminActionType) error
	getInstanceAdminActions    func(ctx context.Context, targetType auditdomain.AdminTargetType, targetID string, params pagination.Params) ([]*auditdomain.AdminAction, int, error)
	getFailedJobs              func(ctx context.Context, params pagination.Params) ([]*auditdomain.FailedJob, int, error)
}

func (m *mockService) GetFailedJobs(ctx context.Context, params pagination.Params) ([]*auditdomain.FailedJob, int, error) {
	if m.getFailedJobs == nil {
		panic("mockService.getFailedJobs not set")
	}
	return m.getFailedJobs(ctx, params)
}

func (m *mockService) GetInstanceAdminActions(ctx context.Context, targetType auditdomain.AdminTargetType, targetID string, params pagination.Params) ([]*auditdomain.AdminAction, int, error) {
	if m.getInstanceAdminActions == nil {
		panic("mockService.getInstanceAdminActions not set")
	}
	return m.getInstanceAdminActions(ctx, targetType, targetID, params)
}

func (m *mockService) GrantUserCourseAccess(ctx context.Context, userID, courseID, adminID string, action auditdomain.AdminActionType) error {
	if m.grantUserCourseAccess == nil {
		panic("mockService.grantUserCourseAccess not set")
	}
	return m.grantUserCourseAccess(ctx, userID, courseID, adminID, action)
}

func (m *mockService) GrantUserContentAccess(ctx context.Context, userID, contentItemID, adminID string, action auditdomain.AdminActionType) error {
	if m.grantUserContentAccess == nil {
		panic("mockService.grantUserContentAccess not set")
	}
	return m.grantUserContentAccess(ctx, userID, contentItemID, adminID, action)
}

func (m *mockService) CreateAnnouncement(ctx context.Context, req admindomain.CreateAnnouncementRequest) (string, error) {
	if m.createAnnouncement == nil {
		panic("mockService.createAnnouncement not set")
	}
	return m.createAnnouncement(ctx, req)
}

func (m *mockService) UpdateAnnouncement(ctx context.Context, req admindomain.UpdateAnnouncementRequest) error {
	if m.updateAnnouncement == nil {
		panic("mockService.updateAnnouncement not set")
	}
	return m.updateAnnouncement(ctx, req)
}

func (m *mockService) ApproveAnnouncement(ctx context.Context, announcementID, userID string) error {
	if m.approveAnnouncement == nil {
		panic("mockService.approveAnnouncement not set")
	}
	return m.approveAnnouncement(ctx, announcementID, userID)
}

func (m *mockService) SetExpiredNowAnnouncement(ctx context.Context, announcementID, userID string) error {
	if m.setExpiredNowAnnouncement == nil {
		panic("mockService.setExpiredNowAnnouncement not set")
	}
	return m.setExpiredNowAnnouncement(ctx, announcementID, userID)
}

func (m *mockService) GetAnnouncements(ctx context.Context, params pagination.Params) ([]*admindomain.Announcement, error) {
	if m.getAnnouncements == nil {
		panic("mockService.getAnnouncements not set")
	}
	return m.getAnnouncements(ctx, params)
}

func (m *mockService) GetUnApprovedAnnouncements(ctx context.Context, params pagination.Params) ([]*admindomain.Announcement, error) {
	if m.getUnApprovedAnnouncements == nil {
		panic("mockService.getUnApprovedAnnouncements not set")
	}
	return m.getUnApprovedAnnouncements(ctx, params)
}

func (m *mockService) GetPublicAnnouncements(ctx context.Context, params pagination.Params, userID string) ([]*admindomain.AnnouncementPublic, error) {
	if m.getPublicAnnouncements == nil {
		panic("mockService.getPublicAnnouncements not set")
	}
	return m.getPublicAnnouncements(ctx, params, userID)
}

func (m *mockService) GetApprovedAnnouncements(ctx context.Context, params pagination.Params) ([]*admindomain.Announcement, error) {
	if m.getApprovedAnnouncements == nil {
		panic("mockService.getApprovedAnnouncements not set")
	}
	return m.getApprovedAnnouncements(ctx, params)
}

func (m *mockService) GetExpiredAnnouncements(ctx context.Context, params pagination.Params) ([]*admindomain.Announcement, error) {
	if m.getExpiredAnnouncements == nil {
		panic("mockService.getExpiredAnnouncements not set")
	}
	return m.getExpiredAnnouncements(ctx, params)
}

func (m *mockService) GetUsersData(ctx context.Context, params pagination.Params) ([]*admindomain.UserData, int, error) {
	if m.getUsersData == nil {
		panic("mockService.getUsersData not set")
	}
	return m.getUsersData(ctx, params)
}

func (m *mockService) GetUserDataByID(ctx context.Context, userID string) (*admindomain.UserData, error) {
	if m.getUserDataByID == nil {
		panic("mockService.getUserDataByID not set")
	}
	return m.getUserDataByID(ctx, userID)
}

func (m *mockService) ChangeUserField(ctx context.Context, userID, adminID string, operationName admindomain.UserAdminOperation) error {
	if m.changeUserField == nil {
		panic("mockService.changeUserField not set")
	}
	return m.changeUserField(ctx, userID, adminID, operationName)
}
