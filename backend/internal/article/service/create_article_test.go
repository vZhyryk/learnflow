package articleservice

import (
	"context"
	"errors"
	articledomain "learnflow_backend/internal/article/domain"
	"learnflow_backend/internal/shared/testutil"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestCreateArticle(t *testing.T) {
	Convey("Create Article", t, func() {
		Convey("GetArticleBySlug error", func() {
			cRepo := &mockArticleRepo{
				getArticleBySlug: alwaysError,
			}

			srv := newTestService(cRepo, nil)
			id, err := srv.CreateArticle(context.Background(), articledomain.CreateArticleRequest{})
			So(err, ShouldNotBeNil)
			So(id, ShouldBeEmpty)
		})

		Convey("GetArticleBySlug already exists", func() {
			cRepo := &mockArticleRepo{
				getArticleBySlug: func(_ context.Context, _ string) (*articledomain.Article, error) {
					return &articledomain.Article{}, nil
				},
			}

			srv := newTestService(cRepo, nil)
			id, err := srv.CreateArticle(context.Background(), articledomain.CreateArticleRequest{})
			So(err, ShouldNotBeNil)
			So(errors.Is(err, articledomain.ErrInvalidSlug), ShouldBeTrue)
			So(id, ShouldBeEmpty)
		})

		Convey("CreateArticle error", func() {
			cRepo := &mockArticleRepo{
				getArticleBySlug: func(_ context.Context, _ string) (*articledomain.Article, error) {
					return nil, nil
				},
				createArticle: func(_ context.Context, _ *articledomain.Article) (*articledomain.Article, error) {
					return nil, testutil.ErrDBUnexpected
				},
			}

			srv := newTestService(cRepo, nil)
			id, err := srv.CreateArticle(context.Background(), articledomain.CreateArticleRequest{})
			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldContainSubstring, "db connection lost")
			So(id, ShouldBeEmpty)
		})

		Convey("Success", func() {
			cRepo := &mockArticleRepo{
				getArticleBySlug: func(_ context.Context, _ string) (*articledomain.Article, error) {
					return nil, nil
				},
				createArticle: func(_ context.Context, _ *articledomain.Article) (*articledomain.Article, error) {
					return &articledomain.Article{ID: "article_ID"}, nil
				},
			}

			srv := newTestService(cRepo, nil)
			id, err := srv.CreateArticle(context.Background(), articledomain.CreateArticleRequest{})
			So(err, ShouldBeNil)
			So(id, ShouldEqual, "article_ID")
		})
	})
}

func TestCreateArticleIsIndexableOverride(t *testing.T) {
	Convey("Create Article with IsIndexable explicitly set to false", t, func() {
		var gotArticle *articledomain.Article
		cRepo := &mockArticleRepo{
			getArticleBySlug: func(_ context.Context, _ string) (*articledomain.Article, error) {
				return nil, nil
			},
			createArticle: func(_ context.Context, Article *articledomain.Article) (*articledomain.Article, error) {
				gotArticle = Article
				return &articledomain.Article{ID: "article_ID"}, nil
			},
		}

		srv := newTestService(cRepo, nil)
		isIndexable := false
		id, err := srv.CreateArticle(context.Background(), articledomain.CreateArticleRequest{IsIndexable: &isIndexable})
		So(err, ShouldBeNil)
		So(id, ShouldEqual, "article_ID")
		So(gotArticle.IsIndexable, ShouldBeFalse)
	})
}
