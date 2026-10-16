package consultationsrepository

import (
	"context"
	"fmt"
	consultationsdomain "learnflow_backend/internal/consultations/domain"
	"learnflow_backend/internal/shared/repository"
	"time"
)

// GetAvailableSlots returns the free slots with from <= date <= to, ordered by date and start time, with their count.
// A non-nil consultantID narrows the list to that consultant.
func (rep *Repository) GetAvailableSlots(ctx context.Context, from, to time.Time, consultantID *string) ([]*consultationsdomain.PublicSlot, int, error) {
	return repository.GetCountAndParseListWithArgs(ctx, &rep.BaseRepository,
		getAvailableSlotsSQL, countAvailableSlotsSQL, "GetAvailableSlots", nil, scanPublicSlot, []any{from, to, consultantID})
}

// GetAllSlots returns every slot with from <= date <= to for the admin calendar, with the holder of booked slots.
// A non-nil consultantID narrows the list to that consultant.
func (rep *Repository) GetAllSlots(ctx context.Context, from, to time.Time, consultantID *string) ([]*consultationsdomain.Slot, int, error) {
	return repository.GetCountAndParseListWithArgs(ctx, &rep.BaseRepository,
		getAllSlotsSQL, countAllSlotsSQL, "GetAllSlots", nil, scanAdminSlot, []any{from, to, consultantID})
}

// GetSlotsByIDsForUpdate locks and returns the existing slots among ids; call it inside a transaction.
func (rep *Repository) GetSlotsByIDsForUpdate(ctx context.Context, ids []string) ([]*consultationsdomain.Slot, error) {
	return repository.GetAndParseListWithArgs(ctx, &rep.BaseRepository,
		getSlotsByIDsForUpdateSQL, "GetSlotsByIDsForUpdate", nil, scanSlot, []any{ids})
}

// UpdateSlotsStatus sets status on every given slot that is not booked and records updatedBy.
// It fails when fewer rows changed than ids given, i.e. a slot was booked or removed after it was locked.
func (rep *Repository) UpdateSlotsStatus(ctx context.Context, ids []string, status consultationsdomain.SlotStatus, updatedBy string) error {
	if len(ids) == 0 {
		return nil
	}

	tag, err := rep.QueryRunner(ctx).Exec(ctx, updateSlotsStatusSQL, status, updatedBy, ids)
	if err != nil {
		return fmt.Errorf("repository.UpdateSlotsStatus: %w", err)
	}

	if int(tag.RowsAffected()) != len(ids) {
		return fmt.Errorf("repository.UpdateSlotsStatus: updated %d of %d slots", tag.RowsAffected(), len(ids))
	}

	return nil
}

// CreateSlotsIfMissing inserts a slot of durationMinutes for every (dates[i], starts[i]) pair of the consultant,
// skipping pairs that already exist. dates and starts must have equal length.
func (rep *Repository) CreateSlotsIfMissing(ctx context.Context, consultantID string, dates []time.Time, starts []string, durationMinutes int) error {
	if len(dates) != len(starts) {
		return fmt.Errorf("repository.CreateSlotsIfMissing: %d dates and %d start times", len(dates), len(starts))
	}

	if len(dates) == 0 {
		return nil
	}

	_, err := rep.QueryRunner(ctx).Exec(ctx, createSlotsIfMissingSQL, consultantID, dates, starts, durationMinutes)
	if err != nil {
		return fmt.Errorf("repository.CreateSlotsIfMissing: %w", err)
	}

	return nil
}

// GetConsultants returns the active admin accounts, who conduct consultations, with their optional profile data.
func (rep *Repository) GetConsultants(ctx context.Context) ([]*consultationsdomain.Consultant, error) {
	return repository.GetAndParseListWithArgs(ctx, &rep.BaseRepository,
		getConsultantsSQL, "GetConsultants", nil, scanConsultant, nil)
}
