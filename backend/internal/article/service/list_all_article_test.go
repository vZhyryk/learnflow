package articleservice

import (
	"context"
	articledomain "learnflow_backend/internal/article/domain"
	"learnflow_backend/internal/shared/pagination"
	"learnflow_backend/internal/shared/testutil"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func getArticleError(_ context.Context, _ pagination.Params) ([]*articledomain.Article, error) {
	return nil, testutil.ErrDBUnexpected
}

func getValidList(_ context.Context, _ pagination.Params) ([]*articledomain.Article, error) {
	return []*articledomain.Article{{ID: "1"}, {ID: "2"}, {ID: "3"}, {ID: "4"}}, nil
}

func testGetAllArticlesByType(t *testing.T, scenario string, status articledomain.ArticleStatus, wireMock func(*mockArticleRepo, func(context.Context, pagination.Params) ([]*articledomain.Article, error))) {
	Convey("GetAllArticles Article", t, func() {
		Convey(scenario+" - error", func() {
			cRepo := &mockArticleRepo{}
			wireMock(cRepo, getArticleError)

			srv := newTestService(cRepo)
			Article, err := srv.GetAllArticles(context.Background(), status, pagination.Params{})
			So(err.Error(), ShouldContainSubstring, "db connection lost")
			So(Article, ShouldBeNil)
		})

		Convey(scenario+" - success", func() {
			cRepo := &mockArticleRepo{}
			wireMock(cRepo, getValidList)

			srv := newTestService(cRepo)
			Article, err := srv.GetAllArticles(context.Background(), status, pagination.Params{})
			So(err, ShouldBeNil)
			So(Article, ShouldNotBeNil)
			So(Article, ShouldNotBeEmpty)
			So(len(Article), ShouldEqual, 4)
		})
	})
}

func TestGetAllArticlesArchived(t *testing.T) {
	testGetAllArticlesByType(t, "getAllArchivedArticles", articledomain.ArchivedStatus, func(r *mockArticleRepo, fn func(context.Context, pagination.Params) ([]*articledomain.Article, error)) {
		r.getAllArchivedArticles = fn
	})
}

func TestGetAllArticlesPublished(t *testing.T) {
	testGetAllArticlesByType(t, "GetAllPublishedArticles", articledomain.PublishedStatus, func(r *mockArticleRepo, fn func(context.Context, pagination.Params) ([]*articledomain.Article, error)) {
		r.getAllPublishedArticles = fn
	})
}

func TestGetAllArticlesDraft(t *testing.T) {
	testGetAllArticlesByType(t, "GetAllDraftArticles", articledomain.DraftStatus, func(r *mockArticleRepo, fn func(context.Context, pagination.Params) ([]*articledomain.Article, error)) {
		r.getAllDraftArticles = fn
	})
}

func TestGetAllArticlesDefault(t *testing.T) {
	testGetAllArticlesByType(t, "default", articledomain.ArticleStatus(""), func(r *mockArticleRepo, fn func(context.Context, pagination.Params) ([]*articledomain.Article, error)) {
		r.getAllArticles = fn
	})
}
