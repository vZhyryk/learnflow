package usersrepository

import (
	"context"
	"learnflow_backend/internal/infrastructure/db"
	usersdomain "learnflow_backend/internal/users/domain"

	"github.com/jackc/pgx/v5/pgxpool"
)

// Repository implements usersdomain.UserProfileRepository using pgxpool.
type Repository struct {
	db db.QueryRunner
}

// NewRepository returns a new Repository backed by the given connection pool.
func NewRepository(pool *pgxpool.Pool) *Repository {
	return &Repository{db: pool}
}

// queryRunner returns ctx's active transaction (see db.ExtractTx) or falls back to the
// pool — lets services wrap calls in a transaction without an explicit tx parameter.
func (rep *Repository) queryRunner(ctx context.Context) db.QueryRunner {
	return db.FallbackQueryRunner(ctx, rep.db)
}

var _ usersdomain.UserProfileRepository = (*Repository)(nil)
