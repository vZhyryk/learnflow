package consultationsdomain_test

import (
	"strings"
	"testing"
	"time"

	. "github.com/smartystreets/goconvey/convey"

	consultationsdomain "learnflow_backend/internal/consultations/domain"
)

const (
	slotID1 = "11111111-1111-4111-8111-111111111111"
	slotID2 = "22222222-2222-4222-8222-222222222222"
)

func TestSlotStatusIsValid(t *testing.T) {
	Convey("Given slot statuses", t, func() {
		Convey("When the status is known", func() {
			So(consultationsdomain.SlotStatusAvailable.IsValid(), ShouldBeTrue)
			So(consultationsdomain.SlotStatusBooked.IsValid(), ShouldBeTrue)
			So(consultationsdomain.SlotStatusBlocked.IsValid(), ShouldBeTrue)
		})

		Convey("When the status is unknown", func() {
			So(consultationsdomain.SlotStatus("hold").IsValid(), ShouldBeFalse)
			So(consultationsdomain.SlotStatus("").IsValid(), ShouldBeFalse)
		})
	})
}

func TestParseSlotRange(t *testing.T) {
	Convey("Given from/to query values", t, func() {
		Convey("When it is a valid single day", func() {
			got, err := consultationsdomain.ParseSlotRange("2026-11-02", "2026-11-02")

			So(err, ShouldBeNil)
			So(got.From.Equal(got.To), ShouldBeTrue)
		})

		Convey("When the range is exactly the maximum", func() {
			_, err := consultationsdomain.ParseSlotRange("2026-11-01", "2027-01-02")

			So(err, ShouldBeNil)
		})

		Convey("When the range is one day over the maximum", func() {
			_, err := consultationsdomain.ParseSlotRange("2026-11-01", "2027-01-03")

			So(err, ShouldEqual, consultationsdomain.ErrDateRangeTooLong)
		})

		Convey("When to is before from", func() {
			_, err := consultationsdomain.ParseSlotRange("2026-11-05", "2026-11-04")

			So(err, ShouldEqual, consultationsdomain.ErrInvalidDateRange)
		})

		Convey("When a value is not a date", func() {
			_, err := consultationsdomain.ParseSlotRange("2026-11-01", "tomorrow")
			So(err, ShouldEqual, consultationsdomain.ErrInvalidDateRange)

			_, err = consultationsdomain.ParseSlotRange("", "2026-11-01")
			So(err, ShouldEqual, consultationsdomain.ErrInvalidDateRange)
		})
	})
}

func TestUpdateSlotsRequestValidate(t *testing.T) {
	Convey("Given a bulk slot update request", t, func() {
		valid := consultationsdomain.UpdateSlotsRequest{Slots: []consultationsdomain.SlotStatusChange{
			{SlotID: slotID1, Status: consultationsdomain.SlotStatusBlocked},
			{SlotID: slotID2, Status: consultationsdomain.SlotStatusAvailable},
		}}

		Convey("When it is valid", func() {
			So(valid.Validate(), ShouldBeNil)
		})

		Convey("When no slots are provided", func() {
			req := consultationsdomain.UpdateSlotsRequest{}

			So(req.Validate(), ShouldEqual, consultationsdomain.ErrEmptySlotUpdate)
		})

		Convey("When the batch is over the limit", func() {
			changes := make([]consultationsdomain.SlotStatusChange, consultationsdomain.MaxSlotUpdates+1)
			for i := range changes {
				changes[i] = consultationsdomain.SlotStatusChange{SlotID: slotID1, Status: consultationsdomain.SlotStatusBlocked}
			}
			req := consultationsdomain.UpdateSlotsRequest{Slots: changes}

			So(req.Validate(), ShouldEqual, consultationsdomain.ErrTooManySlots)
		})

		Convey("When a slot ID is not a UUID", func() {
			req := consultationsdomain.UpdateSlotsRequest{Slots: []consultationsdomain.SlotStatusChange{
				{SlotID: "not-a-uuid", Status: consultationsdomain.SlotStatusBlocked},
			}}

			So(req.Validate(), ShouldEqual, consultationsdomain.ErrInvalidSlotID)
		})

		Convey("When the requested status is booked or unknown", func() {
			for _, status := range []consultationsdomain.SlotStatus{consultationsdomain.SlotStatusBooked, "hold", ""} {
				req := consultationsdomain.UpdateSlotsRequest{Slots: []consultationsdomain.SlotStatusChange{
					{SlotID: slotID1, Status: status},
				}}

				So(req.Validate(), ShouldEqual, consultationsdomain.ErrInvalidSlotStatus)
			}
		})

		Convey("When a slot ID repeats", func() {
			req := consultationsdomain.UpdateSlotsRequest{Slots: []consultationsdomain.SlotStatusChange{
				{SlotID: slotID1, Status: consultationsdomain.SlotStatusBlocked},
				{SlotID: strings.ToUpper(slotID1), Status: consultationsdomain.SlotStatusAvailable},
			}}

			So(req.Validate(), ShouldEqual, consultationsdomain.ErrDuplicateSlotID)
		})
	})
}

