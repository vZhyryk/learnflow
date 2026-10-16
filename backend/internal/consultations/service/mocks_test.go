package consultationservice

import (
	"context"
	"time"

	auditdomain "learnflow_backend/internal/audit/domain"
	consultationsdomain "learnflow_backend/internal/consultations/domain"
	"learnflow_backend/internal/shared/testutil"
)

type mockSlotRepo struct {
	getAvailableSlots      func(ctx context.Context, from, to time.Time, consultantID *string) ([]*consultationsdomain.PublicSlot, int, error)
	getAllSlots            func(ctx context.Context, from, to time.Time, consultantID *string) ([]*consultationsdomain.Slot, int, error)
	updateSlotsStatus      func(ctx context.Context, ids []string, status consultationsdomain.SlotStatus, updatedBy string) error
	getSlotsByIDsForUpdate func(ctx context.Context, ids []string) ([]*consultationsdomain.Slot, error)
	createSlotsIfMissing   func(ctx context.Context, consultantID string, dates []time.Time, starts []string, durationMinutes int) error
	getConsultants         func(ctx context.Context) ([]*consultationsdomain.Consultant, error)
}

func (m *mockSlotRepo) GetAvailableSlots(ctx context.Context, from, to time.Time, consultantID *string) ([]*consultationsdomain.PublicSlot, int, error) {
	if m.getAvailableSlots == nil {
		panic("mockSlotRepo.GetAvailableSlots: not set")
	}

	return m.getAvailableSlots(ctx, from, to, consultantID)
}

func (m *mockSlotRepo) GetAllSlots(ctx context.Context, from, to time.Time, consultantID *string) ([]*consultationsdomain.Slot, int, error) {
	if m.getAllSlots == nil {
		panic("mockSlotRepo.GetAllSlots: not set")
	}

	return m.getAllSlots(ctx, from, to, consultantID)
}

func (m *mockSlotRepo) UpdateSlotsStatus(ctx context.Context, ids []string, status consultationsdomain.SlotStatus, updatedBy string) error {
	if m.updateSlotsStatus == nil {
		panic("mockSlotRepo.UpdateSlotsStatus: not set")
	}

	return m.updateSlotsStatus(ctx, ids, status, updatedBy)
}

func (m *mockSlotRepo) GetSlotsByIDsForUpdate(ctx context.Context, ids []string) ([]*consultationsdomain.Slot, error) {
	if m.getSlotsByIDsForUpdate == nil {
		panic("mockSlotRepo.GetSlotsByIDsForUpdate: not set")
	}

	return m.getSlotsByIDsForUpdate(ctx, ids)
}

func (m *mockSlotRepo) CreateSlotsIfMissing(ctx context.Context, consultantID string, dates []time.Time, starts []string, durationMinutes int) error {
	if m.createSlotsIfMissing == nil {
		panic("mockSlotRepo.CreateSlotsIfMissing: not set")
	}

	return m.createSlotsIfMissing(ctx, consultantID, dates, starts, durationMinutes)
}

func (m *mockSlotRepo) GetConsultants(ctx context.Context) ([]*consultationsdomain.Consultant, error) {
	if m.getConsultants == nil {
		panic("mockSlotRepo.GetConsultants: not set")
	}

	return m.getConsultants(ctx)
}

type mockActionRepo struct {
	createAdminAction func(ctx context.Context, action *auditdomain.AdminAction) error
}

func (m *mockActionRepo) CreateAdminAction(ctx context.Context, action *auditdomain.AdminAction) error {
	if m.createAdminAction == nil {
		panic("mockActionRepo.CreateAdminAction: not set")
	}

	return m.createAdminAction(ctx, action)
}

func newTestService(slotRepo *mockSlotRepo, actionRepo *mockActionRepo) *Service {
	loc, err := time.LoadLocation(consultationsdomain.SlotTimeZone)
	if err != nil {
		panic(err)
	}

	return New(slotRepo, actionRepo, testutil.NoopTransactor{}, consultationsdomain.DefaultSlotSchedule(loc))
}
