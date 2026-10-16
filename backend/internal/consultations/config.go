package consultations

import (
	"fmt"

	consultationsdomain "learnflow_backend/internal/consultations/domain"
	"learnflow_backend/internal/infrastructure/env"

	_ "time/tzdata" // the slot time zone must resolve in minimal container images
)

// LoadSlotSchedule reads the SLOT_* environment variables, falling back to the defaults, and validates them.
func LoadSlotSchedule() (consultationsdomain.SlotSchedule, error) {
	schedule, err := consultationsdomain.ParseSlotSchedule(consultationsdomain.SlotScheduleConfig{
		TimeZone:        env.GetStringEnv("SLOT_TIME_ZONE", consultationsdomain.SlotTimeZone),
		WorkingDays:     env.GetStringEnv("SLOT_WORKING_DAYS", "Mon,Tue,Wed,Thu,Fri,Sat"),
		DayStart:        env.GetStringEnv("SLOT_DAY_START", "09:00"),
		DayEnd:          env.GetStringEnv("SLOT_DAY_END", "19:00"),
		DurationMinutes: env.GetIntEnv("SLOT_DURATION_MINUTES", consultationsdomain.DurationMinutes),
		HorizonDays:     env.GetIntEnv("SLOT_HORIZON_DAYS", consultationsdomain.SlotHorizonDays),
	})
	if err != nil {
		return consultationsdomain.SlotSchedule{}, fmt.Errorf("consultations.LoadSlotSchedule: %w", err)
	}

	return schedule, nil
}
