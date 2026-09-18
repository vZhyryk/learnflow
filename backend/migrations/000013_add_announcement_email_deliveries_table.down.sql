DROP TABLE IF EXISTS announcement_email_deliveries;

ALTER TABLE event_outbox
    ADD COLUMN locked_until timestamptz;

CREATE INDEX idx_event_outbox_locked_until
    ON event_outbox(locked_until) WHERE locked_until IS NOT NULL;
