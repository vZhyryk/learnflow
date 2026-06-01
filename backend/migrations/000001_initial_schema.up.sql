-- 000001: complete initial schema — matches infrastructure/postgres/init/init.sql exactly.
-- All tables, constraints, indexes, and views in one migration.

-- Index naming convention: idx_{table}_{col1}_{col2}[_{qualifier}]
--   qualifier = domain condition key: active, available, booked, pending, open, unread, unresolved, unused

CREATE EXTENSION IF NOT EXISTS pg_stat_statements;

-- ---------------------------------------------------------------------------
-- 1. USERS
-- ---------------------------------------------------------------------------

CREATE TABLE users (
    id                  uuid            PRIMARY KEY DEFAULT gen_random_uuid(),
    email               varchar(254)    NOT NULL,
    password_hash       varchar(60)     NOT NULL,
    role                text            NOT NULL CONSTRAINT users_role_check   CHECK (role IN ('user', 'subadmin', 'admin')),
    status              text            NOT NULL CONSTRAINT users_status_check CHECK (status IN ('active', 'blocked')),
    email_verified_at   timestamptz,
    last_login_at       timestamptz,
    deleted_at          timestamptz,
    created_at          timestamptz     NOT NULL DEFAULT now(),
    updated_at          timestamptz     NOT NULL DEFAULT now(),
    CONSTRAINT users_email_nonempty                   CHECK (btrim(email) <> ''),
    CONSTRAINT users_email_verified_at_after_created  CHECK (email_verified_at IS NULL OR email_verified_at >= created_at),
    CONSTRAINT users_last_login_at_after_created      CHECK (last_login_at IS NULL OR last_login_at >= created_at),
    CONSTRAINT users_deleted_at_after_created         CHECK (deleted_at IS NULL OR deleted_at >= created_at)
);

-- Unique email only among non-deleted users (case-insensitive)
CREATE UNIQUE INDEX idx_users_email_active_unique ON users(LOWER(email)) WHERE deleted_at IS NULL;

-- ---------------------------------------------------------------------------
-- 2. USER PROFILES
-- ---------------------------------------------------------------------------

CREATE TABLE user_profiles (
    user_id         uuid            PRIMARY KEY REFERENCES users(id) ON DELETE RESTRICT,
    first_name      text,
    last_name       text,
    phone_number    varchar(20),
    -- ISO 3166-1 alpha-2; exact 2-char code enforced by CHECK
    country         text            CONSTRAINT user_profiles_country_check CHECK (country IS NULL OR char_length(country) = 2),
    city            text,
    date_of_birth   date,
    gender          text            CONSTRAINT user_profiles_gender_check CHECK (gender IN ('male', 'female', 'other', 'prefer_not_to_say')),
    ui_language     text            NOT NULL DEFAULT 'uk',
    avatar_url      text,
    timezone        text,
    bio             text,
    created_at      timestamptz     NOT NULL DEFAULT now(),
    updated_at      timestamptz     NOT NULL DEFAULT now(),
    CONSTRAINT user_profiles_dob_not_future CHECK (date_of_birth IS NULL OR date_of_birth <= CURRENT_DATE),
    CONSTRAINT user_profiles_dob_min        CHECK (date_of_birth IS NULL OR date_of_birth >= '1900-01-01')
);

-- ---------------------------------------------------------------------------
-- 3. AUTH TOKENS
-- ---------------------------------------------------------------------------

CREATE TABLE email_verification_tokens (
    id          uuid        PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id     uuid        NOT NULL REFERENCES users(id) ON DELETE RESTRICT,
    token_hash  varchar(64)        NOT NULL CONSTRAINT email_verification_tokens_token_hash_nonempty CHECK (length(token_hash) > 0),
    expires_at  timestamptz        NOT NULL,
    used_at     timestamptz,
    created_at  timestamptz NOT NULL DEFAULT now(),
    CONSTRAINT email_verification_tokens_token_hash_unique    UNIQUE (token_hash),
    CONSTRAINT email_verification_tokens_expires_after_created CHECK (expires_at > created_at),
    CONSTRAINT email_verification_tokens_used_after_created   CHECK (used_at IS NULL OR used_at >= created_at)
);

CREATE INDEX idx_email_verification_tokens_user_id_unused
    ON email_verification_tokens(user_id) WHERE used_at IS NULL;

CREATE TABLE password_reset_tokens (
    id          uuid        PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id     uuid        NOT NULL REFERENCES users(id) ON DELETE RESTRICT,
    token_hash  varchar(64)        NOT NULL CONSTRAINT password_reset_tokens_token_hash_nonempty CHECK (length(token_hash) > 0),
    expires_at  timestamptz        NOT NULL,
    used_at     timestamptz,
    created_at  timestamptz        NOT NULL DEFAULT now(),
    CONSTRAINT password_reset_tokens_token_hash_unique    UNIQUE (token_hash),
    CONSTRAINT password_reset_tokens_expires_after_created CHECK (expires_at > created_at),
    CONSTRAINT password_reset_tokens_used_after_created   CHECK (used_at IS NULL OR used_at >= created_at)
);

CREATE INDEX idx_password_reset_tokens_user_id_unused
    ON password_reset_tokens(user_id) WHERE used_at IS NULL;

CREATE TABLE email_change_tokens (
    id          uuid            PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id     uuid            NOT NULL REFERENCES users(id) ON DELETE RESTRICT,
    new_email   varchar(254)    NOT NULL,
    token_hash  varchar(64)     NOT NULL CONSTRAINT email_change_tokens_token_hash_nonempty CHECK (length(token_hash) > 0),
    expires_at  timestamptz     NOT NULL,
    used_at     timestamptz,
    created_at  timestamptz     NOT NULL DEFAULT now(),
    CONSTRAINT email_change_tokens_token_hash_unique    UNIQUE (token_hash),
    CONSTRAINT email_change_tokens_expires_after_created CHECK (expires_at > created_at),
    CONSTRAINT email_change_tokens_used_after_created   CHECK (used_at IS NULL OR used_at >= created_at)
);

CREATE INDEX idx_email_change_tokens_user_id_unused
    ON email_change_tokens(user_id) WHERE used_at IS NULL;

