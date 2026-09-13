package reviewservice

import (
	reviewdomain "learnflow_backend/internal/review/domain"
)

// Service implements reviewdomain.Service.
type Service struct {
	courseRepo    reviewdomain.CourseReviewRepository
	contentRepo   reviewdomain.ContentReviewRepository
	transactor    reviewdomain.Transactor
	accessChecker reviewdomain.AccessChecker
}

var _ reviewdomain.Service = (*Service)(nil)

// New returns a new Service wired to the given repository.
func New(courseRepo reviewdomain.CourseReviewRepository, contentRepo reviewdomain.ContentReviewRepository, transactor reviewdomain.Transactor, accessChecker reviewdomain.AccessChecker) *Service {
	return &Service{courseRepo: courseRepo, contentRepo: contentRepo, transactor: transactor, accessChecker: accessChecker}
}
