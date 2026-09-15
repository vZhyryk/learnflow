package reviewservice

import (
	"context"
	reviewdomain "learnflow_backend/internal/review/domain"
	"learnflow_backend/internal/shared/pagination"
	"learnflow_backend/internal/shared/testutil"
)

type mockReviewRepo struct {
	createCourseReview                 func(ctx context.Context, courseReview *reviewdomain.CourseReview) (*reviewdomain.CourseReview, error)
	updateCourseReview                 func(ctx context.Context, courseReview *reviewdomain.CourseReview) error
	deleteCourseReview                 func(ctx context.Context, reviewID, userID string) error
	getCourseReviewList                func(ctx context.Context, params pagination.Params, courseID string, filter reviewdomain.ReviewFilter) ([]*reviewdomain.CourseReview, error)
	getCourseReviewByID                func(ctx context.Context, reviewID string) (*reviewdomain.CourseReview, error)
	getCourseReviewByUserAndCourseID   func(ctx context.Context, userID, courseID string) (*reviewdomain.CourseReview, error)
	getCourseReviewStats               func(ctx context.Context, courseID string) (rating float64, count int, err error)
	createContentReview                func(ctx context.Context, contentReview *reviewdomain.ContentReview) (*reviewdomain.ContentReview, error)
	updateContentReview                func(ctx context.Context, contentReview *reviewdomain.ContentReview) error
	deleteContentReview                func(ctx context.Context, reviewID, userID string) error
	getContentReviewList               func(ctx context.Context, params pagination.Params, contentID string, filter reviewdomain.ReviewFilter) ([]*reviewdomain.ContentReview, error)
	getContentReviewByID               func(ctx context.Context, reviewID string) (*reviewdomain.ContentReview, error)
	getContentReviewByUserAndContentID func(ctx context.Context, userID, contentID string) (*reviewdomain.ContentReview, error)
	getContentReviewStats              func(ctx context.Context, contentID string) (rating float64, count int, err error)
}

type mockAccessChecker struct {
	hasAccessCourse  func(ctx context.Context, userID, courseID string) (bool, error)
	hasAccessContent func(ctx context.Context, userID, contentID string) (bool, error)
}

func (m *mockReviewRepo) CreateCourseReview(ctx context.Context, courseReview *reviewdomain.CourseReview) (*reviewdomain.CourseReview, error) {
	if m.createCourseReview == nil {
		panic("mockReviewRepo.CreateCourseReview not set")
	}

	return m.createCourseReview(ctx, courseReview)
}
func (m *mockReviewRepo) UpdateCourseReview(ctx context.Context, courseReview *reviewdomain.CourseReview) error {
	if m.updateCourseReview == nil {
		panic("mockReviewRepo.UpdateCourseReview not set")
	}

	return m.updateCourseReview(ctx, courseReview)
}
func (m *mockReviewRepo) DeleteCourseReview(ctx context.Context, reviewID, userID string) error {
	if m.deleteCourseReview == nil {
		panic("mockReviewRepo.DeleteCourseReview not set")
	}

	return m.deleteCourseReview(ctx, reviewID, userID)
}
func (m *mockReviewRepo) GetCourseReviewList(ctx context.Context, params pagination.Params, courseID string, filter reviewdomain.ReviewFilter) ([]*reviewdomain.CourseReview, error) {
	if m.getCourseReviewList == nil {
		panic("mockReviewRepo.GetCourseReviewList not set")
	}

	return m.getCourseReviewList(ctx, params, courseID, filter)
}
func (m *mockReviewRepo) GetCourseReviewByID(ctx context.Context, reviewID string) (*reviewdomain.CourseReview, error) {
	if m.getCourseReviewByID == nil {
		panic("mockReviewRepo.GetCourseReviewByID not set")
	}

	return m.getCourseReviewByID(ctx, reviewID)
}
func (m *mockReviewRepo) GetCourseReviewByUserAndCourseID(ctx context.Context, userID, courseID string) (*reviewdomain.CourseReview, error) {
	if m.getCourseReviewByUserAndCourseID == nil {
		panic("mockReviewRepo.GetCourseReviewByUserAndCourseID not set")
	}

	return m.getCourseReviewByUserAndCourseID(ctx, userID, courseID)
}
func (m *mockReviewRepo) GetCourseReviewStats(ctx context.Context, courseID string) (rating float64, count int, err error) {
	if m.getCourseReviewStats == nil {
		panic("mockReviewRepo.GetCourseReviewStats not set")
	}

	return m.getCourseReviewStats(ctx, courseID)
}

func (m *mockReviewRepo) CreateContentReview(ctx context.Context, contentReview *reviewdomain.ContentReview) (*reviewdomain.ContentReview, error) {
	if m.createContentReview == nil {
		panic("mockReviewRepo.CreateContentReview not set")
	}

	return m.createContentReview(ctx, contentReview)
}
func (m *mockReviewRepo) UpdateContentReview(ctx context.Context, contentReview *reviewdomain.ContentReview) error {
	if m.updateContentReview == nil {
		panic("mockReviewRepo.UpdateContentReview not set")
	}

	return m.updateContentReview(ctx, contentReview)
}
func (m *mockReviewRepo) DeleteContentReview(ctx context.Context, reviewID, userID string) error {
	if m.deleteContentReview == nil {
		panic("mockReviewRepo.DeleteContentReview not set")
	}

	return m.deleteContentReview(ctx, reviewID, userID)
}
func (m *mockReviewRepo) GetContentReviewList(ctx context.Context, params pagination.Params, contentID string, filter reviewdomain.ReviewFilter) ([]*reviewdomain.ContentReview, error) {
	if m.getContentReviewList == nil {
		panic("mockReviewRepo.GetContentReviewList not set")
	}

	return m.getContentReviewList(ctx, params, contentID, filter)
}
func (m *mockReviewRepo) GetContentReviewByID(ctx context.Context, reviewID string) (*reviewdomain.ContentReview, error) {
	if m.getContentReviewByID == nil {
		panic("mockReviewRepo.GetContentReviewByID not set")
	}

	return m.getContentReviewByID(ctx, reviewID)
}
func (m *mockReviewRepo) GetContentReviewByUserAndContentID(ctx context.Context, userID, contentID string) (*reviewdomain.ContentReview, error) {
	if m.getContentReviewByUserAndContentID == nil {
		panic("mockReviewRepo.GetContentReviewByUserAndContentID not set")
	}

	return m.getContentReviewByUserAndContentID(ctx, userID, contentID)
}
func (m *mockReviewRepo) GetContentReviewStats(ctx context.Context, contentID string) (rating float64, count int, err error) {
	if m.getContentReviewStats == nil {
		panic("mockReviewRepo.GetContentReviewStats not set")
	}

	return m.getContentReviewStats(ctx, contentID)
}

func (m *mockAccessChecker) HasAccessCourse(ctx context.Context, userID, courseID string) (bool, error) {
	if m.hasAccessCourse == nil {
		panic("mockAccessChecker.HasAccessCourse not set")
	}

	return m.hasAccessCourse(ctx, userID, courseID)
}

func (m *mockAccessChecker) HasAccessContent(ctx context.Context, userID, contentID string) (bool, error) {
	if m.hasAccessContent == nil {
		panic("mockAccessChecker.HasAccessContent not set")
	}

	return m.hasAccessContent(ctx, userID, contentID)
}

func newTestService(courseRepo *mockReviewRepo, contentRepo *mockReviewRepo, accessChecker *mockAccessChecker) *Service {
	return New(courseRepo, contentRepo, &testutil.NoopTransactor{}, accessChecker)
}
