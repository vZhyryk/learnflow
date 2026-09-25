package adminservice

import (
	admindomain "learnflow_backend/internal/admin/domain"
	"learnflow_backend/internal/events"
)

// Service implements admindomain.Service.
type Service struct {
	announRepo admindomain.AnnouncementRepository
	userRepo   admindomain.UserRepository
	actionRepo admindomain.AdminActionRepository
	transactor admindomain.Transactor
	outbox     *events.OutboxWriter
}

var _ admindomain.Service = (*Service)(nil)

// New returns a new Service wired to the given repository.
func New(
	announRepo admindomain.AnnouncementRepository,
	userRepo admindomain.UserRepository,
	actionRepo admindomain.AdminActionRepository,
	transactor admindomain.Transactor,
	outbox *events.OutboxWriter,
) *Service {
	return &Service{announRepo: announRepo, userRepo: userRepo, actionRepo: actionRepo, transactor: transactor, outbox: outbox}
}
