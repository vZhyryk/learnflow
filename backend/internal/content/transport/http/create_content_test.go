package contenthttp_test

import (
	"context"
	contentdomain "learnflow_backend/internal/content/domain"
	"learnflow_backend/internal/shared/testutil"
	"net/http"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestCreateContentItem(t *testing.T) {
	Convey("POST /api/v1/admin/content", t, func() {
		var svcErr error
		contentItemID := "content_item_id"
		validBody := `{"slug":"test-content","title":"Test content"}`

		svc := &mockService{
			createContentItem: func(_ context.Context, _ contentdomain.CreateContentItemRequest) (string, error) {
				return contentItemID, svcErr
			},
		}

		f := newHTTPFixture(svc, http.MethodPost, "/api/v1/admin/content")
		mux, newReq := f.mux, f.newReq

		Convey("No user in context → panics (middleware invariant violated)", func() {
			So(func() {
				testutil.ServeHTTP(mux, newReq(validBody, nil))
			}, ShouldPanic)
		})

		Convey("Empty body", func() {
			w := testutil.ServeHTTP(mux, withUser(newReq("", nil)))
			So(w.Code, ShouldEqual, http.StatusBadRequest)
		})

		Convey("Service returns an error", func() {
			svcErr = testutil.ErrDBUnexpected
			w := testutil.ServeHTTP(mux, withUser(newReq(validBody, nil)))
			So(w.Code, ShouldEqual, http.StatusInternalServerError)
		})

		Convey("Valid request → 201 with content_item_id", func() {
			w := testutil.ServeHTTP(mux, withUser(newReq(validBody, nil)))
			So(w.Code, ShouldEqual, http.StatusCreated)
			body := decodeBody(t, w.Body.Bytes())
			So(body["content_item_id"], ShouldNotBeNil)
		})

		Convey("Valid request and the success response write fails → does not panic", func() {
			So(func() {
				mux.ServeHTTP(&errWriter{}, withUser(newReq(validBody, nil)))
			}, ShouldNotPanic)
		})
	})
}
