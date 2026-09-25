package articleservice

import (
	articledomain "learnflow_backend/internal/article/domain"
)

// Service implements articledomain.Service.
type Service struct {
	articleRepo articledomain.ArticleRepository
	transactor  articledomain.Transactor
	actionRepo  articledomain.AdminActionRepository
}

var _ articledomain.Service = (*Service)(nil)

// New returns a new Service wired to the given repository.
func New(articleRepo articledomain.ArticleRepository, actionRepo articledomain.AdminActionRepository, transactor articledomain.Transactor) *Service {
	return &Service{articleRepo: articleRepo, transactor: transactor, actionRepo: actionRepo}
}
