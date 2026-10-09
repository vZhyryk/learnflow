package notesdomain

import (
	"errors"
	"strings"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

const (
	validUserID     = "11111111-1111-1111-1111-111111111111"
	validNoteID     = "22222222-2222-2222-2222-222222222222"
	validResourceID = "33333333-3333-3333-3333-333333333333"
)

func ptr[T any](v T) *T { return &v }

func validCreateRequest() CreateNotesRequest {
	return CreateNotesRequest{UserID: validUserID, Title: "title", Body: "body"}
}

func TestCreateNotesRequestValidate(t *testing.T) {
	Convey("CreateNotesRequest.Validate", t, func() {
		Convey("A personal note with only the required fields is valid", func() {
			req := validCreateRequest()
			So(req.Validate(), ShouldBeNil)
		})

		Convey("A note linked to a course or a content item is valid", func() {
			for _, resourceType := range []ResourceType{CourseResourceType, ContentResourceType} {
				req := validCreateRequest()
				req.ResourceType, req.ResourceID = ptr(resourceType), ptr(validResourceID)
				So(req.Validate(), ShouldBeNil)
			}
		})
	})
}

func TestCreateNotesRequestInvalid(t *testing.T) {
	linked := func(resourceType ResourceType, resourceID string) CreateNotesRequest {
		return CreateNotesRequest{UserID: validUserID, Title: "title", Body: "body", ResourceType: &resourceType, ResourceID: &resourceID}
	}

	cases := []struct {
		name string
		req  CreateNotesRequest
		want error
	}{
		{"empty title", CreateNotesRequest{UserID: validUserID, Title: "", Body: "body"}, ErrInvalidTitle},
		{"blank title", CreateNotesRequest{UserID: validUserID, Title: "   ", Body: "body"}, ErrInvalidTitle},
		{"title over 300 runes", CreateNotesRequest{UserID: validUserID, Title: strings.Repeat("a", 301), Body: "body"}, ErrInvalidTitle},
		{"empty body", CreateNotesRequest{UserID: validUserID, Title: "title", Body: ""}, ErrInvalidBody},
		{"body over 10000 runes", CreateNotesRequest{UserID: validUserID, Title: "title", Body: strings.Repeat("a", 10001)}, ErrInvalidBody},
		{"empty description", CreateNotesRequest{UserID: validUserID, Title: "title", Body: "body", Description: ptr("")}, ErrInvalidDescription},
		{"description over 10000 runes", CreateNotesRequest{UserID: validUserID, Title: "title", Body: "body", Description: ptr(strings.Repeat("a", 10001))}, ErrInvalidDescription},
		{"unknown resource type", linked("video", validResourceID), ErrInvalidResourceType},
		{"resource id not a UUID", linked(CourseResourceType, "nope"), ErrInvalidResourceID},
		{"resource type without id", CreateNotesRequest{UserID: validUserID, Title: "title", Body: "body", ResourceType: ptr(CourseResourceType)}, ErrResourceDataMisMatch},
		{"resource id without type", CreateNotesRequest{UserID: validUserID, Title: "title", Body: "body", ResourceID: ptr(validResourceID)}, ErrResourceDataMisMatch},
		{"missing user id", CreateNotesRequest{Title: "title", Body: "body"}, ErrInvalidUserID},
		{"user id not a UUID", CreateNotesRequest{UserID: "nope", Title: "title", Body: "body"}, ErrInvalidUserID},
	}

	for _, tc := range cases {
		Convey("CreateNotesRequest.Validate rejects a request with an "+tc.name, t, func() {
			So(errors.Is(tc.req.Validate(), tc.want), ShouldBeTrue)
		})
	}
}

func TestUpdateNotesRequestValidate(t *testing.T) {
	Convey("UpdateNotesRequest.Validate", t, func() {
		valid := func() UpdateNotesRequest { return UpdateNotesRequest{ID: validNoteID, UserID: validUserID} }

		Convey("A request that changes nothing is valid", func() {
			req := valid()
			So(req.Validate(), ShouldBeNil)
		})

		Convey("A request with valid fields is valid", func() {
			req := valid()
			req.Title, req.Description, req.Body = ptr("new"), ptr("desc"), ptr("body")
			So(req.Validate(), ShouldBeNil)
		})

		cases := []struct {
			name   string
			mutate func(*UpdateNotesRequest)
			want   error
		}{
			{"blank title", func(r *UpdateNotesRequest) { r.Title = ptr(" ") }, ErrInvalidTitle},
			{"empty body", func(r *UpdateNotesRequest) { r.Body = ptr("") }, ErrInvalidBody},
			{"body over 10000 runes", func(r *UpdateNotesRequest) { r.Body = ptr(strings.Repeat("a", 10001)) }, ErrInvalidBody},
			{"empty description", func(r *UpdateNotesRequest) { r.Description = ptr("") }, ErrInvalidDescription},
			{"missing user id", func(r *UpdateNotesRequest) { r.UserID = "" }, ErrInvalidUserID},
			{"note id not a UUID", func(r *UpdateNotesRequest) { r.ID = "nope" }, ErrNoteNotFound},
		}
		for _, tc := range cases {
			Convey("When the request has an "+tc.name+", it is rejected", func() {
				req := valid()
				tc.mutate(&req)

				So(errors.Is(req.Validate(), tc.want), ShouldBeTrue)
			})
		}
	})
}

func TestUpdateNotesRequestApply(t *testing.T) {
	Convey("UpdateNotesRequest.Apply", t, func() {
		oldDescription := "old description"
		note := &UserNotes{Title: "old", Description: &oldDescription, Body: "old body"}

		Convey("When no field is provided, the note is unchanged", func() {
			UpdateNotesRequest{}.Apply(note)

			So(note.Title, ShouldEqual, "old")
			So(*note.Description, ShouldEqual, "old description")
			So(note.Body, ShouldEqual, "old body")
		})

		Convey("When fields are provided, only those are copied", func() {
			UpdateNotesRequest{Title: ptr("new"), Body: ptr("new body")}.Apply(note)

			So(note.Title, ShouldEqual, "new")
			So(note.Body, ShouldEqual, "new body")
			So(*note.Description, ShouldEqual, "old description")
		})
	})
}
