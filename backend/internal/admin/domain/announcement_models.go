package admindomain

import (
	"learnflow_backend/internal/shared/validator"
	"time"
)

// EntityType identifies the kind of entity an announcement is scoped to.
type EntityType string

// Supported announcement entity types.
const (
	ArticleEntityType EntityType = "article"
	ContentEntityType EntityType = "content_item"
	CourseEntityType  EntityType = "course"
)

// IsValid reports whether et is one of the known entity types.
func (et EntityType) IsValid() bool {
	if et != ArticleEntityType && et != ContentEntityType && et != CourseEntityType {
		return false
	}

	return true
}

// Channel identifies a delivery channel for an announcement.
type Channel string

// Supported announcement delivery channels.
const (
	EmailChannel  Channel = "email"
	BannerChannel Channel = "banner"
	InAppChannel  Channel = "inapp"
)

// IsValid reports whether ch is one of the known channels.
func (ch Channel) IsValid() bool {
	if ch != EmailChannel && ch != BannerChannel && ch != InAppChannel {
		return false
	}

	return true
}

// Announcement is a platform-wide or entity-scoped admin announcement.
type Announcement struct {
	ID    string `json:"id"`
	Title string `json:"title"`
	// Body is plaintext only — never render via v-html.
	Body             string      `json:"body"`
	CreatedAt        time.Time   `json:"created_at"`
	CreatedByUserID  string      `json:"created_by_user_id"`
	UpdatedAt        *time.Time  `json:"updated_at"`
	UpdatedByUserID  *string     `json:"updated_by_user_id"`
	ApprovedAt       *time.Time  `json:"approved_at"`
	ApprovedByUserID *string     `json:"approved_by_user_id"`
	ExpiresAt        time.Time   `json:"expires_at"`
	EntityID         *string     `json:"entity_id"`
	EntityType       *EntityType `json:"entity_type"`
	Channels         []Channel   `json:"channels"`
}

// AnnouncementPublic is the user-facing subset of Announcement, without audit fields.
type AnnouncementPublic struct {
	ID    string `json:"id"`
	Title string `json:"title"`
	// Body is plaintext only — never render via v-html.
	Body       string      `json:"body"`
	ApprovedAt time.Time   `json:"approved_at"`
	ExpiresAt  time.Time   `json:"expires_at"`
	EntityID   *string     `json:"entity_id"`
	EntityType *EntityType `json:"entity_type"`
}

// CreateAnnouncementRequest is the input for creating a new Announcement.
type CreateAnnouncementRequest struct {
	Title           string      `json:"title"`
	Body            string      `json:"body"`
	EntityID        *string     `json:"entity_id"`
	EntityType      *EntityType `json:"entity_type"`
	Channels        []Channel   `json:"channels"`
	ExpiresAt       time.Time   `json:"expires_at"`
	CreatedByUserID string      `json:"created_by_user_id"`
}

// Validate checks all fields of the request.
func (req *CreateAnnouncementRequest) Validate() error {
	checks := []func() error{
		req.validateTitle,
		req.validateEntityID,
		req.validateChannels,
		req.validateBody,
		req.validateEntityType,
		req.validateExpiresAt,
		req.validateConstraints,
	}
	for _, check := range checks {
		if err := check(); err != nil {
			return err
		}
	}

	return nil
}

func (req *CreateAnnouncementRequest) validateTitle() error {
	if !validator.IsValidContentTitle(req.Title) {
		return ErrInvalidTitle
	}
	return nil
}

func (req *CreateAnnouncementRequest) validateBody() error {
	if req.Body == "" || !validator.IsValidContentBody(req.Body) {
		return ErrInvalidBody
	}
	return nil
}

func (req *CreateAnnouncementRequest) validateEntityID() error {
	if req.EntityID != nil && *req.EntityID == "" {
		return ErrInvalidEntityID
	}
	return nil
}

func (req *CreateAnnouncementRequest) validateChannels() error {
	for _, ch := range req.Channels {
		if !ch.IsValid() {
			return ErrInvalidChannel
		}
	}
	return nil
}

func (req *CreateAnnouncementRequest) validateEntityType() error {
	if req.EntityType != nil && !req.EntityType.IsValid() {
		return ErrInvalidEntityType
	}
	return nil
}

func (req *CreateAnnouncementRequest) validateExpiresAt() error {
	if req.ExpiresAt.Before(time.Now()) {
		return ErrInvalidExpiresAt
	}
	return nil
}

