ALTER TABLE content_items
    ALTER COLUMN status SET DEFAULT 'draft';


ALTER TABLE articles
    DROP CONSTRAINT articles_status_check,
    ADD CONSTRAINT articles_status_check
        CHECK (status IN ('draft', 'published', 'archived'));
