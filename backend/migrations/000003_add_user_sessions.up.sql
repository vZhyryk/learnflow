-- 000003: add user_sessions table for JWT refresh token storage.

ALTER TABLE user_sessions
    ADD COLUMN token_version          integer NOT NULL DEFAULT 1,
    ADD COLUMN previous_refresh_hash  varchar(64),
    ADD COLUMN failed_attempt_count integer NOT NULL DEFAULT 0,
    ADD COLUMN locked_until timestamptz,
    ADD COLUMN last_attempt_at timestamptz,
    ADD COLUMN revoke_reason VARCHAR(30),
    ADD CONSTRAINT user_sessions_revoke_reason_check
        CHECK (revoke_reason IN ('logout', 'password_changed', 'admin', 'suspicious_activity', 'token_expired')),
    ADD CONSTRAINT user_sessions_token_version_positive CHECK (token_version > 0),
    ADD CONSTRAINT user_sessions_failed_attempts_non_negative CHECK (failed_attempt_count >= 0);


ALTER TABLE users
    DROP CONSTRAINT users_status_check,
    ADD CONSTRAINT users_status_check CHECK (status IN ('active', 'blocked', 'pending_verification')),
    ADD COLUMN password_changed_at timestamptz,
    ADD COLUMN email_changed_at timestamptz

