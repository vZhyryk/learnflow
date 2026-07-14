package router

import (
	"bytes"
	"context"
	"errors"
	"learnflow_backend/cmd/api/app"
	authdomain "learnflow_backend/internal/auth/domain"
	"learnflow_backend/internal/infrastructure/logger"
	appcontext "learnflow_backend/internal/shared/context"
	"learnflow_backend/internal/shared/testutil"
	"learnflow_backend/internal/shared/tokens"
	"net"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	. "github.com/smartystreets/goconvey/convey"
)

func newTestRouteHandler() *RouteHandler {
	a := &app.App{
		Logger: testutil.NewTestLogger(),
		Redis:  testutil.UnreachableRedis(),
	}
	a.Config.Timeouts.RequestTimeout = time.Second
	return &RouteHandler{App: a, token: tokens.NewTokens("test-secret-that-is-long-enough-32b", "", "learnflow", "learnflow-users")}
}

// newTestRequest returns a fresh recorder/GET-request pair for exercising a
// middleware directly (i.e. without going through a mux), saving each Convey
// case the recorder+request construction boilerplate.
func newTestRequest(path string) (*httptest.ResponseRecorder, *http.Request) {
	w := httptest.NewRecorder()
	r := httptest.NewRequestWithContext(context.Background(), http.MethodGet, path, http.NoBody)
	return w, r
}

// --- RecoverPanic ---

func TestRecoverPanic(t *testing.T) {
	Convey("RecoverPanic", t, func() {
		route := newTestRouteHandler()

		Convey("When next handler panics, it recovers and writes 500", func() {
			panicking := http.HandlerFunc(func(_ http.ResponseWriter, _ *http.Request) {
				panic("boom")
			})
			w, r := newTestRequest("/")

			route.RecoverPanic(panicking).ServeHTTP(w, r)

			So(w.Code, ShouldEqual, http.StatusInternalServerError)
			So(w.Header().Get("Connection"), ShouldEqual, "close")
		})

		Convey("When next handler does not panic, response passes through untouched", func() {
			w, r := newTestRequest("/")

			route.RecoverPanic(testutil.OkHandler()).ServeHTTP(w, r)

			So(w.Code, ShouldEqual, http.StatusOK)
		})
	})
}

// --- EnableCORS ---

func TestEnableCORS(t *testing.T) {
	Convey("EnableCORS", t, func() {
		route := newTestRouteHandler()
		route.App.Config.Cors.TrustedOrigins = map[string]struct{}{"https://trusted.example": {}}

		Convey("When origin is untrusted, no CORS headers are set and request passes through", func() {
			w, r := newTestRequest("/")
			r.Header.Set("Origin", "https://evil.example")

			route.EnableCORS(testutil.OkHandler()).ServeHTTP(w, r)

			So(w.Header().Get("Access-Control-Allow-Origin"), ShouldEqual, "")
			So(w.Code, ShouldEqual, http.StatusOK)
		})

		Convey("When origin is trusted, CORS headers are set and request passes through", func() {
			w, r := newTestRequest("/")
			r.Header.Set("Origin", "https://trusted.example")

			route.EnableCORS(testutil.OkHandler()).ServeHTTP(w, r)

			So(w.Header().Get("Access-Control-Allow-Origin"), ShouldEqual, "https://trusted.example")
			So(w.Header().Get("Access-Control-Allow-Credentials"), ShouldEqual, "true")
			So(w.Code, ShouldEqual, http.StatusOK)
		})

		Convey("When trusted origin sends an OPTIONS preflight, it short-circuits with 200", func() {
			w := httptest.NewRecorder()
			r := httptest.NewRequestWithContext(context.Background(), http.MethodOptions, "/", http.NoBody)
			r.Header.Set("Origin", "https://trusted.example")
			r.Header.Set("Access-Control-Request-Method", "POST")

			called := false
			next := http.HandlerFunc(func(_ http.ResponseWriter, _ *http.Request) { called = true })
			route.EnableCORS(next).ServeHTTP(w, r)

			So(called, ShouldBeFalse)
			So(w.Code, ShouldEqual, http.StatusOK)
			So(w.Header().Get("Access-Control-Allow-Methods"), ShouldContainSubstring, "POST")
		})
	})
}

// --- SetSecurityHeaders ---

func TestSetSecurityHeaders(t *testing.T) {
	Convey("SetSecurityHeaders", t, func() {
		route := newTestRouteHandler()
		w, r := newTestRequest("/")

		route.SetSecurityHeaders(testutil.OkHandler()).ServeHTTP(w, r)

		So(w.Header().Get("X-Frame-Options"), ShouldEqual, "DENY")
		So(w.Header().Get("X-Content-Type-Options"), ShouldEqual, "nosniff")
		So(w.Header().Get("Cache-Control"), ShouldEqual, "no-store")
		So(w.Header().Get("Pragma"), ShouldEqual, "no-store")
		So(w.Header().Get("Strict-Transport-Security"), ShouldEqual, "max-age=63072000; includeSubDomains; preload")
		So(w.Header().Get("Content-Security-Policy"), ShouldEqual, "default-src 'self'; object-src 'none'; script-src 'self' 'wasm-unsafe-eval'")
		So(w.Code, ShouldEqual, http.StatusOK)
	})
}

