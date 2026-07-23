package courseservice

import (
	"context"
	coursedomain "learnflow_backend/internal/courses/domain"
	"learnflow_backend/internal/events"
	"learnflow_backend/internal/shared/pagination"
	"learnflow_backend/internal/shared/testutil"
)

// mockCourseRepoRepo implements coursedomain.CourseRepository via function fields.
type mockCourseRepoRepo struct {
	createCourse           func(ctx context.Context, course *coursedomain.Course) (*coursedomain.Course, error)
	publishCourse          func(ctx context.Context, courseID string) error
	archiveCourse          func(ctx context.Context, courseID string) error
	deleteCourse           func(ctx context.Context, courseID string) error
	updateCourse           func(ctx context.Context, course *coursedomain.Course) error
	getAllPublishedCourses func(ctx context.Context, params pagination.Params) ([]*coursedomain.Course, error)
	getAllDraftCourses     func(ctx context.Context, params pagination.Params) ([]*coursedomain.Course, error)
	getAllArchivedCourses  func(ctx context.Context, params pagination.Params) ([]*coursedomain.Course, error)
	getAllCourses          func(ctx context.Context, params pagination.Params) ([]*coursedomain.Course, error)
	getCourseByID          func(ctx context.Context, courseID string) (*coursedomain.Course, error)
	getCourseBySlug        func(ctx context.Context, slug string) (*coursedomain.Course, error)
}

func (m *mockCourseRepoRepo) CreateCourse(ctx context.Context, course *coursedomain.Course) (*coursedomain.Course, error) {
	if m.createCourse == nil {
		panic("mockCourseRepo.createCourse not set")
	}

	return m.createCourse(ctx, course)
}
func (m *mockCourseRepoRepo) PublishCourse(ctx context.Context, courseID string) error {
	if m.publishCourse == nil {
		panic("mockCourseRepo.publishCourse not set")
	}

	return m.publishCourse(ctx, courseID)
}
func (m *mockCourseRepoRepo) ArchiveCourse(ctx context.Context, courseID string) error {
	if m.archiveCourse == nil {
		panic("mockCourseRepo.archiveCourse not set")
	}

	return m.archiveCourse(ctx, courseID)
}
func (m *mockCourseRepoRepo) DeleteCourse(ctx context.Context, courseID string) error {
	if m.deleteCourse == nil {
		panic("mockCourseRepo.deleteCourse not set")
	}

	return m.deleteCourse(ctx, courseID)
}
func (m *mockCourseRepoRepo) UpdateCourse(ctx context.Context, course *coursedomain.Course) error {
	if m.updateCourse == nil {
		panic("mockCourseRepo.updateCourse not set")
	}

	return m.updateCourse(ctx, course)
}
func (m *mockCourseRepoRepo) GetAllPublishedCourses(ctx context.Context, params pagination.Params) ([]*coursedomain.Course, error) {
	if m.getAllPublishedCourses == nil {
		panic("mockCourseRepo.getAllPublishedCourses not set")
	}

	return m.getAllPublishedCourses(ctx, params)
}
func (m *mockCourseRepoRepo) GetAllDraftCourses(ctx context.Context, params pagination.Params) ([]*coursedomain.Course, error) {
	if m.getAllDraftCourses == nil {
		panic("mockCourseRepo.getAllDraftCourses not set")
	}

	return m.getAllDraftCourses(ctx, params)
}
func (m *mockCourseRepoRepo) GetAllArchivedCourses(ctx context.Context, params pagination.Params) ([]*coursedomain.Course, error) {
	if m.getAllArchivedCourses == nil {
		panic("mockCourseRepo.getAllArchivedCourses not set")
	}

	return m.getAllArchivedCourses(ctx, params)
}
func (m *mockCourseRepoRepo) GetAllCourses(ctx context.Context, params pagination.Params) ([]*coursedomain.Course, error) {
	if m.getAllCourses == nil {
		panic("mockCourseRepo.getAllCourses not set")
	}

	return m.getAllCourses(ctx, params)
}
func (m *mockCourseRepoRepo) GetCourseByID(ctx context.Context, courseID string) (*coursedomain.Course, error) {
	if m.getCourseByID == nil {
		panic("mockCourseRepo.getCourseByID not set")
	}

	return m.getCourseByID(ctx, courseID)
}
func (m *mockCourseRepoRepo) GetCourseBySlug(ctx context.Context, slug string) (*coursedomain.Course, error) {
	if m.getCourseBySlug == nil {
		panic("mockCourseRepo.getCourseBySlug not set")
	}

	return m.getCourseBySlug(ctx, slug)
}

func newTestService(repo *mockCourseRepoRepo, outbox *events.OutboxWriter) *Service {
	return New(repo, &testutil.NoopTransactor{}, outbox)
}

func AlwaysError(_ context.Context, _ string) (*coursedomain.Course, error) {
	return nil, testutil.ErrDBUnexpected
}

func AlwaysFailsErr(_ context.Context, _ *coursedomain.Course) error {
	return testutil.ErrDBUnexpected
}

func alwaysSucceedsUpdate(_ context.Context, _ *coursedomain.Course) error {
	return nil
}
