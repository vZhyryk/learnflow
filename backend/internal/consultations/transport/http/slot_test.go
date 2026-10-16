package consultationshttp_test

import (
	"context"
	"net/http"
	"testing"

	consultationsdomain "learnflow_backend/internal/consultations/domain"
	"learnflow_backend/internal/shared/testutil"

	. "github.com/smartystreets/goconvey/convey"
)

const consultantUUID = "bbbbbbbb-bbbb-4bbb-8bbb-bbbbbbbbbbbb"

func TestGetAvailableSlotsRoute(t *testing.T) {
	Convey("GET "+slotPath, t, func() {
		var gotRange consultationsdomain.SlotRange
		var gotConsultant *string
		var svcErr error
		svc := &mockService{getAvailableSlots: func(_ context.Context, r consultationsdomain.SlotRange, id *string) ([]*consultationsdomain.PublicSlot, int, error) {
			gotRange, gotConsultant = r, id
			return []*consultationsdomain.PublicSlot{{ID: "s1"}}, 1, svcErr
		}}
		f := newHTTPFixture(svc, http.MethodGet, slotPath)
		serve := func(params map[string]string) int { return testutil.ServeHTTP(f.Mux, f.NewReq("", params)).Code }

		Convey("A valid range without a consultant → 200 and no consultant filter", func() {
			w := testutil.ServeHTTP(f.Mux, f.NewReq("", map[string]string{"from": validRange, "to": validRange}))

			So(w.Code, ShouldEqual, http.StatusOK)
			So(gotConsultant, ShouldBeNil)
			So(gotRange.From.Format("2006-01-02"), ShouldEqual, validRange)
			body := testutil.DecodeBody(t, w.Body.Bytes())
			So(body["slots"], ShouldHaveLength, 1)
			So(body["total"], ShouldEqual, 1)
		})

		Convey("A consultant_id is passed to the service", func() {
			code := serve(map[string]string{"from": validRange, "to": validRange, "consultant_id": consultantUUID})

			So(code, ShouldEqual, http.StatusOK)
			So(gotConsultant, ShouldNotBeNil)
			So(*gotConsultant, ShouldEqual, consultantUUID)
		})

		Convey("A consultant_id that is not a UUID → 422", func() {
			So(serve(map[string]string{"from": validRange, "to": validRange, "consultant_id": "x"}), ShouldEqual, http.StatusUnprocessableEntity)
		})

		Convey("Missing or reversed dates → 422", func() {
			So(serve(nil), ShouldEqual, http.StatusUnprocessableEntity)
			So(serve(map[string]string{"from": "2026-11-05", "to": "2026-11-02"}), ShouldEqual, http.StatusUnprocessableEntity)
		})

		Convey("A range over the limit → 422", func() {
			So(serve(map[string]string{"from": "2026-01-01", "to": "2026-12-31"}), ShouldEqual, http.StatusUnprocessableEntity)
		})

		Convey("An unexpected service error → 500", func() {
			svcErr = testutil.ErrDBUnexpected
			So(serve(map[string]string{"from": validRange, "to": validRange}), ShouldEqual, http.StatusInternalServerError)
		})
	})
}

func TestGetAllSlotsRoute(t *testing.T) {
	Convey("GET "+adminPath, t, func() {
		var gotConsultant *string
		svc := &mockService{getAllSlots: func(_ context.Context, _ consultationsdomain.SlotRange, id *string) ([]*consultationsdomain.Slot, int, error) {
			gotConsultant = id
			return []*consultationsdomain.Slot{{ID: "s1", Status: consultationsdomain.SlotStatusBooked}}, 1, nil
		}}
		f := newHTTPFixture(svc, http.MethodGet, adminPath)

		Convey("A valid range → 200 with all slots", func() {
			w := testutil.ServeHTTP(f.Mux, f.NewReq("", map[string]string{"from": validRange, "to": validRange, "consultant_id": consultantUUID}))

			So(w.Code, ShouldEqual, http.StatusOK)
			So(gotConsultant, ShouldNotBeNil)
			So(testutil.DecodeBody(t, w.Body.Bytes())["slots"], ShouldHaveLength, 1)
		})

		Convey("Missing dates → 422", func() {
			So(testutil.ServeHTTP(f.Mux, f.NewReq("", nil)).Code, ShouldEqual, http.StatusUnprocessableEntity)
		})
	})
}

