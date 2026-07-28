package articledomain

import (
	"context"
	"learnflow_backend/internal/shared/pagination"
)

// Transactor executes a function within a database transaction.
type Transactor interface {
	InTransaction(ctx context.Context, fn func(ctx context.Context) error) error
}

// ArticleRepository defines persistence operations for Article.
type ArticleRepository interface {
	CreateArticle(ctx context.Context, article *Article) (*Article, error)
	PublishArticle(ctx context.Context, articleID string) error
	ArchiveArticle(ctx context.Context, articleID string) error
	DeleteArticle(ctx context.Context, articleID string) error
	UpdateArticle(ctx context.Context, article *Article) error
	GetAllPublishedArticles(ctx context.Context, params pagination.Params) ([]*Article, error)
	GetAllDraftArticles(ctx context.Context, params pagination.Params) ([]*Article, error)
	GetAllArchivedArticles(ctx context.Context, params pagination.Params) ([]*Article, error)
	GetAllArticles(ctx context.Context, params pagination.Params) ([]*Article, error)
	GetArticleByID(ctx context.Context, articleID string) (*Article, error)
	GetArticleBySlug(ctx context.Context, slug string) (*Article, error)
}

// Service defines the article module's business logic operations.
type Service interface {
	ArchiveArticle(ctx context.Context, articleID string) error
	CreateArticle(ctx context.Context, req CreateArticleRequest) (string, error)
	DeleteArticle(ctx context.Context, articleID string) error
	GetArticleBySlug(ctx context.Context, slug string) (*Article, error)
	PublishArticle(ctx context.Context, articleID string) error
	UpdateArticle(ctx context.Context, req UpdateArticleRequest) error
	GetAllArticles(ctx context.Context, getType ArticleStatus, params pagination.Params) (articleList []*Article, err error)
}
