package adminrepository

const (
	announcementColumns = `
		id, title, body, created_at, created_by_user_id, updated_at, updated_by_user_id,
		approved_at, approved_by_user_id, expires_at, entity_id, entity_type, channels
	`

	createAnnouncementSQL = `
		INSERT INTO announcements (title, body, created_by_user_id, expires_at, entity_id, entity_type, channels)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		RETURNING ` + announcementColumns

	updateAnnouncementSQL = `
		UPDATE announcements
		SET title = $2,
		body = $3,
		entity_id = $4,
		entity_type = $5,
		channels = $6,
		updated_at = now(),
		updated_by_user_id = $7,
		expires_at = $8
		WHERE id = $1
	`

	approveAnnouncementSQL = `
		UPDATE announcements
		SET approved_at = now(),
		approved_by_user_id = $2
		WHERE id = $1 AND (expires_at IS NULL OR expires_at > now()) AND approved_at IS NULL
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
		WHERE approved_at IS NULL AND (expires_at IS NULL OR expires_at > now())
		ORDER BY created_at DESC
		LIMIT $1 OFFSET $2
	`
	getApprovedAnnouncementsSQL = `
		SELECT ` + announcementColumns + `
		FROM announcements
		WHERE approved_at IS NOT NULL AND (expires_at IS NULL OR expires_at > now())
		ORDER BY created_at DESC
		LIMIT $1 OFFSET $2
	`

	// Access predicates mirror access.HasAccessCourseQuery / HasAccessContentQuery.
	getPublicAnnouncementsSQL = `
		SELECT a.id, a.title, a.body, a.approved_at, a.expires_at, a.entity_id, a.entity_type
		FROM announcements a
		WHERE a.approved_at IS NOT NULL AND (a.expires_at IS NULL OR a.expires_at > now()) AND 'banner' = ANY(a.channels)
			AND (
				a.entity_type IS NULL
				OR a.entity_type = 'article'
				OR (a.entity_type = 'course' AND EXISTS (
					SELECT 1 FROM user_course_access uca
					WHERE uca.user_id = $1 AND uca.course_id = a.entity_id AND uca.status = 'active'
						AND (uca.expires_at IS NULL OR uca.expires_at > now())
				))
				OR (a.entity_type = 'content_item' AND (
					EXISTS (
						SELECT 1 FROM user_content_access uca
						WHERE uca.user_id = $1 AND uca.content_item_id = a.entity_id AND uca.status = 'active'
							AND (uca.expires_at IS NULL OR uca.expires_at > now())
					)
					OR EXISTS (
						SELECT 1 FROM user_course_access uca
						JOIN course_content_items cci ON cci.course_id = uca.course_id
						WHERE uca.user_id = $1 AND cci.content_item_id = a.entity_id AND uca.status = 'active'
							AND (uca.expires_at IS NULL OR uca.expires_at > now())
					)
				))
			)
		ORDER BY a.created_at DESC
		LIMIT $2 OFFSET $3
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
