package consultationsrepository

const (
	publicSlotsColumns = `id, consultant_id, date, to_char(start_time, 'HH24:MI'), duration_minutes, status`
	slotsColumns       = `id, consultant_id, date, to_char(start_time, 'HH24:MI'), duration_minutes, status, updated_by, created_at, updated_at`

	// $3 is an optional consultant filter: NULL lists every consultant.
	getAvailableSlotsSQL = `
		SELECT ` + publicSlotsColumns + `
		FROM consultation_slots
		WHERE status = 'available' AND date BETWEEN $1 AND $2 AND ($3::uuid IS NULL OR consultant_id = $3)
		ORDER BY date, start_time, consultant_id
	`

	countAvailableSlotsSQL = `
		SELECT COUNT(*)
		FROM consultation_slots
		WHERE status = 'available' AND date BETWEEN $1 AND $2 AND ($3::uuid IS NULL OR consultant_id = $3)
	`

	// The join condition matches idx_consultation_briefs_slot_id_active_unique, so a slot has at most
	// one live brief and is never listed twice.
	getAllSlotsSQL = `
		SELECT s.id, s.consultant_id, s.date, to_char(s.start_time, 'HH24:MI'), s.duration_minutes, s.status,
			s.updated_by, s.created_at, s.updated_at, b.id, b.user_id
		FROM consultation_slots s
		LEFT JOIN consultation_briefs b
			ON b.slot_id = s.id AND b.deleted_at IS NULL AND b.status <> 'failed'
		WHERE s.date BETWEEN $1 AND $2 AND ($3::uuid IS NULL OR s.consultant_id = $3)
		ORDER BY s.date, s.start_time, s.consultant_id
	`

	countAllSlotsSQL = `
		SELECT COUNT(*)
		FROM consultation_slots
		WHERE date BETWEEN $1 AND $2 AND ($3::uuid IS NULL OR consultant_id = $3)
	`

	// ORDER BY id keeps the lock order stable between concurrent bulk updates.
	getSlotsByIDsForUpdateSQL = `
		SELECT ` + slotsColumns + `
		FROM consultation_slots
		WHERE id = ANY($1::uuid[])
		ORDER BY id
		FOR UPDATE
	`

	// A booked slot is never changed here, even if the caller forgot to filter it out.
	updateSlotsStatusSQL = `
		UPDATE consultation_slots
		SET status = $1, updated_by = $2, updated_at = now()
		WHERE id = ANY($3::uuid[]) AND status <> 'booked'
	`

	// Slots are unique per consultant, date and start time, so a repeated run only adds what is missing.
	createSlotsIfMissingSQL = `
		INSERT INTO consultation_slots (consultant_id, date, start_time, duration_minutes)
		SELECT $1, s.d, s.t, $4
		FROM unnest($2::date[], $3::time[]) AS s(d, t)
		ON CONFLICT (consultant_id, date, start_time) DO NOTHING
	`

	// The profile is optional, so its fields fall back to empty strings.
	getConsultantsSQL = `
		SELECT u.id, COALESCE(up.first_name, ''), COALESCE(up.last_name, ''), COALESCE(up.avatar_url, ''), COALESCE(up.bio, '')
		FROM users u
		LEFT JOIN user_profiles up ON up.user_id = u.id
		WHERE u.role = 'admin' AND u.status = 'active' AND u.deleted_at IS NULL
		ORDER BY u.id
	`
)
