package reviewdomain

import (
	"errors"
	"strings"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

const validUUID = "b3f1c9e2-6a4d-4f2e-9c1a-2d5e7f8a9b0c"

func validCreateCourseReviewRequest() CreateCourseReviewRequest {
	comment := "comment"
	return CreateCourseReviewRequest{
		CourseID: validUUID,
		UserID:   validUUID,
		Rating:   5,
		Comment:  &comment,
	}
}

func validCreateContentReviewRequest() CreateContentReviewRequest {
	comment := "comment"
	return CreateContentReviewRequest{
		ContentID: validUUID,
		UserID:    validUUID,
		Rating:    5,
		Comment:   &comment,
	}
}

func validUpdateCourseReviewRequest() UpdateCourseReviewRequest {
	comment := "comment"
	rating := 5
	return UpdateCourseReviewRequest{
		ReviewID: validUUID,
		UserID:   validUUID,
		Rating:   &rating,
		Comment:  &comment,
	}
}

func validUpdateContentReviewRequest() UpdateContentReviewRequest {
	comment := "comment"
	rating := 5
	return UpdateContentReviewRequest{
		ReviewID: validUUID,
		UserID:   validUUID,
		Rating:   &rating,
		Comment:  &comment,
	}
}

func TestCreateCourseReviewRequestValidateCourseID(t *testing.T) {
	Convey("CreateCourseReviewRequest.validateCourseID", t, func() {
		Convey("valid UUID", func() {
			req := CreateCourseReviewRequest{CourseID: validUUID}
			So(req.validateCourseID(), ShouldBeNil)
		})

		Convey("empty CourseID", func() {
			req := CreateCourseReviewRequest{CourseID: ""}
			So(errors.Is(req.validateCourseID(), ErrInvalidCourseID), ShouldBeTrue)
		})

		Convey("malformed (non-UUID) CourseID", func() {
			req := CreateCourseReviewRequest{CourseID: "valid-course-id"}
			So(errors.Is(req.validateCourseID(), ErrInvalidCourseID), ShouldBeTrue)
		})
	})
}

func TestCreateCourseReviewRequestValidateRating(t *testing.T) {
	Convey("CreateCourseReviewRequest.validateRating", t, func() {
		Convey("valid title", func() {
			req := CreateCourseReviewRequest{Rating: 3}
			So(req.validateRating(), ShouldBeNil)
		})

		Convey("rating > 5", func() {
			req := CreateCourseReviewRequest{Rating: 6}
			So(errors.Is(req.validateRating(), ErrInvalidRating), ShouldBeTrue)
		})

		Convey("rating < 1", func() {
			req := CreateCourseReviewRequest{Rating: 0}
			So(errors.Is(req.validateRating(), ErrInvalidRating), ShouldBeTrue)
		})
	})
}

func TestUpdateCourseReviewRequestValidateCourseID(t *testing.T) {
	Convey("UpdateCourseReviewRequest.validateID", t, func() {
		Convey("valid ReviewID", func() {
			req := UpdateCourseReviewRequest{ReviewID: "b3f1c9e2-6a4d-4f2e-9c1a-2d5e7f8a9b0c"}
			So(req.validateReviewID(), ShouldBeNil)
		})

		Convey("empty ReviewID", func() {
			req := UpdateCourseReviewRequest{ReviewID: ""}
			So(errors.Is(req.validateReviewID(), ErrInvalidReviewID), ShouldBeTrue)
		})

		Convey("malformed (non-UUID) ReviewID", func() {
			req := UpdateCourseReviewRequest{ReviewID: "content-item-id"}
			So(errors.Is(req.validateReviewID(), ErrInvalidReviewID), ShouldBeTrue)
		})
	})
}

func TestUpdateCourseReviewRequestValidateRating(t *testing.T) {
	Convey("UpdateCourseReviewRequest.validateRating", t, func() {
		Convey("nil is valid", func() {
			req := UpdateCourseReviewRequest{Rating: nil}
			So(req.validateRating(), ShouldBeNil)
		})

		Convey("Rating is too low", func() {
			rating := 0
			req := UpdateCourseReviewRequest{Rating: &rating}
			So(errors.Is(req.validateRating(), ErrInvalidRating), ShouldBeTrue)
		})

		Convey("Rating is too high", func() {
			rating := 6
			req := UpdateCourseReviewRequest{Rating: &rating}
			So(errors.Is(req.validateRating(), ErrInvalidRating), ShouldBeTrue)
		})

		Convey("valid rating", func() {
			rating := 5
			req := UpdateCourseReviewRequest{Rating: &rating}
			So(req.validateRating(), ShouldBeNil)
		})
	})
}

func TestUpdateCourseReviewRequestValidateComment(t *testing.T) {
	Convey("UpdateCourseReviewRequest.validateComment", t, func() {
		Convey("nil is valid", func() {
			req := UpdateCourseReviewRequest{Comment: nil}
			So(req.validateComment(), ShouldBeNil)
		})

		Convey("whitespace-only is invalid", func() {
			val := "   "
			req := UpdateCourseReviewRequest{Comment: &val}
			So(errors.Is(req.validateComment(), ErrInvalidComment), ShouldBeTrue)
		})

		Convey("valid Comment", func() {
			val := "Comment"
			req := UpdateCourseReviewRequest{Comment: &val}
			So(req.validateComment(), ShouldBeNil)
		})
	})
}

func TestCreateCourseReviewRequestValidateComment(t *testing.T) {
	Convey("CreateCourseReviewRequest.validateComment", t, func() {
		Convey("nil is valid", func() {
			req := CreateCourseReviewRequest{Comment: nil}
			So(req.validateComment(), ShouldBeNil)
		})

		Convey("empty is invalid", func() {
			empty := ""
			req := CreateCourseReviewRequest{Comment: &empty}
			So(errors.Is(req.validateComment(), ErrInvalidComment), ShouldBeTrue)
		})

		Convey("whitespace-only is invalid", func() {
			val := "   "
			req := CreateCourseReviewRequest{Comment: &val}
			So(errors.Is(req.validateComment(), ErrInvalidComment), ShouldBeTrue)
		})

		Convey("2000 runes is valid (boundary)", func() {
			val := strings.Repeat("A", 2000)
			req := CreateCourseReviewRequest{Comment: &val}
			So(req.validateComment(), ShouldBeNil)
		})
	})
}

func TestCreateCourseReviewRequestValidate(t *testing.T) {
	Convey("CreateCourseReviewRequest.Validate", t, func() {
		Convey("all fields valid", func() {
			req := validCreateCourseReviewRequest()
			So(req.Validate(), ShouldBeNil)
		})

		Convey("invalid course id fails", func() {
			req := validCreateCourseReviewRequest()
			req.CourseID = ""
			So(errors.Is(req.Validate(), ErrInvalidCourseID), ShouldBeTrue)
		})

		Convey("invalid rating lower than 0", func() {
			req := validCreateCourseReviewRequest()
			req.Rating = 0
			So(errors.Is(req.Validate(), ErrInvalidRating), ShouldBeTrue)
		})

		Convey("invalid rating higher than 5", func() {
			req := validCreateCourseReviewRequest()
			req.Rating = 6
			So(errors.Is(req.Validate(), ErrInvalidRating), ShouldBeTrue)
		})

		Convey("invalid comment", func() {
			req := validCreateCourseReviewRequest()
			empty := ""
			req.Comment = &empty
			So(errors.Is(req.Validate(), ErrInvalidComment), ShouldBeTrue)
		})
	})
}

func TestCreateContentReviewRequestValidateContentID(t *testing.T) {
	Convey("CreateContentReviewRequest.validateContentID", t, func() {
		Convey("valid UUID", func() {
			req := CreateContentReviewRequest{ContentID: validUUID}
			So(req.validateContentID(), ShouldBeNil)
		})

		Convey("empty ContentID", func() {
			req := CreateContentReviewRequest{ContentID: ""}
			So(errors.Is(req.validateContentID(), ErrInvalidContentItemID), ShouldBeTrue)
		})

		Convey("malformed (non-UUID) ContentID", func() {
			req := CreateContentReviewRequest{ContentID: "valid-course-id"}
			So(errors.Is(req.validateContentID(), ErrInvalidContentItemID), ShouldBeTrue)
		})
	})
}

func TestCreateContentReviewRequestValidateRating(t *testing.T) {
	Convey("CreateContentReviewRequest.validateRating", t, func() {
		Convey("valid title", func() {
			req := CreateContentReviewRequest{Rating: 3}
			So(req.validateRating(), ShouldBeNil)
		})

		Convey("rating > 5", func() {
			req := CreateContentReviewRequest{Rating: 6}
			So(errors.Is(req.validateRating(), ErrInvalidRating), ShouldBeTrue)
		})

		Convey("rating < 1", func() {
			req := CreateContentReviewRequest{Rating: 0}
			So(errors.Is(req.validateRating(), ErrInvalidRating), ShouldBeTrue)
		})
	})
}

func TestCreateContentReviewRequestValidateComment(t *testing.T) {
	Convey("CreateContentReviewRequest.validateComment", t, func() {
		Convey("nil is valid", func() {
			req := CreateContentReviewRequest{Comment: nil}
			So(req.validateComment(), ShouldBeNil)
		})

		Convey("empty is invalid", func() {
			empty := ""
			req := CreateContentReviewRequest{Comment: &empty}
			So(errors.Is(req.validateComment(), ErrInvalidComment), ShouldBeTrue)
		})

		Convey("whitespace-only is invalid", func() {
			val := "   "
			req := CreateContentReviewRequest{Comment: &val}
			So(errors.Is(req.validateComment(), ErrInvalidComment), ShouldBeTrue)
		})

		Convey("2000 runes is valid (boundary)", func() {
			val := strings.Repeat("A", 2000)
			req := CreateContentReviewRequest{Comment: &val}
			So(req.validateComment(), ShouldBeNil)
		})
	})
}

func TestCreateContentReviewRequestValidate(t *testing.T) {
	Convey("CreateContentReviewRequest.Validate", t, func() {
		Convey("all fields valid", func() {
			req := validCreateContentReviewRequest()
			So(req.Validate(), ShouldBeNil)
		})

		Convey("invalid content id fails", func() {
			req := validCreateContentReviewRequest()
			req.ContentID = ""
			So(errors.Is(req.Validate(), ErrInvalidContentItemID), ShouldBeTrue)
		})

		Convey("invalid rating lower than 0", func() {
			req := validCreateContentReviewRequest()
			req.Rating = 0
			So(errors.Is(req.Validate(), ErrInvalidRating), ShouldBeTrue)
		})

		Convey("invalid rating higher than 5", func() {
			req := validCreateContentReviewRequest()
			req.Rating = 6
			So(errors.Is(req.Validate(), ErrInvalidRating), ShouldBeTrue)
		})

		Convey("invalid comment", func() {
			req := validCreateContentReviewRequest()
			empty := ""
			req.Comment = &empty
			So(errors.Is(req.Validate(), ErrInvalidComment), ShouldBeTrue)
		})
	})
}

func TestUpdateContentReviewRequestValidateReviewID(t *testing.T) {
	Convey("UpdateContentReviewRequest.validateReviewID", t, func() {
		Convey("valid ReviewID", func() {
			req := UpdateContentReviewRequest{ReviewID: "b3f1c9e2-6a4d-4f2e-9c1a-2d5e7f8a9b0c"}
			So(req.validateReviewID(), ShouldBeNil)
		})

		Convey("empty ReviewID", func() {
			req := UpdateContentReviewRequest{ReviewID: ""}
			So(errors.Is(req.validateReviewID(), ErrInvalidReviewID), ShouldBeTrue)
		})

		Convey("malformed (non-UUID) ReviewID", func() {
			req := UpdateContentReviewRequest{ReviewID: "content-item-id"}
			So(errors.Is(req.validateReviewID(), ErrInvalidReviewID), ShouldBeTrue)
		})
	})
}

func TestUpdateContentReviewRequestValidateRating(t *testing.T) {
	Convey("UpdateContentReviewRequest.validateRating", t, func() {
		Convey("nil is valid", func() {
			req := UpdateContentReviewRequest{Rating: nil}
			So(req.validateRating(), ShouldBeNil)
		})

		Convey("Rating is too low", func() {
			rating := 0
			req := UpdateContentReviewRequest{Rating: &rating}
			So(errors.Is(req.validateRating(), ErrInvalidRating), ShouldBeTrue)
		})

		Convey("Rating is too high", func() {
			rating := 6
			req := UpdateContentReviewRequest{Rating: &rating}
			So(errors.Is(req.validateRating(), ErrInvalidRating), ShouldBeTrue)
		})

		Convey("valid rating", func() {
			rating := 5
			req := UpdateContentReviewRequest{Rating: &rating}
			So(req.validateRating(), ShouldBeNil)
		})
	})
}

func TestUpdateContentReviewRequestValidateComment(t *testing.T) {
	Convey("UpdateContentReviewRequest.validateComment", t, func() {
		Convey("nil is valid", func() {
			req := UpdateContentReviewRequest{Comment: nil}
			So(req.validateComment(), ShouldBeNil)
		})

		Convey("whitespace-only is invalid", func() {
			val := "   "
			req := UpdateContentReviewRequest{Comment: &val}
			So(errors.Is(req.validateComment(), ErrInvalidComment), ShouldBeTrue)
		})

		Convey("valid Comment", func() {
			val := "Comment"
			req := UpdateContentReviewRequest{Comment: &val}
			So(req.validateComment(), ShouldBeNil)
		})
	})
}

func TestUpdateCourseReviewRequestValidate(t *testing.T) {
	Convey("UpdateCourseReviewRequest.Validate", t, func() {
		Convey("all fields valid", func() {
			req := validUpdateCourseReviewRequest()
			So(req.Validate(), ShouldBeNil)
		})

		Convey("invalid review id fails", func() {
			req := validUpdateCourseReviewRequest()
			req.ReviewID = ""
			So(errors.Is(req.Validate(), ErrInvalidReviewID), ShouldBeTrue)
		})

		Convey("invalid rating lower than 1", func() {
			req := validUpdateCourseReviewRequest()
			rating := 0
			req.Rating = &rating
			So(errors.Is(req.Validate(), ErrInvalidRating), ShouldBeTrue)
		})

		Convey("invalid rating higher than 5", func() {
			req := validUpdateCourseReviewRequest()
			rating := 6
			req.Rating = &rating
			So(errors.Is(req.Validate(), ErrInvalidRating), ShouldBeTrue)
		})

		Convey("invalid comment", func() {
			req := validUpdateCourseReviewRequest()
			empty := "   "
			req.Comment = &empty
			So(errors.Is(req.Validate(), ErrInvalidComment), ShouldBeTrue)
		})
	})
}

func TestUpdateContentReviewRequestValidate(t *testing.T) {
	Convey("UpdateContentReviewRequest.Validate", t, func() {
		Convey("all fields valid", func() {
			req := validUpdateContentReviewRequest()
			So(req.Validate(), ShouldBeNil)
		})

		Convey("invalid review id fails", func() {
			req := validUpdateContentReviewRequest()
			req.ReviewID = ""
			So(errors.Is(req.Validate(), ErrInvalidReviewID), ShouldBeTrue)
		})

		Convey("invalid rating lower than 1", func() {
			req := validUpdateContentReviewRequest()
			rating := 0
			req.Rating = &rating
			So(errors.Is(req.Validate(), ErrInvalidRating), ShouldBeTrue)
		})

		Convey("invalid rating higher than 5", func() {
			req := validUpdateContentReviewRequest()
			rating := 6
			req.Rating = &rating
			So(errors.Is(req.Validate(), ErrInvalidRating), ShouldBeTrue)
		})

		Convey("invalid comment", func() {
			req := validUpdateContentReviewRequest()
			empty := "   "
			req.Comment = &empty
			So(errors.Is(req.Validate(), ErrInvalidComment), ShouldBeTrue)
		})
	})
}
