package notesdomain

import (
	"learnflow_backend/internal/shared/validator"
	"time"
)

// ResourceType names the kind of resource a note is linked to.
type ResourceType string

// Resource types a note can be linked to.
const (
	ContentResourceType ResourceType = "content_item"
	CourseResourceType  ResourceType = "course"
)

// UserNotes is a note owned by a user, optionally linked to a course or content item.
type UserNotes struct {
	UserID       string        `json:"user_id"`
	ID           string        `json:"id"`
	ResourceType *ResourceType `json:"resource_type"`
	ResourceID   *string       `json:"resource_id"`
	Title        string        `json:"title"`
	Description  *string       `json:"description"`
	Body         string        `json:"body"`
	DeletedAt    *time.Time    `json:"deleted_at"`
	CreatedAt    time.Time     `json:"created_at"`
	UpdatedAt    time.Time     `json:"updated_at"`
}

// CreateNotesRequest is the input for creating a note; UserID comes from the session, never the body.
type CreateNotesRequest struct {
	UserID       string        `json:"-"`
	ResourceType *ResourceType `json:"resource_type"`
	ResourceID   *string       `json:"resource_id"`
	Title        string        `json:"title"`
	Description  *string       `json:"description"`
	Body         string        `json:"body"`
}

// Validate checks the request fields.
func (req *CreateNotesRequest) Validate() error {
	checks := []func() error{
		req.validateTitle,
		req.validateDescription,
		req.validateBody,
		req.validateResourceType,
		req.validateResourceID,
		req.validateUserID,
		req.validateConstraints,
	}
	for _, check := range checks {
		if err := check(); err != nil {
			return err
		}
	}

	return nil
}

func (req *CreateNotesRequest) validateTitle() error {
	if !validator.IsValidContentTitle(req.Title) {
		return ErrInvalidTitle
	}
	return nil
}

func (req *CreateNotesRequest) validateBody() error {
	if req.Body == "" || !validator.IsValidContentBody(req.Body) {
		return ErrInvalidBody
	}
	return nil
}

func (req *CreateNotesRequest) validateDescription() error {
	if req.Description != nil && (*req.Description == "" || !validator.IsValidContentDescription(*req.Description)) {
		return ErrInvalidDescription
	}
	return nil
}

func (req *CreateNotesRequest) validateResourceType() error {
	if req.ResourceType != nil && *req.ResourceType != ContentResourceType && *req.ResourceType != CourseResourceType {
		return ErrInvalidResourceType
	}
	return nil
}

func (req *CreateNotesRequest) validateResourceID() error {
	if req.ResourceID != nil && !validator.IsValidUUID(*req.ResourceID) {
		return ErrInvalidResourceID
	}
	return nil
}

func (req *CreateNotesRequest) validateUserID() error {
	if req.UserID == "" || !validator.IsValidUUID(req.UserID) {
		return ErrInvalidUserID
	}
	return nil
}

func (req *CreateNotesRequest) validateConstraints() error {
	if req.ResourceType != nil && req.ResourceID == nil {
		return ErrResourceDataMisMatch
	}

	if req.ResourceType == nil && req.ResourceID != nil {
		return ErrResourceDataMisMatch
	}
	return nil
}

// UpdateNotesRequest is the input for updating a note; ID comes from the path and UserID from the session.
// The linked resource cannot be changed, so the body does not accept it.
type UpdateNotesRequest struct {
	ID          string  `json:"-"`
	UserID      string  `json:"-"`
	Title       *string `json:"title"`
	Description *string `json:"description"`
	Body        *string `json:"body"`
}

// Validate checks the request fields that were provided.
func (req *UpdateNotesRequest) Validate() error {
	checks := []func() error{
		req.validateTitle,
		req.validateDescription,
		req.validateBody,
		req.validateUserID,
		req.validateID,
	}
	for _, check := range checks {
		if err := check(); err != nil {
			return err
		}
	}

	return nil
}

func (req *UpdateNotesRequest) validateTitle() error {
	if req.Title != nil && !validator.IsValidContentTitle(*req.Title) {
		return ErrInvalidTitle
	}
	return nil
}

func (req *UpdateNotesRequest) validateBody() error {
	if req.Body != nil && (*req.Body == "" || !validator.IsValidContentBody(*req.Body)) {
		return ErrInvalidBody
	}
	return nil
}

func (req *UpdateNotesRequest) validateDescription() error {
	if req.Description != nil && (*req.Description == "" || !validator.IsValidContentDescription(*req.Description)) {
		return ErrInvalidDescription
	}
	return nil
}

func (req *UpdateNotesRequest) validateUserID() error {
	if req.UserID == "" || !validator.IsValidUUID(req.UserID) {
		return ErrInvalidUserID
	}
	return nil
}

func (req *UpdateNotesRequest) validateID() error {
	if req.ID == "" || !validator.IsValidUUID(req.ID) {
		return ErrNoteNotFound
	}
	return nil
}

// Apply copies the provided fields onto p.
func (r UpdateNotesRequest) Apply(p *UserNotes) {
	appliers := []func(*UserNotes){
		r.applyTitle,
		r.applyDescription,
		r.applyBody,
	}
	for _, apply := range appliers {
		apply(p)
	}
}
func (r UpdateNotesRequest) applyTitle(p *UserNotes) {
	if r.Title != nil {
		p.Title = *r.Title
	}
}

func (r UpdateNotesRequest) applyDescription(p *UserNotes) {
	if r.Description != nil {
		p.Description = r.Description
	}
}

func (r UpdateNotesRequest) applyBody(p *UserNotes) {
	if r.Body != nil {
		p.Body = *r.Body
	}
}
