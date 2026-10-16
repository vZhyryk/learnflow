package consultations_test

import (
	"testing"
	"time"

	. "github.com/smartystreets/goconvey/convey"

	"learnflow_backend/internal/consultations"
)

func TestLoadSlotSchedule(t *testing.T) {
	Convey("Given no SLOT_* variables", t, func() {
		Convey("When the schedule is loaded, the defaults apply", func() {
			schedule, err := consultations.LoadSlotSchedule()

			So(err, ShouldBeNil)
			So(schedule.Location.String(), ShouldEqual, "Europe/Warsaw")
			So(schedule.Weekdays, ShouldHaveLength, 6)
			So(schedule.DayStartMinutes, ShouldEqual, 9*60)
			So(schedule.DayEndMinutes, ShouldEqual, 19*60)
			So(schedule.DurationMinutes, ShouldEqual, 120)
			So(schedule.HorizonDays, ShouldEqual, 60)
		})
	})

	Convey("Given custom SLOT_* variables", t, func() {
		t.Setenv("SLOT_TIME_ZONE", "UTC")
		t.Setenv("SLOT_WORKING_DAYS", "mon, WED")
		t.Setenv("SLOT_DAY_START", "10:00")
		t.Setenv("SLOT_DAY_END", "14:00")
		t.Setenv("SLOT_DURATION_MINUTES", "60")
		t.Setenv("SLOT_HORIZON_DAYS", "30")

		Convey("When the schedule is loaded", func() {
			schedule, err := consultations.LoadSlotSchedule()

			So(err, ShouldBeNil)
			So(schedule.Weekdays, ShouldResemble, []time.Weekday{time.Monday, time.Wednesday})
			So(schedule.DayStartMinutes, ShouldEqual, 600)
			So(schedule.DayEndMinutes, ShouldEqual, 840)
			So(schedule.DurationMinutes, ShouldEqual, 60)
			So(schedule.HorizonDays, ShouldEqual, 30)
		})
	})

	Convey("Given an invalid SLOT_* variable", t, func() {
		cases := map[string]string{
			"SLOT_TIME_ZONE":        "Mars/Olympus",
			"SLOT_WORKING_DAYS":     "Mon,Funday",
			"SLOT_DAY_START":        "9am",
			"SLOT_DAY_END":          "25:00",
			"SLOT_DURATION_MINUTES": "700",
			"SLOT_HORIZON_DAYS":     "9999",
		}

		for key, value := range cases {
			Convey("When "+key+" is "+value, func() {
				t.Setenv(key, value)

				_, err := consultations.LoadSlotSchedule()

				So(err, ShouldNotBeNil)
				So(err.Error(), ShouldContainSubstring, "consultations.LoadSlotSchedule")
			})
		}
	})
}
