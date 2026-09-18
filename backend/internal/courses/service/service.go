package courseservice

import (
	coursedomain "learnflow_backend/internal/courses/domain"
)

// Service implements coursedomain.Service.
type Service struct {
	courseRepo coursedomain.CourseRepository
	transactor coursedomain.Transactor
}

// New returns a new Service wired to the given repository.
func New(courseRepo coursedomain.CourseRepository, transactor coursedomain.Transactor) *Service {
	return &Service{courseRepo: courseRepo, transactor: transactor}
}
