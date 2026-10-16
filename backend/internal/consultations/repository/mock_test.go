package consultationsrepository

import (
	consultationsdomain "learnflow_backend/internal/consultations/domain"
	"learnflow_backend/internal/shared/repository"
	"learnflow_backend/internal/shared/testutil"
	"strconv"
	"time"
)

func newTestRepo(runner *testutil.MockQueryRunner) *Repository {
	return &Repository{repository.BaseRepository{DB: runner}}
}

func fakeSlot(n int) *consultationsdomain.Slot {
	updatedBy := "admin-" + strconv.Itoa(n)
	now := time.Date(2026, 10, 15, 12, 0, 0, 0, time.UTC)

	return &consultationsdomain.Slot{
		ID:              "slot-" + strconv.Itoa(n),
		ConsultantID:    "consultant-1",
		Date:            time.Date(2026, 11, n, 0, 0, 0, 0, time.UTC),
		StartTime:       "09:00",
		DurationMinutes: 120,
		Status:          consultationsdomain.SlotStatusAvailable,
		UpdatedBy:       &updatedBy,
		CreatedAt:       now,
		UpdatedAt:       now,
	}
}

func fakePublicSlot(n int) *consultationsdomain.PublicSlot {
	return &consultationsdomain.PublicSlot{
		ID:              "slot-" + strconv.Itoa(n),
		ConsultantID:    "consultant-1",
		Date:            time.Date(2026, 11, n, 0, 0, 0, 0, time.UTC),
		StartTime:       "11:00",
		DurationMinutes: 120,
		Status:          consultationsdomain.SlotStatusAvailable,
	}
}

// fakeSlotScan simulates rows.Scan in the column order of slotsColumns, matching scanSlot.
func fakeSlotScan(slot *consultationsdomain.Slot) func(dest ...any) error {
	return func(dest ...any) error {
		*testutil.CastStr(dest[0], 0) = slot.ID
		*testutil.CastStr(dest[1], 1) = slot.ConsultantID
		*testutil.CastTime(dest[2], 2) = slot.Date
		*testutil.CastStr(dest[3], 3) = slot.StartTime
		*testutil.CastInt(dest[4], 4) = slot.DurationMinutes
		*testutil.CastEnum[consultationsdomain.SlotStatus](dest[5], 5) = slot.Status
		*testutil.CastPtrStr(dest[6], 6) = slot.UpdatedBy
		*testutil.CastTime(dest[7], 7) = slot.CreatedAt
		*testutil.CastTime(dest[8], 8) = slot.UpdatedAt

		return nil
	}
}

// fakeAdminSlotScan is fakeSlotScan plus the brief id and user id of getAllSlotsSQL, matching scanAdminSlot.
func fakeAdminSlotScan(slot *consultationsdomain.Slot, briefID, userID *string) func(dest ...any) error {
	scanSlot := fakeSlotScan(slot)

	return func(dest ...any) error {
		if err := scanSlot(dest...); err != nil {
			return err
		}
		*testutil.CastPtrStr(dest[9], 9) = briefID
		*testutil.CastPtrStr(dest[10], 10) = userID

		return nil
	}
}

// fakePublicSlotScan simulates rows.Scan in the column order of publicSlotsColumns, matching scanPublicSlot.
func fakePublicSlotScan(slot *consultationsdomain.PublicSlot) func(dest ...any) error {
	return func(dest ...any) error {
		*testutil.CastStr(dest[0], 0) = slot.ID
		*testutil.CastStr(dest[1], 1) = slot.ConsultantID
		*testutil.CastTime(dest[2], 2) = slot.Date
		*testutil.CastStr(dest[3], 3) = slot.StartTime
		*testutil.CastInt(dest[4], 4) = slot.DurationMinutes
		*testutil.CastEnum[consultationsdomain.SlotStatus](dest[5], 5) = slot.Status

		return nil
	}
}
