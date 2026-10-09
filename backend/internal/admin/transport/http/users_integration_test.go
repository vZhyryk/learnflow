//go:build integration

package adminhttp_test

import (
	"context"
	adminrepository "learnflow_backend/internal/admin/repository"
	adminservice "learnflow_backend/internal/admin/service"
	adminhttp "learnflow_backend/internal/admin/transport/http"
	auditrepository "learnflow_backend/internal/audit/repository"
	authdomain "learnflow_backend/internal/auth/domain"
	authrepository "learnflow_backend/internal/auth/repository"
	"learnflow_backend/internal/events"
	"learnflow_backend/internal/infrastructure/db"
	redisinfra "learnflow_backend/internal/infrastructure/redis"
	appcontext "learnflow_backend/internal/shared/context"
	"learnflow_backend/internal/shared/testutil"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/justinas/alice"
	"github.com/redis/go-redis/v9"
	. "github.com/smartystreets/goconvey/convey"
)

type httpIntegrationFixture struct {
	pool     *pgxpool.Pool
	mux      *http.ServeMux
	adminID  string
	targetID string
}

// newHTTPIntegrationFixture wires the real handler, service, Postgres and Redis behind the admin routes;
// the chain stands in for JWT + RequireRole by putting a fixed admin into the request context.
func newHTTPIntegrationFixture(t *testing.T) *httpIntegrationFixture {
	t.Helper()

	pool := testutil.NewTestPool(t)
	realRedis := redisinfra.NewInstance(redis.NewClient(&redis.Options{Addr: "localhost:6379"}))
	t.Cleanup(func() { realRedis.Close() }) //nolint:errcheck // Close's error is never actionable in test cleanup

	ctx := context.Background()
	adminID := testutil.InsertTestUser(t, pool, testutil.RandomTestEmail(t, "http-admin"))
	targetID := testutil.InsertTestUser(t, pool, testutil.RandomTestEmail(t, "http-target"))
	if _, err := pool.Exec(ctx, "UPDATE users SET role = 'admin' WHERE id = $1", adminID); err != nil {
		t.Fatalf("promote admin: %v", err)
	}

	t.Cleanup(func() {
		realRedis.UnBlockUser(ctx, targetID)          //nolint:errcheck // best-effort cleanup
		realRedis.ClearUserRoleRevoked(ctx, targetID) //nolint:errcheck // best-effort cleanup
		testutil.DeleteAdminActionsByTarget(t, pool, targetID)
		for _, q := range []string{
			"DELETE FROM event_outbox WHERE aggregate_id = $1",
			"DELETE FROM user_sessions WHERE user_id = $1",
			"DELETE FROM user_profiles WHERE user_id = $1",
			"DELETE FROM users WHERE id = $1",
		} {
			pool.Exec(ctx, q, targetID) //nolint:errcheck // best-effort cleanup of throwaway test data
		}
		pool.Exec(ctx, "DELETE FROM users WHERE id = $1", adminID) //nolint:errcheck // best-effort cleanup
	})

	adminRepo := adminrepository.NewRepository(pool)
	svc := adminservice.New(
		adminservice.Repos{AnnounRepo: adminRepo, UserRepo: adminRepo, ActionRepo: auditrepository.New(pool), SessionRepo: authrepository.NewRepository(pool)},
		adminservice.Utils{Transactor: db.NewTransactor(pool), Outbox: events.NewOutboxWriter(pool), Blocklist: realRedis},
	)

	asAdmin := func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			user := &authdomain.User{ID: adminID, Role: authdomain.RoleAdmin}
			next.ServeHTTP(w, r.WithContext(appcontext.WithUser(r.Context(), user)))
		})
	}

	mux := http.NewServeMux()
	adminhttp.NewHTTPHandler(svc, testutil.NewTestLogger()).RegisterRoutes(mux, alice.New(asAdmin), alice.New())

	return &httpIntegrationFixture{pool: pool, mux: mux, adminID: adminID, targetID: targetID}
}

func (f *httpIntegrationFixture) do(method, path string) *httptest.ResponseRecorder {
	return testutil.ServeHTTP(f.mux, httptest.NewRequestWithContext(context.Background(), method, path, nil))
}

func TestAdminUserRoutes_Integration(t *testing.T) {
	Convey("Given the admin routes on real Postgres and Redis", t, func() {
		f := newHTTPIntegrationFixture(t)
		userPath := "/api/v1/admin/users/" + f.targetID

		Convey("PUT block blocks the user once and answers 409 with a fixed message the second time", func() {
			So(f.do(http.MethodPut, userPath+"/block").Code, ShouldEqual, http.StatusOK)

			second := f.do(http.MethodPut, userPath+"/block")

			So(second.Code, ShouldEqual, http.StatusConflict)
			body := decodeBody(t, second.Body.Bytes())
			So(body["error"], ShouldEqual, "user is not in a valid state for this action")
		})

		Convey("PUT block for an unknown user answers 404", func() {
			w := f.do(http.MethodPut, "/api/v1/admin/users/00000000-0000-0000-0000-000000000000/block")

			So(w.Code, ShouldEqual, http.StatusNotFound)
		})

		Convey("DELETE subadmin for a regular user answers 409", func() {
			So(f.do(http.MethodDelete, userPath+"/subadmin").Code, ShouldEqual, http.StatusConflict)
		})

		Convey("PUT subadmin then DELETE subadmin round-trips", func() {
			So(f.do(http.MethodPut, userPath+"/subadmin").Code, ShouldEqual, http.StatusOK)
			So(f.do(http.MethodDelete, userPath+"/subadmin").Code, ShouldEqual, http.StatusOK)
		})

		Convey("GET audit/actions returns the block written by the PUT, with the admin's name", func() {
			So(f.do(http.MethodPut, userPath+"/block").Code, ShouldEqual, http.StatusOK)

			w := f.do(http.MethodGet, "/api/v1/admin/audit/actions?target_type=user&target_id="+f.targetID)

			So(w.Code, ShouldEqual, http.StatusOK)
			body := decodeBody(t, w.Body.Bytes())
			So(body["total"], ShouldEqual, 1)
			actions, ok := body["actions"].([]any)
			So(ok, ShouldBeTrue)
			So(actions, ShouldHaveLength, 1)
			first, ok := actions[0].(map[string]any)
			So(ok, ShouldBeTrue)
			So(first["action_type"], ShouldEqual, "block_user")
			So(first["admin_user_id"], ShouldEqual, f.adminID)
			So(first["admin_name"], ShouldNotBeEmpty)
		})

		Convey("GET dlq answers 200 with a failed_jobs list", func() {
			w := f.do(http.MethodGet, "/api/v1/admin/dlq")

			So(w.Code, ShouldEqual, http.StatusOK)
			So(decodeBody(t, w.Body.Bytes()), ShouldContainKey, "failed_jobs")
		})
	})
}

