package admindomain

import (
	"errors"
	"strings"
	"testing"
	"time"

	. "github.com/smartystreets/goconvey/convey"
)

func TestEntityTypeIsValid(t *testing.T) {
	Convey("EntityType.IsValid", t, func() {
		Convey("known types are valid", func() {
			So(ArticleEntityType.IsValid(), ShouldBeTrue)
			So(ContentEntityType.IsValid(), ShouldBeTrue)
			So(CourseEntityType.IsValid(), ShouldBeTrue)
		})

		Convey("unknown type is invalid", func() {
			So(EntityType("invalid").IsValid(), ShouldBeFalse)
		})
	})
}

func TestChannelIsValid(t *testing.T) {
	Convey("Channel.IsValid", t, func() {
		Convey("known channels are valid", func() {
			So(EmailChannel.IsValid(), ShouldBeTrue)
			So(BannerChannel.IsValid(), ShouldBeTrue)
			So(InAppChannel.IsValid(), ShouldBeTrue)
		})

		Convey("unknown channel is invalid", func() {
			So(Channel("sms").IsValid(), ShouldBeFalse)
		})
	})
}

// validCreateAnnouncementRequest returns a request that passes every check — tests
// mutate a copy to isolate exactly one field per Convey scenario.
func validCreateAnnouncementRequest() CreateAnnouncementRequest {
	return CreateAnnouncementRequest{
		Title:           "Valid Announcement Title",
		Body:            "Valid announcement body content.",
		EntityID:        nil,
		EntityType:      nil,
		Channels:        []Channel{EmailChannel, BannerChannel},
		ExpiresAt:       time.Now().Add(24 * time.Hour),
		CreatedByUserID: "user-1",
	}
}

func TestCreateAnnouncementRequestValidateSuccess(t *testing.T) {
	Convey("CreateAnnouncementRequest.Validate", t, func() {
		Convey("valid platform-wide request (no entity) succeeds", func() {
			req := validCreateAnnouncementRequest()
			So(req.Validate(), ShouldBeNil)
		})

		Convey("valid entity-scoped request succeeds", func() {
			req := validCreateAnnouncementRequest()
			entityID := "course-123"
			entityType := CourseEntityType
			req.EntityID = &entityID
			req.EntityType = &entityType
			So(req.Validate(), ShouldBeNil)
		})
	})
}

func TestCreateAnnouncementRequestValidateTitle(t *testing.T) {
	Convey("CreateAnnouncementRequest title validation", t, func() {
		Convey("empty title fails", func() {
			req := validCreateAnnouncementRequest()
			req.Title = ""
			So(errors.Is(req.Validate(), ErrInvalidTitle), ShouldBeTrue)
		})

		Convey("title over 300 runes fails", func() {
			req := validCreateAnnouncementRequest()
			req.Title = strings.Repeat("a", 301)
			So(errors.Is(req.Validate(), ErrInvalidTitle), ShouldBeTrue)
		})
	})
}

func TestCreateAnnouncementRequestValidateBody(t *testing.T) {
	Convey("CreateAnnouncementRequest body validation", t, func() {
		Convey("empty body fails", func() {
			req := validCreateAnnouncementRequest()
			req.Body = ""
			So(errors.Is(req.Validate(), ErrInvalidBody), ShouldBeTrue)
		})

		Convey("body over 10000 runes fails", func() {
			req := validCreateAnnouncementRequest()
			req.Body = strings.Repeat("a", 10001)
			So(errors.Is(req.Validate(), ErrInvalidBody), ShouldBeTrue)
		})
	})
}

func TestCreateAnnouncementRequestValidateEntity(t *testing.T) {
	Convey("CreateAnnouncementRequest entity field validation", t, func() {
		Convey("empty-string entity ID (non-nil pointer) fails", func() {
			req := validCreateAnnouncementRequest()
			empty := ""
			req.EntityID = &empty
			entityType := CourseEntityType
			req.EntityType = &entityType
			So(errors.Is(req.Validate(), ErrInvalidEntityID), ShouldBeTrue)
		})

		Convey("nil entity ID does not panic and passes that check", func() {
			req := validCreateAnnouncementRequest()
			req.EntityID = nil
			So(req.Validate(), ShouldBeNil)
		})

		Convey("nil entity type does not panic and passes that check", func() {
			req := validCreateAnnouncementRequest()
			req.EntityType = nil
			So(req.Validate(), ShouldBeNil)
		})

		Convey("invalid entity type fails", func() {
			req := validCreateAnnouncementRequest()
			entityID := "course-123"
			invalid := EntityType("invalid")
			req.EntityID = &entityID
			req.EntityType = &invalid
			So(errors.Is(req.Validate(), ErrInvalidEntityType), ShouldBeTrue)
		})
	})
}

