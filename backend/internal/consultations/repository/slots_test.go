package consultationsrepository

import (
	"context"
	"strconv"
	"strings"
	"testing"
	"time"

	consultationsdomain "learnflow_backend/internal/consultations/domain"
	"learnflow_backend/internal/shared/testutil"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	. "github.com/smartystreets/goconvey/convey"
)

var (
	rangeFrom = time.Date(2026, 11, 1, 0, 0, 0, 0, time.UTC)
	rangeTo   = time.Date(2026, 11, 7, 0, 0, 0, 0, time.UTC)
)

func countRow(total int) *testutil.MockRow {
	return &testutil.MockRow{ScanFn: func(dest ...any) error {
		*testutil.CastInt(dest[0], 0) = total
		return nil
	}}
}

func TestGetAvailableSlots(t *testing.T) {
	Convey("Given a consultations repository", t, func() {
		var countErr error
		var queryErr error
		var rows *testutil.MockRows
		var gotQuery string
		var gotArgs []any
		repo := newTestRepo(&testutil.MockQueryRunner{
			QueryRowFn: func(_ context.Context, _ string, _ ...any) pgx.Row {
				if countErr != nil {
					return &testutil.MockRow{ScanFn: func(_ ...any) error { return countErr }}
				}
				return countRow(2)
			},
			QueryFn: func(_ context.Context, query string, args ...any) (pgx.Rows, error) {
				gotQuery, gotArgs = query, args
				if queryErr != nil {
					return nil, queryErr
				}
				return rows, nil
			},
		})

		Convey("When the count fails", func() {
			countErr = testutil.ErrDBUnexpected

			_, _, err := repo.GetAvailableSlots(context.Background(), rangeFrom, rangeTo, nil)

			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldContainSubstring, "repository.GetAvailableSlots count")
		})

		Convey("When the query fails", func() {
			queryErr = testutil.ErrDBUnexpected

			_, _, err := repo.GetAvailableSlots(context.Background(), rangeFrom, rangeTo, nil)

			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldContainSubstring, "repository.GetAvailableSlots")
		})

		Convey("When two slots are returned, they are scanned in order and the range and no consultant are passed", func() {
			first, second := fakePublicSlot(1), fakePublicSlot(2)
			rows = &testutil.MockRows{Rows: []*testutil.MockRow{
				{ScanFn: fakePublicSlotScan(first)},
				{ScanFn: fakePublicSlotScan(second)},
			}}

			got, total, err := repo.GetAvailableSlots(context.Background(), rangeFrom, rangeTo, nil)

			So(err, ShouldBeNil)
			So(total, ShouldEqual, 2)
			So(got, ShouldResemble, []*consultationsdomain.PublicSlot{first, second})
			So(gotArgs, ShouldResemble, []any{rangeFrom, rangeTo, (*string)(nil)})
			So(gotQuery, ShouldContainSubstring, "status = 'available'")
		})
	})
}

func TestGetAvailableSlotsRows(t *testing.T) {
	Convey("Given a consultations repository returning rows", t, func() {
		var rows *testutil.MockRows
		repo := newTestRepo(&testutil.MockQueryRunner{
			QueryRowFn: func(_ context.Context, _ string, _ ...any) pgx.Row { return countRow(1) },
			QueryFn: func(_ context.Context, _ string, _ ...any) (pgx.Rows, error) {
				return rows, nil
			},
		})

		Convey("When a row fails to scan", func() {
			rows = &testutil.MockRows{Rows: []*testutil.MockRow{{ScanFn: func(_ ...any) error { return testutil.ErrDBUnexpected }}}}

			_, _, err := repo.GetAvailableSlots(context.Background(), rangeFrom, rangeTo, nil)

			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldContainSubstring, "scan")
		})

		Convey("When rows.Err() reports a failure", func() {
			rows = &testutil.MockRows{RowsErr: testutil.ErrDBUnexpected}

			_, _, err := repo.GetAvailableSlots(context.Background(), rangeFrom, rangeTo, nil)

			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldContainSubstring, "rows")
		})
	})
}

