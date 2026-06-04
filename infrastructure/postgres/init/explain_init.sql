-- LearnFlow database initialization — annotated version
-- PostgreSQL 17+; run once on empty volume via /docker-entrypoint-initdb.d/
--
-- GLOBAL DESIGN DECISIONS
-- ───────────────────────
-- UUID PKs          — non-sequential, globally unique, safe to generate in app before INSERT.
--                     No sequential ID enumeration attacks (GET /courses/1, /2, /3 …).
-- text vs varchar   — in PostgreSQL text and varchar are identical on disk; text avoids
--                     arbitrary length decisions and is idiomatic with database/sql.
-- timestamptz       — always stored as UTC, session-displayed in configured timezone.
--                     timestamp (no tz) silently loses location context on DST shifts.
-- bigint for money  — amounts stored as smallest unit (cents/grosze). Avoids IEEE 754
--                     float errors: 0.1 + 0.2 = 0.30000000000000004 in FLOAT.
-- ON DELETE RESTRICT — everywhere. Prevents silent cascading deletes. Explicit deletion
--                     required at the application level. Matches project rule: "never
--                     silently delete production data".
-- CHECK enums       — text + CHECK IN (...) preferred over CREATE TYPE ENUM.
--                     Adding a new value: ALTER TABLE ... ADD CHECK (new_val) vs
--                     ALTER TYPE ... ADD VALUE which cannot be rolled back inside a tx.
-- Soft delete       — deleted_at column instead of physical DELETE. Preserves FK
--                     integrity and full audit trail. Rows excluded by WHERE deleted_at IS NULL.
-- jsonb not json    — binary storage: indexable via GIN, faster read operators,
--                     deduplicates keys. json preserves whitespace/key order (rarely needed).
-- No ORM            — project rule: "raw SQL via database/sql with prepared statements only".
--                     No string concatenation in queries — $1/$2 placeholders only.
-- Named constraints — all CHECK and UNIQUE constraints have explicit names. Unnamed
--                     constraints get auto-generated names (e.g. users_role_check1) that
--                     are unpredictable and fragile in migrations (ALTER TABLE ... DROP
--                     CONSTRAINT requires knowing the exact name).

-- [pg_stat_statements]: Tracks execution statistics for every SQL statement.
-- Enables queries like: "which 10 queries consume the most total time?".
-- Required for query performance analysis in development and production.
-- Pairs with postgres_exporter → Prometheus → Grafana dashboard.
CREATE EXTENSION IF NOT EXISTS pg_stat_statements;

-- ---------------------------------------------------------------------------
-- 1. USERS
-- ---------------------------------------------------------------------------

CREATE TABLE users (
    -- [gen_random_uuid()]: PostgreSQL 13+ built-in (no pgcrypto needed). Generates
    -- a cryptographically random UUID v4. Non-sequential → no ID enumeration.
    id                  uuid        PRIMARY KEY DEFAULT gen_random_uuid(),

    email               varchar(254)    NOT NULL,

    -- [password_hash]: CLAUDE.md rule: "Passwords — bcrypt only." Raw passwords are
    -- never stored. The app hashes with bcrypt before INSERT/UPDATE.
    password_hash   varchar(60)        NOT NULL,

    -- [role CHECK]: Only 'user', 'subadmin', and 'admin' exist in this platform. DB-level
    -- CHECK prevents a bug in the auth service from inserting an invalid role that would
    -- bypass the permission system silently.
    role                text        NOT NULL CONSTRAINT users_role_check   CHECK (role IN ('user', 'subadmin', 'admin')),

    -- [status vs deleted_at]: status = business state (active/blocked/pending_verification).
    -- deleted_at = logical deletion. A blocked user still exists and can log in
    -- to see an "account blocked" message. A deleted user is invisible to queries.
    -- 'pending_verification' — registered but email not yet confirmed; access restricted
    -- until the user completes the email_verification_tokens flow.
    -- ['deleted' status coexists with deleted_at]: status='deleted' triggers immediate
    -- access revocation (checked on every request). deleted_at enables timeline queries
    -- and partial index compatibility (WHERE deleted_at IS NULL). Both are set together.
    status              text        NOT NULL CONSTRAINT users_status_check CHECK (status IN ('active', 'blocked', 'pending_verification', 'deleted')),

    -- [nullable email_verified_at]: NULL = not verified; non-NULL = verified at
    -- that exact timestamp. More informative than a boolean is_verified flag:
    -- gives the verification time for support/audit without an extra column.
    email_verified_at   timestamptz,
    last_login_at       timestamptz,

    -- [deleted_at timestamp over is_deleted boolean]: You get both the deletion
    -- flag and the deletion time in one column. Queries: WHERE deleted_at IS NULL.
    deleted_at          timestamptz,

    created_at          timestamptz NOT NULL DEFAULT now(),
    updated_at          timestamptz NOT NULL DEFAULT now(),

    -- [password_changed_at / email_changed_at]: Track the last credential change.
    -- Used to invalidate sessions issued before the change — any session with
    -- created_at < password_changed_at is considered stale and should be revoked.
    -- NULL = credential never changed since account creation.
    password_changed_at    timestamptz,
    email_changed_at       timestamptz,

    -- [failed_login_count + login_locked_until]: Account-level brute-force protection.
    -- Covers all login vectors (web, mobile, API). Application MUST reset
    -- failed_login_count = 0 atomically on successful login in the same UPDATE as last_login_at.
    -- Complements user_sessions.failed_attempt_count which is per-device/session.
    -- login_locked_until uses timestamptz — time-zone-aware DST-safe lock expiry.
    failed_login_count     integer     NOT NULL DEFAULT 0,
    last_failed_login_at   timestamptz,
    login_locked_until     timestamptz,

    CONSTRAINT users_email_nonempty                        CHECK (btrim(email) <> ''),
    CONSTRAINT users_email_verified_at_after_created       CHECK (email_verified_at IS NULL OR email_verified_at >= created_at),
    CONSTRAINT users_last_login_at_after_created           CHECK (last_login_at IS NULL OR last_login_at >= created_at),
    CONSTRAINT users_deleted_at_after_created              CHECK (deleted_at IS NULL OR deleted_at >= created_at),
    CONSTRAINT users_failed_login_count_non_negative       CHECK (failed_login_count >= 0)
);

-- [LOWER(email) partial unique]: Two protections in one index:
-- 1. LOWER() — emails are case-insensitive by RFC 5321. Without normalization,
--    'User@Example.com' and 'user@example.com' would coexist as separate accounts.
--    The app must also LOWER() emails before querying to hit this index.
-- 2. WHERE deleted_at IS NULL — allows the same email to be reused after soft-delete
--    (e.g., user deletes account, re-registers later). A full unique would block this.
CREATE UNIQUE INDEX idx_users_email_active_unique ON users(LOWER(email)) WHERE deleted_at IS NULL;

-- ---------------------------------------------------------------------------
-- 2. USER PROFILES
-- ---------------------------------------------------------------------------

