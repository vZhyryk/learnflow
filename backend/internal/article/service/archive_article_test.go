package articleservice

import (
	"context"
	"learnflow_backend/internal/shared/testutil"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestArchiveArticle(t *testing.T) {
	Convey("Given a Article service", t, func() {
		Convey("When ArchiveArticle succeeds", func() {
			cRepo := &mockArticleRepo{
				archiveArticle: testutil.AlwaysNil2,
			}

			srv := newTestService(cRepo, nil)
			err := srv.ArchiveArticle(context.Background(), "ArticleID", "user-1")
			So(err, ShouldBeNil)
		})

		Convey("When the repository returns an error", func() {
			cRepo := &mockArticleRepo{
				archiveArticle: testutil.AlwaysFailsDB2,
			}

			srv := newTestService(cRepo, nil)
			err := srv.ArchiveArticle(context.Background(), "ArticleID", "user-1")
			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldContainSubstring, "db connection lost")
		})
	})
}