func (f *httpIntegrationFixture) setActorRole(t *testing.T, role string) {
	t.Helper()

	if _, err := f.pool.Exec(context.Background(), "UPDATE users SET role = $1 WHERE id = $2", role, f.adminID); err != nil {
		t.Fatalf("set actor role: %v", err)
	}
}

func TestAdminJournalRoute_Integration(t *testing.T) {
	Convey("Given the audit journal route on real Postgres and Redis", t, func() {
		f := newHTTPIntegrationFixture(t)
		So(f.do(http.MethodPut, "/api/v1/admin/users/"+f.targetID+"/block").Code, ShouldEqual, http.StatusOK)
		So(f.do(http.MethodPut, "/api/v1/admin/users/"+f.targetID+"/unblock").Code, ShouldEqual, http.StatusOK)

		Convey("An admin filters the journal by admin and action type and gets exactly that action", func() {
			w := f.do(http.MethodGet, "/api/v1/admin/audit?admin_user_id="+f.adminID+"&action_type=unblock_user")

			So(w.Code, ShouldEqual, http.StatusOK)
			body := decodeBody(t, w.Body.Bytes())
			So(body["total"], ShouldEqual, 1)
			actions, ok := body["actions"].([]any)
			So(ok, ShouldBeTrue)
			first, ok := actions[0].(map[string]any)
			So(ok, ShouldBeTrue)
			So(first["action_type"], ShouldEqual, "unblock_user")
			So(first["target_id"], ShouldEqual, f.targetID)
			So(first["admin_name"], ShouldNotBeEmpty)
		})

		Convey("An admin sees both actions of the admin and none outside the date range", func() {
			all := decodeBody(t, f.do(http.MethodGet, "/api/v1/admin/audit?admin_user_id="+f.adminID).Body.Bytes())
			So(all["total"], ShouldEqual, 2)

			future := decodeBody(t, f.do(http.MethodGet, "/api/v1/admin/audit?admin_user_id="+f.adminID+"&from=2999-01-01T00:00:00Z").Body.Bytes())
			So(future["total"], ShouldEqual, 0)
		})

		Convey("An invalid filter answers 400", func() {
			So(f.do(http.MethodGet, "/api/v1/admin/audit?action_type=nope").Code, ShouldEqual, http.StatusBadRequest)
			So(f.do(http.MethodGet, "/api/v1/admin/audit?from=2026-10-02T00:00:00Z&to=2026-10-01T00:00:00Z").Code, ShouldEqual, http.StatusBadRequest)
		})

		Convey("A subadmin is forbidden from the journal, because the DB role is checked, not the token", func() {
			f.setActorRole(t, "subadmin")

			So(f.do(http.MethodGet, "/api/v1/admin/audit").Code, ShouldEqual, http.StatusForbidden)
		})
	})
}

func TestSubadminAuditTrailAccess_Integration(t *testing.T) {
	Convey("Given a subadmin reading the per-target audit trail", t, func() {
		f := newHTTPIntegrationFixture(t)
		f.setActorRole(t, "subadmin")

		Convey("The trail of a plain user is readable", func() {
			So(f.do(http.MethodGet, "/api/v1/admin/audit/actions?target_type=user&target_id="+f.targetID).Code, ShouldEqual, http.StatusOK)
		})

		Convey("The trail of an admin is forbidden", func() {
			f.promoteTarget(t, "admin")

			So(f.do(http.MethodGet, "/api/v1/admin/audit/actions?target_type=user&target_id="+f.targetID).Code, ShouldEqual, http.StatusForbidden)
		})

		Convey("The trail of another subadmin is forbidden", func() {
			f.promoteTarget(t, "subadmin")

			So(f.do(http.MethodGet, "/api/v1/admin/audit/actions?target_type=user&target_id="+f.targetID).Code, ShouldEqual, http.StatusForbidden)
		})

		Convey("The trail of a non-user target needs no role check", func() {
			w := f.do(http.MethodGet, "/api/v1/admin/audit/actions?target_type=course&target_id=00000000-0000-0000-0000-000000000000")

			So(w.Code, ShouldEqual, http.StatusOK)
		})
	})
}

func (f *httpIntegrationFixture) promoteTarget(t *testing.T, role string) {
	t.Helper()

	if _, err := f.pool.Exec(context.Background(), "UPDATE users SET role = $1 WHERE id = $2", role, f.targetID); err != nil {
		t.Fatalf("set target role: %v", err)
	}
}
