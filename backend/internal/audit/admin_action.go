package audit

import (
	"context"
	"fmt"

	auditdomain "learnflow_backend/internal/audit/domain"
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

const createAdminActionSQL = `
	INSERT INTO admin_actions (admin_user_id, action_type, target_type, target_id)
	VALUES ($1, $2, $3, $4)
`

// CreateAdminAction appends an entry to the admin audit trail.
func (r *Audit) CreateAdminAction(ctx context.Context, action *auditdomain.AdminAction) error {
	_, err := r.QueryRunner(ctx).Exec(ctx, createAdminActionSQL, action.AdminUserID, action.ActionType, action.TargetType, action.TargetID)
	if err != nil {
		return fmt.Errorf("audit.CreateAdminAction: %w", err)
	}

	return nil
}
