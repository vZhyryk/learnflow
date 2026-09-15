package reviewrepository

import (
	reviewdomain "learnflow_backend/internal/review/domain"
	"learnflow_backend/internal/shared/repository"
	"learnflow_backend/internal/shared/testutil"
	"time"
)

func newTestRepo(runner *testutil.MockQueryRunner) *Repository {
	return &Repository{repository.BaseRepository{DB: runner}}
}

func fakeCourseReview(now time.Time) *reviewdomain.CourseReview {
	comment := "comment"

	return &reviewdomain.CourseReview{
		ID:              "course-item-123",
		CourseID:        "course_id",
		UserID:          "user_id",
		Rating:          1,
		Comment:         &comment,
		CreatedAt:       now,
		UpdatedAt:       now,
		DeletedAt:       nil,
		DeletedByUserID: nil,
	}
}

func fakeCourseReviewScan(item *reviewdomain.CourseReview) func(dest ...any) error {
	return func(dest ...any) error {
		*testutil.CastStr(dest[0], 0) = item.ID
		*testutil.CastStr(dest[1], 1) = item.CourseID
		*testutil.CastStr(dest[2], 2) = item.UserID
		*testutil.CastInt(dest[3], 3) = item.Rating
		*testutil.CastPtrStr(dest[4], 4) = item.Comment
		*testutil.CastTime(dest[5], 5) = item.CreatedAt
		*testutil.CastTime(dest[6], 6) = item.UpdatedAt
		*testutil.CastPtrTime(dest[7], 7) = item.DeletedAt
		*testutil.CastPtrStr(dest[8], 8) = item.DeletedByUserID
		return nil
	}
}

func fakeContentReview(now time.Time) *reviewdomain.ContentReview {
	comment := "comment"

	return &reviewdomain.ContentReview{
		ID:              "content-item-123",
		ContentID:       "content_id",
		UserID:          "user_id",
		Rating:          1,
		Comment:         &comment,
		CreatedAt:       now,
		UpdatedAt:       now,
		DeletedAt:       nil,
		DeletedByUserID: nil,
	}
}

func fakeContentReviewScan(item *reviewdomain.ContentReview) func(dest ...any) error {
	return func(dest ...any) error {
		*testutil.CastStr(dest[0], 0) = item.ID
		*testutil.CastStr(dest[1], 1) = item.ContentID
		*testutil.CastStr(dest[2], 2) = item.UserID
		*testutil.CastInt(dest[3], 3) = item.Rating
		*testutil.CastPtrStr(dest[4], 4) = item.Comment
		*testutil.CastTime(dest[5], 5) = item.CreatedAt
		*testutil.CastTime(dest[6], 6) = item.UpdatedAt
		*testutil.CastPtrTime(dest[7], 7) = item.DeletedAt
		*testutil.CastPtrStr(dest[8], 8) = item.DeletedByUserID
		return nil
	}
}

func fakeArticleReview(now time.Time) *reviewdomain.ArticleReview {
	comment := "comment"

	return &reviewdomain.ArticleReview{
		ID:              "article-review-123",
		ArticleID:       "article_id",
		UserID:          "user_id",
		Rating:          1,
		Comment:         &comment,
		CreatedAt:       now,
		UpdatedAt:       now,
		DeletedAt:       nil,
		DeletedByUserID: nil,
	}
}

func fakeArticleReviewScan(item *reviewdomain.ArticleReview) func(dest ...any) error {
	return func(dest ...any) error {
		*testutil.CastStr(dest[0], 0) = item.ID
		*testutil.CastStr(dest[1], 1) = item.ArticleID
		*testutil.CastStr(dest[2], 2) = item.UserID
		*testutil.CastInt(dest[3], 3) = item.Rating
		*testutil.CastPtrStr(dest[4], 4) = item.Comment
		*testutil.CastTime(dest[5], 5) = item.CreatedAt
		*testutil.CastTime(dest[6], 6) = item.UpdatedAt
		*testutil.CastPtrTime(dest[7], 7) = item.DeletedAt
		*testutil.CastPtrStr(dest[8], 8) = item.DeletedByUserID
		return nil
	}
}