func (req *CreateAnnouncementRequest) validateConstraints() error {
	if req.EntityType != nil && req.EntityID == nil {
		return ErrEntityDataMisMatch
	}

	if req.EntityType == nil && req.EntityID != nil {
		return ErrEntityDataMisMatch
	}
	return nil
}

// UpdateAnnouncementRequest is the input for updating an existing Announcement.
type UpdateAnnouncementRequest struct {
	ID              string      `json:"id"`
	Title           *string     `json:"title"`
	Body            *string     `json:"body"`
	EntityID        *string     `json:"entity_id"`
	EntityType      *EntityType `json:"entity_type"`
	Channels        *[]Channel  `json:"channels"`
	ExpiresAt       *time.Time  `json:"expires_at"`
	UpdatedByUserID string      `json:"updated_by_user_id"`
}

// Validate checks all fields of the request.
func (req *UpdateAnnouncementRequest) Validate() error {
	checks := []func() error{
		req.validateID,
		req.validateTitle,
		req.validateEntityID,
		req.validateChannels,
		req.validateBody,
		req.validateEntityType,
		req.validateExpiresAt,
		req.validateConstraints,
	}
	for _, check := range checks {
		if err := check(); err != nil {
			return err
		}
	}

	return nil
}

func (req *UpdateAnnouncementRequest) validateID() error {
	if req.ID == "" {
		return ErrInvalidID
	}
	return nil
}

func (req *UpdateAnnouncementRequest) validateTitle() error {
	if req.Title != nil && !validator.IsValidContentTitle(*req.Title) {
		return ErrInvalidTitle
	}
	return nil
}

func (req *UpdateAnnouncementRequest) validateBody() error {
	if req.Body != nil && (*req.Body == "" || !validator.IsValidContentBody(*req.Body)) {
		return ErrInvalidBody
	}
	return nil
}

func (req *UpdateAnnouncementRequest) validateEntityID() error {
	if req.EntityID != nil && *req.EntityID == "" {
		return ErrInvalidEntityID
	}
	return nil
}

func (req *UpdateAnnouncementRequest) validateChannels() error {
	if req.Channels != nil {
		for _, ch := range *req.Channels {
			if !ch.IsValid() {
				return ErrInvalidChannel
			}
		}
	}
	return nil
}

func (req *UpdateAnnouncementRequest) validateEntityType() error {
	if req.EntityType != nil && !req.EntityType.IsValid() {
		return ErrInvalidEntityType
	}
	return nil
}

func (req *UpdateAnnouncementRequest) validateExpiresAt() error {
	if req.ExpiresAt != nil && req.ExpiresAt.Before(time.Now()) {
		return ErrInvalidExpiresAt
	}
	return nil
}

func (req *UpdateAnnouncementRequest) validateConstraints() error {
	if req.EntityType != nil && req.EntityID == nil {
		return ErrEntityDataMisMatch
	}

	if req.EntityType == nil && req.EntityID != nil {
		return ErrEntityDataMisMatch
	}
	return nil
}

// Apply merges the request's non-nil fields onto p.
func (r UpdateAnnouncementRequest) Apply(p *Announcement) {
	appliers := []func(*Announcement){
		r.applyTitle,
		r.applyBody,
		r.applyEntityID,
		r.applyChannels,
		r.applyEntityType,
		r.applyExpiresAt,
		r.applyUpdatedByUserID,
	}
	for _, apply := range appliers {
		apply(p)
	}
}

func (req *UpdateAnnouncementRequest) applyTitle(an *Announcement) {
	if req.Title != nil {
		an.Title = *req.Title
	}
}

func (req *UpdateAnnouncementRequest) applyBody(an *Announcement) {
	if req.Body != nil {
		an.Body = *req.Body
	}
}

func (req *UpdateAnnouncementRequest) applyEntityID(an *Announcement) {
	if req.EntityID != nil {
		an.EntityID = req.EntityID
	}
}

func (req *UpdateAnnouncementRequest) applyChannels(an *Announcement) {
	if req.Channels != nil {
		an.Channels = *req.Channels
	}
}

func (req *UpdateAnnouncementRequest) applyEntityType(an *Announcement) {
	if req.EntityType != nil {
		an.EntityType = req.EntityType
	}
}

func (req *UpdateAnnouncementRequest) applyExpiresAt(an *Announcement) {
	if req.ExpiresAt != nil {
		an.ExpiresAt = *req.ExpiresAt
	}
}

func (req *UpdateAnnouncementRequest) applyUpdatedByUserID(an *Announcement) {
	if req.UpdatedByUserID != "" {
		an.UpdatedByUserID = &req.UpdatedByUserID
	}
}
