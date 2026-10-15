package coursedomain

import (
	"errors"
	"strings"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestCourseStatus(t *testing.T) {
	Convey("Course validation", t, func() {
		Convey("Is Valid", func() {
			So(DraftStatus.Valid(), ShouldBeTrue)
		})
		Convey("is invalid", func() {
			reason := CourseStatus("invalid")
			So(reason.Valid(), ShouldBeFalse)
		})
	})
}

func TestCourseCheckTitleReady(t *testing.T) {
	Convey("checkTitleReady", t, func() {
		Convey("Valid Title", func() {
			course := Course{Title: "title"}
			So(course.checkTitleReady(), ShouldBeNil)
			So(course.ReadyToPublish(), ShouldNotBeNil)
		})

		Convey("invalid Title", func() {
			course := Course{Title: ""}
			So(errors.Is(course.checkTitleReady(), ErrInvalidTitle), ShouldBeTrue)
			So(course.ReadyToPublish(), ShouldNotBeNil)
		})
	})
}

func TestCourseCheckDescriptionReady(t *testing.T) {
	Convey("checkDescriptionReady", t, func() {
		Convey("Valid Description (nil)", func() {
			course := Course{Title: "title", Description: nil}
			So(errors.Is(course.checkDescriptionReady(), ErrInvalidDescription), ShouldBeTrue)
			So(course.ReadyToPublish(), ShouldNotBeNil)
		})

		Convey("invalid Description (empty)", func() {
			empty := ""
			course := Course{Title: "title", Description: &empty}
			So(errors.Is(course.checkDescriptionReady(), ErrInvalidDescription), ShouldBeTrue)
			So(course.ReadyToPublish(), ShouldNotBeNil)
		})

		Convey("Valid Description", func() {
			empty := "not empty"
			course := Course{Title: "title", Description: &empty}
			So(course.checkDescriptionReady(), ShouldBeNil)
			So(course.ReadyToPublish(), ShouldNotBeNil)
		})
	})
}

func TestCourseCheckSeoTitleReady(t *testing.T) {
	Convey("checkSeoTitleReady", t, func() {
		Convey("Valid SeoTitle (nil)", func() {
			notEmpty := "not empty"
			course := Course{Title: "title", Description: &notEmpty, SeoTitle: nil}
			So(errors.Is(course.checkSeoTitleReady(), ErrInvalidSeoTitle), ShouldBeTrue)
			So(course.ReadyToPublish(), ShouldNotBeNil)
		})

		Convey("invalid SeoTitle (empty)", func() {
			empty := ""
			notEmpty := "not empty"
			course := Course{Title: "title", Description: &notEmpty, SeoTitle: &empty}
			So(errors.Is(course.checkSeoTitleReady(), ErrInvalidSeoTitle), ShouldBeTrue)
			So(course.ReadyToPublish(), ShouldNotBeNil)
		})

		Convey("Valid SeoTitle", func() {
			notEmpty := "not empty"
			course := Course{Title: "title", Description: &notEmpty, SeoTitle: &notEmpty}
			So(course.checkSeoTitleReady(), ShouldBeNil)
			So(course.ReadyToPublish(), ShouldNotBeNil)
		})
	})
}

func TestCourseCheckSeoDescriptionReady(t *testing.T) {
	Convey("checkSeoDescriptionReady", t, func() {
		Convey("Valid SeoDescription (nil)", func() {
			notEmpty := "not empty"
			course := Course{Title: "title", Description: &notEmpty, SeoTitle: &notEmpty, SeoDescription: nil}
			So(errors.Is(course.checkSeoDescriptionReady(), ErrInvalidSeoDescription), ShouldBeTrue)
			So(course.ReadyToPublish(), ShouldNotBeNil)
		})

		Convey("invalid SeoDescription (empty)", func() {
			empty := ""
			notEmpty := "not empty"
			course := Course{Title: "title", Description: &notEmpty, SeoTitle: &notEmpty, SeoDescription: &empty}
			So(errors.Is(course.checkSeoDescriptionReady(), ErrInvalidSeoDescription), ShouldBeTrue)
			So(course.ReadyToPublish(), ShouldNotBeNil)
		})

		Convey("Valid SeoDescription", func() {
			notEmpty := "not empty"
			course := Course{Title: "title", Description: &notEmpty, SeoTitle: &notEmpty, SeoDescription: &notEmpty}
			So(course.checkSeoDescriptionReady(), ShouldBeNil)
			So(course.ReadyToPublish(), ShouldNotBeNil)
		})
	})
}

func TestCourseCheckMediaReady(t *testing.T) {
	Convey("checkMediaReady", t, func() {
		Convey("invalid Media (both nil)", func() {
			notEmpty := "not empty"
			course := Course{Title: "title", Description: &notEmpty, SeoTitle: &notEmpty, SeoDescription: &notEmpty, ThumbnailURL: nil, PreviewVideoURL: nil}
			So(errors.Is(course.checkMediaReady(), ErrInvalidThumbnailURL), ShouldBeTrue)
			So(course.ReadyToPublish(), ShouldNotBeNil)
		})

		Convey("invalid Media (both empty)", func() {
			empty := ""
			notEmpty := "not empty"
			course := Course{Title: "title", Description: &notEmpty, SeoTitle: &notEmpty, SeoDescription: &notEmpty, ThumbnailURL: &empty, PreviewVideoURL: &empty}
			So(errors.Is(course.checkMediaReady(), ErrInvalidThumbnailURL), ShouldBeTrue)
			So(course.ReadyToPublish(), ShouldNotBeNil)
		})

		Convey("Valid Media (only ThumbnailURL set)", func() {
			notEmpty := "not empty"
			course := Course{Title: "title", Description: &notEmpty, SeoTitle: &notEmpty, SeoDescription: &notEmpty, ThumbnailURL: &notEmpty, PreviewVideoURL: nil}
			So(course.checkMediaReady(), ShouldBeNil)
			So(course.ReadyToPublish(), ShouldBeNil)
		})

		Convey("Valid Media (only PreviewVideoURL set)", func() {
			notEmpty := "not empty"
			course := Course{Title: "title", Description: &notEmpty, SeoTitle: &notEmpty, SeoDescription: &notEmpty, ThumbnailURL: nil, PreviewVideoURL: &notEmpty}
			So(course.checkMediaReady(), ShouldBeNil)
			So(course.ReadyToPublish(), ShouldBeNil)
		})

		Convey("Valid Media (both set)", func() {
			notEmpty := "not empty"
			course := Course{Title: "title", Description: &notEmpty, SeoTitle: &notEmpty, SeoDescription: &notEmpty, ThumbnailURL: &notEmpty, PreviewVideoURL: &notEmpty}
			So(course.checkMediaReady(), ShouldBeNil)
			So(course.ReadyToPublish(), ShouldBeNil)
		})
	})
}

func validCreateCourseRequest() CreateCourseRequest {
	return CreateCourseRequest{
		Slug:  "valid-slug",
		Title: "Title",
	}
}

func TestCreateCourseRequestValidateSlug(t *testing.T) {
	Convey("CreateCourseRequest.validateSlug", t, func() {
		Convey("valid slug", func() {
			req := CreateCourseRequest{Slug: "valid-slug"}
			So(req.validateSlug(), ShouldBeNil)
		})

		Convey("empty slug", func() {
			req := CreateCourseRequest{Slug: ""}
			So(errors.Is(req.validateSlug(), ErrInvalidSlug), ShouldBeTrue)
		})

		Convey("malformed slug", func() {
			req := CreateCourseRequest{Slug: "Not A Slug!"}
			So(errors.Is(req.validateSlug(), ErrInvalidSlug), ShouldBeTrue)
		})
	})
}

func TestCreateCourseRequestValidateTitle(t *testing.T) {
	Convey("CreateCourseRequest.validateTitle", t, func() {
		Convey("valid title", func() {
			req := CreateCourseRequest{Title: "Title"}
			So(req.validateTitle(), ShouldBeNil)
		})

		Convey("empty title", func() {
			req := CreateCourseRequest{Title: ""}
			So(errors.Is(req.validateTitle(), ErrInvalidTitle), ShouldBeTrue)
		})

		Convey("whitespace-only title", func() {
			req := CreateCourseRequest{Title: "   "}
			So(errors.Is(req.validateTitle(), ErrInvalidTitle), ShouldBeTrue)
		})

		Convey("300 runes is valid (boundary)", func() {
			req := CreateCourseRequest{Title: strings.Repeat("A", 300)}
			So(req.validateTitle(), ShouldBeNil)
		})

		Convey("301 runes is invalid (boundary)", func() {
			req := CreateCourseRequest{Title: strings.Repeat("A", 301)}
			So(errors.Is(req.validateTitle(), ErrInvalidTitle), ShouldBeTrue)
		})
	})
}

func TestCreateCourseRequestValidateDescription(t *testing.T) {
	Convey("CreateCourseRequest.validateDescription", t, func() {
		Convey("nil is valid", func() {
			req := CreateCourseRequest{Description: nil}
			So(req.validateDescription(), ShouldBeNil)
		})

		Convey("empty is invalid", func() {
			empty := ""
			req := CreateCourseRequest{Description: &empty}
			So(errors.Is(req.validateDescription(), ErrInvalidDescription), ShouldBeTrue)
		})

		Convey("10000 runes is valid (boundary)", func() {
			val := strings.Repeat("A", 10000)
			req := CreateCourseRequest{Description: &val}
			So(req.validateDescription(), ShouldBeNil)
		})

		Convey("10001 runes is invalid (boundary)", func() {
			val := strings.Repeat("A", 10001)
			req := CreateCourseRequest{Description: &val}
			So(errors.Is(req.validateDescription(), ErrInvalidDescription), ShouldBeTrue)
		})
	})
}

func TestCreateCourseRequestValidateSeoTitle(t *testing.T) {
	Convey("CreateCourseRequest.validateSeoTitle", t, func() {
		Convey("nil is valid", func() {
			req := CreateCourseRequest{SeoTitle: nil}
			So(req.validateSeoTitle(), ShouldBeNil)
		})

		Convey("empty is invalid", func() {
			empty := ""
			req := CreateCourseRequest{SeoTitle: &empty}
			So(errors.Is(req.validateSeoTitle(), ErrInvalidSeoTitle), ShouldBeTrue)
		})

		Convey("70 runes is valid (boundary)", func() {
			val := strings.Repeat("A", 70)
			req := CreateCourseRequest{SeoTitle: &val}
			So(req.validateSeoTitle(), ShouldBeNil)
		})

		Convey("71 runes is invalid (boundary)", func() {
			val := strings.Repeat("A", 71)
			req := CreateCourseRequest{SeoTitle: &val}
			So(errors.Is(req.validateSeoTitle(), ErrInvalidSeoTitle), ShouldBeTrue)
		})
	})
}

func TestCreateCourseRequestValidateSeoDescription(t *testing.T) {
	Convey("CreateCourseRequest.validateSeoDescription", t, func() {
		Convey("nil is valid", func() {
			req := CreateCourseRequest{SeoDescription: nil}
			So(req.validateSeoDescription(), ShouldBeNil)
		})

		Convey("empty is invalid", func() {
			empty := ""
			req := CreateCourseRequest{SeoDescription: &empty}
			So(errors.Is(req.validateSeoDescription(), ErrInvalidSeoDescription), ShouldBeTrue)
		})

		Convey("160 runes is valid (boundary)", func() {
			val := strings.Repeat("A", 160)
			req := CreateCourseRequest{SeoDescription: &val}
			So(req.validateSeoDescription(), ShouldBeNil)
		})

		Convey("161 runes is invalid (boundary)", func() {
			val := strings.Repeat("A", 161)
			req := CreateCourseRequest{SeoDescription: &val}
			So(errors.Is(req.validateSeoDescription(), ErrInvalidSeoDescription), ShouldBeTrue)
		})
	})
}

func TestCreateCourseRequestValidateEstimatedMinutes(t *testing.T) {
	Convey("CreateCourseRequest.validateEstimatedMinutes", t, func() {
		Convey("nil is valid", func() {
			req := CreateCourseRequest{EstimatedMinutes: nil}
			So(req.validateEstimatedMinutes(), ShouldBeNil)
		})

		Convey("positive value is valid", func() {
			val := 30
			req := CreateCourseRequest{EstimatedMinutes: &val}
			So(req.validateEstimatedMinutes(), ShouldBeNil)
		})

		Convey("zero is invalid", func() {
			val := 0
			req := CreateCourseRequest{EstimatedMinutes: &val}
			So(errors.Is(req.validateEstimatedMinutes(), ErrInvalidEstimatedMinutes), ShouldBeTrue)
		})

		Convey("negative value is invalid", func() {
			val := -1
			req := CreateCourseRequest{EstimatedMinutes: &val}
			So(errors.Is(req.validateEstimatedMinutes(), ErrInvalidEstimatedMinutes), ShouldBeTrue)
		})
	})
}

func TestCreateCourseRequestValidateThumbnailURL(t *testing.T) {
	Convey("CreateCourseRequest.validateThumbnailURL", t, func() {
		Convey("nil is valid", func() {
			req := CreateCourseRequest{ThumbnailURL: nil}
			So(req.validateThumbnailURL(), ShouldBeNil)
		})

		Convey("empty is invalid", func() {
			empty := ""
			req := CreateCourseRequest{ThumbnailURL: &empty}
			So(errors.Is(req.validateThumbnailURL(), ErrInvalidThumbnailURL), ShouldBeTrue)
		})

		Convey("non-https is invalid", func() {
			val := "http://example.com/image.png"
			req := CreateCourseRequest{ThumbnailURL: &val}
			So(errors.Is(req.validateThumbnailURL(), ErrInvalidThumbnailURL), ShouldBeTrue)
		})

		Convey("valid https URL", func() {
			val := "https://example.com/image.png"
			req := CreateCourseRequest{ThumbnailURL: &val}
			So(req.validateThumbnailURL(), ShouldBeNil)
		})
	})
}

func TestCreateCourseRequestValidatePreviewVideoURL(t *testing.T) {
	Convey("CreateCourseRequest.validatePreviewVideoURL", t, func() {
		Convey("nil is valid", func() {
			req := CreateCourseRequest{PreviewVideoURL: nil}
			So(req.validatePreviewVideoURL(), ShouldBeNil)
		})

		Convey("empty is invalid", func() {
			empty := ""
			req := CreateCourseRequest{PreviewVideoURL: &empty}
			So(errors.Is(req.validatePreviewVideoURL(), ErrInvalidPreviewVideoURL), ShouldBeTrue)
		})

		Convey("non-https is invalid", func() {
			val := "http://example.com/video.mp4"
			req := CreateCourseRequest{PreviewVideoURL: &val}
			So(errors.Is(req.validatePreviewVideoURL(), ErrInvalidPreviewVideoURL), ShouldBeTrue)
		})

		Convey("valid https URL", func() {
			val := "https://example.com/video.mp4"
			req := CreateCourseRequest{PreviewVideoURL: &val}
			So(req.validatePreviewVideoURL(), ShouldBeNil)
		})
	})
}

func TestCreateCourseRequestValidateOgImageURL(t *testing.T) {
	Convey("CreateCourseRequest.validateOgImageURL", t, func() {
		Convey("nil is valid", func() {
			req := CreateCourseRequest{OgImageURL: nil}
			So(req.validateOgImageURL(), ShouldBeNil)
		})

		Convey("empty is invalid", func() {
			empty := ""
			req := CreateCourseRequest{OgImageURL: &empty}
			So(errors.Is(req.validateOgImageURL(), ErrInvalidOgImageURL), ShouldBeTrue)
		})

		Convey("non-https is invalid", func() {
			val := "http://example.com/og.png"
			req := CreateCourseRequest{OgImageURL: &val}
			So(errors.Is(req.validateOgImageURL(), ErrInvalidOgImageURL), ShouldBeTrue)
		})

		Convey("valid https URL", func() {
			val := "https://example.com/og.png"
			req := CreateCourseRequest{OgImageURL: &val}
			So(req.validateOgImageURL(), ShouldBeNil)
		})
	})
}

func TestCreateCourseRequestValidateCanonicalURL(t *testing.T) {
	Convey("CreateCourseRequest.validateCanonicalURL", t, func() {
		Convey("nil is valid", func() {
			req := CreateCourseRequest{CanonicalURL: nil}
			So(req.validateCanonicalURL(), ShouldBeNil)
		})

		Convey("empty is invalid", func() {
			empty := ""
			req := CreateCourseRequest{CanonicalURL: &empty}
			So(errors.Is(req.validateCanonicalURL(), ErrInvalidCanonicalURL), ShouldBeTrue)
		})

		Convey("non-https is invalid", func() {
			val := "http://example.com/course"
			req := CreateCourseRequest{CanonicalURL: &val}
			So(errors.Is(req.validateCanonicalURL(), ErrInvalidCanonicalURL), ShouldBeTrue)
		})

		Convey("valid https URL", func() {
			val := "https://example.com/course"
			req := CreateCourseRequest{CanonicalURL: &val}
			So(req.validateCanonicalURL(), ShouldBeNil)
		})
	})
}

func TestCreateCourseRequestValidate(t *testing.T) {
	Convey("CreateCourseRequest.Validate", t, func() {
		Convey("all fields valid", func() {
			req := validCreateCourseRequest()
			So(req.Validate(), ShouldBeNil)
		})

		Convey("invalid slug fails", func() {
			req := validCreateCourseRequest()
			req.Slug = ""
			So(errors.Is(req.Validate(), ErrInvalidSlug), ShouldBeTrue)
		})

		Convey("invalid title fails", func() {
			req := validCreateCourseRequest()
			req.Title = ""
			So(errors.Is(req.Validate(), ErrInvalidTitle), ShouldBeTrue)
		})

		Convey("invalid estimated minutes fails", func() {
			req := validCreateCourseRequest()
			val := 0
			req.EstimatedMinutes = &val
			So(errors.Is(req.Validate(), ErrInvalidEstimatedMinutes), ShouldBeTrue)
		})

		Convey("invalid thumbnail URL fails", func() {
			req := validCreateCourseRequest()
			empty := ""
			req.ThumbnailURL = &empty
			So(errors.Is(req.Validate(), ErrInvalidThumbnailURL), ShouldBeTrue)
		})

		Convey("invalid preview video URL fails", func() {
			req := validCreateCourseRequest()
			empty := ""
			req.PreviewVideoURL = &empty
			So(errors.Is(req.Validate(), ErrInvalidPreviewVideoURL), ShouldBeTrue)
		})

		Convey("invalid og image URL fails", func() {
			req := validCreateCourseRequest()
			empty := ""
			req.OgImageURL = &empty
			So(errors.Is(req.Validate(), ErrInvalidOgImageURL), ShouldBeTrue)
		})

		Convey("invalid canonical URL fails", func() {
			req := validCreateCourseRequest()
			empty := ""
			req.CanonicalURL = &empty
			So(errors.Is(req.Validate(), ErrInvalidCanonicalURL), ShouldBeTrue)
		})
	})
}

func TestUpdateCourseRequestValidateID(t *testing.T) {
	Convey("UpdateCourseRequest.validateID", t, func() {
		Convey("valid ID", func() {
			req := UpdateCourseRequest{ID: "b3f1c9e2-6a4d-4f2e-9c1a-2d5e7f8a9b0c"}
			So(req.validateID(), ShouldBeNil)
		})

		Convey("empty ID", func() {
			req := UpdateCourseRequest{ID: ""}
			So(errors.Is(req.validateID(), ErrInvalidCourseID), ShouldBeTrue)
		})

		Convey("malformed (non-UUID) ID", func() {
			req := UpdateCourseRequest{ID: "course-id"}
			So(errors.Is(req.validateID(), ErrInvalidCourseID), ShouldBeTrue)
		})
	})
}

func TestUpdateCourseRequestValidateSlug(t *testing.T) {
	Convey("UpdateCourseRequest.validateSlug", t, func() {
		Convey("nil is valid", func() {
			req := UpdateCourseRequest{Slug: nil}
			So(req.validateSlug(), ShouldBeNil)
		})

		Convey("empty is invalid", func() {
			empty := ""
			req := UpdateCourseRequest{Slug: &empty}
			So(errors.Is(req.validateSlug(), ErrInvalidSlug), ShouldBeTrue)
		})

		Convey("malformed is invalid", func() {
			val := "Not A Slug!"
			req := UpdateCourseRequest{Slug: &val}
			So(errors.Is(req.validateSlug(), ErrInvalidSlug), ShouldBeTrue)
		})

		Convey("valid slug", func() {
			val := "valid-slug"
			req := UpdateCourseRequest{Slug: &val}
			So(req.validateSlug(), ShouldBeNil)
		})
	})
}

func TestUpdateCourseRequestValidateTitle(t *testing.T) {
	Convey("UpdateCourseRequest.validateTitle", t, func() {
		Convey("nil is valid", func() {
			req := UpdateCourseRequest{Title: nil}
			So(req.validateTitle(), ShouldBeNil)
		})

		Convey("whitespace-only is invalid", func() {
			val := "   "
			req := UpdateCourseRequest{Title: &val}
			So(errors.Is(req.validateTitle(), ErrInvalidTitle), ShouldBeTrue)
		})

		Convey("valid title", func() {
			val := "Title"
			req := UpdateCourseRequest{Title: &val}
			So(req.validateTitle(), ShouldBeNil)
		})

		Convey("300 runes is valid (boundary)", func() {
			val := strings.Repeat("A", 300)
			req := UpdateCourseRequest{Title: &val}
			So(req.validateTitle(), ShouldBeNil)
		})

		Convey("301 runes is invalid (boundary)", func() {
			val := strings.Repeat("A", 301)
			req := UpdateCourseRequest{Title: &val}
			So(errors.Is(req.validateTitle(), ErrInvalidTitle), ShouldBeTrue)
		})
	})
}

func TestUpdateCourseRequestValidateDescription(t *testing.T) {
	Convey("UpdateCourseRequest.validateDescription", t, func() {
		Convey("nil is valid", func() {
			req := UpdateCourseRequest{Description: nil}
			So(req.validateDescription(), ShouldBeNil)
		})

		Convey("empty is invalid", func() {
			empty := ""
			req := UpdateCourseRequest{Description: &empty}
			So(errors.Is(req.validateDescription(), ErrInvalidDescription), ShouldBeTrue)
		})

		Convey("valid description", func() {
			val := "not empty"
			req := UpdateCourseRequest{Description: &val}
			So(req.validateDescription(), ShouldBeNil)
		})

		Convey("10000 runes is valid (boundary)", func() {
			val := strings.Repeat("A", 10000)
			req := UpdateCourseRequest{Description: &val}
			So(req.validateDescription(), ShouldBeNil)
		})

		Convey("10001 runes is invalid (boundary)", func() {
			val := strings.Repeat("A", 10001)
			req := UpdateCourseRequest{Description: &val}
			So(errors.Is(req.validateDescription(), ErrInvalidDescription), ShouldBeTrue)
		})
	})
}

func TestUpdateCourseRequestValidateThumbnailURL(t *testing.T) {
	Convey("UpdateCourseRequest.validateThumbnailURL", t, func() {
		Convey("nil is valid", func() {
			req := UpdateCourseRequest{ThumbnailURL: nil}
			So(req.validateThumbnailURL(), ShouldBeNil)
		})

		Convey("non-https is invalid", func() {
			val := "http://example.com/image.png"
			req := UpdateCourseRequest{ThumbnailURL: &val}
			So(errors.Is(req.validateThumbnailURL(), ErrInvalidThumbnailURL), ShouldBeTrue)
		})

		Convey("valid https URL", func() {
			val := "https://example.com/image.png"
			req := UpdateCourseRequest{ThumbnailURL: &val}
			So(req.validateThumbnailURL(), ShouldBeNil)
		})
	})
}

func TestUpdateCourseRequestValidatePreviewVideoURL(t *testing.T) {
	Convey("UpdateCourseRequest.validatePreviewVideoURL", t, func() {
		Convey("nil is valid", func() {
			req := UpdateCourseRequest{PreviewVideoURL: nil}
			So(req.validatePreviewVideoURL(), ShouldBeNil)
		})

		Convey("non-https is invalid", func() {
			val := "http://example.com/video.mp4"
			req := UpdateCourseRequest{PreviewVideoURL: &val}
			So(errors.Is(req.validatePreviewVideoURL(), ErrInvalidPreviewVideoURL), ShouldBeTrue)
		})

		Convey("valid https URL", func() {
			val := "https://example.com/video.mp4"
			req := UpdateCourseRequest{PreviewVideoURL: &val}
			So(req.validatePreviewVideoURL(), ShouldBeNil)
		})
	})
}

func TestUpdateCourseRequestValidateSeoTitle(t *testing.T) {
	Convey("UpdateCourseRequest.validateSeoTitle", t, func() {
		Convey("nil is valid", func() {
			req := UpdateCourseRequest{SeoTitle: nil}
			So(req.validateSeoTitle(), ShouldBeNil)
		})

		Convey("empty is invalid", func() {
			empty := ""
			req := UpdateCourseRequest{SeoTitle: &empty}
			So(errors.Is(req.validateSeoTitle(), ErrInvalidSeoTitle), ShouldBeTrue)
		})

		Convey("valid seo title", func() {
			val := "not empty"
			req := UpdateCourseRequest{SeoTitle: &val}
			So(req.validateSeoTitle(), ShouldBeNil)
		})

		Convey("70 runes is valid (boundary)", func() {
			val := strings.Repeat("A", 70)
			req := UpdateCourseRequest{SeoTitle: &val}
			So(req.validateSeoTitle(), ShouldBeNil)
		})

		Convey("71 runes is invalid (boundary)", func() {
			val := strings.Repeat("A", 71)
			req := UpdateCourseRequest{SeoTitle: &val}
			So(errors.Is(req.validateSeoTitle(), ErrInvalidSeoTitle), ShouldBeTrue)
		})
	})
}

func TestUpdateCourseRequestValidateSeoDescription(t *testing.T) {
	Convey("UpdateCourseRequest.validateSeoDescription", t, func() {
		Convey("nil is valid", func() {
			req := UpdateCourseRequest{SeoDescription: nil}
			So(req.validateSeoDescription(), ShouldBeNil)
		})

		Convey("empty is invalid", func() {
			empty := ""
			req := UpdateCourseRequest{SeoDescription: &empty}
			So(errors.Is(req.validateSeoDescription(), ErrInvalidSeoDescription), ShouldBeTrue)
		})

		Convey("valid seo description", func() {
			val := "not empty"
			req := UpdateCourseRequest{SeoDescription: &val}
			So(req.validateSeoDescription(), ShouldBeNil)
		})

		Convey("160 runes is valid (boundary)", func() {
			val := strings.Repeat("A", 160)
			req := UpdateCourseRequest{SeoDescription: &val}
			So(req.validateSeoDescription(), ShouldBeNil)
		})

		Convey("161 runes is invalid (boundary)", func() {
			val := strings.Repeat("A", 161)
			req := UpdateCourseRequest{SeoDescription: &val}
			So(errors.Is(req.validateSeoDescription(), ErrInvalidSeoDescription), ShouldBeTrue)
		})
	})
}

func TestUpdateCourseRequestValidateOgImageURL(t *testing.T) {
	Convey("UpdateCourseRequest.validateOgImageURL", t, func() {
		Convey("nil is valid", func() {
			req := UpdateCourseRequest{OgImageURL: nil}
			So(req.validateOgImageURL(), ShouldBeNil)
		})

		Convey("non-https is invalid", func() {
			val := "http://example.com/og.png"
			req := UpdateCourseRequest{OgImageURL: &val}
			So(errors.Is(req.validateOgImageURL(), ErrInvalidOgImageURL), ShouldBeTrue)
		})

		Convey("valid https URL", func() {
			val := "https://example.com/og.png"
			req := UpdateCourseRequest{OgImageURL: &val}
			So(req.validateOgImageURL(), ShouldBeNil)
		})
	})
}

func TestUpdateCourseRequestValidateCanonicalURL(t *testing.T) {
	Convey("UpdateCourseRequest.validateCanonicalURL", t, func() {
		Convey("nil is valid", func() {
			req := UpdateCourseRequest{CanonicalURL: nil}
			So(req.validateCanonicalURL(), ShouldBeNil)
		})

		Convey("non-https is invalid", func() {
			val := "http://example.com/course"
			req := UpdateCourseRequest{CanonicalURL: &val}
			So(errors.Is(req.validateCanonicalURL(), ErrInvalidCanonicalURL), ShouldBeTrue)
		})

		Convey("valid https URL", func() {
			val := "https://example.com/course"
			req := UpdateCourseRequest{CanonicalURL: &val}
			So(req.validateCanonicalURL(), ShouldBeNil)
		})
	})
}

func TestUpdateCourseRequestValidateEstimatedMinutes(t *testing.T) {
	Convey("UpdateCourseRequest.validateEstimatedMinutes", t, func() {
		Convey("nil is valid", func() {
			req := UpdateCourseRequest{EstimatedMinutes: nil}
			So(req.validateEstimatedMinutes(), ShouldBeNil)
		})

		Convey("zero is invalid", func() {
			val := 0
			req := UpdateCourseRequest{EstimatedMinutes: &val}
			So(errors.Is(req.validateEstimatedMinutes(), ErrInvalidEstimatedMinutes), ShouldBeTrue)
		})

		Convey("positive value is valid", func() {
			val := 30
			req := UpdateCourseRequest{EstimatedMinutes: &val}
			So(req.validateEstimatedMinutes(), ShouldBeNil)
		})
	})
}

func TestUpdateCourseRequestValidate(t *testing.T) {
	Convey("UpdateCourseRequest.Validate", t, func() {
		validID := "b3f1c9e2-6a4d-4f2e-9c1a-2d5e7f8a9b0c"

		Convey("all nil fields except ID is valid", func() {
			req := UpdateCourseRequest{ID: validID}
			So(req.Validate(), ShouldBeNil)
		})

		Convey("missing ID fails", func() {
			req := UpdateCourseRequest{ID: ""}
			So(errors.Is(req.Validate(), ErrInvalidCourseID), ShouldBeTrue)
		})

		Convey("malformed (non-UUID) ID fails", func() {
			req := UpdateCourseRequest{ID: "course-id"}
			So(errors.Is(req.Validate(), ErrInvalidCourseID), ShouldBeTrue)
		})

		Convey("invalid slug fails", func() {
			empty := ""
			req := UpdateCourseRequest{ID: validID, Slug: &empty}
			So(errors.Is(req.Validate(), ErrInvalidSlug), ShouldBeTrue)
		})

		Convey("invalid title fails", func() {
			empty := "  "
			req := UpdateCourseRequest{ID: validID, Title: &empty}
			So(errors.Is(req.Validate(), ErrInvalidTitle), ShouldBeTrue)
		})

		Convey("invalid description fails", func() {
			empty := ""
			req := UpdateCourseRequest{ID: validID, Description: &empty}
			So(errors.Is(req.Validate(), ErrInvalidDescription), ShouldBeTrue)
		})

		Convey("invalid seo title fails", func() {
			empty := ""
			req := UpdateCourseRequest{ID: validID, SeoTitle: &empty}
			So(errors.Is(req.Validate(), ErrInvalidSeoTitle), ShouldBeTrue)
		})

		Convey("invalid seo description fails", func() {
			empty := ""
			req := UpdateCourseRequest{ID: validID, SeoDescription: &empty}
			So(errors.Is(req.Validate(), ErrInvalidSeoDescription), ShouldBeTrue)
		})

		Convey("invalid estimated minutes fails", func() {
			val := -5
			req := UpdateCourseRequest{ID: validID, EstimatedMinutes: &val}
			So(errors.Is(req.Validate(), ErrInvalidEstimatedMinutes), ShouldBeTrue)
		})
	})
}

func TestUpdateCourseRequestApply(t *testing.T) {
	Convey("UpdateCourseRequest.Apply", t, func() {
		Convey("nil fields leave course unchanged", func() {
			course := Course{
				ID:               "course-id",
				Slug:             "old-slug",
				Title:            "Old Title",
				EstimatedMinutes: nil,
				IsIndexable:      true,
			}
			req := UpdateCourseRequest{ID: "course-id"}
			req.Apply(&course)

			So(course.Slug, ShouldEqual, "old-slug")
			So(course.Title, ShouldEqual, "Old Title")
			So(course.IsIndexable, ShouldBeTrue)
		})

		Convey("non-nil fields overwrite course fields", func() {
			oldMinutes := 10
			course := Course{
				ID:               "course-id",
				Slug:             "old-slug",
				Title:            "Old Title",
				Description:      nil,
				ThumbnailURL:     nil,
				PreviewVideoURL:  nil,
				EstimatedMinutes: &oldMinutes,
				SeoTitle:         nil,
				SeoDescription:   nil,
				OgImageURL:       nil,
				CanonicalURL:     nil,
				IsIndexable:      false,
			}

			newSlug := "new-slug"
			newTitle := "New Title"
			newDescription := "New Description"
			newThumbnailURL := "https://example.com/thumb.png"
			newPreviewVideoURL := "https://example.com/video.mp4"
			newMinutes := 45
			newSeoTitle := "New SEO Title"
			newSeoDescription := "New SEO Description"
			newOgImageURL := "https://example.com/og.png"
			newCanonicalURL := "https://example.com/course"
			newIsIndexable := true

			req := UpdateCourseRequest{
				ID:               "course-id",
				Slug:             &newSlug,
				Title:            &newTitle,
				Description:      &newDescription,
				ThumbnailURL:     &newThumbnailURL,
				PreviewVideoURL:  &newPreviewVideoURL,
				EstimatedMinutes: &newMinutes,
				SeoTitle:         &newSeoTitle,
				SeoDescription:   &newSeoDescription,
				OgImageURL:       &newOgImageURL,
				CanonicalURL:     &newCanonicalURL,
				IsIndexable:      &newIsIndexable,
			}
			req.Apply(&course)

			So(course.Slug, ShouldEqual, newSlug)
			So(course.Title, ShouldEqual, newTitle)
			So(course.Description, ShouldEqual, &newDescription)
			So(course.ThumbnailURL, ShouldEqual, &newThumbnailURL)
			So(course.PreviewVideoURL, ShouldEqual, &newPreviewVideoURL)
			So(course.EstimatedMinutes, ShouldEqual, &newMinutes)
			So(course.SeoTitle, ShouldEqual, &newSeoTitle)
			So(course.SeoDescription, ShouldEqual, &newSeoDescription)
			So(course.OgImageURL, ShouldEqual, &newOgImageURL)
			So(course.CanonicalURL, ShouldEqual, &newCanonicalURL)
			So(course.IsIndexable, ShouldBeTrue)
		})
	})
}
