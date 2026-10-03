ALTER TABLE admin_actions
    DROP CONSTRAINT admin_actions_action_type_check,
    DROP CONSTRAINT admin_actions_target_type_check;

UPDATE admin_actions
SET action_type = CASE action_type
    WHEN 'grant_course_access' THEN 'grant_item_access'
    WHEN 'publish_article'     THEN 'publish_item'
    WHEN 'delete_article'      THEN 'delete_item'
    WHEN 'deactivate_user'     THEN 'block_user'
END
WHERE action_type IN ('grant_course_access', 'publish_article', 'delete_article', 'deactivate_user');

ALTER TABLE admin_actions
    ADD CONSTRAINT admin_actions_action_type_check CHECK (action_type IN ('confirm_booking', 'cancel_booking', 'grant_item_access', 'issue_refund', 'record_expense', 'block_user', 'unblock_user', 'reschedule_booking', 'close_support_chat', 'assign_subadmin', 'revoke_subadmin', 'delete_user', 'create_gift_coupon', 'revoke_gift_coupon', 'publish_item', 'delete_item', 'archive_item', 'create_item', 'update_item', 'approve_item', 'restore_user')),
    ADD CONSTRAINT admin_actions_target_type_check CHECK (target_type IN ('user', 'booking', 'course', 'failed_job', 'payment', 'support_chat', 'review', 'announcement', 'article', 'gift_coupon', 'content_item', 'expense'));
