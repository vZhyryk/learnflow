package articlerepository

const (
	articleColumns = `
    id, slug, title, description, body, seo_title, seo_description, og_image_url, is_indexable,
	status, created_by_user_id, created_at, updated_at, updated_by_user_id, published_at, published_by_user_id,
	deleted_at, deleted_by_user_id, archived_at, archived_by_user_id
	`

	createDraftArticleSQL = `
		INSERT INTO articles (slug, title, description, body, seo_title, seo_description, og_image_url, is_indexable, created_by_user_id)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
		RETURNING` + articleColumns

	publishArticleSQL = `
		UPDATE articles
		SET status = 'published',
		published_at = now(),
		published_by_user_id = $2
		WHERE id = $1 AND deleted_at IS NULL
	`

	archiveArticleSQL = `
		UPDATE articles
		SET status = 'archived',
		archived_at = now(),
		archived_by_user_id = $2
		WHERE id = $1 AND deleted_at IS NULL
	`

	deleteArticleSQL = `
		UPDATE articles
		SET deleted_at = now(),
		deleted_by_user_id = $2
		WHERE id = $1 AND deleted_at IS NULL
	`

	updateArticleSQL = `
		UPDATE articles
		SET
			slug = $2,
			title = $3,
			description = $4,
			body = $5,
			seo_title = $6,
			seo_description = $7,
			og_image_url = $8,
			is_indexable = $9,
			updated_by_user_id = $10,
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

	checkIfArticleExistsByID = `
		SELECT EXISTS(SELECT 1 FROM articles WHERE id = $1 AND deleted_at IS NULL)
	`
)
