package adminservice

import (
	"context"
	"fmt"
	admindomain "learnflow_backend/internal/admin/domain"
	auditdomain "learnflow_backend/internal/audit/domain"
	"learnflow_backend/internal/shared/pagination"
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

// canChangeUser reports whether actor may run the operation on target: only an active admin may
// change roles, and a subadmin may act solely on plain users; admins are never a valid target.
func canChangeUser(actor, target *admindomain.UserData, operationName string) bool {
	if actor.DeletedAt != nil || actor.Status != admindomain.StatusActive || target.Role == admindomain.RoleAdmin {
		return false
	}

	switch actor.Role {
	case admindomain.RoleAdmin:
		return true
	case admindomain.RoleSubAdmin:
		isRoleChange := operationName == "AssignUserRole" || operationName == "RevokeUserRole"
		return !isRoleChange && target.Role == admindomain.RoleUser
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

		return nil
	})
}