func TestCreateAnnouncementRequestValidateChannelsAndExpiry(t *testing.T) {
	Convey("CreateAnnouncementRequest channels/expiry validation", t, func() {
		Convey("invalid channel fails", func() {
			req := validCreateAnnouncementRequest()
			req.Channels = []Channel{"sms"}
			So(errors.Is(req.Validate(), ErrInvalidChannel), ShouldBeTrue)
		})

		Convey("expiresAt in the past fails", func() {
			req := validCreateAnnouncementRequest()
			req.ExpiresAt = time.Now().Add(-1 * time.Hour)
			So(errors.Is(req.Validate(), ErrInvalidExpiresAt), ShouldBeTrue)
		})
	})
}

func TestCreateAnnouncementRequestValidateConstraints(t *testing.T) {
	Convey("CreateAnnouncementRequest entity pairing constraint", t, func() {
		Convey("entityType set without entityID fails", func() {
			req := validCreateAnnouncementRequest()
			entityType := CourseEntityType
			req.EntityType = &entityType
			req.EntityID = nil
			So(errors.Is(req.Validate(), ErrEntityDataMisMatch), ShouldBeTrue)
		})

		Convey("entityID set without entityType fails", func() {
			req := validCreateAnnouncementRequest()
			entityID := "course-123"
			req.EntityID = &entityID
			req.EntityType = nil
			So(errors.Is(req.Validate(), ErrEntityDataMisMatch), ShouldBeTrue)
		})
	})
}

func validUpdateAnnouncementRequest() UpdateAnnouncementRequest {
	return UpdateAnnouncementRequest{
		ID:              "announcement-123",
		UpdatedByUserID: "user-1",
	}
}

func TestUpdateAnnouncementRequestValidateID(t *testing.T) {
	Convey("UpdateAnnouncementRequest ID validation", t, func() {
		Convey("all-nil optional fields succeed (no-op patch)", func() {
			req := validUpdateAnnouncementRequest()
			So(req.Validate(), ShouldBeNil)
		})

		Convey("empty ID fails", func() {
			req := validUpdateAnnouncementRequest()
			req.ID = ""
			So(errors.Is(req.Validate(), ErrInvalidID), ShouldBeTrue)
		})
	})
}

func TestUpdateAnnouncementRequestValidateTitleAndBody(t *testing.T) {
	Convey("UpdateAnnouncementRequest title/body validation", t, func() {
		Convey("nil title does not panic and passes", func() {
			req := validUpdateAnnouncementRequest()
			req.Title = nil
			So(req.Validate(), ShouldBeNil)
		})

		Convey("empty-string title (non-nil pointer) fails", func() {
			req := validUpdateAnnouncementRequest()
			empty := ""
			req.Title = &empty
			So(errors.Is(req.Validate(), ErrInvalidTitle), ShouldBeTrue)
		})

		Convey("title over 300 runes fails", func() {
			req := validUpdateAnnouncementRequest()
			tooLong := strings.Repeat("a", 301)
			req.Title = &tooLong
			So(errors.Is(req.Validate(), ErrInvalidTitle), ShouldBeTrue)
		})

		Convey("nil body does not panic and passes", func() {
			req := validUpdateAnnouncementRequest()
			req.Body = nil
			So(req.Validate(), ShouldBeNil)
		})

		Convey("empty-string body (non-nil pointer) fails", func() {
			req := validUpdateAnnouncementRequest()
			empty := ""
			req.Body = &empty
			So(errors.Is(req.Validate(), ErrInvalidBody), ShouldBeTrue)
		})
	})
}

func TestUpdateAnnouncementRequestValidateEntity(t *testing.T) {
	Convey("UpdateAnnouncementRequest entity field validation", t, func() {
		Convey("nil entity ID does not panic and passes", func() {
			req := validUpdateAnnouncementRequest()
			req.EntityID = nil
			So(req.Validate(), ShouldBeNil)
		})

		Convey("empty-string entity ID (non-nil pointer) fails", func() {
			req := validUpdateAnnouncementRequest()
			empty := ""
			req.EntityID = &empty
			entityType := CourseEntityType
			req.EntityType = &entityType
			So(errors.Is(req.Validate(), ErrInvalidEntityID), ShouldBeTrue)
		})

		Convey("nil entity type does not panic and passes", func() {
			req := validUpdateAnnouncementRequest()
			req.EntityType = nil
			So(req.Validate(), ShouldBeNil)
		})

		Convey("invalid entity type fails", func() {
			req := validUpdateAnnouncementRequest()
			entityID := "course-123"
			invalid := EntityType("invalid")
			req.EntityID = &entityID
			req.EntityType = &invalid
			So(errors.Is(req.Validate(), ErrInvalidEntityType), ShouldBeTrue)
		})
	})
}

