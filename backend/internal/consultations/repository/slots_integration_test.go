//go:build integration

package consultationsrepository

import (
	"context"
	"testing"
	"time"

	consultationsdomain "learnflow_backend/internal/consultations/domain"
	"learnflow_backend/internal/infrastructure/db"
	"learnflow_backend/internal/shared/repository"
	"learnflow_backend/internal/shared/testutil"

	"github.com/jackc/pgx/v5"
	. "github.com/smartystreets/goconvey/convey"
)

func newIntegrationRepo(tx pgx.Tx) *Repository {
	return &Repository{repository.BaseRepository{DB: tx}}
}

// insertSlot stores a slot far in the future so it cannot collide with other tests' dates.
func insertSlot(ctx context.Context, tx pgx.Tx, consultantID, date, startTime, status string) string {
	var id string
	err := tx.QueryRow(ctx,
		`INSERT INTO consultation_slots (consultant_id, date, start_time, duration_minutes, status) VALUES ($1, $2, $3, 120, $4) RETURNING id`,
		consultantID, date, startTime, status).Scan(&id)
	So(err, ShouldBeNil)

	return id
}

func day(s string) time.Time {
	d, err := time.Parse("2006-01-02", s)
	So(err, ShouldBeNil)

	return d
}

func TestSlotsRepository_Integration(t *testing.T) {
	pool := testutil.NewTestPool(t)

	Convey("Given a consultations repository backed by real Postgres", t, func() {
		Convey("GetAvailableSlots returns only free slots of the range, ordered by date and time", func() {
			testutil.WithTestTx(t, pool, func(ctx context.Context, tx pgx.Tx) {
				repo := newIntegrationRepo(tx)
				consultantID := testutil.InsertRandomTestUser(t, tx)
				late := insertSlot(ctx, tx, consultantID, "2031-05-02", "11:00", "available")
				early := insertSlot(ctx, tx, consultantID, "2031-05-01", "09:00", "available")
				insertSlot(ctx, tx, consultantID, "2031-05-01", "11:00", "blocked")
				insertSlot(ctx, tx, consultantID, "2031-05-01", "13:00", "booked")
				insertSlot(ctx, tx, consultantID, "2031-06-01", "09:00", "available")

				got, total, err := repo.GetAvailableSlots(ctx, day("2031-05-01"), day("2031-05-31"), nil)

				So(err, ShouldBeNil)
				So(total, ShouldEqual, 2)
				So(got, ShouldHaveLength, 2)
				So(got[0].ID, ShouldEqual, early)
				So(got[0].StartTime, ShouldEqual, "09:00")
				So(got[0].DurationMinutes, ShouldEqual, 120)
				So(got[1].ID, ShouldEqual, late)
			})
		})

		Convey("GetAllSlots lists every status and attaches the live brief to a booked slot only", func() {
			testutil.WithTestTx(t, pool, func(ctx context.Context, tx pgx.Tx) {
				repo := newIntegrationRepo(tx)
				consultantID := testutil.InsertRandomTestUser(t, tx)
				userID := testutil.InsertRandomTestUser(t, tx)
				free := insertSlot(ctx, tx, consultantID, "2031-05-01", "09:00", "available")
				booked := insertSlot(ctx, tx, consultantID, "2031-05-01", "11:00", "booked")
				insertSlot(ctx, tx, consultantID, "2031-05-01", "13:00", "blocked")
				var briefID string
				So(tx.QueryRow(ctx,
					`INSERT INTO consultation_briefs (user_id, slot_id, status) VALUES ($1, $2, 'booked') RETURNING id`,
					userID, booked).Scan(&briefID), ShouldBeNil)
				_, err := tx.Exec(ctx,
					`INSERT INTO consultation_briefs (user_id, slot_id, status) VALUES ($1, $2, 'failed')`, userID, free)
				So(err, ShouldBeNil)

				got, total, err := repo.GetAllSlots(ctx, day("2031-05-01"), day("2031-05-01"), nil)

				So(err, ShouldBeNil)
				So(total, ShouldEqual, 3)
				So(got, ShouldHaveLength, 3)
				So(got[0].ID, ShouldEqual, free)
				So(got[0].Booking, ShouldBeNil)
				So(got[1].Status, ShouldEqual, consultationsdomain.SlotStatusBooked)
				So(got[1].Booking, ShouldResemble, &consultationsdomain.SlotBooking{BriefID: briefID, UserID: userID})
				So(got[2].Booking, ShouldBeNil)
			})
		})

		Convey("UpdateSlotsStatus blocks slots, records the admin, and refuses to touch a booked slot", func() {
			testutil.WithTestTx(t, pool, func(ctx context.Context, tx pgx.Tx) {
				repo := newIntegrationRepo(tx)
				consultantID := testutil.InsertRandomTestUser(t, tx)
				adminID := testutil.InsertRandomTestUser(t, tx)
				first := insertSlot(ctx, tx, consultantID, "2031-05-01", "09:00", "available")
				second := insertSlot(ctx, tx, consultantID, "2031-05-01", "11:00", "available")
				booked := insertSlot(ctx, tx, consultantID, "2031-05-01", "13:00", "booked")

				err := repo.UpdateSlotsStatus(ctx, []string{first, second}, consultationsdomain.SlotStatusBlocked, adminID)
				So(err, ShouldBeNil)

				locked, err := repo.GetSlotsByIDsForUpdate(ctx, []string{second, first})
				So(err, ShouldBeNil)
				So(locked, ShouldHaveLength, 2)
				for _, slot := range locked {
					So(slot.Status, ShouldEqual, consultationsdomain.SlotStatusBlocked)
					So(slot.UpdatedBy, ShouldNotBeNil)
					So(*slot.UpdatedBy, ShouldEqual, adminID)
				}

				err = repo.UpdateSlotsStatus(ctx, []string{booked}, consultationsdomain.SlotStatusBlocked, adminID)
				So(err, ShouldNotBeNil)
				So(err.Error(), ShouldContainSubstring, "updated 0 of 1")
			})
		})

		Convey("A slot cannot hold two live briefs, but can be booked again after a failed one", func() {
			testutil.WithTestTx(t, pool, func(ctx context.Context, tx pgx.Tx) {
				userID := testutil.InsertRandomTestUser(t, tx)
				consultantID := testutil.InsertRandomTestUser(t, tx)
				slot := insertSlot(ctx, tx, consultantID, "2031-05-01", "09:00", "booked")
				insertBrief := func(status string) error {
					_, err := tx.Exec(ctx,
						`INSERT INTO consultation_briefs (user_id, slot_id, status) VALUES ($1, $2, $3)`, userID, slot, status)
					return err
				}

				So(insertBrief("failed"), ShouldBeNil)
				So(insertBrief("submitted"), ShouldBeNil)

				_, err := tx.Exec(ctx, "SAVEPOINT second_live")
				So(err, ShouldBeNil)
				err = insertBrief("booked")
				So(err, ShouldNotBeNil)
				So(db.IsUniqueViolation(err, "idx_consultation_briefs_slot_id_active_unique"), ShouldBeTrue)
			})
		})

		Convey("Two consultants can hold the same time, and the consultant filter narrows the list", func() {
			testutil.WithTestTx(t, pool, func(ctx context.Context, tx pgx.Tx) {
				repo := newIntegrationRepo(tx)
				first := testutil.InsertRandomTestUser(t, tx)
				second := testutil.InsertRandomTestUser(t, tx)
				firstSlot := insertSlot(ctx, tx, first, "2031-05-01", "09:00", "available")
				insertSlot(ctx, tx, second, "2031-05-01", "09:00", "available")

				all, total, err := repo.GetAvailableSlots(ctx, day("2031-05-01"), day("2031-05-01"), nil)
				So(err, ShouldBeNil)
				So(total, ShouldEqual, 2)
				So(all, ShouldHaveLength, 2)

				mine, total, err := repo.GetAvailableSlots(ctx, day("2031-05-01"), day("2031-05-01"), &first)
				So(err, ShouldBeNil)
				So(total, ShouldEqual, 1)
				So(mine, ShouldHaveLength, 1)
				So(mine[0].ID, ShouldEqual, firstSlot)
				So(mine[0].ConsultantID, ShouldEqual, first)
			})
		})

		Convey("CreateSlotsIfMissing is idempotent and keeps blocked slots untouched", func() {
			testutil.WithTestTx(t, pool, func(ctx context.Context, tx pgx.Tx) {
				repo := newIntegrationRepo(tx)
				consultantID := testutil.InsertRandomTestUser(t, tx)
				blocked := insertSlot(ctx, tx, consultantID, "2031-05-01", "09:00", "blocked")
				dates := []time.Time{day("2031-05-01"), day("2031-05-01"), day("2031-05-02")}
				starts := []string{"09:00", "11:00", "09:00"}

				created, err := repo.CreateSlotsIfMissing(ctx, consultantID, dates, starts, 120)
				So(err, ShouldBeNil)
				So(created, ShouldEqual, 2)

				again, err := repo.CreateSlotsIfMissing(ctx, consultantID, dates, starts, 120)
				So(err, ShouldBeNil)
				So(again, ShouldEqual, 0)

				got, total, err := repo.GetAllSlots(ctx, day("2031-05-01"), day("2031-05-02"), &consultantID)
				So(err, ShouldBeNil)
				So(total, ShouldEqual, 3)
				So(got, ShouldHaveLength, 3)
				So(got[0].ID, ShouldEqual, blocked)
				So(got[0].Status, ShouldEqual, consultationsdomain.SlotStatusBlocked)
				So(got[1].Status, ShouldEqual, consultationsdomain.SlotStatusAvailable)
				So(got[1].UpdatedBy, ShouldBeNil)
			})
		})

		Convey("GetConsultantsByIDs returns only active admins and subadmins with their names", func() {
			testutil.WithTestTx(t, pool, func(ctx context.Context, tx pgx.Tx) {
				repo := newIntegrationRepo(tx)
				admin := testutil.InsertRandomTestUser(t, tx)
				plainUser := testutil.InsertRandomTestUser(t, tx)
				blocked := testutil.InsertRandomTestUser(t, tx)
				_, err := tx.Exec(ctx, `UPDATE users SET role = 'admin' WHERE id = $1`, admin)
				So(err, ShouldBeNil)
				_, err = tx.Exec(ctx, `UPDATE users SET role = 'subadmin', status = 'blocked' WHERE id = $1`, blocked)
				So(err, ShouldBeNil)
				_, err = tx.Exec(ctx,
					`INSERT INTO user_profiles (user_id, first_name, last_name) VALUES ($1, 'Anna', 'Nowak')
					 ON CONFLICT (user_id) DO UPDATE SET first_name = 'Anna', last_name = 'Nowak'`, admin)
				So(err, ShouldBeNil)

				got, err := repo.GetConsultantsByIDs(ctx, []string{admin, plainUser, blocked})

				So(err, ShouldBeNil)
				So(got, ShouldHaveLength, 1)
				So(got[0], ShouldResemble, &consultationsdomain.Consultant{ID: admin, FirstName: "Anna", LastName: "Nowak"})
			})
		})

		Convey("GetSlotsByIDsForUpdate ignores unknown ids", func() {
			testutil.WithTestTx(t, pool, func(ctx context.Context, tx pgx.Tx) {
				repo := newIntegrationRepo(tx)
				consultantID := testutil.InsertRandomTestUser(t, tx)
				known := insertSlot(ctx, tx, consultantID, "2031-05-01", "09:00", "available")

				got, err := repo.GetSlotsByIDsForUpdate(ctx, []string{known, "99999999-9999-4999-8999-999999999999"})

				So(err, ShouldBeNil)
				So(got, ShouldHaveLength, 1)
				So(got[0].ID, ShouldEqual, known)
			})
		})
	})
}
