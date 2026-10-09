package auditrepository

import (
	"context"
	"errors"
	"fmt"

	auditdomain "learnflow_backend/internal/audit/domain"
	"learnflow_backend/internal/shared/pagination"
	"learnflow_backend/internal/shared/repository"

	"github.com/jackc/pgx/v5"
)

const (
	// adminActionSelectSQL is the shared SELECT ... JOIN part of every admin_actions list query; the admin's display
	// name falls back to the email when the profile has no first name.
	adminActionSelectSQL = `
		SELECT
			a.id,
			a.admin_user_id,
			COALESCE(NULLIF(p.first_name, ''), u.email) AS admin_name,
			a.action_type,
			a.target_type,
			a.target_id,
			a.details_json,
			a.created_at
		FROM admin_actions a
		JOIN users u ON u.id = a.admin_user_id
		LEFT JOIN user_profiles p ON p.user_id = u.id`

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

	getInstanceAdminActionsSQL = adminActionSelectSQL + `
		WHERE a.target_type = $1 AND a.target_id = $2
		ORDER BY a.created_at DESC, a.id DESC
		LIMIT $3 OFFSET $4
	`

	getInstanceAdminActionsCountSQL = `
		SELECT COUNT(*) FROM admin_actions
		WHERE target_type = $1 AND target_id = $2
	`
	adminActionFilterSQL = `
		WHERE ($1::uuid IS NULL OR a.admin_user_id = $1)
			AND ($2::text IS NULL OR a.action_type = $2)
			AND ($3::timestamptz IS NULL OR a.created_at >= $3)
			AND ($4::timestamptz IS NULL OR a.created_at < $4)
	`

	getAdminActionsSQL = adminActionSelectSQL + adminActionFilterSQL + `
		ORDER BY a.created_at DESC, a.id DESC
		LIMIT $5 OFFSET $6
	`

	getAdminActionsCountSQL = `
		SELECT COUNT(*) FROM admin_actions a` + adminActionFilterSQL + `
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
		return fmt.Errorf("repository.CreateAdminAction: %w", err)
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
		return false, fmt.Errorf("repository.WasDeletedByAdmin: %w", err)
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
		return nil, 0, fmt.Errorf("repository.GetInstanceAdminActionsCount: %w", err)
	}

	actions, err := repository.GetAndParseListWithArgs(ctx, &r.BaseRepository, getInstanceAdminActionsSQL, "GetInstanceAdminActions", &params, scanAdminAction, []any{targetType, targetID})
	if err != nil {
		return nil, 0, err
	}

	return actions, count, nil
}

// GetAdminActions returns one page of the general audit journal, newest first, and the total count of the filter.
func (r *Audit) GetAdminActions(ctx context.Context, filter auditdomain.AdminActionFilter, params pagination.Params) ([]*auditdomain.AdminAction, int, error) {
	filterArgs := []any{nilIfEmpty(filter.AdminUserID), nilIfEmpty(string(filter.ActionType)), filter.From, filter.To}

	var count int
	if err := r.QueryRunner(ctx).QueryRow(ctx, getAdminActionsCountSQL, filterArgs...).Scan(&count); err != nil {
		return nil, 0, fmt.Errorf("repository.GetAdminActionsCount: %w", err)
	}

	actions, err := repository.GetAndParseListWithArgs(ctx, &r.BaseRepository, getAdminActionsSQL, "GetAdminActions", &params, scanAdminAction, filterArgs)
	if err != nil {
		return nil, 0, err
	}

	return actions, count, nil
}

// nilIfEmpty passes an unset filter value as SQL NULL so its condition is skipped.
func nilIfEmpty(value string) any {
	if value == "" {
		return nil
	}

	return value
}

// GetFailedJobs returns one page of dead-lettered jobs, newest first, and the total count.
func (r *Audit) GetFailedJobs(ctx context.Context, params pagination.Params) ([]*auditdomain.FailedJob, int, error) {
	var count int
	err := r.QueryRunner(ctx).QueryRow(ctx, getFailedJobsCountSQL).Scan(&count)
	if err != nil {
		return nil, 0, fmt.Errorf("repository.GetFailedJobsCount: %w", err)
	}

	jobs, err := repository.GetAndParseList(ctx, &r.BaseRepository, getFailedJobsSQL, "GetFailedJobs", &params, scanFailedJob)
	if err != nil {
		return nil, 0, err
	}

	return jobs, count, nil
}
