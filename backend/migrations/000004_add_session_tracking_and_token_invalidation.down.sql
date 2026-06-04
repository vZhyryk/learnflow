ALTER TABLE user_sessions
    DROP COLUMN revoked_by_user_id,
    DROP COLUMN last_seen_ip,
    DROP COLUMN last_seen_at;

ALTER TABLE users
    DROP COLUMN failed_login_count,
    DROP COLUMN last_failed_login_at,
    DROP COLUMN login_locked_until,
    DROP CONSTRAINT users_status_check,
    ADD CONSTRAINT users_status_check CHECK (status IN ('active', 'blocked', 'pending_verification')),
    DROP CONSTRAINT users_failed_login_count_non_negative;

ALTER TABLE account_recovery_tokens
    DROP COLUMN invalidated_at,
    DROP COLUMN invalidated_by_user_id,
    DROP CONSTRAINT account_recovery_tokens_user_id_fkey,
    ADD CONSTRAINT account_recovery_tokens_user_id_fkey
        FOREIGN KEY (user_id)
        REFERENCES users(id)
        ON DELETE CASCADE;


ALTER TABLE email_verification_tokens
    DROP COLUMN invalidated_at,
    DROP COLUMN invalidated_by_user_id;

ALTER TABLE password_reset_tokens
    DROP COLUMN invalidated_at,
    DROP COLUMN invalidated_by_user_id;

ALTER TABLE email_change_tokens
    DROP COLUMN invalidated_at,
    DROP COLUMN invalidated_by_user_id;


DROP INDEX IF EXISTS idx_user_content_access_content_item_id_active;
DROP INDEX IF EXISTS idx_user_sessions_previous_refresh_hash;


DROP INDEX IF EXISTS idx_courses_status;
CREATE INDEX idx_courses_status ON courses(status);

DROP INDEX IF EXISTS idx_content_items_content_type_status;
CREATE INDEX idx_content_items_content_type_status ON content_items(content_type, status);

DROP INDEX IF EXISTS idx_consultation_briefs_status;
CREATE INDEX idx_consultation_briefs_status  ON consultation_briefs(status);

ALTER TABLE gift_coupons
    ADD CONSTRAINT gift_coupons_code_unique UNIQUE (code);

DROP INDEX IF EXISTS gift_coupons_code_lower_unique;