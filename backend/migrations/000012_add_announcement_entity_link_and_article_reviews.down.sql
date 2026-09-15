ALTER TABLE articles
    ADD COLUMN announcement text,
    ADD CONSTRAINT articles_announcement_nonempty CHECK (announcement IS NULL OR btrim(announcement) <> ''),
    ADD COLUMN announcement_expires_at timestamptz,
    ADD CONSTRAINT articles_announcement_expires_at_after_created CHECK (announcement_expires_at IS NULL OR announcement_expires_at > created_at);

ALTER TABLE courses
    ADD COLUMN announcement text,
    ADD CONSTRAINT courses_announcement_nonempty CHECK (announcement IS NULL OR btrim(announcement) <> ''),
    ADD COLUMN announcement_expires_at timestamptz,
    ADD CONSTRAINT courses_announcement_expires_at_after_created CHECK (announcement_expires_at IS NULL OR announcement_expires_at > created_at);

ALTER TABLE content_items
    ADD COLUMN announcement text,
    ADD CONSTRAINT content_items_announcement_nonempty CHECK (announcement IS NULL OR btrim(announcement) <> ''),
    ADD COLUMN announcement_expires_at timestamptz,
    ADD CONSTRAINT content_items_announcement_expires_at_after_created CHECK (announcement_expires_at IS NULL OR announcement_expires_at > created_at);

ALTER TABLE announcements
    DROP CONSTRAINT IF EXISTS announcements_entity_type_check,
    DROP CONSTRAINT IF EXISTS announcements_entity_pairing_check,
    DROP CONSTRAINT IF EXISTS announcements_channels_valid,
    DROP CONSTRAINT IF EXISTS announcements_approved_by_user_id_and_approved_at,
    DROP CONSTRAINT IF EXISTS announcements_updated_by_user_id_and_updated_at,
    DROP COLUMN entity_type,
    DROP COLUMN entity_id,
    DROP COLUMN channels,
    DROP COLUMN approved_at,
    DROP COLUMN approved_by_user_id,
    DROP COLUMN updated_by_user_id,
    ALTER COLUMN updated_at SET DEFAULT now(),
    ALTER COLUMN updated_at SET NOT NULL;

ALTER TABLE courses
    DROP CONSTRAINT IF EXISTS courses_published_by_user_id_and_published_at,
    DROP CONSTRAINT IF EXISTS courses_deleted_by_user_id_and_deleted_at,
    DROP CONSTRAINT IF EXISTS courses_updated_by_user_id_and_updated_at,
    DROP CONSTRAINT IF EXISTS courses_archived_by_user_id_and_archived_at,
    DROP COLUMN updated_by_user_id,
    DROP COLUMN published_by_user_id,
    DROP COLUMN deleted_by_user_id,
    DROP COLUMN archived_at,
    DROP COLUMN archived_by_user_id,
    ALTER COLUMN updated_at SET DEFAULT now(),
    ALTER COLUMN updated_at SET NOT NULL;


ALTER TABLE content_items
    DROP CONSTRAINT IF EXISTS content_items_published_by_user_id_and_published_at,
    DROP CONSTRAINT IF EXISTS content_items_deleted_by_user_id_and_deleted_at,
    DROP CONSTRAINT IF EXISTS content_items_updated_by_user_id_and_updated_at,
    DROP CONSTRAINT IF EXISTS content_items_archived_by_user_id_and_archived_at,
    DROP COLUMN updated_by_user_id,
    DROP COLUMN published_by_user_id,
    DROP COLUMN deleted_by_user_id,
    DROP COLUMN archived_at,
    DROP COLUMN archived_by_user_id,
    ALTER COLUMN updated_at SET DEFAULT now(),
    ALTER COLUMN updated_at SET NOT NULL;


ALTER TABLE articles
    DROP CONSTRAINT IF EXISTS articles_published_by_user_id_and_published_at,
    DROP CONSTRAINT IF EXISTS articles_deleted_by_user_id_and_deleted_at,
    DROP CONSTRAINT IF EXISTS articles_updated_by_user_id_and_updated_at,
    DROP CONSTRAINT IF EXISTS articles_archived_by_user_id_and_archived_at,
    DROP COLUMN updated_by_user_id,
    DROP COLUMN published_by_user_id,
    DROP COLUMN deleted_by_user_id,
    DROP COLUMN archived_at,
    DROP COLUMN archived_by_user_id,
    ALTER COLUMN updated_at SET DEFAULT now(),
    ALTER COLUMN updated_at SET NOT NULL;


DROP TABLE IF EXISTS article_reviews;