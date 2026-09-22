package testutil

import (
	"context"
	"testing"
	"time"

	"learnflow_backend/internal/infrastructure/db"
)

const (
	defaultAccessType   = "admin_granted"
	defaultAccessStatus = "active"
)

const grantCourseAccessSQL = `
	INSERT INTO user_course_access (user_id, course_id, access_type, status, created_at, granted_at, expires_at)
	VALUES ($1, $2, $3, $4, COALESCE($5::timestamptz, now()), COALESCE($6::timestamptz, now()), $7)`

const grantContentAccessSQL = `
	INSERT INTO user_content_access (user_id, content_item_id, access_type, status, created_at, granted_at, expires_at)
	VALUES ($1, $2, $3, $4, COALESCE($5::timestamptz, now()), COALESCE($6::timestamptz, now()), $7)`

const linkContentItemToCourseSQL = `
	INSERT INTO course_content_items (course_id, content_item_id, position, is_required)
	VALUES ($1, $2, $3, $4)`

// AccessGrant overrides the defaults of GrantCourseAccess/GrantContentAccess. The zero
// value is an active admin_granted access starting now with no expiry. Set all three
// times together for past-dated rows: the schema requires created_at <= granted_at < expires_at.
type AccessGrant struct {
	AccessType string
	Status     string
	CreatedAt  *time.Time
	GrantedAt  *time.Time
	ExpiresAt  *time.Time
}

func (g AccessGrant) accessType() string {
	if g.AccessType == "" {
		return defaultAccessType
	}
	return g.AccessType
}

func (g AccessGrant) status() string {
	if g.Status == "" {
		return defaultAccessStatus
	}
	return g.Status
}

// GrantCourseAccess inserts a user_course_access row for userID and courseID.
func GrantCourseAccess(t *testing.T, q db.QueryRunner, userID, courseID string, g AccessGrant) {
	t.Helper()

	execFixture(t, q, "GrantCourseAccess", grantCourseAccessSQL,
		userID, courseID, g.accessType(), g.status(), g.CreatedAt, g.GrantedAt, g.ExpiresAt)
}

// GrantContentAccess inserts a user_content_access row for userID and contentItemID.
func GrantContentAccess(t *testing.T, q db.QueryRunner, userID, contentItemID string, g AccessGrant) {
	t.Helper()

	execFixture(t, q, "GrantContentAccess", grantContentAccessSQL,
		userID, contentItemID, g.accessType(), g.status(), g.CreatedAt, g.GrantedAt, g.ExpiresAt)
}

// LinkContentItemToCourse inserts a course_content_items row at the given position.
func LinkContentItemToCourse(t *testing.T, q db.QueryRunner, courseID, contentItemID string, position int, required bool) {
	t.Helper()

	execFixture(t, q, "LinkContentItemToCourse", linkContentItemToCourseSQL, courseID, contentItemID, position, required)
}

func execFixture(t *testing.T, q db.QueryRunner, name, sql string, args ...any) {
	t.Helper()

	if _, err := q.Exec(context.Background(), sql, args...); err != nil {
		t.Fatalf("testutil.%s: %v", name, err)
	}
}
