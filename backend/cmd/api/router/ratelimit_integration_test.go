//go:build integration

package router

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	redisinfra "learnflow_backend/internal/infrastructure/redis"
	"learnflow_backend/internal/shared/testutil"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/redis/go-redis/v9"
	. "github.com/smartystreets/goconvey/convey"
)

func newRealRedisRouteHandler(t *testing.T) *RouteHandler {
	t.Helper()

	client := redis.NewClient(&redis.Options{Addr: "localhost:6379"})
	t.Cleanup(func() { client.Close() }) //nolint:errcheck // Close's error is never actionable in test cleanup

	route := newTestRouteHandler()
	route.App.Redis = &redisinfra.Instance{Client: client}
	route.App.Config.Limiter.Enabled = true

	return route
}

func postJSON(handler http.Handler, remoteAddr, body string) *httptest.ResponseRecorder {
	w := httptest.NewRecorder()
	r := httptest.NewRequestWithContext(context.Background(), http.MethodPost, "/", strings.NewReader(body))
	r.RemoteAddr = remoteAddr
	handler.ServeHTTP(w, r)

	return w
}

func TestRateLimitedChains_Integration(t *testing.T) {
	Convey("Given the auth chains with the limiter enabled on real Redis", t, func() {
		route := newRealRedisRouteHandler(t)
		chains := route.buildChains()
		suffix := time.Now().UnixNano()
		ip := fmt.Sprintf("198.51.100.%d:4321", suffix%200+1)

		var gotBody string
		echo := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			raw, err := io.ReadAll(r.Body)
			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				return
			}
			gotBody = string(raw)
			w.WriteHeader(http.StatusOK)
		})

		Convey("When the same IP and email hit the password-reset chain twice in a row, the second is limited with 429", func() {
			body := fmt.Sprintf(`{"email":"ada-%d@example.com"}`, suffix)
			handler := chains.PassReset.Then(echo)

			first := postJSON(handler, ip, body)
			second := postJSON(handler, ip, body)

			So(first.Code, ShouldEqual, http.StatusOK)
			So(gotBody, ShouldEqual, body)
			So(second.Code, ShouldEqual, http.StatusTooManyRequests)
		})

		Convey("When another email hits the same chain from the same IP, it has its own bucket", func() {
			handler := chains.PassReset.Then(echo)

			a := postJSON(handler, ip, fmt.Sprintf(`{"email":"a-%d@example.com"}`, suffix))
			b := postJSON(handler, ip, fmt.Sprintf(`{"email":"b-%d@example.com"}`, suffix))

			So(a.Code, ShouldEqual, http.StatusOK)
			So(b.Code, ShouldEqual, http.StatusOK)
		})

		Convey("When the email verify chain sees the same token twice, the second is limited", func() {
			body := fmt.Sprintf(`{"token":"tok-%d"}`, suffix)
			handler := chains.EmailVerify.Then(echo)

			first := postJSON(handler, ip, body)
			second := postJSON(handler, ip, body)

			So(first.Code, ShouldEqual, http.StatusOK)
			So(second.Code, ShouldEqual, http.StatusTooManyRequests)
		})

		Convey("When Redis is unreachable, the limiter fails closed with 500", func() {
			route.App.Redis = testutil.UnreachableRedisInstance()

			w := postJSON(route.buildChains().Register.Then(echo), ip, `{}`)

			So(w.Code, ShouldEqual, http.StatusInternalServerError)
		})
	})
}

func TestReadiness_Integration(t *testing.T) {
	type readiness struct {
		Status string `json:"status"`
		Reason string `json:"reason"`
	}
	call := func(route *RouteHandler) (int, readiness) {
		w, r := newTestRequest("/readiness")
		route.Readiness(w, r)

		var body readiness
		So(json.Unmarshal(w.Body.Bytes(), &body), ShouldBeNil)

		return w.Code, body
	}

	Convey("Readiness against real Postgres and Redis", t, func() {
		route := newRealRedisRouteHandler(t)
		route.App.DB = testutil.NewTestPool(t)

		Convey("When both are reachable, it reports ready", func() {
			code, body := call(route)

			So(code, ShouldEqual, http.StatusOK)
			So(body.Status, ShouldEqual, "ready")
		})

		Convey("When Redis is unreachable, it reports 503 with reason redis", func() {
			route.App.Redis = testutil.UnreachableRedisInstance()

			code, body := call(route)

			So(code, ShouldEqual, http.StatusServiceUnavailable)
			So(body, ShouldResemble, readiness{Status: "unavailable", Reason: "redis"})
		})

		Convey("When the database pool is closed, it reports 503 with reason database", func() {
			route.App.DB = testutil.NewTestPool(t)
			route.App.DB.Close()

			code, body := call(route)

			So(code, ShouldEqual, http.StatusServiceUnavailable)
			So(body, ShouldResemble, readiness{Status: "unavailable", Reason: "database"})
		})
	})
}