func TestGetAllSlots(t *testing.T) {
	Convey("Given a consultations repository", t, func() {
		var rows *testutil.MockRows
		var gotArgs []any
		repo := newTestRepo(&testutil.MockQueryRunner{
			QueryRowFn: func(_ context.Context, _ string, _ ...any) pgx.Row { return countRow(2) },
			QueryFn: func(_ context.Context, _ string, args ...any) (pgx.Rows, error) {
				gotArgs = args
				return rows, nil
			},
		})

		Convey("When a booked slot has a holder, the booking is set; a free slot has none", func() {
			free, booked := fakeSlot(1), fakeSlot(2)
			booked.Status = consultationsdomain.SlotStatusBooked
			briefID, userID := "brief-1", "user-1"
			rows = &testutil.MockRows{Rows: []*testutil.MockRow{
				{ScanFn: fakeAdminSlotScan(free, nil, nil)},
				{ScanFn: fakeAdminSlotScan(booked, &briefID, &userID)},
			}}

			got, total, err := repo.GetAllSlots(context.Background(), rangeFrom, rangeTo, nil)

			So(err, ShouldBeNil)
			So(total, ShouldEqual, 2)
			So(got, ShouldHaveLength, 2)
			So(got[0].Booking, ShouldBeNil)
			So(got[1].Booking, ShouldResemble, &consultationsdomain.SlotBooking{BriefID: "brief-1", UserID: "user-1"})
			So(gotArgs, ShouldResemble, []any{rangeFrom, rangeTo, (*string)(nil)})
		})
	})
}

func TestGetSlotsByIDsForUpdate(t *testing.T) {
	Convey("Given a consultations repository", t, func() {
		var queryErr error
		var rows *testutil.MockRows
		var gotQuery string
		var gotArgs []any
		repo := newTestRepo(&testutil.MockQueryRunner{
			QueryFn: func(_ context.Context, query string, args ...any) (pgx.Rows, error) {
				gotQuery, gotArgs = query, args
				if queryErr != nil {
					return nil, queryErr
				}
				return rows, nil
			},
		})
		ids := []string{"slot-1", "slot-2"}

		Convey("When the query fails", func() {
			queryErr = testutil.ErrDBUnexpected

			_, err := repo.GetSlotsByIDsForUpdate(context.Background(), ids)

			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldContainSubstring, "repository.GetSlotsByIDsForUpdate")
		})

		Convey("When slots are found, ids are passed as one array arg and the rows are locked", func() {
			first, second := fakeSlot(1), fakeSlot(2)
			rows = &testutil.MockRows{Rows: []*testutil.MockRow{
				{ScanFn: fakeSlotScan(first)},
				{ScanFn: fakeSlotScan(second)},
			}}

			got, err := repo.GetSlotsByIDsForUpdate(context.Background(), ids)

			So(err, ShouldBeNil)
			So(got, ShouldResemble, []*consultationsdomain.Slot{first, second})
			So(gotArgs, ShouldResemble, []any{ids})
			So(strings.Contains(gotQuery, "FOR UPDATE"), ShouldBeTrue)
		})
	})
}

func TestUpdateSlotsStatus(t *testing.T) {
	Convey("Given a consultations repository", t, func() {
		var execErr error
		var affected int64
		var execCalls int
		var gotQuery string
		var gotArgs []any
		repo := newTestRepo(&testutil.MockQueryRunner{
			ExecFn: func(_ context.Context, query string, args ...any) (pgconn.CommandTag, error) {
				execCalls++
				gotQuery, gotArgs = query, args
				if execErr != nil {
					return pgconn.CommandTag{}, execErr
				}
				return pgconn.NewCommandTag("UPDATE " + strconv.FormatInt(affected, 10)), nil
			},
		})
		ids := []string{"slot-1", "slot-2"}

		Convey("When there is nothing to update, no statement runs", func() {
			err := repo.UpdateSlotsStatus(context.Background(), nil, consultationsdomain.SlotStatusBlocked, "admin-1")

			So(err, ShouldBeNil)
			So(execCalls, ShouldEqual, 0)
		})

		Convey("When every slot is updated, status, admin and ids are passed in order and booked slots are guarded", func() {
			affected = 2

			err := repo.UpdateSlotsStatus(context.Background(), ids, consultationsdomain.SlotStatusBlocked, "admin-1")

			So(err, ShouldBeNil)
			So(gotArgs, ShouldResemble, []any{consultationsdomain.SlotStatusBlocked, "admin-1", ids})
			So(gotQuery, ShouldContainSubstring, "status <> 'booked'")
		})

		Convey("When fewer rows changed than ids given", func() {
			affected = 1

			err := repo.UpdateSlotsStatus(context.Background(), ids, consultationsdomain.SlotStatusAvailable, "admin-1")

			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldContainSubstring, "updated 1 of 2")
		})

		Convey("When the statement fails", func() {
			execErr = testutil.ErrDBUnexpected

			err := repo.UpdateSlotsStatus(context.Background(), ids, consultationsdomain.SlotStatusBlocked, "admin-1")

			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldContainSubstring, "repository.UpdateSlotsStatus")
		})
	})
}

