package reviewdomain

import (
	"context"
	"learnflow_backend/internal/shared/pagination"
)

// Transactor executes a function within a database transaction.
type Transactor interface {
	InTransaction(ctx context.Context, fn func(ctx context.Context) error) error
}

// AccessChecker checks whether a user has access to a course or content item.
type AccessChecker interface {
	HasAccessCourse(ctx context.Context, userID, courseID string) (bool, error)
	HasAccessContent(ctx context.Context, userID, contentID string) (bool, error)
}

// CourseReviewRepository defines persistence operations for CourseReview.
type CourseReviewRepository interface {
	CreateCourseReview(ctx context.Context, courseReview *CourseReview) (*CourseReview, error)
	UpdateCourseReview(ctx context.Context, courseReview *CourseReview) error
	DeleteCourseReview(ctx context.Context, reviewID, userID string) error
	GetCourseReviewList(ctx context.Context, params pagination.Params, courseID string, filter ReviewFilter) ([]*CourseReview, error)
	GetCourseReviewByID(ctx context.Context, reviewID string) (*CourseReview, error)
	GetCourseReviewByUserAndCourseID(ctx context.Context, userID, courseID string) (*CourseReview, error)
	GetCourseReviewStats(ctx context.Context, courseID string) (rating float64, count int, err error)
}

// ContentReviewRepository defines persistence operations for ContentReview.
type ContentReviewRepository interface {
	CreateContentReview(ctx context.Context, contentReview *ContentReview) (*ContentReview, error)
	UpdateContentReview(ctx context.Context, contentReview *ContentReview) error
	DeleteContentReview(ctx context.Context, reviewID, userID string) error
	GetContentReviewList(ctx context.Context, params pagination.Params, contentID string, filter ReviewFilter) ([]*ContentReview, error)
	GetContentReviewByID(ctx context.Context, reviewID string) (*ContentReview, error)
	GetContentReviewByUserAndContentID(ctx context.Context, userID, contentID string) (*ContentReview, error)
	GetContentReviewStats(ctx context.Context, contentID string) (rating float64, count int, err error)
}

type ArticleReviewRepository interface {
	CreateArticleReview(ctx context.Context, articleReview *ArticleReview) (*ArticleReview, error)
	UpdateArticleReview(ctx context.Context, articleReview *ArticleReview) error
	DeleteArticleReview(ctx context.Context, reviewID, userID string) error
	GetArticleReviewList(ctx context.Context, params pagination.Params, articleID string, filter ReviewFilter) ([]*ArticleReview, error)
	GetArticleReviewByID(ctx context.Context, reviewID string) (*ArticleReview, error)
	GetArticleReviewByUserAndArticleID(ctx context.Context, userID, articleID string) (*ArticleReview, error)
	GetArticleReviewStats(ctx context.Context, articleID string) (rating float64, count int, err error)
}

// Service defines the review module's business logic operations.
type Service interface {
	CreateCourseReview(ctx context.Context, req CreateCourseReviewRequest) error
	CreateContentReview(ctx context.Context, req CreateContentReviewRequest) error
	CreateArticleReview(ctx context.Context, req CreateArticleReviewRequest) error
	CreateCourseReviewAdmin(ctx context.Context, req CreateCourseReviewRequest) error
	CreateContentReviewAdmin(ctx context.Context, req CreateContentReviewRequest) error
	CreateArticleReviewAdmin(ctx context.Context, req CreateArticleReviewRequest) error

	UpdateCourseReview(ctx context.Context, req UpdateCourseReviewRequest) error
	UpdateContentReview(ctx context.Context, req UpdateContentReviewRequest) error
	UpdateArticleReview(ctx context.Context, req UpdateArticleReviewRequest) error
	UpdateCourseReviewAdmin(ctx context.Context, req UpdateCourseReviewRequest) error
	UpdateContentReviewAdmin(ctx context.Context, req UpdateContentReviewRequest) error
	UpdateArticleReviewAdmin(ctx context.Context, req UpdateArticleReviewRequest) error

	GetCourseReviews(ctx context.Context, params pagination.Params, courseID string, filter ReviewFilter) ([]*CourseReview, error)
	GetContentReviews(ctx context.Context, params pagination.Params, contentID string, filter ReviewFilter) ([]*ContentReview, error)
	GetArticleReviews(ctx context.Context, params pagination.Params, articleID string, filter ReviewFilter) ([]*ArticleReview, error)
	GetContentReviewStats(ctx context.Context, contentID string) (rating float64, count int, err error)
	GetCourseReviewStats(ctx context.Context, courseID string) (rating float64, count int, err error)
	GetArticleReviewStats(ctx context.Context, articleID string) (rating float64, count int, err error)

	DeleteCourseReview(ctx context.Context, reviewID, userID string) error
	DeleteContentReview(ctx context.Context, reviewID, userID string) error
	DeleteArticleReview(ctx context.Context, articleID, userID string) error
	DeleteContentReviewAdmin(ctx context.Context, reviewID, userID string) error
	DeleteCourseReviewAdmin(ctx context.Context, reviewID, userID string) error
	DeleteArticleReviewAdmin(ctx context.Context, articleID, userID string) error
}