-- ---------------------------------------------------------------------------
-- 4. COURSES
-- ---------------------------------------------------------------------------

CREATE TABLE courses (
    id                  uuid        PRIMARY KEY DEFAULT gen_random_uuid(),
    slug                text        NOT NULL,
    title               text        NOT NULL,
    description         text,
    thumbnail_url       text,
    preview_video_url   text,
    status              text        NOT NULL CONSTRAINT courses_status_check CHECK (status IN ('draft', 'published', 'archived')),
    estimated_minutes   integer     CONSTRAINT courses_estimated_minutes_positive CHECK (estimated_minutes IS NULL OR estimated_minutes > 0),
    seo_title           text,
    seo_description     text,
    og_image_url        text,
    canonical_url       text,
    is_indexable        boolean     NOT NULL DEFAULT true,
    created_by_user_id  uuid        NOT NULL REFERENCES users(id) ON DELETE RESTRICT,
    created_at          timestamptz NOT NULL DEFAULT now(),
    updated_at          timestamptz NOT NULL DEFAULT now(),
    published_at        timestamptz,
    deleted_at          timestamptz,
    CONSTRAINT courses_slug_unique                     UNIQUE (slug),
    CONSTRAINT courses_title_nonempty                  CHECK (btrim(title) <> ''),
    CONSTRAINT courses_slug_nonempty                   CHECK (btrim(slug) <> ''),
    CONSTRAINT courses_published_at_after_created      CHECK (published_at IS NULL OR published_at >= created_at),
    CONSTRAINT courses_deleted_at_after_created        CHECK (deleted_at IS NULL OR deleted_at >= created_at),
    CONSTRAINT courses_published_requires_published_at CHECK (status != 'published' OR published_at IS NOT NULL)
);

CREATE INDEX idx_courses_status ON courses(status);

-- ---------------------------------------------------------------------------
-- 5. CONTENT ITEMS
-- ---------------------------------------------------------------------------

CREATE TABLE content_items (
    id                  uuid        PRIMARY KEY DEFAULT gen_random_uuid(),
    slug                text        NOT NULL,
    title               text        NOT NULL,
    content_type        text        NOT NULL CONSTRAINT content_items_content_type_check CHECK (content_type IN ('video', 'book', 'presentation')),
    description         text,
    body                text,
    video_url           text,
    file_url            text,
    estimated_minutes   integer     CONSTRAINT content_items_estimated_minutes_positive CHECK (estimated_minutes IS NULL OR estimated_minutes > 0),
    estimated_pages     integer     CONSTRAINT content_items_estimated_pages_positive   CHECK (estimated_pages IS NULL OR estimated_pages > 0),
    thumbnail_url       text,
    seo_title           text,
    seo_description     text,
    og_image_url        text,
    canonical_url       text,
    is_indexable        boolean     NOT NULL DEFAULT true,
    status              text        NOT NULL CONSTRAINT content_items_status_check CHECK (status IN ('draft', 'published', 'archived')),
    created_by_user_id  uuid        NOT NULL REFERENCES users(id) ON DELETE RESTRICT,
    created_at          timestamptz NOT NULL DEFAULT now(),
    updated_at          timestamptz NOT NULL DEFAULT now(),
    published_at        timestamptz,
    deleted_at          timestamptz,
    CONSTRAINT content_items_slug_unique                     UNIQUE (slug),
    CONSTRAINT content_items_title_nonempty                  CHECK (btrim(title) <> ''),
    CONSTRAINT content_items_slug_nonempty                   CHECK (btrim(slug) <> ''),
    CONSTRAINT content_items_published_at_after_created      CHECK (published_at IS NULL OR published_at >= created_at),
    CONSTRAINT content_items_deleted_at_after_created        CHECK (deleted_at IS NULL OR deleted_at >= created_at),
    CONSTRAINT content_items_published_requires_published_at CHECK (status != 'published' OR published_at IS NOT NULL)
);

CREATE INDEX idx_content_items_content_type_status ON content_items(content_type, status);

-- ---------------------------------------------------------------------------
-- 6. COURSE <-> CONTENT ITEMS (junction)
-- ---------------------------------------------------------------------------

CREATE TABLE course_content_items (
    id              uuid        PRIMARY KEY DEFAULT gen_random_uuid(),
    course_id       uuid        NOT NULL REFERENCES courses(id) ON DELETE RESTRICT,
    content_item_id uuid        NOT NULL REFERENCES content_items(id) ON DELETE RESTRICT,
    position        integer     NOT NULL CONSTRAINT course_content_items_position_positive CHECK (position > 0),
    is_required     boolean     NOT NULL,
    created_at      timestamptz NOT NULL DEFAULT now(),
    CONSTRAINT course_content_items_unique_pair     UNIQUE (course_id, content_item_id),
    CONSTRAINT course_content_items_unique_position UNIQUE (course_id, position) DEFERRABLE INITIALLY DEFERRED
);

CREATE INDEX idx_course_content_items_content_item_id ON course_content_items(content_item_id);

-- ---------------------------------------------------------------------------
-- 7. COURSE REVIEWS
-- ---------------------------------------------------------------------------

CREATE TABLE course_reviews (
    id          uuid        PRIMARY KEY DEFAULT gen_random_uuid(),
    course_id   uuid        NOT NULL REFERENCES courses(id) ON DELETE RESTRICT,
    user_id     uuid        NOT NULL REFERENCES users(id) ON DELETE RESTRICT,
    rating      integer     NOT NULL CONSTRAINT course_reviews_rating_check CHECK (rating BETWEEN 1 AND 5),
    comment     text,
    created_at  timestamptz NOT NULL DEFAULT now(),
    updated_at  timestamptz NOT NULL DEFAULT now(),
    deleted_at  timestamptz,
    CONSTRAINT course_reviews_deleted_at_after_created CHECK (deleted_at IS NULL OR deleted_at >= created_at)
);

CREATE UNIQUE INDEX idx_course_reviews_user_id_course_id_active_unique
    ON course_reviews(user_id, course_id) WHERE deleted_at IS NULL;

CREATE INDEX idx_course_reviews_course_id_active ON course_reviews(course_id) WHERE deleted_at IS NULL;

-- ---------------------------------------------------------------------------
-- 8. CONTENT REVIEWS
-- ---------------------------------------------------------------------------

