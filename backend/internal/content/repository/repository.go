package contentrepository

import (
	contentdomain "learnflow_backend/internal/content/domain"
	"learnflow_backend/internal/shared/repository"

	"github.com/jackc/pgx/v5/pgxpool"
)

// Repository implements contentdomain.ContentRepository over PostgreSQL.
type Repository struct {
	repository.BaseRepository
}

// NewRepository returns a new Repository backed by the given connection pool.
func NewRepository(pool *pgxpool.Pool) *Repository {
	return &Repository{BaseRepository: *repository.NewBaseRepository(pool)}
}

var _ contentdomain.ContentRepository = (*Repository)(nil)
