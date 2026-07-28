package contentservice

import (
	"context"
	"learnflow_backend/internal/shared/testutil"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestDeleteContentItem(t *testing.T) {
	Convey("DeleteContentItem", t, func() {
		Convey("Success", func() {
			cRepo := &mockContentItemRepo{
				deleteContentItem: testutil.AlwaysNil,
			}

			srv := newTestService(cRepo, nil)
			err := srv.DeleteContentItem(context.Background(), "contentItemID")
			So(err, ShouldBeNil)
		})

		Convey("Error", func() {
			cRepo := &mockContentItemRepo{
				deleteContentItem: testutil.AlwaysFailsDB,
			}

			srv := newTestService(cRepo, nil)
			err := srv.DeleteContentItem(context.Background(), "contentItemID")
			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldContainSubstring, "db connection lost")
		})
	})
}
