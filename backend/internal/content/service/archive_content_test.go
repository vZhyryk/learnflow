package contentservice

import (
	"context"
	"learnflow_backend/internal/shared/testutil"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestArchiveContentItem(t *testing.T) {
	Convey("Given a contentItem service", t, func() {
		Convey("When ArchiveContentItem succeeds", func() {
			cRepo := &mockContentItemRepo{
				archiveContentItem: testutil.AlwaysNil,
			}

			srv := newTestService(cRepo, nil)
			err := srv.ArchiveContentItem(context.Background(), "contentItemID")
			So(err, ShouldBeNil)
		})

		Convey("When the repository returns an error", func() {
			cRepo := &mockContentItemRepo{
				archiveContentItem: testutil.AlwaysFailsDB,
			}

			srv := newTestService(cRepo, nil)
			err := srv.ArchiveContentItem(context.Background(), "contentItemID")
			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldContainSubstring, "db connection lost")
		})
	})
}
