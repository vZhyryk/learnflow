package adminservice

import (
	"context"
	"fmt"
	admindomain "learnflow_backend/internal/admin/domain"
	auditdomain "learnflow_backend/internal/audit/domain"
	"learnflow_backend/internal/shared/pagination"
)

// GetInstanceAdminActions returns one page of the audit trail for a single target, newest first, and the total count.
// A subadmin may only read the trail of plain users, never of admins or other subadmins.
func (srv *Service) GetInstanceAdminActions(ctx context.Context, actorID string, targetType auditdomain.AdminTargetType, targetID string, params pagination.Params) (actions []*auditdomain.AdminAction, count int, err error) {
	err = srv.transactor.InTransaction(ctx, func(ctx context.Context) error {
		if authErr := srv.authorizeInstanceAuditRead(ctx, actorID, targetType, targetID); authErr != nil {
			return fmt.Errorf("service.GetInstanceAdminActions: %w", authErr)
		}

		actions, count, err = srv.actionRepo.GetInstanceAdminActions(ctx, targetType, targetID, params)
		if err != nil {
			return fmt.Errorf("service.GetInstanceAdminActions: %w", err)
		}
		return nil
	})

	return actions, count, err
}

// subAdminAuditTargets lists the non-user trails a subadmin may read: the entities a subadmin manages.
// Everything else (payments, expenses, coupons, bookings, support chats, failed jobs) is admin-only by default.
var subAdminAuditTargets = map[auditdomain.AdminTargetType]struct{}{
	auditdomain.TargetCourse:       {},
	auditdomain.TargetContentItem:  {},
	auditdomain.TargetArticle:      {},
	auditdomain.TargetReview:       {},
	auditdomain.TargetAnnouncement: {},
}

// authorizeInstanceAuditRead lets an admin read any trail; a subadmin reads the trail of plain users and of the
// targets in subAdminAuditTargets, and nothing else.
func (srv *Service) authorizeInstanceAuditRead(ctx context.Context, actorID string, targetType auditdomain.AdminTargetType, targetID string) error {
	actor, err := srv.userRepo.GetUserDataByID(ctx, actorID)
	if err != nil {
		return fmt.Errorf("load actor: %w", err)
	}

	if !isActiveAccount(actor) {
		return admindomain.ErrForbiddenUserAction
	}

	if actor.Role == admindomain.RoleAdmin {
		return nil
	}

	if actor.Role != admindomain.RoleSubAdmin {
		return admindomain.ErrForbiddenUserAction
	}

	if targetType != auditdomain.TargetUser {
		if _, ok := subAdminAuditTargets[targetType]; ok {
			return nil
		}

		return admindomain.ErrForbiddenUserAction
	}

	target, err := srv.userRepo.GetUserDataByID(ctx, targetID)
	if err != nil {
		return fmt.Errorf("load target: %w", err)
	}

	if target.Role == admindomain.RoleUser {
		return nil
	}

	return admindomain.ErrForbiddenUserAction
}

// GetAdminActions returns one page of the general audit journal (all admins, all targets), newest first, and the
// total count of the filter. Only an admin may read it: it also shows what was done to admins.
func (srv *Service) GetAdminActions(ctx context.Context, actorID string, filter auditdomain.AdminActionFilter, params pagination.Params) (actions []*auditdomain.AdminAction, count int, err error) {
	err = srv.transactor.InTransaction(ctx, func(ctx context.Context) error {
		actor, actorErr := srv.userRepo.GetUserDataByID(ctx, actorID)
		if actorErr != nil {
			return fmt.Errorf("service.GetAdminActions: load actor: %w", actorErr)
		}

		if !isActiveAccount(actor) || actor.Role != admindomain.RoleAdmin {
			return admindomain.ErrForbiddenUserAction
		}

		actions, count, err = srv.actionRepo.GetAdminActions(ctx, filter, params)
		if err != nil {
			return fmt.Errorf("service.GetAdminActions: %w", err)
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
