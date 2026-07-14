package helpers_test

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"learnflow_backend/internal/infrastructure/helpers"
	"learnflow_backend/internal/infrastructure/logger"
	"learnflow_backend/internal/infrastructure/sanitizer"
	"learnflow_backend/internal/shared/testutil"

	. "github.com/smartystreets/goconvey/convey"
)

// errWriter is a shared test helper defined once in testutil.
type errWriter = testutil.ErrWriter

// --- WriteJSON ---

func TestWriteJSON(t *testing.T) {
	Convey("WriteJSON", t, func() {
		Convey("When given valid data", func() {
			w := httptest.NewRecorder()
			err := helpers.WriteJSON(w, http.StatusOK, helpers.Envelope{"key": "val"}, nil)
			So(err, ShouldBeNil)
			So(w.Code, ShouldEqual, http.StatusOK)
			So(w.Header().Get("Content-Type"), ShouldEqual, "application/json")
			var body map[string]any
			So(json.Unmarshal(w.Body.Bytes(), &body), ShouldBeNil)
			So(body["key"], ShouldEqual, "val")
		})

		Convey("When extra headers are provided", func() {
			w := httptest.NewRecorder()
			err := helpers.WriteJSON(w, http.StatusCreated, helpers.Envelope{}, http.Header{"X-Custom": []string{"yes"}})
			So(err, ShouldBeNil)
			So(w.Header().Get("X-Custom"), ShouldEqual, "yes")
			So(w.Code, ShouldEqual, http.StatusCreated)
		})

		Convey("When envelope contains a non-serialisable value", func() {
			w := httptest.NewRecorder()
			err := helpers.WriteJSON(w, http.StatusOK, helpers.Envelope{"ch": make(chan int)}, nil)
			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldContainSubstring, "marshaling JSON")
		})

		Convey("When the writer fails", func() {
			err := helpers.WriteJSON(&errWriter{}, http.StatusOK, helpers.Envelope{"k": "v"}, nil)
			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldContainSubstring, "writing JSON response")
		})
	})
}

// --- ReadJSON ---

func newPostRequest(body string) *http.Request {
	return httptest.NewRequestWithContext(context.Background(), http.MethodPost, "/", strings.NewReader(body))
}

func testReadJSONEmpty(postRequestValue, errMsg string) {
	var dst struct{}
	err := helpers.ReadJSON(httptest.NewRecorder(), newPostRequest(postRequestValue), &dst)
	So(err, ShouldNotBeNil)
	So(err.Error(), ShouldContainSubstring, errMsg)
}

func TestReadJSON(t *testing.T) {
	Convey("ReadJSON", t, func() {
		Convey("When body is valid JSON", func() {
			var dst struct {
				Name string `json:"name"`
			}
			err := helpers.ReadJSON(httptest.NewRecorder(), newPostRequest(`{"name":"Alice"}`), &dst)
			So(err, ShouldBeNil)
			So(dst.Name, ShouldEqual, "Alice")
		})

		Convey("When body is empty", func() {
			testReadJSONEmpty("", "must not be empty")
		})

		Convey("When body is malformed JSON", func() {
			testReadJSONEmpty("{bad", "badly-formed JSON")
		})

		Convey("When body has unknown field", func() {
			testReadJSONEmpty(`{"unknown":"x"}`, "unknown key")
		})

		Convey("When body has wrong type for field", func() {
			var dst struct {
				Name string `json:"name"`
			}
			err := helpers.ReadJSON(httptest.NewRecorder(), newPostRequest(`{"name":123}`), &dst)
			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldContainSubstring, "incorrect JSON type")
		})

		Convey("When body contains multiple JSON values", func() {
			testReadJSONEmpty(`{}{}`, "single JSON value")
		})
	})
}

// --- Response helpers ---

func testResponseHelper(status int, errMsg string, helper func(w http.ResponseWriter) error) {
	testResponseHelperAssert(status, errMsg, ShouldEqual, helper)
}

