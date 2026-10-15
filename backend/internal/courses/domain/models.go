package coursedomain

import (
	"time"

	"learnflow_backend/internal/shared/validator"
)

// CourseStatus represents the current lifecycle state of a course.
type CourseStatus string

// CourseStatus values.
const (
	DraftStatus     CourseStatus = "draft"
	PublishedStatus CourseStatus = "published"
	ArchivedStatus  CourseStatus = "archived"
)

// Course represents a course in draft, published, or archived state.
type Course struct {
	ID                    string       `json:"id"`
	Slug                  string       `json:"slug"`
	Title                 string       `json:"title"`
	Description           *string      `json:"description"`
	ThumbnailURL          *string      `json:"thumbnail_url"`
	PreviewVideoURL       *string      `json:"preview_video_url"`
	Status                CourseStatus `json:"status"`
	EstimatedMinutes      *int         `json:"estimated_minutes"`
	SeoTitle              *string      `json:"seo_title"`
	SeoDescription        *string      `json:"seo_description"`
	OgImageURL            *string      `json:"og_image_url"`
	CanonicalURL          *string      `json:"canonical_url"`
	IsIndexable           bool         `json:"is_indexable"`
	CreatedByUserID       string       `json:"created_by_user_id"`
	CreatedAt             time.Time    `json:"created_at"`
	UpdatedAt             *time.Time   `json:"updated_at"`
	UpdatedByUserID       *string      `json:"updated_by_user_id"`
	PublishedAt           *time.Time   `json:"published_at"`
	PublishedByUserID     *string      `json:"published_by_user_id"`
	DeletedAt             *time.Time   `json:"deleted_at"`
	DeletedByUserID       *string      `json:"deleted_by_user_id"`
	ArchivedAt            *time.Time   `json:"archived_at"`
	ArchivedByUserID      *string      `json:"archived_by_user_id"`
	Announcement          *string      `json:"announcement"`
	AnnouncementExpiresAt *time.Time   `json:"announcement_expires_at"`
}

// Valid reports whether r is one of the known CourseStatus values.
func (r CourseStatus) Valid() bool {
	switch r {
	case
		DraftStatus,
		PublishedStatus,
		ArchivedStatus:
		return true
	}
	return false
}

// ReadyToPublish reports whether the course has all fields required to go public.
func (c *Course) ReadyToPublish() error {
	checks := []func() error{
		c.checkTitleReady,
		c.checkDescriptionReady,
		c.checkMediaReady,
		c.checkSeoTitleReady,
		c.checkSeoDescriptionReady,
	}
	for _, check := range checks {
		if err := check(); err != nil {
			return err
		}
	}
	return nil
}

func (c *Course) checkTitleReady() error {
	if c.Title == "" {
		return ErrInvalidTitle
	}
	return nil
}

func (c *Course) checkDescriptionReady() error {
	if c.Description == nil || *c.Description == "" {
		return ErrInvalidDescription
	}
	return nil
}

func (c *Course) checkMediaReady() error {
	if (c.ThumbnailURL == nil || *c.ThumbnailURL == "") && (c.PreviewVideoURL == nil || *c.PreviewVideoURL == "") {
		return ErrInvalidThumbnailURL
	}
	return nil
}

func (c *Course) checkSeoTitleReady() error {
	if c.SeoTitle == nil || *c.SeoTitle == "" {
		return ErrInvalidSeoTitle
	}
	return nil
}

func (c *Course) checkSeoDescriptionReady() error {
	if c.SeoDescription == nil || *c.SeoDescription == "" {
		return ErrInvalidSeoDescription
	}
	return nil
}

// CreateCourseRequest carries the fields needed to create a new draft course.
type CreateCourseRequest struct {
	Slug             string  `json:"slug"`
	Title            string  `json:"title"`
	Description      *string `json:"description"`
	ThumbnailURL     *string `json:"thumbnail_url"`
	PreviewVideoURL  *string `json:"preview_video_url"`
	EstimatedMinutes *int    `json:"estimated_minutes"`
	SeoTitle         *string `json:"seo_title"`
	SeoDescription   *string `json:"seo_description"`
	OgImageURL       *string `json:"og_image_url"`
	CanonicalURL     *string `json:"canonical_url"`
	IsIndexable      *bool   `json:"is_indexable"`
	CreatedByUserID  string  `json:"-"`
}

