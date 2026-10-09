package adminservice

import (
	"context"
	"errors"
	"fmt"
	admindomain "learnflow_backend/internal/admin/domain"
	auditdomain "learnflow_backend/internal/audit/domain"
	"learnflow_backend/internal/events"
	apperrors "learnflow_backend/internal/shared/errors"
	"learnflow_backend/internal/shared/pagination"
	"learnflow_backend/internal/shared/tokens"
	"learnflow_backend/internal/shared/validator"
	"time"
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
func isRoleChange(operationName admindomain.UserAdminOperation) bool {
	return operationName == admindomain.AssignUserRole || operationName == admindomain.RevokeUserRole
}

// canChangeUser reports whether actor may run the operation on target.
// Nobody may touch an admin; an admin may do anything else; a subadmin may only act on plain users
// and never change roles.
func canChangeUser(actor, target *admindomain.UserData, operationName admindomain.UserAdminOperation) bool {
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

// canGrantAccess reports whether actor may give target free access: an admin may grant to anyone,
// a subadmin only to plain users, so access cannot be handed to admins or to other subadmins.
func canGrantAccess(actor, target *admindomain.UserData) bool {
	if !isActiveAccount(actor) {
		return false
	}

	switch actor.Role {
	case admindomain.RoleAdmin:
		return true
	case admindomain.RoleSubAdmin:
		return target.Role == admindomain.RoleUser
	default:
		return false
	}
}

// ChangeUserField applies the named account operation and records it in admin_actions atomically.
func (srv *Service) ChangeUserField(ctx context.Context, userID, adminID string, operationName admindomain.UserAdminOperation) error {
	operation, ok := srv.operations[operationName]
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

// blocksUser reports whether the action must also end the user's sessions and mark them blocked in Redis.
func blocksUser(action auditdomain.AdminActionType) bool {
	return action == auditdomain.ActionBlockUser || action == auditdomain.ActionDeleteUser
}

// unblocksUser reports whether the action must also clear the user's Redis block mark.
func unblocksUser(action auditdomain.AdminActionType) bool {
	return action == auditdomain.ActionUnblockUser || action == auditdomain.ActionRestoreUser
}

// isUserRevokeOperation reports whether the action must also end the user's sessions and mark their role revoked.
func isUserRevokeOperation(action auditdomain.AdminActionType) bool {
	return action == auditdomain.ActionRevokeSubadmin
}

// isUserAssignOperation reports whether the action must also clear the user's role-revoked mark.
func isUserAssignOperation(action auditdomain.AdminActionType) bool {
	return action == auditdomain.ActionAssignSubadmin
}

// runUserOperation authorizes, applies and audits one account operation; it must run inside a transaction.
func (srv *Service) runUserOperation(ctx context.Context, operationName admindomain.UserAdminOperation, operation userOperation, userID, adminID string) error {
	actor, err := srv.userRepo.GetUserDataByID(ctx, adminID)
	if err != nil {
		return fmt.Errorf("service.ChangeUserField: %s: load actor: %w", operationName, err)
	}

	target, err := srv.userRepo.GetUserDataByID(ctx, userID)
	if err != nil {
		return fmt.Errorf("service.ChangeUserField: %s: load target: %w", operationName, err)
	}

	if !canChangeUser(actor, target, operationName) {
		return admindomain.ErrForbiddenUserAction
	}

	if err = operation.apply(ctx, userID); err != nil {
		return fmt.Errorf("service.ChangeUserField: %s: apply: %w", operationName, err)
	}

	err = srv.actionRepo.CreateAdminAction(ctx, &auditdomain.AdminAction{
		AdminUserID: adminID,
		ActionType:  operation.action,
		TargetType:  auditdomain.TargetUser,
		TargetID:    userID,
	})
	if err != nil {
		return fmt.Errorf("service.ChangeUserField: %s: audit: %w", operationName, err)
	}

	if err = srv.sendEmail(ctx, operation.action, target); err != nil {
		return fmt.Errorf("service.ChangeUserField: %s: send email: %w", operationName, err)
	}

	if err = srv.applyChangeEffects(ctx, operation.action, userID, adminID); err != nil {
		return fmt.Errorf("service.ChangeUserField: %s: apply effects: %w", operationName, err)
	}

	return nil
}

// applyChangeEffects runs last, after every in-transaction write: Redis is not rolled back with the DB, so nothing
// that can still fail may run after it. If COMMIT then fails, the mismatch lasts at most tokens.BlockMarkTTL.
func (srv *Service) applyChangeEffects(ctx context.Context, action auditdomain.AdminActionType, userID, adminID string) error {
	if blocksUser(action) {
		return srv.revokeSessionsAndMark(ctx, userID, adminID, "set user_blocked", srv.blocklist.BlockUser)
	}

	if unblocksUser(action) {
		return blocklistError("clear user_blocked", srv.blocklist.UnBlockUser(ctx, userID))
	}

	if isUserRevokeOperation(action) {
		return srv.revokeSessionsAndMark(ctx, userID, adminID, "set user_role_revoked", srv.blocklist.RevokeUserRole)
	}

	if isUserAssignOperation(action) {
		return blocklistError("clear user_role_revoked", srv.blocklist.ClearUserRoleRevoked(ctx, userID))
	}

	return nil
}

// revokeSessionsAndMark revokes the user's sessions and then sets a Redis mark (named by step) for one access-token
// lifetime: user_blocked invalidates every already-issued token, role_revoked makes RequireRole reject tokens issued
// with the old role. The DB status/role keeps the change afterwards.
func (srv *Service) revokeSessionsAndMark(ctx context.Context, userID, adminID, step string, mark func(ctx context.Context, userID string, ttl time.Duration) error) error {
	if err := srv.sessionRepo.RevokeAllUserSessionsAdmin(ctx, userID, adminID); err != nil {
		return fmt.Errorf("revoke sessions: %w", err)
	}

	return blocklistError(step, mark(ctx, userID, tokens.BlockMarkTTL))
}

// blocklistError tags a failed blocklist write with ErrBlocklistUnavailable so the handler can answer 503; nil passes through.
func blocklistError(step string, err error) error {
	if err == nil {
		return nil
	}

	return fmt.Errorf("%s: %w: %w", step, admindomain.ErrBlocklistUnavailable, err)
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

type grantSpec struct {
	method     string
	itemType   admindomain.ItemType
	itemID     string
	detailsKey string
	lookup     func(ctx context.Context, id string) (string, error)
	grant      func(ctx context.Context, userID, itemID string) error
}

// GrantUserCourseAccess gives a user free access to a course, audits it and queues the notification email.
func (srv *Service) GrantUserCourseAccess(ctx context.Context, userID, courseID, adminID string, operationName auditdomain.AdminActionType) error {
	return srv.grantAccess(ctx, userID, adminID, operationName, grantSpec{
		method:     "GrantUserCourseAccess",
		itemType:   admindomain.CourseItemType,
		itemID:     courseID,
		detailsKey: "course_id",
		lookup:     srv.courseRepo.GetCourseTitleByID,
		grant:      srv.userRepo.GrantUserCourseAccess,
	})
}

// GrantUserContentAccess gives a user free access to a content item, audits it and queues the notification email.
func (srv *Service) GrantUserContentAccess(ctx context.Context, userID, contentItemID, adminID string, operationName auditdomain.AdminActionType) error {
	return srv.grantAccess(ctx, userID, adminID, operationName, grantSpec{
		method:     "GrantUserContentAccess",
		itemType:   admindomain.ContentItemType,
		itemID:     contentItemID,
		detailsKey: "content_item_id",
		lookup:     srv.contentItemRepo.GetContentItemTitleByID,
		grant:      srv.userRepo.GrantUserContentAccess,
	})
}

// authorizeGrant loads the acting admin and the target user, checks that the actor may grant access to the target
// and that the target is a live account with an email, and returns the target.
func (srv *Service) authorizeGrant(ctx context.Context, userID, adminID, method string) (*admindomain.UserData, error) {
	actor, err := srv.userRepo.GetUserDataByID(ctx, adminID)
	if err != nil {
		return nil, fmt.Errorf("service.%s: load actor: %w", method, err)
	}

	user, err := srv.userRepo.GetUserDataByID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("service.%s: get user: %w", method, err)
	}

	if !canGrantAccess(actor, user) {
		return nil, admindomain.ErrForbiddenUserAction
	}

	if user.DeletedAt != nil || user.Email == nil {
		return nil, admindomain.ErrInvalidUserState
	}

	return user, nil
}

func (srv *Service) grantAccess(ctx context.Context, userID, adminID string, action auditdomain.AdminActionType, spec grantSpec) error {
	if !validator.IsValidUUID(userID) {
		return admindomain.ErrInvalidID
	}

	return srv.transactor.InTransaction(ctx, func(ctx context.Context) error {
		title, err := spec.lookup(ctx, spec.itemID)
		if errors.Is(err, apperrors.ErrNotFound) {
			return admindomain.ErrItemNotFound
		}
		if err != nil {
			return fmt.Errorf("service.%s: lookup item: %w", spec.method, err)
		}

		user, err := srv.authorizeGrant(ctx, userID, adminID, spec.method)
		if err != nil {
			return err
		}

		if err = spec.grant(ctx, userID, spec.itemID); err != nil {
			return fmt.Errorf("service.%s: %w", spec.method, err)
		}

		err = srv.actionRepo.CreateAdminAction(ctx, &auditdomain.AdminAction{
			AdminUserID: adminID,
			ActionType:  action,
			TargetType:  auditdomain.TargetUser,
			TargetID:    userID,
			Details:     map[string]any{spec.detailsKey: spec.itemID},
		})
		if err != nil {
			return fmt.Errorf("service.%s: %s: audit: %w", spec.method, action, err)
		}

		return srv.outbox.Emit(ctx, events.AggregationTypeUser, userID, events.EventGrantAccess, events.GrantAccessPayload{
			UserID:   userID,
			ItemName: title,
			ItemID:   spec.itemID,
			ItemType: string(spec.itemType),
			UserName: user.DisplayName(),
			Email:    *user.Email,
		})
	})
}
