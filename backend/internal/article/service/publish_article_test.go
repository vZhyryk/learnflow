package articleservice

import (
	"context"
	"errors"
	articledomain "learnflow_backend/internal/article/domain"
	"learnflow_backend/internal/shared/testutil"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func validGetArticleByID(_ context.Context, _ string) (*articledomain.Article, error) {
	seoTitle := "SeoTitle"
	seoDescription := "SeoDescription"
	description := "description"
	return &articledomain.Article{
		Status:         articledomain.DraftStatus,
		Title:          "title",
		SeoTitle:       &seoTitle,
		SeoDescription: &seoDescription,
		Body:           "body",
		Description:    &description,
	}, nil
}

func TestPublishArticle(t *testing.T) {
	Convey("PublishArticle", t, func() {
		Convey("PublishArticle - getArticleByID error", func() {
			cRepo := &mockArticleRepo{
				getArticleByID: alwaysError,
			}

			srv := newTestService(cRepo, nil)
			err := srv.PublishArticle(context.Background(), "article_ID", "user-1")
			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldContainSubstring, "db connection lost")
		})

		Convey("PublishArticle - wrong status", func() {
			cRepo := &mockArticleRepo{
				getArticleByID: func(_ context.Context, _ string) (*articledomain.Article, error) {
					return &articledomain.Article{Status: articledomain.ArchivedStatus}, nil
				},
			}

			srv := newTestService(cRepo, nil)
			err := srv.PublishArticle(context.Background(), "article_ID", "user-1")
			So(err, ShouldNotBeNil)
			So(errors.Is(err, articledomain.ErrInvalidArticleStatus), ShouldBeTrue)
		})

		Convey("PublishArticle - not ready to publish", func() {
			cRepo := &mockArticleRepo{
				getArticleByID: func(_ context.Context, _ string) (*articledomain.Article, error) {
					return &articledomain.Article{Status: articledomain.DraftStatus}, nil
				},
			}

			srv := newTestService(cRepo, nil)
			err := srv.PublishArticle(context.Background(), "article_ID", "user-1")
			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldContainSubstring, "service.PublishArticle")
		})

		Convey("PublishArticle - publish error", func() {
			cRepo := &mockArticleRepo{
				getArticleByID: validGetArticleByID,
				publishArticle: testutil.AlwaysFailsDB2,
			}

			srv := newTestService(cRepo, nil)
			err := srv.PublishArticle(context.Background(), "article_ID", "user-1")
			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldContainSubstring, "service.PublishArticle")
		})

		Convey("PublishArticle - outbox emit error", func() {
			cRepo := &mockArticleRepo{
				getArticleByID: validGetArticleByID,
				publishArticle: testutil.AlwaysNil2,
			}

			srv := newTestService(cRepo, testutil.NewFailingOutbox(testutil.ErrDBUnexpected))
			err := srv.PublishArticle(context.Background(), "article_ID", "user-1")
			So(err, ShouldNotBeNil)
			So(errors.Is(err, testutil.ErrDBUnexpected), ShouldBeTrue)
		})

		Convey("Successful", func() {
			cRepo := &mockArticleRepo{
				getArticleByID: validGetArticleByID,
				publishArticle: testutil.AlwaysNil2,
			}
			var captured []any

			srv := newTestService(cRepo, testutil.NewCapturingOutbox(&captured))
			err := srv.PublishArticle(context.Background(), "article_ID", "user-1")
			So(err, ShouldBeNil)
		})
	})
}
