package reviewrepository

import (
	"fmt"
	reviewdomain "learnflow_backend/internal/review/domain"
	"strings"
)

const filterPlace = "{Filter}"

const (
	courseReviewColumns  = `id, course_id, user_id, rating, comment, created_at, updated_at, deleted_at, deleted_by_user_id`
	contentReviewColumns = `id, content_item_id, user_id, rating, comment, created_at, updated_at, deleted_at, deleted_by_user_id`
	articleReviewColumns = `id, article_id, user_id, rating, comment, created_at, updated_at, deleted_at, deleted_by_user_id`

	createCourseReviewSQL = `
		INSERT INTO course_reviews (course_id, user_id, rating, comment)
		VALUES ($1, $2, $3, $4)
		RETURNING ` + courseReviewColumns

	createContentReviewSQL = `
		INSERT INTO content_reviews (content_item_id, user_id, rating, comment)
		VALUES ($1, $2, $3, $4)
		RETURNING ` + contentReviewColumns

	createArticleReviewSQL = `
		INSERT INTO article_reviews (article_id, user_id, rating, comment)
		VALUES ($1, $2, $3, $4)
		RETURNING ` + articleReviewColumns

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

	updateArticleReviewSQL = `
		UPDATE article_reviews
		SET rating = $2,
		comment = $3,
		updated_at = now()
		WHERE id = $1 AND deleted_at IS NULL
	`

	deleteCourseReviewSQL = `
		UPDATE course_reviews
		SET deleted_at = now(),
		deleted_by_user_id = $2
		WHERE id = $1 AND deleted_at IS NULL
	`

	deleteContentReviewSQL = `
		UPDATE content_reviews
		SET deleted_at = now(),
		deleted_by_user_id = $2
		WHERE id = $1 AND deleted_at IS NULL
	`

	deleteArticleReviewSQL = `
		UPDATE article_reviews
		SET deleted_at = now(),
		deleted_by_user_id = $2
		WHERE id = $1 AND deleted_at IS NULL
	`

	getCourseReviewByCourseIDSQL = `
		SELECT ` + courseReviewColumns + `
		FROM course_reviews WHERE course_id = $1 AND deleted_at IS NULL
		` + filterPlace + `
		ORDER BY created_at DESC
		LIMIT $2 OFFSET $3
	`

	getContentReviewByContentIDSQL = `
		SELECT ` + contentReviewColumns + `
		FROM content_reviews WHERE content_item_id = $1 AND deleted_at IS NULL
		` + filterPlace + `
		ORDER BY created_at DESC
		LIMIT $2 OFFSET $3
	`

	getArticleReviewByArticleIDSQL = `
		SELECT ` + articleReviewColumns + `
		FROM article_reviews WHERE article_id = $1 AND deleted_at IS NULL
		` + filterPlace + `
		ORDER BY created_at DESC
		LIMIT $2 OFFSET $3
	`

	countCourseReviewsByCourseIDSQL = `
		SELECT COUNT(*)
		FROM course_reviews WHERE course_id = $1 AND deleted_at IS NULL
		` + filterPlace + `
	`

	countContentReviewsByContentIDSQL = `
		SELECT COUNT(*)
		FROM content_reviews WHERE content_item_id = $1 AND deleted_at IS NULL
		` + filterPlace + `
	`

	countArticleReviewsByArticleIDSQL = `
		SELECT COUNT(*)
		FROM article_reviews WHERE article_id = $1 AND deleted_at IS NULL
		` + filterPlace + `
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

	getArticleReviewByIDSQL = `
		SELECT ` + articleReviewColumns + `
		FROM article_reviews WHERE id = $1 AND deleted_at IS NULL
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

	getArticleReviewByUserAndArticleIDSQL = `
		SELECT ` + articleReviewColumns + `
		FROM article_reviews WHERE article_id = $1 AND user_id = $2 AND deleted_at IS NULL
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

	getArticleReviewStats = `
		SELECT COALESCE(AVG(rating)::float8, 0), COUNT(*)
  		FROM article_reviews
  		WHERE article_id = $1 AND deleted_at IS NULL
  	`
)

// courseReviewsUserCourseUniqueConstraint / contentReviewsUserContentUniqueConstraint are the
// DB-level backstop for the check-then-insert "one active review per user" race.
const (
	courseReviewsUserCourseUniqueConstraint   = "idx_course_reviews_user_id_course_id_active_unique"
	contentReviewsUserContentUniqueConstraint = "idx_content_reviews_user_id_content_item_id_active_unique"
	articleReviewsUserArticleUniqueConstraint = "idx_article_reviews_user_id_article_id_active_unique"
)

// ratingOperator maps a rating filter op to its SQL comparison operator, or "" when op is unknown.
func ratingOperator(op string) string {
	switch op {
	case "eq":
		return "="
	case "gt":
		return ">"
	case "lt":
		return "<"
	case "gte":
		return ">="
	case "lte":
		return "<="
	default:
		return ""
	}
}

// GenerateFilterQuery returns the rating condition plus ordering and pagination for a rating filter op.
func (rep *Repository) GenerateFilterQuery(op string) string {
	val := ratingOperator(op)
	if val == "" {
		return ""
	}

	return fmt.Sprintf(`
		AND rating %s $2
		ORDER BY created_at DESC
		LIMIT $3 OFFSET $4
	`, val)
}

// applyFilter fills the {Filter} placeholder of a list template (rating condition, ordering, pagination)
// and of its count template (rating condition only), and returns the shared args for both queries.
func (rep *Repository) applyFilter(listSQL, countSQL, id string, filter reviewdomain.ReviewFilter) (query, countQuery string, args []any) {
	args = []any{id}
	idx := strings.Index(listSQL, filterPlace)
	if val := ratingOperator(filter.Op); filter.IsUsed() && idx >= 0 && val != "" {
		args = append(args, filter.Rating)
		query = listSQL[:idx] + rep.GenerateFilterQuery(filter.Op)
		countQuery = strings.Replace(countSQL, filterPlace, fmt.Sprintf("AND rating %s $2", val), 1)
		return query, countQuery, args
	}

	return strings.Replace(listSQL, filterPlace, "", 1), strings.Replace(countSQL, filterPlace, "", 1), args
}