CREATE TABLE content_reviews (
    id              uuid        PRIMARY KEY DEFAULT gen_random_uuid(),
    content_item_id uuid        NOT NULL REFERENCES content_items(id) ON DELETE RESTRICT,
    user_id         uuid        NOT NULL REFERENCES users(id) ON DELETE RESTRICT,
    rating          integer     NOT NULL CONSTRAINT content_reviews_rating_check CHECK (rating BETWEEN 1 AND 5),
    comment         text,
    created_at      timestamptz NOT NULL DEFAULT now(),
    updated_at      timestamptz NOT NULL DEFAULT now(),
    deleted_at      timestamptz,
    CONSTRAINT content_reviews_deleted_at_after_created CHECK (deleted_at IS NULL OR deleted_at >= created_at)
);

CREATE UNIQUE INDEX idx_content_reviews_user_id_content_item_id_active_unique
    ON content_reviews(user_id, content_item_id) WHERE deleted_at IS NULL;

CREATE INDEX idx_content_reviews_content_item_id_active ON content_reviews(content_item_id) WHERE deleted_at IS NULL;

-- ---------------------------------------------------------------------------
-- 9. PAYMENTS
-- (created before user_course_access, user_content_access, and consultation_bookings)
-- ---------------------------------------------------------------------------

CREATE TABLE payments (
    id                  uuid        PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id             uuid        NOT NULL REFERENCES users(id) ON DELETE RESTRICT,
    amount              bigint      NOT NULL CONSTRAINT payments_amount_positive CHECK (amount > 0),
    currency            text        NOT NULL DEFAULT 'PLN',
    status              text        NOT NULL CONSTRAINT payments_status_check         CHECK (status IN ('pending', 'completed', 'failed')),
    payment_method      text        NOT NULL CONSTRAINT payments_payment_method_check CHECK (payment_method IN ('card', 'bank_transfer', 'manual')),
    provider_reference  text        CONSTRAINT payments_provider_reference_check CHECK (provider_reference IS NULL OR length(btrim(provider_reference)) > 0),
    utm_source          text,
    utm_medium          text,
    utm_campaign        text,
    utm_content         text,
    utm_term            text,
    gclid               text,
    created_at          timestamptz NOT NULL DEFAULT now(),
    updated_at          timestamptz NOT NULL DEFAULT now(),
    CONSTRAINT payments_currency_length CHECK (char_length(currency) = 3)
);

CREATE INDEX idx_payments_user_id_created_at ON payments(user_id, created_at DESC);
CREATE INDEX idx_payments_status_created_at  ON payments(status, created_at);
CREATE INDEX idx_payments_utm_campaign       ON payments(utm_campaign)       WHERE utm_campaign IS NOT NULL;
CREATE INDEX idx_payments_provider_reference ON payments(provider_reference) WHERE provider_reference IS NOT NULL;

CREATE TABLE payment_line_items (
    id              uuid        PRIMARY KEY DEFAULT gen_random_uuid(),
    payment_id      uuid        NOT NULL REFERENCES payments(id) ON DELETE RESTRICT,
    resource_type   text        NOT NULL CONSTRAINT payment_line_items_resource_type_check CHECK (resource_type IN ('course', 'consultation', 'content_item')),
    resource_id     uuid        NOT NULL,
    amount          bigint      NOT NULL CONSTRAINT payment_line_items_amount_positive CHECK (amount > 0),
    description     text,
    created_at      timestamptz NOT NULL DEFAULT now()
);

CREATE INDEX idx_payment_line_items_payment_id
    ON payment_line_items(payment_id);
-- Lookup: "which payment covers this resource?"
CREATE INDEX idx_payment_line_items_resource_type_resource_id
    ON payment_line_items(resource_type, resource_id);

CREATE TABLE refunds (
    id                  uuid        PRIMARY KEY DEFAULT gen_random_uuid(),
    payment_id          uuid        NOT NULL REFERENCES payments(id) ON DELETE RESTRICT,
    amount              bigint      NOT NULL CONSTRAINT refunds_amount_positive CHECK (amount > 0),
    reason              text,
    created_by_user_id  uuid        NOT NULL REFERENCES users(id) ON DELETE RESTRICT,
    refunded_at         timestamptz NOT NULL,
    created_at          timestamptz NOT NULL DEFAULT now(),
    CONSTRAINT refunds_refunded_at_after_created CHECK (refunded_at >= created_at)
);

CREATE INDEX idx_refunds_payment_id ON refunds(payment_id);

-- ---------------------------------------------------------------------------
-- 10. USER COURSE ACCESS & PROGRESS
-- ---------------------------------------------------------------------------

CREATE TABLE user_course_access (
    id          uuid        PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id     uuid        NOT NULL REFERENCES users(id) ON DELETE RESTRICT,
    course_id   uuid        NOT NULL REFERENCES courses(id) ON DELETE RESTRICT,
    payment_id  uuid        REFERENCES payments(id) ON DELETE RESTRICT,
    access_type text        NOT NULL CONSTRAINT user_course_access_access_type_check CHECK (access_type IN ('purchased', 'admin_granted', 'gift_coupon')),
    status      text        NOT NULL CONSTRAINT user_course_access_status_check      CHECK (status IN ('active', 'revoked', 'refunded')),
    granted_at  timestamptz NOT NULL,
    expires_at  timestamptz,
    created_at  timestamptz NOT NULL DEFAULT now(),
    updated_at  timestamptz NOT NULL DEFAULT now(),
    CONSTRAINT user_course_access_granted_at_after_created CHECK (granted_at >= created_at),
    CONSTRAINT user_course_access_expires_after_granted   CHECK (expires_at IS NULL OR expires_at > granted_at)
);

-- Partial unique: allows re-purchase after refund, only one active access per course
CREATE UNIQUE INDEX idx_user_course_access_user_id_course_id_active_unique
    ON user_course_access(user_id, course_id) WHERE status = 'active';

-- Admin query: "who has access to this course?"
CREATE INDEX idx_user_course_access_course_id_active
    ON user_course_access(course_id) WHERE status = 'active';

