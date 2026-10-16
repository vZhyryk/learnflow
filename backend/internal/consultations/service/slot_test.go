package consultationservice

import (
	"context"
	"strings"
	"testing"
	"time"

	. "github.com/smartystreets/goconvey/convey"

	auditdomain "learnflow_backend/internal/audit/domain"
	consultationsdomain "learnflow_backend/internal/consultations/domain"
	"learnflow_backend/internal/shared/testutil"
)

const (
	slotA = "11111111-1111-4111-8111-111111111111"
	slotB = "22222222-2222-4222-8222-222222222222"
	slotC = "33333333-3333-4333-8333-333333333333"
	admin = "aaaaaaaa-aaaa-4aaa-8aaa-aaaaaaaaaaaa"
)

func TestGetSlots(t *testing.T) {
	Convey("Given slot listings", t, func() {
		slotRange := consultationsdomain.SlotRange{
			From: time.Date(2026, 11, 2, 0, 0, 0, 0, time.UTC),
			To:   time.Date(2026, 11, 8, 0, 0, 0, 0, time.UTC),
		}

		Convey("When the repository returns available slots", func() {
			want := []*consultationsdomain.PublicSlot{{ID: slotA}, {ID: slotB}}
			svc := newTestService(&mockSlotRepo{
				getAvailableSlots: func(_ context.Context, from, to time.Time, _ *string) ([]*consultationsdomain.PublicSlot, int, error) {
					So(from, ShouldEqual, slotRange.From)
					So(to, ShouldEqual, slotRange.To)
					return want, 2, nil
				},
			}, nil)

			slots, count, err := svc.GetAvailableSlots(context.Background(), slotRange, nil)

			So(err, ShouldBeNil)
			So(count, ShouldEqual, 2)
			So(slots, ShouldResemble, want)
		})

		Convey("When listing available slots fails", func() {
			svc := newTestService(&mockSlotRepo{
				getAvailableSlots: func(context.Context, time.Time, time.Time, *string) ([]*consultationsdomain.PublicSlot, int, error) {
					return nil, 0, testutil.ErrDBUnexpected
				},
			}, nil)

			_, _, err := svc.GetAvailableSlots(context.Background(), slotRange, nil)

			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldContainSubstring, "service.GetAvailableSlots")
		})

		Convey("When the consultant filter is set it reaches the repository", func() {
			consultant := admin
			var got *string
			svc := newTestService(&mockSlotRepo{
				getAllSlots: func(_ context.Context, _, _ time.Time, consultantID *string) ([]*consultationsdomain.Slot, int, error) {
					got = consultantID
					return nil, 0, nil
				},
			}, nil)

			_, _, err := svc.GetAllSlots(context.Background(), slotRange, &consultant)

			So(err, ShouldBeNil)
			So(got, ShouldNotBeNil)
			So(*got, ShouldEqual, admin)
		})

		Convey("When listing all slots fails", func() {
			svc := newTestService(&mockSlotRepo{
				getAllSlots: func(context.Context, time.Time, time.Time, *string) ([]*consultationsdomain.Slot, int, error) {
					return nil, 0, testutil.ErrDBUnexpected
				},
			}, nil)

			_, _, err := svc.GetAllSlots(context.Background(), slotRange, nil)

			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldContainSubstring, "service.GetAllSlots")
		})
	})
}

type updateFixture struct {
	svc      *Service
	updates  map[consultationsdomain.SlotStatus][]string
	updater  string
	audited  []*auditdomain.AdminAction
	lockedBy []string
}

func newUpdateFixture(stored []*consultationsdomain.Slot) *updateFixture {
	fx := &updateFixture{updates: make(map[consultationsdomain.SlotStatus][]string)}
	fx.svc = newTestService(&mockSlotRepo{
		getSlotsByIDsForUpdate: func(_ context.Context, ids []string) ([]*consultationsdomain.Slot, error) {
			fx.lockedBy = ids
			return stored, nil
		},
		updateSlotsStatus: func(_ context.Context, ids []string, status consultationsdomain.SlotStatus, updatedBy string) error {
			fx.updates[status] = ids
			fx.updater = updatedBy
			return nil
		},
	}, &mockActionRepo{
		createAdminAction: func(_ context.Context, action *auditdomain.AdminAction) error {
			fx.audited = append(fx.audited, action)
			return nil
		},
	})

	return fx
}

func change(id string, status consultationsdomain.SlotStatus) consultationsdomain.SlotStatusChange {
	return consultationsdomain.SlotStatusChange{SlotID: id, Status: status}
}

