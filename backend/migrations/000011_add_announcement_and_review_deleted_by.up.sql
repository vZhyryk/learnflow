ALTER TABLE articles
      ADD COLUMN description text;

UPDATE articles SET description = excerpt WHERE excerpt IS NOT NULL;

ALTER TABLE articles
    ADD CONSTRAINT articles_description_nonempty CHECK (description IS NULL OR btrim(description) <> ''),
    DROP CONSTRAINT IF EXISTS articles_excerpt_nonempty,
    DROP COLUMN excerpt,
    ADD COLUMN announcement text,
    ADD CONSTRAINT articles_announcement_nonempty CHECK (announcement IS NULL OR btrim(announcement) <> ''),
    ADD COLUMN announcement_expires_at timestamptz,
    ADD CONSTRAINT articles_announcement_expires_at_after_created CHECK (announcement_expires_at IS NULL OR announcement_expires_at > created_at);

ALTER TABLE courses
    ADD COLUMN announcement text,
    ADD CONSTRAINT courses_announcement_nonempty CHECK (announcement IS NULL OR btrim(announcement) <> ''),
    ADD COLUMN announcement_expires_at timestamptz,
    ADD CONSTRAINT courses_announcement_expires_at_after_created CHECK (announcement_expires_at IS NULL OR announcement_expires_at > created_at),
    ADD CONSTRAINT courses_description_nonempty CHECK (description IS NULL OR btrim(description) <> '');

ALTER TABLE content_items
    ADD COLUMN announcement text,
    ADD CONSTRAINT content_items_announcement_nonempty CHECK (announcement IS NULL OR btrim(announcement) <> ''),
    ADD COLUMN announcement_expires_at timestamptz,
    ADD CONSTRAINT content_items_announcement_expires_at_after_created CHECK (announcement_expires_at IS NULL OR announcement_expires_at > created_at),
    ADD CONSTRAINT content_items_description_nonempty CHECK (description IS NULL OR btrim(description) <> '');

ALTER TABLE course_reviews
    ADD COLUMN deleted_by_user_id uuid REFERENCES users(id);

ALTER TABLE content_reviews
    ADD COLUMN deleted_by_user_id uuid REFERENCES users(id);