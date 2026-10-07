package auditdomain

import (
	"learnflow_backend/internal/shared/validator"
	"time"
)

// AdminActionType is an admin_actions.action_type value.
type AdminActionType string

// AdminTargetType is an admin_actions.target_type value.
type AdminTargetType string

// Admin action types.
const (
	ActionAssignSubadmin    AdminActionType = "assign_subadmin"
	ActionRevokeSubadmin    AdminActionType = "revoke_subadmin"
	ActionDeleteUser        AdminActionType = "delete_user"
	ActionRestoreUser       AdminActionType = "restore_user"
	ActionBlockUser         AdminActionType = "block_user"
	ActionUnblockUser       AdminActionType = "unblock_user"
	ActionConfirmBooking    AdminActionType = "confirm_booking"
	ActionCancelBooking     AdminActionType = "cancel_booking"
	ActionGrantItemAccess   AdminActionType = "grant_item_access"
	ActionIssueRefund       AdminActionType = "issue_refund"
	ActionRecordExpense     AdminActionType = "record_expense"
	ActionRescheduleBooking AdminActionType = "reschedule_booking"
	ActionCloseSupportChat  AdminActionType = "close_support_chat"
	ActionCreateGiftCoupon  AdminActionType = "create_gift_coupon"
	ActionRevokeGiftCoupon  AdminActionType = "revoke_gift_coupon"
	ActionPublishItem       AdminActionType = "publish_item"
	ActionDeleteItem        AdminActionType = "delete_item"
	ActionArchiveItem       AdminActionType = "archive_item"
	ActionCreateItem        AdminActionType = "create_item"
	ActionUpdateItem        AdminActionType = "update_item"
	ActionApproveItem       AdminActionType = "approve_item"
)

// Admin target types.
const (
	TargetUser         AdminTargetType = "user"
	TargetBooking      AdminTargetType = "booking"
	TargetCourse       AdminTargetType = "course"
	TargetFailedJob    AdminTargetType = "failed_job"
	TargetPayment      AdminTargetType = "payment"
	TargetSupportChat  AdminTargetType = "support_chat"
	TargetReview       AdminTargetType = "review"
	TargetAnnouncement AdminTargetType = "announcement"
	TargetArticle      AdminTargetType = "article"
	TargetGiftCoupon   AdminTargetType = "gift_coupon"
	TargetContentItem  AdminTargetType = "content_item"
	TargetExpense      AdminTargetType = "expense"
)

var adminTargetTypes = map[AdminTargetType]struct{}{
	TargetUser:         {},
	TargetBooking:      {},
	TargetCourse:       {},
	TargetFailedJob:    {},
	TargetPayment:      {},
	TargetSupportChat:  {},
	TargetReview:       {},
	TargetAnnouncement: {},
	TargetArticle:      {},
	TargetGiftCoupon:   {},
	TargetContentItem:  {},
	TargetExpense:      {},
}

// AdminAction is an audit-trail entry for an admin operation.
type AdminAction struct {
	ID          string          `json:"id"`
	AdminUserID string          `json:"admin_user_id"`
	AdminName   string          `json:"admin_name"`
	ActionType  AdminActionType `json:"action_type"`
	TargetType  AdminTargetType `json:"target_type"`
	TargetID    string          `json:"target_id"`
	Details     map[string]any  `json:"details"`
	CreatedAt   time.Time       `json:"created_at"`
}

// GetInstanceAdminActionsRequest selects the target whose audit trail is read.
type GetInstanceAdminActionsRequest struct {
	ItemID     string          `json:"item_id"`
	TargetType AdminTargetType `json:"target_type"`
}

// Validate checks that the target id is a UUID and the target type is a known one.
func (req *GetInstanceAdminActionsRequest) Validate() error {
	checks := []func() error{
		req.validateID,
		req.validateItemType,
	}
	for _, check := range checks {
		if err := check(); err != nil {
			return err
		}
	}

	return nil
}

func (req *GetInstanceAdminActionsRequest) validateID() error {
	if req.ItemID == "" || !validator.IsValidUUID(req.ItemID) {
		return ErrInvalidItemID
	}
	return nil
}

func (req *GetInstanceAdminActionsRequest) validateItemType() error {
	if _, exists := adminTargetTypes[req.TargetType]; !exists {
		return ErrInvalidItemType
	}
	return nil
}

// FailedJob is a dead-lettered event; payload_json is deliberately not exposed (it can hold raw tokens).
type FailedJob struct {
	ID             string     `json:"id"`
	EventType      string     `json:"event_type"`
	QueueName      string     `json:"queue_name"`
	AttemptCount   int        `json:"attempt_count"`
	ErrorMessage   *string    `json:"error_message"`
	FailedAt       time.Time  `json:"failed_at"`
	ResolvedAt     *time.Time `json:"resolved_at"`
	ResolutionNote *string    `json:"resolution_note"`
	CreatedAt      time.Time  `json:"created_at"`
	UpdatedAt      time.Time  `json:"updated_at"`
}
