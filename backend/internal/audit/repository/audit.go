package auditrepository

import (
	"learnflow_backend/internal/shared/repository"

	"github.com/jackc/pgx/v5/pgxpool"
)

// Audit writes the admin audit trail; it joins the transaction carried by ctx when present.
type Audit struct {
	repository.BaseRepository
}

// New returns an Audit backed by the given connection pool.
func New(pool *pgxpool.Pool) *Audit {
	return &Audit{BaseRepository: *repository.NewBaseRepository(pool)}
}
