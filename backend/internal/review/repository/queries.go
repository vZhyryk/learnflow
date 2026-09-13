package reviewrepository

const (
	courseReviewColumns  = `id, course_id, user_id, rating, comment, created_at, updated_at, deleted_at`
	contentReviewColumns = `id, content_item_id, user_id, rating, comment, created_at, updated_at, deleted_at`

	createCourseReviewSQL = `
		INSERT INTO course_reviews (course_id, user_id, rating, comment)
		VALUES ($1, $2, $3, $4)
		RETURNING ` + courseReviewColumns

	createContentReviewSQL = `
		INSERT INTO content_reviews (content_item_id, user_id, rating, comment)
		VALUES ($1, $2, $3, $4)
		RETURNING ` + contentReviewColumns

	updateCourseReviewSQL = `
		UPDATE course_reviews
		SET rating = $2,
		comment = $3,
		updated_at = now()
		WHERE id = $1 AND deleted_at IS NULL
	`

	updateContentReviewSQL = `
		UPDATE content_reviews
		SET rating = $2,
		comment = $3,
		updated_at = now()
		WHERE id = $1 AND deleted_at IS NULL
	`

	deleteCourseReviewSQL = `
		UPDATE course_reviews
		SET deleted_at = now()
		WHERE id = $1 AND deleted_at IS NULL
	`

	deleteContentReviewSQL = `
		UPDATE content_reviews
		SET deleted_at = now()
		WHERE id = $1 AND deleted_at IS NULL
	`

	getCourseReviewByCourseIDSQL = `
		SELECT ` + courseReviewColumns + `
		FROM course_reviews WHERE course_id = $1 AND deleted_at IS NULL
		ORDER BY created_at DESC
		LIMIT $2 OFFSET $3
	`

	getContentReviewByContentIDSQL = `
		SELECT ` + contentReviewColumns + `
		FROM content_reviews WHERE content_item_id = $1 AND deleted_at IS NULL
		ORDER BY created_at DESC
		LIMIT $2 OFFSET $3
	`

	getCourseReviewByIDSQL = `
		SELECT ` + courseReviewColumns + `
		FROM course_reviews WHERE id = $1 AND deleted_at IS NULL
		FOR UPDATE

	`

	getContentReviewByIDSQL = `
		SELECT ` + contentReviewColumns + `
		FROM content_reviews WHERE id = $1 AND deleted_at IS NULL
		FOR UPDATE
	`

	getCourseReviewByUserAndCourseIDSQL = `
		SELECT ` + courseReviewColumns + `
		FROM course_reviews WHERE course_id = $1 AND user_id = $2 AND deleted_at IS NULL
		FOR UPDATE
	`

	getContentReviewByUserAndContentIDSQL = `
		SELECT ` + contentReviewColumns + `
		FROM content_reviews WHERE content_item_id = $1 AND user_id = $2 AND deleted_at IS NULL
		FOR UPDATE
	`

	getCourseReviewStats = `
		SELECT COALESCE(AVG(rating)::float8, 0), COUNT(*)
  		FROM course_reviews
  		WHERE course_id = $1 AND deleted_at IS NULL
  	`

	getContentReviewStats = `
		SELECT COALESCE(AVG(rating)::float8, 0), COUNT(*)
  		FROM content_reviews
  		WHERE content_item_id = $1 AND deleted_at IS NULL
  	`
)

// courseReviewsUserCourseUniqueConstraint / contentReviewsUserContentUniqueConstraint are the
// DB-level backstop for the check-then-insert "one active review per user" race.
const (
	courseReviewsUserCourseUniqueConstraint   = "idx_course_reviews_user_id_course_id_active_unique"
	contentReviewsUserContentUniqueConstraint = "idx_content_reviews_user_id_content_item_id_active_unique"
)
