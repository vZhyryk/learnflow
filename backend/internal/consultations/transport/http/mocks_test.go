package consultationshttp_test

import (
	"context"
	"net/http"

	authdomain "learnflow_backend/internal/auth/domain"
	consultationsdomain "learnflow_backend/internal/consultations/domain"
	consultationshttp "learnflow_backend/internal/consultations/transport/http"
	appcontext "learnflow_backend/internal/shared/context"
	"learnflow_backend/internal/shared/testutil"

	"github.com/justinas/alice"
)

const (
	adminID    = "aaaaaaaa-aaaa-4aaa-8aaa-aaaaaaaaaaaa"
	slotPath   = "/v1/consultations/slots"
	adminPath  = "/v1/admin/consultations/slots"
	consPath   = "/v1/consultations/consultants"
	validRange = "2026-11-02"
)

type mockService struct {
	getAvailableSlots func(ctx context.Context, slotRange consultationsdomain.SlotRange, consultantID *string) ([]*consultationsdomain.PublicSlot, int, error)
	getAllSlots       func(ctx context.Context, slotRange consultationsdomain.SlotRange, consultantID *string) ([]*consultationsdomain.Slot, int, error)
	updateSlots       func(ctx context.Context, req consultationsdomain.UpdateSlotsRequest) (*consultationsdomain.UpdateSlotsResult, error)
	getConsultants    func(ctx context.Context) ([]*consultationsdomain.Consultant, error)
}

func (m *mockService) GetAvailableSlots(ctx context.Context, slotRange consultationsdomain.SlotRange, consultantID *string) ([]*consultationsdomain.PublicSlot, int, error) {
	if m.getAvailableSlots == nil {
		panic("mockService.GetAvailableSlots not set")
	}

	return m.getAvailableSlots(ctx, slotRange, consultantID)
}

func (m *mockService) GetAllSlots(ctx context.Context, slotRange consultationsdomain.SlotRange, consultantID *string) ([]*consultationsdomain.Slot, int, error) {
	if m.getAllSlots == nil {
		panic("mockService.GetAllSlots not set")
	}

	return m.getAllSlots(ctx, slotRange, consultantID)
}

func (m *mockService) UpdateSlots(ctx context.Context, req consultationsdomain.UpdateSlotsRequest) (*consultationsdomain.UpdateSlotsResult, error) {
	if m.updateSlots == nil {
		panic("mockService.UpdateSlots not set")
	}

	return m.updateSlots(ctx, req)
}

func (m *mockService) GenerateSlots(context.Context) error {
	panic("mockService.GenerateSlots not set")
}

func (m *mockService) GetConsultants(ctx context.Context) ([]*consultationsdomain.Consultant, error) {
	if m.getConsultants == nil {
		panic("mockService.GetConsultants not set")
	}

	return m.getConsultants(ctx)
}

func withAdmin(r *http.Request) *http.Request {
	return r.WithContext(appcontext.WithUser(r.Context(), &authdomain.User{ID: adminID}))
}

func newHTTPFixture(svc *mockService, method, path string) *testutil.HTTPFixture {
	mux := http.NewServeMux()
	consultationshttp.NewHTTPHandler(svc, testutil.NewTestLogger()).RegisterRoutes(mux, alice.Chain{}, alice.Chain{})

	return testutil.NewHTTPFixture(mux, method, path)
}
