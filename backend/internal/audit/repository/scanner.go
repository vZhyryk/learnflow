package auditrepository

import (
	auditdomain "learnflow_backend/internal/audit/domain"
	"learnflow_backend/internal/shared/repository"
)

func scanAdminAction(row repository.RowScanner) (*auditdomain.AdminAction, error) {
	action := &auditdomain.AdminAction{}
	err := row.Scan(
		&action.ID,
		&action.AdminUserID,
		&action.ActionType,
		&action.TargetType,
		&action.TargetID,
		&action.Details,
		&action.CreatedAt,
	)
	if err != nil {
		return nil, err
	}
	return action, nil
}

func scanFailedJob(row repository.RowScanner) (*auditdomain.FailedJob, error) {
	job := &auditdomain.FailedJob{}
	err := row.Scan(
		&job.ID,
		&job.EventType,
		&job.QueueName,
		&job.AttemptCount,
		&job.ErrorMessage,
		&job.FailedAt,
		&job.ResolvedAt,
		&job.ResolutionNote,
		&job.CreatedAt,
		&job.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	return job, nil
}
