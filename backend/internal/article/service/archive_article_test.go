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
				archiveArticle: testutil.AlwaysNil,
			}

			srv := newTestService(cRepo, nil)
			err := srv.ArchiveArticle(context.Background(), "ArticleID")
			So(err, ShouldBeNil)
		})

		Convey("When the repository returns an error", func() {
			cRepo := &mockArticleRepo{
				archiveArticle: testutil.AlwaysFailsDB,
			}

			srv := newTestService(cRepo, nil)
			err := srv.ArchiveArticle(context.Background(), "ArticleID")
			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldContainSubstring, "db connection lost")
		})
	})
}
