package repository

import (
	"context"
	"learnflow_backend/internal/infrastructure/db"

	"github.com/jackc/pgx/v5/pgxpool"
)

// BaseRepository provides the QueryRunner wiring shared by every domain repository —
// embed it instead of duplicating the pool/transaction plumbing per module.
type BaseRepository struct {
	DB db.QueryRunner
}

// NewBaseRepository returns a new BaseRepository backed by the given connection pool.
func NewBaseRepository(pool *pgxpool.Pool) *BaseRepository {
	return &BaseRepository{DB: pool}
}

// QueryRunner returns the active transaction from ctx if present, otherwise the pool.
func (rep *BaseRepository) QueryRunner(ctx context.Context) db.QueryRunner {
	return db.FallbackQueryRunner(ctx, rep.DB)
}
