package articleservice

import (
	articledomain "learnflow_backend/internal/article/domain"
	"learnflow_backend/internal/events"
)

// Service implements articledomain.Service.
type Service struct {
	articleRepo articledomain.ArticleRepository
	transactor  articledomain.Transactor
	outbox      *events.OutboxWriter
}

var _ articledomain.Service = (*Service)(nil)

// New returns a new Service wired to the given repository.
func New(articleRepo articledomain.ArticleRepository, transactor articledomain.Transactor, outbox *events.OutboxWriter) *Service {
	return &Service{articleRepo: articleRepo, transactor: transactor, outbox: outbox}
}
