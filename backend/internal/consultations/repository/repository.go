package consultationsrepository

import (
	consultationsdomain "learnflow_backend/internal/consultations/domain"
	"learnflow_backend/internal/shared/repository"

	"github.com/jackc/pgx/v5/pgxpool"
)

// Repository implements consultationsdomain.SlotRepository over PostgreSQL.
type Repository struct {
	repository.BaseRepository
}

// NewRepository returns a new Repository backed by the given connection pool.
func NewRepository(pool *pgxpool.Pool) *Repository {
	return &Repository{BaseRepository: *repository.NewBaseRepository(pool)}
}

var (
	_ consultationsdomain.SlotRepository = (*Repository)(nil)
)