CREATE TABLE user_course_progress (
    id                  uuid        PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id             uuid        NOT NULL REFERENCES users(id) ON DELETE RESTRICT,
    course_id           uuid        NOT NULL REFERENCES courses(id) ON DELETE RESTRICT,
    status              text        NOT NULL CONSTRAINT user_course_progress_status_check  CHECK (status IN ('not_started', 'in_progress', 'completed')),
    progress_percent    integer     NOT NULL DEFAULT 0 CONSTRAINT user_course_progress_percent_check CHECK (progress_percent BETWEEN 0 AND 100),
    started_at          timestamptz,
    last_interacted_at  timestamptz,
    completed_at        timestamptz,
    created_at          timestamptz NOT NULL DEFAULT now(),
    updated_at          timestamptz NOT NULL DEFAULT now(),
    CONSTRAINT user_course_progress_unique                       UNIQUE (user_id, course_id),
    CONSTRAINT user_course_progress_started_at_after_created     CHECK (started_at IS NULL OR started_at >= created_at),
    CONSTRAINT user_course_progress_last_interacted_after_created CHECK (last_interacted_at IS NULL OR last_interacted_at >= created_at),
    CONSTRAINT user_course_progress_completed_at_after_created   CHECK (completed_at IS NULL OR completed_at >= created_at),
    CONSTRAINT user_course_progress_started_when_active          CHECK (status = 'not_started' OR started_at IS NOT NULL),
    CONSTRAINT user_course_progress_completed_at_100             CHECK (status != 'completed' OR progress_percent = 100)
);

CREATE TABLE user_content_progress (
    id                  uuid        PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id             uuid        NOT NULL REFERENCES users(id) ON DELETE RESTRICT,
    content_item_id     uuid        NOT NULL REFERENCES content_items(id) ON DELETE RESTRICT,
    status              text        NOT NULL CONSTRAINT user_content_progress_status_check  CHECK (status IN ('not_started', 'in_progress', 'completed')),
    progress_percent    integer     NOT NULL DEFAULT 0 CONSTRAINT user_content_progress_percent_check CHECK (progress_percent BETWEEN 0 AND 100),
    time_spent_minutes  integer,
    started_at          timestamptz,
    last_interacted_at  timestamptz,
    completed_at        timestamptz,
    created_at          timestamptz NOT NULL DEFAULT now(),
    updated_at          timestamptz NOT NULL DEFAULT now(),
    CONSTRAINT user_content_progress_unique                       UNIQUE (user_id, content_item_id),
    CONSTRAINT user_content_progress_time_spent_non_negative      CHECK (time_spent_minutes IS NULL OR time_spent_minutes >= 0),
    CONSTRAINT user_content_progress_started_at_after_created     CHECK (started_at IS NULL OR started_at >= created_at),
    CONSTRAINT user_content_progress_last_interacted_after_created CHECK (last_interacted_at IS NULL OR last_interacted_at >= created_at),
    CONSTRAINT user_content_progress_completed_at_after_created   CHECK (completed_at IS NULL OR completed_at >= created_at),
    CONSTRAINT user_content_progress_started_when_active          CHECK (status = 'not_started' OR started_at IS NOT NULL),
    CONSTRAINT user_content_progress_completed_at_100             CHECK (status != 'completed' OR progress_percent = 100)
);

CREATE TABLE user_video_engagement (
    id                      uuid        PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id                 uuid        NOT NULL REFERENCES users(id) ON DELETE RESTRICT,
    content_item_id         uuid        NOT NULL REFERENCES content_items(id) ON DELETE RESTRICT,
    watched_seconds         integer     NOT NULL DEFAULT 0 CONSTRAINT user_video_engagement_watched_seconds_check CHECK (watched_seconds >= 0),
    last_position_seconds   integer,
    started_at              timestamptz,
    last_watched_at         timestamptz,
    completed_at            timestamptz,
    created_at              timestamptz NOT NULL DEFAULT now(),
    updated_at              timestamptz NOT NULL DEFAULT now(),
    CONSTRAINT user_video_engagement_unique UNIQUE (user_id, content_item_id)
);

CREATE TABLE user_document_engagement (
    id                  uuid        PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id             uuid        NOT NULL REFERENCES users(id) ON DELETE RESTRICT,
    content_item_id     uuid        NOT NULL REFERENCES content_items(id) ON DELETE RESTRICT,
    viewed_pages        integer,
    last_page           integer,
    time_spent_minutes  integer,
    started_at          timestamptz,
    last_viewed_at      timestamptz,
    completed_at        timestamptz,
    created_at          timestamptz NOT NULL DEFAULT now(),
    updated_at          timestamptz NOT NULL DEFAULT now(),
    CONSTRAINT user_document_engagement_unique UNIQUE (user_id, content_item_id)
);

-- ---------------------------------------------------------------------------
-- 11. USER CONTENT ACCESS
-- (individual content item purchases, separate from course-level access)
-- ---------------------------------------------------------------------------

CREATE TABLE user_content_access (
    id              uuid        PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id         uuid        NOT NULL REFERENCES users(id) ON DELETE RESTRICT,
    content_item_id uuid        NOT NULL REFERENCES content_items(id) ON DELETE RESTRICT,
    payment_id      uuid        REFERENCES payments(id) ON DELETE RESTRICT,
    access_type     text        NOT NULL CONSTRAINT user_content_access_access_type_check CHECK (access_type IN ('purchased', 'admin_granted', 'gift_coupon')),
    status          text        NOT NULL CONSTRAINT user_content_access_status_check      CHECK (status IN ('active', 'revoked', 'refunded')),
    granted_at      timestamptz NOT NULL,
    expires_at      timestamptz,
    created_at      timestamptz NOT NULL DEFAULT now(),
    updated_at      timestamptz NOT NULL DEFAULT now()
);

CREATE UNIQUE INDEX idx_user_content_access_user_id_content_item_id_active_unique
    ON user_content_access(user_id, content_item_id) WHERE status = 'active';

-- ---------------------------------------------------------------------------
-- 12. USER NOTES
-- ---------------------------------------------------------------------------

