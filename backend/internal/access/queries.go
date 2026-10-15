package access

// Access-check queries used by Checker.
const (
	// HasAccessContentQuery checks direct content access and course-derived access.
	HasAccessContentQuery = `
		SELECT EXISTS (
			SELECT 1 FROM user_content_access
			WHERE user_id = $1 AND content_item_id = $2 AND status = 'active'
				AND (expires_at IS NULL OR expires_at > now())
		  UNION
			SELECT 1 FROM user_course_access uca
			JOIN course_content_items cci ON cci.course_id = uca.course_id
			WHERE uca.user_id = $1 AND cci.content_item_id = $2 AND uca.status = 'active'
				AND (uca.expires_at IS NULL OR uca.expires_at > now())
		)
	`
	HasAccessCourseQuery = `
		SELECT EXISTS (
			SELECT 1 FROM user_course_access
			WHERE user_id = $1 AND course_id = $2 AND status = 'active'
				AND (expires_at IS NULL OR expires_at > now())
		)
	`
)
