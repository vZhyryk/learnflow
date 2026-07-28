package articledomain

import (
	"errors"
	"strings"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestArticleStatusValid(t *testing.T) {
	Convey("ArticleStatus.Valid", t, func() {
		Convey("known statuses are valid", func() {
			So(DraftStatus.Valid(), ShouldBeTrue)
			So(PublishedStatus.Valid(), ShouldBeTrue)
			So(ArchivedStatus.Valid(), ShouldBeTrue)
		})

		Convey("unknown status is invalid", func() {
			status := ArticleStatus("invalid")
			So(status.Valid(), ShouldBeFalse)
		})
	})
}

func TestArticleCheckTitleReady(t *testing.T) {
	Convey("checkTitleReady", t, func() {
		Convey("valid title", func() {
			item := Article{Title: "title"}
			So(item.checkTitleReady(), ShouldBeNil)
			So(item.ReadyToPublish(), ShouldNotBeNil)
		})

		Convey("invalid title", func() {
			item := Article{Title: ""}
			So(errors.Is(item.checkTitleReady(), ErrInvalidTitle), ShouldBeTrue)
			So(item.ReadyToPublish(), ShouldNotBeNil)
		})
	})
}

func TestArticleCheckSeoTitleReady(t *testing.T) {
	Convey("checkSeoTitleReady", t, func() {
		Convey("nil SeoTitle is invalid", func() {
			item := Article{Title: "title", SeoTitle: nil}
			So(errors.Is(item.checkSeoTitleReady(), ErrInvalidSeoTitle), ShouldBeTrue)
			So(item.ReadyToPublish(), ShouldNotBeNil)
		})

		Convey("empty SeoTitle is invalid", func() {
			empty := ""
			item := Article{Title: "title", SeoTitle: &empty}
			So(errors.Is(item.checkSeoTitleReady(), ErrInvalidSeoTitle), ShouldBeTrue)
			So(item.ReadyToPublish(), ShouldNotBeNil)
		})

		Convey("non-empty SeoTitle is valid", func() {
			notEmpty := "not empty"
			item := Article{Title: "title", SeoTitle: &notEmpty}
			So(item.checkSeoTitleReady(), ShouldBeNil)
			So(item.ReadyToPublish(), ShouldNotBeNil)
		})
	})
}

func TestArticleCheckSeoDescriptionReady(t *testing.T) {
	Convey("checkSeoDescriptionReady", t, func() {
		Convey("nil SeoDescription is invalid", func() {
			item := Article{Title: "title", SeoDescription: nil}
			So(errors.Is(item.checkSeoDescriptionReady(), ErrInvalidSeoDescription), ShouldBeTrue)
			So(item.ReadyToPublish(), ShouldNotBeNil)
		})

		Convey("empty SeoDescription is invalid", func() {
			empty := ""
			item := Article{Title: "title", SeoDescription: &empty}
			So(errors.Is(item.checkSeoDescriptionReady(), ErrInvalidSeoDescription), ShouldBeTrue)
			So(item.ReadyToPublish(), ShouldNotBeNil)
		})

		Convey("non-empty SeoDescription is valid", func() {
			notEmpty := "not empty"
			item := Article{Title: "title", SeoDescription: &notEmpty}
			So(item.checkSeoDescriptionReady(), ShouldBeNil)
			So(item.ReadyToPublish(), ShouldNotBeNil)
		})
	})
}

func TestArticleCheckBody(t *testing.T) {
	Convey("checkBody", t, func() {
		Convey("missing Body is invalid", func() {
			item := Article{Title: "title"}
			So(errors.Is(item.checkBody(), ErrInvalidBody), ShouldBeTrue)
		})

		Convey("empty Body is invalid", func() {
			item := Article{Title: "title", Body: ""}
			So(errors.Is(item.checkBody(), ErrInvalidBody), ShouldBeTrue)
		})

		Convey("non-empty Body is valid", func() {
			item := Article{Title: "title", Body: "article body"}
			So(item.checkBody(), ShouldBeNil)
		})
	})
}

func TestArticleCheckExcerpt(t *testing.T) {
	Convey("checkExcerpt", t, func() {
		Convey("missing Excerpt is invalid", func() {
			item := Article{Title: "title"}
			So(errors.Is(item.checkExcerpt(), ErrInvalidExcerpt), ShouldBeTrue)
		})

		Convey("empty Excerpt is invalid", func() {
			empty := ""
			item := Article{Title: "title", Excerpt: &empty}
			So(errors.Is(item.checkExcerpt(), ErrInvalidExcerpt), ShouldBeTrue)
		})

		Convey("non-empty Excerpt is valid", func() {
			excerpt := "article excerpt"
			item := Article{Title: "title", Excerpt: &excerpt}
			So(item.checkExcerpt(), ShouldBeNil)
		})
	})
}

func TestArticleReadyToPublish(t *testing.T) {
	Convey("ReadyToPublish", t, func() {
		notEmpty := "not empty"

		Convey("all fields valid", func() {
			item := Article{
				Title:          "title",
				Excerpt:        &notEmpty,
				Body:           notEmpty,
				SeoTitle:       &notEmpty,
				SeoDescription: &notEmpty,
			}
			So(item.ReadyToPublish(), ShouldBeNil)
		})

		Convey("missing body fails", func() {
			item := Article{
				Title:          "title",
				Excerpt:        &notEmpty,
				SeoTitle:       &notEmpty,
				SeoDescription: &notEmpty,
			}
			So(errors.Is(item.ReadyToPublish(), ErrInvalidBody), ShouldBeTrue)
		})

		Convey("missing excerpt fails", func() {
			item := Article{
				Title:          "title",
				Body:           notEmpty,
				SeoTitle:       &notEmpty,
				SeoDescription: &notEmpty,
			}
			So(errors.Is(item.ReadyToPublish(), ErrInvalidExcerpt), ShouldBeTrue)
		})
	})
}

func validCreateArticleRequest() CreateArticleRequest {
	excerpt := "article excerpt"
	return CreateArticleRequest{
		Slug:    "valid-slug",
		Title:   "Title",
		Body:    "article body",
		Excerpt: &excerpt,
	}
}

func TestCreateArticleRequestValidateSlug(t *testing.T) {
	Convey("CreateArticleRequest.validateSlug", t, func() {
		Convey("valid slug", func() {
			req := CreateArticleRequest{Slug: "valid-slug"}
			So(req.validateSlug(), ShouldBeNil)
		})

		Convey("empty slug", func() {
			req := CreateArticleRequest{Slug: ""}
			So(errors.Is(req.validateSlug(), ErrInvalidSlug), ShouldBeTrue)
		})

		Convey("malformed slug", func() {
			req := CreateArticleRequest{Slug: "Not A Slug!"}
			So(errors.Is(req.validateSlug(), ErrInvalidSlug), ShouldBeTrue)
		})
	})
}

func TestCreateArticleRequestValidateTitle(t *testing.T) {
	Convey("CreateArticleRequest.validateTitle", t, func() {
		Convey("valid title", func() {
			req := CreateArticleRequest{Title: "Title"}
			So(req.validateTitle(), ShouldBeNil)
		})

		Convey("empty title", func() {
			req := CreateArticleRequest{Title: ""}
			So(errors.Is(req.validateTitle(), ErrInvalidTitle), ShouldBeTrue)
		})

		Convey("whitespace-only title", func() {
			req := CreateArticleRequest{Title: "   "}
			So(errors.Is(req.validateTitle(), ErrInvalidTitle), ShouldBeTrue)
		})

		Convey("300 runes is valid (boundary)", func() {
			req := CreateArticleRequest{Title: strings.Repeat("A", 300)}
			So(req.validateTitle(), ShouldBeNil)
		})

		Convey("301 runes is invalid (boundary)", func() {
			req := CreateArticleRequest{Title: strings.Repeat("A", 301)}
			So(errors.Is(req.validateTitle(), ErrInvalidTitle), ShouldBeTrue)
		})
	})
}

func TestCreateArticleRequestValidateSeoTitle(t *testing.T) {
	Convey("CreateArticleRequest.validateSeoTitle", t, func() {
		Convey("nil is valid", func() {
			req := CreateArticleRequest{SeoTitle: nil}
			So(req.validateSeoTitle(), ShouldBeNil)
		})

		Convey("empty is invalid", func() {
			empty := ""
			req := CreateArticleRequest{SeoTitle: &empty}
			So(errors.Is(req.validateSeoTitle(), ErrInvalidSeoTitle), ShouldBeTrue)
		})

		Convey("70 runes is valid (boundary)", func() {
			val := strings.Repeat("A", 70)
			req := CreateArticleRequest{SeoTitle: &val}
			So(req.validateSeoTitle(), ShouldBeNil)
		})

		Convey("71 runes is invalid (boundary)", func() {
			val := strings.Repeat("A", 71)
			req := CreateArticleRequest{SeoTitle: &val}
			So(errors.Is(req.validateSeoTitle(), ErrInvalidSeoTitle), ShouldBeTrue)
		})
	})
}

func TestCreateArticleRequestValidateSeoDescription(t *testing.T) {
	Convey("CreateArticleRequest.validateSeoDescription", t, func() {
		Convey("nil is valid", func() {
			req := CreateArticleRequest{SeoDescription: nil}
			So(req.validateSeoDescription(), ShouldBeNil)
		})

		Convey("empty is invalid", func() {
			empty := ""
			req := CreateArticleRequest{SeoDescription: &empty}
			So(errors.Is(req.validateSeoDescription(), ErrInvalidSeoDescription), ShouldBeTrue)
		})

		Convey("160 runes is valid (boundary)", func() {
			val := strings.Repeat("A", 160)
			req := CreateArticleRequest{SeoDescription: &val}
			So(req.validateSeoDescription(), ShouldBeNil)
		})

		Convey("161 runes is invalid (boundary)", func() {
			val := strings.Repeat("A", 161)
			req := CreateArticleRequest{SeoDescription: &val}
			So(errors.Is(req.validateSeoDescription(), ErrInvalidSeoDescription), ShouldBeTrue)
		})
	})
}

func TestCreateArticleRequestValidateOgImageURL(t *testing.T) {
	Convey("CreateArticleRequest.validateOgImageURL", t, func() {
		Convey("nil is valid", func() {
			req := CreateArticleRequest{OgImageURL: nil}
			So(req.validateOgImageURL(), ShouldBeNil)
		})

		Convey("empty is invalid", func() {
			empty := ""
			req := CreateArticleRequest{OgImageURL: &empty}
			So(errors.Is(req.validateOgImageURL(), ErrInvalidOgImageURL), ShouldBeTrue)
		})

		Convey("non-https is invalid", func() {
			val := "http://example.com/og.png"
			req := CreateArticleRequest{OgImageURL: &val}
			So(errors.Is(req.validateOgImageURL(), ErrInvalidOgImageURL), ShouldBeTrue)
		})

		Convey("valid https URL", func() {
			val := "https://example.com/og.png"
			req := CreateArticleRequest{OgImageURL: &val}
			So(req.validateOgImageURL(), ShouldBeNil)
		})
	})
}

func TestCreateArticleRequestValidateBody(t *testing.T) {
	Convey("CreateArticleRequest.validateBody", t, func() {
		Convey("empty is invalid", func() {
			req := CreateArticleRequest{Body: ""}
			So(errors.Is(req.validateBody(), ErrInvalidBody), ShouldBeTrue)
		})

		Convey("10000 runes is valid (boundary)", func() {
			req := CreateArticleRequest{Body: strings.Repeat("A", 10000)}
			So(req.validateBody(), ShouldBeNil)
		})

		Convey("10001 runes is invalid (boundary)", func() {
			req := CreateArticleRequest{Body: strings.Repeat("A", 10001)}
			So(errors.Is(req.validateBody(), ErrInvalidBody), ShouldBeTrue)
		})
	})
}

func TestCreateArticleRequestValidateExcerpt(t *testing.T) {
	Convey("CreateArticleRequest.validateExcerpt", t, func() {
		Convey("nil is valid", func() {
			req := CreateArticleRequest{Excerpt: nil}
			So(req.validateExcerpt(), ShouldBeNil)
		})

		Convey("empty is invalid", func() {
			empty := ""
			req := CreateArticleRequest{Excerpt: &empty}
			So(errors.Is(req.validateExcerpt(), ErrInvalidExcerpt), ShouldBeTrue)
		})

		Convey("non-empty is valid", func() {
			val := "excerpt"
			req := CreateArticleRequest{Excerpt: &val}
			So(req.validateExcerpt(), ShouldBeNil)
		})
	})
}

func TestCreateArticleRequestValidate(t *testing.T) {
	Convey("CreateArticleRequest.Validate", t, func() {
		Convey("all fields valid", func() {
			req := validCreateArticleRequest()
			So(req.Validate(), ShouldBeNil)
		})

		Convey("invalid slug fails", func() {
			req := validCreateArticleRequest()
			req.Slug = ""
			So(errors.Is(req.Validate(), ErrInvalidSlug), ShouldBeTrue)
		})

		Convey("invalid title fails", func() {
			req := validCreateArticleRequest()
			req.Title = ""
			So(errors.Is(req.Validate(), ErrInvalidTitle), ShouldBeTrue)
		})

		Convey("invalid og image URL fails", func() {
			req := validCreateArticleRequest()
			empty := ""
			req.OgImageURL = &empty
			So(errors.Is(req.Validate(), ErrInvalidOgImageURL), ShouldBeTrue)
		})

		Convey("invalid body fails", func() {
			req := validCreateArticleRequest()
			req.Body = ""
			So(errors.Is(req.Validate(), ErrInvalidBody), ShouldBeTrue)
		})

		Convey("invalid excerpt fails", func() {
			req := validCreateArticleRequest()
			empty := ""
			req.Excerpt = &empty
			So(errors.Is(req.Validate(), ErrInvalidExcerpt), ShouldBeTrue)
		})
	})
}

func TestUpdateArticleRequestValidateID(t *testing.T) {
	Convey("UpdateArticleRequest.validateID", t, func() {
		Convey("valid ID", func() {
			req := UpdateArticleRequest{ID: "b3f1c9e2-6a4d-4f2e-9c1a-2d5e7f8a9b0c"}
			So(req.validateID(), ShouldBeNil)
		})

		Convey("empty ID", func() {
			req := UpdateArticleRequest{ID: ""}
			So(errors.Is(req.validateID(), ErrInvalidArticleID), ShouldBeTrue)
		})

		Convey("malformed (non-UUID) ID", func() {
			req := UpdateArticleRequest{ID: "article-id"}
			So(errors.Is(req.validateID(), ErrInvalidArticleID), ShouldBeTrue)
		})
	})
}

func TestUpdateArticleRequestValidateSlug(t *testing.T) {
	Convey("UpdateArticleRequest.validateSlug", t, func() {
		Convey("nil is valid", func() {
			req := UpdateArticleRequest{Slug: nil}
			So(req.validateSlug(), ShouldBeNil)
		})

		Convey("empty is invalid", func() {
			empty := ""
			req := UpdateArticleRequest{Slug: &empty}
			So(errors.Is(req.validateSlug(), ErrInvalidSlug), ShouldBeTrue)
		})

		Convey("malformed is invalid", func() {
			val := "Not A Slug!"
			req := UpdateArticleRequest{Slug: &val}
			So(errors.Is(req.validateSlug(), ErrInvalidSlug), ShouldBeTrue)
		})

		Convey("valid slug", func() {
			val := "valid-slug"
			req := UpdateArticleRequest{Slug: &val}
			So(req.validateSlug(), ShouldBeNil)
		})
	})
}

func TestUpdateArticleRequestValidateTitle(t *testing.T) {
	Convey("UpdateArticleRequest.validateTitle", t, func() {
		Convey("nil is valid", func() {
			req := UpdateArticleRequest{Title: nil}
			So(req.validateTitle(), ShouldBeNil)
		})

		Convey("whitespace-only is invalid", func() {
			val := "   "
			req := UpdateArticleRequest{Title: &val}
			So(errors.Is(req.validateTitle(), ErrInvalidTitle), ShouldBeTrue)
		})

		Convey("valid title", func() {
			val := "Title"
			req := UpdateArticleRequest{Title: &val}
			So(req.validateTitle(), ShouldBeNil)
		})

		Convey("300 runes is valid (boundary)", func() {
			val := strings.Repeat("A", 300)
			req := UpdateArticleRequest{Title: &val}
			So(req.validateTitle(), ShouldBeNil)
		})

		Convey("301 runes is invalid (boundary)", func() {
			val := strings.Repeat("A", 301)
			req := UpdateArticleRequest{Title: &val}
			So(errors.Is(req.validateTitle(), ErrInvalidTitle), ShouldBeTrue)
		})
	})
}

func TestUpdateArticleRequestValidateSeoTitle(t *testing.T) {
	Convey("UpdateArticleRequest.validateSeoTitle", t, func() {
		Convey("nil is valid", func() {
			req := UpdateArticleRequest{SeoTitle: nil}
			So(req.validateSeoTitle(), ShouldBeNil)
		})

		Convey("empty is invalid", func() {
			empty := ""
			req := UpdateArticleRequest{SeoTitle: &empty}
			So(errors.Is(req.validateSeoTitle(), ErrInvalidSeoTitle), ShouldBeTrue)
		})

		Convey("valid seo title", func() {
			val := "not empty"
			req := UpdateArticleRequest{SeoTitle: &val}
			So(req.validateSeoTitle(), ShouldBeNil)
		})

		Convey("70 runes is valid (boundary)", func() {
			val := strings.Repeat("A", 70)
			req := UpdateArticleRequest{SeoTitle: &val}
			So(req.validateSeoTitle(), ShouldBeNil)
		})

		Convey("71 runes is invalid (boundary)", func() {
			val := strings.Repeat("A", 71)
			req := UpdateArticleRequest{SeoTitle: &val}
			So(errors.Is(req.validateSeoTitle(), ErrInvalidSeoTitle), ShouldBeTrue)
		})
	})
}

func TestUpdateArticleRequestValidateSeoDescription(t *testing.T) {
	Convey("UpdateArticleRequest.validateSeoDescription", t, func() {
		Convey("nil is valid", func() {
			req := UpdateArticleRequest{SeoDescription: nil}
			So(req.validateSeoDescription(), ShouldBeNil)
		})

		Convey("empty is invalid", func() {
			empty := ""
			req := UpdateArticleRequest{SeoDescription: &empty}
			So(errors.Is(req.validateSeoDescription(), ErrInvalidSeoDescription), ShouldBeTrue)
		})

		Convey("valid seo description", func() {
			val := "not empty"
			req := UpdateArticleRequest{SeoDescription: &val}
			So(req.validateSeoDescription(), ShouldBeNil)
		})

		Convey("160 runes is valid (boundary)", func() {
			val := strings.Repeat("A", 160)
			req := UpdateArticleRequest{SeoDescription: &val}
			So(req.validateSeoDescription(), ShouldBeNil)
		})

		Convey("161 runes is invalid (boundary)", func() {
			val := strings.Repeat("A", 161)
			req := UpdateArticleRequest{SeoDescription: &val}
			So(errors.Is(req.validateSeoDescription(), ErrInvalidSeoDescription), ShouldBeTrue)
		})
	})
}

func TestUpdateArticleRequestValidateOgImageURL(t *testing.T) {
	Convey("UpdateArticleRequest.validateOgImageURL", t, func() {
		Convey("nil is valid", func() {
			req := UpdateArticleRequest{OgImageURL: nil}
			So(req.validateOgImageURL(), ShouldBeNil)
		})

		Convey("non-https is invalid", func() {
			val := "http://example.com/og.png"
			req := UpdateArticleRequest{OgImageURL: &val}
			So(errors.Is(req.validateOgImageURL(), ErrInvalidOgImageURL), ShouldBeTrue)
		})

		Convey("valid https URL", func() {
			val := "https://example.com/og.png"
			req := UpdateArticleRequest{OgImageURL: &val}
			So(req.validateOgImageURL(), ShouldBeNil)
		})
	})
}

func TestUpdateArticleRequestValidateBody(t *testing.T) {
	Convey("UpdateArticleRequest.validateBody", t, func() {
		Convey("nil is valid", func() {
			req := UpdateArticleRequest{Body: nil}
			So(req.validateBody(), ShouldBeNil)
		})

		Convey("empty is invalid", func() {
			empty := ""
			req := UpdateArticleRequest{Body: &empty}
			So(errors.Is(req.validateBody(), ErrInvalidBody), ShouldBeTrue)
		})

		Convey("valid body", func() {
			val := "article body"
			req := UpdateArticleRequest{Body: &val}
			So(req.validateBody(), ShouldBeNil)
		})

		Convey("10000 runes is valid (boundary)", func() {
			val := strings.Repeat("A", 10000)
			req := UpdateArticleRequest{Body: &val}
			So(req.validateBody(), ShouldBeNil)
		})

		Convey("10001 runes is invalid (boundary)", func() {
			val := strings.Repeat("A", 10001)
			req := UpdateArticleRequest{Body: &val}
			So(errors.Is(req.validateBody(), ErrInvalidBody), ShouldBeTrue)
		})
	})
}

func TestUpdateArticleRequestValidateExcerpt(t *testing.T) {
	Convey("UpdateArticleRequest.validateExcerpt", t, func() {
		Convey("nil is valid", func() {
			req := UpdateArticleRequest{Excerpt: nil}
			So(req.validateExcerpt(), ShouldBeNil)
		})

		Convey("empty is invalid", func() {
			empty := ""
			req := UpdateArticleRequest{Excerpt: &empty}
			So(errors.Is(req.validateExcerpt(), ErrInvalidExcerpt), ShouldBeTrue)
		})

		Convey("valid excerpt", func() {
			val := "excerpt"
			req := UpdateArticleRequest{Excerpt: &val}
			So(req.validateExcerpt(), ShouldBeNil)
		})
	})
}

func TestUpdateArticleRequestValidate(t *testing.T) {
	Convey("UpdateArticleRequest.Validate", t, func() {
		validID := "b3f1c9e2-6a4d-4f2e-9c1a-2d5e7f8a9b0c"

		Convey("all nil fields except ID is valid", func() {
			req := UpdateArticleRequest{ID: validID}
			So(req.Validate(), ShouldBeNil)
		})

		Convey("missing ID fails", func() {
			req := UpdateArticleRequest{ID: ""}
			So(errors.Is(req.Validate(), ErrInvalidArticleID), ShouldBeTrue)
		})

		Convey("malformed (non-UUID) ID fails", func() {
			req := UpdateArticleRequest{ID: "article-id"}
			So(errors.Is(req.Validate(), ErrInvalidArticleID), ShouldBeTrue)
		})

		Convey("invalid slug fails", func() {
			empty := ""
			req := UpdateArticleRequest{ID: validID, Slug: &empty}
			So(errors.Is(req.Validate(), ErrInvalidSlug), ShouldBeTrue)
		})

		Convey("invalid title fails", func() {
			empty := "  "
			req := UpdateArticleRequest{ID: validID, Title: &empty}
			So(errors.Is(req.Validate(), ErrInvalidTitle), ShouldBeTrue)
		})

		Convey("invalid seo title fails", func() {
			empty := ""
			req := UpdateArticleRequest{ID: validID, SeoTitle: &empty}
			So(errors.Is(req.Validate(), ErrInvalidSeoTitle), ShouldBeTrue)
		})

		Convey("invalid seo description fails", func() {
			empty := ""
			req := UpdateArticleRequest{ID: validID, SeoDescription: &empty}
			So(errors.Is(req.Validate(), ErrInvalidSeoDescription), ShouldBeTrue)
		})

		Convey("invalid og image URL fails", func() {
			empty := ""
			req := UpdateArticleRequest{ID: validID, OgImageURL: &empty}
			So(errors.Is(req.Validate(), ErrInvalidOgImageURL), ShouldBeTrue)
		})

		Convey("invalid body fails", func() {
			empty := ""
			req := UpdateArticleRequest{ID: validID, Body: &empty}
			So(errors.Is(req.Validate(), ErrInvalidBody), ShouldBeTrue)
		})

		Convey("invalid excerpt fails", func() {
			empty := ""
			req := UpdateArticleRequest{ID: validID, Excerpt: &empty}
			So(errors.Is(req.Validate(), ErrInvalidExcerpt), ShouldBeTrue)
		})
	})
}

func TestUpdateArticleRequestApplyNilFields(t *testing.T) {
	Convey("UpdateArticleRequest.Apply: nil fields leave Article unchanged", t, func() {
		item := Article{
			ID:          "article-id",
			Slug:        "old-slug",
			Title:       "Old Title",
			IsIndexable: true,
		}
		req := UpdateArticleRequest{ID: "article-id"}
		req.Apply(&item)

		So(item.Slug, ShouldEqual, "old-slug")
		So(item.Title, ShouldEqual, "Old Title")
		So(item.IsIndexable, ShouldBeTrue)
	})
}

func TestUpdateArticleRequestApplyNonNilFields(t *testing.T) {
	Convey("UpdateArticleRequest.Apply: non-nil fields overwrite Article fields", t, func() {
		item := Article{
			ID:             "article-id",
			Slug:           "old-slug",
			Title:          "Old Title",
			Excerpt:        nil,
			Body:           "",
			SeoTitle:       nil,
			SeoDescription: nil,
			OgImageURL:     nil,
			IsIndexable:    false,
		}

		newSlug := "new-slug"
		newTitle := "New Title"
		newExcerpt := "New Excerpt"
		newBody := "New body"
		newSeoTitle := "New SEO Title"
		newSeoDescription := "New SEO Description"
		newOgImageURL := "https://example.com/og.png"
		newIsIndexable := true

		req := UpdateArticleRequest{
			ID:             "article-id",
			Slug:           &newSlug,
			Title:          &newTitle,
			Excerpt:        &newExcerpt,
			Body:           &newBody,
			SeoTitle:       &newSeoTitle,
			SeoDescription: &newSeoDescription,
			OgImageURL:     &newOgImageURL,
			IsIndexable:    &newIsIndexable,
		}
		req.Apply(&item)

		So(item.Slug, ShouldEqual, newSlug)
		So(item.Title, ShouldEqual, newTitle)
		So(item.Excerpt, ShouldEqual, &newExcerpt)
		So(item.Body, ShouldEqual, newBody)
		So(item.SeoTitle, ShouldEqual, &newSeoTitle)
		So(item.SeoDescription, ShouldEqual, &newSeoDescription)
		So(item.OgImageURL, ShouldEqual, &newOgImageURL)
		So(item.IsIndexable, ShouldBeTrue)
	})
}
