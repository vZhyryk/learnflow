package contentservice

import (
	contentdomain "learnflow_backend/internal/content/domain"
	"learnflow_backend/internal/events"
)

// Service implements contentdomain.Service.
type Service struct {
	contentRepo contentdomain.ContentRepository
	transactor  contentdomain.Transactor
	outbox      *events.OutboxWriter
}

var _ contentdomain.Service = (*Service)(nil)

// New returns a new Service wired to the given repository.
func New(contentRepo contentdomain.ContentRepository, transactor contentdomain.Transactor, outbox *events.OutboxWriter) *Service {
	return &Service{contentRepo: contentRepo, transactor: transactor, outbox: outbox}
}