func testResponseHelperAssert(status int, errMsg string, assertion func(actual any, expected ...any) string, helper func(w http.ResponseWriter) error) {
	w := httptest.NewRecorder()
	err := helper(w)
	So(err, ShouldBeNil)
	So(w.Code, ShouldEqual, status)
	var body map[string]any
	So(json.Unmarshal(w.Body.Bytes(), &body), ShouldBeNil)
	So(body["error"], assertion, errMsg)
}

func TestInvalidCredentialsResponse(t *testing.T) {
	Convey("InvalidCredentialsResponse", t, func() {
		testResponseHelper(http.StatusUnauthorized, "unauthorized", helpers.InvalidCredentialsResponse)
	})
}

func TestForbiddenResponse(t *testing.T) {
	Convey("ForbiddenResponse", t, func() {
		testResponseHelper(http.StatusForbidden, "no access", func(w http.ResponseWriter) error {
			return helpers.ForbiddenResponse(w, helpers.Envelope{"error": "no access"})
		})
	})
}

func TestServerErrorResponse(t *testing.T) {
	Convey("ServerErrorResponse", t, func() {
		testResponseHelper(http.StatusInternalServerError, "internal server error", helpers.ServerErrorResponse)
	})
}

func TestBadRequestResponse(t *testing.T) {
	Convey("BadRequestResponse", t, func() {
		Convey("When err is non-nil", func() {
			testResponseHelper(http.StatusBadRequest, "invalid input", func(w http.ResponseWriter) error {
				return helpers.BadRequestResponse(w, errors.New("invalid input"))
			})
		})

		Convey("When err is nil", func() {
			testResponseHelper(http.StatusBadRequest, "bad request", func(w http.ResponseWriter) error {
				return helpers.BadRequestResponse(w, nil)
			})
		})
	})
}

func TestNotFoundResponse(t *testing.T) {
	Convey("NotFoundResponse", t, func() {
		testResponseHelperAssert(http.StatusNotFound, "could not be found", ShouldContainSubstring, helpers.NotFoundResponse)
	})
}

func TestErrorResponse(t *testing.T) {
	Convey("ErrorResponse", t, func() {
		testResponseHelper(http.StatusTeapot, "I'm a teapot", func(w http.ResponseWriter) error {
			return helpers.ErrorResponse(w, http.StatusTeapot, "I'm a teapot")
		})
	})
}

func TestRateLimitExceededResponse(t *testing.T) {
	Convey("RateLimitExceededResponse", t, func() {
		w := httptest.NewRecorder()

		err := helpers.RateLimitExceededResponse(w)

		So(err, ShouldBeNil)
		So(w.Code, ShouldEqual, http.StatusTooManyRequests)
		So(w.Header().Get("Retry-After"), ShouldEqual, "60")
		var body map[string]any
		So(json.Unmarshal(w.Body.Bytes(), &body), ShouldBeNil)
		So(body["error"], ShouldEqual, "rate limit exceeded")
	})
}

