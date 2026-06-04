package authrepository

import (
	authdomain "learnflow_backend/internal/auth/domain"

	"github.com/jackc/pgx/v5/pgxpool"
)

// Repository implements UserRepository, SessionRepository, and TokenRepository using pgxpool.
type Repository struct {
	db *pgxpool.Pool
}

// NewRepository returns a new auth Repository backed by the given connection pool.
func NewRepository(db *pgxpool.Pool) *Repository {
	return &Repository{
		db: db,
	}
}

var _ authdomain.SessionRepository = (*Repository)(nil)
var _ authdomain.TokenRepository = (*Repository)(nil)
var _ authdomain.UserRepository = (*Repository)(nil)
