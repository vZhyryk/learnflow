package consultationsdomain

import (
	"fmt"
	"learnflow_backend/internal/shared/validator"
	"strings"
	"time"
)

// SlotStatus is the lifecycle state of a consultation slot.
type SlotStatus string

// Slot statuses; they match the consultation_slots_status_check constraint.
const (
	SlotStatusAvailable SlotStatus = "available"
	SlotStatusBooked    SlotStatus = "booked"
	SlotStatusBlocked   SlotStatus = "blocked"
)

// Limits of the slots API.
const (
	// MaxSlotRangeDays is the longest from..to range a slot listing accepts.
	MaxSlotRangeDays = 62
	// MaxSlotUpdates is the largest batch accepted by a bulk slot status update.
	MaxSlotUpdates = 200
	// DurationMinutes is the default slot length.
	DurationMinutes = 120
	// SlotHorizonDays is the default number of days ahead the generator keeps slots.
	SlotHorizonDays = 60
	DayStartHours   = 9
	DayEndHours     = 19
	// SlotTimeZone is the default zone of the slot schedule.
	SlotTimeZone = "Europe/Warsaw"
)

// Reasons a slot is skipped by a bulk status update.
const (
	SkipReasonBooked   = "booked"
	SkipReasonNotFound = "not_found"
)

// Date and clock layouts used by the slots API.
const (
	slotDateLayout     = "2006-01-02"
	SlotTimeLayout     = "15:04"
	SlotDateTimeLayout = "2006-01-02:15:04"
)

// IsValid reports whether the status is one of the known slot statuses.
func (st SlotStatus) IsValid() bool {
	switch st {
	case SlotStatusAvailable, SlotStatusBooked, SlotStatusBlocked:
		return true
	default:
		return false
	}
}

// Slot is one bookable consultation time of one consultant. Date and StartTime are the local (Europe/Warsaw) wall clock.
type Slot struct {
	ID              string     `json:"id"`
	ConsultantID    string     `json:"consultant_id"`
	Date            time.Time  `json:"date"`
	StartTime       string     `json:"start_time"`
	DurationMinutes int        `json:"duration_minutes"`
	Status          SlotStatus `json:"status"`
	UpdatedBy       *string    `json:"updated_by"`
	CreatedAt       time.Time  `json:"created_at"`
	UpdatedAt       time.Time  `json:"updated_at"`
	// Booking is set only in the admin calendar, for slots held by a brief.
	Booking *SlotBooking `json:"booking,omitempty"`
}

// PublicSlot is what a user sees when browsing free slots.
type PublicSlot struct {
	ID              string     `json:"id"`
	ConsultantID    string     `json:"consultant_id"`
	Date            time.Time  `json:"date"`
	StartTime       string     `json:"start_time"`
	DurationMinutes int        `json:"duration_minutes"`
	Status          SlotStatus `json:"status"`
}

// Consultant is an admin a client can book a consultation with; profile fields are empty when not filled in.
type Consultant struct {
	ID        string `json:"id"`
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
	AvatarURL string `json:"avatar_url"`
	Bio       string `json:"bio"`
}

// SlotBooking identifies who holds a booked slot, for the admin calendar.
type SlotBooking struct {
	BriefID string `json:"brief_id"`
	UserID  string `json:"user_id"`
}

// SlotRange is an inclusive date range for listing slots.
type SlotRange struct {
	From time.Time
	To   time.Time
}

// ParseSlotRange parses from/to query values (YYYY-MM-DD) and checks the range.
func ParseSlotRange(from, to string) (SlotRange, error) {
	fromDate, err := time.Parse(slotDateLayout, from)
	if err != nil {
		return SlotRange{}, ErrInvalidDateRange
	}

	toDate, err := time.Parse(slotDateLayout, to)
	if err != nil {
		return SlotRange{}, ErrInvalidDateRange
	}

	if toDate.Before(fromDate) {
		return SlotRange{}, ErrInvalidDateRange
	}

	if toDate.Sub(fromDate) > MaxSlotRangeDays*24*time.Hour {
		return SlotRange{}, ErrDateRangeTooLong
	}

	return SlotRange{From: fromDate, To: toDate}, nil
}

