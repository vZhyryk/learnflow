package adminhttp_test

import (
	"context"
	admindomain "learnflow_backend/internal/admin/domain"
	"learnflow_backend/internal/shared/pagination"
	"learnflow_backend/internal/shared/testutil"
	"net/http"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

const validAnnouncementID = "11111111-1111-1111-1111-111111111111"

func TestCreateAnnouncement(t *testing.T) {
	Convey("POST /api/v1/admin/announcement", t, func() {
		var svcErr error
		svc := &mockService{
			createAnnouncement: func(_ context.Context, _ admindomain.CreateAnnouncementRequest) (string, error) {
				return "announcement-123", svcErr
			},
		}

		f := newHTTPFixture(svc, http.MethodPost, "/api/v1/admin/announcement")
		mux, newReq := f.mux, f.newReq
		validBody := `{"title":"Test Announcement","body":"Body text","channels":["email"],"expires_at":"2099-01-01T00:00:00Z"}`

		Convey("No user in context → panics (middleware invariant violated)", func() {
			So(func() {
				testutil.ServeHTTP(mux, newReq(validBody, nil))
			}, ShouldPanic)
		})

		Convey("Empty body → 400", func() {
			w := testutil.ServeHTTP(mux, withUser(newReq("", nil)))
			So(w.Code, ShouldEqual, http.StatusBadRequest)
		})

		Convey("Service returns an error → 500", func() {
			svcErr = testutil.ErrDBUnexpected
			w := testutil.ServeHTTP(mux, withUser(newReq(validBody, nil)))
			So(w.Code, ShouldEqual, http.StatusInternalServerError)
		})

		Convey("Valid request → 201 with announcement_id", func() {
			w := testutil.ServeHTTP(mux, withUser(newReq(validBody, nil)))
			So(w.Code, ShouldEqual, http.StatusCreated)
			body := decodeBody(t, w.Body.Bytes())
			So(body["announcement_id"], ShouldEqual, "announcement-123")
		})

		Convey("Valid request and the success response write fails → does not panic", func() {
			So(func() {
				mux.ServeHTTP(&errWriter{}, withUser(newReq(validBody, nil)))
			}, ShouldNotPanic)
		})
	})
}

func TestUpdateAnnouncement(t *testing.T) {
	Convey("PUT /api/v1/admin/announcement", t, func() {
		var svcErr error
		svc := &mockService{
			updateAnnouncement: func(_ context.Context, _ admindomain.UpdateAnnouncementRequest) error {
				return svcErr
			},
		}

		f := newHTTPFixture(svc, http.MethodPut, "/api/v1/admin/announcement")
		mux, newReq := f.mux, f.newReq
		validBody := `{"id":"` + validAnnouncementID + `"}`

		Convey("No user in context → panics (middleware invariant violated)", func() {
			So(func() {
				testutil.ServeHTTP(mux, newReq(validBody, nil))
			}, ShouldPanic)
		})

		Convey("Empty body → 400", func() {
			w := testutil.ServeHTTP(mux, withUser(newReq("", nil)))
			So(w.Code, ShouldEqual, http.StatusBadRequest)
		})

		Convey("Missing id fails validation → 400 (request validation, before it reaches the service)", func() {
			w := testutil.ServeHTTP(mux, withUser(newReq(`{}`, nil)))
			So(w.Code, ShouldEqual, http.StatusBadRequest)
		})

		Convey("Service returns an error → 500", func() {
			svcErr = testutil.ErrDBUnexpected
			w := testutil.ServeHTTP(mux, withUser(newReq(validBody, nil)))
			So(w.Code, ShouldEqual, http.StatusInternalServerError)
		})

		Convey("Valid request → 200 with message", func() {
			w := testutil.ServeHTTP(mux, withUser(newReq(validBody, nil)))
			So(w.Code, ShouldEqual, http.StatusOK)
			body := decodeBody(t, w.Body.Bytes())
			So(body["message"], ShouldNotBeNil)
		})

		Convey("Valid request and the success response write fails → does not panic", func() {
			So(func() {
				mux.ServeHTTP(&errWriter{}, withUser(newReq(validBody, nil)))
			}, ShouldNotPanic)
		})
	})
}

func TestApproveAnnouncement(t *testing.T) {
	Convey("PUT /api/v1/admin/announcement/{id}/approve", t, func() {
		var svcErr error
		svc := &mockService{
			approveAnnouncement: func(_ context.Context, _, _ string) error {
				return svcErr
			},
		}

		f := newHTTPFixture(svc, http.MethodPut, "/api/v1/admin/announcement/"+validAnnouncementID+"/approve")
		mux, newReq := f.mux, f.newReq

		Convey("No user in context → panics (middleware invariant violated)", func() {
			So(func() {
				testutil.ServeHTTP(mux, newReq("", nil))
			}, ShouldPanic)
		})

		Convey("announcement not found → 404", func() {
			svcErr = admindomain.ErrAnnouncementNotFound
			w := testutil.ServeHTTP(mux, withUser(newReq("", nil)))
			So(w.Code, ShouldEqual, http.StatusNotFound)
		})

		Convey("Service returns an unexpected error → 500", func() {
			svcErr = testutil.ErrDBUnexpected
			w := testutil.ServeHTTP(mux, withUser(newReq("", nil)))
			So(w.Code, ShouldEqual, http.StatusInternalServerError)
		})

		Convey("Valid request → 200 with message", func() {
			w := testutil.ServeHTTP(mux, withUser(newReq("", nil)))
			So(w.Code, ShouldEqual, http.StatusOK)
			body := decodeBody(t, w.Body.Bytes())
			So(body["message"], ShouldNotBeNil)
		})

		Convey("Valid request and the success response write fails → does not panic", func() {
			So(func() {
				mux.ServeHTTP(&errWriter{}, withUser(newReq("", nil)))
			}, ShouldNotPanic)
		})
	})
}

func TestGetAnnouncements(t *testing.T) {
	Convey("GET /api/v1/admin/announcement/all", t, func() {
		var svcErr error
		svc := &mockService{
			getAnnouncements: func(_ context.Context, _ pagination.Params) ([]*admindomain.Announcement, error) {
				return []*admindomain.Announcement{}, svcErr
			},
		}

		f := newHTTPFixture(svc, http.MethodGet, "/api/v1/admin/announcement/all")
		mux, newReq := f.mux, f.newReq

		Convey("No user in context → still succeeds (handler does not require user)", func() {
			w := testutil.ServeHTTP(mux, newReq("", nil))
			So(w.Code, ShouldEqual, http.StatusOK)
		})

		Convey("Service returns an error → 500", func() {
			svcErr = testutil.ErrDBUnexpected
			w := testutil.ServeHTTP(mux, newReq("", nil))
			So(w.Code, ShouldEqual, http.StatusInternalServerError)
		})

		Convey("Valid request → 200 with announcements", func() {
			w := testutil.ServeHTTP(mux, newReq("", nil))
			So(w.Code, ShouldEqual, http.StatusOK)
			body := decodeBody(t, w.Body.Bytes())
			So(body["announcements"], ShouldNotBeNil)
		})

		Convey("Valid request and the success response write fails → does not panic", func() {
			So(func() {
				mux.ServeHTTP(&errWriter{}, newReq("", nil))
			}, ShouldNotPanic)
		})
	})
}

func TestGetUnApprovedAnnouncements(t *testing.T) {
	Convey("GET /api/v1/admin/announcement/unapproved", t, func() {
		var svcErr error
		svc := &mockService{
			getUnApprovedAnnouncements: func(_ context.Context, _ pagination.Params) ([]*admindomain.Announcement, error) {
				return []*admindomain.Announcement{}, svcErr
			},
		}

		f := newHTTPFixture(svc, http.MethodGet, "/api/v1/admin/announcement/unapproved")
		mux, newReq := f.mux, f.newReq

		Convey("Service returns an error → 500", func() {
			svcErr = testutil.ErrDBUnexpected
			w := testutil.ServeHTTP(mux, newReq("", nil))
			So(w.Code, ShouldEqual, http.StatusInternalServerError)
		})

		Convey("Valid request → 200 with announcements", func() {
			w := testutil.ServeHTTP(mux, newReq("", nil))
			So(w.Code, ShouldEqual, http.StatusOK)
			body := decodeBody(t, w.Body.Bytes())
			So(body["announcements"], ShouldNotBeNil)
		})
	})
}

func TestGetApprovedAnnouncements(t *testing.T) {
	Convey("GET /api/v1/admin/announcement/approved", t, func() {
		var svcErr error
		svc := &mockService{
			getApprovedAnnouncements: func(_ context.Context, _ pagination.Params) ([]*admindomain.Announcement, error) {
				return []*admindomain.Announcement{}, svcErr
			},
		}

		f := newHTTPFixture(svc, http.MethodGet, "/api/v1/admin/announcement/approved")
		mux, newReq := f.mux, f.newReq

		Convey("Service returns an error → 500", func() {
			svcErr = testutil.ErrDBUnexpected
			w := testutil.ServeHTTP(mux, newReq("", nil))
			So(w.Code, ShouldEqual, http.StatusInternalServerError)
		})

		Convey("Valid request → 200 with announcements", func() {
			w := testutil.ServeHTTP(mux, newReq("", nil))
			So(w.Code, ShouldEqual, http.StatusOK)
			body := decodeBody(t, w.Body.Bytes())
			So(body["announcements"], ShouldNotBeNil)
		})
	})
}

func TestGetExpiredAnnouncements(t *testing.T) {
	Convey("GET /api/v1/admin/announcement/expired", t, func() {
		var svcErr error
		svc := &mockService{
			getExpiredAnnouncements: func(_ context.Context, _ pagination.Params) ([]*admindomain.Announcement, error) {
				return []*admindomain.Announcement{}, svcErr
			},
		}

		f := newHTTPFixture(svc, http.MethodGet, "/api/v1/admin/announcement/expired")
		mux, newReq := f.mux, f.newReq

		Convey("Service returns an error → 500", func() {
			svcErr = testutil.ErrDBUnexpected
			w := testutil.ServeHTTP(mux, newReq("", nil))
			So(w.Code, ShouldEqual, http.StatusInternalServerError)
		})

		Convey("Valid request → 200 with announcements", func() {
			w := testutil.ServeHTTP(mux, newReq("", nil))
			So(w.Code, ShouldEqual, http.StatusOK)
			body := decodeBody(t, w.Body.Bytes())
			So(body["announcements"], ShouldNotBeNil)
		})
	})
}
