package reviewrepository

import (
	reviewdomain "learnflow_backend/internal/review/domain"
	"learnflow_backend/internal/shared/repository"

	"github.com/jackc/pgx/v5/pgxpool"
)

// Repository implements persistence for course and content reviews.
type Repository struct {
	repository.BaseRepository
}

// NewRepository returns a new Repository backed by the given connection pool.
func NewRepository(pool *pgxpool.Pool) *Repository {
	return &Repository{BaseRepository: *repository.NewBaseRepository(pool)}
}

var (
	_ reviewdomain.CourseReviewRepository  = (*Repository)(nil)
	_ reviewdomain.ContentReviewRepository = (*Repository)(nil)
)
