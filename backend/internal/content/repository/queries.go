package contentrepository

const (
	contentColumns = `
    id, slug, title, content_type, description, body, video_url, file_url, estimated_minutes,
	estimated_pages, thumbnail_url, seo_title, seo_description, og_image_url, canonical_url,
    is_indexable, status, created_by_user_id, created_at, updated_at, published_at, deleted_at
	`

	createDraftContentItemSQL = `
		INSERT INTO content_items (slug, title, content_type, description, body, video_url, file_url, thumbnail_url, estimated_minutes, estimated_pages, seo_title, seo_description, og_image_url, canonical_url, is_indexable, created_by_user_id)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16)
		RETURNING` + contentColumns

	publishContentItemSQL = `
		UPDATE content_items
		SET status = 'published',
		published_at = now()
		WHERE id = $1 AND deleted_at IS NULL
	`

	archiveContentItemSQL = `
		UPDATE content_items
		SET status = 'archived'
		WHERE id = $1 AND deleted_at IS NULL
	`

	deleteContentItemSQL = `
		UPDATE content_items
		SET deleted_at = now()
		WHERE id = $1 AND deleted_at IS NULL
	`

	updateContentItemSQL = `
		UPDATE content_items
		SET
			slug = $2,
			title = $3,
			description = $4,
			body = $5,
    		video_url = $6,
    		file_url = $7,
			estimated_minutes = $8,
			estimated_pages = $9,
			thumbnail_url = $10,
			seo_title = $11,
			seo_description = $12,
			og_image_url = $13,
			canonical_url = $14,
			is_indexable = $15,
			updated_at = now()
		WHERE id = $1 AND deleted_at IS NULL
	`

	getAllPublishedContentItemsSQL = `
		SELECT` + contentColumns + `
		FROM content_items WHERE status = 'published' AND deleted_at IS NULL
		ORDER BY created_at DESC
		LIMIT $1 OFFSET $2
	`

	getAllDraftContentItemsSQL = `
		SELECT` + contentColumns + `
		FROM content_items WHERE status = 'draft' AND deleted_at IS NULL
		ORDER BY created_at DESC
		LIMIT $1 OFFSET $2
	`

	getAllArchivedContentItemsSQL = `
		SELECT` + contentColumns + `
		FROM content_items WHERE status = 'archived'
		ORDER BY created_at DESC
		LIMIT $1 OFFSET $2
	`

	getAllContentItemsSQL = `
		SELECT` + contentColumns + `
		FROM content_items
		ORDER BY created_at DESC
		LIMIT $1 OFFSET $2
	`

	getContentItemByIDSQL = `
		SELECT` + contentColumns + `
		FROM content_items WHERE id = $1 AND deleted_at IS NULL
	`

	getContentItemBySlugSQL = `
		SELECT` + contentColumns + `
		FROM content_items WHERE slug = $1 AND deleted_at IS NULL
	`
)
