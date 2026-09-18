
CREATE TABLE announcement_email_deliveries (
    id uuid  PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id uuid NOT NULL REFERENCES users(id) ON DELETE RESTRICT,
    status text DEFAULT 'pending' NOT NULL CONSTRAINT announcement_email_deliveries_status_check CHECK (status IN ('pending', 'sent', 'failed')),
    announcement_id uuid NOT NULL REFERENCES announcements(id) ON DELETE RESTRICT,
    created_at          timestamptz NOT NULL DEFAULT now(),
    updated_at          timestamptz NOT NULL DEFAULT now(),
    last_error      text
);

CREATE INDEX idx_announcement_email_deliveries_pending
    ON announcement_email_deliveries(status) WHERE status = 'pending';

CREATE UNIQUE INDEX idx_announcement_email_deliveries_user_id_announcement_id_unique
    ON announcement_email_deliveries(user_id, announcement_id);


ALTER TABLE event_outbox
    DROP COLUMN locked_until;