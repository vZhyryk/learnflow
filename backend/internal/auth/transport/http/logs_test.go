package authhttp

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"learnflow_backend/internal/infrastructure/logger"
	"learnflow_backend/internal/infrastructure/sanitizer"

	. "github.com/smartystreets/goconvey/convey"
)

func newHandlerForLog() *Handler {
	return &Handler{
		jsonLogger: logger.New(io.Discard, sanitizer.NewSanitizer("***", 100, nil), logger.LevelFatal),
	}
}

func TestLogAuthEvent(t *testing.T) {
	Convey("logAuthEvent", t, func() {
		h := newHandlerForLog()
		r := httptest.NewRequestWithContext(context.Background(), "GET", "/", http.NoBody)

		Convey("nil props does not panic", func() {
			So(func() { h.logAuthEvent(r, "auth.test", nil) }, ShouldNotPanic)
		})

		Convey("populated props does not panic", func() {
			So(func() { h.logAuthEvent(r, "auth.test", map[string]any{"user_id": "u-1"}) }, ShouldNotPanic)
		})
	})
}

func TestLogAuthFailure(t *testing.T) {
	Convey("logAuthFailure", t, func() {
		h := newHandlerForLog()
		r := httptest.NewRequestWithContext(context.Background(), "GET", "/", http.NoBody)

		Convey("nil props does not panic", func() {
			So(func() { h.logAuthFailure(r, "auth.login", "invalid_credentials", nil) }, ShouldNotPanic)
		})

		Convey("populated props does not panic", func() {
			So(func() {
				h.logAuthFailure(r, "auth.login", "account_locked", map[string]any{"user_id": "u-1"})
			}, ShouldNotPanic)
		})
	})
}
