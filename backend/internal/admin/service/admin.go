package adminservice

import (
	"context"
	"fmt"
	auditdomain "learnflow_backend/internal/audit/domain"
	"learnflow_backend/internal/shared/pagination"
)

// GetInstanceAdminActions returns one page of the audit trail for a single target, newest first, and the total count.
func (srv *Service) GetInstanceAdminActions(ctx context.Context, targetType auditdomain.AdminTargetType, targetID string, params pagination.Params) (actions []*auditdomain.AdminAction, count int, err error) {
	err = srv.transactor.InTransaction(ctx, func(ctx context.Context) error {
		actions, count, err = srv.actionRepo.GetInstanceAdminActions(ctx, targetType, targetID, params)
		if err != nil {
			return fmt.Errorf("service.GetInstanceAdminActions: %w", err)
		}
		return nil
	})

	return actions, count, err
}

// GetFailedJobs returns one page of failed jobs, newest first, and the total count.
func (srv *Service) GetFailedJobs(ctx context.Context, params pagination.Params) (failedJobs []*auditdomain.FailedJob, count int, err error) {
	err = srv.transactor.InTransaction(ctx, func(ctx context.Context) error {
		failedJobs, count, err = srv.actionRepo.GetFailedJobs(ctx, params)
		if err != nil {
			return fmt.Errorf("service.GetFailedJobs: %w", err)
		}
		return nil
	})

	return failedJobs, count, err
}