CREATE TABLE user_notes (
    user_id       uuid        NOT NULL REFERENCES users(id) ON DELETE RESTRICT,
    id            uuid        PRIMARY KEY DEFAULT gen_random_uuid(),
    resource_type text        CONSTRAINT user_notes_resource_type_check CHECK (resource_type IN ('course', 'content_item')),
    resource_id   uuid,
    title         text        NOT NULL,
    description   text,
    body          text        NOT NULL,
    deleted_at    timestamptz,
    created_at    timestamptz NOT NULL DEFAULT now(),
    updated_at    timestamptz NOT NULL DEFAULT now(),
    CONSTRAINT user_notes_resource_check CHECK (
        (resource_type IS NULL AND resource_id IS NULL) OR
        (resource_type IS NOT NULL AND resource_id IS NOT NULL)
    )
);

CREATE INDEX idx_user_notes_user_id_active
    ON user_notes(user_id) WHERE deleted_at IS NULL;
CREATE INDEX idx_user_notes_resource_type_resource_id_active
    ON user_notes(resource_type, resource_id) WHERE resource_type IS NOT NULL AND deleted_at IS NULL;

-- ---------------------------------------------------------------------------
-- 13. CONSULTATIONS
-- ---------------------------------------------------------------------------

-- date + start_time stored in Europe/Warsaw timezone; app layer handles DST.
CREATE TABLE consultation_slots (
    id                 uuid        PRIMARY KEY DEFAULT gen_random_uuid(),
    date               date        NOT NULL,
    start_time         time        NOT NULL,
    duration_minutes   integer     NOT NULL CONSTRAINT consultation_slots_duration_positive CHECK (duration_minutes > 0),
    status             text        NOT NULL DEFAULT 'available' CONSTRAINT consultation_slots_status_check CHECK (status IN ('available', 'booked', 'blocked')),
    created_by_user_id uuid        NOT NULL REFERENCES users(id) ON DELETE RESTRICT,
    created_at         timestamptz NOT NULL DEFAULT now(),
    updated_at         timestamptz NOT NULL DEFAULT now(),
    CONSTRAINT consultation_slots_date_time_unique UNIQUE (date, start_time)
);

CREATE INDEX idx_consultation_slots_date_start_time_available
    ON consultation_slots(date, start_time) WHERE status = 'available';

CREATE INDEX idx_consultation_slots_date_start_time_booked
    ON consultation_slots(date, start_time) WHERE status = 'booked';

CREATE TABLE consultation_briefs (
    id            uuid        PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id       uuid        NOT NULL REFERENCES users(id) ON DELETE RESTRICT,
    slot_id       uuid        NOT NULL REFERENCES consultation_slots(id) ON DELETE RESTRICT,
    goal_type     text        CONSTRAINT consultation_briefs_goal_type_check CHECK (goal_type IN ('bring_partner_back', 'recover_after_breakup', 'improve_good_relationship', 'recover_problematic_relationship', 'improve_meeting_skills', 'find_good_partner_skills', 'analyze_situation', 'check_partner')),
    goal          text,
    context       text,
    review_notes  text,
    metadata_json jsonb,
    status        text        NOT NULL CONSTRAINT consultation_briefs_status_check CHECK (status IN ('submitted', 'processing', 'booked', 'failed')),
    submitted_at  timestamptz NOT NULL DEFAULT now(),
    created_at    timestamptz NOT NULL DEFAULT now(),
    updated_at    timestamptz NOT NULL DEFAULT now(),
    deleted_at    timestamptz
);

CREATE INDEX idx_consultation_briefs_user_id ON consultation_briefs(user_id);
CREATE INDEX idx_consultation_briefs_status  ON consultation_briefs(status);

CREATE TABLE consultation_bookings (
    id                          uuid        PRIMARY KEY DEFAULT gen_random_uuid(),
    brief_id                    uuid        NOT NULL REFERENCES consultation_briefs(id) ON DELETE RESTRICT,
    user_id                     uuid        NOT NULL REFERENCES users(id) ON DELETE RESTRICT,
    slot_id                     uuid        NOT NULL REFERENCES consultation_slots(id) ON DELETE RESTRICT,
    payment_id                  uuid        REFERENCES payments(id) ON DELETE RESTRICT,
    status                      text        NOT NULL CONSTRAINT consultation_bookings_status_check CHECK (status IN ('created', 'confirmed', 'cancelled', 'rescheduled', 'completed')),
    -- Snapshot of the slot datetime at booking time; preserved across reschedules for history
    scheduled_at                timestamptz,
    confirmed_at                timestamptz,
    cancelled_at                timestamptz,
    completed_at                timestamptz,
    rescheduled_from_booking_id uuid        REFERENCES consultation_bookings(id),
    handled_by_user_id          uuid        REFERENCES users(id) ON DELETE RESTRICT,
    notes                       text,
    created_at                  timestamptz NOT NULL DEFAULT now(),
    updated_at                  timestamptz NOT NULL DEFAULT now(),
    deleted_at                  timestamptz
);

-- One original booking per brief; reschedules create new rows with rescheduled_from_booking_id set
CREATE UNIQUE INDEX idx_consultation_bookings_brief_id_unique
    ON consultation_bookings(brief_id) WHERE rescheduled_from_booking_id IS NULL;

CREATE INDEX idx_consultation_bookings_brief_id       ON consultation_bookings(brief_id);
CREATE INDEX idx_consultation_bookings_slot_id        ON consultation_bookings(slot_id);
CREATE INDEX idx_consultation_bookings_user_id_status ON consultation_bookings(user_id, status);
CREATE INDEX idx_consultation_bookings_scheduled_at
    ON consultation_bookings(scheduled_at) WHERE scheduled_at IS NOT NULL;

-- ---------------------------------------------------------------------------
-- 14. CONSULTATION OUTCOMES
-- ---------------------------------------------------------------------------

CREATE TABLE consultation_outcomes (
    id              uuid        PRIMARY KEY DEFAULT gen_random_uuid(),
    brief_id        uuid        NOT NULL REFERENCES consultation_briefs(id) ON DELETE RESTRICT,
    recommendation  text,
    knowledge       text,
    created_at      timestamptz NOT NULL DEFAULT now(),
    updated_at      timestamptz NOT NULL DEFAULT now(),
    CONSTRAINT consultation_outcomes_brief_unique UNIQUE (brief_id)
);

