-- Fails if consultation_slots has rows: the restored created_by_user_id is NOT NULL and has no value to backfill.
DROP INDEX IF EXISTS idx_consultation_briefs_slot_id_active_unique;

ALTER TABLE consultation_slots
    DROP CONSTRAINT consultation_slots_consultant_date_time_unique,
    ADD CONSTRAINT consultation_slots_date_time_unique UNIQUE (date, start_time),
    ADD COLUMN created_by_user_id uuid NOT NULL REFERENCES users(id) ON DELETE RESTRICT,
    DROP COLUMN updated_by,
    DROP COLUMN consultant_id;

ALTER TABLE admin_actions
    DROP CONSTRAINT admin_actions_target_type_check,
    ADD CONSTRAINT admin_actions_target_type_check CHECK (target_type IN ('user', 'booking', 'course', 'failed_job', 'payment', 'support_chat', 'review', 'announcement', 'article', 'gift_coupon')) NOT VALID;
