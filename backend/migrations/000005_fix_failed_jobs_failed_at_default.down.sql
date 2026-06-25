ALTER TABLE failed_jobs
    ALTER COLUMN failed_at DROP DEFAULT,
    DROP CONSTRAINT failed_jobs_attempt_count_check,
    ADD CONSTRAINT failed_jobs_attempt_count_check CHECK (attempt_count >= 0);


ALTER TABLE user_sessions
    ALTER COLUMN revoked_by_user_id SET NOT NULL,
    DROP CONSTRAINT user_sessions_revoke_reason_check,
    ADD CONSTRAINT user_sessions_revoke_reason_check
        CHECK (revoke_reason IN ('logout', 'password_changed', 'admin', 'suspicious_activity', 'token_expired'));

DROP INDEX IF EXISTS idx_users_email_deleted_unique;
DROP INDEX IF EXISTS idx_email_verification_tokens_expires_at_unused;
DROP INDEX IF EXISTS idx_password_reset_tokens_expires_at_unused;
DROP INDEX IF EXISTS idx_email_change_tokens_expires_at_unused;
DROP INDEX IF EXISTS idx_account_recovery_tokens_expires_at_unused;