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

// CreateArticleReview persists a new Article review.
func (rep *Repository) CreateArticleReview(ctx context.Context, articleReview *reviewdomain.ArticleReview) (*reviewdomain.ArticleReview, error) {
	created, err := scanArticleReview(rep.QueryRunner(ctx).QueryRow(ctx, createArticleReviewSQL, articleReview.ArticleID, articleReview.UserID, articleReview.Rating, articleReview.Comment))
	if db.IsUniqueViolation(err, articleReviewsUserArticleUniqueConstraint) {
		return nil, reviewdomain.ErrAlreadyReviewed
	}
	if err != nil {
		return nil, fmt.Errorf("repository.CreateArticleReview: %w", err)
	}

	return created, nil
}

// UpdateArticleReview updates an existing Article review's rating and comment.
func (rep *Repository) UpdateArticleReview(ctx context.Context, articleReview *reviewdomain.ArticleReview) error {
	tag, err := rep.QueryRunner(ctx).Exec(ctx, updateArticleReviewSQL, articleReview.ID, articleReview.Rating, articleReview.Comment)
	if err != nil {
		return fmt.Errorf("repository.UpdateArticleReview: %w", err)
	}

	if tag.RowsAffected() == 0 {
		return reviewdomain.ErrReviewNotFound
	}

	return nil
}

// DeleteArticleReview soft-deletes a Article review.
func (rep *Repository) DeleteArticleReview(ctx context.Context, reviewID, userID string) error {
	tag, err := rep.QueryRunner(ctx).Exec(ctx, deleteArticleReviewSQL, reviewID, userID)
	if err != nil {
		return fmt.Errorf("repository.DeleteArticleReview: %w", err)
	}

	if tag.RowsAffected() == 0 {
		return reviewdomain.ErrReviewNotFound
	}

	return nil
}

// GetArticleReviewList returns a paginated list of non-deleted reviews for a Article.
func (rep *Repository) GetArticleReviewList(ctx context.Context, params pagination.Params, articleID string, filter reviewdomain.ReviewFilter) ([]*reviewdomain.ArticleReview, error) {
	args := make([]any, 1)
	args[0] = articleID
	query := getArticleReviewByArticleIDSQL
	if filter.IsUsed() {
		query = query[:strings.Index(query, filterPlace)] + rep.GenerateFilterQuery(filter.Op)
		args = append(args, filter.Rating)
	} else {
		query = strings.Replace(query, filterPlace, "", 1)
	}
	return repository.GetAndParseListWithArgs(ctx, &rep.BaseRepository, query, "GetArticleReviewList", params, scanArticleReview, args)
}

// GetArticleReviewByID retrieves a non-deleted Article review by ID.
func (rep *Repository) GetArticleReviewByID(ctx context.Context, reviewID string) (*reviewdomain.ArticleReview, error) {
	articleReview, err := scanArticleReview(rep.QueryRunner(ctx).QueryRow(ctx, getArticleReviewByIDSQL, reviewID))
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, reviewdomain.ErrReviewNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("repository.GetArticleReviewByID: %w", err)
	}

	return articleReview, nil
}

// GetArticleReviewByUserAndArticleID retrieves a user's non-deleted review for a Article, if any.
func (rep *Repository) GetArticleReviewByUserAndArticleID(ctx context.Context, userID, articleID string) (*reviewdomain.ArticleReview, error) {
	articleReview, err := scanArticleReview(rep.QueryRunner(ctx).QueryRow(ctx, getArticleReviewByUserAndArticleIDSQL, articleID, userID))
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, reviewdomain.ErrReviewNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("repository.GetArticleReviewByUserAndArticleID: %w", err)
	}

	return articleReview, nil
}

// GetArticleReviewStats returns the average rating and review count for a Article.
func (rep *Repository) GetArticleReviewStats(ctx context.Context, articleID string) (rating float64, count int, err error) {
	err = rep.QueryRunner(ctx).QueryRow(ctx, getArticleReviewStats, articleID).Scan(&rating, &count)
	if err != nil {
		return 0, 0, fmt.Errorf("repository.GetArticleReviewStats: %w", err)
	}

	return rating, count, nil
}
