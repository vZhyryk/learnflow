package authrepository

import (
	authdomain "learnflow_backend/internal/auth/domain"
	"learnflow_backend/internal/shared/repository"

	"github.com/jackc/pgx/v5/pgxpool"
)

// Repository implements UserRepository, SessionRepository, and TokenRepository using pgxpool.
type Repository struct {
	repository.BaseRepository
}

// NewRepository returns a new Repository backed by the given connection pool.
func NewRepository(pool *pgxpool.Pool) *Repository {
	return &Repository{BaseRepository: *repository.NewBaseRepository(pool)}
}

var (
	_ authdomain.SessionRepository = (*Repository)(nil)
	_ authdomain.TokenRepository   = (*Repository)(nil)
	_ authdomain.UserRepository    = (*Repository)(nil)
)
