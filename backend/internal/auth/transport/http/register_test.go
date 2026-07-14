package authhttp_test

import (
	"context"
	"net/http"
	"strings"
	"testing"

	authdomain "learnflow_backend/internal/auth/domain"
	"learnflow_backend/internal/shared/testutil"

	. "github.com/smartystreets/goconvey/convey"
)

type registerFixture struct {
	*httpFixture
	svcResult string
	svcErr    error
}

func newRegisterFixture() *registerFixture {
	f := &registerFixture{}
	svc := &mockService{
		register: func(_ context.Context, _ authdomain.RegisterRequest) (string, error) {
			return f.svcResult, f.svcErr
		},
	}
	f.httpFixture = newHTTPFixture(svc, http.MethodPost, "/api/v1/auth/register")
	return f
}

const validRegisterBody = `{"Email":"user@example.com","Password":"password123"}`

func TestRegisterRequestValidation(t *testing.T) {
	Convey("POST /api/v1/auth/register — request validation", t, func() {
		f := newRegisterFixture()

		Convey("Empty body → 400", func() {
			w := f.doRequest("")
			So(w.Code, ShouldEqual, http.StatusBadRequest)
		})

		Convey("Invalid JSON → 400", func() {
			w := f.doRequest("{invalid")
			So(w.Code, ShouldEqual, http.StatusBadRequest)
		})

		Convey("Invalid email format → 400", func() {
			w := f.doRequest(`{"Email":"notanemail","Password":"password123"}`)
			So(w.Code, ShouldEqual, http.StatusBadRequest)
		})

		Convey("Email too short → 400", func() {
			w := f.doRequest(`{"Email":"a@","Password":"password123"}`)
			So(w.Code, ShouldEqual, http.StatusBadRequest)
		})

		Convey("Password too short → 400", func() {
			w := f.doRequest(`{"Email":"user@example.com","Password":"short"}`)
			So(w.Code, ShouldEqual, http.StatusBadRequest)
		})

		Convey("Password too long → 400", func() {
			w := f.doRequest(`{"Email":"user@example.com","Password":"` + strings.Repeat("a", 73) + `"}`)
			So(w.Code, ShouldEqual, http.StatusBadRequest)
		})
	})
}

func TestRegisterServiceOutcomes(t *testing.T) {
	Convey("POST /api/v1/auth/register — service outcomes", t, func() {
		f := newRegisterFixture()

		Convey("Service ErrUserAlreadyExists → 202 (email enumeration guard)", func() {
			f.svcErr = authdomain.ErrUserAlreadyExists
			w := f.doRequest(validRegisterBody)
			So(w.Code, ShouldEqual, http.StatusAccepted)
		})

		Convey("Unexpected service error → 500", func() {
			f.svcErr = testutil.ErrDBUnexpected
			w := f.doRequest(validRegisterBody)
			So(w.Code, ShouldEqual, http.StatusInternalServerError)
		})

		Convey("Valid request → 201 with message", func() {
			f.svcResult = "user-123"
			w := f.doRequest(validRegisterBody)
			So(w.Code, ShouldEqual, http.StatusCreated)
			body := decodeBody(t, w.Body.Bytes())
			So(body["message"], ShouldNotBeNil)
		})

		Convey("Valid request and the success response write fails → does not panic", func() {
			f.svcResult = "user-123"
			So(func() { f.doRequestWithWriter(&errWriter{}, validRegisterBody) }, ShouldNotPanic)
		})
	})
}
