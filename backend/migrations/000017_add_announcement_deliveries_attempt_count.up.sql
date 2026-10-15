ALTER TABLE announcement_email_deliveries
    ADD COLUMN attempt_count integer NOT NULL DEFAULT 0 CONSTRAINT announcement_email_deliveries_attempt_count_check CHECK (attempt_count >= 0);
