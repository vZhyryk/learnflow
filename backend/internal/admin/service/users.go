package adminservice

import (
	"context"
	"fmt"
	admindomain "learnflow_backend/internal/admin/domain"
	auditdomain "learnflow_backend/internal/audit/domain"
	"learnflow_backend/internal/events"
	"learnflow_backend/internal/shared/pagination"
	"learnflow_backend/internal/shared/tokens"
	"learnflow_backend/internal/shared/validator"
)

type userOperation struct {
	apply  func(ctx context.Context, userID string) error
	action auditdomain.AdminActionType
}

// GetUsersData returns a page of users and the total user count.
func (srv *Service) GetUsersData(ctx context.Context, params pagination.Params) (users []*admindomain.UserData, total int, err error) {
	err = srv.transactor.InTransaction(ctx, func(ctx context.Context) error {
		users, total, err = srv.userRepo.GetUsersData(ctx, params)
		if err != nil {
			return fmt.Errorf("service.GetUsersData: %w", err)
		}

		return nil
	})

	return users, total, err
}

// GetUserDataByID returns a single user's admin view.
func (srv *Service) GetUserDataByID(ctx context.Context, userID string) (*admindomain.UserData, error) {
	if !validator.IsValidUUID(userID) {
		return nil, admindomain.ErrInvalidID
	}

	user, err := srv.userRepo.GetUserDataByID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("service.GetUserDataByID: %w", err)
	}

	return user, nil
}

// isActiveAccount reports whether the account can act: not soft-deleted and not blocked/pending.
func isActiveAccount(user *admindomain.UserData) bool {
	return user.DeletedAt == nil && user.Status == admindomain.StatusActive
}

// isRoleChange reports whether the operation assigns or revokes the subadmin role.
func isRoleChange(operationName string) bool {
	return operationName == "AssignUserRole" || operationName == "RevokeUserRole"
}

// canChangeUser reports whether actor may run the operation on target.
// Nobody may touch an admin; an admin may do anything else; a subadmin may only act on plain users
// and never change roles.
func canChangeUser(actor, target *admindomain.UserData, operationName string) bool {
	if !isActiveAccount(actor) {
		return false
	}

	if target.Role == admindomain.RoleAdmin {
		return false
	}

	switch actor.Role {
	case admindomain.RoleAdmin:
		return true
	case admindomain.RoleSubAdmin:
		return target.Role == admindomain.RoleUser && !isRoleChange(operationName)
	default:
		return false
	}
}

// ChangeUserField applies the named account operation and records it in admin_actions atomically.
func (srv *Service) ChangeUserField(ctx context.Context, operationName, userID, adminID string) error {
	operations := map[string]userOperation{
		"RevokeUserRole": {srv.userRepo.RevokeUserRole, auditdomain.ActionRevokeSubadmin},
		"AssignUserRole": {srv.userRepo.AssignUserRole, auditdomain.ActionAssignSubadmin},
		"DeleteUser":     {srv.userRepo.DeleteUser, auditdomain.ActionDeleteUser},
		"RestoreUser":    {srv.userRepo.RestoreUser, auditdomain.ActionRestoreUser},
		"BlockUser":      {srv.userRepo.BlockUser, auditdomain.ActionBlockUser},
		"UnBlockUser":    {srv.userRepo.UnBlockUser, auditdomain.ActionUnblockUser},
	}

	operation, ok := operations[operationName]
	if !ok {
		return fmt.Errorf("service.ChangeUserField: invalid operation name: %s", operationName)
	}

	if !validator.IsValidUUID(userID) {
		return admindomain.ErrInvalidID
	}

	if userID == adminID {
		return admindomain.ErrForbiddenUserAction
	}

	return srv.transactor.InTransaction(ctx, func(ctx context.Context) error {
		return srv.runUserOperation(ctx, operationName, operation, userID, adminID)
	})
}

// inUserBlocked reports whether the action must also end the user's sessions and mark them blocked in Redis.
func inUserBlocked(action auditdomain.AdminActionType) bool {
	return action == auditdomain.ActionBlockUser || action == auditdomain.ActionDeleteUser
}

