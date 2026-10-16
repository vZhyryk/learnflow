package consultationsrepository

import (
	consultationsdomain "learnflow_backend/internal/consultations/domain"
	"learnflow_backend/internal/shared/repository"
)

func scanSlot(row repository.RowScanner) (*consultationsdomain.Slot, error) {
	slot := &consultationsdomain.Slot{}
	err := row.Scan(
		&slot.ID,
		&slot.ConsultantID,
		&slot.Date,
		&slot.StartTime,
		&slot.DurationMinutes,
		&slot.Status,
		&slot.UpdatedBy,
		&slot.CreatedAt,
		&slot.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}

	return slot, nil
}

func scanAdminSlot(row repository.RowScanner) (*consultationsdomain.Slot, error) {
	slot := &consultationsdomain.Slot{}
	var briefID, userID *string
	err := row.Scan(
		&slot.ID,
		&slot.ConsultantID,
		&slot.Date,
		&slot.StartTime,
		&slot.DurationMinutes,
		&slot.Status,
		&slot.UpdatedBy,
		&slot.CreatedAt,
		&slot.UpdatedAt,
		&briefID,
		&userID,
	)
	if err != nil {
		return nil, err
	}

	if briefID != nil && userID != nil {
		slot.Booking = &consultationsdomain.SlotBooking{BriefID: *briefID, UserID: *userID}
	}

	return slot, nil
}

func scanPublicSlot(row repository.RowScanner) (*consultationsdomain.PublicSlot, error) {
	slot := &consultationsdomain.PublicSlot{}
	err := row.Scan(
		&slot.ID,
		&slot.ConsultantID,
		&slot.Date,
		&slot.StartTime,
		&slot.DurationMinutes,
		&slot.Status,
	)
	if err != nil {
		return nil, err
	}

	return slot, nil
}

func scanConsultant(row repository.RowScanner) (*consultationsdomain.Consultant, error) {
	consultant := &consultationsdomain.Consultant{}
	err := row.Scan(
		&consultant.ID,
		&consultant.FirstName,
		&consultant.LastName,
		&consultant.AvatarURL,
		&consultant.Bio,
	)
	if err != nil {
		return nil, err
	}

	return consultant, nil
}
