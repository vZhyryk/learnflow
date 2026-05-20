-- 000001 rollback: drop entire schema in reverse dependency order.

DROP VIEW IF EXISTS monthly_pnl;
DROP VIEW IF EXISTS monthly_revenue;

DROP TABLE IF EXISTS gift_coupons;
DROP TABLE IF EXISTS articles;
DROP TABLE IF EXISTS account_recovery_tokens;
DROP TABLE IF EXISTS support_messages;
DROP TABLE IF EXISTS support_chats;
DROP TABLE IF EXISTS admin_actions;
DROP TABLE IF EXISTS failed_jobs;
DROP TABLE IF EXISTS event_outbox;
DROP TABLE IF EXISTS activity_log;
DROP TABLE IF EXISTS announcements;
DROP TABLE IF EXISTS notifications;
DROP TABLE IF EXISTS notification_preferences;
DROP TABLE IF EXISTS consultation_outcomes;
DROP TABLE IF EXISTS consultation_bookings;
DROP TABLE IF EXISTS consultation_briefs;
DROP TABLE IF EXISTS consultation_slots;
DROP TABLE IF EXISTS user_notes;
DROP TABLE IF EXISTS user_content_access;
DROP TABLE IF EXISTS user_document_engagement;
DROP TABLE IF EXISTS user_video_engagement;
DROP TABLE IF EXISTS user_content_progress;
DROP TABLE IF EXISTS user_course_progress;
DROP TABLE IF EXISTS user_course_access;
DROP TABLE IF EXISTS expenses;
DROP TABLE IF EXISTS campaigns;
DROP TABLE IF EXISTS refunds;
DROP TABLE IF EXISTS payment_line_items;
DROP TABLE IF EXISTS payments;
DROP TABLE IF EXISTS content_reviews;
DROP TABLE IF EXISTS course_reviews;
DROP TABLE IF EXISTS course_content_items;
DROP TABLE IF EXISTS content_items;
DROP TABLE IF EXISTS courses;
DROP TABLE IF EXISTS email_change_tokens;
DROP TABLE IF EXISTS password_reset_tokens;
DROP TABLE IF EXISTS email_verification_tokens;
DROP TABLE IF EXISTS user_profiles;
DROP TABLE IF EXISTS users;

DROP EXTENSION IF EXISTS pg_stat_statements;