// inUserUnBlocked reports whether the action must also clear the user's Redis block mark.
func inUserUnBlocked(action auditdomain.AdminActionType) bool {
	return action == auditdomain.ActionUnblockUser || action == auditdomain.ActionRestoreUser
}

// runUserOperation authorizes, applies and audits one account operation; it must run inside a transaction.
func (srv *Service) runUserOperation(ctx context.Context, operationName string, operation userOperation, userID, adminID string) error {
	actor, err := srv.userRepo.GetUserDataByID(ctx, adminID)
	if err != nil {
		return fmt.Errorf("service.ChangeUserField: %s: load actor: %w", operationName, err)
	}

	target, err := srv.userRepo.GetUserDataByID(ctx, userID)
	if err != nil {
		return fmt.Errorf("service.ChangeUserField: %s: %w", operationName, err)
	}

	if !canChangeUser(actor, target, operationName) {
		return admindomain.ErrForbiddenUserAction
	}

	if err = operation.apply(ctx, userID); err != nil {
		return fmt.Errorf("service.ChangeUserField: %s: %w", operationName, err)
	}

	err = srv.actionRepo.CreateAdminAction(ctx, &auditdomain.AdminAction{
		AdminUserID: adminID,
		ActionType:  operation.action,
		TargetType:  auditdomain.TargetUser,
		TargetID:    userID,
	})
	if err != nil {
		return fmt.Errorf("service.ChangeUserField: %s audit: %w", operationName, err)
	}

	if err = srv.sendEmail(ctx, operation.action, target); err != nil {
		return fmt.Errorf("service.ChangeUserField: %s: %w", operationName, err)
	}

	if err = srv.applyRedisEffects(ctx, operation.action, userID, adminID); err != nil {
		return fmt.Errorf("service.ChangeUserField: %s: %w", operationName, err)
	}

	return nil
}

// applyRedisEffects runs last, after every in-transaction write: Redis is not rolled back with the DB, so nothing
// that can still fail may run after it.
func (srv *Service) applyRedisEffects(ctx context.Context, action auditdomain.AdminActionType, userID, adminID string) error {
	if inUserBlocked(action) {
		return srv.blockUserRedis(ctx, userID, adminID)
	}

	if inUserUnBlocked(action) {
		return srv.blocklist.UnBlockUser(ctx, userID)
	}

	return nil
}

// blockUserRedis revokes the user's sessions and marks the user blocked in Redis for one access-token lifetime,
// which is enough to invalidate every already-issued token; the DB status keeps the user blocked afterwards.
func (srv *Service) blockUserRedis(ctx context.Context, userID, adminID string) error {
	if err := srv.sessionRepo.RevokeAllUserSessionsAdmin(ctx, userID, adminID); err != nil {
		return fmt.Errorf("revoke sessions: %w", err)
	}

	if err := srv.blocklist.BlockUser(ctx, userID, tokens.AccessTokenTTL); err != nil {
		return fmt.Errorf("set user_blocked: %w", err)
	}

	return nil
}

// sendEmail emits the notification event for the action; a target without an email is skipped so a notification
// problem never blocks the admin action itself.
func (srv *Service) sendEmail(ctx context.Context, action auditdomain.AdminActionType, target *admindomain.UserData) error {
	if target == nil || target.Email == nil || *target.Email == "" {
		return nil
	}

	var eventType events.EventType
	switch action {
	case auditdomain.ActionUnblockUser:
		eventType = events.EventUserUnBlocked
	case auditdomain.ActionBlockUser:
		eventType = events.EventUserBlocked
	case auditdomain.ActionDeleteUser:
		eventType = events.EventUserDeleted
	case auditdomain.ActionRestoreUser:
		eventType = events.EventUserRestored
	default:
		return nil
	}

	_, hash, err := tokens.GenerateSecureToken()
	if err != nil {
		return fmt.Errorf("service.sendEmail: %w", err)
	}

	return srv.outbox.Emit(ctx, events.AggregationTypeUser, target.UserID, eventType, events.UserNotificationPayload{
		UserID:   target.UserID,
		Email:    *target.Email,
		UserName: target.DisplayName(),
		EventID:  hash,
	})
}
