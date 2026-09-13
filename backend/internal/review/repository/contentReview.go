package reviewrepository

import (
	"context"
	"errors"
	"fmt"
	"learnflow_backend/internal/infrastructure/db"
	reviewdomain "learnflow_backend/internal/review/domain"
	"learnflow_backend/internal/shared/pagination"
	"learnflow_backend/internal/shared/repository"

	"github.com/jackc/pgx/v5"
)

// CreateContentReview persists a new content item review.
func (rep *Repository) CreateContentReview(ctx context.Context, contentReview *reviewdomain.ContentReview) (*reviewdomain.ContentReview, error) {
	created, err := scanContentReview(rep.QueryRunner(ctx).QueryRow(ctx, createContentReviewSQL, contentReview.ContentID, contentReview.UserID, contentReview.Rating, contentReview.Comment))
	if db.IsUniqueViolation(err, contentReviewsUserContentUniqueConstraint) {
		return nil, reviewdomain.ErrAlreadyReviewed
	}
	if err != nil {
		return nil, fmt.Errorf("repository.CreateContentReview: %w", err)
	}

	return created, nil
}

// UpdateContentReview updates an existing content review's rating and comment.
func (rep *Repository) UpdateContentReview(ctx context.Context, contentItem *reviewdomain.ContentReview) error {
	tag, err := rep.QueryRunner(ctx).Exec(ctx, updateContentReviewSQL, contentItem.ID, contentItem.Rating, contentItem.Comment)
	if err != nil {
		return fmt.Errorf("repository.UpdateContentReview: %w", err)
	}

	if tag.RowsAffected() == 0 {
		return reviewdomain.ErrReviewNotFound
	}

	return nil
}

// DeleteContentReview soft-deletes a content review.
func (rep *Repository) DeleteContentReview(ctx context.Context, reviewID string) error {
	tag, err := rep.QueryRunner(ctx).Exec(ctx, deleteContentReviewSQL, reviewID)
	if err != nil {
		return fmt.Errorf("repository.DeleteContentReview: %w", err)
	}

	if tag.RowsAffected() == 0 {
		return reviewdomain.ErrReviewNotFound
	}

	return nil
}

// GetContentReviewList returns a paginated list of non-deleted reviews for a content item.
func (rep *Repository) GetContentReviewList(ctx context.Context, params pagination.Params, contentID string) ([]*reviewdomain.ContentReview, error) {
	args := make([]any, 1)
	args[0] = contentID
	return repository.GetAndParseListWithArgs(ctx, &rep.BaseRepository, getContentReviewByContentIDSQL, "GetContentReviewList", params, scanContentReview, args)
}

// GetContentReviewByID retrieves a non-deleted content review by ID.
func (rep *Repository) GetContentReviewByID(ctx context.Context, reviewID string) (*reviewdomain.ContentReview, error) {
	contentReview, err := scanContentReview(rep.QueryRunner(ctx).QueryRow(ctx, getContentReviewByIDSQL, reviewID))
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, reviewdomain.ErrReviewNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("repository.GetContentReviewByID: %w", err)
	}

	return contentReview, nil
}

// GetContentReviewByUserAndContentID retrieves a user's non-deleted review for a content item, if any.
func (rep *Repository) GetContentReviewByUserAndContentID(ctx context.Context, userID, contentID string) (*reviewdomain.ContentReview, error) {
	contentReview, err := scanContentReview(rep.QueryRunner(ctx).QueryRow(ctx, getContentReviewByUserAndContentIDSQL, contentID, userID))
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, reviewdomain.ErrReviewNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("repository.GetContentReviewByUserAndContentID: %w", err)
	}

	return contentReview, nil
}

func (rep *Repository) GetContentReviewStats(ctx context.Context, contentID string) (rating float64, count int, err error) {
	err = rep.QueryRunner(ctx).QueryRow(ctx, getContentReviewStats, contentID).Scan(&rating, &count)
	if err != nil {
		return 0, 0, fmt.Errorf("repository.GetContentReviewStats: %w", err)
	}

	return rating, count, nil
}
