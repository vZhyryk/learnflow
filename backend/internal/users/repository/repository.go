package usersrepository

import (
	"learnflow_backend/internal/shared/repository"
	usersdomain "learnflow_backend/internal/users/domain"

	"github.com/jackc/pgx/v5/pgxpool"
)

// Repository implements usersdomain.UserProfileRepository using pgxpool.
type Repository struct {
	repository.BaseRepository
}

// NewRepository returns a new Repository backed by the given connection pool.
func NewRepository(pool *pgxpool.Pool) *Repository {
	return &Repository{BaseRepository: *repository.NewBaseRepository(pool)}
}

var _ usersdomain.UserProfileRepository = (*Repository)(nil)
