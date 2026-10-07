package admin

import (
	"context"
	admindomain "learnflow_backend/internal/admin/domain"
	"learnflow_backend/internal/shared/testutil"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/justinas/alice"
	. "github.com/smartystreets/goconvey/convey"
)

// stubService satisfies admindomain.Service without implementing it: routes are only matched, never served.
type stubService struct{ admindomain.Service }

func TestRegisterAdminRoutes(t *testing.T) {
	Convey("Given a mux with the admin routes registered", t, func() {
		mux := http.NewServeMux()
		RegisterAdminRoutes(mux, stubService{}, alice.Chain{}, alice.Chain{}, testutil.NewTestLogger())

		routes := []struct{ method, path, pattern string }{
			{http.MethodGet, "/api/v1/admin/users", "GET /api/v1/admin/users"},
			{http.MethodGet, "/api/v1/admin/users/11111111-1111-1111-1111-111111111111", "GET /api/v1/admin/users/{id}"},
			{http.MethodPost, "/api/v1/admin/users/11111111-1111-1111-1111-111111111111/course-access", "POST /api/v1/admin/users/{id}/course-access"},
			{http.MethodPut, "/api/v1/admin/announcements/11111111-1111-1111-1111-111111111111/expired", "PUT /api/v1/admin/announcements/{id}/expired"},
			{http.MethodGet, "/api/v1/admin/audit/actions", "GET /api/v1/admin/audit/actions"},
			{http.MethodGet, "/api/v1/admin/dlq", "GET /api/v1/admin/dlq"},
			{http.MethodGet, "/api/v1/announcements", "GET /api/v1/announcements"},
		}

		for _, route := range routes {
			Convey(route.method+" "+route.path+" is routed to the admin handlers", func() {
				req := httptest.NewRequestWithContext(context.Background(), route.method, route.path, http.NoBody)

				_, pattern := mux.Handler(req)

				So(pattern, ShouldEqual, route.pattern)
			})
		}

		Convey("An unknown admin path is not routed", func() {
			req := httptest.NewRequestWithContext(context.Background(), http.MethodGet, "/api/v1/admin/nope", http.NoBody)

			_, pattern := mux.Handler(req)

			So(pattern, ShouldBeEmpty)
		})
	})
}
