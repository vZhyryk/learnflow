ALTER TABLE articles
    DROP CONSTRAINT IF EXISTS articles_announcement_nonempty,
    DROP CONSTRAINT IF EXISTS articles_announcement_expires_at_after_created,
    DROP COLUMN announcement,
    DROP COLUMN announcement_expires_at;

ALTER TABLE courses
    DROP CONSTRAINT IF EXISTS courses_announcement_nonempty,
    DROP CONSTRAINT IF EXISTS courses_announcement_expires_at_after_created,
    DROP COLUMN announcement,
    DROP COLUMN announcement_expires_at;

ALTER TABLE content_items
    DROP CONSTRAINT IF EXISTS content_items_announcement_nonempty,
    DROP CONSTRAINT IF EXISTS content_items_announcement_expires_at_after_created,
    DROP COLUMN announcement,
    DROP COLUMN announcement_expires_at;


ALTER TABLE announcements
    ADD COLUMN entity_type text,
    ADD COLUMN entity_id   uuid,
    ADD CONSTRAINT announcements_entity_type_check CHECK (entity_type IS NULL OR entity_type IN ('course', 'content_item', 'article')),
    ADD CONSTRAINT announcements_entity_pairing_check CHECK ((entity_type IS NULL) = (entity_id IS NULL)),
    ADD COLUMN channels text[] NOT NULL DEFAULT '{}',
    ADD CONSTRAINT announcements_channels_valid CHECK (channels <@ ARRAY['email','banner','inapp']::text[]),
    ADD COLUMN approved_at timestamptz,
    ADD COLUMN approved_by_user_id uuid REFERENCES users(id) ON DELETE RESTRICT,
    ADD COLUMN updated_by_user_id uuid REFERENCES users(id) ON DELETE RESTRICT,
    ADD CONSTRAINT announcements_approved_by_user_id_and_approved_at CHECK ((approved_at IS NULL) = (approved_by_user_id IS NULL)),
    ADD CONSTRAINT announcements_updated_by_user_id_and_updated_at CHECK ((updated_at IS NULL) = (updated_by_user_id IS NULL)),
    ALTER COLUMN updated_at DROP DEFAULT,
    ALTER COLUMN updated_at DROP NOT NULL;

ALTER TABLE courses
    ADD COLUMN updated_by_user_id uuid REFERENCES users(id) ON DELETE RESTRICT,
    ADD COLUMN published_by_user_id uuid REFERENCES users(id) ON DELETE RESTRICT,
    ADD COLUMN deleted_by_user_id uuid REFERENCES users(id) ON DELETE RESTRICT,
    ADD COLUMN archived_by_user_id uuid REFERENCES users(id) ON DELETE RESTRICT,
    ADD COLUMN archived_at timestamptz,
    ADD CONSTRAINT courses_published_by_user_id_and_published_at CHECK ((published_at IS NULL) = (published_by_user_id IS NULL)),
    ADD CONSTRAINT courses_deleted_by_user_id_and_deleted_at CHECK ((deleted_at IS NULL) = (deleted_by_user_id IS NULL)),
    ADD CONSTRAINT courses_updated_by_user_id_and_updated_at CHECK ((updated_at IS NULL) = (updated_by_user_id IS NULL)),
    ADD CONSTRAINT courses_archived_by_user_id_and_archived_at CHECK ((archived_at IS NULL) = (archived_by_user_id IS NULL)),
    ALTER COLUMN updated_at DROP DEFAULT,
    ALTER COLUMN updated_at DROP NOT NULL;

ALTER TABLE content_items
    ADD COLUMN updated_by_user_id uuid REFERENCES users(id) ON DELETE RESTRICT,
    ADD COLUMN published_by_user_id uuid REFERENCES users(id) ON DELETE RESTRICT,
    ADD COLUMN deleted_by_user_id uuid REFERENCES users(id) ON DELETE RESTRICT,
    ADD COLUMN archived_by_user_id uuid REFERENCES users(id) ON DELETE RESTRICT,
    ADD COLUMN archived_at timestamptz,
    ADD CONSTRAINT content_items_published_by_user_id_and_published_at CHECK ((published_at IS NULL) = (published_by_user_id IS NULL)),
    ADD CONSTRAINT content_items_deleted_by_user_id_and_deleted_at CHECK ((deleted_at IS NULL) = (deleted_by_user_id IS NULL)),
    ADD CONSTRAINT content_items_updated_by_user_id_and_updated_at CHECK ((updated_at IS NULL) = (updated_by_user_id IS NULL)),
    ADD CONSTRAINT content_items_archived_by_user_id_and_archived_at CHECK ((archived_at IS NULL) = (archived_by_user_id IS NULL)),
    ALTER COLUMN updated_at DROP DEFAULT,
    ALTER COLUMN updated_at DROP NOT NULL;


ALTER TABLE articles
    ADD COLUMN updated_by_user_id uuid REFERENCES users(id) ON DELETE RESTRICT,
    ADD COLUMN published_by_user_id uuid REFERENCES users(id) ON DELETE RESTRICT,
    ADD COLUMN deleted_by_user_id uuid REFERENCES users(id) ON DELETE RESTRICT,
    ADD COLUMN archived_by_user_id uuid REFERENCES users(id) ON DELETE RESTRICT,
    ADD COLUMN archived_at timestamptz,
    ADD CONSTRAINT articles_published_by_user_id_and_published_at CHECK ((published_at IS NULL) = (published_by_user_id IS NULL)),
    ADD CONSTRAINT articles_deleted_by_user_id_and_deleted_at CHECK ((deleted_at IS NULL) = (deleted_by_user_id IS NULL)),
    ADD CONSTRAINT articles_archived_by_user_id_and_archived_at CHECK ((archived_at IS NULL) = (archived_by_user_id IS NULL)),
    ADD CONSTRAINT articles_updated_by_user_id_and_updated_at CHECK ((updated_at IS NULL) = (updated_by_user_id IS NULL)),
    ALTER COLUMN updated_at DROP DEFAULT,
    ALTER COLUMN updated_at DROP NOT NULL;

CREATE TABLE article_reviews (
    id              uuid        PRIMARY KEY DEFAULT gen_random_uuid(),
    article_id      uuid        NOT NULL REFERENCES articles(id) ON DELETE RESTRICT,
    user_id         uuid        NOT NULL REFERENCES users(id) ON DELETE RESTRICT,
    rating          integer     NOT NULL CONSTRAINT article_reviews_rating_check CHECK (rating BETWEEN 1 AND 5),
    comment         text,
    created_at      timestamptz NOT NULL DEFAULT now(),
    updated_at      timestamptz NOT NULL DEFAULT now(),
    deleted_at      timestamptz,
    deleted_by_user_id uuid REFERENCES users(id),
    CONSTRAINT article_reviews_deleted_at_after_created CHECK (deleted_at IS NULL OR deleted_at >= created_at),
    CONSTRAINT article_reviews_comment_length_check CHECK (comment IS NULL OR char_length(comment) <= 2000)
);

CREATE UNIQUE INDEX idx_article_reviews_user_id_article_id_active_unique
    ON article_reviews(user_id, article_id) WHERE deleted_at IS NULL;

CREATE INDEX idx_article_reviews_article_id_active ON article_reviews(article_id) WHERE deleted_at IS NULL;