// SlotStatusChange asks for one slot to be set to Status.
type SlotStatusChange struct {
	SlotID string     `json:"slot_id"`
	Status SlotStatus `json:"status"`
}

// UpdateSlotsRequest is the input of a bulk slot status update; AdminUserID comes from the session.
type UpdateSlotsRequest struct {
	AdminUserID string             `json:"-"`
	Slots       []SlotStatusChange `json:"slots"`
}

// Validate checks the batch size, IDs, duplicates and that only blocked/available can be requested.
func (req *UpdateSlotsRequest) Validate() error {
	if len(req.Slots) == 0 {
		return ErrEmptySlotUpdate
	}

	if len(req.Slots) > MaxSlotUpdates {
		return ErrTooManySlots
	}

	seen := make(map[string]struct{}, len(req.Slots))
	for _, change := range req.Slots {
		if !validator.IsValidUUID(change.SlotID) {
			return ErrInvalidSlotID
		}

		if change.Status != SlotStatusBlocked && change.Status != SlotStatusAvailable {
			return ErrInvalidSlotStatus
		}

		key := strings.ToLower(change.SlotID)
		if _, dup := seen[key]; dup {
			return ErrDuplicateSlotID
		}
		seen[key] = struct{}{}
	}

	return nil
}

// GetIDList returns the slot IDs of the request in order.
func (req *UpdateSlotsRequest) GetIDList() []string {
	ids := make([]string, 0, len(req.Slots))
	for _, slot := range req.Slots {
		ids = append(ids, slot.SlotID)
	}

	return ids
}

// SkippedSlot is a slot a bulk update did not change, with the reason.
type SkippedSlot struct {
	SlotID string `json:"slot_id"`
	Reason string `json:"reason"`
}

// UpdateSlotsResult reports which slots a bulk update changed and which it skipped.
type UpdateSlotsResult struct {
	Applied []string      `json:"applied"`
	Skipped []SkippedSlot `json:"skipped"`
}

// SlotSchedule describes when slots exist and how far ahead they are generated.
type SlotSchedule struct {
	Location        *time.Location
	Weekdays        []time.Weekday
	DayStartMinutes int
	DayEndMinutes   int
	DurationMinutes int
	HorizonDays     int
}

// DefaultSlotSchedule is Monday–Saturday, 09:00–19:00, 120-minute slots, in loc.
func DefaultSlotSchedule(loc *time.Location) SlotSchedule {
	return SlotSchedule{
		Location: loc,
		Weekdays: []time.Weekday{
			time.Monday, time.Tuesday, time.Wednesday, time.Thursday, time.Friday, time.Saturday,
		},
		DayStartMinutes: DayStartHours * 60,
		DayEndMinutes:   DayEndHours * 60,
		DurationMinutes: DurationMinutes,
		HorizonDays:     SlotHorizonDays,
	}
}

// SlotScheduleConfig is the raw, unparsed schedule as read from configuration.
type SlotScheduleConfig struct {
	TimeZone        string
	WorkingDays     string
	DayStart        string
	DayEnd          string
	DurationMinutes int
	HorizonDays     int
}

// MaxSlotHorizonDays caps how far ahead slots may be generated.
const MaxSlotHorizonDays = 365

var weekdayByName = map[string]time.Weekday{
	"mon": time.Monday, "tue": time.Tuesday, "wed": time.Wednesday, "thu": time.Thursday,
	"fri": time.Friday, "sat": time.Saturday, "sun": time.Sunday,
}

