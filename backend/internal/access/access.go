package access

import (
	"context"
	"fmt"
	"learnflow_backend/internal/infrastructure/db"
)

// Checker checks whether a user has access to a course or content item.
type Checker struct {
	db db.QueryRunner
}

// New returns a new Checker.
func New(dbConn db.QueryRunner) *Checker { return &Checker{db: dbConn} }

// HasAccessCourse reports whether the user has access to the given course.
func (c *Checker) HasAccessCourse(ctx context.Context, userID, courseID string) (bool, error) {
	var exists bool
	err := c.db.QueryRow(ctx, HasAccessCourseQuery, userID, courseID).Scan(&exists)
	if err != nil {
		return false, fmt.Errorf("access.HasAccessCourse: %w", err)
	}
	return exists, nil
}

// HasAccessContent reports whether the user has access to the given content item.
func (c *Checker) HasAccessContent(ctx context.Context, userID, contentID string) (bool, error) {
	var exists bool
	err := c.db.QueryRow(ctx, HasAccessContentQuery, userID, contentID).Scan(&exists)
	if err != nil {
		return false, fmt.Errorf("access.HasAccessContent: %w", err)
	}
	return exists, nil
}
