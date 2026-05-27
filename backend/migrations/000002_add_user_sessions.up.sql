-- 000002: add user_sessions table for JWT refresh token storage.

CREATE TABLE user_sessions (
    id              uuid        PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id         uuid        NOT NULL REFERENCES users(id) ON DELETE RESTRICT,
    refresh_hash    text        NOT NULL,
    user_agent      text,
    ip              varchar(45),
    expires_at      timestamptz NOT NULL,
    revoked_at      timestamptz,
    created_at      timestamptz NOT NULL DEFAULT now(),
    CONSTRAINT user_sessions_refresh_hash_unique        UNIQUE (refresh_hash),
    CONSTRAINT user_sessions_expires_after_created      CHECK (expires_at > created_at),
    CONSTRAINT user_sessions_revoked_after_created      CHECK (revoked_at IS NULL OR revoked_at >= created_at)
);

-- Primary lookup: validate refresh token
CREATE INDEX idx_user_sessions_refresh_hash_active
    ON user_sessions(refresh_hash)
    WHERE revoked_at IS NULL;

-- Admin/user: list active sessions per user + RevokeAllForUser
CREATE INDEX idx_user_sessions_user_id_active
    ON user_sessions(user_id)
    WHERE revoked_at IS NULL;

-- Cleanup job: expired sessions
CREATE INDEX idx_user_sessions_expires_at
    ON user_sessions(expires_at)
    WHERE revoked_at IS NULL;
