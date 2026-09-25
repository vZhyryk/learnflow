package auditdomain

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

// AdminAction is an audit-trail entry for an admin operation.
type AdminAction struct {
	AdminUserID string
	ActionType  AdminActionType
	TargetType  AdminTargetType
	TargetID    string
}