func TestUpdateSlotsOutcomes(t *testing.T) {
	Convey("Given stored slots", t, func() {
		stored := []*consultationsdomain.Slot{
			{ID: slotA, Status: consultationsdomain.SlotStatusAvailable},
			{ID: slotB, Status: consultationsdomain.SlotStatusBooked},
			{ID: slotC, Status: consultationsdomain.SlotStatusBlocked},
		}
		fx := newUpdateFixture(stored)

		Convey("When available, booked and unknown slots are requested", func() {
			unknown := "44444444-4444-4444-8444-444444444444"
			req := consultationsdomain.UpdateSlotsRequest{AdminUserID: admin, Slots: []consultationsdomain.SlotStatusChange{
				change(slotA, consultationsdomain.SlotStatusBlocked),
				change(slotB, consultationsdomain.SlotStatusBlocked),
				change(unknown, consultationsdomain.SlotStatusAvailable),
			}}

			res, err := fx.svc.UpdateSlots(context.Background(), req)

			So(err, ShouldBeNil)
			So(res.Applied, ShouldResemble, []string{slotA})
			So(res.Skipped, ShouldResemble, []consultationsdomain.SkippedSlot{
				{SlotID: slotB, Reason: consultationsdomain.SkipReasonBooked},
				{SlotID: unknown, Reason: consultationsdomain.SkipReasonNotFound},
			})
			So(fx.updates, ShouldResemble, map[consultationsdomain.SlotStatus][]string{
				consultationsdomain.SlotStatusBlocked: {slotA},
			})
			So(fx.updater, ShouldEqual, admin)
			So(fx.lockedBy, ShouldResemble, []string{slotA, slotB, unknown})
		})

		Convey("When a slot is already in the requested status", func() {
			req := consultationsdomain.UpdateSlotsRequest{AdminUserID: admin, Slots: []consultationsdomain.SlotStatusChange{
				change(slotC, consultationsdomain.SlotStatusBlocked),
			}}

			res, err := fx.svc.UpdateSlots(context.Background(), req)

			So(err, ShouldBeNil)
			So(res.Applied, ShouldResemble, []string{slotC})
			So(res.Skipped, ShouldBeEmpty)
			So(fx.updates, ShouldBeEmpty)
			So(fx.audited, ShouldBeEmpty)
		})

		Convey("When the request ID differs from the stored ID only by case", func() {
			upper := strings.ToUpper(slotA)
			req := consultationsdomain.UpdateSlotsRequest{AdminUserID: admin, Slots: []consultationsdomain.SlotStatusChange{
				change(upper, consultationsdomain.SlotStatusBlocked),
			}}

			res, err := fx.svc.UpdateSlots(context.Background(), req)

			So(err, ShouldBeNil)
			So(res.Applied, ShouldResemble, []string{upper})
			So(res.Skipped, ShouldBeEmpty)
		})

		Convey("When nothing is applied or skipped the lists are empty, not nil", func() {
			req := consultationsdomain.UpdateSlotsRequest{AdminUserID: admin}

			res, err := fx.svc.UpdateSlots(context.Background(), req)

			So(err, ShouldBeNil)
			So(res.Applied, ShouldNotBeNil)
			So(res.Skipped, ShouldNotBeNil)
		})
	})
}

func TestUpdateSlotsAudit(t *testing.T) {
	Convey("Given stored slots", t, func() {
		fx := newUpdateFixture([]*consultationsdomain.Slot{
			{ID: slotA, Status: consultationsdomain.SlotStatusAvailable},
			{ID: slotB, Status: consultationsdomain.SlotStatusBooked},
			{ID: slotC, Status: consultationsdomain.SlotStatusBlocked},
		})

		Convey("When slots are changed, one audit entry per changed slot is written", func() {
			req := consultationsdomain.UpdateSlotsRequest{AdminUserID: admin, Slots: []consultationsdomain.SlotStatusChange{
				change(slotA, consultationsdomain.SlotStatusBlocked),
				change(slotC, consultationsdomain.SlotStatusAvailable),
				change(slotB, consultationsdomain.SlotStatusBlocked),
			}}

			_, err := fx.svc.UpdateSlots(context.Background(), req)

			So(err, ShouldBeNil)
			So(fx.audited, ShouldHaveLength, 2)
			targets := []string{fx.audited[0].TargetID, fx.audited[1].TargetID}
			So(targets, ShouldContain, slotA)
			So(targets, ShouldContain, slotC)
			for _, action := range fx.audited {
				So(action.AdminUserID, ShouldEqual, admin)
				So(action.TargetType, ShouldEqual, auditdomain.TargetConsultationSlot)
				So(action.ActionType, ShouldEqual, auditdomain.ActionUpdateItem)
			}
		})
	})
}

