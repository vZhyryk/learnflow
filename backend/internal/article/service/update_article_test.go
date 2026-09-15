package articleservice

import (
	"context"
	"errors"
	articledomain "learnflow_backend/internal/article/domain"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func getValidArticle(_ context.Context, _ string) (*articledomain.Article, error) {
	return &articledomain.Article{
		Slug: "old slug",
		ID:   "article_id",
	}, nil
}

func TestUpdateArticle(t *testing.T) {
	Convey("UpdateArticle", t, func() {
		Convey("UpdateArticle - GetArticleByID error", func() {
			cRepo := &mockArticleRepo{
				getArticleByID: alwaysError,
			}

			srv := newTestService(cRepo, nil)
			err := srv.UpdateArticle(context.Background(), articledomain.UpdateArticleRequest{ID: "article_id"}, "user-1")
			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldContainSubstring, "db connection lost")
		})

		Convey("UpdateArticle - getArticleBySlug error", func() {
			cRepo := &mockArticleRepo{
				getArticleByID:   getValidArticle,
				getArticleBySlug: alwaysError,
			}

			srv := newTestService(cRepo, nil)
			slug := "New Slug"
			err := srv.UpdateArticle(context.Background(), articledomain.UpdateArticleRequest{ID: "article_id", Slug: &slug}, "user-1")
			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldContainSubstring, "db connection lost")
		})

		Convey("UpdateArticle - Same Slug different ID error", func() {
			cRepo := &mockArticleRepo{
				getArticleByID: getValidArticle,
				getArticleBySlug: func(_ context.Context, _ string) (*articledomain.Article, error) {
					return &articledomain.Article{Slug: "old slug", ID: "article_ID_2"}, nil
				},
			}

			srv := newTestService(cRepo, nil)
			slug := "New Slug"
			err := srv.UpdateArticle(context.Background(), articledomain.UpdateArticleRequest{ID: "article_id", Slug: &slug}, "user-1")
			So(err, ShouldNotBeNil)
			So(errors.Is(err, articledomain.ErrInvalidSlug), ShouldBeTrue)
		})

		Convey("UpdateArticle - updateArticle ID Match error ", func() {
			cRepo := &mockArticleRepo{
				getArticleByID:   getValidArticle,
				getArticleBySlug: getValidArticle,
				updateArticle:    alwaysFailsErr,
			}

			srv := newTestService(cRepo, nil)
			slug := "New Slug"
			err := srv.UpdateArticle(context.Background(), articledomain.UpdateArticleRequest{ID: "article_id", Slug: &slug}, "user-1")
			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldContainSubstring, "db connection lost")
		})

		Convey("UpdateArticle - nil slug skips uniqueness check, update failure propagates", func() {
			cRepo := &mockArticleRepo{
				getArticleByID:   getValidArticle,
				getArticleBySlug: getValidArticle,
				updateArticle:    alwaysFailsErr,
			}

			srv := newTestService(cRepo, nil)
			err := srv.UpdateArticle(context.Background(), articledomain.UpdateArticleRequest{ID: "article_id", Slug: nil}, "user-1")
			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldContainSubstring, "db connection lost")
		})

		Convey("Success", func() {
			cRepo := &mockArticleRepo{
				getArticleByID:   getValidArticle,
				getArticleBySlug: getValidArticle,
				updateArticle:    alwaysSucceedsUpdate,
			}

			srv := newTestService(cRepo, nil)
			err := srv.UpdateArticle(context.Background(), articledomain.UpdateArticleRequest{ID: "article_id", Slug: nil}, "user-1")
			So(err, ShouldBeNil)
		})
	})
}