// --- Timeout ---

func TestTimeout(t *testing.T) {
	Convey("Timeout", t, func() {
		route := newTestRouteHandler()

		Convey("When next handler respects context, no deadline error occurs", func() {
			route.App.Config.Timeouts.RequestTimeout = time.Second
			w, r := newTestRequest("/")

			var ctxErr error
			next := http.HandlerFunc(func(_ http.ResponseWriter, r *http.Request) {
				ctxErr = r.Context().Err()
			})
			route.Timeout(next).ServeHTTP(w, r)

			So(ctxErr, ShouldBeNil)
		})

		Convey("When handler outlives the timeout, the request context is canceled with DeadlineExceeded", func() {
			route.App.Config.Timeouts.RequestTimeout = 20 * time.Millisecond
			w, r := newTestRequest("/")

			var ctxErr error
			next := http.HandlerFunc(func(_ http.ResponseWriter, r *http.Request) {
				<-r.Context().Done()
				ctxErr = r.Context().Err()
			})
			route.Timeout(next).ServeHTTP(w, r)

			So(errors.Is(ctxErr, context.DeadlineExceeded), ShouldBeTrue)
		})
	})
}

// --- RequestLogger ---

func TestRequestLogger(t *testing.T) {
	Convey("RequestLogger", t, func() {
		route := newTestRouteHandler()
		var logBuf *bytes.Buffer
		route.App.Logger, logBuf = testutil.NewBufferLogger(logger.LevelInfo)

		w, r := newTestRequest("/ping")

		next := http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			w.WriteHeader(http.StatusTeapot)
		})
		route.RequestLogger(next).ServeHTTP(w, r)

		So(w.Code, ShouldEqual, http.StatusTeapot)
		So(logBuf.String(), ShouldContainSubstring, `"path":"/ping"`)
		So(logBuf.String(), ShouldContainSubstring, `"status":418`)
	})
}

// --- SetRequestID ---

func TestSetRequestID(t *testing.T) {
	Convey("SetRequestID", t, func() {
		route := newTestRouteHandler()
		w, r := newTestRequest("/")

		var idFromCtx string
		next := http.HandlerFunc(func(_ http.ResponseWriter, r *http.Request) {
			idFromCtx = appcontext.RequestIDFromContext(r.Context())
		})
		route.SetRequestID(next).ServeHTTP(w, r)

		So(w.Header().Get(requestIDHeader), ShouldNotBeEmpty)
		So(idFromCtx, ShouldEqual, w.Header().Get(requestIDHeader))
	})
}

// --- SetIPAddress ---

func TestSetIPAddress(t *testing.T) {
	Convey("SetIPAddress", t, func() {
		route := newTestRouteHandler()
		route.App.Config.TrustedProxies = nil

		w, r := newTestRequest("/")
		r.RemoteAddr = "203.0.113.5:54321"

		var ip string
		next := http.HandlerFunc(func(_ http.ResponseWriter, r *http.Request) {
			ip = appcontext.IPAddressFromContext(r.Context())
		})
		route.SetIPAddress(next).ServeHTTP(w, r)

		So(ip, ShouldEqual, "203.0.113.5")
	})
}

// --- realClientIP / ipFromProxyHeaders / parseIP ---

func mustParseCIDR(t *testing.T, s string) net.IPNet {
	t.Helper()
	_, cidr, err := net.ParseCIDR(s)
	if err != nil {
		t.Fatalf("invalid CIDR %q: %v", s, err)
	}
	return *cidr
}

