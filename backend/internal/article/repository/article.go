package articlerepository

import (
	"context"
	"errors"
	"fmt"
	articledomain "learnflow_backend/internal/article/domain"
	"learnflow_backend/internal/infrastructure/db"
	"learnflow_backend/internal/shared/pagination"
	"learnflow_backend/internal/shared/repository"

	"github.com/jackc/pgx/v5"
)

// articleSlugUniqueConstraint is the DB-level backstop for the check-then-insert slug race.
const articleSlugUniqueConstraint = "articles_slug_unique"

// CreateArticle inserts a new draft Article.
func (rep *Repository) CreateArticle(ctx context.Context, article *articledomain.Article) (*articledomain.Article, error) {
	article, err := scanArticle(rep.QueryRunner(ctx).QueryRow(ctx, createDraftArticleSQL, article.Slug, article.Title, article.Description, article.Body, article.SeoTitle, article.SeoDescription, article.OgImageURL, article.IsIndexable, article.CreatedByUserID))
	if db.IsUniqueViolation(err, articleSlugUniqueConstraint) {
		return nil, articledomain.ErrInvalidSlug
	}
	if err != nil {
		return nil, fmt.Errorf("repository.CreateArticle: %w", err)
	}

	return article, nil
}

// PublishArticle marks a Article as published.
func (rep *Repository) PublishArticle(ctx context.Context, articleID, userID string) error {
	return repository.ExecUpdateByID(ctx, &rep.BaseRepository, publishArticleSQL, "PublishArticle", articleID, userID, articledomain.ErrArticleNotFound)
}

// ArchiveArticle marks a Article as archived.
func (rep *Repository) ArchiveArticle(ctx context.Context, articleID, userID string) error {
	return repository.ExecUpdateByID(ctx, &rep.BaseRepository, archiveArticleSQL, "ArchiveArticle", articleID, userID, articledomain.ErrArticleNotFound)
}

// DeleteArticle soft-deletes a Article.
func (rep *Repository) DeleteArticle(ctx context.Context, articleID, userID string) error {
	return repository.ExecUpdateByID(ctx, &rep.BaseRepository, deleteArticleSQL, "DeleteArticle", articleID, userID, articledomain.ErrArticleNotFound)
}

// UpdateArticle persists changes to an existing Article.
func (rep *Repository) UpdateArticle(ctx context.Context, article *articledomain.Article, userID string) error {
	tag, err := rep.QueryRunner(ctx).Exec(ctx, updateArticleSQL, article.ID, article.Slug, article.Title, article.Description, article.Body, article.SeoTitle, article.SeoDescription, article.OgImageURL, article.IsIndexable, userID)
	if db.IsUniqueViolation(err, articleSlugUniqueConstraint) {
		return articledomain.ErrInvalidSlug
	}
	if err != nil {
		return fmt.Errorf("repository.UpdateArticle: %w", err)
	}

	if tag.RowsAffected() == 0 {
		return articledomain.ErrArticleNotFound
	}

	return nil
}

// GetAllPublishedArticles returns every non-deleted published Article.
func (rep *Repository) GetAllPublishedArticles(ctx context.Context, params pagination.Params) ([]*articledomain.Article, error) {
	return repository.GetAndParseList(ctx, &rep.BaseRepository, getAllPublishedArticlesSQL, "GetAllPublishedArticles", params, scanArticle)
}

// GetAllDraftArticles returns every non-deleted draft Article.
func (rep *Repository) GetAllDraftArticles(ctx context.Context, params pagination.Params) ([]*articledomain.Article, error) {
	return repository.GetAndParseList(ctx, &rep.BaseRepository, getAllDraftArticlesSQL, "GetAllDraftArticles", params, scanArticle)
}

// GetAllArchivedArticles returns every archived Article, including soft-deleted ones.
func (rep *Repository) GetAllArchivedArticles(ctx context.Context, params pagination.Params) ([]*articledomain.Article, error) {
	return repository.GetAndParseList(ctx, &rep.BaseRepository, getAllArchivedArticlesSQL, "GetAllArchivedArticles", params, scanArticle)
}

// GetAllArticles returns every Article regardless of status, including soft-deleted ones.
func (rep *Repository) GetAllArticles(ctx context.Context, params pagination.Params) ([]*articledomain.Article, error) {
	return repository.GetAndParseList(ctx, &rep.BaseRepository, getAllArticlesSQL, "GetAllArticles", params, scanArticle)
}

// GetArticleByID retrieves a non-deleted Article by ID.
func (rep *Repository) GetArticleByID(ctx context.Context, articleID string) (*articledomain.Article, error) {
	article, err := scanArticle(rep.QueryRunner(ctx).QueryRow(ctx, getArticleByIDSQL, articleID))
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, articledomain.ErrArticleNotFound
	}

	if err != nil {
		return nil, fmt.Errorf("repository.GetArticleByID: %w", err)
	}

	return article, nil
}

// GetArticleBySlug retrieves a non-deleted Article by slug.
func (rep *Repository) GetArticleBySlug(ctx context.Context, slug string) (*articledomain.Article, error) {
	article, err := scanArticle(rep.QueryRunner(ctx).QueryRow(ctx, getArticleBySlugSQL, slug))
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, articledomain.ErrArticleNotFound
	}

	if err != nil {
		return nil, fmt.Errorf("repository.GetArticleBySlug: %w", err)
	}

	return article, nil
}
