package adminrepository

import (
	admindomain "learnflow_backend/internal/admin/domain"
	"learnflow_backend/internal/shared/repository"

	"github.com/jackc/pgx/v5/pgxpool"
)

// Repository implements admindomain.AnnouncementRepository over PostgreSQL.
type Repository struct {
	repository.BaseRepository
}

// NewRepository returns a new Repository backed by the given connection pool.
func NewRepository(pool *pgxpool.Pool) *Repository {
	return &Repository{BaseRepository: *repository.NewBaseRepository(pool)}
}

var _ admindomain.AnnouncementRepository = (*Repository)(nil)
