package articlehttp_test

import (
	"context"
	articledomain "learnflow_backend/internal/article/domain"
	"learnflow_backend/internal/shared/testutil"
	"net/http"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestCreateArticle(t *testing.T) {
	Convey("POST /api/v1/admin/articles", t, func() {
		var svcErr error
		ArticleID := "article_id"
		validBody := `{"slug":"test-article","title":"Test article","body":"Article body"}`

		svc := &mockService{
			createArticle: func(_ context.Context, _ articledomain.CreateArticleRequest) (string, error) {
				return ArticleID, svcErr
			},
		}

		f := newHTTPFixture(svc, http.MethodPost, "/api/v1/admin/articles")
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

		Convey("Valid request → 201 with article_id", func() {
			w := testutil.ServeHTTP(mux, withUser(newReq(validBody, nil)))
			So(w.Code, ShouldEqual, http.StatusCreated)
			body := decodeBody(t, w.Body.Bytes())
			So(body["article_id"], ShouldNotBeNil)
		})

		Convey("Valid request and the success response write fails → does not panic", func() {
			So(func() {
				mux.ServeHTTP(&errWriter{}, withUser(newReq(validBody, nil)))
			}, ShouldNotPanic)
		})
	})
}
