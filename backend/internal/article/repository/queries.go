package articlerepository

const (
	articleColumns = `
    id, slug, title, excerpt, body, seo_title, seo_description, og_image_url, is_indexable,
	status, created_by_user_id, created_at, updated_at, published_at, deleted_at
	`

	createDraftArticleSQL = `
		INSERT INTO articles (slug, title, excerpt, body, seo_title, seo_description, og_image_url, is_indexable, created_by_user_id)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
		RETURNING` + articleColumns

	publishArticleSQL = `
		UPDATE articles
		SET status = 'published',
		published_at = now()
		WHERE id = $1 AND deleted_at IS NULL
	`

	archiveArticleSQL = `
		UPDATE articles
		SET status = 'archived'
		WHERE id = $1 AND deleted_at IS NULL
	`

	deleteArticleSQL = `
		UPDATE articles
		SET deleted_at = now()
		WHERE id = $1 AND deleted_at IS NULL
	`

	updateArticleSQL = `
		UPDATE articles
		SET
			slug = $2,
			title = $3,
			excerpt = $4,
			body = $5,
			seo_title = $6,
			seo_description = $7,
			og_image_url = $8,
			is_indexable = $9,
			updated_at = now()
		WHERE id = $1 AND deleted_at IS NULL
	`

	getAllPublishedArticlesSQL = `
		SELECT` + articleColumns + `
		FROM articles WHERE status = 'published' AND deleted_at IS NULL
		ORDER BY created_at DESC
		LIMIT $1 OFFSET $2
	`

	getAllDraftArticlesSQL = `
		SELECT` + articleColumns + `
		FROM articles WHERE status = 'draft' AND deleted_at IS NULL
		ORDER BY created_at DESC
		LIMIT $1 OFFSET $2
	`

	getAllArchivedArticlesSQL = `
		SELECT` + articleColumns + `
		FROM articles WHERE status = 'archived'
		ORDER BY created_at DESC
		LIMIT $1 OFFSET $2
	`

	getAllArticlesSQL = `
		SELECT` + articleColumns + `
		FROM articles
		ORDER BY created_at DESC
		LIMIT $1 OFFSET $2
	`

	getArticleByIDSQL = `
		SELECT` + articleColumns + `
		FROM articles WHERE id = $1 AND deleted_at IS NULL
	`

	getArticleBySlugSQL = `
		SELECT` + articleColumns + `
		FROM articles WHERE slug = $1 AND deleted_at IS NULL
	`
)