-- ---------------------------------------------------------------------------
-- 15. CAMPAIGNS
-- ---------------------------------------------------------------------------

-- id is the external campaign ID from the ad platform (e.g. Google Ads campaign ID),
-- stored as text because external IDs are not UUIDs.
CREATE TABLE campaigns (
    id          text        PRIMARY KEY,
    name        text        NOT NULL CONSTRAINT campaigns_name_nonempty CHECK (btrim(name) <> ''),
    platform    text        NOT NULL DEFAULT 'google' CONSTRAINT campaigns_platform_check CHECK (platform IN ('google', 'meta', 'tiktok', 'other')),
    created_at  timestamptz NOT NULL DEFAULT now(),
    updated_at  timestamptz NOT NULL DEFAULT now()
);

-- ---------------------------------------------------------------------------
-- 16. EXPENSES
-- ---------------------------------------------------------------------------

CREATE TABLE expenses (
    id                  uuid        PRIMARY KEY DEFAULT gen_random_uuid(),
    category            text        NOT NULL CONSTRAINT expenses_category_check CHECK (category IN ('hosting', 'tooling', 'contractor', 'marketing', 'other')),
    amount              bigint      NOT NULL CONSTRAINT expenses_amount_positive CHECK (amount > 0),
    currency            text        NOT NULL DEFAULT 'PLN',
    description         text,
    occurred_at         timestamptz NOT NULL,
    vendor              text,
    campaign_id         text        REFERENCES campaigns(id) ON DELETE RESTRICT,
    course_id           uuid        REFERENCES courses(id) ON DELETE RESTRICT,
    content_item_id     uuid        REFERENCES content_items(id) ON DELETE RESTRICT,
    external_reference  text,
    metadata_json       jsonb,
    created_by_user_id  uuid        NOT NULL REFERENCES users(id) ON DELETE RESTRICT,
    created_at          timestamptz NOT NULL DEFAULT now(),
    updated_at          timestamptz NOT NULL DEFAULT now(),
    CONSTRAINT expenses_resource_exclusive CHECK (course_id IS NULL OR content_item_id IS NULL)
);

CREATE INDEX idx_expenses_occurred_at     ON expenses(occurred_at);
CREATE INDEX idx_expenses_campaign_id     ON expenses(campaign_id)     WHERE campaign_id IS NOT NULL;
CREATE INDEX idx_expenses_course_id       ON expenses(course_id)       WHERE course_id IS NOT NULL;
CREATE INDEX idx_expenses_content_item_id ON expenses(content_item_id) WHERE content_item_id IS NOT NULL;

-- ---------------------------------------------------------------------------
-- 17. NOTIFICATION PREFERENCES
-- ---------------------------------------------------------------------------

CREATE TABLE notification_preferences (
    user_id                      uuid        PRIMARY KEY REFERENCES users(id) ON DELETE RESTRICT,
    email_on_booking_confirmed   boolean     NOT NULL DEFAULT true,
    email_on_reminder            boolean     NOT NULL DEFAULT true,
    email_on_new_content         boolean     NOT NULL DEFAULT true,
    email_on_support_reply       boolean     NOT NULL DEFAULT true,
    email_on_consultation_booked boolean     NOT NULL DEFAULT true,
    email_on_support_message     boolean     NOT NULL DEFAULT true,
    email_on_announcement        boolean     NOT NULL DEFAULT true,
    email_on_failed_job          boolean     NOT NULL DEFAULT true,
    created_at                   timestamptz NOT NULL DEFAULT now(),
    updated_at                   timestamptz NOT NULL DEFAULT now()
);

-- ---------------------------------------------------------------------------
-- 18. NOTIFICATIONS
-- ---------------------------------------------------------------------------

CREATE TABLE notifications (
    id              uuid        PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id         uuid        NOT NULL REFERENCES users(id) ON DELETE RESTRICT,
    kind            text        NOT NULL CONSTRAINT notifications_kind_check            CHECK (kind IN ('booking_created', 'booking_confirmed', 'payment_confirmed', 'refund_processed', 'reminder', 'admin_message', 'new_content', 'consultation_booked_admin', 'support_message_admin', 're_engagement', 'failed_job_alert', 'announcement_sent')),
    channel         text        NOT NULL CONSTRAINT notifications_channel_check         CHECK (channel IN ('email', 'in_app')),
    title           text,
    body            text,
    delivery_status text        NOT NULL CONSTRAINT notifications_delivery_status_check CHECK (delivery_status IN ('pending', 'skipped', 'sent', 'failed')),
    payload_json    jsonb,
    scheduled_for   timestamptz,
    read_at         timestamptz,
    sent_at         timestamptz,
    created_at      timestamptz NOT NULL DEFAULT now()
);

CREATE INDEX idx_notifications_user_id_created_at ON notifications(user_id, created_at DESC);
CREATE INDEX idx_notifications_scheduled_for_pending
    ON notifications(scheduled_for) WHERE delivery_status = 'pending' AND scheduled_for IS NOT NULL;
-- Unread count query: SELECT COUNT(*) FROM notifications WHERE user_id = $1 AND read_at IS NULL
CREATE INDEX idx_notifications_user_id_unread
    ON notifications(user_id) WHERE read_at IS NULL;

-- ---------------------------------------------------------------------------
-- 19. ANNOUNCEMENTS
-- ---------------------------------------------------------------------------

CREATE TABLE announcements (
    id                 uuid        PRIMARY KEY DEFAULT gen_random_uuid(),
    title              text        NOT NULL,
    body               text        NOT NULL,
    created_by_user_id uuid        NOT NULL REFERENCES users(id) ON DELETE RESTRICT,
    expires_at         timestamptz,
    created_at         timestamptz NOT NULL DEFAULT now(),
    updated_at         timestamptz NOT NULL DEFAULT now()
);

CREATE INDEX idx_announcements_created_at ON announcements(created_at DESC);
CREATE INDEX idx_announcements_expires_at ON announcements(expires_at) WHERE expires_at IS NOT NULL;

-- ---------------------------------------------------------------------------
-- 20. ACTIVITY LOG
-- ---------------------------------------------------------------------------