// Validate checks that the create course request fields meet format requirements.
func (req *CreateCourseRequest) Validate() error {
	checks := []func() error{
		req.validateSlug,
		req.validateTitle,
		req.validateDescription,
		req.validateEstimatedMinutes,
		req.validateThumbnailURL,
		req.validatePreviewVideoURL,
		req.validateSeoTitle,
		req.validateSeoDescription,
		req.validateOgImageURL,
		req.validateCanonicalURL,
	}
	for _, check := range checks {
		if err := check(); err != nil {
			return err
		}
	}

	return nil
}

func (req *CreateCourseRequest) validateSlug() error {
	if req.Slug == "" || !validator.IsValidSlug(req.Slug) {
		return ErrInvalidSlug
	}
	return nil
}

func (req *CreateCourseRequest) validateTitle() error {
	if !validator.IsValidContentTitle(req.Title) {
		return ErrInvalidTitle
	}
	return nil
}

func (req *CreateCourseRequest) validateDescription() error {
	return validator.RequireOptionalContentDescription(req.Description, ErrInvalidDescription)
}

func (req *CreateCourseRequest) validateSeoTitle() error {
	return validator.RequireOptionalSeoTitle(req.SeoTitle, ErrInvalidSeoTitle)
}

func (req *CreateCourseRequest) validateSeoDescription() error {
	return validator.RequireOptionalSeoDescription(req.SeoDescription, ErrInvalidSeoDescription)
}

func (req *CreateCourseRequest) validateEstimatedMinutes() error {
	return validator.RequireOptionalPositiveInt(req.EstimatedMinutes, ErrInvalidEstimatedMinutes)
}

func (req *CreateCourseRequest) validateThumbnailURL() error {
	return validator.RequireOptionalHTTPSURL(req.ThumbnailURL, ErrInvalidThumbnailURL)
}

func (req *CreateCourseRequest) validatePreviewVideoURL() error {
	return validator.RequireOptionalHTTPSURL(req.PreviewVideoURL, ErrInvalidPreviewVideoURL)
}

func (req *CreateCourseRequest) validateOgImageURL() error {
	return validator.RequireOptionalHTTPSURL(req.OgImageURL, ErrInvalidOgImageURL)
}

func (req *CreateCourseRequest) validateCanonicalURL() error {
	return validator.RequireOptionalHTTPSURL(req.CanonicalURL, ErrInvalidCanonicalURL)
}

// UpdateCourseRequest carries the fields to patch onto an existing course; nil fields
// are left unchanged.
type UpdateCourseRequest struct {
	ID               string  `json:"id"`
	Slug             *string `json:"slug"`
	Title            *string `json:"title"`
	Description      *string `json:"description"`
	ThumbnailURL     *string `json:"thumbnail_url"`
	PreviewVideoURL  *string `json:"preview_video_url"`
	EstimatedMinutes *int    `json:"estimated_minutes"`
	SeoTitle         *string `json:"seo_title"`
	SeoDescription   *string `json:"seo_description"`
	OgImageURL       *string `json:"og_image_url"`
	CanonicalURL     *string `json:"canonical_url"`
	IsIndexable      *bool   `json:"is_indexable"`
}

// Validate checks that the update course request fields meet format requirements.
func (req *UpdateCourseRequest) Validate() error {
	checks := []func() error{
		req.validateID,
		req.validateSlug,
		req.validateTitle,
		req.validateDescription,
		req.validateThumbnailURL,
		req.validatePreviewVideoURL,
		req.validateSeoTitle,
		req.validateSeoDescription,
		req.validateOgImageURL,
		req.validateCanonicalURL,
		req.validateEstimatedMinutes,
	}
	for _, check := range checks {
		if err := check(); err != nil {
			return err
		}
	}

	return nil
}

func (req *UpdateCourseRequest) validateID() error {
	if req.ID == "" || !validator.IsValidUUID(req.ID) {
		return ErrInvalidCourseID
	}
	return nil
}

func (req *UpdateCourseRequest) validateSlug() error {
	if req.Slug != nil && (*req.Slug == "" || !validator.IsValidSlug(*req.Slug)) {
		return ErrInvalidSlug
	}
	return nil
}

