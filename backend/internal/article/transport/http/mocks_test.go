package articlehttp_test

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"

	articledomain "learnflow_backend/internal/article/domain"
	articlehttp "learnflow_backend/internal/article/transport/http"
	"learnflow_backend/internal/shared/pagination"
	"learnflow_backend/internal/shared/testutil"

	"github.com/justinas/alice"
)

type errWriter = testutil.ErrWriter

var decodeBody = testutil.DecodeBody
var withUser = testutil.WithUser

func newAuthMux(svc *mockService) *http.ServeMux {
	h := articlehttp.NewHTTPHandler(svc, testutil.NewTestLogger())
	mux := http.NewServeMux()
	h.RegisterRoutes(mux, alice.Chain{}, alice.Chain{})
	return mux
}

// httpFixture wires a mockService-backed mux and a request builder for a single
// route, shared by every per-handler fixture in this package (loginFixture,
// registerFixture, ...). Embed it and add the handler-specific svcResult/svcErr
// fields on top.
type httpFixture struct {
	mux    *http.ServeMux
	newReq func(body string, urlParams map[string]string) *http.Request
}

func newHTTPFixture(svc *mockService, method, path string) *httpFixture {
	return &httpFixture{
		mux: newAuthMux(svc),
		newReq: func(body string, urlParams map[string]string) *http.Request {
			if len(urlParams) > 0 {
				path += "?"
			}

			for key, value := range urlParams {
				if value != "" && key != "" {
					path += fmt.Sprintf("%s=%s", key, url.QueryEscape(value))
				}
			}

			return httptest.NewRequestWithContext(context.Background(), method, path, strings.NewReader(body))
		},
	}
}

type mockService struct {
	archiveArticle   func(ctx context.Context, articleID string) error
	createArticle    func(ctx context.Context, req articledomain.CreateArticleRequest) (string, error)
	deleteArticle    func(ctx context.Context, articleID string) error
	getArticleBySlug func(ctx context.Context, slug string) (*articledomain.Article, error)
	publishArticle   func(ctx context.Context, articleID string) error
	updateArticle    func(ctx context.Context, req articledomain.UpdateArticleRequest) error
	getAllArticles   func(ctx context.Context, getType articledomain.ArticleStatus, params pagination.Params) (articleList []*articledomain.Article, err error)
}

func (m *mockService) ArchiveArticle(ctx context.Context, articleID string) error {
	if m.archiveArticle == nil {
		panic("mockService.archiveArticle not set")
	}
	return m.archiveArticle(ctx, articleID)
}
func (m *mockService) CreateArticle(ctx context.Context, req articledomain.CreateArticleRequest) (string, error) {
	if m.createArticle == nil {
		panic("mockService.createArticle not set")
	}
	return m.createArticle(ctx, req)
}

func (m *mockService) DeleteArticle(ctx context.Context, articleID string) error {
	if m.deleteArticle == nil {
		panic("mockService.deleteArticle not set")
	}
	return m.deleteArticle(ctx, articleID)
}

func (m *mockService) GetArticleBySlug(ctx context.Context, slug string) (*articledomain.Article, error) {
	if m.getArticleBySlug == nil {
		panic("mockService.getArticleBySlug not set")
	}
	return m.getArticleBySlug(ctx, slug)
}
func (m *mockService) PublishArticle(ctx context.Context, articleID string) error {
	if m.publishArticle == nil {
		panic("mockService.publishArticle not set")
	}
	return m.publishArticle(ctx, articleID)
}
func (m *mockService) UpdateArticle(ctx context.Context, req articledomain.UpdateArticleRequest) error {
	if m.updateArticle == nil {
		panic("mockService.updateArticle not set")
	}
	return m.updateArticle(ctx, req)
}
func (m *mockService) GetAllArticles(ctx context.Context, getType articledomain.ArticleStatus, params pagination.Params) (articleList []*articledomain.Article, err error) {
	if m.getAllArticles == nil {
		panic("mockService.getAllArticles not set")
	}
	return m.getAllArticles(ctx, getType, params)
}
