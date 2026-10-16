ALTER TABLE consultation_slots
    ADD COLUMN consultant_id uuid NOT NULL REFERENCES users(id) ON DELETE RESTRICT,
    ADD COLUMN updated_by uuid REFERENCES users(id) ON DELETE RESTRICT,
    DROP COLUMN created_by_user_id,
    DROP CONSTRAINT consultation_slots_date_time_unique,
    ADD CONSTRAINT consultation_slots_consultant_date_time_unique UNIQUE (consultant_id, date, start_time);

CREATE UNIQUE INDEX idx_consultation_briefs_slot_id_active_unique
    ON consultation_briefs(slot_id) WHERE deleted_at IS NULL AND status <> 'failed';

ALTER TABLE admin_actions
    DROP CONSTRAINT admin_actions_target_type_check,
    ADD CONSTRAINT admin_actions_target_type_check CHECK (target_type IN ('user', 'booking', 'course', 'failed_job', 'payment', 'support_chat', 'review', 'announcement', 'article', 'gift_coupon', 'consultation_slot')) NOT VALID;