func TestUpdateSlotsErrors(t *testing.T) {
	Convey("Given a request that changes one slot", t, func() {
		stored := []*consultationsdomain.Slot{{ID: slotA, Status: consultationsdomain.SlotStatusAvailable}}
		req := consultationsdomain.UpdateSlotsRequest{AdminUserID: admin, Slots: []consultationsdomain.SlotStatusChange{
			change(slotA, consultationsdomain.SlotStatusBlocked),
		}}

		Convey("When locking the slots fails", func() {
			svc := newTestService(&mockSlotRepo{
				getSlotsByIDsForUpdate: func(context.Context, []string) ([]*consultationsdomain.Slot, error) {
					return nil, testutil.ErrDBUnexpected
				},
			}, nil)

			res, err := svc.UpdateSlots(context.Background(), req)

			So(res, ShouldBeNil)
			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldContainSubstring, "lock slots")
		})

		Convey("When updating the status fails", func() {
			svc := newTestService(&mockSlotRepo{
				getSlotsByIDsForUpdate: func(context.Context, []string) ([]*consultationsdomain.Slot, error) { return stored, nil },
				updateSlotsStatus: func(context.Context, []string, consultationsdomain.SlotStatus, string) error {
					return testutil.ErrDBUnexpected
				},
			}, nil)

			res, err := svc.UpdateSlots(context.Background(), req)

			So(res, ShouldBeNil)
			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldContainSubstring, "set status")
		})

		Convey("When writing the audit entry fails", func() {
			svc := newTestService(&mockSlotRepo{
				getSlotsByIDsForUpdate: func(context.Context, []string) ([]*consultationsdomain.Slot, error) { return stored, nil },
				updateSlotsStatus: func(context.Context, []string, consultationsdomain.SlotStatus, string) error {
					return nil
				},
			}, &mockActionRepo{
				createAdminAction: func(context.Context, *auditdomain.AdminAction) error { return testutil.ErrDBUnexpected },
			})

			res, err := svc.UpdateSlots(context.Background(), req)

			So(res, ShouldBeNil)
			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldContainSubstring, "audit slot")
		})
	})
}

func TestGenerateSlots(t *testing.T) {
	Convey("Given two consultants", t, func() {
		consultants := []*consultationsdomain.Consultant{{ID: admin}, {ID: "bbbbbbbb-bbbb-4bbb-8bbb-bbbbbbbbbbbb"}}

		Convey("When slots are generated", func() {
			created := make(map[string]int)
			var gotDuration int
			svc := newTestService(&mockSlotRepo{
				getConsultants: func(context.Context) ([]*consultationsdomain.Consultant, error) { return consultants, nil },
				createSlotsIfMissing: func(_ context.Context, id string, dates []time.Time, starts []string, duration int) error {
					So(dates, ShouldHaveLength, len(starts))
					So(dates, ShouldNotBeEmpty)
					created[id] = len(dates)
					gotDuration = duration
					return nil
				},
			}, nil)

			err := svc.GenerateSlots(context.Background())

			So(err, ShouldBeNil)
			So(created, ShouldHaveLength, 2)
			So(created[consultants[0].ID], ShouldEqual, created[consultants[1].ID])
			So(gotDuration, ShouldEqual, consultationsdomain.DurationMinutes)
		})

		Convey("When loading consultants fails", func() {
			svc := newTestService(&mockSlotRepo{
				getConsultants: func(context.Context) ([]*consultationsdomain.Consultant, error) {
					return nil, testutil.ErrDBUnexpected
				},
			}, nil)

			err := svc.GenerateSlots(context.Background())

			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldContainSubstring, "get consultants")
		})

		Convey("When creating slots fails the cause and consultant are in the error", func() {
			svc := newTestService(&mockSlotRepo{
				getConsultants: func(context.Context) ([]*consultationsdomain.Consultant, error) { return consultants, nil },
				createSlotsIfMissing: func(context.Context, string, []time.Time, []string, int) error {
					return testutil.ErrDBUnexpected
				},
			}, nil)

			err := svc.GenerateSlots(context.Background())

			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldContainSubstring, consultants[0].ID)
			So(err.Error(), ShouldContainSubstring, testutil.ErrDBUnexpected.Error())
		})
	})
}

func TestGetConsultants(t *testing.T) {
	Convey("Given the consultants list", t, func() {
		Convey("When the repository returns consultants", func() {
			want := []*consultationsdomain.Consultant{{ID: admin, FirstName: "Anna"}}
			svc := newTestService(&mockSlotRepo{
				getConsultants: func(context.Context) ([]*consultationsdomain.Consultant, error) { return want, nil },
			}, nil)

			got, err := svc.GetConsultants(context.Background())

			So(err, ShouldBeNil)
			So(got, ShouldResemble, want)
		})

		Convey("When the repository fails", func() {
			svc := newTestService(&mockSlotRepo{
				getConsultants: func(context.Context) ([]*consultationsdomain.Consultant, error) {
					return nil, testutil.ErrDBUnexpected
				},
			}, nil)

			_, err := svc.GetConsultants(context.Background())

			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldContainSubstring, "service.GetConsultants")
		})
	})
}
