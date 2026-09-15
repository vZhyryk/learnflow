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
				deleteArticle: testutil.AlwaysNil2,
			}

			srv := newTestService(cRepo, nil)
			err := srv.DeleteArticle(context.Background(), "ArticleID", "user-1")
			So(err, ShouldBeNil)
		})

		Convey("Error", func() {
			cRepo := &mockArticleRepo{
				deleteArticle: testutil.AlwaysFailsDB2,
			}

			srv := newTestService(cRepo, nil)
			err := srv.DeleteArticle(context.Background(), "ArticleID", "user-1")
			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldContainSubstring, "db connection lost")
		})
	})
}
