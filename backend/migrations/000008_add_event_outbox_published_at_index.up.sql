CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_event_outbox_published_at_published
    ON event_outbox(published_at) WHERE status = 'published';
