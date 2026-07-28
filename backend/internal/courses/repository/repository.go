package courserepository

import (
	coursedomain "learnflow_backend/internal/courses/domain"
	"learnflow_backend/internal/shared/repository"

	"github.com/jackc/pgx/v5/pgxpool"
)

// Repository implements coursedomain.CourseRepository over PostgreSQL.
type Repository struct {
	repository.BaseRepository
}

// NewRepository returns a new Repository backed by the given connection pool.
func NewRepository(pool *pgxpool.Pool) *Repository {
	return &Repository{BaseRepository: *repository.NewBaseRepository(pool)}
}

var _ coursedomain.CourseRepository = (*Repository)(nil)
