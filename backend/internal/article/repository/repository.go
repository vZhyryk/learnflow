package articlerepository

import (
	articledomain "learnflow_backend/internal/article/domain"
	"learnflow_backend/internal/shared/repository"

	"github.com/jackc/pgx/v5/pgxpool"
)

// Repository implements articledomain.ArticleRepository over PostgreSQL.
type Repository struct {
	repository.BaseRepository
}

// NewRepository returns a new Repository backed by the given connection pool.
func NewRepository(pool *pgxpool.Pool) *Repository {
	return &Repository{BaseRepository: *repository.NewBaseRepository(pool)}
}

var _ articledomain.ArticleRepository = (*Repository)(nil)
