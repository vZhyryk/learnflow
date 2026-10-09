CREATE FUNCTION admin_actions_reject_modification() RETURNS trigger AS $$
BEGIN
    RAISE EXCEPTION 'admin_actions is append-only: % is not allowed', TG_OP
        USING ERRCODE = 'restrict_violation';
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER admin_actions_append_only
    BEFORE UPDATE OR DELETE ON admin_actions
    FOR EACH ROW EXECUTE FUNCTION admin_actions_reject_modification();
