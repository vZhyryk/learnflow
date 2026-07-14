package authhttp

import (
	"context"
	authdomain "learnflow_backend/internal/auth/domain"
	"learnflow_backend/internal/infrastructure/helpers"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	. "github.com/smartystreets/goconvey/convey"
)

func TestHandleErrorResponse(t *testing.T) {
	Convey("handleErrorResponse", t, func() {
		h := newHandlerForLog()
		w := httptest.NewRecorder()
		r := httptest.NewRequestWithContext(context.Background(), "GET", "/", http.NoBody)

		Convey("known error does not panic", func() {
			So(func() { h.handleErrorResponse(w, r, authdomain.ErrUserAlreadyExists) }, ShouldNotPanic)
		})

		Convey("nil error falls through to default does not panic", func() {
			So(func() { h.handleErrorResponse(w, r, nil) }, ShouldNotPanic)
		})
	})
}

func TestHandleErrorRespond(t *testing.T) {
	Convey("handleErrorRespond", t, func() {
		h := newHandlerForLog()
		w := httptest.NewRecorder()
		r := httptest.NewRequestWithContext(context.Background(), "GET", "/", http.NoBody)

		Convey("populated props does not panic", func() {
			So(func() {
				h.handleErrorRespond(r, "invalid_account_state", func() error {
					return helpers.WriteJSON(w, http.StatusOK, helpers.Envelope{"message": "if your account is eligible, you will receive an email"}, nil)
				})
			}, ShouldNotPanic)
		})
	})
}

func TestSetAccountLockHeader(t *testing.T) {
	Convey("setAccountLockHeader", t, func() {
		Convey("default", func() {
			So(setAccountLockHeader(nil), ShouldEqual, "900")
		})

		Convey("difference", func() {
			lockErr := &authdomain.ErrAccountLockedError{LockedUntil: time.Now().Add(10 * time.Second)}
			So(setAccountLockHeader(lockErr), ShouldEqual, "9")
		})
	})
}
