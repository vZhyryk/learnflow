package contentservice

import (
	contentdomain "learnflow_backend/internal/content/domain"
)

// Service implements contentdomain.Service.
type Service struct {
	contentRepo contentdomain.ContentRepository
	transactor  contentdomain.Transactor
	actionRepo  contentdomain.AdminActionRepository
}

var _ contentdomain.Service = (*Service)(nil)

// New returns a new Service wired to the given repository.
func New(contentRepo contentdomain.ContentRepository, actionRepo contentdomain.AdminActionRepository, transactor contentdomain.Transactor) *Service {
	return &Service{contentRepo: contentRepo, transactor: transactor, actionRepo: actionRepo}
}
