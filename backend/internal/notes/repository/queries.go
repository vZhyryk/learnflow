package notesrepository

const (
	notesColumns = `
		id, user_id, resource_type, resource_id, title, description, body, deleted_at, created_at, updated_at
	`
	getUserNotesByIDSQL = `
		SELECT ` + notesColumns + `
		FROM user_notes
		WHERE id = $1 AND user_id = $2 AND deleted_at IS NULL
	`

	getUserNotesByIDForUpdateSQL = getUserNotesByIDSQL + `FOR UPDATE`

	getUserAllNotesByUserIDSQL = `
		SELECT ` + notesColumns + `
		FROM user_notes
		WHERE user_id = $1 AND deleted_at IS NULL
			AND ($2::text = '' OR title ILIKE '%' || $2 || '%' OR body ILIKE '%' || $2 || '%')
		ORDER BY created_at DESC
		LIMIT $3 OFFSET $4
	`

	createUserNotesSQL = `
		INSERT INTO user_notes (user_id, resource_type, resource_id, title, description, body)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING ` + notesColumns

	updateUserNotesSQL = `
		UPDATE user_notes
		SET title = $1,
			description = $2,
			body = $3,
			updated_at = NOW()
		WHERE id = $4 AND user_id = $5 AND deleted_at IS NULL
	`

	deleteUserNotesSQL = `
		UPDATE user_notes
		SET deleted_at = NOW()
		WHERE id = $1 AND user_id = $2 AND deleted_at IS NULL
	`
)
