package courserepository

const (
	courseColumns = `
	    id, slug, title, description, thumbnail_url, preview_video_url, status, estimated_minutes,
		seo_title, seo_description, og_image_url, canonical_url, is_indexable,
		created_by_user_id, created_at, updated_at, published_at, deleted_at
	`

	createDraftCourseSQL = `
		INSERT INTO courses (slug, title, description, thumbnail_url, preview_video_url, estimated_minutes, seo_title, seo_description, og_image_url, canonical_url, is_indexable, created_by_user_id)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)
		RETURNING` + courseColumns

	publishCourseSQL = `
		UPDATE courses
		SET status = 'published',
		published_at = now()
		WHERE id = $1 AND deleted_at IS NULL
	`

	archiveCourseSQL = `
		UPDATE courses
		SET status = 'archived'
		WHERE id = $1 AND deleted_at IS NULL
	`

	deleteCourseSQL = `
		UPDATE courses
		SET deleted_at = now()
		WHERE id = $1 AND deleted_at IS NULL
	`

	updateCourseSQL = `
		UPDATE courses
		SET
			slug = $2,
			title = $3,
			description = $4,
			thumbnail_url = $5,
			preview_video_url = $6,
			estimated_minutes = $7,
			seo_title = $8,
			seo_description = $9,
			og_image_url = $10,
			canonical_url = $11,
			is_indexable = $12,
			updated_at = now()
		WHERE id = $1 AND deleted_at IS NULL
	`

	getAllPublishedCoursesSQL = `
		SELECT` + courseColumns + `
		FROM courses WHERE status = 'published' AND deleted_at IS NULL
		ORDER BY created_at DESC
	`

	getAllDraftCoursesSQL = `
		SELECT` + courseColumns + `
		FROM courses WHERE status = 'draft' AND deleted_at IS NULL
		ORDER BY created_at DESC
	`

	getAllArchivedCoursesSQL = `
		SELECT` + courseColumns + `
		FROM courses WHERE status = 'archived'
		ORDER BY created_at DESC
	`

	getAllCoursesSQL = `
		SELECT` + courseColumns + `
		FROM courses
		ORDER BY created_at DESC
	`

	getCourseByIDSQL = `
		SELECT` + courseColumns + `
		FROM courses WHERE id = $1 AND deleted_at IS NULL
	`

	getCourseBySlugSQL = `
		SELECT` + courseColumns + `
		FROM courses WHERE slug = $1 AND deleted_at IS NULL
	`
)
