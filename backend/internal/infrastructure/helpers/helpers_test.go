package helpers_test

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"learnflow_backend/internal/infrastructure/helpers"

	. "github.com/smartystreets/goconvey/convey"
)

// errWriter is an http.ResponseWriter whose Write always fails.
type errWriter struct {
	httptest.ResponseRecorder
}

func (e *errWriter) Write(_ []byte) (int, error) {
	return 0, errors.New("write failed")
}

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
			var dst struct{}
			err := helpers.ReadJSON(httptest.NewRecorder(), newPostRequest(""), &dst)
			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldContainSubstring, "must not be empty")
		})

		Convey("When body is malformed JSON", func() {
			var dst struct{}
			err := helpers.ReadJSON(httptest.NewRecorder(), newPostRequest("{bad"), &dst)
			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldContainSubstring, "badly-formed JSON")
		})

		Convey("When body has unknown field", func() {
			var dst struct{}
			err := helpers.ReadJSON(httptest.NewRecorder(), newPostRequest(`{"unknown":"x"}`), &dst)
			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldContainSubstring, "unknown key")
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
			var dst struct{}
			err := helpers.ReadJSON(httptest.NewRecorder(), newPostRequest(`{}{}`), &dst)
			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldContainSubstring, "single JSON value")
		})
	})
}

// --- Response helpers ---

func TestInvalidCredentialsResponse(t *testing.T) {
	Convey("InvalidCredentialsResponse", t, func() {
		w := httptest.NewRecorder()
		err := helpers.InvalidCredentialsResponse(w)
		So(err, ShouldBeNil)
		So(w.Code, ShouldEqual, http.StatusUnauthorized)
		var body map[string]any
		So(json.Unmarshal(w.Body.Bytes(), &body), ShouldBeNil)
		So(body["error"], ShouldEqual, "unauthorized")
	})
}

func TestForbiddenResponse(t *testing.T) {
	Convey("ForbiddenResponse", t, func() {
		w := httptest.NewRecorder()
		err := helpers.ForbiddenResponse(w, helpers.Envelope{"error": "no access"})
		So(err, ShouldBeNil)
		So(w.Code, ShouldEqual, http.StatusForbidden)
		var body map[string]any
		So(json.Unmarshal(w.Body.Bytes(), &body), ShouldBeNil)
		So(body["error"], ShouldEqual, "no access")
	})
}

func TestServerErrorResponse(t *testing.T) {
	Convey("ServerErrorResponse", t, func() {
		w := httptest.NewRecorder()
		err := helpers.ServerErrorResponse(w)
		So(err, ShouldBeNil)
		So(w.Code, ShouldEqual, http.StatusInternalServerError)
		var body map[string]any
		So(json.Unmarshal(w.Body.Bytes(), &body), ShouldBeNil)
		So(body["error"], ShouldEqual, "internal server error")
	})
}

func TestBadRequestResponse(t *testing.T) {
	Convey("BadRequestResponse", t, func() {
		Convey("When err is non-nil", func() {
			w := httptest.NewRecorder()
			err := helpers.BadRequestResponse(w, errors.New("invalid input"))
			So(err, ShouldBeNil)
			So(w.Code, ShouldEqual, http.StatusBadRequest)
			var body map[string]any
			So(json.Unmarshal(w.Body.Bytes(), &body), ShouldBeNil)
			So(body["error"], ShouldEqual, "invalid input")
		})

		Convey("When err is nil", func() {
			w := httptest.NewRecorder()
			err := helpers.BadRequestResponse(w, nil)
			So(err, ShouldBeNil)
			So(w.Code, ShouldEqual, http.StatusBadRequest)
			var body map[string]any
			So(json.Unmarshal(w.Body.Bytes(), &body), ShouldBeNil)
			So(body["error"], ShouldEqual, "bad request")
		})
	})
}

func TestNotFoundResponse(t *testing.T) {
	Convey("NotFoundResponse", t, func() {
		w := httptest.NewRecorder()
		err := helpers.NotFoundResponse(w)
		So(err, ShouldBeNil)
		So(w.Code, ShouldEqual, http.StatusNotFound)
		var body map[string]any
		So(json.Unmarshal(w.Body.Bytes(), &body), ShouldBeNil)
		So(body["error"], ShouldContainSubstring, "could not be found")
	})
}

func TestErrorResponse(t *testing.T) {
	Convey("ErrorResponse", t, func() {
		w := httptest.NewRecorder()
		err := helpers.ErrorResponse(w, http.StatusTeapot, "I'm a teapot")
		So(err, ShouldBeNil)
		So(w.Code, ShouldEqual, http.StatusTeapot)
		var body map[string]any
		So(json.Unmarshal(w.Body.Bytes(), &body), ShouldBeNil)
		So(body["error"], ShouldEqual, "I'm a teapot")
	})
}
