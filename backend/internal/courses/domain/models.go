package coursedomain

import (
	"strings"
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
	ID               string       `json:"id"`
	Slug             string       `json:"slug"`
	Title            string       `json:"title"`
	Description      *string      `json:"description"`
	ThumbnailURL     *string      `json:"thumbnail_url"`
	PreviewVideoURL  *string      `json:"preview_video_url"`
	Status           CourseStatus `json:"status"`
	EstimatedMinutes *int         `json:"estimated_minutes"`
	SeoTitle         *string      `json:"seo_title"`
	SeoDescription   *string      `json:"seo_description"`
	OgImageURL       *string      `json:"og_image_url"`
	CanonicalURL     *string      `json:"canonical_url"`
	IsIndexable      bool         `json:"is_indexable"`
	CreatedByUserID  string       `json:"created_by_user_id"`
	CreatedAt        time.Time    `json:"created_at"`
	UpdatedAt        time.Time    `json:"updated_at"`
	PublishedAt      *time.Time   `json:"published_at"`
	DeletedAt        *time.Time   `json:"deleted_at"`
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
		req.validateEstimatedMinutes,
		req.validateThumbnailURL,
		req.validatePreviewVideoURL,
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
	if strings.TrimSpace(req.Title) == "" {
		return ErrInvalidTitle
	}
	return nil
}

func (req *CreateCourseRequest) validateEstimatedMinutes() error {
	if req.EstimatedMinutes != nil && *req.EstimatedMinutes <= 0 {
		return ErrInvalidEstimatedMinutes
	}
	return nil
}

func (req *CreateCourseRequest) validateThumbnailURL() error {
	if req.ThumbnailURL != nil && (*req.ThumbnailURL == "" || !validator.IsValidHTTPSURL(*req.ThumbnailURL)) {
		return ErrInvalidThumbnailURL
	}
	return nil
}

func (req *CreateCourseRequest) validatePreviewVideoURL() error {
	if req.PreviewVideoURL != nil && (*req.PreviewVideoURL == "" || !validator.IsValidHTTPSURL(*req.PreviewVideoURL)) {
		return ErrInvalidPreviewVideoURL
	}
	return nil
}

func (req *CreateCourseRequest) validateOgImageURL() error {
	if req.OgImageURL != nil && (*req.OgImageURL == "" || !validator.IsValidHTTPSURL(*req.OgImageURL)) {
		return ErrInvalidOgImageURL
	}
	return nil
}

func (req *CreateCourseRequest) validateCanonicalURL() error {
	if req.CanonicalURL != nil && (*req.CanonicalURL == "" || !validator.IsValidHTTPSURL(*req.CanonicalURL)) {
		return ErrInvalidCanonicalURL
	}
	return nil
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
	if req.ID == "" {
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
	if req.Title != nil && strings.TrimSpace(*req.Title) == "" {
		return ErrInvalidTitle
	}
	return nil
}

func (req *UpdateCourseRequest) validateDescription() error {
	if req.Description != nil && *req.Description == "" {
		return ErrInvalidDescription
	}
	return nil
}

func (req *UpdateCourseRequest) validateThumbnailURL() error {
	if req.ThumbnailURL != nil && (*req.ThumbnailURL == "" || !validator.IsValidHTTPSURL(*req.ThumbnailURL)) {
		return ErrInvalidThumbnailURL
	}
	return nil
}

func (req *UpdateCourseRequest) validatePreviewVideoURL() error {
	if req.PreviewVideoURL != nil && (*req.PreviewVideoURL == "" || !validator.IsValidHTTPSURL(*req.PreviewVideoURL)) {
		return ErrInvalidPreviewVideoURL
	}
	return nil
}

func (req *UpdateCourseRequest) validateSeoTitle() error {
	if req.SeoTitle != nil && *req.SeoTitle == "" {
		return ErrInvalidSeoTitle
	}
	return nil
}

func (req *UpdateCourseRequest) validateSeoDescription() error {
	if req.SeoDescription != nil && *req.SeoDescription == "" {
		return ErrInvalidSeoDescription
	}
	return nil
}

func (req *UpdateCourseRequest) validateOgImageURL() error {
	if req.OgImageURL != nil && (*req.OgImageURL == "" || !validator.IsValidHTTPSURL(*req.OgImageURL)) {
		return ErrInvalidOgImageURL
	}
	return nil
}

func (req *UpdateCourseRequest) validateCanonicalURL() error {
	if req.CanonicalURL != nil && (*req.CanonicalURL == "" || !validator.IsValidHTTPSURL(*req.CanonicalURL)) {
		return ErrInvalidCanonicalURL
	}
	return nil
}

func (req *UpdateCourseRequest) validateEstimatedMinutes() error {
	if req.EstimatedMinutes != nil && *req.EstimatedMinutes <= 0 {
		return ErrInvalidEstimatedMinutes
	}
	return nil
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
