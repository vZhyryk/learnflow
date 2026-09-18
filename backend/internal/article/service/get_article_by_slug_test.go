package articleservice

import (
	"context"
	"errors"
	articledomain "learnflow_backend/internal/article/domain"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestGetArticleBySlug(t *testing.T) {
	Convey("Given a Article service", t, func() {
		Convey("When the Article is published", func() {
			cRepo := &mockArticleRepo{
				getArticleBySlug: func(_ context.Context, _ string) (*articledomain.Article, error) {
					return &articledomain.Article{Status: articledomain.PublishedStatus}, nil
				},
			}

			srv := newTestService(cRepo)
			Article, err := srv.GetArticleBySlug(context.Background(), "Article_slug")
			So(err, ShouldBeNil)
			So(Article, ShouldNotBeNil)
			So(Article.Status, ShouldEqual, articledomain.PublishedStatus)
		})

		Convey("When the repository returns an error", func() {
			cRepo := &mockArticleRepo{
				getArticleBySlug: alwaysError,
			}

			srv := newTestService(cRepo)
			Article, err := srv.GetArticleBySlug(context.Background(), "Article_slug")
			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldContainSubstring, "db connection lost")
			So(Article, ShouldBeNil)
		})

		Convey("When the Article is not published", func() {
			cRepo := &mockArticleRepo{
				getArticleBySlug: func(_ context.Context, _ string) (*articledomain.Article, error) {
					return &articledomain.Article{Status: articledomain.DraftStatus}, nil
				},
			}

			srv := newTestService(cRepo)
			Article, err := srv.GetArticleBySlug(context.Background(), "Article_slug")
			So(err, ShouldNotBeNil)
			So(errors.Is(err, articledomain.ErrArticleNotFound), ShouldBeTrue)
			So(Article, ShouldBeNil)
		})
	})
}
