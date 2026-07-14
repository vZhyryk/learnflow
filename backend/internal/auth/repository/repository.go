package authrepository

import (
	"context"

	authdomain "learnflow_backend/internal/auth/domain"
	"learnflow_backend/internal/infrastructure/db"

	"github.com/jackc/pgx/v5/pgxpool"
)

// Repository implements UserRepository, SessionRepository, and TokenRepository using pgxpool.
type Repository struct {
	db db.QueryRunner
}

// NewRepository returns a new auth Repository backed by the given connection pool.
func NewRepository(pool *pgxpool.Pool) *Repository {
	return &Repository{db: pool}
}

// queryRunner returns the active transaction from ctx if one was started by the
// caller (see db.ExtractTx), otherwise falls back to the connection pool. This lets
// service-layer code wrap multiple repository calls in a single transaction
// transparently, without every method taking an explicit tx parameter.
func (rep *Repository) queryRunner(ctx context.Context) db.QueryRunner {
	return db.FallbackQueryRunner(ctx, rep.db)
}

var (
	_ authdomain.SessionRepository = (*Repository)(nil)
	_ authdomain.TokenRepository   = (*Repository)(nil)
	_ authdomain.UserRepository    = (*Repository)(nil)
)
