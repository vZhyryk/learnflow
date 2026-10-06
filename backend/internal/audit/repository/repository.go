package auditrepository

import (
	"context"
	"errors"
	"fmt"

	auditdomain "learnflow_backend/internal/audit/domain"
	"learnflow_backend/internal/shared/pagination"

	"github.com/jackc/pgx/v5"
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

	getInstanceAdminActionsSQL = `
		SELECT id, admin_user_id, action_type, target_type, target_id, details_json, created_at
		FROM admin_actions
		WHERE target_type = $1 AND target_id = $2
		ORDER BY created_at DESC, id DESC
		LIMIT $3 OFFSET $4
	`

	getInstanceAdminActionsCountSQL = `
		SELECT COUNT(*) FROM admin_actions
		WHERE target_type = $1 AND target_id = $2
	`
	getFailedJobsSQL = `
		SELECT
			id,
			event_type,
			queue_name,
			attempt_count,
			error_message,
			failed_at,
			resolved_at,
			resolution_note,
			created_at,
			updated_at
		FROM failed_jobs
		ORDER BY created_at DESC, id DESC
		LIMIT $1 OFFSET $2
	`

	getFailedJobsCountSQL = `
		SELECT COUNT(*) FROM failed_jobs
	`
)

// CreateAdminAction appends an entry to the admin audit trail.
func (r *Audit) CreateAdminAction(ctx context.Context, action *auditdomain.AdminAction) error {
	_, err := r.QueryRunner(ctx).Exec(ctx, createAdminActionSQL, action.AdminUserID, action.ActionType, action.TargetType, action.TargetID, detailsArg(action.Details))
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

// detailsArg passes an untyped nil for absent details so the column is always SQL NULL, never JSON null.
func detailsArg(details map[string]any) any {
	if details == nil {
		return nil
	}

	return details
}

// GetInstanceAdminActions returns one page of the audit trail for a target, newest first, and the total count.
func (r *Audit) GetInstanceAdminActions(ctx context.Context, targetType auditdomain.AdminTargetType, targetID string, params pagination.Params) ([]*auditdomain.AdminAction, int, error) {
	var count int
	err := r.QueryRunner(ctx).QueryRow(ctx, getInstanceAdminActionsCountSQL, targetType, targetID).Scan(&count)
	if err != nil {
		return nil, 0, fmt.Errorf("audit.GetInstanceAdminActionsCount: %w", err)
	}

	rows, err := r.QueryRunner(ctx).Query(ctx, getInstanceAdminActionsSQL, targetType, targetID, params.Limit(), params.Offset())
	if err != nil {
		return nil, 0, fmt.Errorf("audit.GetInstanceAdminActions: %w", err)
	}
	defer rows.Close()

	actions := make([]*auditdomain.AdminAction, 0)
	for rows.Next() {
		action, err := scanAdminAction(rows)
		if err != nil {
			return nil, 0, fmt.Errorf("audit.GetInstanceAdminActions: %w", err)
		}
		actions = append(actions, action)
	}

	if err := rows.Err(); err != nil {
		return nil, 0, fmt.Errorf("audit.GetInstanceAdminActions: %w", err)
	}

	return actions, count, nil
}

func (r *Audit) GetFailedJobs(ctx context.Context, params pagination.Params) ([]*auditdomain.FailedJob, int, error) {
	var count int
	err := r.QueryRunner(ctx).QueryRow(ctx, getFailedJobsCountSQL).Scan(&count)
	if err != nil {
		return nil, 0, fmt.Errorf("audit.GetFailedJobsCount: %w", err)
	}

	rows, err := r.QueryRunner(ctx).Query(ctx, getFailedJobsSQL, params.Limit(), params.Offset())
	if err != nil {
		return nil, 0, fmt.Errorf("audit.GetFailedJobs: %w", err)
	}
	defer rows.Close()

	jobs := make([]*auditdomain.FailedJob, 0)
	for rows.Next() {
		job, err := scanFailedJob(rows)
		if err != nil {
			return nil, 0, fmt.Errorf("audit.GetFailedJobs: %w", err)
		}
		jobs = append(jobs, job)
	}

	if err := rows.Err(); err != nil {
		return nil, 0, fmt.Errorf("audit.GetFailedJobs: %w", err)
	}

	return jobs, count, nil
}
