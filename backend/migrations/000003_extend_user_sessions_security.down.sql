ALTER TABLE user_sessions
    DROP COLUMN token_version,
    DROP COLUMN previous_refresh_hash,
    DROP COLUMN failed_attempt_count,
    DROP COLUMN locked_until,
    DROP COLUMN last_attempt_at,
    DROP COLUMN revoke_reason,
    DROP CONSTRAINT IF EXISTS user_sessions_token_version_positive,
    DROP CONSTRAINT IF EXISTS user_sessions_failed_attempts_non_negative,
    DROP CONSTRAINT IF EXISTS user_sessions_revoke_reason_check;

ALTER TABLE users
    DROP CONSTRAINT users_status_check,
    ADD CONSTRAINT users_status_check CHECK (status IN ('active', 'blocked')),
    DROP COLUMN password_changed_at,
    DROP COLUMN email_changed_at;