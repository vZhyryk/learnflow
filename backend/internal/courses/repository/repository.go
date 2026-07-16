package courserepository

import (
	"context"
	coursedomain "learnflow_backend/internal/courses/domain"
	"learnflow_backend/internal/infrastructure/db"

	"github.com/jackc/pgx/v5/pgxpool"
)

// Repository implements coursedomain.CourseRepository over PostgreSQL.
type Repository struct {
	db db.QueryRunner
}

// NewRepository returns a new Repository backed by the given connection pool.
func NewRepository(pool *pgxpool.Pool) *Repository {
	return &Repository{db: pool}
}

func (rep *Repository) queryRunner(ctx context.Context) db.QueryRunner {
	return db.FallbackQueryRunner(ctx, rep.db)
}

var _ coursedomain.CourseRepository = (*Repository)(nil)
