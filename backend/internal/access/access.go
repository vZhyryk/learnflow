package access

import (
	"context"
	"fmt"
	"learnflow_backend/internal/infrastructure/db"
)

type Checker struct {
	db db.QueryRunner
}

func New(db db.QueryRunner) *Checker { return &Checker{db: db} }

func (c *Checker) HasAccessCourse(ctx context.Context, userID, courseID string) (bool, error) {
	var exists bool
	err := c.db.QueryRow(ctx, HasAccessCourseQuery, userID, courseID).Scan(&exists)
	if err != nil {
		return false, fmt.Errorf("access.HasAccessCourse: %w", err)
	}
	return exists, nil
}

func (c *Checker) HasAccessContent(ctx context.Context, userID, contentID string) (bool, error) {
	var exists bool
	err := c.db.QueryRow(ctx, HasAccessContentQuery, userID, contentID).Scan(&exists)
	if err != nil {
		return false, fmt.Errorf("access.HasAccessContent: %w", err)
	}
	return exists, nil
}