func (req *UpdateCourseRequest) validateTitle() error {
	if req.Title != nil && !validator.IsValidContentTitle(*req.Title) {
		return ErrInvalidTitle
	}
	return nil
}

func (req *UpdateCourseRequest) validateDescription() error {
	return validator.RequireOptionalContentDescription(req.Description, ErrInvalidDescription)
}

func (req *UpdateCourseRequest) validateThumbnailURL() error {
	return validator.RequireOptionalHTTPSURL(req.ThumbnailURL, ErrInvalidThumbnailURL)
}

func (req *UpdateCourseRequest) validatePreviewVideoURL() error {
	return validator.RequireOptionalHTTPSURL(req.PreviewVideoURL, ErrInvalidPreviewVideoURL)
}

func (req *UpdateCourseRequest) validateSeoTitle() error {
	return validator.RequireOptionalSeoTitle(req.SeoTitle, ErrInvalidSeoTitle)
}

func (req *UpdateCourseRequest) validateSeoDescription() error {
	return validator.RequireOptionalSeoDescription(req.SeoDescription, ErrInvalidSeoDescription)
}

func (req *UpdateCourseRequest) validateOgImageURL() error {
	return validator.RequireOptionalHTTPSURL(req.OgImageURL, ErrInvalidOgImageURL)
}

func (req *UpdateCourseRequest) validateCanonicalURL() error {
	return validator.RequireOptionalHTTPSURL(req.CanonicalURL, ErrInvalidCanonicalURL)
}

func (req *UpdateCourseRequest) validateEstimatedMinutes() error {
	return validator.RequireOptionalPositiveInt(req.EstimatedMinutes, ErrInvalidEstimatedMinutes)
}

// Apply copies every non-nil field from r onto p.
func (r UpdateCourseRequest) Apply(p *Course) {
	appliers := []func(*Course){
		r.applySlug,
		r.applyTitle,
		r.applyDescription,
		r.applyThumbnailURL,
		r.applyPreviewVideoURL,
		r.applyEstimatedMinutes,
		r.applySeoTitle,
		r.applySeoDescription,
		r.applyOgImageURL,
		r.applyCanonicalURL,
		r.applyIsIndexable,
	}
	for _, apply := range appliers {
		apply(p)
	}
}

func (r UpdateCourseRequest) applySlug(p *Course) {
	if r.Slug != nil {
		p.Slug = *r.Slug
	}
}

func (r UpdateCourseRequest) applyTitle(p *Course) {
	if r.Title != nil {
		p.Title = *r.Title
	}
}

func (r UpdateCourseRequest) applyDescription(p *Course) {
	if r.Description != nil {
		p.Description = r.Description
	}
}

func (r UpdateCourseRequest) applyThumbnailURL(p *Course) {
	if r.ThumbnailURL != nil {
		p.ThumbnailURL = r.ThumbnailURL
	}
}

func (r UpdateCourseRequest) applyPreviewVideoURL(p *Course) {
	if r.PreviewVideoURL != nil {
		p.PreviewVideoURL = r.PreviewVideoURL
	}
}

func (r UpdateCourseRequest) applyEstimatedMinutes(p *Course) {
	if r.EstimatedMinutes != nil {
		p.EstimatedMinutes = r.EstimatedMinutes
	}
}

func (r UpdateCourseRequest) applySeoTitle(p *Course) {
	if r.SeoTitle != nil {
		p.SeoTitle = r.SeoTitle
	}
}

func (r UpdateCourseRequest) applySeoDescription(p *Course) {
	if r.SeoDescription != nil {
		p.SeoDescription = r.SeoDescription
	}
}

func (r UpdateCourseRequest) applyOgImageURL(p *Course) {
	if r.OgImageURL != nil {
		p.OgImageURL = r.OgImageURL
	}
}

func (r UpdateCourseRequest) applyCanonicalURL(p *Course) {
	if r.CanonicalURL != nil {
		p.CanonicalURL = r.CanonicalURL
	}
}

func (r UpdateCourseRequest) applyIsIndexable(p *Course) {
	if r.IsIndexable != nil {
		p.IsIndexable = *r.IsIndexable
	}
}