func TestSlotScheduleSlotsBetween(t *testing.T) {
	Convey("Given the default Warsaw schedule", t, func() {
		warsaw := time.FixedZone("Europe/Warsaw", 3600)
		schedule := consultationsdomain.DefaultSlotSchedule(warsaw)

		Convey("When a Monday is requested", func() {
			day := time.Date(2026, 11, 2, 0, 0, 0, 0, time.UTC)

			dates, starts := schedule.SlotsBetween(day, day)

			So(starts, ShouldResemble, []string{"09:00", "11:00", "13:00", "15:00", "17:00"})
			So(dates, ShouldHaveLength, 5)
			for _, date := range dates {
				So(date, ShouldEqual, day)
			}
		})

		Convey("When a Sunday is requested", func() {
			sunday := time.Date(2026, 11, 1, 0, 0, 0, 0, time.UTC)

			dates, starts := schedule.SlotsBetween(sunday, sunday)

			So(dates, ShouldBeEmpty)
			So(starts, ShouldBeEmpty)
		})

		Convey("When a whole week is requested", func() {
			monday := time.Date(2026, 11, 2, 0, 0, 0, 0, time.UTC)
			nextSunday := time.Date(2026, 11, 8, 0, 0, 0, 0, time.UTC)

			dates, starts := schedule.SlotsBetween(monday, nextSunday)

			So(dates, ShouldHaveLength, 6*5)
			So(starts, ShouldHaveLength, 6*5)
		})

		Convey("When from and to carry a non-UTC clock time, only the date counts", func() {
			from := time.Date(2026, 11, 2, 23, 30, 0, 0, warsaw)

			dates, _ := schedule.SlotsBetween(from, from)

			So(dates, ShouldHaveLength, 5)
			So(dates[0], ShouldEqual, time.Date(2026, 11, 2, 0, 0, 0, 0, time.UTC))
		})

		Convey("When the range is reversed", func() {
			from := time.Date(2026, 11, 5, 0, 0, 0, 0, time.UTC)
			to := time.Date(2026, 11, 2, 0, 0, 0, 0, time.UTC)

			dates, starts := schedule.SlotsBetween(from, to)

			So(dates, ShouldBeEmpty)
			So(starts, ShouldBeEmpty)
		})
	})

	Convey("Given a schedule with a non-positive duration", t, func() {
		schedule := consultationsdomain.SlotSchedule{DayStartMinutes: 540, DayEndMinutes: 1140}
		day := time.Date(2026, 11, 2, 0, 0, 0, 0, time.UTC)

		dates, starts := schedule.SlotsBetween(day, day)

		So(dates, ShouldBeNil)
		So(starts, ShouldBeNil)
	})
}

func validScheduleConfig() consultationsdomain.SlotScheduleConfig {
	return consultationsdomain.SlotScheduleConfig{
		TimeZone: "Europe/Warsaw", WorkingDays: "Mon,Sat", DayStart: "09:00", DayEnd: "13:00", DurationMinutes: 120, HorizonDays: 30,
	}
}

func TestParseSlotSchedule(t *testing.T) {
	Convey("Given a valid config", t, func() {
		Convey("When it is parsed", func() {
			schedule, err := consultationsdomain.ParseSlotSchedule(validScheduleConfig())

			So(err, ShouldBeNil)
			So(schedule.Weekdays, ShouldResemble, []time.Weekday{time.Monday, time.Saturday})
			So(schedule.DayStartMinutes, ShouldEqual, 540)
			So(schedule.DayEndMinutes, ShouldEqual, 780)
			So(schedule.HorizonDays, ShouldEqual, 30)

			dates, starts := schedule.SlotsBetween(time.Date(2026, 11, 2, 0, 0, 0, 0, time.UTC), time.Date(2026, 11, 2, 0, 0, 0, 0, time.UTC))
			So(dates, ShouldHaveLength, 2)
			So(starts, ShouldResemble, []string{"09:00", "11:00"})
		})
	})
}

func TestParseSlotScheduleInvalid(t *testing.T) {
	Convey("Given an invalid config", t, func() {
		mutations := map[string]func(*consultationsdomain.SlotScheduleConfig){
			"unknown time zone":  func(c *consultationsdomain.SlotScheduleConfig) { c.TimeZone = "Mars/Olympus" },
			"empty working days": func(c *consultationsdomain.SlotScheduleConfig) { c.WorkingDays = "" },
			"unknown day":        func(c *consultationsdomain.SlotScheduleConfig) { c.WorkingDays = "Mon,Xyz" },
			"repeated day":       func(c *consultationsdomain.SlotScheduleConfig) { c.WorkingDays = "Mon,mon" },
			"bad day start":      func(c *consultationsdomain.SlotScheduleConfig) { c.DayStart = "9" },
			"bad day end":        func(c *consultationsdomain.SlotScheduleConfig) { c.DayEnd = "24:30" },
			"end before start":   func(c *consultationsdomain.SlotScheduleConfig) { c.DayEnd = "08:00" },
			"zero duration":      func(c *consultationsdomain.SlotScheduleConfig) { c.DurationMinutes = 0 },
			"duration too long":  func(c *consultationsdomain.SlotScheduleConfig) { c.DurationMinutes = 300 },
			"zero horizon":       func(c *consultationsdomain.SlotScheduleConfig) { c.HorizonDays = 0 },
			"horizon over the cap": func(c *consultationsdomain.SlotScheduleConfig) {
				c.HorizonDays = consultationsdomain.MaxSlotHorizonDays + 1
			},
		}

		for name, mutate := range mutations {
			Convey("When there is a "+name, func() {
				cfg := validScheduleConfig()
				mutate(&cfg)

				_, err := consultationsdomain.ParseSlotSchedule(cfg)

				So(err, ShouldNotBeNil)
			})
		}
	})
}
