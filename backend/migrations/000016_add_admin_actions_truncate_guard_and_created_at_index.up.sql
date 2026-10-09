CREATE TRIGGER admin_actions_no_truncate
    BEFORE TRUNCATE ON admin_actions
    FOR EACH STATEMENT EXECUTE FUNCTION admin_actions_reject_modification();

CREATE INDEX idx_admin_actions_created_at_id
    ON admin_actions(created_at DESC, id DESC);