// ParseSlotSchedule validates cfg and builds the schedule: working days are three-letter names (Mon,Tue,...),
// times are HH:MM, and at least one slot must fit into a working day.
func ParseSlotSchedule(cfg SlotScheduleConfig) (SlotSchedule, error) {
	loc, err := time.LoadLocation(cfg.TimeZone)
	if err != nil {
		return SlotSchedule{}, fmt.Errorf("slot schedule: time zone %q: %w", cfg.TimeZone, err)
	}

	weekdays, err := parseWeekdays(cfg.WorkingDays)
	if err != nil {
		return SlotSchedule{}, err
	}

	start, err := parseClockMinutes(cfg.DayStart)
	if err != nil {
		return SlotSchedule{}, fmt.Errorf("slot schedule: day start: %w", err)
	}

	end, err := parseClockMinutes(cfg.DayEnd)
	if err != nil {
		return SlotSchedule{}, fmt.Errorf("slot schedule: day end: %w", err)
	}

	if cfg.DurationMinutes <= 0 || start+cfg.DurationMinutes > end {
		return SlotSchedule{}, fmt.Errorf("slot schedule: duration %d min does not fit into %s-%s", cfg.DurationMinutes, cfg.DayStart, cfg.DayEnd)
	}

	if cfg.HorizonDays < 1 || cfg.HorizonDays > MaxSlotHorizonDays {
		return SlotSchedule{}, fmt.Errorf("slot schedule: horizon %d days is outside 1..%d", cfg.HorizonDays, MaxSlotHorizonDays)
	}

	return SlotSchedule{
		Location:        loc,
		Weekdays:        weekdays,
		DayStartMinutes: start,
		DayEndMinutes:   end,
		DurationMinutes: cfg.DurationMinutes,
		HorizonDays:     cfg.HorizonDays,
	}, nil
}

func parseWeekdays(list string) ([]time.Weekday, error) {
	var weekdays []time.Weekday
	seen := make(map[time.Weekday]struct{})

	for _, name := range strings.Split(list, ",") {
		day, ok := weekdayByName[strings.ToLower(strings.TrimSpace(name))]
		if !ok {
			return nil, fmt.Errorf("slot schedule: unknown working day %q", name)
		}

		if _, dup := seen[day]; dup {
			return nil, fmt.Errorf("slot schedule: working day %q repeats", name)
		}
		seen[day] = struct{}{}
		weekdays = append(weekdays, day)
	}

	return weekdays, nil
}

func parseClockMinutes(value string) (int, error) {
	clock, err := time.Parse(SlotTimeLayout, value)
	if err != nil {
		return 0, fmt.Errorf("%q is not HH:MM", value)
	}

	return clock.Hour()*60 + clock.Minute(), nil
}

// SlotsBetween returns parallel date and start-time slices for every working day in [from, to], inclusive.
// Only the calendar date of from and to is used.
func (sc SlotSchedule) SlotsBetween(from, to time.Time) (dates []time.Time, starts []string) {
	if sc.DurationMinutes <= 0 {
		return nil, nil
	}

	last := time.Date(to.Year(), to.Month(), to.Day(), 0, 0, 0, 0, time.UTC)

	for date := time.Date(from.Year(), from.Month(), from.Day(), 0, 0, 0, 0, time.UTC); !date.After(last); date = date.AddDate(0, 0, 1) {
		if !sc.worksOn(date.Weekday()) {
			continue
		}

		for start := sc.DayStartMinutes; start+sc.DurationMinutes <= sc.DayEndMinutes; start += sc.DurationMinutes {
			dates = append(dates, date)
			starts = append(starts, time.Date(0, 1, 1, start/60, start%60, 0, 0, time.UTC).Format(SlotTimeLayout))
		}
	}

	return dates, starts
}

func (sc SlotSchedule) worksOn(day time.Weekday) bool {
	for _, wd := range sc.Weekdays {
		if wd == day {
			return true
		}
	}

	return false
}
