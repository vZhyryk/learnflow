package router

import (
	"context"
	"errors"
	"io"
	appcontext "learnflow_backend/internal/shared/context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

const testClientIP = "203.0.113.7"

func newBodyRequest(body io.Reader) *http.Request {
	r := httptest.NewRequestWithContext(context.Background(), http.MethodPost, "/", body)
	return r.WithContext(appcontext.WithIPAddress(r.Context(), testClientIP))
}

type failingReader struct{}

func (failingReader) Read([]byte) (int, error) { return 0, errors.New("read failed") }

func TestBodyRateLimitKey(t *testing.T) {
	route := newTestRouteHandler()
	extractEmail := func(body []byte) (string, bool) {
		if !strings.Contains(string(body), "email") {
			return "", false
		}
		return "ada@example.com", true
	}

	Convey("bodyRateLimitKey", t, func() {
		Convey("When the field is extracted, the key is its hash and the body is still readable downstream", func() {
			r := newBodyRequest(strings.NewReader(`{"email":"ada@example.com"}`))

			key := route.bodyRateLimitKey(r, extractEmail)

			So(key, ShouldEqual, route.rateLimitDigest("ada@example.com"))
			rest, err := io.ReadAll(r.Body)
			So(err, ShouldBeNil)
			So(string(rest), ShouldEqual, `{"email":"ada@example.com"}`)
		})

		Convey("When the field cannot be extracted, the key falls back to the client IP", func() {
			r := newBodyRequest(strings.NewReader(`{"other":"x"}`))

			So(route.bodyRateLimitKey(r, extractEmail), ShouldEqual, testClientIP)
		})

		Convey("When the extracted value is empty, the key falls back to the client IP", func() {
			r := newBodyRequest(strings.NewReader(`{"email":""}`))

			key := route.bodyRateLimitKey(r, func([]byte) (string, bool) { return "", true })

			So(key, ShouldEqual, testClientIP)
		})

		Convey("When the body is at the size limit, the key is a fixed oversized marker", func() {
			r := newBodyRequest(strings.NewReader(strings.Repeat("a", rateLimitBodyLimit)))

			So(route.bodyRateLimitKey(r, extractEmail), ShouldEqual, "oversized-body")
		})

		Convey("When the body is one byte under the limit, it is still parsed", func() {
			r := newBodyRequest(strings.NewReader("email" + strings.Repeat("x", rateLimitBodyLimit-1-len("email"))))

			So(route.bodyRateLimitKey(r, extractEmail), ShouldEqual, route.rateLimitDigest("ada@example.com"))
		})

		Convey("When reading the body fails, the key is empty", func() {
			r := newBodyRequest(failingReader{})

			So(route.bodyRateLimitKey(r, extractEmail), ShouldBeEmpty)
		})
	})
}

func TestGetEmailAndTokenFromBody(t *testing.T) {
	route := newTestRouteHandler()

	Convey("getEmailFromBody", t, func() {
		Convey("returns the hash of the email field", func() {
			r := newBodyRequest(strings.NewReader(`{"email":"ada@example.com"}`))

			So(route.getEmailFromBody(r), ShouldEqual, route.rateLimitDigest("ada@example.com"))
		})

		Convey("ignores case and surrounding spaces, so spelling variants share one bucket", func() {
			r := newBodyRequest(strings.NewReader(`{"email":"  Ada@Example.COM "}`))

			So(route.getEmailFromBody(r), ShouldEqual, route.rateLimitDigest("ada@example.com"))
		})

		Convey("falls back to the IP for malformed JSON", func() {
			r := newBodyRequest(strings.NewReader(`{not json`))

			So(route.getEmailFromBody(r), ShouldEqual, testClientIP)
		})

		Convey("falls back to the IP when the email is missing", func() {
			r := newBodyRequest(strings.NewReader(`{}`))

			So(route.getEmailFromBody(r), ShouldEqual, testClientIP)
		})
	})

	Convey("getTokenFromBody", t, func() {
		Convey("returns the hash of the token field", func() {
			r := newBodyRequest(strings.NewReader(`{"token":"raw-token"}`))

			So(route.getTokenFromBody(r), ShouldEqual, route.rateLimitDigest("raw-token"))
		})

		Convey("falls back to the IP for malformed JSON", func() {
			r := newBodyRequest(strings.NewReader(`[]`))

			So(route.getTokenFromBody(r), ShouldEqual, testClientIP)
		})

		Convey("falls back to the IP when the token is missing", func() {
			r := newBodyRequest(strings.NewReader(`{"email":"ada@example.com"}`))

			So(route.getTokenFromBody(r), ShouldEqual, testClientIP)
		})
	})
}

func TestRespondRateLimitExceeded(t *testing.T) {
	Convey("respondRateLimitExceeded", t, func() {
		route := newTestRouteHandler()
		w, r := newTestRequest("/")

		route.respondRateLimitExceeded(w, r)

		So(w.Code, ShouldEqual, http.StatusTooManyRequests)
	})
}

func TestBuildChains(t *testing.T) {
	Convey("Given the auth route chains with the limiter disabled", t, func() {
		route := newTestRouteHandler()
		route.App.Config.Limiter.Enabled = false
		chains := route.buildChains()
		called := false
		next := http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			called = true
			w.WriteHeader(http.StatusOK)
		})

		for name, chain := range map[string]func(http.Handler) http.Handler{
			"Static":      chains.Static.Then,
			"Login":       chains.Login.Then,
			"Register":    chains.Register.Then,
			"PassReset":   chains.PassReset.Then,
			"EmailVerify": chains.EmailVerify.Then,
		} {
			Convey("The "+name+" chain reaches the handler and adds the security headers", func() {
				called = false
				w, r := newTestRequest("/")

				chain(next).ServeHTTP(w, r)

				So(called, ShouldBeTrue)
				So(w.Code, ShouldEqual, http.StatusOK)
				So(w.Header().Get("X-Frame-Options"), ShouldEqual, "DENY")
				So(w.Header().Get("X-Request-Id"), ShouldNotBeEmpty)
			})
		}

		Convey("The StaticWithAuth chain rejects a request without a bearer token", func() {
			w, r := newTestRequest("/")

			chains.StaticWithAuth.Then(next).ServeHTTP(w, r)

			So(called, ShouldBeFalse)
			So(w.Code, ShouldEqual, http.StatusUnauthorized)
		})
	})
}

func TestRateLimitDigest(t *testing.T) {
	Convey("rateLimitDigest", t, func() {
		route := newTestRouteHandler()
		route.App.Config.Secret.JWTSecret = "first-secret-that-is-long-enough-32b"

		Convey("When the value is the same, the digest is stable and hex SHA-256 sized", func() {
			So(route.rateLimitDigest("ada@example.com"), ShouldEqual, route.rateLimitDigest("ada@example.com"))
			So(route.rateLimitDigest("ada@example.com"), ShouldHaveLength, 64)
		})

		Convey("When the server secret differs, the digest differs, so it cannot be rebuilt from a dictionary of emails", func() {
			first := route.rateLimitDigest("ada@example.com")
			route.App.Config.Secret.JWTSecret = "second-secret-that-is-long-enough-32"

			So(route.rateLimitDigest("ada@example.com"), ShouldNotEqual, first)
		})
	})
}
