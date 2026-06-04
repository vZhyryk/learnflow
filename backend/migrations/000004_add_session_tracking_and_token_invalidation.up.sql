-- 000004: extend user_sessions table for enhanced security and token management.

ALTER TABLE user_sessions
    ADD COLUMN revoked_by_user_id uuid REFERENCES users(id) ON DELETE RESTRICT,
    ADD COLUMN last_seen_ip VARCHAR(45),
    ADD COLUMN last_seen_at timestamptz;

ALTER TABLE users
    ADD COLUMN failed_login_count integer NOT NULL DEFAULT 0,
    ADD COLUMN last_failed_login_at timestamptz,
    ADD COLUMN login_locked_until timestamptz,
    DROP CONSTRAINT users_status_check,
    ADD CONSTRAINT users_status_check CHECK (status IN ('active', 'blocked', 'pending_verification', 'deleted')),
    ADD CONSTRAINT users_failed_login_count_non_negative CHECK (failed_login_count >= 0);


ALTER TABLE account_recovery_tokens
    ADD COLUMN invalidated_at timestamptz,
    ADD COLUMN invalidated_by_user_id uuid REFERENCES users(id) ON DELETE RESTRICT,
    DROP CONSTRAINT account_recovery_tokens_user_id_fkey,
    ADD CONSTRAINT account_recovery_tokens_user_id_fkey
        FOREIGN KEY (user_id)
        REFERENCES users(id)
        ON DELETE RESTRICT;

ALTER TABLE email_verification_tokens
    ADD COLUMN invalidated_at timestamptz,
    ADD COLUMN invalidated_by_user_id uuid REFERENCES users(id) ON DELETE RESTRICT;

ALTER TABLE password_reset_tokens
    ADD COLUMN invalidated_at timestamptz,
    ADD COLUMN invalidated_by_user_id uuid REFERENCES users(id) ON DELETE RESTRICT;

ALTER TABLE email_change_tokens
    ADD COLUMN invalidated_at timestamptz,
    ADD COLUMN invalidated_by_user_id uuid REFERENCES users(id) ON DELETE RESTRICT;


CREATE INDEX idx_user_content_access_content_item_id_active
    ON user_content_access(content_item_id) WHERE status = 'active';

CREATE INDEX idx_user_sessions_previous_refresh_hash
    ON user_sessions(previous_refresh_hash)
    WHERE previous_refresh_hash IS NOT NULL;


DROP INDEX IF EXISTS idx_courses_status;
CREATE INDEX idx_courses_status ON courses(status) WHERE deleted_at IS NULL;

DROP INDEX IF EXISTS idx_content_items_content_type_status;
CREATE INDEX idx_content_items_content_type_status ON content_items(content_type, status) WHERE deleted_at IS NULL;

DROP INDEX IF EXISTS idx_consultation_briefs_status;
CREATE INDEX idx_consultation_briefs_status  ON consultation_briefs(status) WHERE deleted_at IS NULL;

ALTER TABLE gift_coupons
    DROP CONSTRAINT gift_coupons_code_unique;

CREATE UNIQUE INDEX gift_coupons_code_lower_unique
    ON gift_coupons (LOWER(code));