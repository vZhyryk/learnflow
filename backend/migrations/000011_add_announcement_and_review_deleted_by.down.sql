ALTER TABLE articles
      ADD COLUMN excerpt text;

UPDATE articles SET excerpt = description WHERE description IS NOT NULL;

ALTER TABLE articles
    DROP CONSTRAINT IF EXISTS articles_description_nonempty,
    DROP COLUMN description,
    ADD CONSTRAINT articles_excerpt_nonempty CHECK (excerpt IS NULL OR btrim(excerpt) <> ''),
    DROP COLUMN announcement,
    DROP CONSTRAINT IF EXISTS articles_announcement_nonempty,
    DROP COLUMN announcement_expires_at,
    DROP CONSTRAINT IF EXISTS articles_announcement_expires_at_after_created;

ALTER TABLE courses
    DROP CONSTRAINT IF EXISTS courses_announcement_nonempty,
    DROP COLUMN announcement,
    DROP COLUMN announcement_expires_at,
    DROP CONSTRAINT IF EXISTS courses_announcement_expires_at_after_created,
    DROP CONSTRAINT IF EXISTS courses_description_nonempty;



ALTER TABLE content_items
    DROP CONSTRAINT IF EXISTS content_items_announcement_nonempty,
    DROP COLUMN announcement,
    DROP COLUMN announcement_expires_at,
    DROP CONSTRAINT IF EXISTS content_items_announcement_expires_at_after_created,
    DROP CONSTRAINT IF EXISTS content_items_description_nonempty;

ALTER TABLE course_reviews
    DROP COLUMN deleted_by_user_id;

ALTER TABLE content_reviews
    DROP COLUMN deleted_by_user_id;