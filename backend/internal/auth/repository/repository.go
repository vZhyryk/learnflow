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
func NewRepository(db *pgxpool.Pool) *Repository {
	return &Repository{db: db}
}

func (rep *Repository) queryRunner(ctx context.Context) db.QueryRunner {
	if tx, ok := db.ExtractTx(ctx); ok {
		return tx
	}
	return rep.db
}

var _ authdomain.SessionRepository = (*Repository)(nil)
var _ authdomain.TokenRepository = (*Repository)(nil)
var _ authdomain.UserRepository = (*Repository)(nil)