func TestUpdateSlotsRoute(t *testing.T) {
	Convey("PATCH "+adminPath, t, func() {
		var gotReq consultationsdomain.UpdateSlotsRequest
		var svcErr error
		svc := &mockService{updateSlots: func(_ context.Context, req consultationsdomain.UpdateSlotsRequest) (*consultationsdomain.UpdateSlotsResult, error) {
			gotReq = req
			return &consultationsdomain.UpdateSlotsResult{Applied: []string{"s1"}, Skipped: []consultationsdomain.SkippedSlot{}}, svcErr
		}}
		f := newHTTPFixture(svc, http.MethodPatch, adminPath)
		body := `{"slots":[{"slot_id":"11111111-1111-4111-8111-111111111111","status":"blocked"}]}`

		Convey("A valid batch → 200 and the admin comes from the session", func() {
			w := testutil.ServeHTTP(f.Mux, withAdmin(f.NewReq(body, nil)))

			So(w.Code, ShouldEqual, http.StatusOK)
			So(gotReq.AdminUserID, ShouldEqual, adminID)
			So(gotReq.Slots, ShouldHaveLength, 1)
			result, ok := testutil.DecodeBody(t, w.Body.Bytes())["result"].(map[string]any)
			So(ok, ShouldBeTrue)
			So(result["applied"], ShouldHaveLength, 1)
		})

		Convey("An invalid batch (booked status requested) → 4xx, service not called", func() {
			bad := `{"slots":[{"slot_id":"11111111-1111-4111-8111-111111111111","status":"booked"}]}`

			code := testutil.ServeHTTP(f.Mux, withAdmin(f.NewReq(bad, nil))).Code

			So(code, ShouldBeBetweenOrEqual, 400, 422)
			So(gotReq.AdminUserID, ShouldBeEmpty)
		})

		Convey("Malformed JSON → 400", func() {
			So(testutil.ServeHTTP(f.Mux, withAdmin(f.NewReq(`{`, nil))).Code, ShouldEqual, http.StatusBadRequest)
		})

		Convey("An unexpected service error → 500", func() {
			svcErr = testutil.ErrDBUnexpected
			So(testutil.ServeHTTP(f.Mux, withAdmin(f.NewReq(body, nil))).Code, ShouldEqual, http.StatusInternalServerError)
		})
	})
}

func TestGetConsultantsRoute(t *testing.T) {
	Convey("GET "+consPath, t, func() {
		var svcErr error
		svc := &mockService{getConsultants: func(context.Context) ([]*consultationsdomain.Consultant, error) {
			return []*consultationsdomain.Consultant{{ID: "c1", FirstName: "Anna", Bio: "b"}}, svcErr
		}}
		f := newHTTPFixture(svc, http.MethodGet, consPath)

		Convey("Consultants are returned under the consultants key with bio", func() {
			w := testutil.ServeHTTP(f.Mux, f.NewReq("", nil))

			So(w.Code, ShouldEqual, http.StatusOK)
			list, ok := testutil.DecodeBody(t, w.Body.Bytes())["consultants"].([]any)
			So(ok, ShouldBeTrue)
			So(list, ShouldHaveLength, 1)
			first, ok := list[0].(map[string]any)
			So(ok, ShouldBeTrue)
			So(first["bio"], ShouldEqual, "b")
		})

		Convey("An unexpected service error → 500", func() {
			svcErr = testutil.ErrDBUnexpected
			So(testutil.ServeHTTP(f.Mux, f.NewReq("", nil)).Code, ShouldEqual, http.StatusInternalServerError)
		})
	})
}
