package adminrepository

const (
	announcementColumns = `
    id, title, body, created_at, created_by_user_id, updated_at, updated_by_user_id,
	approved_at, approved_by_user_id, expires_at, entity_id, entity_type, channels
	`

	createAnnouncementSQL = `
		INSERT INTO announcements (title, body, created_by_user_id, expires_at, entity_id, entity_type, channels)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		RETURNING` + announcementColumns

	updateAnnouncementSQL = `
		UPDATE announcements
		SET title = $2,
		body = $3,
		entity_id = $4,
		entity_type = $5,
		channels = $6,
		updated_at = now(),
		updated_by_user_id = $7
		WHERE id = $1
	`

	approveAnnouncementSQL = `
		UPDATE announcements
		SET approved_at = now(),
		approved_by_user_id = $2
		WHERE id = $1 AND expires_at > now()
	`

	getAnnouncementsSQL = `
		SELECT ` + announcementColumns + `
		FROM announcements
		ORDER BY created_at DESC
		LIMIT $1 OFFSET $2
	`
	getUnApprovedAnnouncementsSQL = `
		SELECT ` + announcementColumns + `
		FROM announcements
		WHERE approved_at IS NULL AND expires_at > now()
		ORDER BY created_at DESC
		LIMIT $1 OFFSET $2
	`
	getApprovedAnnouncementsSQL = `
		SELECT ` + announcementColumns + `
		FROM announcements
		WHERE approved_at IS NOT NULL AND expires_at > now()
		ORDER BY created_at DESC
		LIMIT $1 OFFSET $2
	`

	getExpiredAnnouncementsSQL = `
		SELECT ` + announcementColumns + `
		FROM announcements
		WHERE approved_at IS NOT NULL AND expires_at < now()
		ORDER BY created_at DESC
		LIMIT $1 OFFSET $2
	`

	getAnnouncementByIDSQL = `
		SELECT ` + announcementColumns + `
		FROM announcements
		WHERE id = $1
	`
)
