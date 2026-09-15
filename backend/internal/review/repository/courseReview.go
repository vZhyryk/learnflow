package reviewrepository

import (
	"context"
	"errors"
	"fmt"
	"learnflow_backend/internal/infrastructure/db"
	reviewdomain "learnflow_backend/internal/review/domain"
	"learnflow_backend/internal/shared/pagination"
	"learnflow_backend/internal/shared/repository"
	"strings"

	"github.com/jackc/pgx/v5"
)

// CreateCourseReview persists a new course review.
func (rep *Repository) CreateCourseReview(ctx context.Context, courseReview *reviewdomain.CourseReview) (*reviewdomain.CourseReview, error) {
	created, err := scanCourseReview(rep.QueryRunner(ctx).QueryRow(ctx, createCourseReviewSQL, courseReview.CourseID, courseReview.UserID, courseReview.Rating, courseReview.Comment))
	if db.IsUniqueViolation(err, courseReviewsUserCourseUniqueConstraint) {
		return nil, reviewdomain.ErrAlreadyReviewed
	}
	if err != nil {
		return nil, fmt.Errorf("repository.CreateCourseReview: %w", err)
	}

	return created, nil
}

// UpdateCourseReview updates an existing course review's rating and comment.
func (rep *Repository) UpdateCourseReview(ctx context.Context, contentItem *reviewdomain.CourseReview) error {
	tag, err := rep.QueryRunner(ctx).Exec(ctx, updateCourseReviewSQL, contentItem.ID, contentItem.Rating, contentItem.Comment)
	if err != nil {
		return fmt.Errorf("repository.UpdateCourseReview: %w", err)
	}

	if tag.RowsAffected() == 0 {
		return reviewdomain.ErrReviewNotFound
	}

	return nil
}

// DeleteCourseReview soft-deletes a course review.
func (rep *Repository) DeleteCourseReview(ctx context.Context, reviewID, userID string) error {
	tag, err := rep.QueryRunner(ctx).Exec(ctx, deleteCourseReviewSQL, reviewID, userID)
	if err != nil {
		return fmt.Errorf("repository.DeleteCourseReview: %w", err)
	}

	if tag.RowsAffected() == 0 {
		return reviewdomain.ErrReviewNotFound
	}

	return nil
}

// GetCourseReviewList returns a paginated list of non-deleted reviews for a course.
func (rep *Repository) GetCourseReviewList(ctx context.Context, params pagination.Params, courseID string, filter reviewdomain.ReviewFilter) ([]*reviewdomain.CourseReview, error) {
	args := make([]any, 1)
	args[0] = courseID
	var query string = getCourseReviewByCourseIDSQL
	if filter.IsUsed() {
		query = query[:strings.Index(query, filterPlace)] + rep.GenerateFilterQuery(filter.Op)
		args = append(args, filter.Rating)
	} else {
		query = strings.Replace(query, filterPlace, "", 1)
	}
	return repository.GetAndParseListWithArgs(ctx, &rep.BaseRepository, query, "GetCourseReviewList", params, scanCourseReview, args)
}

// GetCourseReviewByID retrieves a non-deleted course review by ID.
func (rep *Repository) GetCourseReviewByID(ctx context.Context, reviewID string) (*reviewdomain.CourseReview, error) {
	courseReview, err := scanCourseReview(rep.QueryRunner(ctx).QueryRow(ctx, getCourseReviewByIDSQL, reviewID))
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, reviewdomain.ErrReviewNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("repository.GetCourseReviewByID: %w", err)
	}

	return courseReview, nil
}

// GetCourseReviewByUserAndCourseID retrieves a user's non-deleted review for a course, if any.
func (rep *Repository) GetCourseReviewByUserAndCourseID(ctx context.Context, userID, courseID string) (*reviewdomain.CourseReview, error) {
	courseReview, err := scanCourseReview(rep.QueryRunner(ctx).QueryRow(ctx, getCourseReviewByUserAndCourseIDSQL, courseID, userID))
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, reviewdomain.ErrReviewNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("repository.GetCourseReviewByUserAndCourseID: %w", err)
	}

	return courseReview, nil
}

// GetCourseReviewStats returns the average rating and review count for a course.
func (rep *Repository) GetCourseReviewStats(ctx context.Context, courseID string) (rating float64, count int, err error) {
	err = rep.QueryRunner(ctx).QueryRow(ctx, getCourseReviewStats, courseID).Scan(&rating, &count)
	if err != nil {
		return 0, 0, fmt.Errorf("repository.GetCourseReviewStats: %w", err)
	}

	return rating, count, nil
}
