package coursedomain

import (
	"context"
	"learnflow_backend/internal/shared/pagination"
)

// Transactor executes a function within a database transaction.
type Transactor interface {
	InTransaction(ctx context.Context, fn func(ctx context.Context) error) error
}

// CourseRepository defines persistence operations for Course.
type CourseRepository interface {
	CreateCourse(ctx context.Context, course *Course) (*Course, error)
	PublishCourse(ctx context.Context, courseID, userID string) error
	ArchiveCourse(ctx context.Context, courseID, userID string) error
	DeleteCourse(ctx context.Context, courseID, userID string) error
	UpdateCourse(ctx context.Context, course *Course, userID string) error
	GetAllPublishedCourses(ctx context.Context, params pagination.Params) ([]*Course, error)
	GetAllDraftCourses(ctx context.Context, params pagination.Params) ([]*Course, error)
	GetAllArchivedCourses(ctx context.Context, params pagination.Params) ([]*Course, error)
	GetAllCourses(ctx context.Context, params pagination.Params) ([]*Course, error)
	GetCourseByID(ctx context.Context, courseID string) (*Course, error)
	GetCourseBySlug(ctx context.Context, slug string) (*Course, error)
}

// Service defines the courses module's business logic operations.
type Service interface {
	ArchiveCourse(ctx context.Context, courseID, userID string) error
	CreateCourse(ctx context.Context, req CreateCourseRequest) (string, error)
	DeleteCourse(ctx context.Context, courseID, userID string) error
	GetCourseBySlug(ctx context.Context, slug string) (*Course, error)
	PublishCourse(ctx context.Context, courseID, userID string) error
	UpdateCourse(ctx context.Context, req UpdateCourseRequest, userID string) error
	GetAllCourses(ctx context.Context, getType CourseStatus, params pagination.Params) (courseList []*Course, err error)
}
