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

func (rep *Repository) queryRunner(ctx context.Context) db.QueryRunner {
	if tx, ok := db.ExtractTx(ctx); ok {
		return tx
	}
	return rep.db
}

var _ usersdomain.UserProfileRepository = (*Repository)(nil)