func TestLogRespondError(t *testing.T) {
	Convey("LogRespondError", t, func() {
		var buf bytes.Buffer
		jsonLogger := logger.New(&buf, nil, logger.LevelFatal)
		r := httptest.NewRequestWithContext(context.Background(), http.MethodGet, "/some/path", nil)

		Convey("When fn succeeds", func() {
			helpers.LogRespondError(jsonLogger, r, "test case", nil, func() error { return nil })

			So(buf.String(), ShouldEqual, "")
		})

		Convey("When fn fails", func() {
			helpers.LogRespondError(jsonLogger, r, "test case", nil, func() error { return errors.New("write failed") })

			var entry map[string]any
			So(json.Unmarshal(buf.Bytes(), &entry), ShouldBeNil)
			So(entry["message"], ShouldEqual, "write failed")
			props, ok := entry["properties"].(map[string]any)
			So(ok, ShouldBeTrue)
			So(props["case"], ShouldEqual, "test case")
			So(props["path"], ShouldEqual, "/some/path")
		})

		Convey("When fn fails with a wrapped error, the full chain reaches the log", func() {
			wrapped := fmt.Errorf("outer: %w", fmt.Errorf("inner: %w", errors.New("root cause")))
			helpers.LogRespondError(jsonLogger, r, "wrapped case", nil, func() error { return wrapped })

			var entry map[string]any
			So(json.Unmarshal(buf.Bytes(), &entry), ShouldBeNil)
			So(entry["message"], ShouldEqual, "outer: inner: root cause")
		})

		Convey("When the request path embeds a raw secret (path segment, not key=value)", func() {
			// MaskInlineSecrets only catches "key=value"/"key: value" patterns,
			// so a token embedded directly in a path segment (e.g. a password
			// reset or email verification link) has no such marker. The "path"
			// property is special-cased in sanitizer.sanitizeMapValue to go
			// through SanitizePath instead, which redacts opaque-looking
			// segments regardless of marker syntax.
			var sanitizedBuf bytes.Buffer
			realSanitizer := sanitizer.NewSanitizer("***REDACTED***", 2000, sanitizer.DefaultSensitiveKeys())
			sanitizedLogger := logger.New(&sanitizedBuf, realSanitizer, logger.LevelFatal)
			secretPath := "/auth/reset-password/eyJhbGciOiJIUzI1NiJ9.super-secret-token"
			req := httptest.NewRequestWithContext(context.Background(), http.MethodGet, secretPath, nil)

			helpers.LogRespondError(sanitizedLogger, req, "reset case", nil, func() error { return errors.New("boom") })

			var entry map[string]any
			So(json.Unmarshal(sanitizedBuf.Bytes(), &entry), ShouldBeNil)
			props, ok := entry["properties"].(map[string]any)
			So(ok, ShouldBeTrue)
			So(props["path"], ShouldEqual, "/auth/reset-password/***REDACTED***")
			So(props["path"], ShouldNotContainSubstring, "super-secret-token")
		})

		Convey("When the request path has a UUID resource ID, it is left untouched for log correlation", func() {
			var sanitizedBuf bytes.Buffer
			realSanitizer := sanitizer.NewSanitizer("***REDACTED***", 2000, sanitizer.DefaultSensitiveKeys())
			sanitizedLogger := logger.New(&sanitizedBuf, realSanitizer, logger.LevelFatal)
			uuidPath := "/users/550e8400-e29b-41d4-a716-446655440000/profile"
			req := httptest.NewRequestWithContext(context.Background(), http.MethodGet, uuidPath, nil)

			helpers.LogRespondError(sanitizedLogger, req, "profile case", nil, func() error { return errors.New("boom") })

			var entry map[string]any
			So(json.Unmarshal(sanitizedBuf.Bytes(), &entry), ShouldBeNil)
			props, ok := entry["properties"].(map[string]any)
			So(ok, ShouldBeTrue)
			So(props["path"], ShouldEqual, uuidPath)
		})
	})
}

func TestLogRespondErrorWithProps(t *testing.T) {
	Convey("LogRespondError merges extra props with case and path", t, func() {
		var buf bytes.Buffer
		jsonLogger := logger.New(&buf, nil, logger.LevelFatal)
		r := httptest.NewRequestWithContext(context.Background(), http.MethodGet, "/some/path", nil)

		helpers.LogRespondError(jsonLogger, r, "test case", map[string]any{"method": "POST", "ip": "1.2.3.4"}, func() error {
			return errors.New("write failed")
		})

		var entry map[string]any
		So(json.Unmarshal(buf.Bytes(), &entry), ShouldBeNil)
		props, ok := entry["properties"].(map[string]any)
		So(ok, ShouldBeTrue)
		So(props["case"], ShouldEqual, "test case")
		So(props["path"], ShouldEqual, "/some/path")
		So(props["method"], ShouldEqual, "POST")
		So(props["ip"], ShouldEqual, "1.2.3.4")
	})
}
