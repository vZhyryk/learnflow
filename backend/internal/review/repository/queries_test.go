package reviewrepository

import (
	"strings"
	"testing"

	reviewdomain "learnflow_backend/internal/review/domain"

	. "github.com/smartystreets/goconvey/convey"
)

func TestApplyFilter(t *testing.T) {
	Convey("Given a review repository", t, func() {
		repo := newTestRepo(nil)

		Convey("When no filter is set, the placeholder is dropped and only the id is bound", func() {
			query, countQuery, args := repo.applyFilter(getCourseReviewByCourseIDSQL, countCourseReviewsByCourseIDSQL, "course-1", reviewdomain.ReviewFilter{})

			So(args, ShouldResemble, []any{"course-1"})
			So(query, ShouldNotContainSubstring, filterPlace)
			So(query, ShouldContainSubstring, "LIMIT $2 OFFSET $3")
			So(countQuery, ShouldNotContainSubstring, filterPlace)
			So(countQuery, ShouldNotContainSubstring, "LIMIT")
		})

		Convey("When a valid filter is set, list and count share the rating condition and args", func() {
			filter := reviewdomain.ReviewFilter{Op: "gte", Rating: 4}

			query, countQuery, args := repo.applyFilter(getCourseReviewByCourseIDSQL, countCourseReviewsByCourseIDSQL, "course-1", filter)

			So(args, ShouldResemble, []any{"course-1", 4})
			So(query, ShouldContainSubstring, "AND rating >= $2")
			So(query, ShouldContainSubstring, "LIMIT $3 OFFSET $4")
			So(countQuery, ShouldContainSubstring, "AND rating >= $2")
			So(strings.Contains(countQuery, "LIMIT"), ShouldBeFalse)
			So(strings.Contains(countQuery, "ORDER BY"), ShouldBeFalse)
		})

		Convey("When the filter is invalid, it is ignored", func() {
			filter := reviewdomain.ReviewFilter{Op: "bogus", Rating: 4}

			query, countQuery, args := repo.applyFilter(getCourseReviewByCourseIDSQL, countCourseReviewsByCourseIDSQL, "course-1", filter)

			So(args, ShouldResemble, []any{"course-1"})
			So(query, ShouldNotContainSubstring, "AND rating")
			So(countQuery, ShouldNotContainSubstring, "AND rating")
		})
	})
}
