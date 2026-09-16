package adminhttp_test

import (
	"context"
	"net/http"

	admindomain "learnflow_backend/internal/admin/domain"
	adminhttp "learnflow_backend/internal/admin/transport/http"
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
	h.RegisterRoutes(mux, alice.Chain{})
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
	getAnnouncements           func(ctx context.Context, params pagination.Params) ([]*admindomain.Announcement, error)
	getUnApprovedAnnouncements func(ctx context.Context, params pagination.Params) ([]*admindomain.Announcement, error)
	getApprovedAnnouncements   func(ctx context.Context, params pagination.Params) ([]*admindomain.Announcement, error)
	getExpiredAnnouncements    func(ctx context.Context, params pagination.Params) ([]*admindomain.Announcement, error)
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