func TestRealClientIP(t *testing.T) {
	Convey("realClientIP", t, func() {
		trustedCIDR := mustParseCIDR(t, "10.0.0.0/8")

		Convey("When RemoteAddr is not a trusted proxy, RemoteAddr IP wins even with spoofed headers", func() {
			r := httptest.NewRequestWithContext(context.Background(), http.MethodGet, "/", http.NoBody)
			r.RemoteAddr = "203.0.113.5:1234"
			r.Header.Set("X-Forwarded-For", "1.2.3.4")

			ip := realClientIP(r, []net.IPNet{trustedCIDR})
			So(ip, ShouldEqual, "203.0.113.5")
		})

		Convey("When RemoteAddr is a trusted proxy, X-Forwarded-For is honored (first entry)", func() {
			r := httptest.NewRequestWithContext(context.Background(), http.MethodGet, "/", http.NoBody)
			r.RemoteAddr = "10.0.0.1:1234"
			r.Header.Set("X-Forwarded-For", "1.2.3.4, 5.6.7.8")

			ip := realClientIP(r, []net.IPNet{trustedCIDR})
			So(ip, ShouldEqual, "1.2.3.4")
		})

		Convey("When RemoteAddr is a trusted proxy and only X-Real-Ip is set, it is honored", func() {
			r := httptest.NewRequestWithContext(context.Background(), http.MethodGet, "/", http.NoBody)
			r.RemoteAddr = "10.0.0.1:1234"
			r.Header.Set("X-Real-Ip", "9.9.9.9")

			ip := realClientIP(r, []net.IPNet{trustedCIDR})
			So(ip, ShouldEqual, "9.9.9.9")
		})

		Convey("When RemoteAddr is a trusted proxy but proxy headers are malformed, it falls back to RemoteAddr", func() {
			r := httptest.NewRequestWithContext(context.Background(), http.MethodGet, "/", http.NoBody)
			r.RemoteAddr = "10.0.0.1:1234"
			r.Header.Set("X-Forwarded-For", "not-an-ip")

			ip := realClientIP(r, []net.IPNet{trustedCIDR})
			So(ip, ShouldEqual, "10.0.0.1")
		})

		Convey("When RemoteAddr has no port, RemoteAddr is returned verbatim", func() {
			r := httptest.NewRequestWithContext(context.Background(), http.MethodGet, "/", http.NoBody)
			r.RemoteAddr = "not-a-valid-addr"

			ip := realClientIP(r, []net.IPNet{trustedCIDR})
			So(ip, ShouldEqual, "not-a-valid-addr")
		})
	})
}

// --- AuthenticateUser ---

func TestAuthenticateUser(t *testing.T) {
	Convey("AuthenticateUser", t, func() {
		route := newTestRouteHandler()
		called := false
		next := http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			called = true
			w.WriteHeader(http.StatusOK)
		})

		Convey("When Authorization header is missing, it rejects with 401 and does not call next", func() {
			w, r := newTestRequest("/")

			route.AuthenticateUser(next).ServeHTTP(w, r)

			So(called, ShouldBeFalse)
			So(w.Code, ShouldEqual, http.StatusUnauthorized)
		})

		Convey("When Authorization header has no Bearer prefix, it rejects with 401", func() {
			w, r := newTestRequest("/")
			r.Header.Set("Authorization", "Basic abc123")

			route.AuthenticateUser(next).ServeHTTP(w, r)

			So(called, ShouldBeFalse)
			So(w.Code, ShouldEqual, http.StatusUnauthorized)
		})

		Convey("When the token is malformed/invalid, it rejects with 401 before touching Redis", func() {
			w, r := newTestRequest("/")
			r.Header.Set("Authorization", "Bearer not-a-real-jwt")

			route.AuthenticateUser(next).ServeHTTP(w, r)

			So(called, ShouldBeFalse)
			So(w.Code, ShouldEqual, http.StatusUnauthorized)
		})

		Convey("When the token is valid but Redis is unreachable, it fails closed with 500 and does not call next", func() {
			user := &authdomain.User{ID: "user-1", Role: authdomain.UserRole("student")}
			signed, err := route.token.GenerateAccessToken(user, time.Minute)
			So(err, ShouldBeNil)

			w, r := newTestRequest("/")
			r.Header.Set("Authorization", "Bearer "+signed)

			route.AuthenticateUser(next).ServeHTTP(w, r)

			So(called, ShouldBeFalse)
			So(w.Code, ShouldEqual, http.StatusInternalServerError)
		})
	})
}

// --- NewRouteRateLimiter ---

func TestNewRouteRateLimiter(t *testing.T) {
	Convey("NewRouteRateLimiter", t, func() {
		route := newTestRouteHandler()
		called := false
		next := http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			called = true
			w.WriteHeader(http.StatusOK)
		})
		limiter := route.NewRouteRateLimiter(1, time.Second, 1, func(_ *http.Request) string {
			return "test-key"
		})

		Convey("When the limiter is disabled, it always calls next without touching Redis", func() {
			route.App.Config.Limiter.Enabled = false
			w, r := newTestRequest("/")

			limiter(next).ServeHTTP(w, r)

			So(called, ShouldBeTrue)
			So(w.Code, ShouldEqual, http.StatusOK)
		})

		Convey("When the limiter is enabled and Redis is unreachable, it fails closed with 500 and does not call next", func() {
			route.App.Config.Limiter.Enabled = true
			w, r := newTestRequest("/")

			limiter(next).ServeHTTP(w, r)

			So(called, ShouldBeFalse)
			So(w.Code, ShouldEqual, http.StatusInternalServerError)
		})
	})
}
