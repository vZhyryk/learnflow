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

// queryRunner returns the active transaction from ctx if one was started by the
// caller (see db.ExtractTx), otherwise falls back to the connection pool. This lets
// service-layer code wrap multiple repository calls in a single transaction
// transparently, without every method taking an explicit tx parameter.
func (rep *Repository) queryRunner(ctx context.Context) db.QueryRunner {
	if tx, ok := db.ExtractTx(ctx); ok {
		return tx
	}
	return rep.db
}

var _ usersdomain.UserProfileRepository = (*Repository)(nil)