CREATE TABLE activity_log (
    id              uuid        PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id         uuid        REFERENCES users(id) ON DELETE RESTRICT,
    event_type      text        NOT NULL,
    resource_type   text        NOT NULL,
    resource_id     uuid        NOT NULL,
    metadata_json   jsonb,
    created_at      timestamptz NOT NULL DEFAULT now()
);

CREATE INDEX idx_activity_log_user_id_created_at        ON activity_log(user_id, created_at DESC);
CREATE INDEX idx_activity_log_resource_type_resource_id ON activity_log(resource_type, resource_id);

-- ---------------------------------------------------------------------------
-- 21. EVENT RELIABILITY
-- ---------------------------------------------------------------------------

CREATE TABLE event_outbox (
    id              uuid        PRIMARY KEY DEFAULT gen_random_uuid(),
    aggregate_type  text        NOT NULL,
    aggregate_id    uuid        NOT NULL,
    event_type      text        NOT NULL,
    payload_json    jsonb       NOT NULL,
    status          text        NOT NULL CONSTRAINT event_outbox_status_check        CHECK (status IN ('pending', 'published', 'failed')),
    attempt_count   integer     NOT NULL DEFAULT 0 CONSTRAINT event_outbox_attempt_count_check CHECK (attempt_count >= 0),
    available_at    timestamptz NOT NULL DEFAULT now(),
    locked_until    timestamptz,
    published_at    timestamptz,
    last_error      text,
    created_at      timestamptz NOT NULL DEFAULT now(),
    updated_at      timestamptz NOT NULL DEFAULT now()
);

CREATE INDEX idx_event_outbox_status_available_at_pending
    ON event_outbox(status, available_at) WHERE status = 'pending';
CREATE INDEX idx_event_outbox_locked_until
    ON event_outbox(locked_until) WHERE locked_until IS NOT NULL;

CREATE TABLE failed_jobs (
    id              uuid        PRIMARY KEY DEFAULT gen_random_uuid(),
    event_type      text        NOT NULL,
    queue_name      text        NOT NULL,
    payload_json    jsonb,
    attempt_count   integer     NOT NULL DEFAULT 0 CONSTRAINT failed_jobs_attempt_count_check CHECK (attempt_count >= 0),
    error_message   text,
    failed_at       timestamptz NOT NULL,
    resolved_at     timestamptz,
    resolution_note text,
    created_at      timestamptz NOT NULL DEFAULT now(),
    updated_at      timestamptz NOT NULL DEFAULT now()
);

CREATE INDEX idx_failed_jobs_failed_at_unresolved
    ON failed_jobs(failed_at DESC) WHERE resolved_at IS NULL;

-- ---------------------------------------------------------------------------
-- 22. ADMIN ACTIONS
-- ---------------------------------------------------------------------------

CREATE TABLE admin_actions (
    id              uuid        PRIMARY KEY DEFAULT gen_random_uuid(),
    admin_user_id   uuid        NOT NULL REFERENCES users(id) ON DELETE RESTRICT,
    action_type     text        NOT NULL CONSTRAINT admin_actions_action_type_check CHECK (action_type IN ('confirm_booking', 'cancel_booking', 'grant_course_access', 'issue_refund', 'record_expense', 'block_user', 'reschedule_booking', 'close_support_chat', 'assign_subadmin', 'revoke_subadmin', 'deactivate_user', 'delete_user', 'create_gift_coupon', 'revoke_gift_coupon', 'publish_article', 'delete_article')),
    target_type     text        NOT NULL CONSTRAINT admin_actions_target_type_check CHECK (target_type IN ('user', 'booking', 'course', 'failed_job', 'payment', 'support_chat', 'review', 'announcement', 'article', 'gift_coupon')),
    target_id       uuid        NOT NULL,
    details_json    jsonb,
    created_at      timestamptz NOT NULL DEFAULT now()
);

CREATE INDEX idx_admin_actions_admin_user_id_created_at ON admin_actions(admin_user_id, created_at DESC);
CREATE INDEX idx_admin_actions_target_type_target_id    ON admin_actions(target_type, target_id);

-- ---------------------------------------------------------------------------
-- 23. SUPPORT CHAT
-- ---------------------------------------------------------------------------

CREATE TABLE support_chats (
    id          uuid        PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id     uuid        NOT NULL REFERENCES users(id) ON DELETE RESTRICT,
    status      text        NOT NULL DEFAULT 'open'
        CONSTRAINT support_chats_status_check CHECK (status IN ('open', 'closed')),
    created_at  timestamptz NOT NULL DEFAULT now(),
    updated_at  timestamptz NOT NULL DEFAULT now()
);

-- One active (open) chat per user at a time; closed chats are retained for history
CREATE UNIQUE INDEX idx_support_chats_user_id_open_unique
    ON support_chats(user_id) WHERE status = 'open';

CREATE INDEX idx_support_chats_status_open ON support_chats(status) WHERE status = 'open';

CREATE TABLE support_messages (
    id          uuid        PRIMARY KEY DEFAULT gen_random_uuid(),
    chat_id     uuid        NOT NULL REFERENCES support_chats(id) ON DELETE RESTRICT,
    sender_id   uuid        NOT NULL REFERENCES users(id) ON DELETE RESTRICT,
    body        text        NOT NULL,
    read_at     timestamptz,
    created_at  timestamptz NOT NULL DEFAULT now()
);

CREATE INDEX idx_support_messages_chat_id_created_at ON support_messages(chat_id, created_at ASC);
CREATE INDEX idx_support_messages_chat_id_unread      ON support_messages(chat_id) WHERE read_at IS NULL;

-- ---------------------------------------------------------------------------
-- 24. P&L VIEWS
-- ---------------------------------------------------------------------------

CREATE VIEW monthly_revenue AS
SELECT
    date_trunc('month', p.created_at)                           AS period,
    p.currency,
    SUM(p.amount)                                               AS gross_revenue_cents,
    COALESCE(SUM(r.total_refunded), 0)                          AS refunded_cents,
    SUM(p.amount) - COALESCE(SUM(r.total_refunded), 0)         AS net_revenue_cents
FROM payments p
LEFT JOIN (
    SELECT payment_id, SUM(amount) AS total_refunded
    FROM refunds
    GROUP BY payment_id
) r ON r.payment_id = p.id
WHERE p.status = 'completed'
GROUP BY 1, 2;