func TestUpdateAnnouncementRequestValidateChannelsAndExpiry(t *testing.T) {
	Convey("UpdateAnnouncementRequest channels/expiry validation", t, func() {
		Convey("nil channels does not panic and passes", func() {
			req := validUpdateAnnouncementRequest()
			req.Channels = nil
			So(req.Validate(), ShouldBeNil)
		})

		Convey("invalid channel in non-nil slice fails", func() {
			req := validUpdateAnnouncementRequest()
			channels := []Channel{"sms"}
			req.Channels = &channels
			So(errors.Is(req.Validate(), ErrInvalidChannel), ShouldBeTrue)
		})

		Convey("nil expiresAt does not panic and passes", func() {
			req := validUpdateAnnouncementRequest()
			req.ExpiresAt = nil
			So(req.Validate(), ShouldBeNil)
		})

		Convey("expiresAt in the past fails", func() {
			req := validUpdateAnnouncementRequest()
			past := time.Now().Add(-1 * time.Hour)
			req.ExpiresAt = &past
			So(errors.Is(req.Validate(), ErrInvalidExpiresAt), ShouldBeTrue)
		})
	})
}

func TestUpdateAnnouncementRequestValidateConstraints(t *testing.T) {
	Convey("UpdateAnnouncementRequest entity pairing constraint", t, func() {
		Convey("entityType set without entityID fails", func() {
			req := validUpdateAnnouncementRequest()
			entityType := CourseEntityType
			req.EntityType = &entityType
			So(errors.Is(req.Validate(), ErrEntityDataMisMatch), ShouldBeTrue)
		})

		Convey("entityID set without entityType fails", func() {
			req := validUpdateAnnouncementRequest()
			entityID := "course-123"
			req.EntityID = &entityID
			So(errors.Is(req.Validate(), ErrEntityDataMisMatch), ShouldBeTrue)
		})
	})
}

func TestUpdateAnnouncementRequestApply(t *testing.T) {
	Convey("UpdateAnnouncementRequest.Apply", t, func() {
		Convey("nil fields leave the existing announcement untouched", func() {
			expiresAt := time.Now().Add(48 * time.Hour)
			original := &Announcement{
				Title:     "Original Title",
				Body:      "Original body",
				ExpiresAt: expiresAt,
			}
			req := UpdateAnnouncementRequest{ID: "announcement-123", UpdatedByUserID: "user-1"}

			req.Apply(original)

			So(original.Title, ShouldEqual, "Original Title")
			So(original.Body, ShouldEqual, "Original body")
			So(original.ExpiresAt, ShouldEqual, expiresAt)
			So(original.EntityID, ShouldBeNil)
			So(*original.UpdatedByUserID, ShouldEqual, "user-1")
		})

		Convey("non-nil fields overwrite the existing announcement", func() {
			original := &Announcement{
				Title: "Original Title",
				Body:  "Original body",
			}
			newTitle := "New Title"
			newBody := "New body"
			newEntityID := "course-123"
			newEntityType := CourseEntityType
			newChannels := []Channel{EmailChannel}
			newExpiresAt := time.Now().Add(72 * time.Hour)
			req := UpdateAnnouncementRequest{
				ID:              "announcement-123",
				Title:           &newTitle,
				Body:            &newBody,
				EntityID:        &newEntityID,
				EntityType:      &newEntityType,
				Channels:        &newChannels,
				ExpiresAt:       &newExpiresAt,
				UpdatedByUserID: "user-1",
			}

			req.Apply(original)

			So(original.Title, ShouldEqual, "New Title")
			So(original.Body, ShouldEqual, "New body")
			So(*original.EntityID, ShouldEqual, "course-123")
			So(*original.EntityType, ShouldEqual, CourseEntityType)
			So(original.Channels, ShouldResemble, newChannels)
			So(original.ExpiresAt, ShouldEqual, newExpiresAt)
			So(*original.UpdatedByUserID, ShouldEqual, "user-1")
		})

		Convey("empty UpdatedByUserID does not overwrite", func() {
			original := &Announcement{Title: "Original Title"}
			req := UpdateAnnouncementRequest{ID: "announcement-123", UpdatedByUserID: ""}

			req.Apply(original)

			So(original.UpdatedByUserID, ShouldBeNil)
		})
	})
}
