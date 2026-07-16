package coursedomain

import (
	"context"
)

// Transactor executes a function within a database transaction.
type Transactor interface {
	InTransaction(ctx context.Context, fn func(ctx context.Context) error) error
}

// CourseRepository defines persistence operations for Course.
type CourseRepository interface {
	CreateCourse(ctx context.Context, course *Course) (*Course, error)
	PublishCourse(ctx context.Context, courseID string) error
	ArchiveCourse(ctx context.Context, courseID string) error
	DeleteCourse(ctx context.Context, courseID string) error
	UpdateCourse(ctx context.Context, course *Course) error
	GetAllPublishedCourses(ctx context.Context) ([]*Course, error)
	GetAllDraftCourses(ctx context.Context) ([]*Course, error)
	GetAllArchivedCourses(ctx context.Context) ([]*Course, error)
	GetAllCourses(ctx context.Context) ([]*Course, error)
	GetCourseByID(ctx context.Context, courseID string) (*Course, error)
	GetCourseBySlug(ctx context.Context, slug string) (*Course, error)
}

// Service defines the courses module's business logic operations.
type Service interface {
	ArchiveCourse(ctx context.Context, courseID string) error
	CreateCourse(ctx context.Context, req CreateCourseRequest) (string, error)
	DeleteCourse(ctx context.Context, courseID string) error
	GetCourseBySlug(ctx context.Context, slug string) (*Course, error)
	GetAllCourses(ctx context.Context, getType CourseStatus) (courseList []*Course, err error)
	PublishCourse(ctx context.Context, courseID string) error
	UpdateCourse(ctx context.Context, req UpdateCourseRequest) error
}
