ALTER TABLE failed_jobs
    ALTER COLUMN failed_at SET DEFAULT now(),
    DROP CONSTRAINT failed_jobs_attempt_count_check,
    ADD CONSTRAINT failed_jobs_attempt_count_check CHECK (attempt_count >= 0 AND attempt_count <= 10);

ALTER TABLE user_sessions
    ALTER COLUMN revoked_by_user_id DROP NOT NULL,
    DROP CONSTRAINT user_sessions_revoke_reason_check,
    ADD CONSTRAINT user_sessions_revoke_reason_check
        CHECK (revoke_reason IN ('logout', 'password_changed', 'admin', 'suspicious_activity', 'token_expired', 'password_reset', 'email_change'));


CREATE INDEX idx_users_email_deleted ON users(LOWER(email)) WHERE deleted_at IS NOT NULL;

CREATE INDEX idx_email_verification_tokens_expires_at_unused ON email_verification_tokens(expires_at) WHERE used_at IS NULL;
CREATE INDEX idx_password_reset_tokens_expires_at_unused ON password_reset_tokens(expires_at) WHERE used_at IS NULL;
CREATE INDEX idx_email_change_tokens_expires_at_unused ON email_change_tokens(expires_at) WHERE used_at IS NULL;
CREATE INDEX idx_account_recovery_tokens_expires_at_unused ON account_recovery_tokens(expires_at) WHERE used_at IS NULL;