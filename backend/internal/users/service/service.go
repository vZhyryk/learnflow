package usersservice

import (
	usersdomain "learnflow_backend/internal/users/domain"
)

// Service implements usersdomain.Service.
type Service struct {
	usersRepo usersdomain.UserProfileRepository
}

// New returns a new Service wired to the given repository.
func New(usersRepo usersdomain.UserProfileRepository) *Service {
	return &Service{usersRepo: usersRepo}
}
