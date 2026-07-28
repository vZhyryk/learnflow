package articleservice

import (
	"context"
	"learnflow_backend/internal/shared/testutil"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestDeleteArticle(t *testing.T) {
	Convey("DeleteArticle", t, func() {
		Convey("Success", func() {
			cRepo := &mockArticleRepo{
				deleteArticle: testutil.AlwaysNil,
			}

			srv := newTestService(cRepo, nil)
			err := srv.DeleteArticle(context.Background(), "ArticleID")
			So(err, ShouldBeNil)
		})

		Convey("Error", func() {
			cRepo := &mockArticleRepo{
				deleteArticle: testutil.AlwaysFailsDB,
			}

			srv := newTestService(cRepo, nil)
			err := srv.DeleteArticle(context.Background(), "ArticleID")
			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldContainSubstring, "db connection lost")
		})
	})
}
