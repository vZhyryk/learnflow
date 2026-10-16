package consultationservice

import (
	consultationsdomain "learnflow_backend/internal/consultations/domain"
)

// Service implements consultationsdomain.Service.
type Service struct {
	slotRepo   consultationsdomain.SlotRepository
	transactor consultationsdomain.Transactor
	actionRepo consultationsdomain.AdminActionRepository
	schedule   consultationsdomain.SlotSchedule
}

var _ consultationsdomain.Service = (*Service)(nil)

// New returns a new Service wired to the given repositories and slot schedule.
func New(slotRepo consultationsdomain.SlotRepository, actionRepo consultationsdomain.AdminActionRepository, transactor consultationsdomain.Transactor, schedule consultationsdomain.SlotSchedule) *Service {
	return &Service{slotRepo: slotRepo, transactor: transactor, actionRepo: actionRepo, schedule: schedule}
}
