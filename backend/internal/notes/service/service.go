package notesservice

import (
	notesdomain "learnflow_backend/internal/notes/domain"
)

// Service implements notesdomain.Service.
type Service struct {
	notesRepo   notesdomain.NotesRepository
	courseRepo  notesdomain.CourseRepository
	contentRepo notesdomain.ContentRepository

	transactor notesdomain.Transactor
}

var _ notesdomain.Service = (*Service)(nil)

// New returns a new Service wired to the given repository.
func New(notesRepo notesdomain.NotesRepository, courseRepo notesdomain.CourseRepository, contentRepo notesdomain.ContentRepository, transactor notesdomain.Transactor) *Service {
	return &Service{notesRepo: notesRepo, courseRepo: courseRepo, contentRepo: contentRepo, transactor: transactor}
}
