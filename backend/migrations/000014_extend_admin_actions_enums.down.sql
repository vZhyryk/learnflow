-- Lossy: rows with action/target types that have no pre-000014 equivalent stay in the table;
-- the old CHECKs are added NOT VALID so the rollback does not fail on them.
ALTER TABLE admin_actions
    DROP CONSTRAINT admin_actions_action_type_check,
    DROP CONSTRAINT admin_actions_target_type_check;

UPDATE admin_actions
SET action_type = CASE action_type
    WHEN 'grant_item_access' THEN 'grant_course_access'
    WHEN 'publish_item'      THEN 'publish_article'
    WHEN 'delete_item'       THEN 'delete_article'
END
WHERE action_type IN ('grant_item_access', 'publish_item', 'delete_item');

ALTER TABLE admin_actions
    ADD CONSTRAINT admin_actions_action_type_check CHECK (action_type IN ('confirm_booking', 'cancel_booking', 'grant_course_access', 'issue_refund', 'record_expense', 'block_user', 'reschedule_booking', 'close_support_chat', 'assign_subadmin', 'revoke_subadmin', 'deactivate_user', 'delete_user', 'create_gift_coupon', 'revoke_gift_coupon', 'publish_article', 'delete_article')) NOT VALID,
    ADD CONSTRAINT admin_actions_target_type_check CHECK (target_type IN ('user', 'booking', 'course', 'failed_job', 'payment', 'support_chat', 'review', 'announcement', 'article', 'gift_coupon')) NOT VALID;