-- FULL OUTER JOIN ensures months with expenses but zero revenue appear as losses (not hidden)
CREATE VIEW monthly_pnl AS
WITH revenue AS (
    SELECT
        date_trunc('month', p.created_at)                       AS period,
        p.currency,
        SUM(p.amount) - COALESCE(SUM(r.total_refunded), 0)     AS net_revenue_cents
    FROM payments p
    LEFT JOIN (
        SELECT payment_id, SUM(amount) AS total_refunded
        FROM refunds
        GROUP BY payment_id
    ) r ON r.payment_id = p.id
    WHERE p.status = 'completed'
    GROUP BY 1, 2
),
expense_totals AS (
    SELECT
        date_trunc('month', occurred_at)    AS period,
        currency,
        SUM(amount)                         AS total_expenses_cents
    FROM expenses
    GROUP BY 1, 2
)
SELECT
    COALESCE(rev.period,   exp.period)   AS period,
    COALESCE(rev.currency, exp.currency) AS currency,
    COALESCE(rev.net_revenue_cents,       0) AS net_revenue_cents,
    COALESCE(exp.total_expenses_cents,    0) AS total_expenses_cents,
    COALESCE(rev.net_revenue_cents, 0) - COALESCE(exp.total_expenses_cents, 0) AS profit_cents
FROM revenue rev
FULL OUTER JOIN expense_totals exp ON exp.period = rev.period AND exp.currency = rev.currency;

-- ---------------------------------------------------------------------------
-- 25. ACCOUNT RECOVERY TOKENS
-- ---------------------------------------------------------------------------

CREATE TABLE account_recovery_tokens (
    id          uuid        PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id     uuid        NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    token_hash  varchar(64)        NOT NULL,
    expires_at  timestamptz NOT NULL,
    used_at     timestamptz,
    created_at  timestamptz NOT NULL DEFAULT now(),
    CONSTRAINT account_recovery_tokens_token_hash_unique     UNIQUE (token_hash),
    CONSTRAINT account_recovery_tokens_token_hash_nonempty   CHECK (length(token_hash) > 0),
    CONSTRAINT account_recovery_tokens_expires_after_created CHECK (expires_at > created_at),
    CONSTRAINT account_recovery_tokens_used_after_created    CHECK (used_at IS NULL OR used_at >= created_at)
);

CREATE INDEX idx_account_recovery_tokens_user_id_unused
    ON account_recovery_tokens(user_id) WHERE used_at IS NULL;

-- ---------------------------------------------------------------------------
-- 26. ARTICLES
-- ---------------------------------------------------------------------------

CREATE TABLE articles (
    id                  uuid        PRIMARY KEY DEFAULT gen_random_uuid(),
    slug                text        NOT NULL,
    title               text        NOT NULL,
    excerpt             text,
    body                text        NOT NULL,
    seo_title           text,
    seo_description     text,
    og_image_url        text,
    is_indexable        boolean     NOT NULL DEFAULT true,
    status              text        NOT NULL DEFAULT 'draft',
    published_at        timestamptz,
    created_by_user_id  uuid        NOT NULL REFERENCES users(id) ON DELETE RESTRICT,
    deleted_at          timestamptz,
    created_at          timestamptz NOT NULL DEFAULT now(),
    updated_at          timestamptz NOT NULL DEFAULT now(),
    CONSTRAINT articles_slug_unique                     UNIQUE (slug),
    CONSTRAINT articles_status_check                    CHECK (status IN ('draft', 'published')),
    CONSTRAINT articles_title_nonempty                  CHECK (btrim(title) <> ''),
    CONSTRAINT articles_slug_nonempty                   CHECK (btrim(slug) <> ''),
    CONSTRAINT articles_body_nonempty                   CHECK (btrim(body) <> ''),
    CONSTRAINT articles_excerpt_nonempty                CHECK (excerpt IS NULL OR btrim(excerpt) <> ''),
    CONSTRAINT articles_published_at_after_created      CHECK (published_at IS NULL OR published_at >= created_at),
    CONSTRAINT articles_deleted_at_after_created        CHECK (deleted_at IS NULL OR deleted_at >= created_at),
    CONSTRAINT articles_published_requires_published_at CHECK (status != 'published' OR published_at IS NOT NULL)
);

CREATE INDEX idx_articles_created_at ON articles(created_at DESC);
CREATE INDEX idx_articles_published_at
    ON articles(published_at DESC) WHERE status = 'published' AND deleted_at IS NULL;

-- ---------------------------------------------------------------------------
-- 27. GIFT COUPONS
-- ---------------------------------------------------------------------------

CREATE TABLE gift_coupons (
    id                  uuid        PRIMARY KEY DEFAULT gen_random_uuid(),
    code                text        NOT NULL,
    course_id           uuid        REFERENCES courses(id) ON DELETE RESTRICT,
    expires_at          timestamptz,
    redeemed_by_user_id uuid        REFERENCES users(id) ON DELETE RESTRICT,
    redeemed_at         timestamptz,
    created_by_user_id  uuid        NOT NULL REFERENCES users(id) ON DELETE RESTRICT,
    created_at          timestamptz NOT NULL DEFAULT now(),
    updated_at          timestamptz NOT NULL DEFAULT now(),
    CONSTRAINT gift_coupons_code_unique               UNIQUE (code),
    CONSTRAINT gift_coupons_code_nonempty             CHECK (btrim(code) <> ''),
    CONSTRAINT gift_coupons_expires_after_created     CHECK (expires_at IS NULL OR expires_at > created_at),
    CONSTRAINT gift_coupons_redeemed_at_after_created CHECK (redeemed_at IS NULL OR redeemed_at >= created_at),
    CONSTRAINT gift_coupons_redemption_consistency    CHECK (
        (redeemed_by_user_id IS NULL AND redeemed_at IS NULL) OR
        (redeemed_by_user_id IS NOT NULL AND redeemed_at IS NOT NULL)
    )
);

CREATE INDEX idx_gift_coupons_unredeemed
    ON gift_coupons(created_at DESC) WHERE redeemed_at IS NULL;
CREATE INDEX idx_gift_coupons_course_id
    ON gift_coupons(course_id) WHERE course_id IS NOT NULL;
