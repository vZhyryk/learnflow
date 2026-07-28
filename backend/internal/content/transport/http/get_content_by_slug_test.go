package contenthttp_test

import (
	"context"
	contentdomain "learnflow_backend/internal/content/domain"
	"learnflow_backend/internal/shared/testutil"
	"net/http"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestGetContentItemBySlug(t *testing.T) {
	Convey("GET /api/v1/content/{slug}", t, func() {
		var svcErr error

		svc := &mockService{
			getContentItemBySlug: func(_ context.Context, _ string) (*contentdomain.ContentItem, error) {
				return &contentdomain.ContentItem{}, svcErr
			},
		}

		f := newHTTPFixture(svc, http.MethodGet, "/api/v1/content/{slug}")
		mux, newReq := f.mux, f.newReq

		Convey("No user in context → still succeeds (public route)", func() {
			w := testutil.ServeHTTP(mux, newReq("", nil))
			So(w.Code, ShouldEqual, http.StatusOK)
		})

		Convey("Unexpected service error → 500", func() {
			svcErr = testutil.ErrDBUnexpected
			w := testutil.ServeHTTP(mux, withUser(newReq("", nil)))
			So(w.Code, ShouldEqual, http.StatusInternalServerError)
		})

		Convey("Content not found → 404", func() {
			svcErr = contentdomain.ErrContentItemNotFound
			w := testutil.ServeHTTP(mux, withUser(newReq("", nil)))
			So(w.Code, ShouldEqual, http.StatusNotFound)
		})

		Convey("Valid request → 200 with message", func() {
			w := testutil.ServeHTTP(mux, withUser(newReq("", nil)))
			So(w.Code, ShouldEqual, http.StatusOK)
			body := decodeBody(t, w.Body.Bytes())
			So(body["content_item"], ShouldNotBeNil)
		})

		Convey("Valid request and the success response write fails → does not panic", func() {
			So(func() {
				mux.ServeHTTP(&errWriter{}, withUser(newReq("", nil)))
			}, ShouldNotPanic)
		})
	})
}
