package contentdomain

import (
	"errors"
	"strings"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestContentItemStatusValid(t *testing.T) {
	Convey("ContentItemStatus.Valid", t, func() {
		Convey("known statuses are valid", func() {
			So(DraftStatus.Valid(), ShouldBeTrue)
			So(PublishedStatus.Valid(), ShouldBeTrue)
			So(ArchivedStatus.Valid(), ShouldBeTrue)
		})

		Convey("unknown status is invalid", func() {
			status := ContentItemStatus("invalid")
			So(status.Valid(), ShouldBeFalse)
		})
	})
}

func TestContentTypeValid(t *testing.T) {
	Convey("ContentType.Valid", t, func() {
		Convey("known types are valid", func() {
			So(VideoContent.Valid(), ShouldBeTrue)
			So(BookContent.Valid(), ShouldBeTrue)
			So(PresentationContent.Valid(), ShouldBeTrue)
		})

		Convey("unknown type is invalid", func() {
			contentType := ContentType("invalid")
			So(contentType.Valid(), ShouldBeFalse)
		})
	})
}

func TestContentItemCheckTitleReady(t *testing.T) {
	Convey("checkTitleReady", t, func() {
		Convey("valid title", func() {
			item := ContentItem{Title: "title"}
			So(item.checkTitleReady(), ShouldBeNil)
			So(item.ReadyToPublish(), ShouldNotBeNil)
		})

		Convey("invalid title", func() {
			item := ContentItem{Title: ""}
			So(errors.Is(item.checkTitleReady(), ErrInvalidTitle), ShouldBeTrue)
			So(item.ReadyToPublish(), ShouldNotBeNil)
		})
	})
}

func TestContentItemCheckDescriptionReady(t *testing.T) {
	Convey("checkDescriptionReady", t, func() {
		Convey("nil description is invalid", func() {
			item := ContentItem{Title: "title", Description: nil}
			So(errors.Is(item.checkDescriptionReady(), ErrInvalidDescription), ShouldBeTrue)
			So(item.ReadyToPublish(), ShouldNotBeNil)
		})

		Convey("empty description is invalid", func() {
			empty := ""
			item := ContentItem{Title: "title", Description: &empty}
			So(errors.Is(item.checkDescriptionReady(), ErrInvalidDescription), ShouldBeTrue)
			So(item.ReadyToPublish(), ShouldNotBeNil)
		})

		Convey("non-empty description is valid", func() {
			notEmpty := "not empty"
			item := ContentItem{Title: "title", Description: &notEmpty}
			So(item.checkDescriptionReady(), ShouldBeNil)
			So(item.ReadyToPublish(), ShouldNotBeNil)
		})
	})
}

func TestContentItemCheckThumbnailURL(t *testing.T) {
	Convey("checkThumbnailURL", t, func() {
		Convey("nil ThumbnailURL is invalid", func() {
			item := ContentItem{Title: "title", ThumbnailURL: nil}
			So(errors.Is(item.checkThumbnailURL(), ErrInvalidThumbnailURL), ShouldBeTrue)
			So(item.ReadyToPublish(), ShouldNotBeNil)
		})

		Convey("empty ThumbnailURL is invalid", func() {
			empty := ""
			item := ContentItem{Title: "title", ThumbnailURL: &empty}
			So(errors.Is(item.checkThumbnailURL(), ErrInvalidThumbnailURL), ShouldBeTrue)
			So(item.ReadyToPublish(), ShouldNotBeNil)
		})

		Convey("non-empty ThumbnailURL is valid", func() {
			notEmpty := "https://example.com/thumb.png"
			item := ContentItem{Title: "title", ThumbnailURL: &notEmpty}
			So(item.checkThumbnailURL(), ShouldBeNil)
			So(item.ReadyToPublish(), ShouldNotBeNil)
		})
	})
}

func TestContentItemCheckSeoTitleReady(t *testing.T) {
	Convey("checkSeoTitleReady", t, func() {
		Convey("nil SeoTitle is invalid", func() {
			item := ContentItem{Title: "title", SeoTitle: nil}
			So(errors.Is(item.checkSeoTitleReady(), ErrInvalidSeoTitle), ShouldBeTrue)
			So(item.ReadyToPublish(), ShouldNotBeNil)
		})

		Convey("empty SeoTitle is invalid", func() {
			empty := ""
			item := ContentItem{Title: "title", SeoTitle: &empty}
			So(errors.Is(item.checkSeoTitleReady(), ErrInvalidSeoTitle), ShouldBeTrue)
			So(item.ReadyToPublish(), ShouldNotBeNil)
		})

		Convey("non-empty SeoTitle is valid", func() {
			notEmpty := "not empty"
			item := ContentItem{Title: "title", SeoTitle: &notEmpty}
			So(item.checkSeoTitleReady(), ShouldBeNil)
			So(item.ReadyToPublish(), ShouldNotBeNil)
		})
	})
}

func TestContentItemCheckSeoDescriptionReady(t *testing.T) {
	Convey("checkSeoDescriptionReady", t, func() {
		Convey("nil SeoDescription is invalid", func() {
			item := ContentItem{Title: "title", SeoDescription: nil}
			So(errors.Is(item.checkSeoDescriptionReady(), ErrInvalidSeoDescription), ShouldBeTrue)
			So(item.ReadyToPublish(), ShouldNotBeNil)
		})

		Convey("empty SeoDescription is invalid", func() {
			empty := ""
			item := ContentItem{Title: "title", SeoDescription: &empty}
			So(errors.Is(item.checkSeoDescriptionReady(), ErrInvalidSeoDescription), ShouldBeTrue)
			So(item.ReadyToPublish(), ShouldNotBeNil)
		})

		Convey("non-empty SeoDescription is valid", func() {
			notEmpty := "not empty"
			item := ContentItem{Title: "title", SeoDescription: &notEmpty}
			So(item.checkSeoDescriptionReady(), ShouldBeNil)
			So(item.ReadyToPublish(), ShouldNotBeNil)
		})
	})
}

func TestContentItemCheckMediaReadyVideo(t *testing.T) {
	Convey("checkMediaReady: video content type", t, func() {
		Convey("missing VideoURL is invalid", func() {
			item := ContentItem{Title: "title", ContentType: VideoContent}
			So(errors.Is(item.checkMediaReady(), ErrInvalidMedia), ShouldBeTrue)
			So(item.ReadyToPublish(), ShouldNotBeNil)
		})

		Convey("empty VideoURL is invalid", func() {
			empty := ""
			item := ContentItem{Title: "title", ContentType: VideoContent, VideoURL: &empty}
			So(errors.Is(item.checkMediaReady(), ErrInvalidMedia), ShouldBeTrue)
			So(item.ReadyToPublish(), ShouldNotBeNil)
		})

		Convey("non-empty VideoURL is valid", func() {
			videoURL := "https://example.com/video.mp4"
			item := ContentItem{Title: "title", ContentType: VideoContent, VideoURL: &videoURL}
			So(item.checkMediaReady(), ShouldBeNil)
		})
	})
}

func TestContentItemCheckMediaReadyBook(t *testing.T) {
	Convey("checkMediaReady: book content type", t, func() {
		Convey("missing Body is invalid", func() {
			item := ContentItem{Title: "title", ContentType: BookContent}
			So(errors.Is(item.checkMediaReady(), ErrInvalidMedia), ShouldBeTrue)
		})

		Convey("empty Body is invalid", func() {
			empty := ""
			item := ContentItem{Title: "title", ContentType: BookContent, Body: &empty}
			So(errors.Is(item.checkMediaReady(), ErrInvalidMedia), ShouldBeTrue)
		})

		Convey("non-empty Body is valid", func() {
			body := "book contents"
			item := ContentItem{Title: "title", ContentType: BookContent, Body: &body}
			So(item.checkMediaReady(), ShouldBeNil)
		})
	})
}

func TestContentItemCheckMediaReadyPresentation(t *testing.T) {
	Convey("checkMediaReady: presentation content type", t, func() {
		Convey("missing FileURL is invalid", func() {
			item := ContentItem{Title: "title", ContentType: PresentationContent}
			So(errors.Is(item.checkMediaReady(), ErrInvalidMedia), ShouldBeTrue)
		})

		Convey("empty FileURL is invalid", func() {
			empty := ""
			item := ContentItem{Title: "title", ContentType: PresentationContent, FileURL: &empty}
			So(errors.Is(item.checkMediaReady(), ErrInvalidMedia), ShouldBeTrue)
		})

		Convey("non-empty FileURL is valid", func() {
			fileURL := "https://example.com/presentation.pdf"
			item := ContentItem{Title: "title", ContentType: PresentationContent, FileURL: &fileURL}
			So(item.checkMediaReady(), ShouldBeNil)
		})
	})
}

func TestContentItemCheckMediaReadyUnknownType(t *testing.T) {
	Convey("checkMediaReady: unknown/empty content type is invalid", t, func() {
		item := ContentItem{Title: "title", ContentType: ContentType("")}
		So(errors.Is(item.checkMediaReady(), ErrInvalidMedia), ShouldBeTrue)
	})
}

func TestContentItemReadyToPublish(t *testing.T) {
	Convey("ReadyToPublish", t, func() {
		notEmpty := "not empty"
		videoURL := "https://example.com/video.mp4"

		Convey("all fields valid for video content type", func() {
			item := ContentItem{
				Title:          "title",
				Description:    &notEmpty,
				ThumbnailURL:   &notEmpty,
				SeoTitle:       &notEmpty,
				SeoDescription: &notEmpty,
				ContentType:    VideoContent,
				VideoURL:       &videoURL,
			}
			So(item.ReadyToPublish(), ShouldBeNil)
		})

		Convey("missing media fails", func() {
			item := ContentItem{
				Title:          "title",
				Description:    &notEmpty,
				ThumbnailURL:   &notEmpty,
				SeoTitle:       &notEmpty,
				SeoDescription: &notEmpty,
				ContentType:    VideoContent,
			}
			So(errors.Is(item.ReadyToPublish(), ErrInvalidMedia), ShouldBeTrue)
		})

		Convey("missing thumbnail fails", func() {
			item := ContentItem{
				Title:          "title",
				Description:    &notEmpty,
				SeoTitle:       &notEmpty,
				SeoDescription: &notEmpty,
				ContentType:    VideoContent,
				VideoURL:       &videoURL,
			}
			So(errors.Is(item.ReadyToPublish(), ErrInvalidThumbnailURL), ShouldBeTrue)
		})
	})
}

func validCreateContentItemRequest() CreateContentItemRequest {
	videoURL := "https://example.com/video.mp4"
	return CreateContentItemRequest{
		Slug:        "valid-slug",
		Title:       "Title",
		ContentType: VideoContent,
		VideoURL:    &videoURL,
	}
}

func TestCreateContentItemRequestValidateSlug(t *testing.T) {
	Convey("CreateContentItemRequest.validateSlug", t, func() {
		Convey("valid slug", func() {
			req := CreateContentItemRequest{Slug: "valid-slug"}
			So(req.validateSlug(), ShouldBeNil)
		})

		Convey("empty slug", func() {
			req := CreateContentItemRequest{Slug: ""}
			So(errors.Is(req.validateSlug(), ErrInvalidSlug), ShouldBeTrue)
		})

		Convey("malformed slug", func() {
			req := CreateContentItemRequest{Slug: "Not A Slug!"}
			So(errors.Is(req.validateSlug(), ErrInvalidSlug), ShouldBeTrue)
		})
	})
}

func TestCreateContentItemRequestValidateTitle(t *testing.T) {
	Convey("CreateContentItemRequest.validateTitle", t, func() {
		Convey("valid title", func() {
			req := CreateContentItemRequest{Title: "Title"}
			So(req.validateTitle(), ShouldBeNil)
		})

		Convey("empty title", func() {
			req := CreateContentItemRequest{Title: ""}
			So(errors.Is(req.validateTitle(), ErrInvalidTitle), ShouldBeTrue)
		})

		Convey("whitespace-only title", func() {
			req := CreateContentItemRequest{Title: "   "}
			So(errors.Is(req.validateTitle(), ErrInvalidTitle), ShouldBeTrue)
		})

		Convey("300 runes is valid (boundary)", func() {
			req := CreateContentItemRequest{Title: strings.Repeat("A", 300)}
			So(req.validateTitle(), ShouldBeNil)
		})

		Convey("301 runes is invalid (boundary)", func() {
			req := CreateContentItemRequest{Title: strings.Repeat("A", 301)}
			So(errors.Is(req.validateTitle(), ErrInvalidTitle), ShouldBeTrue)
		})
	})
}

func TestCreateContentItemRequestValidateDescription(t *testing.T) {
	Convey("CreateContentItemRequest.validateDescription", t, func() {
		Convey("nil is valid", func() {
			req := CreateContentItemRequest{Description: nil}
			So(req.validateDescription(), ShouldBeNil)
		})

		Convey("empty is invalid", func() {
			empty := ""
			req := CreateContentItemRequest{Description: &empty}
			So(errors.Is(req.validateDescription(), ErrInvalidDescription), ShouldBeTrue)
		})

		Convey("10000 runes is valid (boundary)", func() {
			val := strings.Repeat("A", 10000)
			req := CreateContentItemRequest{Description: &val}
			So(req.validateDescription(), ShouldBeNil)
		})

		Convey("10001 runes is invalid (boundary)", func() {
			val := strings.Repeat("A", 10001)
			req := CreateContentItemRequest{Description: &val}
			So(errors.Is(req.validateDescription(), ErrInvalidDescription), ShouldBeTrue)
		})
	})
}

func TestCreateContentItemRequestValidateSeoTitle(t *testing.T) {
	Convey("CreateContentItemRequest.validateSeoTitle", t, func() {
		Convey("nil is valid", func() {
			req := CreateContentItemRequest{SeoTitle: nil}
			So(req.validateSeoTitle(), ShouldBeNil)
		})

		Convey("empty is invalid", func() {
			empty := ""
			req := CreateContentItemRequest{SeoTitle: &empty}
			So(errors.Is(req.validateSeoTitle(), ErrInvalidSeoTitle), ShouldBeTrue)
		})

		Convey("70 runes is valid (boundary)", func() {
			val := strings.Repeat("A", 70)
			req := CreateContentItemRequest{SeoTitle: &val}
			So(req.validateSeoTitle(), ShouldBeNil)
		})

		Convey("71 runes is invalid (boundary)", func() {
			val := strings.Repeat("A", 71)
			req := CreateContentItemRequest{SeoTitle: &val}
			So(errors.Is(req.validateSeoTitle(), ErrInvalidSeoTitle), ShouldBeTrue)
		})
	})
}

func TestCreateContentItemRequestValidateSeoDescription(t *testing.T) {
	Convey("CreateContentItemRequest.validateSeoDescription", t, func() {
		Convey("nil is valid", func() {
			req := CreateContentItemRequest{SeoDescription: nil}
			So(req.validateSeoDescription(), ShouldBeNil)
		})

		Convey("empty is invalid", func() {
			empty := ""
			req := CreateContentItemRequest{SeoDescription: &empty}
			So(errors.Is(req.validateSeoDescription(), ErrInvalidSeoDescription), ShouldBeTrue)
		})

		Convey("160 runes is valid (boundary)", func() {
			val := strings.Repeat("A", 160)
			req := CreateContentItemRequest{SeoDescription: &val}
			So(req.validateSeoDescription(), ShouldBeNil)
		})

		Convey("161 runes is invalid (boundary)", func() {
			val := strings.Repeat("A", 161)
			req := CreateContentItemRequest{SeoDescription: &val}
			So(errors.Is(req.validateSeoDescription(), ErrInvalidSeoDescription), ShouldBeTrue)
		})
	})
}

func TestCreateContentItemRequestValidateEstimatedMinutes(t *testing.T) {
	Convey("CreateContentItemRequest.validateEstimatedMinutes", t, func() {
		Convey("nil is valid", func() {
			req := CreateContentItemRequest{EstimatedMinutes: nil}
			So(req.validateEstimatedMinutes(), ShouldBeNil)
		})

		Convey("positive value is valid", func() {
			val := 30
			req := CreateContentItemRequest{EstimatedMinutes: &val}
			So(req.validateEstimatedMinutes(), ShouldBeNil)
		})

		Convey("zero is invalid", func() {
			val := 0
			req := CreateContentItemRequest{EstimatedMinutes: &val}
			So(errors.Is(req.validateEstimatedMinutes(), ErrInvalidEstimatedMinutes), ShouldBeTrue)
		})

		Convey("negative value is invalid", func() {
			val := -1
			req := CreateContentItemRequest{EstimatedMinutes: &val}
			So(errors.Is(req.validateEstimatedMinutes(), ErrInvalidEstimatedMinutes), ShouldBeTrue)
		})
	})
}

func TestCreateContentItemRequestValidateEstimatedPages(t *testing.T) {
	Convey("CreateContentItemRequest.validateEstimatedPages", t, func() {
		Convey("nil is valid", func() {
			req := CreateContentItemRequest{EstimatedPages: nil}
			So(req.validateEstimatedPages(), ShouldBeNil)
		})

		Convey("positive value is valid", func() {
			val := 120
			req := CreateContentItemRequest{EstimatedPages: &val}
			So(req.validateEstimatedPages(), ShouldBeNil)
		})

		Convey("zero is invalid", func() {
			val := 0
			req := CreateContentItemRequest{EstimatedPages: &val}
			So(errors.Is(req.validateEstimatedPages(), ErrInvalidEstimatedPages), ShouldBeTrue)
		})

		Convey("negative value is invalid", func() {
			val := -1
			req := CreateContentItemRequest{EstimatedPages: &val}
			So(errors.Is(req.validateEstimatedPages(), ErrInvalidEstimatedPages), ShouldBeTrue)
		})
	})
}

func TestCreateContentItemRequestValidateThumbnailURL(t *testing.T) {
	Convey("CreateContentItemRequest.validateThumbnailURL", t, func() {
		Convey("nil is valid", func() {
			req := CreateContentItemRequest{ThumbnailURL: nil}
			So(req.validateThumbnailURL(), ShouldBeNil)
		})

		Convey("empty is invalid", func() {
			empty := ""
			req := CreateContentItemRequest{ThumbnailURL: &empty}
			So(errors.Is(req.validateThumbnailURL(), ErrInvalidThumbnailURL), ShouldBeTrue)
		})

		Convey("non-https is invalid", func() {
			val := "http://example.com/image.png"
			req := CreateContentItemRequest{ThumbnailURL: &val}
			So(errors.Is(req.validateThumbnailURL(), ErrInvalidThumbnailURL), ShouldBeTrue)
		})

		Convey("valid https URL", func() {
			val := "https://example.com/image.png"
			req := CreateContentItemRequest{ThumbnailURL: &val}
			So(req.validateThumbnailURL(), ShouldBeNil)
		})
	})
}

func TestCreateContentItemRequestValidateOgImageURL(t *testing.T) {
	Convey("CreateContentItemRequest.validateOgImageURL", t, func() {
		Convey("nil is valid", func() {
			req := CreateContentItemRequest{OgImageURL: nil}
			So(req.validateOgImageURL(), ShouldBeNil)
		})

		Convey("empty is invalid", func() {
			empty := ""
			req := CreateContentItemRequest{OgImageURL: &empty}
			So(errors.Is(req.validateOgImageURL(), ErrInvalidOgImageURL), ShouldBeTrue)
		})

		Convey("non-https is invalid", func() {
			val := "http://example.com/og.png"
			req := CreateContentItemRequest{OgImageURL: &val}
			So(errors.Is(req.validateOgImageURL(), ErrInvalidOgImageURL), ShouldBeTrue)
		})

		Convey("valid https URL", func() {
			val := "https://example.com/og.png"
			req := CreateContentItemRequest{OgImageURL: &val}
			So(req.validateOgImageURL(), ShouldBeNil)
		})
	})
}

func TestCreateContentItemRequestValidateCanonicalURL(t *testing.T) {
	Convey("CreateContentItemRequest.validateCanonicalURL", t, func() {
		Convey("nil is valid", func() {
			req := CreateContentItemRequest{CanonicalURL: nil}
			So(req.validateCanonicalURL(), ShouldBeNil)
		})

		Convey("empty is invalid", func() {
			empty := ""
			req := CreateContentItemRequest{CanonicalURL: &empty}
			So(errors.Is(req.validateCanonicalURL(), ErrInvalidCanonicalURL), ShouldBeTrue)
		})

		Convey("non-https is invalid", func() {
			val := "http://example.com/content-item"
			req := CreateContentItemRequest{CanonicalURL: &val}
			So(errors.Is(req.validateCanonicalURL(), ErrInvalidCanonicalURL), ShouldBeTrue)
		})

		Convey("valid https URL", func() {
			val := "https://example.com/content-item"
			req := CreateContentItemRequest{CanonicalURL: &val}
			So(req.validateCanonicalURL(), ShouldBeNil)
		})
	})
}

func TestCreateContentItemRequestValidateFileURL(t *testing.T) {
	Convey("CreateContentItemRequest.validateFileURL", t, func() {
		Convey("nil is valid", func() {
			req := CreateContentItemRequest{FileURL: nil}
			So(req.validateFileURL(), ShouldBeNil)
		})

		Convey("empty is invalid", func() {
			empty := ""
			req := CreateContentItemRequest{FileURL: &empty}
			So(errors.Is(req.validateFileURL(), ErrInvalidMedia), ShouldBeTrue)
		})

		Convey("non-https is invalid", func() {
			val := "http://example.com/book.pdf"
			req := CreateContentItemRequest{FileURL: &val}
			So(errors.Is(req.validateFileURL(), ErrInvalidMedia), ShouldBeTrue)
		})

		Convey("valid https URL", func() {
			val := "https://example.com/book.pdf"
			req := CreateContentItemRequest{FileURL: &val}
			So(req.validateFileURL(), ShouldBeNil)
		})
	})
}

func TestCreateContentItemRequestValidateVideoURL(t *testing.T) {
	Convey("CreateContentItemRequest.validateVideoURL", t, func() {
		Convey("nil is valid", func() {
			req := CreateContentItemRequest{VideoURL: nil}
			So(req.validateVideoURL(), ShouldBeNil)
		})

		Convey("empty is invalid", func() {
			empty := ""
			req := CreateContentItemRequest{VideoURL: &empty}
			So(errors.Is(req.validateVideoURL(), ErrInvalidMedia), ShouldBeTrue)
		})

		Convey("non-https is invalid", func() {
			val := "http://example.com/video.mp4"
			req := CreateContentItemRequest{VideoURL: &val}
			So(errors.Is(req.validateVideoURL(), ErrInvalidMedia), ShouldBeTrue)
		})

		Convey("valid https URL", func() {
			val := "https://example.com/video.mp4"
			req := CreateContentItemRequest{VideoURL: &val}
			So(req.validateVideoURL(), ShouldBeNil)
		})
	})
}

func TestCreateContentItemRequestValidateBody(t *testing.T) {
	Convey("CreateContentItemRequest.validateBody", t, func() {
		Convey("nil is valid", func() {
			req := CreateContentItemRequest{Body: nil}
			So(req.validateBody(), ShouldBeNil)
		})

		Convey("empty is invalid", func() {
			empty := ""
			req := CreateContentItemRequest{Body: &empty}
			So(errors.Is(req.validateBody(), ErrInvalidMedia), ShouldBeTrue)
		})

		Convey("10000 runes is valid (boundary)", func() {
			val := strings.Repeat("A", 10000)
			req := CreateContentItemRequest{Body: &val}
			So(req.validateBody(), ShouldBeNil)
		})

		Convey("10001 runes is invalid (boundary)", func() {
			val := strings.Repeat("A", 10001)
			req := CreateContentItemRequest{Body: &val}
			So(errors.Is(req.validateBody(), ErrInvalidMedia), ShouldBeTrue)
		})
	})
}

func TestCreateContentItemRequestValidate(t *testing.T) {
	Convey("CreateContentItemRequest.Validate", t, func() {
		Convey("all fields valid", func() {
			req := validCreateContentItemRequest()
			So(req.Validate(), ShouldBeNil)
		})

		Convey("invalid slug fails", func() {
			req := validCreateContentItemRequest()
			req.Slug = ""
			So(errors.Is(req.Validate(), ErrInvalidSlug), ShouldBeTrue)
		})

		Convey("invalid title fails", func() {
			req := validCreateContentItemRequest()
			req.Title = ""
			So(errors.Is(req.Validate(), ErrInvalidTitle), ShouldBeTrue)
		})

		Convey("invalid estimated minutes fails", func() {
			req := validCreateContentItemRequest()
			val := 0
			req.EstimatedMinutes = &val
			So(errors.Is(req.Validate(), ErrInvalidEstimatedMinutes), ShouldBeTrue)
		})

		Convey("invalid estimated pages fails", func() {
			req := validCreateContentItemRequest()
			val := 0
			req.EstimatedPages = &val
			So(errors.Is(req.Validate(), ErrInvalidEstimatedPages), ShouldBeTrue)
		})

		Convey("invalid thumbnail URL fails", func() {
			req := validCreateContentItemRequest()
			empty := ""
			req.ThumbnailURL = &empty
			So(errors.Is(req.Validate(), ErrInvalidThumbnailURL), ShouldBeTrue)
		})

		Convey("invalid og image URL fails", func() {
			req := validCreateContentItemRequest()
			empty := ""
			req.OgImageURL = &empty
			So(errors.Is(req.Validate(), ErrInvalidOgImageURL), ShouldBeTrue)
		})

		Convey("invalid canonical URL fails", func() {
			req := validCreateContentItemRequest()
			empty := ""
			req.CanonicalURL = &empty
			So(errors.Is(req.Validate(), ErrInvalidCanonicalURL), ShouldBeTrue)
		})

		Convey("invalid file URL fails", func() {
			req := validCreateContentItemRequest()
			empty := ""
			req.FileURL = &empty
			So(errors.Is(req.Validate(), ErrInvalidMedia), ShouldBeTrue)
		})

		Convey("invalid video URL fails", func() {
			req := validCreateContentItemRequest()
			empty := ""
			req.VideoURL = &empty
			So(errors.Is(req.Validate(), ErrInvalidMedia), ShouldBeTrue)
		})

		Convey("invalid body fails", func() {
			req := validCreateContentItemRequest()
			empty := ""
			req.Body = &empty
			So(errors.Is(req.Validate(), ErrInvalidMedia), ShouldBeTrue)
		})
	})
}

func TestUpdateContentItemRequestValidateID(t *testing.T) {
	Convey("UpdateContentItemRequest.validateID", t, func() {
		Convey("valid ID", func() {
			req := UpdateContentItemRequest{ID: "b3f1c9e2-6a4d-4f2e-9c1a-2d5e7f8a9b0c"}
			So(req.validateID(), ShouldBeNil)
		})

		Convey("empty ID", func() {
			req := UpdateContentItemRequest{ID: ""}
			So(errors.Is(req.validateID(), ErrInvalidContentItemID), ShouldBeTrue)
		})

		Convey("malformed (non-UUID) ID", func() {
			req := UpdateContentItemRequest{ID: "content-item-id"}
			So(errors.Is(req.validateID(), ErrInvalidContentItemID), ShouldBeTrue)
		})
	})
}

func TestUpdateContentItemRequestValidateSlug(t *testing.T) {
	Convey("UpdateContentItemRequest.validateSlug", t, func() {
		Convey("nil is valid", func() {
			req := UpdateContentItemRequest{Slug: nil}
			So(req.validateSlug(), ShouldBeNil)
		})

		Convey("empty is invalid", func() {
			empty := ""
			req := UpdateContentItemRequest{Slug: &empty}
			So(errors.Is(req.validateSlug(), ErrInvalidSlug), ShouldBeTrue)
		})

		Convey("malformed is invalid", func() {
			val := "Not A Slug!"
			req := UpdateContentItemRequest{Slug: &val}
			So(errors.Is(req.validateSlug(), ErrInvalidSlug), ShouldBeTrue)
		})

		Convey("valid slug", func() {
			val := "valid-slug"
			req := UpdateContentItemRequest{Slug: &val}
			So(req.validateSlug(), ShouldBeNil)
		})
	})
}

func TestUpdateContentItemRequestValidateTitle(t *testing.T) {
	Convey("UpdateContentItemRequest.validateTitle", t, func() {
		Convey("nil is valid", func() {
			req := UpdateContentItemRequest{Title: nil}
			So(req.validateTitle(), ShouldBeNil)
		})

		Convey("whitespace-only is invalid", func() {
			val := "   "
			req := UpdateContentItemRequest{Title: &val}
			So(errors.Is(req.validateTitle(), ErrInvalidTitle), ShouldBeTrue)
		})

		Convey("valid title", func() {
			val := "Title"
			req := UpdateContentItemRequest{Title: &val}
			So(req.validateTitle(), ShouldBeNil)
		})

		Convey("300 runes is valid (boundary)", func() {
			val := strings.Repeat("A", 300)
			req := UpdateContentItemRequest{Title: &val}
			So(req.validateTitle(), ShouldBeNil)
		})

		Convey("301 runes is invalid (boundary)", func() {
			val := strings.Repeat("A", 301)
			req := UpdateContentItemRequest{Title: &val}
			So(errors.Is(req.validateTitle(), ErrInvalidTitle), ShouldBeTrue)
		})
	})
}

func TestUpdateContentItemRequestValidateDescription(t *testing.T) {
	Convey("UpdateContentItemRequest.validateDescription", t, func() {
		Convey("nil is valid", func() {
			req := UpdateContentItemRequest{Description: nil}
			So(req.validateDescription(), ShouldBeNil)
		})

		Convey("empty is invalid", func() {
			empty := ""
			req := UpdateContentItemRequest{Description: &empty}
			So(errors.Is(req.validateDescription(), ErrInvalidDescription), ShouldBeTrue)
		})

		Convey("valid description", func() {
			val := "not empty"
			req := UpdateContentItemRequest{Description: &val}
			So(req.validateDescription(), ShouldBeNil)
		})

		Convey("10000 runes is valid (boundary)", func() {
			val := strings.Repeat("A", 10000)
			req := UpdateContentItemRequest{Description: &val}
			So(req.validateDescription(), ShouldBeNil)
		})

		Convey("10001 runes is invalid (boundary)", func() {
			val := strings.Repeat("A", 10001)
			req := UpdateContentItemRequest{Description: &val}
			So(errors.Is(req.validateDescription(), ErrInvalidDescription), ShouldBeTrue)
		})
	})
}

func TestUpdateContentItemRequestValidateThumbnailURL(t *testing.T) {
	Convey("UpdateContentItemRequest.validateThumbnailURL", t, func() {
		Convey("nil is valid", func() {
			req := UpdateContentItemRequest{ThumbnailURL: nil}
			So(req.validateThumbnailURL(), ShouldBeNil)
		})

		Convey("non-https is invalid", func() {
			val := "http://example.com/image.png"
			req := UpdateContentItemRequest{ThumbnailURL: &val}
			So(errors.Is(req.validateThumbnailURL(), ErrInvalidThumbnailURL), ShouldBeTrue)
		})

		Convey("valid https URL", func() {
			val := "https://example.com/image.png"
			req := UpdateContentItemRequest{ThumbnailURL: &val}
			So(req.validateThumbnailURL(), ShouldBeNil)
		})
	})
}

func TestUpdateContentItemRequestValidateSeoTitle(t *testing.T) {
	Convey("UpdateContentItemRequest.validateSeoTitle", t, func() {
		Convey("nil is valid", func() {
			req := UpdateContentItemRequest{SeoTitle: nil}
			So(req.validateSeoTitle(), ShouldBeNil)
		})

		Convey("empty is invalid", func() {
			empty := ""
			req := UpdateContentItemRequest{SeoTitle: &empty}
			So(errors.Is(req.validateSeoTitle(), ErrInvalidSeoTitle), ShouldBeTrue)
		})

		Convey("valid seo title", func() {
			val := "not empty"
			req := UpdateContentItemRequest{SeoTitle: &val}
			So(req.validateSeoTitle(), ShouldBeNil)
		})

		Convey("70 runes is valid (boundary)", func() {
			val := strings.Repeat("A", 70)
			req := UpdateContentItemRequest{SeoTitle: &val}
			So(req.validateSeoTitle(), ShouldBeNil)
		})

		Convey("71 runes is invalid (boundary)", func() {
			val := strings.Repeat("A", 71)
			req := UpdateContentItemRequest{SeoTitle: &val}
			So(errors.Is(req.validateSeoTitle(), ErrInvalidSeoTitle), ShouldBeTrue)
		})
	})
}

func TestUpdateContentItemRequestValidateSeoDescription(t *testing.T) {
	Convey("UpdateContentItemRequest.validateSeoDescription", t, func() {
		Convey("nil is valid", func() {
			req := UpdateContentItemRequest{SeoDescription: nil}
			So(req.validateSeoDescription(), ShouldBeNil)
		})

		Convey("empty is invalid", func() {
			empty := ""
			req := UpdateContentItemRequest{SeoDescription: &empty}
			So(errors.Is(req.validateSeoDescription(), ErrInvalidSeoDescription), ShouldBeTrue)
		})

		Convey("valid seo description", func() {
			val := "not empty"
			req := UpdateContentItemRequest{SeoDescription: &val}
			So(req.validateSeoDescription(), ShouldBeNil)
		})

		Convey("160 runes is valid (boundary)", func() {
			val := strings.Repeat("A", 160)
			req := UpdateContentItemRequest{SeoDescription: &val}
			So(req.validateSeoDescription(), ShouldBeNil)
		})

		Convey("161 runes is invalid (boundary)", func() {
			val := strings.Repeat("A", 161)
			req := UpdateContentItemRequest{SeoDescription: &val}
			So(errors.Is(req.validateSeoDescription(), ErrInvalidSeoDescription), ShouldBeTrue)
		})
	})
}

func TestUpdateContentItemRequestValidateOgImageURL(t *testing.T) {
	Convey("UpdateContentItemRequest.validateOgImageURL", t, func() {
		Convey("nil is valid", func() {
			req := UpdateContentItemRequest{OgImageURL: nil}
			So(req.validateOgImageURL(), ShouldBeNil)
		})

		Convey("non-https is invalid", func() {
			val := "http://example.com/og.png"
			req := UpdateContentItemRequest{OgImageURL: &val}
			So(errors.Is(req.validateOgImageURL(), ErrInvalidOgImageURL), ShouldBeTrue)
		})

		Convey("valid https URL", func() {
			val := "https://example.com/og.png"
			req := UpdateContentItemRequest{OgImageURL: &val}
			So(req.validateOgImageURL(), ShouldBeNil)
		})
	})
}

func TestUpdateContentItemRequestValidateCanonicalURL(t *testing.T) {
	Convey("UpdateContentItemRequest.validateCanonicalURL", t, func() {
		Convey("nil is valid", func() {
			req := UpdateContentItemRequest{CanonicalURL: nil}
			So(req.validateCanonicalURL(), ShouldBeNil)
		})

		Convey("non-https is invalid", func() {
			val := "http://example.com/content-item"
			req := UpdateContentItemRequest{CanonicalURL: &val}
			So(errors.Is(req.validateCanonicalURL(), ErrInvalidCanonicalURL), ShouldBeTrue)
		})

		Convey("valid https URL", func() {
			val := "https://example.com/content-item"
			req := UpdateContentItemRequest{CanonicalURL: &val}
			So(req.validateCanonicalURL(), ShouldBeNil)
		})
	})
}

func TestUpdateContentItemRequestValidateEstimatedMinutes(t *testing.T) {
	Convey("UpdateContentItemRequest.validateEstimatedMinutes", t, func() {
		Convey("nil is valid", func() {
			req := UpdateContentItemRequest{EstimatedMinutes: nil}
			So(req.validateEstimatedMinutes(), ShouldBeNil)
		})

		Convey("zero is invalid", func() {
			val := 0
			req := UpdateContentItemRequest{EstimatedMinutes: &val}
			So(errors.Is(req.validateEstimatedMinutes(), ErrInvalidEstimatedMinutes), ShouldBeTrue)
		})

		Convey("positive value is valid", func() {
			val := 30
			req := UpdateContentItemRequest{EstimatedMinutes: &val}
			So(req.validateEstimatedMinutes(), ShouldBeNil)
		})
	})
}

func TestUpdateContentItemRequestValidateEstimatedPages(t *testing.T) {
	Convey("UpdateContentItemRequest.validateEstimatedPages", t, func() {
		Convey("nil is valid", func() {
			req := UpdateContentItemRequest{EstimatedPages: nil}
			So(req.validateEstimatedPages(), ShouldBeNil)
		})

		Convey("zero is invalid", func() {
			val := 0
			req := UpdateContentItemRequest{EstimatedPages: &val}
			So(errors.Is(req.validateEstimatedPages(), ErrInvalidEstimatedPages), ShouldBeTrue)
		})

		Convey("positive value is valid", func() {
			val := 120
			req := UpdateContentItemRequest{EstimatedPages: &val}
			So(req.validateEstimatedPages(), ShouldBeNil)
		})
	})
}

func TestUpdateContentItemRequestValidateFileURL(t *testing.T) {
	Convey("UpdateContentItemRequest.validateFileURL", t, func() {
		Convey("nil is valid", func() {
			req := UpdateContentItemRequest{FileURL: nil}
			So(req.validateFileURL(), ShouldBeNil)
		})

		Convey("non-https is invalid", func() {
			val := "http://example.com/book.pdf"
			req := UpdateContentItemRequest{FileURL: &val}
			So(errors.Is(req.validateFileURL(), ErrInvalidMedia), ShouldBeTrue)
		})

		Convey("valid https URL", func() {
			val := "https://example.com/book.pdf"
			req := UpdateContentItemRequest{FileURL: &val}
			So(req.validateFileURL(), ShouldBeNil)
		})
	})
}

func TestUpdateContentItemRequestValidateVideoURL(t *testing.T) {
	Convey("UpdateContentItemRequest.validateVideoURL", t, func() {
		Convey("nil is valid", func() {
			req := UpdateContentItemRequest{VideoURL: nil}
			So(req.validateVideoURL(), ShouldBeNil)
		})

		Convey("non-https is invalid", func() {
			val := "http://example.com/video.mp4"
			req := UpdateContentItemRequest{VideoURL: &val}
			So(errors.Is(req.validateVideoURL(), ErrInvalidMedia), ShouldBeTrue)
		})

		Convey("valid https URL", func() {
			val := "https://example.com/video.mp4"
			req := UpdateContentItemRequest{VideoURL: &val}
			So(req.validateVideoURL(), ShouldBeNil)
		})
	})
}

func TestUpdateContentItemRequestValidateBody(t *testing.T) {
	Convey("UpdateContentItemRequest.validateBody", t, func() {
		Convey("nil is valid", func() {
			req := UpdateContentItemRequest{Body: nil}
			So(req.validateBody(), ShouldBeNil)
		})

		Convey("empty is invalid", func() {
			empty := ""
			req := UpdateContentItemRequest{Body: &empty}
			So(errors.Is(req.validateBody(), ErrInvalidMedia), ShouldBeTrue)
		})

		Convey("valid body", func() {
			val := "slide contents"
			req := UpdateContentItemRequest{Body: &val}
			So(req.validateBody(), ShouldBeNil)
		})

		Convey("10000 runes is valid (boundary)", func() {
			val := strings.Repeat("A", 10000)
			req := UpdateContentItemRequest{Body: &val}
			So(req.validateBody(), ShouldBeNil)
		})

		Convey("10001 runes is invalid (boundary)", func() {
			val := strings.Repeat("A", 10001)
			req := UpdateContentItemRequest{Body: &val}
			So(errors.Is(req.validateBody(), ErrInvalidMedia), ShouldBeTrue)
		})
	})
}

func TestUpdateContentItemRequestValidate(t *testing.T) {
	Convey("UpdateContentItemRequest.Validate", t, func() {
		validID := "b3f1c9e2-6a4d-4f2e-9c1a-2d5e7f8a9b0c"

		Convey("all nil fields except ID is valid", func() {
			req := UpdateContentItemRequest{ID: validID}
			So(req.Validate(), ShouldBeNil)
		})

		Convey("missing ID fails", func() {
			req := UpdateContentItemRequest{ID: ""}
			So(errors.Is(req.Validate(), ErrInvalidContentItemID), ShouldBeTrue)
		})

		Convey("malformed (non-UUID) ID fails", func() {
			req := UpdateContentItemRequest{ID: "content-item-id"}
			So(errors.Is(req.Validate(), ErrInvalidContentItemID), ShouldBeTrue)
		})

		Convey("invalid slug fails", func() {
			empty := ""
			req := UpdateContentItemRequest{ID: validID, Slug: &empty}
			So(errors.Is(req.Validate(), ErrInvalidSlug), ShouldBeTrue)
		})

		Convey("invalid title fails", func() {
			empty := "  "
			req := UpdateContentItemRequest{ID: validID, Title: &empty}
			So(errors.Is(req.Validate(), ErrInvalidTitle), ShouldBeTrue)
		})

		Convey("invalid description fails", func() {
			empty := ""
			req := UpdateContentItemRequest{ID: validID, Description: &empty}
			So(errors.Is(req.Validate(), ErrInvalidDescription), ShouldBeTrue)
		})

		Convey("invalid seo title fails", func() {
			empty := ""
			req := UpdateContentItemRequest{ID: validID, SeoTitle: &empty}
			So(errors.Is(req.Validate(), ErrInvalidSeoTitle), ShouldBeTrue)
		})

		Convey("invalid seo description fails", func() {
			empty := ""
			req := UpdateContentItemRequest{ID: validID, SeoDescription: &empty}
			So(errors.Is(req.Validate(), ErrInvalidSeoDescription), ShouldBeTrue)
		})

		Convey("invalid estimated minutes fails", func() {
			val := -5
			req := UpdateContentItemRequest{ID: validID, EstimatedMinutes: &val}
			So(errors.Is(req.Validate(), ErrInvalidEstimatedMinutes), ShouldBeTrue)
		})

		Convey("invalid estimated pages fails", func() {
			val := -5
			req := UpdateContentItemRequest{ID: validID, EstimatedPages: &val}
			So(errors.Is(req.Validate(), ErrInvalidEstimatedPages), ShouldBeTrue)
		})

		Convey("invalid file URL fails", func() {
			empty := ""
			req := UpdateContentItemRequest{ID: validID, FileURL: &empty}
			So(errors.Is(req.Validate(), ErrInvalidMedia), ShouldBeTrue)
		})

		Convey("invalid video URL fails", func() {
			empty := ""
			req := UpdateContentItemRequest{ID: validID, VideoURL: &empty}
			So(errors.Is(req.Validate(), ErrInvalidMedia), ShouldBeTrue)
		})

		Convey("invalid body fails", func() {
			empty := ""
			req := UpdateContentItemRequest{ID: validID, Body: &empty}
			So(errors.Is(req.Validate(), ErrInvalidMedia), ShouldBeTrue)
		})
	})
}

func TestUpdateContentItemRequestApplyNilFields(t *testing.T) {
	Convey("UpdateContentItemRequest.Apply: nil fields leave ContentItem unchanged", t, func() {
		item := ContentItem{
			ID:               "content-item-id",
			Slug:             "old-slug",
			Title:            "Old Title",
			ContentType:      VideoContent,
			EstimatedMinutes: nil,
			IsIndexable:      true,
		}
		req := UpdateContentItemRequest{ID: "content-item-id"}
		req.Apply(&item)

		So(item.Slug, ShouldEqual, "old-slug")
		So(item.Title, ShouldEqual, "Old Title")
		So(item.ContentType, ShouldEqual, VideoContent)
		So(item.IsIndexable, ShouldBeTrue)
	})
}

func TestUpdateContentItemRequestApplyNonNilFields(t *testing.T) {
	Convey("UpdateContentItemRequest.Apply: non-nil fields overwrite ContentItem fields", t, func() {
		oldMinutes := 10
		oldPages := 5
		item := ContentItem{
			ID:               "content-item-id",
			Slug:             "old-slug",
			Title:            "Old Title",
			ContentType:      VideoContent,
			Description:      nil,
			ThumbnailURL:     nil,
			Body:             nil,
			VideoURL:         nil,
			FileURL:          nil,
			EstimatedMinutes: &oldMinutes,
			EstimatedPages:   &oldPages,
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
		newBody := "New body"
		newVideoURL := "https://example.com/video.mp4"
		newFileURL := "https://example.com/book.pdf"
		newMinutes := 45
		newPages := 90
		newSeoTitle := "New SEO Title"
		newSeoDescription := "New SEO Description"
		newOgImageURL := "https://example.com/og.png"
		newCanonicalURL := "https://example.com/content-item"
		newIsIndexable := true

		req := UpdateContentItemRequest{
			ID:               "content-item-id",
			Slug:             &newSlug,
			Title:            &newTitle,
			Description:      &newDescription,
			ThumbnailURL:     &newThumbnailURL,
			Body:             &newBody,
			VideoURL:         &newVideoURL,
			FileURL:          &newFileURL,
			EstimatedMinutes: &newMinutes,
			EstimatedPages:   &newPages,
			SeoTitle:         &newSeoTitle,
			SeoDescription:   &newSeoDescription,
			OgImageURL:       &newOgImageURL,
			CanonicalURL:     &newCanonicalURL,
			IsIndexable:      &newIsIndexable,
		}
		req.Apply(&item)

		So(item.Slug, ShouldEqual, newSlug)
		So(item.Title, ShouldEqual, newTitle)
		So(item.ContentType, ShouldEqual, VideoContent) // content_type is create-only, not patchable
		So(item.Description, ShouldEqual, &newDescription)
		So(item.ThumbnailURL, ShouldEqual, &newThumbnailURL)
		So(item.Body, ShouldEqual, &newBody)
		So(item.VideoURL, ShouldEqual, &newVideoURL)
		So(item.FileURL, ShouldEqual, &newFileURL)
		So(item.EstimatedMinutes, ShouldEqual, &newMinutes)
		So(item.EstimatedPages, ShouldEqual, &newPages)
		So(item.SeoTitle, ShouldEqual, &newSeoTitle)
		So(item.SeoDescription, ShouldEqual, &newSeoDescription)
		So(item.OgImageURL, ShouldEqual, &newOgImageURL)
		So(item.CanonicalURL, ShouldEqual, &newCanonicalURL)
		So(item.IsIndexable, ShouldBeTrue)
	})
}
