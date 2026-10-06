package auditrepository

import (
	"context"
	"errors"
	"fmt"

	auditdomain "learnflow_backend/internal/audit/domain"
	"learnflow_backend/internal/shared/repository"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

const (
	createAdminActionSQL = `
		INSERT INTO admin_actions (admin_user_id, action_type, target_type, target_id, details_json)
		VALUES ($1, $2, $3, $4, $5::jsonb)
	`

	wasDeletedByAdminSQL = `
		SELECT action_type = 'delete_user'
		FROM admin_actions
		WHERE target_type = 'user' AND target_id = $1 AND action_type IN ('delete_user', 'restore_user')
		ORDER BY created_at DESC, id DESC
		LIMIT 1
	`
)

// Audit writes the admin audit trail; it joins the transaction carried by ctx when present.
type Audit struct {
	repository.BaseRepository
}

// New returns an Audit backed by the given connection pool.
func New(pool *pgxpool.Pool) *Audit {
	return &Audit{BaseRepository: *repository.NewBaseRepository(pool)}
}

// CreateAdminAction appends an entry to the admin audit trail.
func (r *Audit) CreateAdminAction(ctx context.Context, action *auditdomain.AdminAction) error {
	_, err := r.QueryRunner(ctx).Exec(ctx, createAdminActionSQL, action.AdminUserID, action.ActionType, action.TargetType, action.TargetID, action.Details)
	if err != nil {
		return fmt.Errorf("audit.CreateAdminAction: %w", err)
	}

	return nil
}

// WasDeletedByAdmin reports whether the latest admin delete/restore action on targetID is a delete,
// so a user who deleted their own account after an admin restored it can still recover it.
func (r *Audit) WasDeletedByAdmin(ctx context.Context, targetID string) (bool, error) {
	var exists bool
	err := r.QueryRunner(ctx).QueryRow(ctx, wasDeletedByAdminSQL, targetID).Scan(&exists)
	if errors.Is(err, pgx.ErrNoRows) {
		return false, nil
	}

	if err != nil {
		return false, fmt.Errorf("audit.WasDeletedByAdmin: %w", err)
	}
	return exists, nil
}
