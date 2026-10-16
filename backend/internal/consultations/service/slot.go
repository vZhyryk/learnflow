package consultationservice

import (
	"context"
	"fmt"
	"strings"
	"time"

	auditdomain "learnflow_backend/internal/audit/domain"
	consultationsdomain "learnflow_backend/internal/consultations/domain"
)

// GetAvailableSlots returns the free slots of the range, optionally of one consultant.
func (srv *Service) GetAvailableSlots(ctx context.Context, slotRange consultationsdomain.SlotRange, consultantID *string) ([]*consultationsdomain.PublicSlot, int, error) {
	slots, count, err := srv.slotRepo.GetAvailableSlots(ctx, slotRange.From, slotRange.To, consultantID)
	if err != nil {
		return nil, 0, fmt.Errorf("service.GetAvailableSlots: %w", err)
	}

	return slots, count, nil
}

// GetAllSlots returns the slots of the range in every status for the admin calendar.
func (srv *Service) GetAllSlots(ctx context.Context, slotRange consultationsdomain.SlotRange, consultantID *string) ([]*consultationsdomain.Slot, int, error) {
	slots, count, err := srv.slotRepo.GetAllSlots(ctx, slotRange.From, slotRange.To, consultantID)
	if err != nil {
		return nil, 0, fmt.Errorf("service.GetAllSlots: %w", err)
	}

	return slots, count, nil
}

// UpdateSlots blocks or frees slots in one transaction; booked and unknown slots are reported as skipped.
func (srv *Service) UpdateSlots(ctx context.Context, req consultationsdomain.UpdateSlotsRequest) (*consultationsdomain.UpdateSlotsResult, error) {
	var res *consultationsdomain.UpdateSlotsResult

	err := srv.transactor.InTransaction(ctx, func(ctx context.Context) error {
		slots, err := srv.slotRepo.GetSlotsByIDsForUpdate(ctx, req.GetIDList())
		if err != nil {
			return fmt.Errorf("service.UpdateSlots: lock slots: %w", err)
		}

		var toUpdate map[consultationsdomain.SlotStatus][]string
		res, toUpdate = planSlotUpdates(req.Slots, slots)

		return srv.applySlotUpdates(ctx, req.AdminUserID, toUpdate)
	})
	if err != nil {
		return nil, err
	}

	return res, nil
}

func (srv *Service) applySlotUpdates(ctx context.Context, adminID string, toUpdate map[consultationsdomain.SlotStatus][]string) error {
	for _, status := range []consultationsdomain.SlotStatus{consultationsdomain.SlotStatusBlocked, consultationsdomain.SlotStatusAvailable} {
		ids := toUpdate[status]
		if len(ids) == 0 {
			continue
		}

		if err := srv.slotRepo.UpdateSlotsStatus(ctx, ids, status, adminID); err != nil {
			return fmt.Errorf("service.UpdateSlots: set status %s: %w", status, err)
		}

		for _, id := range ids {
			action := &auditdomain.AdminAction{
				AdminUserID: adminID,
				ActionType:  auditdomain.ActionUpdateItem,
				TargetType:  auditdomain.TargetConsultationSlot,
				TargetID:    id,
			}
			if err := srv.actionRepo.CreateAdminAction(ctx, action); err != nil {
				return fmt.Errorf("service.UpdateSlots: audit slot %s: %w", id, err)
			}
		}
	}

	return nil
}

// planSlotUpdates sorts the requested changes into applied, skipped and the IDs that need an UPDATE per status.
// A slot already in the requested status counts as applied without an UPDATE.
func planSlotUpdates(changes []consultationsdomain.SlotStatusChange, slots []*consultationsdomain.Slot) (res *consultationsdomain.UpdateSlotsResult, toUpdate map[consultationsdomain.SlotStatus][]string) {
	byID := make(map[string]*consultationsdomain.Slot, len(slots))
	for _, slot := range slots {
		byID[strings.ToLower(slot.ID)] = slot
	}

	res = &consultationsdomain.UpdateSlotsResult{
		Applied: []string{},
		Skipped: []consultationsdomain.SkippedSlot{},
	}
	toUpdate = make(map[consultationsdomain.SlotStatus][]string)

	for _, change := range changes {
		slot, found := byID[strings.ToLower(change.SlotID)]

		switch {
		case !found:
			res.Skipped = append(res.Skipped, consultationsdomain.SkippedSlot{SlotID: change.SlotID, Reason: consultationsdomain.SkipReasonNotFound})
		case slot.Status == consultationsdomain.SlotStatusBooked:
			res.Skipped = append(res.Skipped, consultationsdomain.SkippedSlot{SlotID: change.SlotID, Reason: consultationsdomain.SkipReasonBooked})
		case slot.Status == change.Status:
			res.Applied = append(res.Applied, change.SlotID)
		default:
			toUpdate[change.Status] = append(toUpdate[change.Status], change.SlotID)
			res.Applied = append(res.Applied, change.SlotID)
		}
	}

	return res, toUpdate
}

// GenerateSlots tops up the scheduled slots for every consultant over the schedule horizon.
func (srv *Service) GenerateSlots(ctx context.Context) error {
	consultants, err := srv.slotRepo.GetConsultants(ctx)
	if err != nil {
		return fmt.Errorf("service.GenerateSlots: get consultants: %w", err)
	}

	from := time.Now().In(srv.schedule.Location)
	dates, starts := srv.schedule.SlotsBetween(from, from.AddDate(0, 0, srv.schedule.HorizonDays))

	for _, c := range consultants {
		if err = srv.slotRepo.CreateSlotsIfMissing(ctx, c.ID, dates, starts, srv.schedule.DurationMinutes); err != nil {
			return fmt.Errorf("service.GenerateSlots: create slots for consultant %s: %w", c.ID, err)
		}
	}

	return nil
}

// GetConsultants returns the consultants a client can choose from.
func (srv *Service) GetConsultants(ctx context.Context) ([]*consultationsdomain.Consultant, error) {
	consultants, err := srv.slotRepo.GetConsultants(ctx)
	if err != nil {
		return nil, fmt.Errorf("service.GetConsultants: %w", err)
	}

	return consultants, nil
}