func TestGetAvailableSlotsConsultantFilter(t *testing.T) {
	Convey("Given a consultations repository", t, func() {
		var gotArgs []any
		repo := newTestRepo(&testutil.MockQueryRunner{
			QueryRowFn: func(_ context.Context, _ string, _ ...any) pgx.Row { return countRow(0) },
			QueryFn: func(_ context.Context, _ string, args ...any) (pgx.Rows, error) {
				gotArgs = args
				return &testutil.MockRows{}, nil
			},
		})

		Convey("When a consultant is given, it is passed as the third arg", func() {
			consultantID := "consultant-1"

			_, _, err := repo.GetAvailableSlots(context.Background(), rangeFrom, rangeTo, &consultantID)

			So(err, ShouldBeNil)
			So(gotArgs, ShouldResemble, []any{rangeFrom, rangeTo, &consultantID})
		})
	})
}

func TestCreateSlotsIfMissing(t *testing.T) {
	Convey("Given a consultations repository", t, func() {
		var execErr error
		var execCalls int
		var gotArgs []any
		repo := newTestRepo(&testutil.MockQueryRunner{
			ExecFn: func(_ context.Context, _ string, args ...any) (pgconn.CommandTag, error) {
				execCalls++
				gotArgs = args
				if execErr != nil {
					return pgconn.CommandTag{}, execErr
				}
				return pgconn.NewCommandTag("INSERT 0 3"), nil
			},
		})
		dates := []time.Time{rangeFrom, rangeFrom, rangeTo}
		starts := []string{"09:00", "11:00", "09:00"}

		Convey("When there is nothing to create, no statement runs", func() {
			err := repo.CreateSlotsIfMissing(context.Background(), "consultant-1", nil, nil, 120)

			So(err, ShouldBeNil)
			So(execCalls, ShouldEqual, 0)
		})

		Convey("When dates and start times differ in length", func() {
			err := repo.CreateSlotsIfMissing(context.Background(), "consultant-1", dates, starts[:2], 120)

			So(err, ShouldNotBeNil)
			So(execCalls, ShouldEqual, 0)
		})

		Convey("When slots are inserted, args are passed in order", func() {
			err := repo.CreateSlotsIfMissing(context.Background(), "consultant-1", dates, starts, 120)

			So(err, ShouldBeNil)
			So(gotArgs, ShouldResemble, []any{"consultant-1", dates, starts, 120})
		})

		Convey("When the statement fails", func() {
			execErr = testutil.ErrDBUnexpected

			err := repo.CreateSlotsIfMissing(context.Background(), "consultant-1", dates, starts, 120)

			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldContainSubstring, "repository.CreateSlotsIfMissing")
		})
	})
}

func TestGetConsultants(t *testing.T) {
	Convey("Given a consultations repository", t, func() {
		var queryErr error
		var rows *testutil.MockRows
		var gotQuery string
		repo := newTestRepo(&testutil.MockQueryRunner{
			QueryFn: func(_ context.Context, query string, _ ...any) (pgx.Rows, error) {
				gotQuery = query
				if queryErr != nil {
					return nil, queryErr
				}
				return rows, nil
			},
		})

		Convey("When the query fails", func() {
			queryErr = testutil.ErrDBUnexpected

			_, err := repo.GetConsultants(context.Background())

			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldContainSubstring, "repository.GetConsultants")
		})

		Convey("When a row fails to scan", func() {
			rows = &testutil.MockRows{Rows: []*testutil.MockRow{{ScanFn: func(_ ...any) error { return testutil.ErrDBUnexpected }}}}

			_, err := repo.GetConsultants(context.Background())

			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldContainSubstring, "scan")
		})

		Convey("When consultants are found, only live active admins are queried", func() {
			scan := func(c *consultationsdomain.Consultant) func(dest ...any) error {
				return func(dest ...any) error {
					*testutil.CastStr(dest[0], 0) = c.ID
					*testutil.CastStr(dest[1], 1) = c.FirstName
					*testutil.CastStr(dest[2], 2) = c.LastName
					*testutil.CastStr(dest[3], 3) = c.AvatarURL
					*testutil.CastStr(dest[4], 4) = c.Bio
					return nil
				}
			}
			first := &consultationsdomain.Consultant{ID: "a", FirstName: "Anna", LastName: "K", AvatarURL: "u", Bio: "b"}
			second := &consultationsdomain.Consultant{ID: "b"}
			rows = &testutil.MockRows{Rows: []*testutil.MockRow{{ScanFn: scan(first)}, {ScanFn: scan(second)}}}

			got, err := repo.GetConsultants(context.Background())

			So(err, ShouldBeNil)
			So(got, ShouldResemble, []*consultationsdomain.Consultant{first, second})
			So(gotQuery, ShouldContainSubstring, "role = 'admin'")
			So(gotQuery, ShouldContainSubstring, "status = 'active'")
			So(gotQuery, ShouldContainSubstring, "deleted_at IS NULL")
		})
	})
}
