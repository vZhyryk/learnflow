package adminservice

import (
	admindomain "learnflow_backend/internal/admin/domain"
	"learnflow_backend/internal/events"
)

// Service implements admindomain.Service.
type Service struct {
	announRepo admindomain.AnnouncementRepository
	transactor admindomain.Transactor
	outbox     *events.OutboxWriter
}

var _ admindomain.Service = (*Service)(nil)

// New returns a new Service wired to the given repository.
func New(announRepo admindomain.AnnouncementRepository, transactor admindomain.Transactor, outbox *events.OutboxWriter) *Service {
	return &Service{announRepo: announRepo, transactor: transactor, outbox: outbox}
}