-- [Separate table]: Profile data (bio, avatar, timezone) is only needed on profile/
-- settings pages. Keeping it separate avoids loading many nullable columns on every
-- auth check, permission guard, or JOIN that only needs id + email + role.
-- 1:1 relation enforced by PRIMARY KEY = FK to users.
CREATE TABLE user_profiles (
    user_id         uuid        PRIMARY KEY REFERENCES users(id) ON DELETE RESTRICT,

    -- [Split full_name into first_name + last_name]: A single full_name column is
    -- convenient to store but painful to consume. The notification service needs
    -- "Dear Vitaliy" (first name only). The admin panel sorts by last name.
    -- Splitting at the source avoids unreliable string parsing later.
    -- Both are nullable — a user may fill in only what they want.
    first_name      text,
    last_name       text,

    -- [phone_number as text]: Phone numbers include country codes, leading zeros,
    -- and formatting characters (+48 500 123 456). Storing as text avoids integer
    -- truncation and normalization headaches. Validation belongs in the app layer.
    phone_number    varchar(20),
    country         text            CONSTRAINT user_profiles_country_check CHECK (country IS NULL OR char_length(country) = 2),
    city            text,

    -- [date type for date_of_birth]: Pure calendar date — no time component needed.
    -- Using timestamptz would add timezone ambiguity (midnight in which zone?).
    -- The 'date' type is exactly the right semantic here.
    date_of_birth   date,

    -- [gender CHECK with 'prefer_not_to_say']: Provides typed values for analytics
    -- and localization while including the privacy-respecting option. NULL = not set
    -- (user hasn't visited the field). 'prefer_not_to_say' = explicit opt-out.
    -- The constraint is named for clean DROP CONSTRAINT in future migrations.
    gender          text        CONSTRAINT user_profiles_gender_check CHECK (gender IN ('male', 'female', 'other', 'prefer_not_to_say')),

    -- [ui_language NOT NULL DEFAULT 'uk']: Controls the Nuxt 3 SSR locale for this
    -- user. NOT NULL with DEFAULT avoids null-checks in every i18n lookup. Default
    -- 'uk' matches the platform's primary audience (Ukrainian speakers).
    ui_language     text        NOT NULL DEFAULT 'uk',

    avatar_url      text,

    -- [timezone as text]: Stored as IANA string (e.g. 'Europe/Warsaw', 'UTC').
    -- No CHECK — the full IANA list (600+ zones) belongs in the app, not a DB constraint.
    timezone        text,
    bio             text,
    created_at      timestamptz NOT NULL DEFAULT now(),
    updated_at      timestamptz NOT NULL DEFAULT now()
);

-- ---------------------------------------------------------------------------
-- 3. AUTH TOKENS
-- ---------------------------------------------------------------------------

-- [token_hash, not raw token]: If the tokens table is ever leaked or accessed by a
-- compromised read replica, attackers cannot use hashed values directly. The flow:
--   1. App generates a random token (crypto/rand, 32 bytes)
--   2. Sends raw token to user (email link)
--   3. Stores SHA-256(token) in the DB
--   4. On verification: hash the submitted token, look up by hash
-- CLAUDE.md: "JWT secret from env only" — same security-first philosophy.
--
-- [UNIQUE on token_hash]: Enables O(log n) lookup: WHERE token_hash = $1.
-- Without UNIQUE, every verification request is a full sequential scan.
-- Also prevents hash collision persistence (astronomically rare but defensive).
--
-- [Separate tables per token type]: email_verification vs password_reset vs
-- email_change have different expiry policies, different invalidation logic,
-- and different audit needs. A shared 'tokens' table with a 'type' discriminator
-- creates coupling and complicates the partial indexes (every index would need
-- WHERE type = '...').
CREATE TABLE email_verification_tokens (
    id          uuid               PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id     uuid               NOT NULL REFERENCES users(id) ON DELETE RESTRICT,
    token_hash  varchar(64)        NOT NULL CONSTRAINT email_verification_tokens_token_hash_nonempty CHECK (length(token_hash) > 0),
    expires_at  timestamptz        NOT NULL,

    -- [used_at timestamp over is_used boolean]: Records when the token was consumed.
    -- Useful for support: "when did the user verify their email?".
    -- NULL = unused; non-NULL = used at that time.
    used_at                 timestamptz,
    -- [invalidated_at + invalidated_by_user_id]: Explicit admin/system invalidation
    -- distinct from natural expiry. Who invalidated this token and when?
    -- NULL invalidated_by_user_id = natural expiry; non-NULL = explicit action (admin/system).
    invalidated_at          timestamptz,
    invalidated_by_user_id  uuid        REFERENCES users(id) ON DELETE RESTRICT,
    created_at              timestamptz NOT NULL DEFAULT now(),
    CONSTRAINT email_verification_tokens_token_hash_unique    UNIQUE (token_hash),
    CONSTRAINT email_verification_tokens_expires_after_created CHECK (expires_at > created_at),
    CONSTRAINT email_verification_tokens_used_after_created   CHECK (used_at IS NULL OR used_at >= created_at)
);

-- [Partial index WHERE used_at IS NULL]: The only query by user_id is
-- "find this user's pending tokens" (to invalidate old ones before issuing a new one).
-- Used tokens are never queried by user_id again. Partial index = smaller, faster.
CREATE INDEX idx_email_verification_tokens_user_id_unused
    ON email_verification_tokens(user_id) WHERE used_at IS NULL;

CREATE TABLE password_reset_tokens (
    id          uuid               PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id     uuid               NOT NULL REFERENCES users(id) ON DELETE RESTRICT,
    token_hash  varchar(64)        NOT NULL CONSTRAINT password_reset_tokens_token_hash_nonempty CHECK (length(token_hash) > 0),
    expires_at  timestamptz        NOT NULL,
    used_at                 timestamptz,
    invalidated_at          timestamptz,
    invalidated_by_user_id  uuid        REFERENCES users(id) ON DELETE RESTRICT,
    created_at              timestamptz NOT NULL DEFAULT now(),
    CONSTRAINT password_reset_tokens_token_hash_unique    UNIQUE (token_hash),
    CONSTRAINT password_reset_tokens_expires_after_created CHECK (expires_at > created_at),
    CONSTRAINT password_reset_tokens_used_after_created   CHECK (used_at IS NULL OR used_at >= created_at)
);

CREATE INDEX idx_password_reset_tokens_user_id_unused
    ON password_reset_tokens(user_id) WHERE used_at IS NULL;

-- [email_change_tokens — dedicated table for email change flow]:
-- Email change is security-sensitive: the user must confirm both old and new email.
-- Storing new_email in the token row (not in users) avoids a race condition:
-- if the user starts two change requests simultaneously, only the one with the
-- matching token_hash succeeds. The users.email column stays unchanged until
-- the token is consumed and validated.
CREATE TABLE email_change_tokens (
    id          uuid        PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id     uuid        NOT NULL REFERENCES users(id) ON DELETE RESTRICT,

    -- [new_email stored in token row]: The pending new address lives here, not in
    -- users. This avoids partial state: users.email is authoritative and only
    -- updated atomically when the token is consumed.
    new_email   varchar(254)    NOT NULL,
    token_hash  varchar(64)     NOT NULL CONSTRAINT email_change_tokens_token_hash_nonempty CHECK (length(token_hash) > 0),
    expires_at  timestamptz        NOT NULL,
    used_at                 timestamptz,
    invalidated_at          timestamptz,
    invalidated_by_user_id  uuid        REFERENCES users(id) ON DELETE RESTRICT,
    created_at              timestamptz NOT NULL DEFAULT now(),
    CONSTRAINT email_change_tokens_token_hash_unique    UNIQUE (token_hash),
    CONSTRAINT email_change_tokens_expires_after_created CHECK (expires_at > created_at),
    CONSTRAINT email_change_tokens_used_after_created   CHECK (used_at IS NULL OR used_at >= created_at)
);

-- [Partial index WHERE used_at IS NULL]: Same rationale as other auth token tables —
-- only pending (unused) tokens are ever looked up by user_id.
CREATE INDEX idx_email_change_tokens_user_id_unused
    ON email_change_tokens(user_id) WHERE used_at IS NULL;

-- ---------------------------------------------------------------------------
-- 4. COURSES
-- ---------------------------------------------------------------------------

-- CLAUDE.md domain: "courses/ — Courses, landing pages, SEO metadata"
-- The SEO columns (seo_title, og_image_url, canonical_url, is_indexable) support
-- the landing page module. All are nullable — not every course needs custom SEO.
CREATE TABLE courses (
    id                  uuid        PRIMARY KEY DEFAULT gen_random_uuid(),

    -- [slug table-level UNIQUE constraint]: Course URLs are /courses/{slug}.
    -- Named constraint enables clean ALTER TABLE ... DROP CONSTRAINT in migrations.
    -- The UNIQUE constraint creates the B-tree index needed for O(log n) slug lookup.
    slug                text        NOT NULL,
    title               text        NOT NULL,
    description         text,
    thumbnail_url       text,
    preview_video_url   text,

    -- [status CHECK, not ENUM type]: Adding 'suspended' later:
    --   ENUM: ALTER TYPE course_status ADD VALUE 'suspended' — cannot be rolled back
    --         inside a transaction if other migrations follow.
    --   text+CHECK: ALTER TABLE courses ADD CONSTRAINT ... CHECK (..., 'suspended')
    --         — fully transactional, rollback-safe.
    status              text        NOT NULL CONSTRAINT courses_status_check CHECK (status IN ('draft', 'published', 'archived')),

    -- [CHECK estimated_minutes > 0]: A course advertised as "0 minutes" or negative is a
    -- data entry error. NULL is allowed — not all courses have a duration estimate.
    estimated_minutes   integer     CONSTRAINT courses_estimated_minutes_positive CHECK (estimated_minutes IS NULL OR estimated_minutes > 0),
    seo_title           text,
    seo_description     text,
    og_image_url        text,
    canonical_url       text,
    is_indexable        boolean     NOT NULL DEFAULT true,
    created_by_user_id  uuid        NOT NULL REFERENCES users(id) ON DELETE RESTRICT,
    created_at          timestamptz NOT NULL DEFAULT now(),
    updated_at          timestamptz NOT NULL DEFAULT now(),

    -- [nullable published_at]: Set when status transitions to 'published'.
    -- Used in SEO (sitemap lastmod), analytics (time-to-publish), and feeds.
    published_at        timestamptz,

    -- [deleted_at on courses]: Soft delete mirrors the users pattern. Allows a course
    -- to be "removed" without breaking FK references from payments, access records,
    -- and progress rows that depend on it. WHERE deleted_at IS NULL in all listing queries.
    deleted_at          timestamptz,
    CONSTRAINT courses_slug_unique UNIQUE (slug)
);

-- [status index]: Most common listing query: WHERE status = 'published'.
-- Without this, every course listing is a full table scan.
-- [Partial WHERE deleted_at IS NULL]: Soft-delete tables have almost all rows non-deleted.
-- Partial index covers only non-deleted courses — much smaller and faster than a full index.
-- Migration 000004 dropped the original full index and recreated it as partial.
CREATE INDEX idx_courses_status ON courses(status) WHERE deleted_at IS NULL;

-- ---------------------------------------------------------------------------
-- 5. CONTENT ITEMS
-- ---------------------------------------------------------------------------

-- CLAUDE.md domain: "content/ — Learning content items (video, book, presentation)"
-- content_type is the discriminator for which fields are meaningful:
--   video       → video_url, estimated_minutes, (engagement tracked in user_video_engagement)
--   book        → file_url, estimated_pages, (engagement tracked in user_document_engagement)
--   presentation → file_url, estimated_pages
--
-- [SEO fields — same set as courses]: Content items have their own landing pages
-- (/content/{slug}), so they need the same SEO surface as courses:
--   thumbnail_url   — cover image shown in listings and link previews
--   seo_title       — custom <title> tag; falls back to title if NULL
--   seo_description — <meta name="description">
--   og_image_url    — Open Graph image for social sharing
--   canonical_url   — explicit canonical to prevent duplicate-content penalties
--                     (e.g., same content accessible via /courses/{slug}/content/{slug})
--   is_indexable    — false for draft/preview pages; robots noindex
CREATE TABLE content_items (
    id                  uuid        PRIMARY KEY DEFAULT gen_random_uuid(),
    slug                text        NOT NULL,
    title               text        NOT NULL,
    content_type        text        NOT NULL CONSTRAINT content_items_content_type_check CHECK (content_type IN ('video', 'book', 'presentation')),

    -- [description, not summary]: Name follows the platform-wide convention used on courses
    -- and articles — 'description' is the short editorial intro; 'body' is the full content.
    description         text,

    -- [body text]: Markdown/rich text stored directly. Acceptable at this scale.
    -- For very large content, migrate to object storage (Cloudflare R2 in prod,
    -- MinIO locally — see infrastructure/storage in CLAUDE.md).
    body                text,
    video_url           text,
    file_url            text,
    -- [CHECK > 0 on estimates]: A video with 0 or negative estimated minutes is a bug.
    -- NULL is allowed — not all items have a duration/page estimate yet.
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

    -- [deleted_at on content_items]: Same rationale as courses. Content items may be
    -- referenced by user_content_progress, user_video_engagement, course_content_items,
    -- and user_content_access. Physical delete would violate all these FKs.
    deleted_at          timestamptz,
    CONSTRAINT content_items_slug_unique UNIQUE (slug)
);

-- [Composite (content_type, status)]: Covers two query patterns with one index:
--   1. "Show all published videos"          → content_type='video' AND status='published'
--   2. "Show all items of a type"           → content_type='book' (leftmost prefix)
-- ORDER matters: content_type first because it has lower cardinality (3 values vs 3),
-- making it a better filter for the first step of index scan.
-- Partial WHERE deleted_at IS NULL: same rationale as idx_courses_status.
CREATE INDEX idx_content_items_content_type_status ON content_items(content_type, status) WHERE deleted_at IS NULL;

-- ---------------------------------------------------------------------------
-- 6. COURSE ↔ CONTENT ITEMS (junction)
-- ---------------------------------------------------------------------------

CREATE TABLE course_content_items (
    id              uuid        PRIMARY KEY DEFAULT gen_random_uuid(),
    course_id       uuid        NOT NULL REFERENCES courses(id) ON DELETE RESTRICT,
    content_item_id uuid        NOT NULL REFERENCES content_items(id) ON DELETE RESTRICT,

    -- [position integer]: Explicit sort order. Allows reordering without touching
    -- sibling rows (gaps are fine: 10, 20, 30 → insert at 15 without renumbering).
    -- [CHECK > 0]: Position 0 would be ambiguous (is it "first" or "unset"?).
    -- Positions start at 1. Enforced here so no content ordering bug can persist.
    position        integer     NOT NULL CONSTRAINT course_content_items_position_positive CHECK (position > 0),
    is_required     boolean     NOT NULL,
    created_at      timestamptz NOT NULL DEFAULT now(),
    CONSTRAINT course_content_items_unique_pair     UNIQUE (course_id, content_item_id),

    -- [DEFERRABLE INITIALLY DEFERRED on position]: Without DEFERRABLE, swapping
    -- positions 1 and 2 in a single transaction fails:
    --   UPDATE SET position=2 WHERE position=1  ← violates unique (2 already exists)
    --   UPDATE SET position=1 WHERE position=2  ← never reached
    -- With DEFERRABLE, the uniqueness check runs at COMMIT, not per-statement,
    -- so intermediate states are allowed. The transaction either fully succeeds or
    -- fully rolls back.
    CONSTRAINT course_content_items_unique_position UNIQUE (course_id, position) DEFERRABLE INITIALLY DEFERRED
);

-- [content_item_id index]: UNIQUE (course_id, content_item_id) covers forward lookup
-- (items in a course) but NOT the reverse: "which courses contain this item?".
-- Needed for: content impact analysis, content deletion safety checks.
-- PostgreSQL does NOT auto-create FK indexes — explicit index required.
CREATE INDEX idx_course_content_items_content_item_id ON course_content_items(content_item_id);

-- ---------------------------------------------------------------------------
-- 7. COURSE REVIEWS
-- ---------------------------------------------------------------------------

-- [course_reviews — separate table from courses]: Ratings are an aggregate of user
-- submissions, not a property of the course itself. Storing avg_rating on courses
-- would require update triggers or app-level synchronization. Instead, the courses
-- service computes AVG(rating) via query when needed, and can cache it as needed.
-- Separating also gives full history: who rated what and when.
--
-- [deleted_at on reviews]: Soft delete allows admins to hide inappropriate or
-- fraudulent reviews without physical deletion. Preserves FK integrity and audit trail.
-- WHERE deleted_at IS NULL in all public-facing review queries.
CREATE TABLE course_reviews (
    id          uuid        PRIMARY KEY DEFAULT gen_random_uuid(),
    course_id   uuid        NOT NULL REFERENCES courses(id) ON DELETE RESTRICT,
    user_id     uuid        NOT NULL REFERENCES users(id) ON DELETE RESTRICT,

    -- [CHECK BETWEEN 1 AND 5]: Standard 5-star scale. Enforced at DB level so
    -- a bug in the frontend (submitting rating=0 or rating=6) cannot corrupt
    -- the computed average. BETWEEN is inclusive on both ends.
    rating      integer     NOT NULL CONSTRAINT course_reviews_rating_check CHECK (rating BETWEEN 1 AND 5),
    comment     text,
    created_at  timestamptz NOT NULL DEFAULT now(),
    updated_at  timestamptz NOT NULL DEFAULT now(),
    deleted_at  timestamptz
);

-- [Partial unique WHERE deleted_at IS NULL]: One active review per user per course.
-- A hard UNIQUE (user_id, course_id) would block re-reviews after soft-delete:
--   user deletes review (deleted_at set) → tries to re-submit → INSERT fails because
--   the deleted row still occupies the unique slot.
-- Partial unique excludes deleted rows, allowing a new review after soft-delete.
CREATE UNIQUE INDEX idx_course_reviews_user_id_course_id_active_unique ON course_reviews(user_id, course_id) WHERE deleted_at IS NULL;

-- [Partial index WHERE deleted_at IS NULL]: Public queries only need active reviews.
-- Soft-deleted reviews are excluded from rating aggregations and display.
CREATE INDEX idx_course_reviews_course_id_active ON course_reviews(course_id) WHERE deleted_at IS NULL;

-- ---------------------------------------------------------------------------
-- 8. CONTENT REVIEWS
-- ---------------------------------------------------------------------------

-- [content_reviews — mirrors course_reviews for individual content items]:
-- Content items (videos, books, presentations) can be purchased and consumed
-- independently of a course. User feedback on individual items is valuable for:
--   1. Content quality signals — which videos/books users find most useful
--   2. Purchase decisions — buyers see ratings before buying standalone content
--   3. Creator feedback — content authors see per-item ratings separately from course ratings
-- Structure intentionally mirrors course_reviews for consistency.
CREATE TABLE content_reviews (
    id              uuid        PRIMARY KEY DEFAULT gen_random_uuid(),
    content_item_id uuid        NOT NULL REFERENCES content_items(id) ON DELETE RESTRICT,
    user_id         uuid        NOT NULL REFERENCES users(id) ON DELETE RESTRICT,
    rating          integer     NOT NULL CONSTRAINT content_reviews_rating_check CHECK (rating BETWEEN 1 AND 5),
    comment         text,
    created_at      timestamptz NOT NULL DEFAULT now(),
    updated_at      timestamptz NOT NULL DEFAULT now(),
    deleted_at      timestamptz
);

-- [Partial unique WHERE deleted_at IS NULL]: Same rationale as course_reviews —
-- allows re-review after soft-delete without hard unique blocking the INSERT.
CREATE UNIQUE INDEX idx_content_reviews_user_id_content_item_id_active_unique ON content_reviews(user_id, content_item_id) WHERE deleted_at IS NULL;

-- [Partial index WHERE deleted_at IS NULL]: Same rationale as idx_course_reviews_course_id_active.
CREATE INDEX idx_content_reviews_content_item_id_active ON content_reviews(content_item_id) WHERE deleted_at IS NULL;

-- ---------------------------------------------------------------------------
-- 9. PAYMENTS
-- (created before user_course_access, user_content_access, and consultation_bookings — all ref it)
-- ---------------------------------------------------------------------------

-- CLAUDE.md domain: "payments/ — Payments, Stripe, Przelewy24, refunds"
CREATE TABLE payments (
    id                  uuid        PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id             uuid        NOT NULL REFERENCES users(id) ON DELETE RESTRICT,

    -- [bigint for amount]: Stores smallest currency unit (1 PLN = 100 units).
    -- IEEE 754 float: 0.1 + 0.2 = 0.30000000000000004. A PLN 1999.99 payment
    -- stored as FLOAT would drift after arithmetic. bigint = exact always.
    -- [CHECK amount > 0]: A zero or negative payment is a bug. Mirrors the same
    -- constraint on refunds and expenses — all financial amounts are positive.
    amount              bigint      NOT NULL CONSTRAINT payments_amount_positive CHECK (amount > 0),

    -- [currency column]: Platform supports PLN (Przelewy24) and potentially USD/EUR
    -- (Stripe). Required for correct P&L grouping — see monthly_revenue view.
    -- Without this, SUM(amount) across currencies produces nonsense.
    -- [DEFAULT 'PLN']: Primary market is Poland. Avoids requiring the app to always
    -- pass currency explicitly for the common case. EUR/USD flows set it explicitly.
    currency            text        NOT NULL DEFAULT 'PLN',

    status              text        NOT NULL CONSTRAINT payments_status_check         CHECK (status IN ('pending', 'completed', 'failed')),

    -- ['manual' in payment_method]: Admins can record offline bank transfers and
    -- manual payments. Matches admin_actions 'record_expense' pattern.
    payment_method      text        NOT NULL CONSTRAINT payments_payment_method_check CHECK (payment_method IN ('card', 'bank_transfer', 'manual')),

    -- [nullable provider_reference]: Set by Stripe/P24 webhook after confirmation.
    -- NULL during 'pending'. Webhook handlers look up by this reference.
    provider_reference  text        CONSTRAINT payments_provider_reference_check CHECK (provider_reference IS NULL OR length(btrim(provider_reference)) > 0),

    -- [UTM fields denormalized here]: Captured at payment moment for conversion
    -- attribution. Denormalizing onto the payment avoids a JOIN to a sessions table
    -- on every P&L/attribution query. Trade-off: slight duplication vs query simplicity.
    utm_source          text,
    utm_medium          text,
    utm_campaign        text,
    utm_content         text,
    utm_term            text,

    -- [gclid separate from UTMs]: Google Click ID is a separate Google Ads attribution
    -- identifier, not part of the standard UTM parameter set.
    gclid               text,
    created_at          timestamptz NOT NULL DEFAULT now(),
    updated_at          timestamptz NOT NULL DEFAULT now()
);

CREATE INDEX idx_payments_user_id_created_at ON payments(user_id, created_at DESC);
CREATE INDEX idx_payments_status_created_at ON payments(status, created_at);
CREATE INDEX idx_payments_utm_campaign ON payments(utm_campaign) WHERE utm_campaign IS NOT NULL;

-- [provider_reference partial index]: Stripe/P24 webhook handlers look up:
--   SELECT * FROM payments WHERE provider_reference = $1
-- Partial (WHERE NOT NULL) excludes all manual payments — keeps the index small
-- and only covers rows that webhooks will ever query.
CREATE INDEX idx_payments_provider_reference ON payments(provider_reference) WHERE provider_reference IS NOT NULL;

-- [Polymorphic resource_id, no FK]: A line item can be a course or a consultation.
-- PostgreSQL cannot enforce a FK to multiple tables simultaneously. Application
-- logic in the payments service validates resource existence before INSERT.
-- resource_type acts as a discriminator for which table to join.
CREATE TABLE payment_line_items (
    id              uuid        PRIMARY KEY DEFAULT gen_random_uuid(),
    payment_id      uuid        NOT NULL REFERENCES payments(id) ON DELETE RESTRICT,
    -- ['content_item' included]: Users can purchase individual content items (PYMT-07).
    -- Three resource types map to three separate access tables:
    --   course        → user_course_access
    --   consultation  → consultation_bookings
    --   content_item  → user_content_access
    resource_type   text        NOT NULL CONSTRAINT payment_line_items_resource_type_check CHECK (resource_type IN ('course', 'consultation', 'content_item')),
    resource_id     uuid        NOT NULL,

    -- [CHECK amount > 0]: A line item amount of zero or negative is a billing bug.
    -- Mirrors the same constraint on payments.amount and refunds.amount.
    amount          bigint      NOT NULL CONSTRAINT payment_line_items_amount_positive CHECK (amount > 0),
    description     text,
    created_at      timestamptz NOT NULL DEFAULT now()
);

-- [FK index — critical]: PostgreSQL does NOT auto-create indexes for FK columns.
-- Every query "get line items for payment X" would be a full seq scan without this.
-- Also required for the JOIN in revenue reports.
CREATE INDEX idx_payment_line_items_payment_id ON payment_line_items(payment_id);

-- [resource_type + resource_id index]: "Which payment covers this resource?"
-- (e.g. find payment for a given course purchase). Polymorphic FK lookup.
CREATE INDEX idx_payment_line_items_resource_type_resource_id ON payment_line_items(resource_type, resource_id);

CREATE TABLE refunds (
    id                  uuid        PRIMARY KEY DEFAULT gen_random_uuid(),
    payment_id          uuid        NOT NULL REFERENCES payments(id) ON DELETE RESTRICT,

    -- [CHECK amount > 0]: A zero or negative refund is a bug. Catching it at the
    -- DB level prevents incorrect financial records regardless of what the refund
    -- service calculates.
    amount              bigint      NOT NULL CONSTRAINT refunds_amount_positive CHECK (amount > 0),
    reason              text,
    created_by_user_id  uuid        NOT NULL REFERENCES users(id) ON DELETE RESTRICT,
    refunded_at         timestamptz NOT NULL,
    created_at          timestamptz NOT NULL DEFAULT now()
);

-- [FK index]: monthly_revenue view JOINs refunds → payments per payment_id.
-- Without this index, the P&L view would scan all refunds for each payment.
CREATE INDEX idx_refunds_payment_id ON refunds(payment_id);

-- ---------------------------------------------------------------------------
-- 10. USER COURSE ACCESS & PROGRESS
-- ---------------------------------------------------------------------------

CREATE TABLE user_course_access (
    id          uuid        PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id     uuid        NOT NULL REFERENCES users(id) ON DELETE RESTRICT,
    course_id   uuid        NOT NULL REFERENCES courses(id) ON DELETE RESTRICT,

    -- [nullable payment_id]: Admin-granted and gift-coupon access have no payment.
    -- access_type = 'admin_granted' → payment_id NULL.
    -- access_type = 'gift_coupon'   → payment_id NULL.
    -- access_type = 'purchased'     → payment_id NOT NULL (enforced at app level).
    payment_id  uuid        REFERENCES payments(id) ON DELETE RESTRICT,

    -- ['gift_coupon' access_type]: Users can redeem a gift coupon (gift_coupons table)
    -- to gain course access without a payment. Treated the same as 'admin_granted'
    -- for access control but tracked separately for reporting and coupon audit.
    -- 'granted' (generic free tier) removed — not in scope for this platform.
    access_type text        NOT NULL CONSTRAINT user_course_access_access_type_check CHECK (access_type IN ('purchased', 'admin_granted', 'gift_coupon')),
    status      text        NOT NULL CONSTRAINT user_course_access_status_check      CHECK (status IN ('active', 'revoked', 'refunded')),
    granted_at  timestamptz NOT NULL,
    expires_at  timestamptz,
    created_at  timestamptz NOT NULL DEFAULT now(),
    updated_at  timestamptz NOT NULL DEFAULT now(),
    CONSTRAINT user_course_access_granted_at_after_created CHECK (granted_at >= created_at),
    CONSTRAINT user_course_access_expires_after_granted   CHECK (expires_at IS NULL OR expires_at > granted_at)
);

-- [Partial unique, NOT table-level UNIQUE]: A table-level UNIQUE (user_id, course_id)
-- would prevent re-purchase after refund:
--   1. User buys course  → row inserted (status='active')
--   2. User gets refund  → row updated (status='refunded')
--   3. User buys again   → INSERT fails! (user_id, course_id) pair already exists
-- Partial unique WHERE status='active' enforces "only one active access per course"
-- while allowing multiple historical rows (refunded, revoked).
CREATE UNIQUE INDEX idx_user_course_access_user_id_course_id_active_unique
    ON user_course_access(user_id, course_id) WHERE status = 'active';

CREATE INDEX idx_user_course_access_course_id_active
    ON user_course_access(course_id) WHERE status = 'active';

CREATE TABLE user_course_progress (
    id                  uuid        PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id             uuid        NOT NULL REFERENCES users(id) ON DELETE RESTRICT,
    course_id           uuid        NOT NULL REFERENCES courses(id) ON DELETE RESTRICT,
    status              text        NOT NULL CONSTRAINT user_course_progress_status_check  CHECK (status IN ('not_started', 'in_progress', 'completed')),

    -- [CHECK 0-100]: Catches calculation bugs in the progress service before they
    -- persist. A progress_percent of 101 or -5 indicates a logic error.
    progress_percent    integer     NOT NULL DEFAULT 0 CONSTRAINT user_course_progress_percent_check CHECK (progress_percent BETWEEN 0 AND 100),
    started_at          timestamptz,
    last_interacted_at  timestamptz,
    completed_at        timestamptz,
    created_at          timestamptz NOT NULL DEFAULT now(),
    updated_at          timestamptz NOT NULL DEFAULT now(),
    -- [UNIQUE (user_id, course_id)]: One progress row per user per course.
    -- UPSERT pattern: INSERT ... ON CONFLICT (user_id, course_id) DO UPDATE SET ...
    -- The UNIQUE constraint also creates the index needed for those UPSERTs.
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

-- [Separate video/document engagement tables]: Video and document engagement have
-- fundamentally different metrics:
--   video    → seconds watched, playhead position (watched_seconds, last_position_seconds)
--   document → pages viewed, last page (viewed_pages, last_page)
-- A single engagement table with nullable columns for each type would be sparse,
-- ambiguous (is NULL "not applicable" or "not recorded"?), and harder to index.
CREATE TABLE user_video_engagement (
    id                      uuid        PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id                 uuid        NOT NULL REFERENCES users(id) ON DELETE RESTRICT,
    content_item_id         uuid        NOT NULL REFERENCES content_items(id) ON DELETE RESTRICT,

    -- [CHECK >= 0]: Watched seconds cannot be negative. Defensive against clock
    -- drift bugs or subtraction errors in the frontend tracking code.
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

-- [user_content_access vs user_course_access]: A user can purchase a single
-- content item (video, book) without buying the full course. This is a separate
-- access model — course access grants access to all course_content_items,
-- while content access is per-item. The content/ module checks both tables
-- before returning a content item.
--
-- [access_type]: 'purchased' (via payment), 'admin_granted' (manual override),
-- 'gift_coupon' (coupon redemption). No generic 'granted' free tier for content items.
-- Keeping the set tight prevents unintended access type creep.
CREATE TABLE user_content_access (
    id              uuid        PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id         uuid        NOT NULL REFERENCES users(id) ON DELETE RESTRICT,
    content_item_id uuid        NOT NULL REFERENCES content_items(id) ON DELETE RESTRICT,

    -- [nullable payment_id]: admin_granted and gift_coupon access have no payment.
    payment_id      uuid        REFERENCES payments(id) ON DELETE RESTRICT,
    access_type     text        NOT NULL CONSTRAINT user_content_access_access_type_check CHECK (access_type IN ('purchased', 'admin_granted', 'gift_coupon')),
    status          text        NOT NULL CONSTRAINT user_content_access_status_check CHECK (status IN ('active', 'revoked', 'refunded')),
    granted_at      timestamptz NOT NULL,

    -- [expires_at nullable]: Content access can be time-limited (e.g., promotional coupon).
    -- NULL = permanent.
    expires_at      timestamptz,
    created_at      timestamptz NOT NULL DEFAULT now(),
    updated_at      timestamptz NOT NULL DEFAULT now()
);

-- [Partial unique WHERE status='active']: Same pattern as user_course_access.
-- Allows re-purchase after refund while enforcing one active access at a time.
CREATE UNIQUE INDEX idx_user_content_access_user_id_content_item_id_active_unique
    ON user_content_access(user_id, content_item_id) WHERE status = 'active';

-- ---------------------------------------------------------------------------
-- 12. USER NOTES
-- ---------------------------------------------------------------------------

-- [user_notes — personal note-taking attached to learning content]:
-- Users can annotate their learning experience at three levels:
--   1. Course-level note   → resource_type='course',        resource_id=course.id
--   2. Content-level note  → resource_type='content_item',  resource_id=content_item.id
--   3. Standalone note     → resource_type=NULL,            resource_id=NULL
-- A single table handles all three cases. The CHECK constraint enforces that
-- resource_type and resource_id are either both NULL (standalone) or both set (attached).
--
-- [No FK on resource_id — polymorphic]: resource_id may point to courses or
-- content_items. PostgreSQL cannot enforce a FK to multiple tables simultaneously.
-- The app validates resource existence based on resource_type before INSERT.
--
-- [deleted_at]: Soft delete — notes are personal data. Physical deletion would
-- break any future "restore deleted notes" feature and loses audit history.
CREATE TABLE user_notes (
    id            uuid        PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id       uuid        NOT NULL REFERENCES users(id) ON DELETE RESTRICT,
    resource_type text        CONSTRAINT user_notes_resource_type_check CHECK (resource_type IN ('course', 'content_item')),
    resource_id   uuid,

    -- [title NOT NULL]: Required for note listings, search results, and any UI preview
    -- that shows notes without loading the full body. A note without a title is
    -- unusable in a list context.
    title         text        NOT NULL,

    -- [description nullable]: Optional short preview or subtitle. Useful for showing
    -- a note summary in cards without rendering the full markdown body.
    description   text,

    body          text        NOT NULL,
    deleted_at    timestamptz,
    created_at    timestamptz NOT NULL DEFAULT now(),
    updated_at    timestamptz NOT NULL DEFAULT now(),

    -- [Paired nullability constraint]: Prevents half-set state where resource_type
    -- is set but resource_id is NULL (or vice versa). Either both are present
    -- (note attached to a resource) or both are NULL (standalone note).
    CONSTRAINT user_notes_resource_check CHECK (
        (resource_type IS NULL AND resource_id IS NULL) OR
        (resource_type IS NOT NULL AND resource_id IS NOT NULL)
    )
);

-- [user_id partial index]: "Give me all active notes for user X" — primary query.
-- Partial index excludes deleted notes from the working set.
CREATE INDEX idx_user_notes_user_id_active ON user_notes(user_id) WHERE deleted_at IS NULL;

-- [resource index]: "Give me all active notes for course Y" or "for content item Z".
-- Enables the resource-attached note display in course/content pages.
CREATE INDEX idx_user_notes_resource_type_resource_id_active ON user_notes(resource_type, resource_id) WHERE resource_type IS NOT NULL AND deleted_at IS NULL;

-- ---------------------------------------------------------------------------
-- 13. CONSULTATIONS
-- ---------------------------------------------------------------------------

-- CLAUDE.md domain: "consultations/ — 1:1 consultation sessions, availability slots"
--
-- [consultation_slots — admin-managed availability calendar]:
-- The admin creates slots (date + start_time + duration) in advance. Users pick
-- from available slots when submitting a brief. This separates scheduling supply
-- (admin creates slots) from demand (users request bookings). This enables
-- future auto-booking without manual coordination.
CREATE TABLE consultation_slots (
    id                 uuid        PRIMARY KEY DEFAULT gen_random_uuid(),
    date               date        NOT NULL,

    -- [time type for start_time]: Pure wall-clock time without date. The slot's
    -- full datetime is reconstructed as (date + start_time) in the app, cast to
    -- the admin's configured timezone. Storing as a separate date+time pair
    -- allows querying "all slots on date X" and "all slots starting at 10:00"
    -- independently.
    start_time         time        NOT NULL,

    -- [CHECK duration_minutes > 0]: Zero-duration or negative-duration slots are
    -- nonsensical. Prevents a data entry mistake from creating unbookable slots.
    duration_minutes   integer     NOT NULL CONSTRAINT consultation_slots_duration_positive CHECK (duration_minutes > 0),

    -- [DEFAULT 'available']: New slots are immediately available. Status transitions:
    -- available → booked (when a booking is confirmed for this slot)
    -- available → blocked (admin blocks the slot, e.g., holiday)
    -- booked → available would happen on cancellation (handled in app).
    status             text        NOT NULL DEFAULT 'available' CONSTRAINT consultation_slots_status_check CHECK (status IN ('available', 'booked', 'blocked')),
    created_by_user_id uuid        NOT NULL REFERENCES users(id) ON DELETE RESTRICT,
    created_at         timestamptz NOT NULL DEFAULT now(),
    updated_at         timestamptz NOT NULL DEFAULT now(),

    -- [UNIQUE (date, start_time)]: No two slots can start at the same time on the
    -- same day. Prevents accidental double-booking at the DB level.
    CONSTRAINT consultation_slots_date_time_unique UNIQUE (date, start_time)
);

-- [Partial index WHERE status='available']: The primary query on this table is:
--   SELECT * FROM consultation_slots WHERE status='available' ORDER BY date, start_time
-- Partial index excludes booked and blocked slots — the relevant working set shrinks
-- significantly over time as slots fill up.
CREATE INDEX idx_consultation_slots_date_start_time_available
    ON consultation_slots(date, start_time) WHERE status = 'available';

-- Admin calendar: view booked slots; also used for bulk-blocking date ranges (e.g. vacation)
CREATE INDEX idx_consultation_slots_date_start_time_booked
    ON consultation_slots(date, start_time) WHERE status = 'booked';

-- CLAUDE.md event flow: POST /briefs → Save Brief → Emit "brief_submitted"
--   → BriefSubmittedWorker: Create Booking → Emit "booking_created"
--
-- [Separate brief and booking tables]: A brief is the user's request (goal, context).
-- A booking is the scheduled session that results from processing the brief.
-- They have different lifecycles:
--   brief   → submitted once, transitions: submitted → processing → booked/failed
--   booking → may be rescheduled (each reschedule = new row, not mutation)
CREATE TABLE consultation_briefs (
    id            uuid        PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id       uuid        NOT NULL REFERENCES users(id) ON DELETE RESTRICT,

    -- [slot_id NOT NULL]: Brief submission requires a slot selection (SLOT-03).
    -- The user picks a slot from available slots before submitting. The worker
    -- atomically marks the slot as 'booked' when processing the brief.
    -- NOT NULL enforces this at the DB level — a brief without a slot cannot exist.
    slot_id       uuid        NOT NULL REFERENCES consultation_slots(id) ON DELETE RESTRICT,

    -- [goal_type — typed consultation category]: Structured field for analytics
    -- and admin routing. Enables: "how many 'recover_after_breakup' consultations
    -- this month?" without parsing free-text. Nullable — users may not always pick
    -- a category, and new categories require only a migration, not code changes.
    -- The free-text 'goal' field below captures the personal details.
    goal_type     text        CONSTRAINT consultation_briefs_goal_type_check CHECK (goal_type IN ('bring_partner_back', 'recover_after_breakup', 'improve_good_relationship', 'recover_problematic_relationship', 'improve_meeting_skills', 'find_good_partner_skills', 'analyze_situation', 'check_partner')),

    goal          text,
    context       text,

    -- [review_notes — admin preparation notes]: Private field visible only to
    -- admin/subadmin. Used to jot down preparation notes before the consultation:
    -- key concerns, prior history, approach. NOT visible to the user.
    -- Stored on the brief (not a separate table) since there is one set of notes
    -- per brief and no history tracking is required for now.
    review_notes  text,

    -- [jsonb metadata]: Flexible extra data — calendar preferences, timezone hints,
    -- referral source. Structure evolves without schema migrations. JSONB is
    -- indexable (GIN) if querying by specific keys becomes necessary later.
    metadata_json jsonb,
    status        text        NOT NULL CONSTRAINT consultation_briefs_status_check CHECK (status IN ('submitted', 'processing', 'booked', 'failed')),

    -- [NOT NULL DEFAULT now()]: No draft state in this model — a brief is always
    -- submitted on creation. submitted_at equals created_at here but is kept for
    -- explicit domain clarity (could diverge if draft support is added later).
    submitted_at  timestamptz NOT NULL DEFAULT now(),
    created_at    timestamptz NOT NULL DEFAULT now(),
    updated_at    timestamptz NOT NULL DEFAULT now(),

    -- [deleted_at on briefs]: Soft delete allows admin to "remove" a brief from
    -- the working queue without losing the FK chain to its bookings and history.
    deleted_at    timestamptz
);

-- [user_id index]: User's consultation history:
--   SELECT * FROM consultation_briefs WHERE user_id = $1 ORDER BY created_at DESC
CREATE INDEX idx_consultation_briefs_user_id ON consultation_briefs(user_id);

-- [status index]: Worker processing queue:
--   SELECT * FROM consultation_briefs WHERE status = 'submitted' LIMIT 10 FOR UPDATE SKIP LOCKED
-- Partial WHERE deleted_at IS NULL: worker/admin queries only active briefs.
CREATE INDEX idx_consultation_briefs_status ON consultation_briefs(status) WHERE deleted_at IS NULL;

CREATE TABLE consultation_bookings (
    id                          uuid        PRIMARY KEY DEFAULT gen_random_uuid(),
    brief_id                    uuid        NOT NULL REFERENCES consultation_briefs(id) ON DELETE RESTRICT,
    user_id                     uuid        NOT NULL REFERENCES users(id) ON DELETE RESTRICT,

    -- [slot_id NOT NULL on booking]: Each booking (original or reschedule) references
    -- the specific slot it occupies. This is critical for the reschedule flow:
    --   1. User reschedules → new booking row created with new slot_id
    --   2. Old booking status → 'rescheduled'; old slot freed (status → 'available')
    --   3. New booking holds new slot_id (status → 'booked')
    -- Without slot_id on booking, it would be impossible to track which slot each
    -- booking in the reschedule chain occupies, making slot liberation ambiguous.
    slot_id                     uuid        NOT NULL REFERENCES consultation_slots(id) ON DELETE RESTRICT,

    -- [nullable payment_id]: Admin-booked or free consultations have no payment.
    payment_id                  uuid        REFERENCES payments(id) ON DELETE RESTRICT,
    status                      text        NOT NULL CONSTRAINT consultation_bookings_status_check CHECK (status IN ('created', 'confirmed', 'cancelled', 'rescheduled', 'completed')),
    scheduled_at                timestamptz,
    confirmed_at                timestamptz,
    cancelled_at                timestamptz,
    completed_at                timestamptz,

    -- [Self-referential rescheduled_from_booking_id]:
    -- Reschedules create a NEW booking row. The original booking status → 'rescheduled'.
    -- The new booking has rescheduled_from_booking_id = original booking's id.
    -- Benefits: full reschedule history preserved; easy to count reschedules per brief;
    -- audit trail of all scheduled times. Mutating a single row would lose this history.
    rescheduled_from_booking_id uuid        REFERENCES consultation_bookings(id),
    handled_by_user_id          uuid        REFERENCES users(id) ON DELETE RESTRICT,
    notes                       text,
    created_at                  timestamptz NOT NULL DEFAULT now(),
    updated_at                  timestamptz NOT NULL DEFAULT now(),

    -- [deleted_at on bookings]: Bookings are financial records. Physical deletion
    -- would break payment_line_items, refunds, and activity_log FK references.
    deleted_at                  timestamptz
);

-- [Partial unique WHERE rescheduled_from_booking_id IS NULL]:
-- The constraint reads: "only one original booking per brief".
-- Original bookings have rescheduled_from_booking_id = NULL → covered by the unique.
-- Rescheduled bookings have a non-NULL predecessor → excluded from the unique.
-- This allows: 1 original + N reschedule rows per brief without constraint violations.
CREATE UNIQUE INDEX idx_consultation_bookings_brief_id_unique
    ON consultation_bookings(brief_id) WHERE rescheduled_from_booking_id IS NULL;

CREATE INDEX idx_consultation_bookings_brief_id ON consultation_bookings(brief_id);
CREATE INDEX idx_consultation_bookings_slot_id ON consultation_bookings(slot_id);
CREATE INDEX idx_consultation_bookings_user_id_status ON consultation_bookings(user_id, status);

-- [scheduled_at index]: Admin dashboard "upcoming sessions" query:
--   SELECT * FROM consultation_bookings WHERE scheduled_at > now() ORDER BY scheduled_at ASC
-- Partial (WHERE NOT NULL) skips unscheduled rows (status='created', not yet confirmed).
CREATE INDEX idx_consultation_bookings_scheduled_at
    ON consultation_bookings(scheduled_at) WHERE scheduled_at IS NOT NULL;

-- ---------------------------------------------------------------------------
-- 14. CONSULTATION OUTCOMES
-- ---------------------------------------------------------------------------

-- [consultation_outcomes — post-consultation results for the user]:
-- Separated from consultation_briefs because outcomes represent a different lifecycle
-- phase: briefs are pre-consultation (request + preparation), outcomes are
-- post-consultation (what was learned and recommended).
-- Mixing pre- and post-consultation data in one table makes the brief's lifecycle
-- ambiguous — a brief with NULL recommendation could mean "not consulted yet" or
-- "consulted but admin forgot to fill in". A separate table makes the state explicit:
-- no row = consultation not yet completed; row exists = outcomes recorded.
--
-- [UNIQUE (brief_id)]: One outcome per brief. Enforced at DB level so the
-- consultation service cannot accidentally create duplicate outcome rows.
--
-- [recommendation — visible to user after consultation]: Personal recommendations
-- written by the admin/consultant. The user can return to read them anytime
-- from their consultation history.
--
-- [knowledge — private insights gained during consultation]: Personal notes the
-- user received (strategies, frameworks, action steps). Stored permanently so
-- the user never loses access to what they learned, even months later.
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

-- [campaigns table — 3NF normalization of expenses]:
-- campaign_name is functionally dependent on campaign_id, not on the expense itself
-- (a 3NF violation: transitive dependency). Separating into a campaigns table ensures:
--   1. Campaign names are consistent across all expenses for the same campaign
--   2. Renaming a campaign requires one UPDATE, not N UPDATEs across all expense rows
--   3. campaign_id FK enforces referential integrity — no typos in campaign IDs
--
-- [text PRIMARY KEY]: campaign_id is an external identifier (Google Ads campaign ID,
-- Meta campaign ID, etc.). Using the external ID as PK avoids a surrogate UUID
-- and makes joins from expenses readable without an extra lookup.
CREATE TABLE campaigns (
    id          text        PRIMARY KEY,
    name        text        NOT NULL,

    -- [platform CHECK]: Which ad platform this campaign runs on. Typed for analytics
    -- (ROAS per platform, budget allocation). DEFAULT 'google' — the primary ad channel.
    -- New platforms (e.g. linkedin) require only a migration, not code changes.
    platform    text        NOT NULL DEFAULT 'google' CONSTRAINT campaigns_platform_check CHECK (platform IN ('google', 'meta', 'tiktok', 'other')),

    created_at  timestamptz NOT NULL DEFAULT now(),
    updated_at  timestamptz NOT NULL DEFAULT now()
);

-- ---------------------------------------------------------------------------
-- 16. EXPENSES
-- ---------------------------------------------------------------------------

-- CLAUDE.md domain: "pnl/ — P&L reporting, expenses, campaign costs"
-- Expenses feed the monthly_pnl view: profit = revenue - expenses.
-- Stored in DB (not a spreadsheet) to make them queryable and auditable.
CREATE TABLE expenses (
    id                  uuid        PRIMARY KEY DEFAULT gen_random_uuid(),
    category            text        NOT NULL CONSTRAINT expenses_category_check CHECK (category IN ('hosting', 'tooling', 'contractor', 'marketing', 'other')),

    -- [CHECK amount > 0]: An expense of zero or negative amount is nonsensical.
    amount              bigint      NOT NULL CONSTRAINT expenses_amount_positive CHECK (amount > 0),

    -- [DEFAULT 'PLN']: Same rationale as payments.currency — primary market is Poland.
    currency            text        NOT NULL DEFAULT 'PLN',
    description         text,
    occurred_at         timestamptz NOT NULL,
    vendor              text,

    -- [campaign_id FK → campaigns]: Links expense to a campaign record.
    -- campaign_name removed — now derived via JOIN to campaigns table (3NF).
    campaign_id         text        REFERENCES campaigns(id) ON DELETE RESTRICT,

    -- [nullable course_id FK]: Marketing/hosting expenses can be attributed to a
    -- specific course for per-course P&L. NULL = platform-wide expense.
    course_id           uuid        REFERENCES courses(id) ON DELETE RESTRICT,

    -- [nullable content_item_id FK]: Expenses can also be attributed to an individual
    -- content item (e.g., production cost for a specific video). Mutually exclusive
    -- with course_id — an expense targets either a course, a content item, or neither
    -- (platform-wide). The CHECK constraint enforces this at the DB level.
    content_item_id     uuid        REFERENCES content_items(id) ON DELETE RESTRICT,

    external_reference  text,
    metadata_json       jsonb,
    created_by_user_id  uuid        NOT NULL REFERENCES users(id) ON DELETE RESTRICT,
    created_at          timestamptz NOT NULL DEFAULT now(),

    -- [updated_at]: Expenses may be corrected after entry (wrong amount, category).
    -- updated_at lets dependent caches detect stale P&L calculations.
    updated_at          timestamptz NOT NULL DEFAULT now(),

    -- [expenses_resource_exclusive]: An expense belongs to at most one resource.
    -- course_id and content_item_id cannot both be set simultaneously. Both can be
    -- NULL (platform-wide expense). This prevents ambiguous attribution in P&L reports.
    CONSTRAINT expenses_resource_exclusive CHECK (course_id IS NULL OR content_item_id IS NULL)
);

CREATE INDEX idx_expenses_occurred_at ON expenses(occurred_at);
CREATE INDEX idx_expenses_campaign_id ON expenses(campaign_id) WHERE campaign_id IS NOT NULL;
CREATE INDEX idx_expenses_course_id ON expenses(course_id) WHERE course_id IS NOT NULL;
CREATE INDEX idx_expenses_content_item_id ON expenses(content_item_id) WHERE content_item_id IS NOT NULL;

-- ---------------------------------------------------------------------------
-- 17. NOTIFICATION PREFERENCES
-- ---------------------------------------------------------------------------

-- [notification_preferences — separate table from users]:
-- Preferences are a settings concern, not an identity concern. Keeping them in a
-- separate table avoids adding many boolean columns to the users table, which is
-- loaded on every auth check. Rows are created lazily (first time user visits settings)
-- or eagerly (on registration). The notification worker reads this table before
-- deciding whether to enqueue an email delivery.
--
-- [PRIMARY KEY = user_id]: 1:1 relation — one preferences row per user. Using
-- user_id as PK (not a surrogate UUID) removes a JOIN step: SELECT directly by
-- user_id, no need to look up a separate id first.
--
-- [boolean NOT NULL DEFAULT true]: All notification types default to opt-in.
-- This is the safest default for a new platform — users expect to receive relevant
-- notifications. They can opt out explicitly in settings.
CREATE TABLE notification_preferences (
    user_id                      uuid        PRIMARY KEY REFERENCES users(id) ON DELETE RESTRICT,
    email_on_booking_confirmed   boolean     NOT NULL DEFAULT true,
    email_on_reminder            boolean     NOT NULL DEFAULT true,
    email_on_new_content         boolean     NOT NULL DEFAULT true,

    -- [email_on_support_reply — user-facing]: Notifies the user when admin/subadmin
    -- replies to their support message.
    email_on_support_reply       boolean     NOT NULL DEFAULT true,

    -- [email_on_consultation_booked + email_on_support_message — admin-facing]:
    -- These two are admin-targeted notifications (when a user books a consultation,
    -- or sends a support message, the admin/subadmin should be notified).
    -- Separate columns allow admins to configure their own notification preferences
    -- independently from user-facing notification preferences.
    email_on_consultation_booked boolean     NOT NULL DEFAULT true,
    email_on_support_message     boolean     NOT NULL DEFAULT true,

    -- [email_on_announcement — user-facing]: Platform-wide announcement sent to all users.
    -- Opt-out available; defaults to true (users expect to hear about new features).
    email_on_announcement        boolean     NOT NULL DEFAULT true,

    -- [email_on_failed_job — admin-facing]: Notifies admin when an event exhausts retries
    -- and lands in the DLQ (failed_jobs). Critical operational signal — admin should
    -- review failed_jobs dashboard. Default true because missed failures = silent data loss.
    email_on_failed_job          boolean     NOT NULL DEFAULT true,

    created_at                   timestamptz NOT NULL DEFAULT now(),
    updated_at                   timestamptz NOT NULL DEFAULT now()
);

-- ---------------------------------------------------------------------------
-- 18. NOTIFICATIONS
-- ---------------------------------------------------------------------------

-- CLAUDE.md domain: "notifications/ — Notification delivery"
-- CLAUDE.md event flow: NotificationWorker: Send email → Emit "notification_sent"
CREATE TABLE notifications (
    id              uuid        PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id         uuid        NOT NULL REFERENCES users(id) ON DELETE RESTRICT,

    -- [kind CHECK]: Typed notification kinds let the frontend render kind-specific UI.
    -- e.g. booking_created → show "View your booking" CTA with booking_id from payload_json.
    -- 'consultation_booked_admin' and 'support_message_admin' are admin-targeted.
    -- 're_engagement' is sent by the scheduled re-engagement worker when a user has
    -- in_progress content/course with last_interacted_at older than 14 days.
    -- 'failed_job_alert' — DLQ alert to admin when an event exhausts retries.
    -- 'announcement_sent' — platform-wide announcement broadcast to all active users.
    kind            text        NOT NULL CONSTRAINT notifications_kind_check            CHECK (kind IN ('booking_created', 'booking_confirmed', 'payment_confirmed', 'refund_processed', 'reminder', 'admin_message', 'new_content', 'consultation_booked_admin', 'support_message_admin', 're_engagement', 'failed_job_alert', 'announcement_sent')),
    channel         text        NOT NULL CONSTRAINT notifications_channel_check         CHECK (channel IN ('email', 'in_app')),
    title           text,
    body            text,

    -- ['skipped' instead of 'stubbed']: In local dev, the email module is a no-op stub
    -- (CLAUDE.md infrastructure: "email/ — Email sending — Resend (prod), stub (local dev)").
    -- 'skipped' is an implementation-neutral term — the delivery was intentionally bypassed.
    -- 'stubbed' would leak that there's a stub implementation in the schema.
    delivery_status text        NOT NULL CONSTRAINT notifications_delivery_status_check CHECK (delivery_status IN ('pending', 'skipped', 'sent', 'failed')),

    -- [jsonb payload_json]: Notification workers need context to render emails:
    -- booking ID, course title, scheduled time, etc. JSONB stores this without a
    -- separate notification_data table. Queryable if needed: payload_json->>'booking_id'.
    payload_json    jsonb,
    scheduled_for   timestamptz,

    -- [read_at]: In-app notification read tracking. NULL = unread. Not meaningful for
    -- email channel (would require open-pixel tracking — out of scope here).
    read_at         timestamptz,
    sent_at         timestamptz,
    created_at      timestamptz NOT NULL DEFAULT now()
);

CREATE INDEX idx_notifications_user_id_created_at ON notifications(user_id, created_at DESC);

-- [Partial index for scheduled+pending]: Workers query:
--   SELECT * FROM notifications WHERE delivery_status='pending' AND scheduled_for <= now()
-- Partial index excludes: already sent, failed, and non-scheduled rows.
-- Much smaller and faster than a full index on (scheduled_for).
CREATE INDEX idx_notifications_scheduled_for_pending
    ON notifications(scheduled_for) WHERE delivery_status = 'pending' AND scheduled_for IS NOT NULL;

-- [Unread count query]: SELECT COUNT(*) FROM notifications WHERE user_id = $1 AND read_at IS NULL
CREATE INDEX idx_notifications_user_id_unread
    ON notifications(user_id) WHERE read_at IS NULL;

-- ---------------------------------------------------------------------------
-- 19. ANNOUNCEMENTS
-- ---------------------------------------------------------------------------

-- [announcements — platform-wide broadcasts from admin]:
-- CLAUDE.md domain: "admin/ — User management, announcements, audit trail".
-- Announcements are different from notifications: they are not per-user events
-- but platform-wide messages (new features, scheduled maintenance, promotions).
-- They are read by the frontend at page load (visible banner or modal),
-- filtered by expires_at to show only current ones.
--
-- [expires_at nullable]: A NULL expires_at means the announcement is permanent
-- (e.g., a persistent welcome banner). Non-NULL means it auto-expires at that time.
-- The partial index below leverages this to only index active announcements.
CREATE TABLE announcements (
    id                 uuid        PRIMARY KEY DEFAULT gen_random_uuid(),
    title              text        NOT NULL,
    body               text        NOT NULL,
    created_by_user_id uuid        NOT NULL REFERENCES users(id) ON DELETE RESTRICT,
    expires_at         timestamptz,
    created_at         timestamptz NOT NULL DEFAULT now(),
    updated_at         timestamptz NOT NULL DEFAULT now()
);

-- [Two indexes, not one partial with now()]:
-- The original design considered a partial index with `expires_at > now()` in the
-- predicate, but PostgreSQL requires IMMUTABLE functions in partial index predicates.
-- now() is STABLE (same value per transaction, not per index build) — PostgreSQL
-- rejects such a CREATE INDEX with "functions in index predicate must be marked IMMUTABLE".
--
-- Solution: two plain indexes; the WHERE clause stays in the query, not the index.
--   idx_announcements_created_at  — covers ORDER BY created_at DESC for all announcements
--   idx_announcements_expires_at  — covers range filtering on expires_at for non-NULL rows
-- The frontend query: SELECT ... WHERE expires_at IS NULL OR expires_at > now()
-- ORDER BY created_at DESC — uses both indexes via bitmap scan or index scan.
CREATE INDEX idx_announcements_created_at ON announcements(created_at DESC);
CREATE INDEX idx_announcements_expires_at ON announcements(expires_at) WHERE expires_at IS NOT NULL;

-- ---------------------------------------------------------------------------
-- 20. ACTIVITY LOG
-- ---------------------------------------------------------------------------

-- CLAUDE.md domain: "analytics/ — UTM attribution, activity tracking"
-- activity_log ≠ notifications: this records ALL user actions for analytics/audit
-- (lesson starts, video watches, page views). Notifications are a communication channel.
-- Different consumers: analytics/ reads activity_log; notifications/ reads notifications.
CREATE TABLE activity_log (
    id              uuid        PRIMARY KEY DEFAULT gen_random_uuid(),

    -- [nullable user_id]: Anonymous events (unauthenticated page views, bot crawls)
    -- have no user. The FK is still enforced for non-NULL user_id values.
    user_id         uuid        REFERENCES users(id) ON DELETE RESTRICT,
    event_type      text        NOT NULL,
    resource_type   text        NOT NULL,
    resource_id     uuid        NOT NULL,
    metadata_json   jsonb,
    created_at      timestamptz NOT NULL DEFAULT now()
);

CREATE INDEX idx_activity_log_user_id_created_at ON activity_log(user_id, created_at DESC);

-- [resource index]: Enables queries like "all events on booking #X" or "all activity
-- for course #Y". Without this, resource history scans the entire log table.
CREATE INDEX idx_activity_log_resource_type_resource_id ON activity_log(resource_type, resource_id);

-- ---------------------------------------------------------------------------
-- 21. EVENT RELIABILITY
-- ---------------------------------------------------------------------------

-- CLAUDE.md event system: "Workers use BLPop (blocking pop, 5s timeout).
-- On failure: retry 3× with exponential backoff, then → DLQ."
--
-- [Transactional Outbox Pattern — event_outbox]:
-- Problem without outbox: service saves a brief → crashes before emitting to Redis
--   → brief exists in DB but no worker ever processes it → silent data loss.
-- Solution: write the event to event_outbox in the SAME database transaction as
--   the business object. A separate relay process reads pending events and publishes
--   to Redis. If the relay crashes, events remain in 'pending' state and are retried.
--   At-least-once delivery guaranteed: DB commit = both business data and event saved atomically.
CREATE TABLE event_outbox (
    id              uuid        PRIMARY KEY DEFAULT gen_random_uuid(),
    aggregate_type  text        NOT NULL,
    aggregate_id    uuid        NOT NULL,
    event_type      text        NOT NULL,

    -- [NOT NULL payload_json]: An event without a payload cannot be processed correctly.
    -- NULL payload would cause silent failures or panics in the worker. DB rejects it.
    payload_json    jsonb       NOT NULL,
    status          text        NOT NULL CONSTRAINT event_outbox_status_check        CHECK (status IN ('pending', 'published', 'failed')),

    -- [CHECK >= 0 attempt_count]: Starts at 0, increments on each attempt.
    -- A negative count indicates a bug in worker increment logic — rejected at DB level.
    attempt_count   integer     NOT NULL DEFAULT 0 CONSTRAINT event_outbox_attempt_count_check CHECK (attempt_count >= 0),

    -- [available_at for exponential backoff]: After a failed attempt, the worker sets:
    --   available_at = now() + interval '1 second' * power(2, attempt_count)
    -- Workers only pick up events WHERE available_at <= now(), implementing backoff
    -- without a separate scheduler process.
    available_at    timestamptz NOT NULL DEFAULT now(),

    -- [locked_until for distributed locking]: Multiple worker instances could pick
    -- up the same event simultaneously without this. A worker claims an event by
    -- setting locked_until = now() + 30s before processing. If the worker crashes,
    -- the lock expires and another worker picks it up (at-least-once semantics).
    locked_until    timestamptz,
    published_at    timestamptz,
    last_error      text,
    created_at      timestamptz NOT NULL DEFAULT now(),
    updated_at      timestamptz NOT NULL DEFAULT now()
);

-- [Partial index WHERE status='pending']: Workers ONLY query pending events.
-- Published and failed events are never polled. Partial index = fraction of total rows.
CREATE INDEX idx_event_outbox_status_available_at_pending
    ON event_outbox(status, available_at) WHERE status = 'pending';

-- [locked_until index]: Cleanup job resets stale locks:
--   UPDATE event_outbox SET locked_until=NULL WHERE locked_until < now() AND status='pending'
-- Without this index, the cleanup job scans all events.
CREATE INDEX idx_event_outbox_locked_until
    ON event_outbox(locked_until) WHERE locked_until IS NOT NULL;

-- [Dead Letter Queue — failed_jobs]:
-- CLAUDE.md: "On failure: retry 3× with exponential backoff, then → DLQ".
-- Events that exhausted retries move here. Stored in PostgreSQL (not Redis) for:
--   1. Durability — persists across Redis restarts
--   2. Queryability — admins can inspect, filter, replay failed jobs
--   3. Auditability — resolution_note records what was done with each failure
CREATE TABLE failed_jobs (
    id              uuid        PRIMARY KEY DEFAULT gen_random_uuid(),
    event_type      text        NOT NULL,
    queue_name      text        NOT NULL,
    payload_json    jsonb,
    attempt_count   integer     NOT NULL DEFAULT 0 CONSTRAINT failed_jobs_attempt_count_check CHECK (attempt_count >= 0),
    error_message   text,
    failed_at       timestamptz NOT NULL,

    -- [resolved_at + resolution_note]: Admin resolves a DLQ entry by either:
    --   1. Replaying it (re-inserting into event_outbox)
    --   2. Skipping it (marking resolved with a reason)
    -- These fields track resolution for compliance and post-mortems.
    resolved_at     timestamptz,
    resolution_note text,
    created_at      timestamptz NOT NULL DEFAULT now(),
    updated_at      timestamptz NOT NULL DEFAULT now()
);

-- [Partial index WHERE resolved_at IS NULL]: DLQ management query:
--   SELECT * FROM failed_jobs WHERE resolved_at IS NULL ORDER BY failed_at DESC
-- Partial index covers only unresolved jobs — the active working set.
-- Resolved jobs are archived state and never queried this way again.
CREATE INDEX idx_failed_jobs_failed_at_unresolved
    ON failed_jobs(failed_at DESC) WHERE resolved_at IS NULL;

-- ---------------------------------------------------------------------------
-- 22. ADMIN ACTIONS
-- ---------------------------------------------------------------------------

-- CLAUDE.md domain: "admin/ — Admin operations"
-- Every admin operation is recorded here. High-impact operations (confirming bookings,
-- granting course access, issuing refunds, blocking users) need a paper trail for:
--   1. Compliance and accountability
--   2. Debugging unexpected state ("who granted this access?")
--   3. Reversal context ("what did the admin intend?")
CREATE TABLE admin_actions (
    id              uuid        PRIMARY KEY DEFAULT gen_random_uuid(),
    admin_user_id   uuid        NOT NULL REFERENCES users(id) ON DELETE RESTRICT,

    -- [Extended action_type list]: Covers all admin operations in scope.
    -- 'create_gift_coupon' / 'revoke_gift_coupon' — gift coupon lifecycle management.
    -- 'publish_article' / 'delete_article' — admin-managed editorial content.
    -- admin_user_id references users.role which includes 'subadmin' — so subadmin
    -- actions are also captured here using the same table.
    action_type     text        NOT NULL CONSTRAINT admin_actions_action_type_check CHECK (action_type IN ('confirm_booking', 'cancel_booking', 'grant_course_access', 'issue_refund', 'record_expense', 'block_user', 'reschedule_booking', 'close_support_chat', 'assign_subadmin', 'revoke_subadmin', 'deactivate_user', 'delete_user', 'create_gift_coupon', 'revoke_gift_coupon', 'publish_article', 'delete_article')),

    -- [Extended target_type list]: 'review' and 'announcement' cover course_reviews/
    -- content_reviews and announcements. 'article' and 'gift_coupon' added for
    -- the new editorial and coupon management flows.
    -- The polymorphic pattern (no FK, discriminated by target_type) is the same
    -- as payment_line_items.resource_id.
    target_type     text        NOT NULL CONSTRAINT admin_actions_target_type_check CHECK (target_type IN ('user', 'booking', 'course', 'failed_job', 'payment', 'support_chat', 'review', 'announcement', 'article', 'gift_coupon')),

    -- [no FK on target_id — polymorphic reference]:
    -- target_id may reference users, bookings, courses, failed_jobs, payments, etc.
    -- PostgreSQL cannot enforce a FK to multiple tables. Validated in the admin service
    -- based on target_type.
    target_id       uuid        NOT NULL,
    details_json    jsonb,
    created_at      timestamptz NOT NULL DEFAULT now()
);

-- [admin_user_id + created_at DESC]: Standard audit query: "show all actions by admin X".
-- Descending order in the index avoids a filesort on ORDER BY created_at DESC.
CREATE INDEX idx_admin_actions_admin_user_id_created_at ON admin_actions(admin_user_id, created_at DESC);

-- [target_type + target_id]: "Show all admin actions on booking #X" — per-resource audit.
CREATE INDEX idx_admin_actions_target_type_target_id ON admin_actions(target_type, target_id);

-- ---------------------------------------------------------------------------
-- 23. SUPPORT CHAT
-- ---------------------------------------------------------------------------

-- CLAUDE.md domain: "notifications/ — reminder scheduling" (support replies trigger notifications)
-- Support chat is not a general messaging system — it is a 1:1 channel between a user
-- and the support team. Only one OPEN chat per user is allowed at any time; however,
-- a user may have multiple CLOSED (archived) chats over time (e.g., previous support sessions).
-- This is enforced via a partial unique index rather than a table-level UNIQUE:
--
-- [WHY partial unique instead of UNIQUE (user_id)]:
-- A table-level UNIQUE (user_id) would prevent a user from ever opening a second
-- support chat after closing the first. With a partial index WHERE status='open',
-- only one open chat per user is enforced, while closed chats can accumulate as history.
-- This is the correct business model: chat history is preserved, not overwritten.
CREATE TABLE support_chats (
    id          uuid        PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id     uuid        NOT NULL REFERENCES users(id) ON DELETE RESTRICT,
    status      text        NOT NULL DEFAULT 'open'
        CONSTRAINT support_chats_status_check CHECK (status IN ('open', 'closed')),
    created_at  timestamptz NOT NULL DEFAULT now(),
    updated_at  timestamptz NOT NULL DEFAULT now()
);

-- [Partial unique WHERE status='open']: Allows one open chat per user at a time.
-- Closed chats are excluded — a user can open a new chat after their previous one is closed.
CREATE UNIQUE INDEX idx_support_chats_user_id_open_unique
    ON support_chats(user_id) WHERE status = 'open';

CREATE INDEX idx_support_chats_status_open ON support_chats(status) WHERE status = 'open';

-- [support_messages.sender_id]: Identifies who sent each message — user or admin/subadmin.
-- JOIN to users on sender_id + users.role distinguishes user messages from staff replies.
-- No separate 'response_sender_id' needed: sender_id already captures this information,
-- and multiple staff members can reply in the same chat over time.
CREATE TABLE support_messages (
    id          uuid        PRIMARY KEY DEFAULT gen_random_uuid(),
    chat_id     uuid        NOT NULL REFERENCES support_chats(id) ON DELETE RESTRICT,
    sender_id   uuid        NOT NULL REFERENCES users(id) ON DELETE RESTRICT,
    body        text        NOT NULL,
    read_at     timestamptz,
    created_at  timestamptz NOT NULL DEFAULT now()
);

CREATE INDEX idx_support_messages_chat_id_created_at ON support_messages(chat_id, created_at ASC);
CREATE INDEX idx_support_messages_chat_id_unread ON support_messages(chat_id) WHERE read_at IS NULL;

-- ---------------------------------------------------------------------------
-- 24. P&L VIEWS
-- ---------------------------------------------------------------------------

-- [Regular VIEWs, not MATERIALIZED VIEWs]: P&L data must be current (no staleness).
-- Regular views compute on read. At expected volume (SaaS learning platform, thousands
-- of payments/month) the query is fast enough. Upgrade to MATERIALIZED VIEW + refresh
-- job (pg_cron or app-level) if query time becomes unacceptable.
--
-- [GROUP BY period AND currency]: Platform supports PLN + potentially EUR/USD.
-- Without currency in GROUP BY, SUM(amount) mixes PLN and USD — the resulting
-- "revenue" number would be financially meaningless.

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

-- [FULL OUTER JOIN, not LEFT JOIN]:
-- LEFT JOIN monthly_revenue → expenses would silently drop months that have expenses
-- but NO revenue (e.g., a month where the platform paid hosting but earned nothing).
-- Those months would show profit = NULL instead of being recorded as a loss.
-- FULL OUTER JOIN ensures both directions are captured:
--   revenue month with no expenses  → profit = net_revenue (expenses COALESCE to 0)
--   expense month with no revenue   → profit = -total_expenses (revenue COALESCE to 0)
-- This is the correct P&L representation for a platform with ongoing fixed costs.
--
-- [CTE instead of view-on-view]: Inlining monthly_revenue logic here (rather than
-- joining the monthly_revenue view) avoids double-planning and makes the FULL OUTER
-- JOIN semantics explicit and reviewable in one place.
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
    COALESCE(rev.period, exp.period)                AS period,
    COALESCE(rev.currency, exp.currency)            AS currency,
    COALESCE(rev.net_revenue_cents, 0)              AS net_revenue_cents,
    COALESCE(exp.total_expenses_cents, 0)           AS total_expenses_cents,
    COALESCE(rev.net_revenue_cents, 0) - COALESCE(exp.total_expenses_cents, 0) AS profit_cents
FROM revenue rev
FULL OUTER JOIN expense_totals exp
    ON exp.period = rev.period AND exp.currency = rev.currency;

-- ---------------------------------------------------------------------------
-- 25. ACCOUNT RECOVERY TOKENS
-- ---------------------------------------------------------------------------

-- [account_recovery_tokens — separate from password_reset_tokens]:
-- Account recovery targets soft-deleted users who want to reclaim their account.
-- This is a fundamentally different flow from password reset:
--   password_reset: user exists, is active, has forgotten their password
--   account_recovery: user's account has deleted_at set, they want it restored
-- Separate table keeps each flow's logic and expiry policy independent.
-- Recovery sets users.deleted_at = NULL atomically on token consumption.
--
-- [ON DELETE RESTRICT on user_id]: Prevents deleting a user who has an active recovery
-- token. Race condition: admin deletes user mid-recovery flow → token becomes orphan,
-- user loses their only way back. RESTRICT forces explicit cleanup of token table first.
-- Migration 000004 changed this from CASCADE to RESTRICT for this safety reason.
CREATE TABLE account_recovery_tokens (
    id                      uuid        PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id                 uuid        NOT NULL REFERENCES users(id) ON DELETE RESTRICT,
    token_hash              varchar(64) NOT NULL,
    expires_at              timestamptz NOT NULL,
    used_at                 timestamptz,
    -- [invalidated_at + invalidated_by_user_id]: Same audit pattern as other token tables.
    -- Captures who explicitly invalidated this token (admin, system) vs natural expiry.
    invalidated_at          timestamptz,
    invalidated_by_user_id  uuid        REFERENCES users(id) ON DELETE RESTRICT,
    created_at              timestamptz NOT NULL DEFAULT now(),
    CONSTRAINT account_recovery_tokens_token_hash_unique     UNIQUE (token_hash),

    -- [length > 0]: Guards against storing an empty string hash.
    -- The app generates tokens via crypto/rand — this check is a last-resort defense.
    CONSTRAINT account_recovery_tokens_token_hash_nonempty   CHECK (length(token_hash) > 0),

    -- [expires_at > created_at]: A token that expires before or at creation time is
    -- immediately invalid. Catches misconfigured token TTL in the auth service.
    CONSTRAINT account_recovery_tokens_expires_after_created CHECK (expires_at > created_at),

    -- [used_at >= created_at]: A token cannot have been used before it was created.
    -- Defensive against clock skew bugs or incorrect field assignment.
    CONSTRAINT account_recovery_tokens_used_after_created    CHECK (used_at IS NULL OR used_at >= created_at)
);

-- [Partial index WHERE used_at IS NULL]: Only unused tokens are ever looked up
-- by user_id (to validate a recovery request). Used tokens are archived state.
CREATE INDEX idx_account_recovery_tokens_user_id_unused
    ON account_recovery_tokens(user_id) WHERE used_at IS NULL;

-- ---------------------------------------------------------------------------
-- 26. ARTICLES
-- ---------------------------------------------------------------------------

-- [articles — admin-managed editorial content]:
-- CLAUDE.md domain: "articles/ — Admin-managed articles (draft/published), SEO metadata,
-- soft delete, public SSR pages". Articles are long-form content authored by admin/subadmin.
-- Unlike courses and content items, articles do NOT have a user access model —
-- they are publicly readable (gated by status=published AND deleted_at IS NULL).
--
-- [Two-state status, not three]: Articles have 'draft' and 'published' only.
-- No 'archived' — articles are either visible or soft-deleted. Archiving a published
-- article (making it temporarily invisible without deleting) can be done by
-- reverting to 'draft' status.
--
-- [articles_published_requires_published_at]: Enforces that a published article
-- always has a published_at timestamp. Without this, a status='published' row with
-- NULL published_at would break sitemap generation and SEO lastmod logic.
--
-- [deleted_at on articles]: Soft delete preserves SEO history and allows recovery.
-- Hard DELETE would cause 404 on cached/indexed URLs, which search engines penalize.
CREATE TABLE articles (
    id                  uuid        PRIMARY KEY DEFAULT gen_random_uuid(),
    slug                text        NOT NULL,
    title               text        NOT NULL,

    -- [excerpt nullable]: Optional short summary shown in article listings and OG tags.
    -- If NULL, the frontend falls back to the first N characters of body.
    excerpt             text,

    body                text        NOT NULL,
    seo_title           text,
    seo_description     text,
    og_image_url        text,
    is_indexable        boolean     NOT NULL DEFAULT true,

    -- [DEFAULT 'draft']: New articles start as drafts. Explicit admin action required
    -- to publish. Prevents accidentally exposing unfinished content.
    status              text        NOT NULL DEFAULT 'draft',
    published_at        timestamptz,
    created_by_user_id  uuid        NOT NULL REFERENCES users(id) ON DELETE RESTRICT,
    deleted_at          timestamptz,
    created_at          timestamptz NOT NULL DEFAULT now(),
    updated_at          timestamptz NOT NULL DEFAULT now(),

    CONSTRAINT articles_slug_unique                     UNIQUE (slug),
    CONSTRAINT articles_status_check                    CHECK (status IN ('draft', 'published')),

    -- [nonempty CHECKs]: Guards against inserting whitespace-only values.
    -- btrim() strips leading/trailing whitespace before the empty-string comparison.
    CONSTRAINT articles_title_nonempty                  CHECK (btrim(title) <> ''),
    CONSTRAINT articles_slug_nonempty                   CHECK (btrim(slug) <> ''),
    CONSTRAINT articles_body_nonempty                   CHECK (btrim(body) <> ''),
    CONSTRAINT articles_excerpt_nonempty                CHECK (excerpt IS NULL OR btrim(excerpt) <> ''),

    -- [Chronological ordering constraints]: Prevents impossible timestamp combinations
    -- caused by clock bugs or incorrect field assignments.
    CONSTRAINT articles_published_at_after_created      CHECK (published_at IS NULL OR published_at >= created_at),
    CONSTRAINT articles_deleted_at_after_created        CHECK (deleted_at IS NULL OR deleted_at >= created_at),

    -- [published_requires_published_at]: A published article without published_at would
    -- break sitemap lastmod and SEO tooling. This constraint closes the gap.
    CONSTRAINT articles_published_requires_published_at CHECK (status != 'published' OR published_at IS NOT NULL)
);

-- [created_at DESC]: Default article listing order — newest first.
CREATE INDEX idx_articles_created_at ON articles(created_at DESC);

-- [Partial index for public listings]: Only published and non-deleted articles appear
-- on the public SSR site. Partial index excludes drafts and deleted rows from the
-- working set, matching the exact WHERE clause used by Nuxt SSR pages.
CREATE INDEX idx_articles_published_at
    ON articles(published_at DESC) WHERE status = 'published' AND deleted_at IS NULL;

-- ---------------------------------------------------------------------------
-- 27. GIFT COUPONS
-- ---------------------------------------------------------------------------

-- [gift_coupons — admin-created single-use access codes]:
-- CLAUDE.md domain: "payments/ — gift coupons".
-- A gift coupon grants course access (access_type='gift_coupon') without a payment.
-- The admin creates a coupon with a unique code; the user redeems it via a separate
-- endpoint that atomically marks the coupon as redeemed and inserts user_course_access.
--
-- [course_id nullable]: A coupon can be course-specific (most common) or
-- platform-general (redeemable for any course — resolved at redemption time).
-- Nullable FK is the simplest model: NULL = generic, non-NULL = course-specific.
--
-- [redemption_consistency CHECK]: Prevents half-set redemption state:
-- redeemed_by_user_id and redeemed_at must both be NULL (not yet redeemed) or
-- both be NOT NULL (redeemed). A row with one field set and the other NULL indicates
-- a bug in the redemption transaction — this constraint catches it at the DB level.
--
-- [No deleted_at]: Gift coupons are operational records. Admins revoke them by
-- recording a 'revoke_gift_coupon' admin_actions row and marking the coupon as
-- redeemed (with admin as the redeemed_by_user_id). Physical deletion is not needed —
-- the coupon code's uniqueness must be preserved in history.
CREATE TABLE gift_coupons (
    id                  uuid        PRIMARY KEY DEFAULT gen_random_uuid(),

    -- [code text]: Human-readable redemption code (e.g., 'GIFT-ABC123').
    -- Case-insensitive uniqueness enforced by UNIQUE INDEX on LOWER(code) below —
    -- prevents 'SAVE10' and 'save10' from being different coupons.
    -- Migration 000003 dropped the original UNIQUE CONSTRAINT (case-sensitive) and
    -- replaced it with a function-based unique index.
    code                text        NOT NULL,

    -- [nullable course_id]: NULL = generic coupon (any course). Non-NULL = locked to one course.
    course_id           uuid        REFERENCES courses(id) ON DELETE RESTRICT,
    expires_at          timestamptz,
    redeemed_by_user_id uuid        REFERENCES users(id) ON DELETE RESTRICT,
    redeemed_at         timestamptz,
    created_by_user_id  uuid        NOT NULL REFERENCES users(id) ON DELETE RESTRICT,
    created_at          timestamptz NOT NULL DEFAULT now(),
    updated_at          timestamptz NOT NULL DEFAULT now(),

    -- [code_nonempty]: Prevents whitespace-only codes that would be invisible in URLs.
    CONSTRAINT gift_coupons_code_nonempty             CHECK (btrim(code) <> ''),

    -- [expires_after_created]: A coupon that expired before it was created is immediately
    -- invalid — catches misconfigured expiry settings in the admin panel.
    CONSTRAINT gift_coupons_expires_after_created     CHECK (expires_at IS NULL OR expires_at > created_at),

    -- [redeemed_at_after_created]: Cannot be redeemed before it was created.
    CONSTRAINT gift_coupons_redeemed_at_after_created CHECK (redeemed_at IS NULL OR redeemed_at >= created_at),

    -- [redemption_consistency]: See design note above. Both fields NULL or both NOT NULL.
    -- Prevents partial redemption state (redeemed_by_user_id set but redeemed_at NULL, or vice versa).
    CONSTRAINT gift_coupons_redemption_consistency    CHECK (
        (redeemed_by_user_id IS NULL AND redeemed_at IS NULL) OR
        (redeemed_by_user_id IS NOT NULL AND redeemed_at IS NOT NULL)
    )
);

-- [Partial index WHERE redeemed_at IS NULL]: The redemption endpoint queries:
--   SELECT * FROM gift_coupons WHERE code = $1 AND redeemed_at IS NULL
-- Partial index covers only unredeemed coupons — the active working set.
-- Redeemed coupons are never queried this way again.
CREATE INDEX idx_gift_coupons_unredeemed
    ON gift_coupons(created_at DESC) WHERE redeemed_at IS NULL;

-- [course_id index]: Admin query "which coupons are for this course?".
-- Partial (WHERE NOT NULL) skips generic (course-unspecific) coupons.
CREATE INDEX idx_gift_coupons_course_id
    ON gift_coupons(course_id) WHERE course_id IS NOT NULL;

-- [Function-based unique index]: Enforces case-insensitive code uniqueness.
-- UNIQUE CONSTRAINT on (code) would allow 'SAVE10' ≠ 'save10'. LOWER() collapses them.
CREATE UNIQUE INDEX gift_coupons_code_lower_unique ON gift_coupons (LOWER(code));

-- ---------------------------------------------------------------------------
-- 28. USER SESSIONS  (migrations 000002_add_user_sessions, 000003_add_user_sessions)
-- ---------------------------------------------------------------------------

-- [user_sessions — refresh token storage for JWT authentication]:
-- This project uses a two-token JWT scheme:
--   Access token  — short-lived (15 min), stateless, never stored in DB.
--                   Contains user_id and role, verified by signature only.
--   Refresh token — long-lived (7–30 days), stateful, hash stored here.
--                   Used solely to obtain a new access token.
--
-- [Why store refresh tokens in DB instead of relying on stateless JWT]:
-- Stateless JWTs cannot be revoked without a denylist. If a refresh token is
-- compromised (e.g., cookie leak), it must be invalidatable.
-- user_sessions enables:
--   1. Revoking a specific session (logout from one device)
--   2. Revoking all sessions (logout everywhere, password change)
--   3. Showing the user their active sessions ("where am I logged in")
--   4. Detecting suspicious activity (unfamiliar IP or user_agent)
--
-- [Hash, not the raw token]: POST /auth/refresh sends the raw token in a cookie.
-- The DB stores SHA-256(token) — if the DB leaks, the attacker cannot use
-- hashes directly as tokens (would require a SHA-256 preimage attack).

CREATE TABLE user_sessions (
    -- [gen_random_uuid()]: UUID v4 generated by PostgreSQL. Project rule
    -- (db-conventions.md): UUIDs are generated in the DB, not in Go.
    id              uuid        PRIMARY KEY DEFAULT gen_random_uuid(),

    -- [REFERENCES users(id) ON DELETE RESTRICT]: A user row cannot be physically
    -- deleted while their sessions exist. RESTRICT forces explicit session revocation
    -- before user deletion. (ON DELETE CASCADE would silently remove sessions on
    -- user delete, but the project uses soft-delete on users, making RESTRICT safer.)
    user_id         uuid        NOT NULL REFERENCES users(id) ON DELETE RESTRICT,

    -- [SHA-256 hash of the refresh token]: The raw token is never stored — only its hash.
    -- Verification flow: client sends raw token → hash(token) → look up
    --   WHERE refresh_hash = hash(token) AND revoked_at IS NULL.
    -- text instead of varchar — hash length depends on the algorithm; text is flexible.
    -- Covered by UNIQUE constraint below — O(log n) lookup, no full scan.
    refresh_hash    varchar(64)        NOT NULL,

    -- [HTTP User-Agent string]:
    -- Examples: "Mozilla/5.0 (Macintosh; ...)", "Dart/3.0 (dart:io)", "curl/8.1"
    -- Nullable — some clients (curl, API integrations) omit this header.
    -- Used only for the "active sessions" UI display — not for security decisions.
    -- (Easily spoofed by the client, so not a reliable identifier.)
    user_agent      varchar(255),

    -- [Client IP address, varchar(45)]:
    -- Length 45 = maximum possible IP address representation:
    --   IPv4:             max 15 chars  (255.255.255.255)
    --   IPv6:             max 39 chars  (2001:0db8:85a3:0000:0000:8a2e:0370:7334)
    --   IPv6-mapped IPv4: max 45 chars  (::ffff:255.255.255.255)
    -- varchar(45) instead of text — the length cap rejects clearly malformed values.
    -- Nullable — may be absent behind a proxy or in tests without a real request.
    ip              varchar(45),

    -- [Session expiry]:
    -- NOT NULL — sessions always have a bounded lifetime. Indefinite sessions are a
    -- security risk. timestamptz — stored as UTC, compared without timezone ambiguity.
    -- The CHECK expires_at > created_at (below) ensures the token is not already
    -- expired at creation time (e.g., due to a misconfigured TTL setting).
    expires_at      timestamptz NOT NULL,

    -- [Explicit revocation timestamp]:
    -- NULL     = session is active (token accepted if not expired).
    -- NOT NULL = session revoked; refresh token is no longer accepted.
    -- Revocation occurs on: logout, /auth/revoke-all, password change,
    -- admin blocking the user.
    -- The row is kept after revocation for audit and statistics
    -- (e.g., "how many sessions did a password change revoke?"). A cleanup job
    -- eventually purges old rows. CHECK revoked_at >= created_at (below) prevents
    -- impossible state where revocation precedes creation.
    revoked_at      timestamptz,

    -- [Row creation timestamp]:
    -- DEFAULT now() — set automatically by PostgreSQL on INSERT.
    -- NOT NULL — always present, no need to pass it explicitly from Go.
    -- Used in CHECK constraints (expires_at > created_at), in the cleanup job
    -- index, and for audit ("when did the user log in?").
    created_at      timestamptz NOT NULL DEFAULT now(),

    -- [CONSTRAINT: unique hash]:
    -- No two rows can share the same refresh_hash. Protects against:
    --   1. Hash collisions (astronomically rare, but defensive)
    --   2. Duplicate rows from concurrent INSERTs (race condition in auth service)
    -- Also creates a B-tree index used for the primary lookup:
    --   WHERE refresh_hash = $1  (O(log n) instead of O(n) full scan)
    CONSTRAINT user_sessions_refresh_hash_unique        UNIQUE (refresh_hash),

    -- [CONSTRAINT: expires_at > created_at]:
    -- Token must expire AFTER creation. Strict > (not >=):
    -- expires_at = created_at means the token is already expired at birth — rejected.
    -- Catches: zero-second TTL, arithmetic errors in the auth service.
    CONSTRAINT user_sessions_expires_after_created      CHECK (expires_at > created_at),

    -- [CONSTRAINT: revoked_at >= created_at]:
    -- A session cannot be revoked before it was created.
    -- OR revoked_at IS NULL — constraint allows NULL (not yet revoked).
    -- Catches: clock skew bugs, incorrect field assignment in the service layer.
    CONSTRAINT user_sessions_revoked_after_created      CHECK (revoked_at IS NULL OR revoked_at >= created_at),

    -- [token_version]: Rotated on every successful refresh. The client's stored version
    -- is compared on the next refresh — mismatch indicates a stale or replayed token.
    -- NOT NULL DEFAULT 1 — every session starts at version 1.
    token_version          integer     NOT NULL DEFAULT 1,

    -- [previous_refresh_hash]: Solves concurrent refresh race condition.
    -- During token rotation two simultaneous requests may both arrive with the old token.
    -- Without this field the second request is rejected (old token already replaced),
    -- forcing re-login. With previous_refresh_hash: accept the old hash briefly as fallback.
    -- Cleared when the new token is first used or when the session expires.
    previous_refresh_hash  varchar(64),

    -- [failed_attempt_count / locked_until]: Brute-force protection for the refresh endpoint.
    -- After N consecutive failures, locked_until is set. Requests are rejected until
    -- locked_until < now(). Resets to 0 on a successful refresh.
    failed_attempt_count   integer     NOT NULL DEFAULT 0,
    locked_until           timestamptz,

    -- [last_attempt_at]: Timestamp of the most recent refresh attempt (success or failure).
    -- Enables sliding-window rate limiting and activity audit.
    last_attempt_at        timestamptz,

    -- [revoke_reason]: Typed cause of revocation. NULL = session still active.
    -- Enables targeted session management — e.g., password_changed → revoke all other
    -- sessions while preserving the current one.
    revoke_reason           varchar(30),

    -- [revoked_by_user_id ON DELETE RESTRICT]: Audit trail — who revoked this session.
    -- RESTRICT: cannot delete the revoking user while they are recorded as having revoked
    -- another user's session (e.g., admin action). Preserves audit integrity.
    revoked_by_user_id      uuid        REFERENCES users(id) ON DELETE RESTRICT,

    -- [last_seen_ip + last_seen_at]: Security audit — where and when session was last active.
    -- Supports geolocation anomaly detection. Updated on each authenticated request.
    last_seen_ip            varchar(45),
    last_seen_at            timestamptz,

    CONSTRAINT user_sessions_revoke_reason_check
        CHECK (revoke_reason IN ('logout', 'password_changed', 'admin', 'suspicious_activity', 'token_expired')),
    CONSTRAINT user_sessions_token_version_positive CHECK (token_version > 0),
    CONSTRAINT user_sessions_failed_attempts_non_negative CHECK (failed_attempt_count >= 0)
);

-- ---------------------------------------------------------------------------
-- INDEXES — user_sessions
-- ---------------------------------------------------------------------------

-- [idx_user_sessions_refresh_hash_active — primary lookup]:
-- Query: POST /auth/refresh → WHERE refresh_hash = $1 AND revoked_at IS NULL
-- Partial index (WHERE revoked_at IS NULL):
--   - Covers only active sessions — exactly what this query needs.
--   - Revoked sessions are excluded from the index → index stays small and fast.
--   - Over time revoked sessions accumulate, but the index does not grow with them.
-- The UNIQUE constraint already created a B-tree on refresh_hash, but that is a
-- full index. This partial index is smaller and more efficient for active-only lookups.
CREATE INDEX idx_user_sessions_refresh_hash_active
    ON user_sessions(refresh_hash)
    WHERE revoked_at IS NULL;

-- [idx_user_sessions_user_id_active — sessions per user]:
-- Queries:
--   GET /users/me/sessions → WHERE user_id = $1 AND revoked_at IS NULL
--   POST /auth/revoke-all  → UPDATE ... WHERE user_id = $1 AND revoked_at IS NULL
-- Partial index (WHERE revoked_at IS NULL):
--   An active user typically has 1–5 active sessions (phone, laptop, tablet).
--   But may have hundreds of revoked sessions over the years. Partial index covers
--   only the active subset — the working set stays small regardless of history size.
CREATE INDEX idx_user_sessions_user_id_active
    ON user_sessions(user_id)
    WHERE revoked_at IS NULL;

-- [idx_user_sessions_expires_at — cleanup job]:
-- Query (cron worker or scheduled job):
--   DELETE FROM user_sessions WHERE expires_at < now() AND revoked_at IS NULL
-- Or for soft cleanup:
--   SELECT id FROM user_sessions WHERE expires_at < now() - interval '30 days'
-- Partial index (WHERE revoked_at IS NULL):
--   Already-revoked sessions do not need cleanup — they are historical records.
--   Partial index covers only rows that the cleanup job will ever touch.
-- Without this index, the cleanup job requires a full table scan (O(n)).
CREATE INDEX idx_user_sessions_expires_at
    ON user_sessions(expires_at)
    WHERE revoked_at IS NULL;

-- [Partial WHERE previous_refresh_hash IS NOT NULL]: Only set during token rotation
-- window (brief). Partial index = only rows where this lookup is actually needed.
CREATE INDEX idx_user_sessions_previous_refresh_hash
    ON user_sessions(previous_refresh_hash)
    WHERE previous_refresh_hash IS NOT NULL;
