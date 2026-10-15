ALTER TABLE content_items
    ALTER COLUMN status DROP DEFAULT;

ALTER TABLE articles
    DROP CONSTRAINT articles_status_check,
    ADD CONSTRAINT articles_status_check
        CHECK (status IN ('draft', 'published'));