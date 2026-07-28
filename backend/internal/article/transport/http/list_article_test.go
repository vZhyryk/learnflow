package articlehttp_test

import (
	"context"
	articledomain "learnflow_backend/internal/article/domain"
	"learnflow_backend/internal/shared/pagination"
	"learnflow_backend/internal/shared/testutil"
	"net/http"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func AestListarticle(t *testing.T) {
	Convey("GET /api/v1/articles", t, func() {
		var svcErr error

		svc := &mockService{
			getAllArticles: func(_ context.Context, _ articledomain.ArticleStatus, _ pagination.Params) ([]*articledomain.Article, error) {
				return []*articledomain.Article{}, svcErr
			},
		}

		f := newHTTPFixture(svc, http.MethodGet, "/api/v1/articles")
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

		Convey("Valid request → 200 with message", func() {
			w := testutil.ServeHTTP(mux, withUser(newReq("", nil)))
			So(w.Code, ShouldEqual, http.StatusOK)
			body := decodeBody(t, w.Body.Bytes())
			So(body["article_list"], ShouldNotBeNil)
		})

		Convey("Valid request and the success response write fails → does not panic", func() {
			So(func() {
				mux.ServeHTTP(&errWriter{}, withUser(newReq("", nil)))
			}, ShouldNotPanic)
		})
	})
}
