package courseservice

import (
	coursedomain "learnflow_backend/internal/courses/domain"
	"learnflow_backend/internal/events"
)

// Service implements coursedomain.Service.
type Service struct {
	courseRepo coursedomain.CourseRepository
	transactor coursedomain.Transactor
	outbox     *events.OutboxWriter
}

// New returns a new Service wired to the given repository.
func New(courseRepo coursedomain.CourseRepository, transactor coursedomain.Transactor, outbox *events.OutboxWriter) *Service {
	return &Service{courseRepo: courseRepo, transactor: transactor, outbox: outbox}
}
