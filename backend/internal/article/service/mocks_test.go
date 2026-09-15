package articleservice

import (
	"context"
	articledomain "learnflow_backend/internal/article/domain"
	"learnflow_backend/internal/events"
	"learnflow_backend/internal/shared/pagination"
	"learnflow_backend/internal/shared/testutil"
)

type mockArticleRepo struct {
	createArticle           func(ctx context.Context, Article *articledomain.Article) (*articledomain.Article, error)
	publishArticle          func(ctx context.Context, ArticleID, userID string) error
	archiveArticle          func(ctx context.Context, ArticleID, userID string) error
	deleteArticle           func(ctx context.Context, ArticleID, userID string) error
	updateArticle           func(ctx context.Context, Article *articledomain.Article, userID string) error
	getAllPublishedArticles func(ctx context.Context, params pagination.Params) ([]*articledomain.Article, error)
	getAllDraftArticles     func(ctx context.Context, params pagination.Params) ([]*articledomain.Article, error)
	getAllArchivedArticles  func(ctx context.Context, params pagination.Params) ([]*articledomain.Article, error)
	getAllArticles          func(ctx context.Context, params pagination.Params) ([]*articledomain.Article, error)
	getArticleByID          func(ctx context.Context, ArticleID string) (*articledomain.Article, error)
	getArticleBySlug        func(ctx context.Context, slug string) (*articledomain.Article, error)
}

func (m *mockArticleRepo) CreateArticle(ctx context.Context, article *articledomain.Article) (*articledomain.Article, error) {
	if m.createArticle == nil {
		panic("mockArticleRepo.CreateArticle not set")
	}

	return m.createArticle(ctx, article)
}
func (m *mockArticleRepo) PublishArticle(ctx context.Context, articleID, userID string) error {
	if m.publishArticle == nil {
		panic("mockArticleRepo.publishArticle not set")
	}

	return m.publishArticle(ctx, articleID, userID)
}
func (m *mockArticleRepo) ArchiveArticle(ctx context.Context, articleID, userID string) error {
	if m.archiveArticle == nil {
		panic("mockArticleRepo.archiveArticle not set")
	}

	return m.archiveArticle(ctx, articleID, userID)
}
func (m *mockArticleRepo) DeleteArticle(ctx context.Context, articleID, userID string) error {
	if m.deleteArticle == nil {
		panic("mockArticleRepo.deleteArticle not set")
	}

	return m.deleteArticle(ctx, articleID, userID)
}
func (m *mockArticleRepo) UpdateArticle(ctx context.Context, article *articledomain.Article, userID string) error {
	if m.updateArticle == nil {
		panic("mockArticleRepo.updateArticle not set")
	}

	return m.updateArticle(ctx, article, userID)
}
func (m *mockArticleRepo) GetAllPublishedArticles(ctx context.Context, params pagination.Params) ([]*articledomain.Article, error) {
	if m.getAllPublishedArticles == nil {
		panic("mockArticleRepo.getAllPublishedArticles not set")
	}

	return m.getAllPublishedArticles(ctx, params)
}
func (m *mockArticleRepo) GetAllDraftArticles(ctx context.Context, params pagination.Params) ([]*articledomain.Article, error) {
	if m.getAllDraftArticles == nil {
		panic("mockArticleRepo.getAllDraftArticles not set")
	}

	return m.getAllDraftArticles(ctx, params)
}
func (m *mockArticleRepo) GetAllArchivedArticles(ctx context.Context, params pagination.Params) ([]*articledomain.Article, error) {
	if m.getAllArchivedArticles == nil {
		panic("mockArticleRepo.getAllArchivedArticles not set")
	}

	return m.getAllArchivedArticles(ctx, params)
}
func (m *mockArticleRepo) GetAllArticles(ctx context.Context, params pagination.Params) ([]*articledomain.Article, error) {
	if m.getAllArticles == nil {
		panic("mockArticleRepo.getAllArticles not set")
	}

	return m.getAllArticles(ctx, params)
}
func (m *mockArticleRepo) GetArticleByID(ctx context.Context, articleID string) (*articledomain.Article, error) {
	if m.getArticleByID == nil {
		panic("mockArticleRepo.getArticleByID not set")
	}

	return m.getArticleByID(ctx, articleID)
}
func (m *mockArticleRepo) GetArticleBySlug(ctx context.Context, slug string) (*articledomain.Article, error) {
	if m.getArticleBySlug == nil {
		panic("mockArticleRepo.getArticleBySlug not set")
	}

	return m.getArticleBySlug(ctx, slug)
}

func newTestService(repo *mockArticleRepo, outbox *events.OutboxWriter) *Service {
	return New(repo, &testutil.NoopTransactor{}, outbox)
}

// alwaysError is a getArticleByID/getArticleBySlug stub that always fails.
func alwaysError(_ context.Context, _ string) (*articledomain.Article, error) {
	return nil, testutil.ErrDBUnexpected
}

// alwaysFailsErr is an updateArticle stub that always fails.
func alwaysFailsErr(_ context.Context, _ *articledomain.Article, _ string) error {
	return testutil.ErrDBUnexpected
}

func alwaysSucceedsUpdate(_ context.Context, _ *articledomain.Article, _ string) error {
	return nil
}
